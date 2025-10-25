package merge

import (
	"fmt"
	"hocon-go/common"
	"hocon-go/raw"
	"strings"
)

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
