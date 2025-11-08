package config

import (
	"fmt"
	"hocon-go/common"
	"hocon-go/merge"
	"hocon-go/raw"
	"math"
	"sort"
)

func buildMergeObject(parent *common.Path, obj *raw.Object) (*merge.Object, error) {
	if obj == nil {
		return merge.NewObject(make(map[string]merge.Value), true), nil
	}
	result := merge.NewObject(make(map[string]merge.Value), false)
	for _, field := range obj.Fields {
		switch f := field.(type) {
		case *raw.KeyValueField:
			parts := f.Key.AsPath()
			if len(parts) == 0 {
				return nil, fmt.Errorf("object key is empty")
			}
			fullPath := appendPathParts(parent, parts)
			val, err := valueFromRaw(fullPath, f.Value)
			if err != nil {
				return nil, err
			}
			fieldObj := objectForPath(parts, val)
			if err := result.Merge(fieldObj, parent); err != nil {
				return nil, err
			}
		case *raw.InclusionField:
			if f.Inclusion.Val == nil {
				continue
			}
			child, err := buildMergeObject(parent, f.Inclusion.Val)
			if err != nil {
				return nil, err
			}
			if err := result.Merge(child, parent); err != nil {
				return nil, err
			}
		case *raw.NewlineCommentField:
			continue
		default:
			return nil, fmt.Errorf("unsupported object field %T", f)
		}
	}
	return result, nil
}

func objectForPath(parts []string, value merge.Value) *merge.Object {
	current := value
	for i := len(parts) - 1; i >= 0; i-- {
		obj := merge.NewObject(map[string]merge.Value{
			parts[i]: current,
		}, false)
		current = obj
	}
	return current.(*merge.Object)
}

func valueFromRaw(path *common.Path, rv raw.Value) (merge.Value, error) {
	switch v := rv.(type) {
	case *raw.Object:
		return buildMergeObject(path, v)
	case *raw.Array:
		values := make([]merge.Value, len(v.Values))
		for i, item := range v.Values {
			itemPath := appendIndex(path, uint(i))
			val, err := valueFromRaw(itemPath, item)
			if err != nil {
				return nil, err
			}
			values[i] = val
		}
		return merge.NewArray(values, false), nil
	case *raw.Boolean:
		return merge.NewBoolean(v.Val), nil
	case *raw.Null:
		return &merge.Null{}, nil
	case *raw.QuotedString:
		return merge.NewString(v.String()), nil
	case *raw.UnquotedString:
		return merge.NewString(v.String()), nil
	case *raw.MultilineString:
		return merge.NewString(v.String()), nil
	case *raw.PathExpressionString:
		return merge.NewString(v.String()), nil
	case *raw.PosInt, *raw.NegInt, *raw.Float:
		if number, ok := v.(raw.Number); ok {
			return merge.NewNumber(number), nil
		}
		return nil, fmt.Errorf("unknown number type %T", v)
	case *raw.Substitution:
		subPath, err := stringsToPath(v.Path.AsPath())
		if err != nil {
			return nil, err
		}
		return merge.NewSubstitution(subPath, v.Optional), nil
	case *raw.Concat:
		values := make([]merge.Value, len(v.Values))
		for i, val := range v.Values {
			itemPath := appendIndex(path, uint(i))
			merged, err := valueFromRaw(itemPath, val)
			if err != nil {
				return nil, err
			}
			values[i] = merged
		}
		spaces := make([]*string, len(v.Spaces))
		for i, space := range v.Spaces {
			if space == nil {
				continue
			}
			s := *space
			spaces[i] = &s
		}
		concat, err := merge.NewConcat(values, spaces)
		if err != nil {
			return nil, err
		}
		return concat, nil
	case *raw.AddAssign:
		val, err := valueFromRaw(path, v.Val)
		if err != nil {
			return nil, err
		}
		return merge.NewAddAssign(val), nil
	default:
		return nil, fmt.Errorf("unsupported raw value %T", rv)
	}
}

func objectToInterface(obj *merge.Object) (map[string]interface{}, error) {
	if obj == nil || len(obj.Values) == 0 {
		return map[string]interface{}{}, nil
	}
	result := make(map[string]interface{}, len(obj.Values))
	keys := make([]string, 0, len(obj.Values))
	for k := range obj.Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		val, err := valueToInterface(obj.Values[key])
		if err != nil {
			return nil, err
		}
		result[key] = val
	}
	return result, nil
}

func valueToInterface(value merge.Value) (interface{}, error) {
	switch v := value.(type) {
	case *merge.Object:
		return objectToInterface(v)
	case *merge.Array:
		result := make([]interface{}, len(v.Values))
		for i, item := range v.Values {
			converted, err := valueToInterface(item)
			if err != nil {
				return nil, err
			}
			result[i] = converted
		}
		return result, nil
	case *merge.String:
		return v.Val, nil
	case *merge.Boolean:
		return v.Val, nil
	case *merge.Number:
		switch num := v.N.(type) {
		case *raw.PosInt:
			if num.Val <= math.MaxInt64 {
				return int64(num.Val), nil
			}
			return num.Val, nil
		case *raw.NegInt:
			return num.Val, nil
		case *raw.Float:
			return num.Val, nil
		default:
			return nil, fmt.Errorf("unknown number representation %T", num)
		}
	case *merge.Null, *merge.None:
		return nil, nil
	default:
		return nil, fmt.Errorf("value of type %T is not resolved", value)
	}
}

func appendPathParts(parent *common.Path, parts []string) *common.Path {
	var path *common.Path
	if parent != nil {
		path = clonePath(parent)
	}
	for _, part := range parts {
		key := common.NewStrKey(part)
		if path == nil {
			path = common.NewPath(key, nil)
		} else {
			path = appendKey(path, key)
		}
	}
	return path
}

func stringsToPath(parts []string) (*common.Path, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("path is empty")
	}
	var path *common.Path
	for _, part := range parts {
		key := common.NewStrKey(part)
		if path == nil {
			path = common.NewPath(key, nil)
		} else {
			path = appendKey(path, key)
		}
	}
	return path, nil
}

func appendIndex(path *common.Path, idx uint) *common.Path {
	key := common.NewIndexKey(idx)
	if path == nil {
		return common.NewPath(key, nil)
	}
	return appendKey(path, key)
}

func clonePath(path *common.Path) *common.Path {
	if path == nil {
		return nil
	}
	return &common.Path{
		First:     cloneKey(path.First),
		Remainder: clonePath(path.Remainder),
	}
}

func appendKey(path *common.Path, key common.Key) *common.Path {
	if path == nil {
		return common.NewPath(key, nil)
	}
	tail := path
	for tail.Remainder != nil {
		tail = tail.Remainder
	}
	tail.Remainder = common.NewPath(key, nil)
	return path
}

func cloneKey(key common.Key) common.Key {
	switch k := key.(type) {
	case *common.StrKey:
		return common.NewStrKey(k.Str)
	case *common.IndexKey:
		return common.NewIndexKey(k.Index)
	default:
		return key
	}
}
