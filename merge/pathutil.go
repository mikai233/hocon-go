package merge

import "hocon-go/common"

func clonePath(p *common.Path) *common.Path {
	if p == nil {
		return nil
	}
	return &common.Path{
		First:     cloneKey(p.First),
		Remainder: clonePath(p.Remainder),
	}
}

func appendPath(path *common.Path, key common.Key) *common.Path {
	if path == nil {
		return common.NewPath(key, nil)
	}
	clone := clonePath(path)
	tail := clone
	for tail.Remainder != nil {
		tail = tail.Remainder
	}
	tail.Remainder = common.NewPath(key, nil)
	return clone
}

func pathsEqual(a, b *common.Path) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if !keysEqual(a.First, b.First) {
		return false
	}
	return pathsEqual(a.Remainder, b.Remainder)
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

func keysEqual(a, b common.Key) bool {
	switch ka := a.(type) {
	case *common.StrKey:
		kb, ok := b.(*common.StrKey)
		return ok && ka.Str == kb.Str
	case *common.IndexKey:
		kb, ok := b.(*common.IndexKey)
		return ok && ka.Index == kb.Index
	default:
		return a == b
	}
}
