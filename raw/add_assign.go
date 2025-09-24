package raw

type AddAssign struct {
	Val Value
}

func (AddAssign) Type() string {
	return AddAssignType
}

func (AddAssign) isRawValue() {
}

func (a AddAssign) String() string {
	return a.Val.String()
}

func NewAddAssign(v Value) *AddAssign {
	return &AddAssign{Val: v}
}
