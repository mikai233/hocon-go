package parser

import (
	"errors"
	"fmt"
	"hocon-go/raw"
	"io"
	"strings"
	"unicode/utf8"
)

type Parser struct {
	reader  *reader
	scratch []byte
	options ConfigOptions
	depth   int
}

func NewParser(data []byte) *Parser {
	return &Parser{
		reader:  newReader(data),
		scratch: make([]byte, 0, 64),
		options: DefaultConfigOptions(),
	}
}

func (p *Parser) WithOptions(opts ConfigOptions) *Parser {
	p.options = opts
	return p
}

func (p *Parser) Parse() (*raw.Object, error) {
	if err := p.dropWhitespaceAndComments(); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	var obj *raw.Object
	ch, err := p.reader.peek()
	switch {
	case err == nil && ch == '{':
		obj, err = p.parseObject(false)
	case err == nil:
		obj, err = p.parseBracesOmittedObject()
	case errors.Is(err, io.EOF):
		return raw.NewObject(nil), nil
	default:
	}
	if err != nil {
		return nil, err
	}
	if err := p.dropWhitespaceAndComments(); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	if _, err := p.reader.peek(); err == nil {
		return nil, fmt.Errorf("unexpected trailing content")
	}
	return obj, nil
}

func (p *Parser) dropWhitespace() error {
	for {
		rn, size, err := p.reader.peekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if !isWhitespace(rn) {
			return nil
		}
		if err := p.reader.discard(size); err != nil {
			return err
		}
	}
}

func (p *Parser) dropHorizontalWhitespace() error {
	for {
		rn, size, err := p.reader.peekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if !isHorizontalWhitespace(rn) {
			return nil
		}
		if err := p.reader.discard(size); err != nil {
			return err
		}
	}
}

func (p *Parser) parseHorizontalWhitespace(into *[]byte) error {
	for {
		rn, size, err := p.reader.peekRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if !isHorizontalWhitespace(rn) {
			return nil
		}
		buf := p.reader.remaining()[:size]
		*into = append(*into, buf...)
		if err := p.reader.discard(size); err != nil {
			return err
		}
	}
}

func (p *Parser) dropWhitespaceAndComments() error {
	for {
		if err := p.dropWhitespace(); err != nil {
			return err
		}
		consumed, err := p.dropComment()
		if err != nil {
			return err
		}
		if !consumed {
			return nil
		}
	}
}

func (p *Parser) dropComment() (bool, error) {
	ch, err := p.reader.peek()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		return false, err
	}
	if ch == '#' {
		if err := p.reader.discard(1); err != nil {
			return false, err
		}
		return true, p.discardUntilNewline()
	}
	if ch == '/' {
		if _, ch2, err := p.reader.peek2(); err == nil && ch2 == '/' {
			if err := p.reader.discard(2); err != nil {
				return false, err
			}
			return true, p.discardUntilNewline()
		}
	}
	return false, nil
}

func (p *Parser) discardUntilNewline() error {
	for {
		ch, err := p.reader.next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if ch == '\n' {
			return nil
		}
		if ch == '\r' {
			// swallow optional LF
			if next, err := p.reader.peek(); err == nil && next == '\n' {
				_ = p.reader.discard(1)
			}
			return nil
		}
	}
}

func (p *Parser) parseObjectField() (raw.ObjectField, error) {
	ch, err := p.reader.peek()
	if err != nil {
		return nil, err
	}
	if ch == 'i' {
		// check include prefix
		if bytes, err := p.reader.peekN(7); err == nil && string(bytes) == "include" {
			inclusion, err := p.parseInclude()
			if err != nil {
				return nil, err
			}
			return raw.NewInclusionField(*inclusion), nil
		}
	}
	key, value, err := p.parseKeyValue()
	if err != nil {
		return nil, err
	}
	return raw.NewKeyValueField(key, value), nil
}

