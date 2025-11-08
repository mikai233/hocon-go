package merge

import (
	"fmt"
	"hocon-go/common"
	"hocon-go/raw"
	"log"
)

type Value interface {
	Type() string
	String() string
	isMergeValue()
}

func Replace(path *common.Path, left Value, right Value) (Value, error) {
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
				return NewDelayReplacement([]Value{left, rr}), nil
			default:
				return rr, nil
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
	log.Printf("concatenate: %v <- %v", left, right)

	var val Value

	switch l := left.(type) {
	// --- Object concatenation ---
	case *Object:
		switch r := right.(type) {
		case *None:
			val = l
		case *Object:
			if err := l.Merge(r, path); err != nil {
				return nil, err
			}
			val = l
		case *Null, *Array, *Boolean, *String, *Number, *AddAssign:
			return nil, &common.ConcatenateDifferentType{Path: path.String(), LeftType: left.Type(), RightType: right.Type()}
		case *Substitution, *DelayReplacement:
			val = NewConcatTwo(left, space, right)
		case *Concat:
			r.PushFront(left, space)
			val = r
		default:
			return nil, fmt.Errorf("unsupported right type: %T", right)
		}

	// --- Array concatenation ---
	case *Array:
		if r, ok := right.(*Array); ok {
			l.Values = append(l.Values, r.Values...)
			val = l
		} else {
			return nil, &common.ConcatenateDifferentType{Path: path.String(), LeftType: left.Type(), RightType: right.Type()}
		}

	// --- None ---
	case *None:
		if space != nil {
			switch r := right.(type) {
			case *Null, *Boolean, *String, *Number:
				s := *space + r.String()
				val = NewString(s)
			case *None:
				val = NewString(*space)
			case *Substitution:
				val = NewConcatTwo(left, space, right)
			default:
				val = right
			}
		} else {
			val = right
		}

	// --- Primitive (Null, Bool, String, Number) ---
	case *Null, *Boolean, *String, *Number:
		switch r := right.(type) {
		case *Boolean, *Null, *String, *Number:
			s := l.String()
			if space != nil {
				s += *space
			}
			s += r.String()
			val = NewString(s)
		case *None:
			s := l.String()
			if space != nil {
				s += *space
			}
			val = NewString(s)
		case *Substitution:
			val = NewConcatTwo(left, space, right)
		default:
			return nil, &common.ConcatenateDifferentType{Path: path.String(), LeftType: left.Type(), RightType: right.Type()}
		}

	// --- Substitution or DelayReplacement ---
	case *Substitution, *DelayReplacement:
		val = NewConcatTwo(left, space, right)

	// --- Concat ---
	case *Concat:
		l.PushBack(space, right)
		val = l

	// --- AddAssign (invalid) ---
	case *AddAssign:
		return nil, &common.ConcatenateDifferentType{Path: path.String(), LeftType: left.Type(), RightType: right.Type()}

	default:
		return nil, fmt.Errorf("unknown left type: %T", left)
	}

	log.Printf("concatenate result: %v = %v", path, val)
	return val, nil
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

func CloneValue(value Value) Value {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case *Object:
		copied := make(map[string]Value, len(v.Values))
		for k, child := range v.Values {
			copied[k] = CloneValue(child)
		}
		return &Object{
			Values:   copied,
			IsMerged: v.IsMerged,
		}
	case *Array:
		values := make([]Value, len(v.Values))
		for i, child := range v.Values {
			values[i] = CloneValue(child)
		}
		return &Array{
			Values:   values,
			IsMerged: v.IsMerged,
		}
	case *Boolean:
		return &Boolean{Val: v.Val}
	case *String:
		return &String{Val: v.Val}
	case *Number:
		return &Number{N: cloneNumber(v.N)}
	case *Null:
		return &Null{}
	case *None:
		return &None{}
	case *Substitution:
		return &Substitution{
			Path:     clonePath(v.Path),
			Optional: v.Optional,
		}
	case *Concat:
		values := make([]Value, len(v.values))
		for i, child := range v.values {
			values[i] = CloneValue(child)
		}
		spaces := make([]*string, len(v.spaces))
		for i, space := range v.spaces {
			if space != nil {
				s := *space
				spaces[i] = &s
			}
		}
		newConcat, err := NewConcat(values, spaces)
		if err != nil {
			panic(err)
		}
		return newConcat
	case *AddAssign:
		return &AddAssign{Val: CloneValue(v.Val)}
	case *DelayReplacement:
		values := make([]Value, len(v.Values))
		for i, child := range v.Values {
			values[i] = CloneValue(child)
		}
		return NewDelayReplacement(values)
	default:
		return value
	}
}

func cloneNumber(number raw.Number) raw.Number {
	switch n := number.(type) {
	case *raw.PosInt:
		return &raw.PosInt{Val: n.Val}
	case *raw.NegInt:
		return &raw.NegInt{Val: n.Val}
	case *raw.Float:
		return &raw.Float{Val: n.Val}
	default:
		return number
	}
}
