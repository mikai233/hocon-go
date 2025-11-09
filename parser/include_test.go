package parser

import "testing"

func TestParseIncludeValid(t *testing.T) {
	cases := []struct {
		input    string
		expected string
		rest     string
	}{
		{`include "demo".conf`, `include "demo"`, `.conf`},
		{`include"demo.conf"`, `include "demo.conf"`, ``},
		{`include   required(    "demo.conf" )`, `include required("demo.conf")`, ``},
		{`include   required(  file(  "demo.conf" ))`, `include required(file("demo.conf"))`, ``},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			p := newTestParser(tc.input)
			inc, err := p.parseInclude()
			if err != nil {
				t.Fatalf("parseInclude(%s) error: %v", tc.input, err)
			}
			if inc.String() != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, inc.String())
			}
			if rest := remainingInput(p); rest != tc.rest {
				t.Fatalf("expected rest %q, got %q", tc.rest, rest)
			}
		})
	}
}

func TestParseIncludeInvalid(t *testing.T) {
	cases := []string{
		"includedemo",
		`include required ("demo")`,
		`include required("demo",)`,
		`include required("demo"`,
		`include required1("demo")`,
		`include classpat("demo")`,
		`include classpath(file("demo"))`,
		`include classpath(required("demo"))`,
	}
	for _, input := range cases {
		p := newTestParser(input)
		if _, err := p.parseInclude(); err == nil {
			t.Fatalf("expected error for %q", input)
		}
	}
}
