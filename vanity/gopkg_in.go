package vanity

import "regexp"

// gopkg.in parser

type gopkgInParser struct {
	parserBase
}

func newGopkgInParser() *gopkgInParser {
	return &gopkgInParser{
		parserBase{
			"gopkg.in",

			// gopkg.in/pkg.v3 -> github.com/go-pkg/pkg
			// gopkg.in/user/pkg.v3 -> github.com/user/pkg
			regexp.MustCompile(`\Agopkg\.in/([0-9A-Za-z][-0-9A-Za-z]+)(?:\.v.+)?(?:/([0-9A-Za-z][-0-9A-Za-z]+)(?:\.v.+))?\z`),
		},
	}
}

func (p *gopkgInParser) Parse(pkg string) (string, string) {
	sm := p.rx.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return "", ""
	}
	if sm[0][2] == "" {
		return "go-" + sm[0][1], sm[0][1]
	}
	return sm[0][1], sm[0][2]
}