func (p *Parser) parseObject(verifyDelimiter bool) (*raw.Object, error) {
	if verifyDelimiter {
		ch, err := p.reader.peek()
		if err != nil {
			return nil, err
		}
		if ch != '{' {
			return nil, &unexpectedTokenError{Expected: "{", Found: ch}
		}
	}
	if err := p.reader.discard(1); err != nil {
		return nil, err
	}
	obj, err := p.parseBracesOmittedObject()
	if err != nil {
		return nil, err
	}
	ch, err := p.reader.peek()
	if err != nil {
		return nil, err
	}
	if ch != '}' {
		return nil, &unexpectedTokenError{Expected: "}", Found: ch}
	}
	if err := p.reader.discard(1); err != nil {
		return nil, err
	}
	return obj, nil
}

func (p *Parser) parseBracesOmittedObject() (*raw.Object, error) {
	fields := make([]raw.ObjectField, 0)
	for {
		if err := p.dropWhitespaceAndComments(); err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		ch, err := p.reader.peek()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if ch == '}' {
			break
		}
		field, err := p.parseObjectField()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		fields = append(fields, field)
		if err := p.dropWhitespaceAndComments(); err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		stop, err := p.dropCommaSeparator()
		if err != nil {
			return nil, err
		}
		if stop {
			break
		}
	}
	return raw.NewObject(fields), nil
}

func (p *Parser) dropCommaSeparator() (bool, error) {
	ch, err := p.reader.peek()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return true, nil
		}
		return false, err
	}
	if ch == ',' {
		if err := p.reader.discard(1); err != nil {
			return false, err
		}
	}
	return false, nil
}

func (p *Parser) parseArray(verifyDelimiter bool) (*raw.Array, error) {
	if verifyDelimiter {
		ch, err := p.reader.peek()
		if err != nil {
			return nil, err
		}
		if ch != '[' {
			return nil, &unexpectedTokenError{Expected: "[", Found: ch}
		}
	}
	if err := p.reader.discard(1); err != nil {
		return nil, err
	}
	values := make([]raw.Value, 0)
	for {
		if err := p.dropWhitespaceAndComments(); err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		ch, err := p.reader.peek()
		if err != nil {
			return nil, err
		}
		if ch == ']' {
			_ = p.reader.discard(1)
			break
		}
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
		if err := p.dropWhitespaceAndComments(); err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		stop, err := p.dropCommaSeparator()
		if err != nil {
			return nil, err
		}
		if stop {
			break
		}
	}
	return raw.NewRawArray(values), nil
}

func (p *Parser) parseValue() (raw.Value, error) {
	if err := p.dropWhitespace(); err != nil {
		return nil, err
	}
	var values []raw.Value
	var spaces []*string
	var prevSpace *string
	push := func(val raw.Value) {
		if len(values) > 0 {
			spaces = append(spaces, prevSpace)
			prevSpace = nil
		}
		values = append(values, val)
	}

	for {
		ch, err := p.reader.peek()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		switch ch {
		case '[':
			if err := p.increaseDepth(); err != nil {
				return nil, err
			}
			arr, err := p.parseArray(false)
			p.decreaseDepth()
			if err != nil {
				return nil, err
			}
			push(arr)
		case '{':
			if err := p.increaseDepth(); err != nil {
				return nil, err
			}
			obj, err := p.parseObject(false)
			p.decreaseDepth()
			if err != nil {
				return nil, err
			}
			push(obj)
		case '"':
			strVal, err := p.parsePossibleMultilineString()
			if err != nil {
				return nil, err
			}
			push(strVal)
		case '$':
			subst, err := p.parseSubstitution()
			if err != nil {
				return nil, err
			}
			push(subst)
		case '}', ']':
			goto done
		case ',', '#', '\n', '\r':
			if len(values) == 0 {
				return nil, &unexpectedTokenError{Expected: "value", Found: ch}
			}
			goto done
		case '/':
			if _, ch2, err := p.reader.peek2(); err == nil && ch2 == '/' {
				if len(values) == 0 {
					return nil, &unexpectedTokenError{Expected: "value", Found: ch}
				}
				goto done
			}
			fallthrough
		default:
			if p.startsWithHorizontalWhitespace() {
				p.scratch = p.scratch[:0]
				if err := p.parseHorizontalWhitespace(&p.scratch); err != nil {
					return nil, err
				}
				if len(p.scratch) == 0 {
					prevSpace = nil
				} else {
					s := string(p.scratch)
					prevSpace = &s
				}
			} else {
				unquoted, err := p.parseUnquotedString()
				if err != nil {
					return nil, err
				}
				push(unquoted)
			}
		}
	}
done:
	switch len(values) {
	case 0:
		return nil, errors.New("expected value")
	case 1:
		return p.resolveUnquotedString(values[0]), nil
	default:
		concat, err := raw.NewConcat(values, spaces)
		if err != nil {
			return nil, err
		}
		return concat, nil
	}
}

