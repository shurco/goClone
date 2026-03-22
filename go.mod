module github.com/shurco/goClone

go 1.26

require (
	github.com/PuerkitoBio/goquery v1.12.0
	github.com/geziyor/geziyor v0.0.0-20240812061556-229b8ca83ac1
	github.com/spf13/cobra v1.10.2
)

require (
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chromedp/cdproto v0.0.0-20260321001828-e3e3800016bc // indirect
	github.com/chromedp/chromedp v0.15.0 // indirect
	github.com/chromedp/sysutil v1.1.0 // indirect
	github.com/go-json-experiment/json v0.0.0-20260214004413-d219187c3433 // indirect
	github.com/go-kit/kit v0.13.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/procfs v0.20.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/temoto/robotstxt v1.1.2 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

// Pin cdproto to the revision chromedp v0.15.0 targets; a slightly newer cdproto changed
// css.GetComputedStyleForNode(...).Do to return three values and breaks chromedp's query.go.
replace github.com/chromedp/cdproto => github.com/chromedp/cdproto v0.0.0-20260320225252-cf654f46fc63
