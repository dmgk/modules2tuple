package tuple

type Source interface {
	String() string
}

type GithubSource string

func (s GithubSource) String() string {
	return string(s)
}

type GitlabSource string

func (s GitlabSource) String() string {
	return string(s)
}

// GH is Github default source
var GH GithubSource

// GL is Gitlab default source
var GL GitlabSource

func sourceVarName(s Source) string {
	switch s.(type) {
	case GithubSource:
		return "GH_TUPLE"
	case GitlabSource:
		return "GL_TUPLE"
	default:
		panic("unknown source type")
	}
}