func (p *Parser) startsWithHorizontalWhitespace() bool {
	rn, _, err := p.reader.peekRune()
	if err != nil {
		return false
	}
	return isHorizontalWhitespace(rn)
}

func (p *Parser) increaseDepth() error {
	p.depth++
	if p.depth > p.options.MaxDepth {
		return &depthExceededError{Limit: p.options.MaxDepth}
	}
	return nil
}

func (p *Parser) decreaseDepth() {
	if p.depth > 0 {
		p.depth--
	}
}

func (p *Parser) resolveUnquotedString(val raw.Value) raw.Value {
	switch v := val.(type) {
	case *raw.UnquotedString:
		switch strings.ToLower(v.Value) {
		case "true":
			return raw.NewBoolean(true)
		case "false":
			return raw.NewBoolean(false)
		case "null":
			return &raw.NULL
		default:
			if number, err := raw.ParseNumber(v.Value); err == nil {
				switch n := number.(type) {
				case *raw.PosInt:
					return n
				case *raw.NegInt:
					return n
				case *raw.Float:
					return n
				}
			}
			return v
		}
	default:
		return val
	}
}

func (p *Parser) parseKeyValue() (raw.String, raw.Value, error) {
	if err := p.dropWhitespace(); err != nil {
		return nil, nil, err
	}
	key, err := p.parseKey()
	if err != nil {
		return nil, nil, err
	}
	if err := p.dropWhitespace(); err != nil {
		return nil, nil, err
	}
	addAssign, err := p.dropKVSeparator()
	if err != nil {
		return nil, nil, err
	}
	if err := p.dropWhitespace(); err != nil {
		return nil, nil, err
	}
	val, err := p.parseValue()
	if err != nil {
		return nil, nil, err
	}
	if addAssign {
		val = raw.NewAddAssign(val)
	}
	return key, val, nil
}

func (p *Parser) dropKVSeparator() (bool, error) {
	ch, err := p.reader.peek()
	if err != nil {
		return false, err
	}
	switch ch {
	case ':', '=':
		return false, p.reader.discard(1)
	case '+':
		if _, ch2, err := p.reader.peek2(); err == nil && ch2 == '=' {
			if err := p.reader.discard(2); err != nil {
				return false, err
			}
			return true, nil
		}
		return false, &unexpectedTokenError{Expected: "=", Found: ch}
	default:
		return false, &unexpectedTokenError{Expected: ": or =", Found: ch}
	}
}

func (p *Parser) parseKey() (raw.String, error) {
	if err := p.dropHorizontalWhitespace(); err != nil {
		return nil, err
	}
	return p.parsePathExpression()
}

func (p *Parser) parseInclude() (*raw.Inclusion, error) {
	const token = "include"
	if err := p.reader.discard(len(token)); err != nil {
		return nil, err
	}
	if err := p.dropHorizontalWhitespace(); err != nil {
		return nil, err
	}
	required, err := p.parseRequiredToken()
	if err != nil {
		return nil, err
	}
	if err := p.dropHorizontalWhitespace(); err != nil {
		return nil, err
	}
	location, err := p.parseLocationToken()
	if err != nil {
		return nil, err
	}
	if err := p.dropHorizontalWhitespace(); err != nil {
		return nil, err
	}
	path, err := p.parseQuotedString(true)
	if err != nil {
		return nil, err
	}
	if location != nil {
		if err := p.dropHorizontalWhitespace(); err != nil {
			return nil, err
		}
		if err := p.expectChar(')'); err != nil {
			return nil, err
		}
	}
	if required {
		if err := p.dropHorizontalWhitespace(); err != nil {
			return nil, err
		}
		if err := p.expectChar(')'); err != nil {
			return nil, err
		}
	}
	return raw.NewInclusion(path, required, location, nil), nil
}

