package main

import (
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/yankeguo/rg"
)

const (
	DirOut = "out"
)

var (
	AllowedExtensions = []string{
		".html",
		".htm",
	}
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	panic(err)
}

func calculateFilename(dir string, u *url.URL) string {
	relPath := strings.TrimSuffix(strings.TrimPrefix(u.Path, "/"), "/")
	if ext := path.Ext(relPath); ext == "" {
		relPath = relPath + "/__index.html"
	}
	return filepath.Join(dir, u.Hostname(), filepath.FromSlash(relPath))
}

func shouldSkip(u *url.URL, base *url.URL) bool {
	if u.Host != base.Host || !strings.HasPrefix(u.Path, base.Path) {
		return true
	}

	if ext := strings.ToLower(path.Ext(path.Base(u.Path))); ext != "" {
		for _, allowed := range AllowedExtensions {
			if ext == allowed {
				goto extPassed
			}
		}
		return true
	}

extPassed:

	if filename := calculateFilename(DirOut, u); fileExists(filename) {
		return true
	}

	return false
}

func main() {
	args := os.Args[1:]

	if len(args) != 1 {
		log.Println("mallam-scrape: requires 1 and only 1 url")
		return
	}

	urlBase := rg.Must(url.Parse(args[0]))

	c := colly.NewCollector(
		colly.Async(true),
		colly.IgnoreRobotsTxt(),
	)

	rg.Must0(c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       time.Millisecond * 500,
		RandomDelay: time.Millisecond * 200,
		Parallelism: 4,
	}))

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		href := strings.TrimSpace(e.Attr("href"))

		if href == "" {
			return
		}

		var (
			u   *url.URL
			err error
		)

		if u, err = url.Parse(href); err != nil {
			log.Println("failed parsing href:", href)
			return
		}

		u = e.Request.URL.ResolveReference(u)

		if shouldSkip(u, urlBase) {
			log.Println("skip:", u.String())
			return
		}

		e.Request.Visit(u.String())
	})

	c.OnResponse(func(res *colly.Response) {
		filename := calculateFilename(DirOut, res.Request.URL)
		dirname := filepath.Dir(filename)
		rg.Must0(os.MkdirAll(dirname, 0755))
		rg.Must0(os.WriteFile(filename, res.Body, 0640))
		log.Println("done:", res.Request.URL.String())
	})

	rg.Must0(c.Visit(urlBase.String()))

	c.Wait()
}
