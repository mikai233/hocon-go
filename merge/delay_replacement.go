package merge

// DelayReplacement is a container for values that cannot be immediately merged during a replacement operation.
// When merging HOCON values, a substitution expression (${...}) might be encountered. Since the final value
// of the substitution is unknown until the entire configuration is parsed, these pending values are stored here.
type DelayReplacement struct {
	Values []Value
}

// NewDelayReplacement creates a new DelayReplacement from a slice of Values.
// Nested DelayReplacement entries are flattened to mirror the Rust semantics.
func NewDelayReplacement(values []Value) *DelayReplacement {
	return &DelayReplacement{Values: flattenDelayReplacementValues(values)}
}

// FromIter creates a DelayReplacement from any iterable (slice) of Values.
func FromIter(values []Value) *DelayReplacement {
	ptrs := make([]Value, len(values))
	copy(ptrs, values)
	return NewDelayReplacement(ptrs)
}

// IntoInner returns the underlying slice of Values.
func (d *DelayReplacement) IntoInner() []Value {
	return d.Values
}

// Flatten recursively flattens nested DelayReplacement values into a single DelayReplacement.
func (d *DelayReplacement) Flatten() *DelayReplacement {
	var flat []Value
	for _, v := range d.Values {
		switch val := v.(type) {
		case *DelayReplacement:
			flat = append(flat, val.Flatten().Values...)
		default:
			flat = append(flat, v)
		}
	}
	return &DelayReplacement{Values: flat}
}

// String implements fmt.Stringer for debugging or display.
func (d *DelayReplacement) String() string {
	s := "DelayReplacement("
	for i, v := range d.Values {
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

func (d *DelayReplacement) PushFront(val Value) {
	d.Values = append([]Value{val}, d.Values...)
}

func flattenDelayReplacementValues(values []Value) []Value {
	if len(values) == 0 {
		return nil
	}
	flat := make([]Value, 0, len(values))
	for _, v := range values {
		if dr, ok := v.(*DelayReplacement); ok {
			flat = append(flat, dr.Flatten().Values...)
		} else {
			flat = append(flat, v)
		}
	}
	return flat
}
