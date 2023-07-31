# 🌱 goclone

<a href="https://github.com/shurco/goclone/releases"><img src="https://img.shields.io/github/v/release/shurco/goclone?sort=semver&label=Release&color=651FFF"></a>
<a href="https://goreportcard.com/report/github.com/shurco/goclone"><img src="https://goreportcard.com/badge/github.com/shurco/goclone"></a>
<a href="https://www.codefactor.io/repository/github/shurco/goclone"><img src="https://www.codefactor.io/repository/github/shurco/goclone/badge" alt="CodeFactor" /></a>
<a href="https://github.com/shurco/goclone/actions/workflows/release.yml"><img src="https://github.com/shurco/goclone/actions/workflows/release.yml/badge.svg"></a>
<a href="https://github.com/shurco/goclone/blob/master/LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg"></a>


Goclone is a powerful utility that enables you to effortlessly download entire websites from the Internet and save them to your local directory. With Goclone, you can easily obtain HTML, CSS, JavaScript, images, and other associated files directly from the server and store them on your computer.

One of the standout features of Goclone is its ability to accurately preserve the original website's relative link structure. This means that when you open any page of the "mirrored" website in your browser, you can seamlessly navigate through the site by following links just as if you were browsing it online.

Goclone empowers you to have offline access to websites, making it convenient for various purposes such as research, archiving, or simply enjoying a website without an internet connection.

So go ahead, give Goclone a try and experience the freedom of having your favorite websites at your fingertips, even when you're offline!

![Example](/.github/media/example.gif)

<a name="manual"></a>

### Manual

```bash
# go get :)
go get github.com/shurco/goclone
# change to project directory using your GOPATH
cd $GOPATH/src/github.com/shurco/goclone/cmd
# build and install application
go install
```

<a name="examples"></a>

## Examples

```bash
# goclone <url>
goclone https://domain.com
```

<a name="usage"></a>

## Usage

```
Usage:
  goclone <url> [flags]

Flags:
  -b, --browser_endpoint string   chrome headless browser WS endpoint
  -c, --cookie                    if set true, cookies won't send
  -h, --help                      help for goclone
  -o, --open                      automatically open project in default browser
  -p, --proxy_string string       proxy connection string
  -r, --robots                    disable robots.txt checks
  -s, --serve                     serve the generated files using gofiber
  -P, --servePort int             serve port number (default 8088)
  -u, --user_agent string         custom User-Agent (default "goclone")
  -v, --version                   version for goclone
```

## Making JS Rendered Requests

JS Rendered requests can be made using ```-b``` flag. For example start image :  


``` bash
docker run -d -p 9222:9222 --rm --name headless-shell chromedp/headless-shell
```

then run goclone: 

```bash
goclone -b "ws://localhost:9222" https://domain.com
```

