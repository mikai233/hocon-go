package merge

import (
	"fmt"
	"hocon-go/common"
	"hocon-go/raw"
	"os"
	"strings"
)

const maxSubstitutionDepth = 32

type Object struct {
	Values   map[string]Value
	IsMerged bool
}

func (o *Object) Type() string {
	return raw.ObjectType
}

func (o *Object) String() string {
	if o.Values == nil {
		return "{}"
	}
	parts := make([]string, 0, len(o.Values))
	for k, v := range o.Values {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v.String()))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func (o *Object) isMergeValue() {}

func NewObject(values map[string]Value, isMerged bool) *Object {
	return &Object{
		Values:   values,
		IsMerged: isMerged,
	}
}

// Merge merges another Object into the current one.
// It mirrors the Rust logic of recursively merging nested objects and
// replacing non-object values via Replace().
func (o *Object) Merge(other *Object, parent *common.Path) error {
	// Determine whether both sides were merged
	bothMerged := o.IsMerged && other.IsMerged

	// Iterate over keys in right-hand object
	for k, vRight := range other.Values {
		var subPath *common.Path
		if parent == nil {
			subPath = common.NewPath(common.NewStrKey(k), nil)
		} else {
			parentCopy := *parent
			parentCopy.Join(common.NewStrKey(k))
			subPath = &parentCopy
		}

		if vLeft, ok := o.Values[k]; ok {
			// both sides have the same key
			switch l := vLeft.(type) {
			case *Object:
				if r, ok := vRight.(*Object); ok {
					// both are objects → recursive merge
					if err := l.Merge(r, subPath); err != nil {
						return err
					}
					continue
				}
				// else fallthrough to replacement below
			}

			// general replace logic
			replaced, err := Replace(subPath, vLeft, vRight)
			if err != nil {
				return fmt.Errorf("merge replace failed at key %q: %w", k, err)
			}

			// mimic obj.resolve_add_assign() from Rust
			if obj, ok := replaced.(*Object); ok {
				obj.ResolveAddAssign()
			}

			o.Values[k] = replaced

		} else {
			// key only exists in right object
			replaced, err := Replace(subPath, &None{}, vRight)
			if err != nil {
				return fmt.Errorf("merge new key replace failed at key %q: %w", k, err)
			}
			if obj, ok := replaced.(*Object); ok {
				obj.ResolveAddAssign()
			}
			o.Values[k] = replaced
		}
	}

	// Merge-state tracking
	if !bothMerged {
		o.IsMerged = true
	} else {
		o.IsMerged = false
	}

	return nil
}

func (o *Object) ResolveAddAssign() {
	if o.IsMerged {
		return
	}
	for k := range o.Values {
		ResolveAddAssign(k, o.Values)
	}
}

func (o *Object) TryBecomeMerged() {
	if o == nil {
		return
	}
	for _, v := range o.Values {
		if !IsMerged(v) {
			o.IsMerged = false
			return
		}
	}
	o.IsMerged = true
}

func (o *Object) Substitute() error {
	if o == nil {
		return nil
	}
	memo := &Memo{}
	for key, val := range o.Values {
		path := common.NewPath(common.NewStrKey(key), nil)
		resolved, err := o.substituteValue(path, val, memo)
		if err != nil {
			return err
		}
		o.Values[key] = resolved
	}
	o.TryBecomeMerged()
	return nil
}

func (o *Object) substituteValue(path *common.Path, value Value, memo *Memo) (Value, error) {
	if memo == nil {
		memo = &Memo{}
	}
	memo.SubstitutionCounter++
	if memo.SubstitutionCounter > maxSubstitutionDepth {
		return nil, &common.SubstitutionDepthExceeded{MaxDepth: maxSubstitutionDepth}
	}
	defer func() { memo.SubstitutionCounter-- }()

	switch v := value.(type) {
	case *Object:
		for key, child := range v.Values {
			subPath := appendPath(path, common.NewStrKey(key))
			resolved, err := o.substituteValue(subPath, child, memo)
			if err != nil {
				return nil, err
			}
			v.Values[key] = resolved
		}
		v.TryBecomeMerged()
		return v, nil
	case *Array:
		return o.handleArray(path, v, memo)
	case *Boolean, *Null, *None, *String, *Number:
		return v, nil
	case *Substitution:
		return o.handleSubstitution(path, v, memo)
	case *Concat:
		return o.handleConcat(path, v, memo)
	case *AddAssign:
		return o.handleAddAssign(path, v, memo)
	case *DelayReplacement:
		return o.handleDelayReplacement(path, v, memo)
	default:
		return value, nil
	}
}

