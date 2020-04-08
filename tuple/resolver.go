package tuple

import (
	"fmt"
	"regexp"
	"strings"
)

type SourceError string

func (err SourceError) Error() string {
	return string(err)
}

func Resolve(pkg, version, subdir, link string) (*Tuple, error) {
	t := &Tuple{
		Package: pkg,
		Version: version,
		Subdir:  subdir,
		Link:    link,
		Group:   "group_name",
	}

	for _, r := range resolvers {
		ok, err := r(t)
		if err != nil {
			return nil, err
		}
		if ok {
			break
		}
	}

	if !t.IsResolved() {
		return nil, SourceError(t.String())
	}

	group := t.Account + "_" + t.Project
	if t.Submodule != "" {
		group = group + "_" + t.Submodule
	}
	group = underscoreRe.ReplaceAllString(group, "_")
	group = strings.Trim(group, "_")
	t.Group = strings.ToLower(group)

	linkSuffix := underscoreRe.ReplaceAllString(t.Link, "_")
	linkSuffix = strings.Trim(linkSuffix, "_")
	t.LinkSuffix = strings.ToLower(linkSuffix)

	return t, nil
}

var underscoreRe = regexp.MustCompile(`[^\w]+`)

type mirrorResolver interface {
	resolve(t *Tuple) bool
}

type mirror struct {
	source    Source
	account   string
	project   string
	submodule string
}

func (m mirror) resolve(t *Tuple) bool {
	t.Source = m.source
	t.Account = m.account
	t.Project = m.project
	return true
}

type mirrorFn func(t *Tuple) bool

func (f mirrorFn) resolve(t *Tuple) bool {
	return f(t)
}

var mirrors = map[string]mirrorResolver{
	"contrib.go.opencensus.io/exporter/ocagent": mirror{GH, "census-ecosystem", "opencensus-go-exporter-ocagent", ""},

	"bazil.org":                   mirrorFn(bazilOrgResolver),
	"camlistore.org":              mirror{GH, "perkeep", "perkeep", ""},
	"cloud.google.com":            mirrorFn(cloudGoogleComResolver),
	"code.cloudfoundry.org":       mirrorFn(codeCloudfoundryOrgResolver),
	"docker.io/go-docker":         mirror{GH, "docker", "go-docker", ""},
	"git.apache.org/thrift.git":   mirror{GH, "apache", "thrift", ""},
	"go.bug.st/serial.v1":         mirror{GH, "bugst", "go-serial", ""},
	"go.elastic.co/apm":           mirror{GH, "elastic", "apm-agent-go", ""},
	"go.elastic.co/fastjson":      mirror{GH, "elastic", "go-fastjson", ""},
	"go.etcd.io":                  mirrorFn(goEtcdIoResolver),
	"go.mongodb.org/mongo-driver": mirror{GH, "mongodb", "mongo-go-driver", ""},
	"go.mozilla.org":              mirrorFn(goMozillaOrgResolver),
	"go.opencensus.io":            mirror{GH, "census-instrumentation", "opencensus-go", ""},
	"go.uber.org":                 mirrorFn(goUberOrgResolver),
	"go4.org":                     mirror{GH, "go4org", "go4", ""},
	"gocloud.dev":                 mirror{GH, "google", "go-cloud", ""},
	"golang.org":                  mirrorFn(golangOrgResolver),
	"google.golang.org/api":       mirror{GH, "googleapis", "google-api-go-client", ""},
	"google.golang.org/appengine": mirror{GH, "golang", "appengine", ""},
	"google.golang.org/genproto":  mirror{GH, "google", "go-genproto", ""},
	"google.golang.org/grpc":      mirror{GH, "grpc", "grpc-go", ""},
	"gopkg.in":                    mirrorFn(gopkgInResolver),
	"gotest.tools":                mirror{GH, "gotestyourself", "gotest.tools", ""},
	"honnef.co/go/tools":          mirror{GH, "dominikh", "go-tools", ""},
	"howett.net/plist":            mirror{GitlabSource("https://gitlab.howett.net"), "go", "plist", ""},
	"k8s.io":                      mirrorFn(k8sIoResolver),
	"layeh.com/radius":            mirror{GH, "layeh", "radius", ""},
	"mvdan.cc":                    mirrorFn(mvdanCcResolver),
	"rsc.io":                      mirrorFn(rscIoResolver),
	"sigs.k8s.io/yaml":            mirror{GH, "kubernetes-sigs", "yaml", ""},
	"tinygo.org/x/go-llvm":        mirror{GH, "tinygo-org", "go-llvm", ""},
}

