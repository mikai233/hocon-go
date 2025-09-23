package raw

type Null struct {
}

func (Null) Type() string {
	return RAW_NULL_TYPE
}

func (Null) String() string {
	return "null"
}

func (Null) isRawValue() {
}

func NewNull() Null {
	return Null{}
}
