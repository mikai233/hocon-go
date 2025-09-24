package parser

import "hocon-go/raw"

type Value struct {
	Values   []raw.Value
	Spaces   []*string
	PreSpace *string
}

func NewValue() *Value {
	return &Value{
		Values:   []raw.Value{},
		Spaces:   []*string{},
		PreSpace: nil,
	}
}

func (v *Value) PushValue(value raw.Value) {
	if len(v.Values) > 0 {
		v.Spaces = append(v.Spaces, v.PreSpace)
		v.PreSpace = nil
	}
	v.Values = append(v.Values, value)
}

// ---------------- Separator ----------------

type Separator int

const (
	Assign Separator = iota
	AddAssign
)

// ---------------- Entry ----------------

type Entry struct {
	Key       *raw.String
	Separator *Separator
	Value     *Value
}

func NewEntry() *Entry {
	return &Entry{}
}

// ---------------- Frame ----------------

type Frame interface {
	TypeName() string
	ExpectValue() bool
}

type FrameObject struct {
	Entries   []raw.ObjectField
	NextEntry *Entry
}

func (f *FrameObject) TypeName() string {
	return "Frame::Object"
}

func (f *FrameObject) ExpectValue() bool {
	if f.NextEntry != nil {
		return f.NextEntry.Key != nil && f.NextEntry.Separator != nil
	}
	return false
}

type FrameArray struct {
	Elements    []raw.Value
	NextElement *Value
}

func (f *FrameArray) TypeName() string {
	return "Frame::Array"
}

func (f *FrameArray) ExpectValue() bool {
	return true
}
