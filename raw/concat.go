package raw

import (
	"fmt"
	"strings"
)

type Concat struct {
	Values []Value
	Spaces []*string
}

func (*Concat) Type() string {
	return ConcatType
}

func (*Concat) isRawValue() {
}

func (c *Concat) String() string {
	parts := make([]string, len(c.Values))
	for i, v := range c.Values {
		parts[i] = v.String()
	}
	return strings.Join(parts, " ")
}

func NewConcat(values []Value, spaces []*string) (*Concat, error) {
	if len(values) != len(spaces)+1 {
		return nil, fmt.Errorf("invalid concat: %d values, %d spaces", len(values), len(spaces))
	}

	for _, v := range values {
		switch v.(type) {
		case *Concat, *AddAssign:
			return nil, fmt.Errorf("invalid value type for concat: %s", v.Type())
		}
	}

	return &Concat{
		Values: values,
		Spaces: spaces,
	}, nil
}
