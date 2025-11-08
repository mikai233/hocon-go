package parser

import (
	"fmt"
)

type unexpectedTokenError struct {
	Expected string
	Found    byte
}

func (e *unexpectedTokenError) Error() string {
	return fmt.Sprintf("unexpected token: expected %s, found %q", e.Expected, e.Found)
}

type depthExceededError struct {
	Limit int
}

func (e *depthExceededError) Error() string {
	return fmt.Sprintf("recursion depth exceeded limit of %d", e.Limit)
}

type invalidEscapeError struct{}

func (invalidEscapeError) Error() string {
	return "invalid escape sequence"
}
