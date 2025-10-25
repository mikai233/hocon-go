package common

import "fmt"

type ConcatenateDifferentType struct {
	Path      string
	LeftType  string
	RightType string
}

func (c *ConcatenateDifferentType) Error() string {
	return fmt.Sprintf("cannot concatenate different type %s and %s at %s", c.LeftType, c.RightType, c.Path)
}
