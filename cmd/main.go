package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/shurco/goclone/pkg/crawler"
)

var (
	gitCommit = "00000000"
	version   = "0.0.1"
	buildDate = "07.07.2023"
)

func main() {
	flags := crawler.Flags{}

	rootCmd := &cobra.Command{
		Use:     "goclone <url>",
		Short:   "Clone a website with ease!",
		Long:    `Copy websites to your computer! goclone is a utility that allows you to download a website from the Internet to a local directory. Get html, css, js, images, and other files from the server to your computer. goclone arranges the original site's relative link-structure. Simply open a page of the "mirrored" website in your browser, and you can browse the site from link to link, as if you were viewing it online.`,
		Args:    cobra.ArbitraryArgs,
		Version: fmt.Sprintf("%s (%s), %s", version, gitCommit, buildDate),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				if err := cmd.Usage(); err != nil {
					log.Fatal(err)
				}
				return
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			if err := crawler.CloneSite(ctx, args, flags); err != nil {
				log.Printf("%+v", err)
				os.Exit(1)
			}
		},
	}

	pf := rootCmd.PersistentFlags()
	pf.BoolVarP(&flags.Open, "open", "o", false, "automatically open project in default browser")
	pf.BoolVarP(&flags.Serve, "serve", "s", false, "serve the generated files using gofiber")
	pf.IntVarP(&flags.ServePort, "servePort", "P", 8088, "serve port number")
	pf.StringVarP(&flags.ProxyString, "proxy_string", "p", "", "proxy connection string")
	pf.StringVarP(&flags.UserAgent, "user_agent", "u", "goclone", "custom User-Agent")
	pf.BoolVarP(&flags.Cookies, "cookie", "c", false, "if set true, cookies won't send")
	pf.BoolVarP(&flags.Robots, "robots", "r", false, "disable robots.txt checks")
	pf.StringVarP(&flags.BrowserEndpoint, "browser_endpoint", "b", "", "chrome headless browser WS endpoint")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
