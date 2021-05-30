package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/sbuckle/m3u8"
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
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	resp, err := client.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	//if resp.StatusCode/100!= 2 {
	//	log.Fatal(fmt.Errorf("Failed to fetch playlist. Got a %d response\n", resp.StatusCode))
	//}
	p, err := m3u8.Parse(resp.Body)
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
