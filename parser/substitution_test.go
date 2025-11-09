package parser

import "testing"

func TestParseSubstitutionValid(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{`${a}`, `${a}`},
		{`${foo .bar }`, `${foo .bar}`},
		{`${a. b."c"}`, `${a. b.c}`},
		{`${? a. b."c"}`, `${?a. b.c}`},
		{`${? """a""". b."c"}`, `${?a. b.c}`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			p := newTestParser(tc.input)
			sub, err := p.parseSubstitution()
			if err != nil {
				t.Fatalf("parseSubstitution(%s) error: %v", tc.input, err)
			}
			if sub.String() != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, sub.String())
			}
		})
	}
}

func TestParseSubstitutionInvalid(t *testing.T) {
	cases := []string{
		`${foo .bar`,
		`${ ?foo.bar}`,
		`${?foo.bar.}`,
		`${?foo.bar`,
	}
	for _, input := range cases {
		p := newTestParser(input)
		if _, err := p.parseSubstitution(); err == nil {
			t.Fatalf("expected error for %q", input)
		}
	}
}