// bazil.org/fuse -> github.com/bazil/fuse
var bazilOrgRe = regexp.MustCompile(`\Abazil\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func bazilOrgResolver(t *Tuple) bool {
	if !bazilOrgRe.MatchString(t.Package) {
		return false
	}
	sm := bazilOrgRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.Source = GH
	t.Account = "bazil"
	t.Project = sm[0][1]
	t.Submodule = ""
	return true
}

// cloud.google.com/go/* -> github.com/googleapis/google-cloud-go
var cloudGoogleComRe = regexp.MustCompile(`\Acloud\.google\.com/go(/([0-9A-Za-z][-0-9A-Za-z]+))?\z`)

func cloudGoogleComResolver(t *Tuple) bool {
	if !cloudGoogleComRe.MatchString(t.Package) {
		return false
	}
	t.Source = GH
	t.Account = "googleapis"
	t.Project = "google-cloud-go"
	t.Submodule = ""
	return true
}

// code.cloudfoundry.org/gofileutils -> github.com/cloudfoundry/gofileutils
var codeCloudfoundryOrgRe = regexp.MustCompile(`\Acode\.cloudfoundry\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func codeCloudfoundryOrgResolver(t *Tuple) bool {
	if !codeCloudfoundryOrgRe.MatchString(t.Package) {
		return false
	}
	sm := codeCloudfoundryOrgRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.Source = GH
	t.Account = "cloudfoundry"
	t.Project = sm[0][1]
	t.Submodule = ""
	return true
}

// go.etcd.io/etcd -> github.com/etcd-io/etcd
var goEtcdIoRe = regexp.MustCompile(`\Ago\.etcd\.io/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func goEtcdIoResolver(t *Tuple) bool {
	if !goEtcdIoRe.MatchString(t.Package) {
		return false
	}
	sm := goEtcdIoRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.Source = GH
	t.Account = "etcd-io"
	t.Project = sm[0][1]
	t.Submodule = ""
	return true
}

// go.mozilla.org/gopgagent -> github.com/mozilla-services/gopgagent
var goMozillaOrgRe = regexp.MustCompile(`\Ago\.mozilla\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func goMozillaOrgResolver(t *Tuple) bool {
	if !goMozillaOrgRe.MatchString(t.Package) {
		return false
	}
	sm := goMozillaOrgRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.Source = GH
	t.Account = "mozilla-services"
	t.Project = sm[0][1]
	t.Submodule = ""
	return true
}

// go.uber.org/zap -> github.com/uber-go/zap
var goUberOrgRe = regexp.MustCompile(`\Ago\.uber\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func goUberOrgResolver(t *Tuple) bool {
	if !goUberOrgRe.MatchString(t.Package) {
		return false
	}
	sm := goUberOrgRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.Source = GH
	t.Account = "uber-go"
	t.Project = sm[0][1]
	t.Submodule = ""
	return true
}

// golang.org/x/pkg -> github.com/golang/pkg
var golangOrgRe = regexp.MustCompile(`\Agolang\.org/x/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func golangOrgResolver(t *Tuple) bool {
	if !golangOrgRe.MatchString(t.Package) {
		return false
	}
	sm := golangOrgRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.Source = GH
	t.Account = "golang"
	t.Project = sm[0][1]
	t.Submodule = ""
	return true
}

// gopkg.in/pkg.v3 -> github.com/go-pkg/pkg
// gopkg.in/user/pkg.v3 -> github.com/user/pkg
var gopkgInRe = regexp.MustCompile(`\Agopkg\.in/([0-9A-Za-z][-0-9A-Za-z]+)(?:\.v.+)?(?:/([0-9A-Za-z][-0-9A-Za-z]+)(?:\.v.+))?\z`)

func gopkgInResolver(t *Tuple) bool {
	if !gopkgInRe.MatchString(t.Package) {
		return false
	}
	// fsnotify is a special case in gopkg.in
	if t.Package == "gopkg.in/fsnotify.v1" {
		t.Source = GH
		t.Account = "fsnotify"
		t.Project = "fsnotify"
		t.Submodule = ""
		return true
	}
	sm := gopkgInRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.Source = GH
	if sm[0][2] == "" {
		t.Account = "go-" + sm[0][1]
		t.Project = sm[0][1]
	} else {
		t.Account = sm[0][1]
		t.Project = sm[0][2]
	}
	t.Submodule = ""
	return true
}

// k8s.io/api -> github.com/kubernetes/api
var k8sIoRe = regexp.MustCompile(`\Ak8s\.io/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func k8sIoResolver(t *Tuple) bool {
	if !k8sIoRe.MatchString(t.Package) {
		return false
	}
	sm := k8sIoRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.Source = GH
	t.Account = "kubernetes"
	t.Project = sm[0][1]
	t.Submodule = ""
	return true
}

// mvdan.cc/editorconfig -> github.com/mvdan/editconfig
var mvdanCcRe = regexp.MustCompile(`\Amvdan\.cc/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func mvdanCcResolver(t *Tuple) bool {
	if !mvdanCcRe.MatchString(t.Package) {
		return false
	}
	sm := mvdanCcRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.Source = GH
	t.Account = "mvdan"
	t.Project = sm[0][1]
	t.Submodule = ""
	return true
}

// rsc.io/pdf -> github.com/rsc/pdf
var rscIoRe = regexp.MustCompile(`\Arsc\.io/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func rscIoResolver(t *Tuple) bool {
	if !rscIoRe.MatchString(t.Package) {
		return false
	}
	sm := rscIoRe.FindAllStringSubmatch(t.Package, -1)
	if len(sm) == 0 {
		return false
	}
	t.Source = GH
	t.Account = "rsc"
	t.Project = sm[0][1]
	t.Submodule = ""
	return true
}

func tryMirror(t *Tuple) (bool, error) {
	// TODO: lookup online unless -offline was given

	for k, v := range mirrors {
		if strings.HasPrefix(t.Package, k) {
			return v.resolve(t), nil
		}
	}
	return false, nil
}

func tryGithub(t *Tuple) (bool, error) {
	if !strings.HasPrefix(t.Package, "github.com") {
		return false, nil
	}
	parts := strings.SplitN(t.Package, "/", 4)
	if len(parts) < 3 {
		return false, fmt.Errorf("unexpected Github package name: %q", t.Package)
	}
	t.Source = GH
	t.Account = parts[1]
	t.Project = parts[2]
	if len(parts) == 4 {
		t.Submodule = parts[3]
	}
	return true, nil
}

func tryGitlab(t *Tuple) (bool, error) {
	if !strings.HasPrefix(t.Package, "gitlab.com") {
		return false, nil
	}
	parts := strings.SplitN(t.Package, "/", 4)
	if len(parts) < 3 {
		return false, fmt.Errorf("unexpected Gitlab package name: %q", t.Package)
	}
	t.Source = GL
	t.Account = parts[1]
	t.Project = parts[2]
	if len(parts) == 4 {
		t.Submodule = parts[3]
	}
	return true, nil
}

var resolvers = []func(*Tuple) (bool, error){
	tryMirror,
	tryGithub,
	tryGitlab,
}

// func tryMirror(pkg string) (*Tuple, error) {
//     for k, v := range mirrors {
//         if strings.HasPrefix(pkg, k) {
//             m := v.parse(pkg)
//             return newTuple(m.source, pkg, m.account, m.project), nil
//         }
//     }
//     return nil, nil
// }
