package merge

type None struct{}

func (o *None) Type() string {
	return "none"
}

func (o *None) String() string {
	return "none"
}

func (o *None) isMergeValue() {}

var NONE = None{}
