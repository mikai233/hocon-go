package raw

import "fmt"

type ObjectField interface {
	String() string
	isObjectField()
	SetComment(comment Comment)
}

type InclusionField struct {
	Inclusion Inclusion
	Comment   *Comment
}

func (*InclusionField) isObjectField() {}

func (i *InclusionField) SetComment(comment Comment) {
	i.Comment = &comment
}
func (i *InclusionField) String() string {
	if i.Comment != nil {
		return fmt.Sprintf("%s %s", i.Inclusion.String(), i.Comment.String())
	}
	return i.Inclusion.String()
}

type KeyValueField struct {
	Key     String
	Value   Value
	Comment *Comment
}

func (*KeyValueField) isObjectField() {}

func (k *KeyValueField) SetComment(comment Comment) {
	k.Comment = &comment
}
func (k *KeyValueField) String() string {
	if k.Comment != nil {
		return fmt.Sprintf("%s: %s %s", k.Key.String(), k.Value.String(), k.Comment.String())
	}
	return fmt.Sprintf("%s: %s", k.Key.String(), k.Value.String())
}

type NewlineCommentField struct {
	Comment Comment
}

func (*NewlineCommentField) isObjectField() {}
func (n *NewlineCommentField) SetComment(comment Comment) {
	n.Comment = comment
}
func (n *NewlineCommentField) String() string {
	return n.Comment.String()
}

func NewInclusionField(inclusion Inclusion) ObjectField {
	return &InclusionField{Inclusion: inclusion}
}

func NewInclusionFieldWithComment(inclusion Inclusion, comment Comment) ObjectField {
	return &InclusionField{Inclusion: inclusion, Comment: &comment}
}

func NewKeyValueField(key String, value Value) ObjectField {
	return &KeyValueField{Key: key, Value: value}
}

func NewKeyValueFieldWithComment(key String, value Value, comment Comment) ObjectField {
	return &KeyValueField{Key: key, Value: value, Comment: &comment}
}

func NewNewlineCommentField(comment Comment) ObjectField {
	return &NewlineCommentField{Comment: comment}
}
