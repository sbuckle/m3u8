package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Playlist struct {
	EndOfList      bool
	Version        int
	TargetDuration int
	MediaSequence  int
	Segments       []Segment
	Variants       []Variant
	Media          []Media
}

type Key struct {
	Method string
	Url    string
	IV     string
}

type Segment struct {
	Duration float64
	Url      string
	Title    string
	Length   int
	Offset   int
	Key      *Key // optional
}

type Media struct {
	Type     string
	Url      string
	GroupID  string
	Language string
	Name     string
	Default  bool
	Forced   bool
}

type Variant struct {
	Bandwidth        int
	AverageBandwidth int
	Url              string
	Codecs           string
	Resolution       string
	Audio            string
	Video            string
	Subtitles        string
}

var re = regexp.MustCompile(`([-A-Z0-9]+)=("[^"\x0A\x0D]+"|[^",\s]+)`)

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
	var key *Key
	var val string
	var variant Variant
	var duration float64
	var title string
	var isSegment, isVariant bool
	var offset, length int

	linenum := 1
	for s.Scan() {
		linenum += 1
		line := s.Text()
		if line == "" || isComment(line) {
			continue
		} else if startsWith(line, "#EXT-X-VERSION:", &val) {
			if v, err := strconv.Atoi(val); err == nil {
				pl.Version = v
			}
		} else if startsWith(line, "#EXT-X-MEDIA-SEQUENCE:", &val) {
			if ms, err := strconv.Atoi(val); err == nil {
				pl.MediaSequence = ms
			}
		} else if startsWith(line, "#EXT-X-BYTERANGE:", &val) {
			length, offset = parseByteRange(val)
		} else if startsWith(line, "#EXT-X-STREAM-INF:", &val) {
			variant = parseVariant(val)
			isVariant = true
		} else if startsWith(line, "#EXT-X-MEDIA:", &val) {
			pl.Media = append(pl.Media, parseMedia(val))
		} else if startsWith(line, "#EXT-X-TARGETDURATION:", &val) {
			if t, err := strconv.Atoi(val); err == nil {
				pl.TargetDuration = t
			}
		} else if line == "#EXT-X-ENDLIST" {
			pl.EndOfList = true
		} else if startsWith(line, "#EXT-X-KEY:", &val) {
			key = parseKey(val)
		} else if startsWith(line, "#EXTINF:", &val) {
			duration, title, _ = parseExtInf(val)
			isSegment = true
		} else if isSegment {
			segment := Segment{
				Duration: duration,
				Title:    title,
				Url:      line,
				Length:   length,
				Offset:   offset,
			}
			if key != nil {
				segment.Key = key
			}
			pl.Segments = append(pl.Segments, segment)
			isSegment = false
		} else if isVariant {
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

func parseKey(val string) *Key {
	key := new(Key)
	for k, v := range parseAttributeList(val) {
		switch k {
		case "METHOD":
			key.Method = v
		case "IV":
			key.IV = v
		case "URI":
			key.Url = v
		}
	}
	return key
}

func parseMedia(val string) Media {
	m := Media{}
	for k, v := range parseAttributeList(val) {
		switch k {
		case "GROUP-ID":
			m.GroupID = v
		case "TYPE":
			m.Type = v
		case "LANGUAGE":
			m.Language = v
		case "DEFAULT":
			if v == "YES" {
				m.Default = true
			}
		case "FORCED":
			if v == "YES" {
				m.Forced = true
			}
		case "URI":
			m.Url = v
		case "NAME":
			m.Name = v
		}
	}
	return m
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
		case "AUDIO":
			variant.Audio = v
		case "VIDEO":
			variant.Video = v
		case "SUBTITLES":
			variant.Subtitles = v
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

func parseByteRange(value string) (length int, offset int) {
	if value == "" {
		return
	}
	res := strings.Split(value, "@")
	if n, err := strconv.Atoi(res[0]); err == nil {
		length = n
	}
	if len(res) > 1 {
		if o, err := strconv.Atoi(res[1]); err == nil {
			offset = o
		}
	}
	return
}

func parseAttributeList(value string) map[string]string {
	attrs := make(map[string]string)
	for _, result := range re.FindAllStringSubmatch(value, -1) {
		attrs[result[1]] = result[2]
	}
	return attrs
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
		fmt.Printf("%s %f %d %d\n", s.Url, s.Duration, s.Length, s.Offset)
		if s.Key != nil {
			fmt.Printf("%s\n", s.Key.Url)
		}
	}
	for _, v := range pl.Variants {
		fmt.Printf("%v\n", v)
	}

	for _, m := range pl.Media {
		fmt.Printf("%v\n", m)
	}
}
