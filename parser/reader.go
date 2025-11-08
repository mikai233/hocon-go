package parser

import (
	"errors"
	"io"
	"unicode/utf8"
)

var errEOF = io.EOF

type reader struct {
	data []byte
	idx  int
}

func newReader(data []byte) *reader {
	return &reader{data: data}
}

func (r *reader) remaining() []byte {
	if r.idx >= len(r.data) {
		return nil
	}
	return r.data[r.idx:]
}

func (r *reader) peek() (byte, error) {
	if r.idx >= len(r.data) {
		return 0, errEOF
	}
	return r.data[r.idx], nil
}

func (r *reader) peekN(n int) ([]byte, error) {
	if n <= 0 {
		return nil, errors.New("peekN requires n>0")
	}
	if r.idx+n > len(r.data) {
		return nil, errEOF
	}
	return r.data[r.idx : r.idx+n], nil
}

func (r *reader) peek2() (byte, byte, error) {
	buf, err := r.peekN(2)
	if err != nil {
		return 0, 0, err
	}
	return buf[0], buf[1], nil
}

func (r *reader) next() (byte, error) {
	if r.idx >= len(r.data) {
		return 0, errEOF
	}
	ch := r.data[r.idx]
	r.idx++
	return ch, nil
}

func (r *reader) discard(n int) error {
	if n < 0 {
		return errors.New("cannot discard negative amount")
	}
	if r.idx+n > len(r.data) {
		return errEOF
	}
	r.idx += n
	return nil
}

func (r *reader) peekRune() (rune, int, error) {
	if r.idx >= len(r.data) {
		return 0, 0, errEOF
	}
	rn, size := utf8.DecodeRune(r.data[r.idx:])
	if rn == utf8.RuneError && size == 1 {
		return 0, 0, errors.New("invalid utf-8")
	}
	return rn, size, nil
}
