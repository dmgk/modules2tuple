package tuple

import "regexp"

type vanityParser func(*Tuple) bool

var vanity = map[string]vanityParser{
	"go.mozilla.org": goMozillaOrgParser,
	"go.uber.org":    goUberOrgParser,
	"golang.org":     golangOrgParser,
	"gopkg.in":       gopkgInParser,
	"k8s.io":         k8sIoParser,
}

func (t *Tuple) fromVanity() bool {
	for _, fn := range vanity {
		if fn(t) {
			return true
		}
	}
	return false
}

// go.mozilla.org/gopgagent -> github.com/mozilla-services/gopgagent
var goMozillaOrgRe = regexp.MustCompile(`\Ago\.mozilla\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func goMozillaOrgParser(t *Tuple) bool {
	if !goMozillaOrgRe.MatchString(t.Package) {
		return false
	}
	sm := goMozillaOrgRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.setSource(SourceGithub, "mozilla-services", sm[0][1])
	return true
}

// go.uber.org/zap -> github.com/uber-go/zap
var goUberOrgRe = regexp.MustCompile(`\Ago\.uber\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func goUberOrgParser(t *Tuple) bool {
	if !goUberOrgRe.MatchString(t.Package) {
		return false
	}
	sm := goUberOrgRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.setSource(SourceGithub, "uber-go", sm[0][1])
	return true
}

// golang.org/x/pkg -> github.com/golang/pkg
var golangOrgRe = regexp.MustCompile(`\Agolang\.org/x/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func golangOrgParser(t *Tuple) bool {
	if !golangOrgRe.MatchString(t.Package) {
		return false
	}
	sm := golangOrgRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.setSource(SourceGithub, "golang", sm[0][1])
	return true
}

// gopkg.in/pkg.v3 -> github.com/go-pkg/pkg
// gopkg.in/user/pkg.v3 -> github.com/user/pkg
var gopkgInRe = regexp.MustCompile(`\Agopkg\.in/([0-9A-Za-z][-0-9A-Za-z]+)(?:\.v.+)?(?:/([0-9A-Za-z][-0-9A-Za-z]+)(?:\.v.+))?\z`)

func gopkgInParser(t *Tuple) bool {
	if !gopkgInRe.MatchString(t.Package) {
		return false
	}
	sm := gopkgInRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	if sm[0][2] == "" {
		t.setSource(SourceGithub, "go-"+sm[0][1], sm[0][1])
	} else {
		t.setSource(SourceGithub, sm[0][1], sm[0][2])
	}
	return true
}

// k8s.io/api -> github.com/kubernetes/api
var k8sIoRe = regexp.MustCompile(`\Ak8s\.io/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func k8sIoParser(t *Tuple) bool {
	if !k8sIoRe.MatchString(t.Package) {
		return false
	}
	sm := k8sIoRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.setSource(SourceGithub, "kubernetes", sm[0][1])
	return true
}
