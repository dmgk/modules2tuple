package vanity

import "regexp"

// go.uber.org parser

type goUberOrgParser struct {
	parserBase
}

func newGoUberOrgParser() *goUberOrgParser {
	return &goUberOrgParser{
		parserBase{
			"go.uber.org",
			// go.uber.org/zap -> github.com/uber-go/zap
			regexp.MustCompile(`\Ago\.uber\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`),
		},
	}
}

func (p *goUberOrgParser) Parse(pkg string) (string, string) {
	sm := p.rx.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return "", ""
	}
	return "uber-go", sm[0][1]
}
