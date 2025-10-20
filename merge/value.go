package merge

type Value interface {
	Type() string
	String() string
	isMergeValue()
}
