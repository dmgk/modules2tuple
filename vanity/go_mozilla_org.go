package vanity

import "regexp"

type goMozillaOrgParser struct {
	parserBase
}

func newGoMozillaOrgParser() *goMozillaOrgParser {
	return &goMozillaOrgParser{
		parserBase{
			"go.mozilla.org",
			// go.mozilla.org/gopgagent -> github.com/mozilla-services/gopgagent
			regexp.MustCompile(`\Ago\.mozilla\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`),
		},
	}
}

func (p *goMozillaOrgParser) Parse(pkg string) (string, string) {
	sm := p.rx.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return "", ""
	}
	return "mozilla-services", sm[0][1]
}
