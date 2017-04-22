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
			Name:  "config, C",
			Value: filepath.Join(dir, "config.toml"),
		},
		cli.BoolFlag{
			Name: "debug, D",
		},
	}
	app.Action = crawling
	app.Run(os.Args)
}

func crawling(c *cli.Context) error {
	wg := new(sync.WaitGroup)
	m := new(sync.Mutex)

	ua := c.String("useragent")
	buf, err := ioutil.ReadFile(c.String("config"))
	if err != nil {
		return err
	}
	file := &configFile{}
	err = toml.Unmarshal(buf, file)
	if err != nil {
		return err
	}

	for _, url := range file.Urls {
		wg.Add(1)
		s := &requestSetting{
			url:       url,
			useragent: ua,
		}
		go getUrl(wg, m, s)
	}
	wg.Wait()
	return nil
}

func getUrl(wg *sync.WaitGroup, m *sync.Mutex, s *requestSetting) {
	client := &http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	req, _ := http.NewRequest("GET", s.url, nil)
	req.Header.Set("User-Agent", s.useragent)

	resp, _ := client.Do(req)
	status := strings.Split(resp.Status, " ")
	code, _ := strconv.Atoi(status[0])

	m.Lock()
	fmt.Print(s.url + "\t")

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
	m.Unlock()
	_, _ := getLinks(resp)
	wg.Done()
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
