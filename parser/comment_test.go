package parser

import (
	"hocon-go/raw"
	"testing"
)

func TestParseComment(t *testing.T) {
	cases := []struct {
		input    string
		expected raw.CommentType
		content  string
		rest     string
	}{
		{"#你好👌\r\r\n", raw.Hash, "你好👌\r", "\r\n"},
		{"#你好👌\r\n", raw.Hash, "你好👌", "\r\n"},
		{"#HelloWo\nrld👌\r\n", raw.Hash, "HelloWo", "\nrld👌\r\n"},
		{"//Hello//World\n", raw.DoubleSlash, "Hello//World", "\n"},
		{"//\r\n", raw.DoubleSlash, "", "\r\n"},
		{"#\n", raw.Hash, "", "\n"},
		{"//Hello//World", raw.DoubleSlash, "Hello//World", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			p := newTestParser(tc.input)
			ty, content, err := p.parseComment()
			if err != nil {
				t.Fatalf("parseComment(%s) error: %v", tc.input, err)
			}
			if ty != tc.expected {
				t.Fatalf("expected type %v, got %v", tc.expected, ty)
			}
			if content != tc.content {
				t.Fatalf("expected content %q, got %q", tc.content, content)
			}
			if rest := remainingInput(p); rest != tc.rest {
				t.Fatalf("expected rest %q, got %q", tc.rest, rest)
			}
		})
	}
}
