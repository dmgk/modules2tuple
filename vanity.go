package main

import "regexp"

type vanityParser interface {
	Name() string
	Match(pkg string) bool
	Parse(pkg string) (string, string)
}

type vanityParserBase struct {
	name string // for failure reporting in tests
	rx   *regexp.Regexp
}

func (p vanityParserBase) Name() string {
	return p.name
}

func (p vanityParserBase) Match(pkg string) bool {
	return p.rx.MatchString(pkg)
}

// gopkg.in

type gopkgInParser struct {
	vanityParserBase
}

func newGopkgInParser() *gopkgInParser {
	return &gopkgInParser{
		vanityParserBase{
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

// golang.org

type golangOrgParser struct {
	vanityParserBase
}

func newGolangOrgParser() *golangOrgParser {
	return &golangOrgParser{
		vanityParserBase{
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

// k8s.io

type k8sIoParser struct {
	vanityParserBase
}

func newK8sIoParser() *k8sIoParser {
	return &k8sIoParser{
		vanityParserBase{
			"k8s.io",
			// k8s.io/api -> github.com/kubernetes/api
			regexp.MustCompile(`\Ak8s\.io/([0-9A-Za-z][-0-9A-Za-z]+)\z`),
		},
	}
}

func (p *k8sIoParser) Parse(pkg string) (string, string) {
	sm := p.rx.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return "", ""
	}
	return "kubernetes", sm[0][1]
}

// go.uber.org

type goUberOrgParser struct {
	vanityParserBase
}

func newGoUberOrgParser() *goUberOrgParser {
	return &goUberOrgParser{
		vanityParserBase{
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

var vanityParsers = []vanityParser{
	newGopkgInParser(),
	newGolangOrgParser(),
	newK8sIoParser(),
	newGoUberOrgParser(),
}
