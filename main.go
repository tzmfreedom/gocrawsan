package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

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
		cli.StringFlag{
			Name: "selector",
		},
		cli.StringFlag{
			Name: "pick-type",
		},
		cli.StringFlag{
			Name: "attribute",
		},
		cli.BoolFlag{
			Name: "no-error",
		},
		cli.IntFlag{
			Name: "timeout",
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
		err := validate(c)
		if err != nil {
			return err
		}
		cr := NewCrawler()
		cr.useragent = c.String("useragent")
		client := &http.Client{}
		client.Timeout = time.Duration(time.Duration(c.Int("timeout")) * time.Second)
		if c.Bool("no-redirect") {
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}
		cr.client = client

		file, err := readOrCreateConfigFile(c)
		if err != nil {
			return err
		}
		if file == nil {
			return nil
		}
		var f func(string, *http.Response)
		if c.String("selector") != "" {
			f = cr.printWithSelector(c.String("selector"), c.String("pick-type"), c.String("attribute"))
		} else {
			f = cr.printHttpStatus
		}
		cr.crawl(file.Urls, f, c.Int("depth"))
		if len(cr.errors) > 0 {
			return &multipleError{errors: cr.errors}
		}
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
	errors       []error
}

func NewCrawler() *Crawler {
	c := &Crawler{
		wg:           new(sync.WaitGroup),
		m:            new(sync.Mutex),
		accessedUrls: make(map[string]struct{}),
		errors:       []error{},
	}
	return c
}

func (c *Crawler) crawl(urls []string, f func(string, *http.Response), depth int) {
	for _, url := range urls {
		c.m.Lock()
		if _, ok := c.accessedUrls[url]; ok {
			c.m.Unlock()
			continue
		}
		c.accessedUrls[url] = struct{}{}
		c.m.Unlock()

		c.wg.Add(1)
		go c.getUrl(url, f, depth)
	}
	c.wg.Wait()
}

func (c *Crawler) getUrl(url string, f func(string, *http.Response), d int) {
	defer c.wg.Done()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", c.useragent)
	resp, err := c.client.Do(req)
	if err != nil {
		c.errors = append(c.errors, err)
		return
	}
	d -= 1
	f(url, resp)
	c.accessToNext(resp, f, d)
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

func (c *Crawler) printWithSelector(selector string, pickType string, pickValue string) func(string, *http.Response) {
	return func(url string, resp *http.Response) {
		c.m.Lock()
		printWithSelector(selector, pickType, pickValue, url, resp)
		c.m.Unlock()
	}
}

func printWithSelector(selector string, pickType string, pickValue string, url string, resp *http.Response) {
	doc, _ := goquery.NewDocumentFromResponse(resp)
	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		var text string
		if pickType == "text" {
			text = s.Text()
		} else if pickType == "attr" {
			text, _ = s.Attr(pickValue)
		}
		fmt.Println(text)
	})
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
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Do you create configfile in " + config + "?(y/N): ")
		answer, _ := reader.ReadString('\n')
		if answer == "y" || answer == "Y" {
			err = ioutil.WriteFile(config, []byte("urls = [\"https://example.com\"]"), 0644)
			if err != nil {
				return "", err
			}
			fmt.Println("successful to create config file.")
		}
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

type multipleError struct {
	errors []error
}

func (e *multipleError) Error() string {
	errorStrings := []string{}
	for _, err := range e.errors {
		errorStrings = append(errorStrings, err.Error())
	}
	return strings.Join(errorStrings, "\n")
}

func readOrCreateConfigFile(c *cli.Context) (*configFile, error) {
	var config string
	var err error
	if c.String("config") == "" {
		config, err = createConfigFile()
		if err != nil {
			return nil, err
		}
		if config == "" {
			return nil, nil
		}
	} else {
		config = c.String("config")
	}

	buf, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, err
	}
	file := &configFile{}
	if err = toml.Unmarshal(buf, file); err != nil {
		return nil, err
	}
	return file, nil
}

func validate(c *cli.Context) error {
	pickType := map[string]bool{
		"text": true,
		"attr": true,
	}
	if c.String("pick-type") != "" && !pickType[c.String("pick-type")] {
		return errors.New("Invalid pick-type. please set 'text' or 'attr'")
	}
	if c.String("selector") != "" && c.String("pick-type") == "" {
		return errors.New("if you set selector option, please set pick-type option too")
	}
	if c.String("pick-type") == "attr" && c.String("attribute") == "" {
		return errors.New("if your set 'attr' to pick-type option, please set attribute")
	}
	return nil
}

