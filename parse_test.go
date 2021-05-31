package m3u8

import (
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
	}
	for _, test := range tests {
		if got := parseAttributeList(test.input); !reflect.DeepEqual(test.want, got) {
			t.Errorf("parseAttributeList(%q) = %q", test.input, got)
		}
	}
}
