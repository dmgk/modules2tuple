package tuple

import "regexp"

type vanityParser func(string, string) *Tuple

var vanity = map[string]vanityParser{
	"code.cloudfoundry.org": codeCloudfoundryOrgParser,
	"go.etcd.io":            goEtcdIoParser,
	"go.mozilla.org":        goMozillaOrgParser,
	"go.uber.org":           goUberOrgParser,
	"golang.org":            golangOrgParser,
	"gopkg.in":              gopkgInParser,
	"k8s.io":                k8sIoParser,
}

func tryVanity(pkg, packagePrefix string) (*Tuple, error) {
	for _, fn := range vanity {
		if t := fn(pkg, packagePrefix); t != nil {
			return t, nil
		}
	}
	return nil, nil
}

// code.cloudfoundry.org/gofileutils -> github.com/cloudfoundry/gofileutils
var codeCloudfoundryOrgRe = regexp.MustCompile(`\Acode\.cloudfoundry\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func codeCloudfoundryOrgParser(pkg, packagePrefix string) *Tuple {
	if !codeCloudfoundryOrgRe.MatchString(pkg) {
		return nil
	}
	sm := codeCloudfoundryOrgRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil
	}
	return newTuple(GH{}, pkg, "cloudfoundry", sm[0][1], packagePrefix)
}

// go.etcd.io/etcd -> github.com/etcd-io/etcd
var goEtcdIoRe = regexp.MustCompile(`\Ago\.etcd\.io/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func goEtcdIoParser(pkg, packagePrefix string) *Tuple {
	if !goEtcdIoRe.MatchString(pkg) {
		return nil
	}
	sm := goEtcdIoRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil
	}
	return newTuple(GH{}, pkg, "etcd-io", sm[0][1], packagePrefix)
}

// go.mozilla.org/gopgagent -> github.com/mozilla-services/gopgagent
var goMozillaOrgRe = regexp.MustCompile(`\Ago\.mozilla\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func goMozillaOrgParser(pkg, packagePrefix string) *Tuple {
	if !goMozillaOrgRe.MatchString(pkg) {
		return nil
	}
	sm := goMozillaOrgRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil
	}
	return newTuple(GH{}, pkg, "mozilla-services", sm[0][1], packagePrefix)
}

// go.uber.org/zap -> github.com/uber-go/zap
var goUberOrgRe = regexp.MustCompile(`\Ago\.uber\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func goUberOrgParser(pkg, packagePrefix string) *Tuple {
	if !goUberOrgRe.MatchString(pkg) {
		return nil
	}
	sm := goUberOrgRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil
	}
	return newTuple(GH{}, pkg, "uber-go", sm[0][1], packagePrefix)
}

// golang.org/x/pkg -> github.com/golang/pkg
var golangOrgRe = regexp.MustCompile(`\Agolang\.org/x/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func golangOrgParser(pkg, packagePrefix string) *Tuple {
	if !golangOrgRe.MatchString(pkg) {
		return nil
	}
	sm := golangOrgRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil
	}
	return newTuple(GH{}, pkg, "golang", sm[0][1], packagePrefix)
}

// gopkg.in/pkg.v3 -> github.com/go-pkg/pkg
// gopkg.in/user/pkg.v3 -> github.com/user/pkg
var gopkgInRe = regexp.MustCompile(`\Agopkg\.in/([0-9A-Za-z][-0-9A-Za-z]+)(?:\.v.+)?(?:/([0-9A-Za-z][-0-9A-Za-z]+)(?:\.v.+))?\z`)

func gopkgInParser(pkg, packagePrefix string) *Tuple {
	if !gopkgInRe.MatchString(pkg) {
		return nil
	}
	sm := gopkgInRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil
	}
	if sm[0][2] == "" {
		return newTuple(GH{}, pkg, "go-"+sm[0][1], sm[0][1], packagePrefix)
	}
	return newTuple(GH{}, pkg, sm[0][1], sm[0][2], packagePrefix)
}

// k8s.io/api -> github.com/kubernetes/api
var k8sIoRe = regexp.MustCompile(`\Ak8s\.io/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func k8sIoParser(pkg, packagePrefix string) *Tuple {
	if !k8sIoRe.MatchString(pkg) {
		return nil
	}
	sm := k8sIoRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil
	}
	return newTuple(GH{}, pkg, "kubernetes", sm[0][1], packagePrefix)
}
