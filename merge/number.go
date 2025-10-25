package merge

import "hocon-go/raw"

type Number struct {
	N raw.Number
}

func (*Number) Type() string {
	return raw.NumberType
}

func (*Number) String() string {
	return "number"
}

func (*Number) isMergeValue() {}

func NewNumber(n raw.Number) *Number {
	return &Number{
		N: n,
	}
}
