package raw

import "hocon-go/common"

type Object struct {
	Fields []ObjectField
}

func (o *Object) Type() string {
	return RAW_OBJECT_TYPE
}

func (o *Object) isRawValue() {
}

func (o *Object) String() string {
	result := "{"
	for i, f := range o.Fields {
		if i > 0 {
			result += ", "
		}
		result += f.String()
	}
	result += "}"
	return result
}

func NewObject(fields []ObjectField) Object {
	return Object{Fields: fields}
}

func ObjectFromEntries(entries []KeyValueField) Object {
	fields := make([]ObjectField, 0, len(entries))
	for _, kv := range entries {
		fields = append(fields, NewKeyValueField(kv.Key, kv.Value))
	}
	return NewObject(fields)
}

func (o *Object) RemoveByPath(path *common.Path) *ObjectField {
	var removeIndex *int
	for i := len(o.Fields) - 1; i >= 0; i-- {
		field := o.Fields[i]
		switch f := field.(type) {
		case *InclusionField:
			if f.Inclusion.Val != nil {
				return f.Inclusion.Val.RemoveByPath(path)
			}
		case *KeyValueField:
			k := f.Key.AsPath()
			if path.StartsWith1(k) {
				sub := path.SubPath(len(k))
				if sub == nil {
					removeIndex = &i
					break
				} else if obj, ok := f.Value.(*Object); ok {
					return obj.RemoveByPath(sub)
				}
			}
		case *NewlineCommentField:
			// ignore
		}
	}

	if removeIndex != nil {
		field := o.Fields[*removeIndex]
		o.Fields = append(o.Fields[:*removeIndex], o.Fields[*removeIndex+1:]...)
		return &field
	}
	return nil
}

func (o *Object) RemoveAllByPath(path *common.Path) []ObjectField {
	var results []ObjectField
	var removeIndices []int
	for i := len(o.Fields) - 1; i >= 0; i-- {
		field := o.Fields[i]
		switch f := field.(type) {
		case *InclusionField:
			if f.Inclusion.Val != nil {
				results = append(results, f.Inclusion.Val.RemoveAllByPath(path)...)
			}
		case *KeyValueField:
			k := f.Key.AsPath()
			if path.StartsWith1(k) {
				sub := path.SubPath(len(k))
				if sub == nil {
					removeIndices = append(removeIndices, i)
				} else if obj, ok := f.Value.(*Object); ok {
					results = append(results, obj.RemoveAllByPath(sub)...)
				}
			}
		case *NewlineCommentField:
			// ignore
		}
	}

	for _, idx := range removeIndices {
		results = append(results, o.Fields[idx])
		o.Fields = append(o.Fields[:idx], o.Fields[idx+1:]...)
	}

	return results
}

func (o *Object) GetByPath(path *common.Path) Value {
	for i := len(o.Fields) - 1; i >= 0; i-- {
		field := o.Fields[i]
		switch f := field.(type) {
		case *InclusionField:
			if f.Inclusion.Val != nil {
				return f.Inclusion.Val.GetByPath(path)
			}
		case *KeyValueField:
			k := f.Key.AsPath()
			if path.StartsWith1(k) {
				sub := path.SubPath(len(k))
				if sub == nil {
					return f.Value
				} else if obj, ok := f.Value.(*Object); ok {
					return obj.GetByPath(sub)
				}
			}
		case *NewlineCommentField:
			// ignore
		}
	}
	return nil
}

func MergeObject(left, right Object) Object {
	left.Fields = append(left.Fields, right.Fields...)
	return left
}
