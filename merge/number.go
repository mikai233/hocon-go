package merge

import "hocon-go/raw"

type Number struct {
	N raw.Number
}

func (*Number) Type() string {
	return raw.NumberType
}

func (n *Number) String() string {
	if n == nil || n.N == nil {
		return "0"
	}
	if s, ok := n.N.(interface{ String() string }); ok {
		return s.String()
	}
	return "number"
}

func (*Number) isMergeValue() {}

func NewNumber(n raw.Number) *Number {
	return &Number{
		N: n,
	}
}
