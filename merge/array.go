package merge

import (
	"hocon-go/raw"
)

type Array struct {
	Values   []Value
	IsMerged bool
}

func (o *Array) Type() string {
	return raw.ArrayType
}

func (o *Array) String() string {
	if o.Values == nil {
		return "[]"
	}
	var elements string
	for i, v := range o.Values {
		if i > 0 {
			elements += ", "
		}
		elements += v.String()
	}
	return "[" + elements + "]"
}

func (o *Array) isMergeValue() {}

func NewArray(values []Value, isMerged bool) *Array {
	return &Array{
		Values:   values,
		IsMerged: isMerged,
	}
}
