package vanity

import "regexp"

type Parser interface {
	Name() string // for testing
	Match(pkg string) bool
	Parse(pkg string) (string, string)
}

type parserBase struct {
	name string
	rx   *regexp.Regexp
}

func (p parserBase) Name() string {
	return p.name
}

func (p parserBase) Match(pkg string) bool {
	return p.rx.MatchString(pkg)
}

var Parsers = []Parser{
	newGopkgInParser(),
	newGolangOrgParser(),
	newK8sIoParser(),
	newGoUberOrgParser(),
}
