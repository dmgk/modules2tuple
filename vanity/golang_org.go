package vanity

import "regexp"

// golang.org parser

type golangOrgParser struct {
	parserBase
}

func newGolangOrgParser() *golangOrgParser {
	return &golangOrgParser{
		parserBase{
			"golang.org",
			// golang.org/x/pkg -> github.com/golang/pkg
			regexp.MustCompile(`\Agolang\.org/x/([0-9A-Za-z][-0-9A-Za-z]+)\z`),
		},
	}
}

func (p *golangOrgParser) Parse(pkg string) (string, string) {
	sm := p.rx.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return "", ""
	}
	return "golang", sm[0][1]
}
