package merge

import (
	"fmt"
	"hocon-go/common"
	"hocon-go/raw"
)

// Concat represents a concatenation of evaluated HOCON values during the merge phase.
// Similar to Rust's Concat, but uses Go slices and pointers instead of VecDeque and RefCell.
type Concat struct {
	values []Value   // slice of value pointers
	spaces []*string // slice of optional spaces; len(values) == len(spaces)+1
}

// NewConcat creates a new Concat instance.
// Returns error if len(values) != len(spaces)+1
func NewConcat(values []Value, spaces []*string) (*Concat, error) {
	if len(values) != len(spaces)+1 {
		return nil, fmt.Errorf("invalid concat: values=%d, spaces=%d", len(values), len(spaces))
	}
	return &Concat{values: values, spaces: spaces}, nil
}

// Two creates a minimal Concat with two values and one optional space.
func NewConcatTwo(left Value, space *string, right Value) *Concat {
	return &Concat{
		values: []Value{left, right},
		spaces: []*string{space},
	}
}

// PushBack appends a new value and its preceding space.
func (c *Concat) PushBack(space *string, val Value) {
	if len(c.values) == 0 {
		if space != nil {
			panic("space must be nil when values are empty")
		}
		c.values = append(c.values, val)
	} else {
		c.values = append(c.values, val)
		c.spaces = append(c.spaces, space)
	}
	if len(c.values) != len(c.spaces)+1 {
		panic("invariant violated: len(values) != len(spaces)+1")
	}
}

// PopBack removes and returns the last value and its preceding space.
func (c *Concat) PopBack() (*string, Value) {
	if len(c.values) == 0 {
		if len(c.spaces) != 0 {
			panic("spaces should be empty if values are empty")
		}
		return nil, nil
	}
	val := c.values[len(c.values)-1]
	c.values = c.values[:len(c.values)-1]

	var space *string
	if len(c.values) == 0 {
		space = nil
	} else {
		space = c.spaces[len(c.spaces)-1]
		c.spaces = c.spaces[:len(c.spaces)-1]
	}
	if len(c.values) != len(c.spaces)+1 {
		panic("invariant violated: len(values) != len(spaces)+1")
	}
	return space, val
}

// PushFront inserts a new value and its following space at the beginning.
func (c *Concat) PushFront(val Value, space *string) {
	if len(c.values) == 0 {
		if space != nil {
			panic("space must be nil when values are empty")
		}
		c.values = append([]Value{val}, c.values...)
	} else {
		c.values = append([]Value{val}, c.values...)
		c.spaces = append([]*string{space}, c.spaces...)
	}
	if len(c.values) != len(c.spaces)+1 {
		panic("invariant violated: len(values) != len(spaces)+1")
	}
}

// Len returns the number of values in the Concat.
func (c *Concat) Len() int {
	return len(c.values)
}

// Values returns all values.
func (c *Concat) Values() []Value {
	return c.values
}

// TryResolve attempts to merge all values into a single Value.
// Returns ValueNone if empty, single value if only one, or concatenated otherwise.
func (c *Concat) TryResolve(path *common.Path) (Value, error) {
	if len(c.values) == 0 {
		return &None{}, nil
	} else if len(c.values) == 1 {
		_, v := c.PopBack()
		return v, nil
	} else {
		firstVal := c.values[0]
		firstSpace := c.spaces[0]
		c.values = c.values[1:]
		c.spaces = c.spaces[1:]

		result := firstVal
		space := firstSpace

		for len(c.values) > 0 {
			second := c.values[0]
			var secondSpace *string
			if len(c.spaces) > 0 {
				secondSpace = c.spaces[0]
			}
			c.values = c.values[1:]
			if len(c.spaces) > 0 {
				c.spaces = c.spaces[1:]
			}
			var err error
			result, err = Concatenate(path, result, space, second)
			if err != nil {
				return nil, err
			}
			space = secondSpace
		}
		return result, nil
	}
}

// String returns a human-readable representation of the Concat.
func (c *Concat) String() string {
	s := "Concat("
	for i, val := range c.values {
		if i > 0 {
			s += ", "
		}
		s += val.String()
	}
	s += ")"
	return s
}

func (c *Concat) Type() string {
	return raw.ConcatType
}

func (o *Concat) isMergeValue() {}
