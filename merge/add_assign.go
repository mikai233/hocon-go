package merge

import (
	"hocon-go/common"
	"hocon-go/raw"
)

type AddAssign struct {
	Val Value
}

func (*AddAssign) Type() string {
	return raw.AddAssignType
}

func (*AddAssign) isMergeValue() {
}

func (a *AddAssign) String() string {
	return a.Val.String()
}

func (a *AddAssign) TryResolve(path *common.Path) (Value, error) {
	if IsMerged(a.Val) {
		return a.Val, nil
	} else if concat, ok := a.Val.(*Concat); ok {
		return concat.TryResolve(path)
	} else {
		return a.Val, nil
	}
}

func NewAddAssign(v Value) *AddAssign {
	return &AddAssign{Val: v}
}
