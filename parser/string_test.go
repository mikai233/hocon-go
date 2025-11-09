package parser

import "testing"

func TestParseQuotedString(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		rest     string
	}{
		{`"hello"`, "hello", ""},
		{`"hello\nworld"`, "hello\nworld", ""},
		{`"line1\nline2\tindent\\slash\"quote"`, "line1\nline2\tindent\\slash\"quote", ""},
		{`"\u4F60\u597D"`, "你好", ""},
		{`"\uD83D\uDE00"`, "😀", ""},
		{`"Hello \u4F60\u597D \n 😀!"`, "Hello 你好 \n 😀!", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			p := newTestParser(tc.input)
			got, err := p.parseQuotedString(true)
			if err != nil {
				t.Fatalf("parseQuotedString(%s) error: %v", tc.input, err)
			}
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
			if rest := remainingInput(p); rest != tc.rest {
				t.Fatalf("expected rest %q, got %q", tc.rest, rest)
			}
		})
	}
}

func TestParseQuotedStringInvalid(t *testing.T) {
	cases := []string{
		`"Hello \`,
		`"\uD83D\u0041"`,
		``,
		`"`,
		`"\u`,
		`"\uD83`,
		`"\uD83D\u004\` + "`" + `"`,
		`"\uD83D\u004"`,
	}
	for _, input := range cases {
		p := newTestParser(input)
		if _, err := p.parseQuotedString(true); err == nil {
			t.Fatalf("expected error for %q", input)
		}
	}
}

func TestParseUnquotedString(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		rest     string
	}{
		{"a.b.c", "a.b.c", ""},
		{"a.b.c//", "a.b.c", "//"},
		{"a.b.c/b", "a.b.c/b", ""},
		{"hello#world", "hello", "#world"},
		{"你 好", "你", " 好"},
		{"你 \\r\n不好", "你", " \\r\n不好"},
		{"你 \r\n不好", "你", " \r\n不好"},
		{"a，\n", "a，", "\n"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			p := newTestParser(tc.input)
			val, err := p.parseUnquotedString()
			if err != nil {
				t.Fatalf("parseUnquotedString(%s) error: %v", tc.input, err)
			}
			str := val.(interface{ String() string }).String()
			if str != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, str)
			}
			if rest := remainingInput(p); rest != tc.rest {
				t.Fatalf("expected rest %q, got %q", tc.rest, rest)
			}
		})
	}
}

func TestParseMultilineString(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		rest     string
	}{
		{`"""a.bbc"""`, "a.bbc", ""},
		{`"""a.bbc😀"""😀`, "a.bbc😀", "😀"},
		{`"""a.b\r\nbc"""`, "a.b\\r\\nbc", ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			p := newTestParser(tc.input)
			got, err := p.parseMultilineString(true)
			if err != nil {
				t.Fatalf("parseMultilineString(%s) error: %v", tc.input, err)
			}
			if got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
			if rest := remainingInput(p); rest != tc.rest {
				t.Fatalf("expected rest %q, got %q", tc.rest, rest)
			}
		})
	}
}

func TestParseMultilineStringInvalid(t *testing.T) {
	cases := []string{
		`"`,
		`"""Hello"`,
		`""Hello"""`,
		`"Hello"""`,
	}
	for _, input := range cases {
		p := newTestParser(input)
		if _, err := p.parseMultilineString(true); err == nil {
			t.Fatalf("expected error for %q", input)
		}
	}
}

func TestParsePathExpression(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		rest     string
	}{
		{`a.b.c `, "a.b.c", ""},
		{`a. b.c `, "a. b.c", ""},
		{`a. "..".c `, "a. ...c", ""},
		{`a.b.c :`, "a.b.c", ":"},
		{`a.b.c =`, "a.b.c", "="},
		{`a.b.c{`, "a.b.c", "{"},
		{`a. """b""" . c }`, "a. b . c", "}"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			p := newTestParser(tc.input)
			str, err := p.parsePathExpression()
			if err != nil {
				t.Fatalf("parsePathExpression(%s) error: %v", tc.input, err)
			}
			if str.String() != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, str.String())
			}
			if rest := remainingInput(p); rest != tc.rest {
				t.Fatalf("expected rest %q, got %q", tc.rest, rest)
			}
		})
	}
}
