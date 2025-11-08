package merge

import (
	"hocon-go/common"
	"hocon-go/raw"
)

type Substitution struct {
	Path     *common.Path
	Optional bool
}

func NewSubstitution(path *common.Path, optional bool) *Substitution {
	return &Substitution{
		Path:     path,
		Optional: optional,
	}
}

func (s *Substitution) Type() string {
	return raw.SubstitutionType
}

func (s *Substitution) isMergeValue() {
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

func (s *Substitution) FullPath() string {
	if s.Path == nil {
		return ""
	}
	return s.Path.String()
}
