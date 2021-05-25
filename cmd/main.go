package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sbuckle/hls"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <url>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	p, err := hls.ParsePlaylist(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range p.Segments {
		fmt.Printf("%s\n", s.Url)
	}
}
