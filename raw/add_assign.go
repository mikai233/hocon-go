package raw

type AddAssign struct {
	Val Value
}

func (AddAssign) Type() string {
	return RAW_ADD_ASSIGN_TYPE
}

func (AddAssign) isRawValue() {
}

func (a AddAssign) String() string {
	return a.Val.String()
}

func NewAddAssign(v Value) AddAssign {
	return AddAssign{Val: v}
}
