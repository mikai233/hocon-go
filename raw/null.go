package raw

type Null struct {
}

func (Null) Type() string {
	return NullType
}

func (Null) String() string {
	return "null"
}

func (Null) isRawValue() {
}

func NewNull() Null {
	return Null{}
}
