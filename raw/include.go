package raw

import "fmt"

type Location int

const (
	File Location = iota
	Url
	Classpath
)

func (l Location) String() string {
	switch l {
	case File:
		return "file"
	case Url:
		return "url"
	case Classpath:
		return "classpath"
	default:
		return "unknown"
	}
}

type Inclusion struct {
	Path     string
	Required bool
	Location *Location
	Val      *Object
}

func (inc Inclusion) String() string {
	result := "include "
	if inc.Required {
		result += "required("
	}

	if inc.Location == nil {
		result += fmt.Sprintf("\"%s\"", inc.Path)
	} else {
		result += fmt.Sprintf("%s(\"%s\")", inc.Location.String(), inc.Path)
	}

	if inc.Required {
		result += ")"
	}
	return result
}

func NewInclusion(path string, required bool, location *Location, val *Object) Inclusion {
	return Inclusion{
		Path:     path,
		Required: required,
		Location: location,
		Val:      val,
	}
}
