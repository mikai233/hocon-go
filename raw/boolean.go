package raw

import "strconv"

type Boolean struct {
	Val bool
}

func (*Boolean) Type() string {
	return BooleanType
}

func (b *Boolean) String() string {
	return strconv.FormatBool(b.Val)
}

func (b *Boolean) isRawValue() {
}

func NewBoolean(val bool) *Boolean {
	return &Boolean{Val: val}
}
