# 🌱 goClone

<a href="https://github.com/shurco/goClone/releases"><img src="https://img.shields.io/github/v/release/shurco/goclone?sort=semver&label=Release&color=651FFF"></a>
<a href="https://goreportcard.com/report/github.com/shurco/goClone"><img src="https://goreportcard.com/badge/github.com/shurco/goClone"></a>
<a href="https://www.codefactor.io/repository/github/shurco/goclone"><img src="https://www.codefactor.io/repository/github/shurco/goclone/badge" alt="CodeFactor" /></a>
<a href="https://github.com/shurco/goClone/actions/workflows/release.yml"><img src="https://github.com/shurco/goClone/actions/workflows/release.yml/badge.svg"></a>
<a href="https://github.com/shurco/goClone/blob/master/LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg"></a>


goClone is a powerful utility that enables you to effortlessly download entire websites from the Internet and save them to your local directory. With goClone, you can easily obtain HTML, CSS, JavaScript, images, and other associated files directly from the server and store them on your computer.

One of the standout features of goClone is its ability to accurately preserve the original website's relative link structure. This means that when you open any page of the "mirrored" website in your browser, you can seamlessly navigate through the site by following links just as if you were browsing it online.

goClone empowers you to have offline access to websites, making it convenient for various purposes such as research, archiving, or simply enjoying a website without an internet connection.

So go ahead, give goClone a try and experience the freedom of having your favorite websites at your fingertips, even when you're offline!

![Example](/.github/media/example.gif)

<a name="macos"></a>
## MacOS installing

```shell
$ brew install shurco/tap/goclone
```

Alternately, you can configure the tap and install the package separately:

``` shell
$ brew tap shurco/tap
$ brew install goclone
```


<a name="manual"></a>

## Manual

```bash
# go get :)
go get github.com/shurco/goClone
# change to project directory using your GOPATH
cd $GOPATH/src/github.com/shurco/goClone/cmd
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
  -b, --browser_endpoint string        chrome headless browser WS endpoint
  -c, --cookie                         if set true, cookies won't send
  -h, --help                           help for goclone
  -o, --open                           automatically open project in default browser
  -p, --proxy_string string            proxy connection string
  -r, --robots                         disable robots.txt checks
  -s, --serve                          serve the generated files using gofiber
  -P, --servePort int                  serve port number (default 8088)
  -u, --user_agent string              custom User-Agent (default "goclone")
      --assets_root string             root directory for downloaded assets (default "assets")
      --max_concurrent_downloads int   maximum number of concurrent downloads (default 8)
      --max_download_mb int            maximum size of a downloaded asset in MB (default 50)
      --http_timeout_seconds int       HTTP request timeout for asset downloads (seconds) (default 20)
  -v, --verbose                        enable verbose logging
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

## JS-rendered pages (browser_endpoint)

To crawl JS-rendered pages, run with a Chrome DevTools endpoint and pass it via `-b/--browser_endpoint`.

Example:

```bash
# Start Chrome headless with DevTools (Docker)
docker run -d -p 9222:9222 --rm --name headless chromedp/headless-shell:stable
# Retrieve full DevTools WS URL (optional)
curl -s http://127.0.0.1:9222/json/version | jq -r .webSocketDebuggerUrl
# Run goclone with rendering
goclone -b ws://127.0.0.1:9222 https://example.com
```

### Linux notes (Docker networking)
If you run Chrome inside Docker on Linux, the container might not be able to reach `http://127.0.0.1:<port>` served on the host. Use one of the following approaches:

- Host networking (simple and recommended for local testing):
  ```bash
  docker rm -f headless || true
  docker run -d --net=host --rm --name headless chromedp/headless-shell:stable
  goclone -b ws://127.0.0.1:9222 http://127.0.0.1:18088
  ```
- Shared Docker network with a server in a container:
  ```bash
  docker network create goclone-net || true
  docker run -d --network goclone-net --name headless chromedp/headless-shell:stable
  docker run -d --network goclone-net --name web -p 18088:18088 your/web:latest
  goclone -b ws://127.0.0.1:9222 http://web:18088
  ```
- Use the host IP address reachable from the container (instead of 127.0.0.1):
  ```bash
  HOST_IP=$(ip route | awk '/default/ {print $3}')
  goclone -b ws://127.0.0.1:9222 http://$HOST_IP:18088
  ```

If rendering produces errors like `ERR_CONNECTION_REFUSED` or `ERR_NAME_NOT_RESOLVED`, adjust Docker networking as above and ensure you pass the full DevTools WebSocket URL from `/json/version` if necessary.

