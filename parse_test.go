package m3u8

import (
	"os"
	"reflect"
	"testing"
)

func TestParseByteRange(t *testing.T) {
	var tests = []struct {
		input  string
		length int
		offset int
	}{
		{"1234@5678", 1234, 5678},
		{"1234", 1234, 0},
		{"", 0, 0},
	}
	for _, test := range tests {
		if l, o := parseByteRange(test.input); l != test.length || o != test.offset {
			t.Errorf("parseByteRange(%q) = %d %d", test.input, l, o)
		}
	}
}

func TestParseAttributeList(t *testing.T) {
	var tests = []struct {
		input string
		want  map[string]string
	}{
		{
			`BANDWIDTH=3389529,CODECS="mp4a.40.2,avc1.4d401e"`,
			map[string]string{"BANDWIDTH": "3389529", "CODECS": `mp4a.40.2,avc1.4d401e`},
		},
		{
			`URI="https://priv.example.com/key.php?r=52"`,
			map[string]string{"URI": `https://priv.example.com/key.php?r=52`},
		},
		{
			"",
			map[string]string{},
		},
	}
	for _, test := range tests {
		if got := parseAttributeList(test.input); !reflect.DeepEqual(test.want, got) {
			t.Errorf("parseAttributeList(%q) = %q", test.input, got)
		}
	}
}

func TestParseMasterPlaylist(t *testing.T) {
	f, err := os.Open("testdata/master.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p, err := Parse(f)
	if err != nil {
		t.Errorf("Failed to parse playlist: %v", err)
	}
	if p.Version != 4 {
		t.Errorf("Playlist version is wrong. Got %d, expected 4", p.Version)
	}
	if len(p.Variants) != 4 {
		t.Errorf("Not all variants in the master playlist parsed. Got %d, expected 4", len(p.Variants))
	}
}

func TestParseMediaPlaylist(t *testing.T) {
	f, err := os.Open("testdata/media.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p, err := Parse(f)
	if err != nil {
		t.Errorf("Failed to parse playlist: %v", err)
	}
	if p.Version != 3 {
		t.Errorf("Playlist version is wrong. Got %d, expected 4", p.Version)
	}
	if p.TargetDuration != 10 {
		t.Errorf("Target duration. Got %d, expected 10", p.TargetDuration)
	}
	if len(p.Segments) != 3 {
		t.Errorf("Not all segments in the playlist parsed. Got %d, expected 3", len(p.Segments))
	}
	expected := []Segment{
		{Url: "http://media.example.com/first.ts", Duration: 9.009},
		{Url: "http://media.example.com/second.ts", Duration: 9.009},
		{Url: "http://media.example.com/third.ts", Duration: 3.003},
	}
	for i, segment := range p.Segments {
		if !reflect.DeepEqual(segment, expected[i]) {
			t.Errorf("Got: %+v Expected: %+v", segment, expected[i])
		}
	}
}
