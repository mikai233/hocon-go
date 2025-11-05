package raw

import (
	"errors"
	"strconv"
	"strings"
)

type Number interface {
	isNumber()
}

type PosInt struct {
	Val uint64
}

func (*PosInt) isNumber() {}

func (p *PosInt) Type() string {
	return NumberType
}

func (p *PosInt) String() string {
	return strconv.FormatUint(p.Val, 10)
}

func (p *PosInt) isRawValue() {
}

type NegInt struct {
	Val int64
}

func (*NegInt) isNumber() {}

func (n *NegInt) Type() string {
	return NumberType
}

func (n *NegInt) String() string {
	return strconv.FormatInt(n.Val, 10)
}

func (n *NegInt) isRawValue() {
}

type Float struct {
	Val float64
}

func (*Float) isNumber() {}

func (f *Float) Type() string {
	return NumberType
}

func (f *Float) String() string {
	return strconv.FormatFloat(f.Val, 'f', -1, 64)
}

func (f *Float) isRawValue() {
}

func NewFloat(val float64) *Float {
	return &Float{Val: val}
}

func NewPosInt(val uint64) *PosInt {
	return &PosInt{Val: val}
}

func NewNegInt(val int64) *NegInt {
	return &NegInt{Val: val}
}

// ParseNumber turns a textual HOCON number into one of the Number implementations.
func ParseNumber(text string) (Number, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, errors.New("number literal is empty")
	}

	// Float literals (anything with a decimal point or exponent).
	if strings.ContainsAny(trimmed, ".eE") {
		f, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, err
		}
		return &Float{Val: f}, nil
	}

	// Negative integers.
	if trimmed[0] == '-' {
		i, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return nil, err
		}
		return &NegInt{Val: i}, nil
	}

	// Positive integers (also accepts a leading '+', strip it first).
	if trimmed[0] == '+' {
		trimmed = trimmed[1:]
	}
	u, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil {
		return nil, err
	}
	return &PosInt{Val: u}, nil
}
