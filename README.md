# Gocrawsan

Simple web crawler with golang

## Install

For Linux or macOS user
```bash
$ curl -sL http://install.freedom-man.com/goc.sh | bash
```
If you want to install zsh completion, add --zsh-completion option
```bash
$ curl -sL http://install.freedom-man.com/goc.sh | bash -s -- --zsh-completion
```
or if you get lastest version, execute following command
```bash
$ go get github.com/tzmfreedom/gocrawsan
```

## Usage

```
NAME:
   gocrawsan

USAGE:
   goc [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --useragent value, -U value
   --config value, -C value
   --no-redirect
   --selector value, -S value
   --pick-type value, -P value
   --attribute value, -A value
   --no-error
   --timeout value              (default: 10)
   --depth value, -D value      (default: 1)
   --help, -h                   show help
   --version, -v                print the version
```

You should create config file. By default, gocrawsan reads `~/.config/gocrawsan/config.toml` as config file.
```toml
urls = [
  "https://www.google.co.jp",
  "https://www.example.com",
]
```

Then, execute following command.
```bash
$ goc
```

### Crawling Depth

By default, gocrawsan crawl only urls that is configured by file (depth = 1).
If you want to recursively crawl urls, set depth option to integer value greather than 1.

For example, following command crawl urls that is configured by file and links that these contents have.
```bash
$ goc --depth 2
```


### Extract Element By Selector

By default, gocrawsan crawl and print http status code with url.
Additionaly, gocrawsan can extract element from html document by css selector.

This command extract "href" attribute on "a" tag.
```bash
$ goc --selector a --pick-type attr --attribute href
```

If you want to text value, set text to pick-type option.
```bash
$ goc --selector a --pick-type text
```

### Other Option

You can timeout for http request with timeout option.
```bash
$ goc --timeout 10 # timeout with 10 seconds
```

The user-agent option allows you to set User Agent for http request.
```bash
$ goc --useragent "Mozilla/5.0 (X11; Linux i686) AppleWebKit/535.1 (KHTML, like Gecko) Ubuntu/11.04 Chromium/14.0.825.0 Chrome/14.0.825.0 Safari/535.1"
```
