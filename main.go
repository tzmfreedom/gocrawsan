package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

var (
	Version  string
	Revision string
)

type configFile struct {
	Urls      []string `toml:"urls"`
	UserAgent string   `toml:"useragent"`
}

type requestSetting struct {
	url       string
	useragent string
	redirect  bool
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("version=%s revision=%s\n", c.App.Version, Revision)
	}
	home, err := homedir.Dir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, ".config", "gocrawsan")
	if err = os.MkdirAll(dir, 0700); err != nil {
		return
	}
	app := cli.NewApp()
	app.Name = "gocrawsan"
	app.Usage = "web crawling command utility"
	app.Version = Version
	app.Usage = ""
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name: "useragent, U",
		},
		cli.StringFlag{
			Name: "config, C",
			Value: filepath.Join(dir, "config.toml"),
		},
		cli.BoolFlag{
			Name: "redirect",
		},
		cli.BoolFlag{
			Name: "debug, D",
		},
	}
	app.Action = func (c *cli.Context) error {
		cr := NewCrawler()
		ua := c.String("useragent")
		cr.useragent = ua
		cr.redirect = c.Bool("redirect")
		buf, err := ioutil.ReadFile(c.String("config"))
		if err != nil {
			return err
		}
		file := &configFile{}
		err = toml.Unmarshal(buf, file)
		if err != nil {
			return err
		}
		cr.crawling(file.Urls)
		return nil
	}
	app.Run(os.Args)
}

type Crawler struct {
	m *sync.Mutex
	wg *sync.WaitGroup
	useragent string
	redirect bool
}

func NewCrawler() *Crawler {
	c := &Crawler{
		wg: new(sync.WaitGroup),
		m: new(sync.Mutex),
	}
	return c
}

func (c *Crawler) crawling(urls []string) error {
	for _, url := range urls {
		c.wg.Add(1)
		go c.getUrl(url, c.printHttpStatus, 1)
	}
	c.wg.Wait()
	return nil
}

func (c *Crawler) getUrl(url string, f func(string, *http.Response), d int) {
	client := &http.Client{}
	if c.redirect {
		client.CheckRedirect = func (req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
		}
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", c.useragent)

	resp, _ := client.Do(req)
	d -= 1
	f(url, resp)
	c.accessToNext(resp, d)
	c.wg.Done()
}

func (c *Crawler) printHttpStatus(url string, resp *http.Response) {
	c.m.Lock()
	status := strings.Split(resp.Status, " ")
	code, _ := strconv.Atoi(status[0])
	fmt.Print(url + "\t")
	switch code / 100 {
	case 2:
		color.Cyan(resp.Status)
	case 3:
		color.Yellow(resp.Status)
	case 4:
		color.Red(resp.Status)
	default:
		fmt.Println(resp.Status)
	}
	c.m.Unlock()
}

func (c *Crawler) accessToNext(resp *http.Response, d int) {
	if d > 0 {
		links, err := getLinks(resp)
    if err == nil {
      for _, link := range links {
        c.wg.Add(1)
        go c.getUrl(link, c.printHttpStatus, d)
      }
    }
	}
}
func getLinks(res *http.Response) ([]string, error) {
	urls := []string{}
	doc, _ := goquery.NewDocumentFromResponse(res)
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		url, _ := s.Attr("href")
		r := regexp.MustCompile(`^(https|http)://(.*)`)
		if !r.MatchString(url) {
			url = res.Request.URL.Scheme + "://" + res.Request.URL.Host + url
		}
		urls = append(urls, url)
	})
	return urls, nil
}
