package merge

type Null struct{}

func (*Null) Type() string {
	return "null"
}

func (*Null) String() string {
	return "null"
}

func (*Null) isMergeValue() {}

var NULL = Null{}
