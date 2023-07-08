Goclone is a powerful utility that enables you to effortlessly download entire websites from the Internet and save them to your local directory. With Goclone, you can easily obtain HTML, CSS, JavaScript, images, and other associated files directly from the server and store them on your computer.

One of the standout features of Goclone is its ability to accurately preserve the original website's relative link structure. This means that when you open any page of the "mirrored" website in your browser, you can seamlessly navigate through the site by following links just as if you were browsing it online.

Goclone empowers you to have offline access to websites, making it convenient for various purposes such as research, archiving, or simply enjoying a website without an internet connection.

So go ahead, give Goclone a try and experience the freedom of having your favorite websites at your fingertips, even when you're offline!


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



## Usage

```
Usage:
  goclone <url> [flags]

Flags:
  -c, --cookie                if set true, cookies won't send.
  -h, --help                  help for goclone
  -o, --open                  automatically open project in default browser
  -p, --proxy_string string   proxy connection string.
  -s, --serve                 serve the generated files using gofiber.
  -P, --servePort int         serve port number. (default 8088)
  -u, --user_agent string     custom User Agent. (default "gocloner")
  -v, --version               version for goclone
```
