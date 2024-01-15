package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/karrick/godirwalk"
	"github.com/yankeguo/rg"
)

var (
	dirBase    = filepath.Join("out", "www.marxists.org", "archive", "marx", "works")
	fileOutput = filepath.Join("out", "text-marx.json")

	regexpLeading4Digits = regexp.MustCompile(`^[0-9]{4}`)
	regexpNoPrintable    = regexp.MustCompile(`[^ -~]+`)
	regexpWhitespaces    = regexp.MustCompile(`\s+`)
	regexpCitation       = regexp.MustCompile(`\[(\s|\d)+\]`)
)

func cleanContent(s string) string {
	s = regexpNoPrintable.ReplaceAllLiteralString(s, " ")
	s = regexpWhitespaces.ReplaceAllLiteralString(s, " ")
	s = regexpCitation.ReplaceAllLiteralString(s, "")
	return s
}

func main() {
	rg.Must0(os.RemoveAll(fileOutput))

	f := rg.Must(os.OpenFile(fileOutput, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0640))
	defer f.Close()

	enc := json.NewEncoder(f)

	godirwalk.Walk(dirBase, &godirwalk.Options{
		ErrorCallback: func(s string, err error) godirwalk.ErrorAction {
			return godirwalk.Halt
		},
		FollowSymbolicLinks: true,
		Callback: func(filename string, entry *godirwalk.Dirent) error {
			relPath := strings.TrimPrefix(strings.TrimPrefix(filename, dirBase), "/")

			if entry.IsDir() {
				if relPath != "" && !regexpLeading4Digits.MatchString(relPath) {
					return godirwalk.SkipThis
				}
				return nil
			}

			name := filepath.Base(filename)

			if strings.Contains(name, "index") {
				return nil
			}

			log.Println(relPath)

			buf := rg.Must(os.ReadFile(filename))
			doc := rg.Must(goquery.NewDocumentFromReader(bytes.NewReader(buf)))

			var lines []string

			doc.Find("p").Each(func(i int, sel *goquery.Selection) {
				if _, ok := sel.Attr("class"); ok {
					return
				}
				lines = append(lines, strings.TrimSpace(sel.Text()))
			})

			var title string
			doc.Find("title").Each(func(i int, selection *goquery.Selection) {
				title += selection.Text()
			})

			title = strings.TrimSpace(title)

			content := strings.Join(lines, "\n\n")
			content = cleanContent(content)

			if title == "" || content == "" {
				return nil
			}

			return enc.Encode(map[string]any{
				"title":   title,
				"content": content,
				"url":     "https://www.marxists.org/archive/marx/works/" + strings.TrimPrefix(relPath, "/"),
			})
		},
	})

}
