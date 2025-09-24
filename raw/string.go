package raw

import (
	"fmt"
	"strings"
)

type PathExpression struct {
	Paths []String
}

func NewPathExpression(paths []String) PathExpression {
	return PathExpression{Paths: paths}
}

func (p *PathExpression) String() string {
	parts := make([]string, len(p.Paths))
	for i, s := range p.Paths {
		parts[i] = s.String()
	}
	return strings.Join(parts, ".")
}

func (p *PathExpression) Debug() string {
	parts := make([]string, len(p.Paths))
	for i, s := range p.Paths {
		parts[i] = fmt.Sprintf("%q", s.String())
	}
	return strings.Join(parts, ".")
}

type String interface {
	String() string
	Type() string
	AsPath() []string
	isRawString() // marker method
}

type QuotedString struct{ Value string }

func (*QuotedString) isRawString()       {}
func (s *QuotedString) String() string   { return s.Value }
func (*QuotedString) Type() string       { return QuotedStringType }
func (s *QuotedString) AsPath() []string { return []string{s.Value} }

type UnquotedString struct{ Value string }

func (*UnquotedString) isRawString()       {}
func (s *UnquotedString) String() string   { return s.Value }
func (s *UnquotedString) Type() string     { return UnquotedStringType }
func (s *UnquotedString) AsPath() []string { return []string{s.Value} }

type MultilineString struct{ Value string }

func (*MultilineString) isRawString()       {}
func (s *MultilineString) String() string   { return s.Value }
func (s *MultilineString) Type() string     { return MultilineStringType }
func (s *MultilineString) AsPath() []string { return []string{s.Value} }

type PathExpressionString struct{ Value PathExpression }

func (*PathExpressionString) isRawString()     {}
func (s *PathExpressionString) String() string { return s.Value.String() }
func (s *PathExpressionString) Type() string   { return ConcatStringType }
func (s *PathExpressionString) AsPath() []string {
	var result []string
	for _, p := range s.Value.Paths {
		result = append(result, p.AsPath()...)
	}
	return result
}

func NewQuotedString(val string) *QuotedString {
	return &QuotedString{Value: val}
}

func NewUnquotedString(val string) *UnquotedString {
	return &UnquotedString{Value: val}
}

func NewMultilineString(val string) *MultilineString {
	return &MultilineString{Value: val}
}

func NewPathExpressionString(paths []String) *PathExpressionString {
	return &PathExpressionString{Value: NewPathExpression(paths)}
}
