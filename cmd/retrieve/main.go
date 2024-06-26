package main

import (
	"context"
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
	"time"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
)

const SEARCH_TERM = "food wastage in restaurants and other corporate places"

func main() {
	os.Mkdir("results", 0777)

	extractor, err := extract.NewExtractor()
	if err != nil {
		log.Fatal(err)
	}

	fetch, err := getter.NewGetter("cache.db")
	if err != nil {
		log.Fatal(err)
	}

	googleClient := google.NewClient(fetch)
	searchResults, err := googleClient.Search(SEARCH_TERM, 2)
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

	f, err := os.Create("results/retrieved.md")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	vectordb, err := weaviate.NewClient(weaviate.Config{
		Host:   "localhost:8080",
		Scheme: "http",
	})
	if err != nil {
		log.Fatal(err)
	}

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

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		_, err = vectordb.Data().
			Creator().
			WithClassName("Website").
			WithProperties(map[string]string{
				"url":     searchResults[i].String(),
				"content": text,
			}).
			Do(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}
}
