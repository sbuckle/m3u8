package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/sbuckle/hls"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <url>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	u, err := url.Parse(os.Args[1]) // Check
	if err != nil {
		log.Fatalf("Invalid argument: %v\n", err)
	}
	p, err := hls.ParsePlaylist(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range p.Segments {
		fmt.Printf("%s\n", s.Url)
	}

	for _, v := range p.Variants {
		rel, err := u.Parse(v.Url)
		if err != nil {
			continue
		}
		fmt.Printf("%s\n", rel.String())
	}
}
