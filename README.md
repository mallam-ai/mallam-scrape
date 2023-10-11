# mallam-scrape

website scrapping tool for mallam-ai

## Pre-requisites

* Install `go` from https://go.dev
* Execute `go get ./...` to install dependencies

## Tool `mallam-scrape`

```
go run ./cmd/mallam-scrape "https://www.marxists.org/archive/marx/"
```

This will scrape all urls and save to `out/www.marxists.org/../..` directory

## Tool `mallam-extract-text-marx`

```
go run ./cmd/mallam-extract-text-marx
```

This will read all HTML files in `out/www.marxists.org/archive/marx/works` and save plain text to `out/text-marx.txt`

**Internal Logic**

1. Iterate subdirectories in `archive/marx/works` with 4-digits prefixed
2. Ignore `index.htm` files
3. Collect `<p>` element without `class`
4. Combine all text together

## Credits

MALLAM Developers, MIT License
