# Gocrawsan

Simple web crawler with golang(WIP)

## Install

```bash
$ curl -sL http://install.freedom-man.com/goc.sh | bash
```
if you want to install zsh completion for gocrawsan, add --zsh-completion option
```bash
$ curl -sL http://install.freedom-man.com/goc.sh | bash -s -- --zsh-completion
```

```bash
$ go get github.com/tzmfreedom/gocrawsan
```

## Usage

```
NAME:
   gocrawsans

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
   --selector value
   --pick-type value
   --attribute value
   --no-error
   --timeout value              (default: 0)
   --depth value                (default: 1)
   --debug, -D
   --help, -h                   show help
   --version, -v                print the version
```

~/.config/gocrawsan/config.toml
```
urls = ["https://www.google.co.jp", "https://www.example.com"]
```

run following command
```
$ goc
```

