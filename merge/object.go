package merge

import (
	"fmt"
	"hocon-go/raw"
	"strings"
)

type Object struct {
	Values   map[string]Value
	IsMerged bool
}

func (o *Object) Type() string {
	return raw.ObjectType
}

func (o *Object) String() string {
	if o.Values == nil {
		return "{}"
	}
	parts := make([]string, 0, len(o.Values))
	for k, v := range o.Values {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v.String()))
	}
	return strings.Join(parts, ",")
}

func (o *Object) isMergeValue() {}
