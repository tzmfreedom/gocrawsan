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

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("version=%s revision=%s\n", c.App.Version, Revision)
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
		},
		cli.BoolFlag{
			Name: "no-redirect",
		},
		cli.IntFlag{
			Name:  "depth",
			Value: 1,
		},
		cli.BoolFlag{
			Name: "debug, D",
		},
	}
	app.Action = func(c *cli.Context) error {
		var config string
		var err error
		if c.String("config") == "" {
			config, err = createConfigFile()
			if err != nil {
				return err
			}
		} else {
			config = c.String("config")
		}

		cr := NewCrawler()
		cr.useragent = c.String("useragent")
		client := &http.Client{}
		if c.Bool("no-redirect") {
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}
		cr.client = client

		buf, err := ioutil.ReadFile(config)
		if err != nil {
			return err
		}
		file := &configFile{}
		err = toml.Unmarshal(buf, file)
		if err != nil {
			return err
		}
		cr.crawling(file.Urls, c.Int("depth"))
		return nil
	}
	app.Run(os.Args)
}

type Crawler struct {
	m            *sync.Mutex
	wg           *sync.WaitGroup
	useragent    string
	client       *http.Client
	accessedUrls map[string]struct{}
}

func NewCrawler() *Crawler {
	c := &Crawler{
		wg: new(sync.WaitGroup),
		m:  new(sync.Mutex),
		accessedUrls: make(map[string]struct{}),
	}
	return c
}

func (c *Crawler) crawling(urls []string, depth int) error {
	for _, url := range urls {
		c.m.Lock()
		if _, ok := c.accessedUrls[url]; ok {
			c.m.Unlock()
			continue
		}
		c.accessedUrls[url] = struct{}{}
		c.m.Unlock()

		c.wg.Add(1)
		go c.getUrl(url, c.printHttpStatus, depth)
	}
	c.wg.Wait()
	return nil
}

func (c *Crawler) getUrl(url string, f func(string, *http.Response), d int) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", c.useragent)
	resp, _ := c.client.Do(req)

	d -= 1
	f(url, resp)
	c.accessToNext(resp, f, d)
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

func (c *Crawler) accessToNext(resp *http.Response, f func(string, *http.Response), d int) error {
	if d <= 0 {
		return nil
	}
	links, err := getLinks(resp)
	if err != nil {
		return err
	}
	for _, link := range links {
		c.m.Lock()
		if _, ok := c.accessedUrls[link]; ok {
			c.m.Unlock()
			continue
		}
		c.accessedUrls[link] = struct{}{}
		c.m.Unlock()
		c.wg.Add(1)
		go c.getUrl(link, f, d)
	}
	return nil
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

func createConfigFile() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	config := filepath.Join(dir, "config.toml")
	if _, err := os.Stat(config); err != nil {
		fmt.Println("create " + config)
		ioutil.WriteFile(config, []byte("urls = [\"https://example.com\"]"), 0644)
	}
	return config, nil
}

func configDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "gocrawsan")
	return dir, nil
}