func (p *Parser) parseRequiredToken() (bool, error) {
	const token = "required("
	ch, err := p.reader.peek()
	if err != nil {
		return false, err
	}
	if ch != 'r' {
		return false, nil
	}
	if err := p.reader.discard(len(token)); err != nil {
		return false, err
	}
	return true, nil
}

func (p *Parser) parseLocationToken() (*raw.Location, error) {
	ch, err := p.reader.peek()
	if err != nil {
		return nil, err
	}
	var loc raw.Location
	var token string
	switch ch {
	case 'f':
		loc = raw.File
		token = "file("
	case 'c':
		loc = raw.Classpath
		token = "classpath("
	case 'u':
		loc = raw.Url
		token = "url("
	case '"':
		return nil, nil
	default:
		return nil, &unexpectedTokenError{Expected: "file( or classpath( or url(", Found: ch}
	}
	if err := p.reader.discard(len(token)); err != nil {
		return nil, err
	}
	return &loc, nil
}

func (p *Parser) expectChar(ch byte) error {
	next, err := p.reader.peek()
	if err != nil {
		return err
	}
	if next != ch {
		return &unexpectedTokenError{Expected: string(ch), Found: next}
	}
	return p.reader.discard(1)
}

func (p *Parser) parsePathExpression() (raw.String, error) {
	segments := make([]raw.String, 0)
	for {
		if err := p.dropHorizontalWhitespace(); err != nil {
			return nil, err
		}
		ch, err := p.reader.peek()
		if err != nil {
			if len(segments) == 0 {
				return nil, &unexpectedTokenError{Expected: "path", Found: 0}
			}
			break
		}
		var segStr string
		switch ch {
		case '"':
			str, err := p.parsePossibleMultilineText()
			if err != nil {
				return nil, err
			}
			segStr = str
		default:
			str, err := p.parseUnquotedPathSegment()
			if err != nil {
				return nil, err
			}
			segStr = str
		}
		segments = append(segments, raw.NewQuotedString(segStr))
		if err := p.dropHorizontalWhitespace(); err != nil {
			return nil, err
		}
		ch, err = p.reader.peek()
		if err != nil {
			break
		}
		if ch == '.' {
			_ = p.reader.discard(1)
			continue
		}
		if strings.ContainsRune(":=+{", rune(ch)) || ch == '}' {
			break
		}
		if isWhitespace(rune(ch)) {
			break
		}
	}
	if len(segments) == 0 {
		return nil, &unexpectedTokenError{Expected: "path", Found: 0}
	}
	if len(segments) == 1 {
		return segments[0], nil
	}
	return raw.NewPathExpressionString(segments), nil
}

func (p *Parser) parsePossibleMultilineString() (raw.Value, error) {
	if bytes, err := p.reader.peekN(3); err == nil && string(bytes) == "\"\"\"" {
		str, err := p.parseMultilineString(true)
		if err != nil {
			return nil, err
		}
		return raw.NewMultilineString(str), nil
	}
	str, err := p.parseQuotedString(true)
	if err != nil {
		return nil, err
	}
	return raw.NewQuotedString(str), nil
}

func (p *Parser) parsePossibleMultilineText() (string, error) {
	if bytes, err := p.reader.peekN(3); err == nil && string(bytes) == "\"\"\"" {
		return p.parseMultilineString(true)
	}
	return p.parseQuotedString(true)
}

func (p *Parser) parseQuotedString(require bool) (string, error) {
	if require {
		if err := p.expectChar('"'); err != nil {
			return "", err
		}
	} else {
		if err := p.reader.discard(1); err != nil {
			return "", err
		}
	}
	var b strings.Builder
	for {
		ch, err := p.reader.next()
		if err != nil {
			return "", err
		}
		if ch == '"' {
			break
		}
		if ch == '\\' {
			escaped, err := p.parseEscapedChar()
			if err != nil {
				return "", err
			}
			b.WriteRune(escaped)
			continue
		}
		b.WriteByte(ch)
	}
	return b.String(), nil
}

