package main

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"research-toolkit/lib/extract"
	mdrender "research-toolkit/lib/md-render"
)

func main() {
	extractor, err := extract.NewExtractor()
	if err != nil {
		log.Fatal(err)
	}

	u, err := url.Parse("https://en.wikipedia.org/wiki/Polar_bear")
	if err != nil {
		log.Fatal(err)
	}

	mdTree, axTree, err := extractor.Extract(u)
	if err != nil {
		log.Fatal(err)
	}

	serializedAXTree, err := json.Marshal(axTree)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("ax-tree.json", []byte(serializedAXTree), 0777)
	if err != nil {
		log.Fatal(err)
	}

	serializedMDTree, err := json.Marshal(mdTree)
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("md-tree.json", []byte(serializedMDTree), 0777)
	if err != nil {
		log.Fatal(err)
	}

	serialized := mdrender.Render(mdTree)
	err = os.WriteFile("extracted.md", []byte(serialized), 0777)
	if err != nil {
		log.Fatal(err)
	}
}
