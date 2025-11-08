package common

import (
	"fmt"
	"strings"
)

type ConcatenateDifferentType struct {
	Path      string
	LeftType  string
	RightType string
}

func (c *ConcatenateDifferentType) Error() string {
	return fmt.Sprintf("cannot concatenate different type %s and %s at %s", c.LeftType, c.RightType, c.Path)
}

type SubstitutionNotFound struct {
	Path string
}

func (e *SubstitutionNotFound) Error() string {
	return fmt.Sprintf("substitution %s not found", e.Path)
}

type SubstitutionCycle struct {
	Current   string
	Backtrace []string
}

func (e *SubstitutionCycle) Error() string {
	if len(e.Backtrace) == 0 {
		return fmt.Sprintf("substitution cycle detected at %s", e.Current)
	}
	return fmt.Sprintf("substitution cycle: %s -> %s (cycle closed)", strings.Join(e.Backtrace, " -> "), e.Current)
}

type SubstitutionDepthExceeded struct {
	MaxDepth int
}

func (e *SubstitutionDepthExceeded) Error() string {
	return fmt.Sprintf("substitution depth exceeded the limit of %d levels", e.MaxDepth)
}
