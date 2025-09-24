package raw

import "fmt"

type Substitution struct {
	Path     String
	Optional bool
}

func NewSubstitution(path String, optional bool) *Substitution {
	return &Substitution{
		Path:     path,
		Optional: optional,
	}
}

func (s *Substitution) Type() string {
	return SubstitutionType
}

func (s *Substitution) isRawValue() {
}

func (s *Substitution) String() string {
	result := "${"
	if s.Optional {
		result += "?"
	}
	result += s.Path.String()
	result += "}"
	return result
}

func (s *Substitution) Debug() string {
	result := "${"
	if s.Optional {
		result += "?"
	}
	result += fmt.Sprintf("%#v", s.Path)
	result += "}"
	return result
}
