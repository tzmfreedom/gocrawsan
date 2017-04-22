package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	_ "github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

var (
	Version   string
	Revision  string
	useragent string = "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1)"
)

type configFile struct {
	Urls []string `toml:"urls"`
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
			Name: "urls",
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
	buf, err := ioutil.ReadFile(c.String("config"))
	if err != nil {
		return err
	}
	file := &configFile{}
	err = toml.Unmarshal(buf, file)

	for _, url := range file.Urls {
		wg.Add(1)
		go getUrl(wg, m, url)
	}
	wg.Wait()
	return nil
}

func getUrl(wg *sync.WaitGroup, m *sync.Mutex, url string) {
	client := &http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", useragent)

	resp, _ := client.Do(req)
	status := strings.Split(resp.Status, " ")
	code, _ := strconv.Atoi(status[0])

	m.Lock()
	defer m.Unlock()
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
	wg.Done()
}
