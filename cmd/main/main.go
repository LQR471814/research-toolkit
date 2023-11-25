package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"research-toolkit/lib/extract"
	"research-toolkit/lib/getter"
	"research-toolkit/lib/google"
	mdrender "research-toolkit/lib/md-render"
	"research-toolkit/lib/utils"
)

func main() {
	os.Mkdir("results", 0777)

	extractor, err := extract.NewExtractor()
	if err != nil {
		log.Fatal(err)
	}
	extractor.Preprocess = extract.ExtractMain

	fetch, err := getter.NewGetter("cache.db")
	if err != nil {
		log.Fatal(err)
	}

	client := google.NewClient(fetch)
	searchResults, err := client.Search("polar bears", 2)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("search results", "results", searchResults)

	parsed := utils.ParallelMap[*url.URL, string](searchResults, func(u *url.URL) string {
		mdTree, _, err := extractor.Extract(u)
		if err != nil {
			slog.Warn("failed to extract webpage contents:", "url", u, "err", err.Error())
			return ""
		}
		return mdrender.Render(mdTree)
	})

	f, err := os.Create("results/main.md")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for i, text := range parsed {
		if text == "" {
			continue
		}

		_, err := f.Write([]byte(fmt.Sprintf(
			"<hr>\nExtracted from: %s\n\n%s\n",
			searchResults[i].String(),
			text,
		)))
		if err != nil {
			log.Fatal(err)
		}
	}
}
