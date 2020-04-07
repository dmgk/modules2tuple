package tuple

import "fmt"

type Source interface {
	Site() string
	IsDefaultSite() bool
	VarName() string
	String() string
}

type GH struct{}

func (s GH) Site() string {
	return ""
}

func (s GH) IsDefaultSite() bool {
	return true // there's only one Github
}

func (s GH) VarName() string {
	return "GH_TUPLE"
}

func (s GH) String() string {
	return "GH"
}

type GL struct {
	string
}

const defaultSiteGitlab = "https://gitlab.com"

func (s GL) Site() string {
	if s.string != "" {
		return s.string
	}
	return defaultSiteGitlab
}

func (s GL) IsDefaultSite() bool {
	return s.string == "" || s.string == defaultSiteGitlab
}

func (s GL) VarName() string {
	return "GL_TUPLE"
}

func (s GL) String() string {
	return fmt.Sprintf("GH{%q}", s.string)
}
