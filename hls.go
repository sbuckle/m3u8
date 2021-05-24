package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Playlist struct {
	EndOfList      bool
	Version        int
	TargetDuration int
	Segments       []Segment
	Variants       []Variant
}

type Segment struct {
	Duration float64
	Url      string
	Title    string
}

type Variant struct {
	Bandwidth        int
	AverageBandwidth int
	Url              string
	Codecs           string
	Resolution       string
}

func ParsePlaylist(url string) (*Playlist, error) {
	content, err := fetch(url)
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(strings.NewReader(content))
	if !s.Scan() {
		return nil, s.Err()
	}
	if s.Text() != "#EXTM3U" {
		return nil, errors.New("Playlist MUST start with an #EXTM3U tag")
	}

	var pl Playlist
	// State variables
	var val string
	var variant Variant
	var duration float64
	var title string
	var isSegment, isVariant bool

	linenum := 1
	for s.Scan() {
		linenum += 1
		line := s.Text()
		if line == "" || isComment(line) {
			continue
		} else if startsWith(line, "#EXT-X-STREAM-INF:", &val) {
			variant = parseVariant(val)
			isVariant = true
		} else if startsWith(line, "#EXT-X-TARGETDURATION:", &val) {
			t, _ := strconv.Atoi(val)
			pl.TargetDuration = t
		} else if line == "#EXT-X-ENDLIST" {
			pl.EndOfList = true
		} else if startsWith(line, "#EXTINF:", &val) {
			duration, title, _ = parseExtInf(val)
			isSegment = true
		} else if isSegment {
			if !isAbs(line) {
				line = makeAbsoluteURL(url, line)
			}
			segment := Segment{
				Duration: duration,
				Title:    title,
				Url:      line,
			}
			pl.Segments = append(pl.Segments, segment)
			isSegment = false
		} else if isVariant {
			if !isAbs(line) {
				line = makeAbsoluteURL(url, line)
			}
			variant.Url = line
			pl.Variants = append(pl.Variants, variant)
			isVariant = false
		} else {
			log.Printf("Unknown: %s\n", line)
		}
	}
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}
	return &pl, nil
}

func isComment(line string) bool {
	return line[0] == '#' && !strings.HasPrefix(line, "#EXT")
}

func parseVariant(val string) Variant {
	variant := Variant{}
	attrs := parseAttributeList(val)
	for k, v := range attrs {
		switch k {
		case "BANDWIDTH":
			if b, err := strconv.Atoi(v); err == nil {
				variant.Bandwidth = b
			}
		case "CODECS":
			variant.Codecs = v
		case "AVERAGE-BANDWIDTH":
			if ab, err := strconv.Atoi(v); err == nil {
				variant.AverageBandwidth = ab
			}
		case "RESOLUTION":
			variant.Resolution = v
		}
	}
	return variant
}

func parseExtInf(value string) (d float64, title string, err error) {
	result := strings.Split(value, ",")
	d, err = strconv.ParseFloat(result[0], 64)
	if err != nil {
		return
	}
	if len(result) == 2 {
		title = result[1]
	}
	return
}

func parseByteRange(value string) (int, int, error) {
	var length, offset int
	var err error
	idx := strings.Index(value, "@")
	if idx == -1 {
		idx = len(value)
	} else {
		if offset, err = strconv.Atoi(value[idx+1:]); err != nil {
			return 0, 0, err
		}
	}
	if length, err = strconv.Atoi(value[:idx]); err != nil {
		return 0, 0, err
	}
	return length, offset, nil
}

func scanAttributeList(data []byte, atEOF bool) (int, []byte, error) {
	insideQuote := false
	width := 0
	for i := 0; i < len(data); i += width {
		var r rune
		r, width = utf8.DecodeRune(data[i:])
		if r == '"' {
			insideQuote = !insideQuote
		} else if r == ',' && !insideQuote {
			return i + width, data[0:i], nil
		}
	}
	if atEOF && len(data) > 0 {
		return len(data), data[0:], nil
	}
	return 0, nil, nil
}

func parseAttributeList(value string) map[string]string {
	attrs := make(map[string]string)
	s := bufio.NewScanner(strings.NewReader(value))
	s.Split(scanAttributeList)
	for s.Scan() {
		result := strings.Split(s.Text(), "=")
		attrs[result[0]] = result[1]
	}
	return attrs
}

func isAbs(s string) bool {
	return strings.HasPrefix(s, "https") || strings.HasPrefix(s, "http")
}

func makeAbsoluteURL(base, rel string) string {
	p, err := url.Parse(base)
	if err != nil {
		return ""
	}
	u, err := p.Parse(rel)
	if err != nil {
		return ""
	}
	return u.String()
}

func startsWith(line, prefix string, ptr *string) bool {
	b := strings.HasPrefix(line, prefix)
	if b && ptr != nil {
		*ptr = line[len(prefix):]
	}
	return b
}

func fetch(url string) (string, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func main() {
	pl, err := ParsePlaylist(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range pl.Segments {
		fmt.Printf("%s %f\n", s.Url, s.Duration)
	}
	for _, v := range pl.Variants {
		fmt.Printf("%v\n", v)
	}
}
