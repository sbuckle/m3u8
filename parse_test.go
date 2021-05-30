package m3u8

import (
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
