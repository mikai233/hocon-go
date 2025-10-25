package merge

import (
	"hocon-go/raw"
	"strconv"
)

type Boolean struct {
	Val bool
}

func (*Boolean) Type() string {
	return raw.BooleanType
}

func (b *Boolean) String() string {
	return strconv.FormatBool(b.Val)
}

func (b *Boolean) isMergeValue() {
}

func NewBoolean(val bool) *Boolean {
	return &Boolean{Val: val}
}
