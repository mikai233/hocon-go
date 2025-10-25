package common

import (
	"errors"
	"strconv"
	"strings"
)

type Key interface {
	IsKey()
	String() string
}

type IndexKey struct {
	Index uint
}

func NewIndexKey(index uint) *IndexKey {
	return &IndexKey{
		Index: index,
	}
}

func (*IndexKey) IsKey() {}

func (i *IndexKey) String() string {
	return strconv.FormatUint(uint64(i.Index), 10)
}

type StrKey struct {
	Str string
}

func NewStrKey(str string) *StrKey {
	return &StrKey{
		Str: str,
	}
}

func (*StrKey) IsKey() {}

func (s *StrKey) String() string {
	return s.Str
}

type Path struct {
	First     Key
	Remainder *Path
}

func NewPath(first Key, remainder *Path) *Path {
	return &Path{First: first, Remainder: remainder}
}

func FromStr(s string) (*Path, error) {
	parts := strings.Split(s, ".")
	if len(parts) == 0 {
		return nil, errors.New("the path is empty")
	}

	dummy := &Path{}
	curr := dummy
	for _, p := range parts {
		curr.Remainder = &Path{First: NewStrKey(p)}
		curr = curr.Remainder
	}
	return dummy.Remainder, nil
}

func (p *Path) Len() int {
	length := 1
	curr := p
	for curr.Remainder != nil {
		length++
		curr = curr.Remainder
	}
	return length
}

func (p *Path) SubPath(n int) *Path {
	curr := p
	for i := 0; i < n && curr != nil; i++ {
		curr = curr.Remainder
	}
	return curr
}

func (p *Path) Next() *Path {
	return p.Remainder
}

func (p *Path) PushBack(path *Path) {
	tail := p.Tail()
	tail.Remainder = path
}

func (p *Path) Tail() *Path {
	curr := p
	for curr.Remainder != nil {
		curr = curr.Remainder
	}
	return curr
}

func (p *Path) StartsWith0(other *Path) bool {
	left, right := p, other
	for left != nil && right != nil {
		if left.First != right.First {
			return false
		}
		left = left.Remainder
		right = right.Remainder
	}
	return right == nil
}

func (p *Path) StartsWith1(parts []string) bool {
	if len(parts) == 0 {
		return false
	}

	curr := p
	for _, part := range parts {
		if curr == nil {
			return false
		}

		key, ok := curr.First.(*StrKey)
		if !ok || key.Str != part {
			return false
		}

		curr = curr.Remainder
	}

	return true
}

func (p *Path) Join(key Key) {
	p.Remainder = NewPath(key, nil)
}

func (p *Path) String() string {
	parts := []string{p.First.String()}
	curr := p.Remainder
	for curr != nil {
		parts = append(parts, curr.First.String())
		curr = curr.Remainder
	}
	return strings.Join(parts, ".")
}

func (p *Path) Iter() <-chan *Path {
	ch := make(chan *Path)
	go func() {
		curr := p
		for curr != nil {
			ch <- curr
			curr = curr.Remainder
		}
		close(ch)
	}()
	return ch
}
