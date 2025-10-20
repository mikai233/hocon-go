package merge

// DelayReplacement is a container for values that cannot be immediately merged during a replacement operation.
// When merging HOCON values, a substitution expression (${...}) might be encountered. Since the final value
// of the substitution is unknown until the entire configuration is parsed, these pending values are stored here.
type DelayReplacement struct {
	values []Value
}

// NewDelayReplacement creates a new DelayReplacement from a slice of Values.
func NewDelayReplacement(values []Value) *DelayReplacement {
	return &DelayReplacement{values: values}
}

// FromIter creates a DelayReplacement from any iterable (slice) of Values.
func FromIter(values []Value) *DelayReplacement {
	ptrs := make([]Value, len(values))
	copy(ptrs, values)
	return &DelayReplacement{values: ptrs}
}

// IntoInner returns the underlying slice of Values.
func (d *DelayReplacement) IntoInner() []Value {
	return d.values
}

// Flatten recursively flattens nested DelayReplacement values into a single DelayReplacement.
func (d *DelayReplacement) Flatten() *DelayReplacement {
	var flat []Value
	for _, v := range d.values {
		switch val := v.(type) {
		case *DelayReplacement:
			flat = append(flat, val.Flatten().values...)
		default:
			flat = append(flat, v)
		}
	}
	return &DelayReplacement{values: flat}
}

// String implements fmt.Stringer for debugging or display.
func (d *DelayReplacement) String() string {
	s := "DelayReplacement("
	for i, v := range d.values {
		if i > 0 {
			s += ", "
		}
		s += v.String()
	}
	s += ")"
	return s
}

func (d *DelayReplacement) Type() string {
	return "delay_replacement"
}

func (d *DelayReplacement) isMergeValue() {}
