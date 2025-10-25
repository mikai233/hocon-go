package merge

import (
	"fmt"
	"hocon-go/common"
)

type Value interface {
	Type() string
	String() string
	isMergeValue()
}

func Replace(path *common.Path, left Value, right Value) (Value, error) {
	// trace
	// fmt.Printf("replace: `%s`: `%s` <- `%s`\n", *path, left.String(), right.String())

	switch l := left.(type) {

	// LEFT = OBJECT
	case *Object:
		switch r := right.(type) {
		case *Object:
			if err := l.Merge(r, path); err != nil {
				return nil, err
			}
			return l, nil
		case *Array, *Boolean, *Null, *None, *String, *Number:
			return right, nil
		case *Substitution:
			// wrap as DelayReplacement([left, right])
			return NewDelayReplacement([]Value{left, right}), nil
		case *Concat:
			resolved, err := r.TryResolve(path)
			if err != nil {
				return nil, err
			}
			switch rr := resolved.(type) {
			case *Object:
				if err := l.Merge(rr, path); err != nil {
					return nil, err
				}
				return l, nil
			case *Concat:
				rr.PushFront(left, nil)
				return rr, nil
			default:
				return rr, nil
			}
		case *AddAssign:
			return nil, &common.ConcatenateDifferentType{Path: path.String(), LeftType: "object", RightType: r.Type()}
		case *DelayReplacement:
			r.PushFront(left)
			return r, nil
		default:
			return right, nil
		}

	// LEFT = ARRAY
	case *Array:
		switch r := right.(type) {
		case *Substitution, *DelayReplacement:
			return NewDelayReplacement([]Value{left, right}), nil
		case *Concat:
			resolved, err := r.TryResolve(path)
			if err != nil {
				return nil, err
			}
			switch rr := resolved.(type) {
			case *Array:
				return Concatenate(path, l, nil, rr)
			case *Concat:
				return NewDelayReplacement([]Value{left, right}), nil
			default:
				return right, nil
			}
		case *AddAssign:
			inner := r.Val
			isUnmerged := !IsMerged(inner)
			l.Values = append(l.Values, inner)
			if isUnmerged {
				l.IsMerged = false
			}
			return l, nil
		default:
			return right, nil
		}

	// LEFT = NULL
	case *Null:
		if _, ok := right.(*AddAssign); ok {
			return nil, &common.ConcatenateDifferentType{Path: path.String(), LeftType: l.Type(), RightType: right.Type()}
		}
		return right, nil

	// LEFT = NONE
	case *None:
		if add, ok := right.(*AddAssign); ok {
			val, err := add.TryResolve(path)
			if err != nil {
				return nil, err
			}
			arr := NewArray([]Value{val}, IsMerged(val))
			return arr, nil
		}
		return right, nil

	// LEFT = PRIMITIVES (boolean, string, number)
	case *Boolean, *String, *Number:
		switch r := right.(type) {
		case *Substitution:
			return NewDelayReplacement([]Value{left, right}), nil
		case *Concat:
			resolved, err := r.TryResolve(path)
			if err != nil {
				return nil, err
			}
			switch rr := resolved.(type) {
			case *Concat:
				return NewDelayReplacement([]Value{left, rr}), nil
			case *AddAssign:
				return nil, &common.ConcatenateDifferentType{Path: path.String(), LeftType: left.Type(), RightType: rr.Type()}
			default:
				return rr, nil
			}
		case *AddAssign:
			return nil, &common.ConcatenateDifferentType{Path: path.String(), LeftType: left.Type(), RightType: right.Type()}
		default:
			return right, nil
		}

	// LEFT = ADDASSIGN (should not happen)
	case *AddAssign:
		return nil, &common.ConcatenateDifferentType{Path: path.String(), LeftType: left.Type(), RightType: right.Type()}

	// LEFT = SUBSTITUTION | CONCAT | DELAY_REPLACEMENT
	case *Substitution, *Concat, *DelayReplacement:
		return NewDelayReplacement([]Value{left, right}), nil

	default:
		return nil, fmt.Errorf("unknown left value type: %T", left)
	}
}

func ResolveAddAssign(k string, values map[string]Value) {
	v := values[k]
	switch val := v.(type) {
	case *Object:
		val.ResolveAddAssign()
	case *AddAssign:
		array := NewArray([]Value{val.Val}, false)
		values[k] = array
	}
}

func Concatenate(path *common.Path, left Value, space *string, right Value) (Value, error) {
	// implement concatenation logic according to HOCON semantics
	panic("unimplement")
}

func IsMerged(value Value) bool {
	switch val := value.(type) {
	case *Object:
		return val.IsMerged
	case *Array:
		return val.IsMerged
	case *Boolean:
	case *String:
	case *Number:
	case *Null:
	case *None:
		return true
	case *Substitution:
	case *Concat:
	case *AddAssign:
	case *DelayReplacement:
		return false
	}
	return false
}
