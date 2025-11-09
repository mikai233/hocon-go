package parser

func newTestParser(input string) *Parser {
	return NewParser([]byte(input))
}

func remainingInput(p *Parser) string {
	if p == nil || p.reader == nil {
		return ""
	}
	rem := p.reader.remaining()
	if rem == nil {
		return ""
	}
	return string(rem)
}