func (o *Object) handleArray(path *common.Path, array *Array, memo *Memo) (Value, error) {
	allMerged := true
	for idx, element := range array.Values {
		subPath := appendPath(path, common.NewIndexKey(uint(idx)))
		resolved, err := o.substituteValue(subPath, element, memo)
		if err != nil {
			return nil, err
		}
		array.Values[idx] = resolved
		if !IsMerged(resolved) {
			allMerged = false
		}
	}
	array.IsMerged = allMerged
	return array, nil
}

func (o *Object) handleAddAssign(path *common.Path, add *AddAssign, memo *Memo) (Value, error) {
	resolved, err := o.substituteValue(path, add.Val, memo)
	if err != nil {
		return nil, err
	}
	add.Val = resolved
	return add, nil
}

func (o *Object) handleSubstitution(path *common.Path, substitution *Substitution, memo *Memo) (Value, error) {
	if err := pushTrackerPath(memo, path); err != nil {
		return nil, err
	}
	defer popTrackerPath(memo)

	if target, ok := o.getValueByPath(substitution.Path); ok {
		if pathsEqual(substitution.Path, path) {
			if _, isSub := target.(*Substitution); isSub {
				if substitution.Optional {
					return &None{}, nil
				}
				return nil, &common.SubstitutionCycle{
					Current:   substitution.String(),
					Backtrace: []string{substitution.String()},
				}
			}
		}
		clone := CloneValue(target)
		resolved, err := o.substituteValue(clonePath(substitution.Path), clone, memo)
		if err != nil {
			return nil, err
		}
		return resolved, nil
	}

	if envVal, ok := os.LookupEnv(substitution.FullPath()); ok {
		return NewString(envVal), nil
	}
	if substitution.Optional {
		return &None{}, nil
	}
	return nil, &common.SubstitutionNotFound{Path: substitution.FullPath()}
}

func (o *Object) handleConcat(path *common.Path, concat *Concat, memo *Memo) (Value, error) {
	for idx, element := range concat.values {
		subPath := appendPath(path, common.NewIndexKey(uint(idx)))
		resolved, err := o.substituteValue(subPath, element, memo)
		if err != nil {
			return nil, err
		}
		concat.values[idx] = resolved
	}
	return concat.TryResolve(path)
}

func (o *Object) handleDelayReplacement(path *common.Path, delay *DelayReplacement, memo *Memo) (Value, error) {
	for idx, element := range delay.Values {
		subPath := appendPath(path, common.NewIndexKey(uint(idx)))
		resolved, err := o.substituteValue(subPath, element, memo)
		if err != nil {
			return nil, err
		}
		delay.Values[idx] = resolved
	}
	if len(delay.Values) == 0 {
		return &None{}, nil
	}
	result := delay.Values[len(delay.Values)-1]
	for i := len(delay.Values) - 2; i >= 0; i-- {
		var err error
		result, err = Replace(path, delay.Values[i], result)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (o *Object) getValueByPath(path *common.Path) (Value, bool) {
	if path == nil {
		return nil, false
	}
	key, ok := path.First.(*common.StrKey)
	if !ok {
		return nil, false
	}
	val, ok := o.Values[key.Str]
	if !ok {
		return nil, false
	}
	if path.Remainder == nil {
		return val, true
	}
	return getValueFromPath(val, path.Remainder)
}

func getValueFromPath(val Value, path *common.Path) (Value, bool) {
	if path == nil {
		return val, true
	}
	switch v := val.(type) {
	case *Object:
		key, ok := path.First.(*common.StrKey)
		if !ok {
			return nil, false
		}
		child, ok := v.Values[key.Str]
		if !ok {
			return nil, false
		}
		return getValueFromPath(child, path.Remainder)
	case *Array:
		indexKey, ok := path.First.(*common.IndexKey)
		if !ok {
			return nil, false
		}
		idx := int(indexKey.Index)
		if idx < 0 || idx >= len(v.Values) {
			return nil, false
		}
		return getValueFromPath(v.Values[idx], path.Remainder)
	default:
		if path.Remainder == nil {
			return val, true
		}
		return nil, false
	}
}

func pushTrackerPath(memo *Memo, path *common.Path) error {
	if memo == nil || path == nil {
		return nil
	}
	pathStr := path.String()
	for idx := len(memo.Tracker) - 1; idx >= 0; idx-- {
		if memo.Tracker[idx] == pathStr {
			backtrace := append([]string{}, memo.Tracker[idx:]...)
			return &common.SubstitutionCycle{
				Current:   pathStr,
				Backtrace: backtrace,
			}
		}
	}
	memo.Tracker = append(memo.Tracker, pathStr)
	return nil
}

func popTrackerPath(memo *Memo) {
	if memo == nil {
		return
	}
	if len(memo.Tracker) == 0 {
		return
	}
	memo.Tracker = memo.Tracker[:len(memo.Tracker)-1]
}
