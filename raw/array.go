package raw

import "strings"

type Array struct {
	Values []Value
}

func (*Array) Type() string {
	return ArrayType
}

func (*Array) isRawValue() {
}

func (a *Array) String() string {
	parts := make([]string, len(a.Values))
	for i, v := range a.Values {
		parts[i] = v.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func NewRawArray(values []Value) *Array {
	return &Array{Values: values}
}
