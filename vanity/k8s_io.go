package vanity

import "regexp"

type k8sIoParser struct {
	parserBase
}

func newK8sIoParser() *k8sIoParser {
	return &k8sIoParser{
		parserBase{
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