func (p *Parser) parseEscapedChar() (rune, error) {
	ch, err := p.reader.next()
	if err != nil {
		return 0, err
	}
	switch ch {
	case '"':
		return '"', nil
	case '\\':
		return '\\', nil
	case '/':
		return '/', nil
	case 'b':
		return '\b', nil
	case 'f':
		return '\f', nil
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 't':
		return '\t', nil
	case 'u':
		return p.parseUnicodeEscape()
	default:
		return 0, invalidEscapeError{}
	}
}

func (p *Parser) parseUnicodeEscape() (rune, error) {
	readHex := func() (rune, error) {
		var val rune
		for i := 0; i < 4; i++ {
			ch, err := p.reader.next()
			if err != nil {
				return 0, err
			}
			val <<= 4
			switch {
			case ch >= '0' && ch <= '9':
				val |= rune(ch - '0')
			case ch >= 'a' && ch <= 'f':
				val |= rune(ch-'a') + 10
			case ch >= 'A' && ch <= 'F':
				val |= rune(ch-'A') + 10
			default:
				return 0, invalidEscapeError{}
			}
		}
		return val, nil
	}
	code, err := readHex()
	if err != nil {
		return 0, err
	}
	if code >= 0xD800 && code <= 0xDBFF {
		// high surrogate, expect low surrogate
		if ch, err := p.reader.next(); err != nil || ch != '\\' {
			return 0, invalidEscapeError{}
		}
		if ch, err := p.reader.next(); err != nil || ch != 'u' {
			return 0, invalidEscapeError{}
		}
		low, err := readHex()
		if err != nil {
			return 0, err
		}
		if low < 0xDC00 || low > 0xDFFF {
			return 0, invalidEscapeError{}
		}
		code = 0x10000 + ((code-0xD800)<<10 | (low - 0xDC00))
	}
	return rune(code), nil
}

func (p *Parser) parseMultilineString(verify bool) (string, error) {
	if verify {
		bytes, err := p.reader.peekN(3)
		if err != nil {
			return "", err
		}
		if string(bytes) != "\"\"\"" {
			return "", &unexpectedTokenError{Expected: "\"\"\"", Found: bytes[0]}
		}
	}
	if err := p.reader.discard(3); err != nil {
		return "", err
	}
	var b strings.Builder
	for {
		bytes, err := p.reader.peekN(3)
		if err == nil && string(bytes) == "\"\"\"" {
			_ = p.reader.discard(3)
			break
		}
		ch, err := p.reader.next()
		if err != nil {
			return "", err
		}
		b.WriteByte(ch)
	}
	return b.String(), nil
}

func (p *Parser) parseUnquotedString() (raw.Value, error) {
	var b strings.Builder
	for {
		ch, err := p.reader.peek()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if ch == '/' {
			if _, ch2, err := p.reader.peek2(); err == nil && ch2 == '/' {
				break
			}
		}
		if ch < utf8.RuneSelf {
			if ForbiddenTable[ch] || ch == '#' || ch == ',' || ch == ':' || ch == '=' || ch == '+' || ch == '{' || ch == '}' || ch == '[' || ch == ']' {
				break
			}
			if ch <= ' ' {
				break
			}
			b.WriteByte(ch)
			_ = p.reader.discard(1)
			continue
		}
		rn, size, err := p.reader.peekRune()
		if err != nil {
			return nil, err
		}
		if isWhitespace(rn) {
			break
		}
		b.WriteString(string(rn))
		_ = p.reader.discard(size)
	}
	if b.Len() == 0 {
		return nil, errors.New("invalid unquoted string")
	}
	return raw.NewUnquotedString(b.String()), nil
}

func (p *Parser) parseUnquotedPathSegment() (string, error) {
	val, err := p.parseUnquotedString()
	if err != nil {
		return "", err
	}
	return val.(*raw.UnquotedString).Value, nil
}

func (p *Parser) parseSubstitution() (raw.Value, error) {
	if err := p.expectChar('$'); err != nil {
		return nil, err
	}
	if err := p.expectChar('{'); err != nil {
		return nil, err
	}
	optional := false
	if ch, err := p.reader.peek(); err == nil && ch == '?' {
		optional = true
		_ = p.reader.discard(1)
	}
	path, err := p.parsePathExpression()
	if err != nil {
		return nil, err
	}
	if err := p.expectChar('}'); err != nil {
		return nil, err
	}
	return raw.NewSubstitution(path, optional), nil
}
