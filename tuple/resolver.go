package tuple

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dmgk/modules2tuple/config"
	"github.com/dmgk/modules2tuple/debug"
)

type SourceError string

func (err SourceError) Error() string {
	return string(err)
}

// Resolve looks up mirrors and parses tuple account and project.
func Resolve(pkg, version, subdir, link_target string) (*Tuple, error) {
	t := &Tuple{
		pkg:      pkg,
		version:  version,
		subdir:   subdir,
		link_tgt: link_target,
		group:    "group_name",
	}

	var done bool
	for {
		// try static mirror lookup first
		for _, r := range resolvers {
			if strings.HasPrefix(pkg, r.prefix) {
				m, err := r.resolver.resolve(pkg)
				if err != nil {
					return nil, err
				}
				if m != nil {
					t.makeResolved(m.source, m.account, m.project, m.module)
					return t, nil
				}
			}
		}

		if done || config.Offline {
			break
		}

		// try looking up missing mirror online
		m, err := discoverMirrors(pkg)
		if err != nil {
			return nil, err
		}
		debug.Printf("[tuple.Resolve] discovered mirror %q for %q\n", m, pkg)
		pkg = m
		done = true
	}

	return nil, SourceError(fmt.Sprintf("%s (from %s@%s)", t.String(), pkg, version))
}

type mirror struct {
	source  Source
	account string
	project string
	module  string
}

type resolver interface {
	resolve(pkg string) (*mirror, error)
}

func (m *mirror) resolve(string) (*mirror, error) {
	return m, nil
}

type mirrorFn func(pkg string) (*mirror, error)

func (f mirrorFn) resolve(pkg string) (*mirror, error) {
	return f(pkg)
}

var resolvers = []struct {
	prefix   string
	resolver resolver
}{
	// Docker is a special snowflake
	{"github.com/docker/docker", &mirror{GH, "moby", "moby", ""}},

	{"github.com", mirrorFn(githubResolver)},
	{"gitlab.com", mirrorFn(gitlabResolver)},

	{"contrib.go.opencensus.io/exporter/ocagent", &mirror{GH, "census-ecosystem", "opencensus-go-exporter-ocagent", ""}},
	{"aletheia.icu/broccoli/fs", &mirror{GH, "aletheia-icu", "broccoli", "fs"}},
	{"bazil.org", mirrorFn(bazilOrgResolver)},
	{"camlistore.org", &mirror{GH, "perkeep", "perkeep", ""}},
	{"cloud.google.com", mirrorFn(cloudGoogleComResolver)},
	{"docker.io/go-docker", &mirror{GH, "docker", "go-docker", ""}},
	{"git.apache.org/thrift.git", &mirror{GH, "apache", "thrift", ""}},
	{"go.bug.st/serial.v1", &mirror{GH, "bugst", "go-serial", ""}},
	{"go.elastic.co/apm", mirrorFn(goElasticCoResolver)},
	{"go.elastic.co/fastjson", &mirror{GH, "elastic", "go-fastjson", ""}},
	{"go.etcd.io", mirrorFn(goEtcdIoResolver)},
	{"go.mongodb.org/mongo-driver", &mirror{GH, "mongodb", "mongo-go-driver", ""}},
	{"go.mozilla.org", mirrorFn(goMozillaOrgResolver)},
	{"go.opencensus.io", &mirror{GH, "census-instrumentation", "opencensus-go", ""}},
	{"go.uber.org", mirrorFn(goUberOrgResolver)},
	{"go4.org", &mirror{GH, "go4org", "go4", ""}},
	{"gocloud.dev", &mirror{GH, "google", "go-cloud", ""}},
	{"golang.org", mirrorFn(golangOrgResolver)},
	{"golang.zx2c4.com/wireguard", &mirror{GH, "wireguard", "wireguard-go", ""}},
	{"google.golang.org/api", &mirror{GH, "googleapis", "google-api-go-client", ""}},
	{"google.golang.org/appengine", &mirror{GH, "golang", "appengine", ""}},
	{"google.golang.org/genproto", &mirror{GH, "google", "go-genproto", ""}},
	{"google.golang.org/grpc", &mirror{GH, "grpc", "grpc-go", ""}},
	{"google.golang.org/protobuf", &mirror{GH, "protocolbuffers", "protobuf-go", ""}},
	{"gopkg.in", mirrorFn(gopkgInResolver)},
	{"gotest.tools", mirrorFn(gotestToolsResolver)},
	{"honnef.co/go/tools", &mirror{GH, "dominikh", "go-tools", ""}},
	{"howett.net/plist", &mirror{GitlabSource("https://gitlab.howett.net"), "go", "plist", ""}},
	{"k8s.io", mirrorFn(k8sIoResolver)},
	{"launchpad.net/gocheck", &mirror{GH, "go-check", "check", ""}},
	{"layeh.com/radius", &mirror{GH, "layeh", "radius", ""}},
	{"mvdan.cc", mirrorFn(mvdanCcResolver)},
	{"rsc.io", mirrorFn(rscIoResolver)},
	{"sigs.k8s.io/yaml", &mirror{GH, "kubernetes-sigs", "yaml", ""}},
	{"tinygo.org/x/go-llvm", &mirror{GH, "tinygo-org", "go-llvm", ""}},
}

func githubResolver(pkg string) (*mirror, error) {
	if !strings.HasPrefix(pkg, "github.com") {
		return nil, nil
	}
	parts := strings.SplitN(pkg, "/", 4)
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected Github package name: %q", pkg)
	}
	var module string
	if len(parts) == 4 {
		module = parts[3]
	}
	return &mirror{GH, parts[1], parts[2], module}, nil
}

func gitlabResolver(pkg string) (*mirror, error) {
	if !strings.HasPrefix(pkg, "gitlab.com") {
		return nil, nil
	}
	parts := strings.SplitN(pkg, "/", 4)
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected Gitlab package name: %q", pkg)
	}
	var module string
	if len(parts) == 4 {
		module = parts[3]
	}
	return &mirror{GL, parts[1], parts[2], module}, nil
}

// bazil.org/fuse -> github.com/bazil/fuse
var bazilOrgRe = regexp.MustCompile(`\Abazil\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func bazilOrgResolver(pkg string) (*mirror, error) {
	if !bazilOrgRe.MatchString(pkg) {
		return nil, nil
	}
	sm := bazilOrgRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil, nil
	}
	return &mirror{GH, "bazil", sm[0][1], ""}, nil
}

// cloud.google.com/go/* -> github.com/googleapis/google-cloud-go
var cloudGoogleComRe = regexp.MustCompile(`\Acloud\.google\.com/go(?:/([0-9A-Za-z][-0-9A-Za-z]+))?\z`)

func cloudGoogleComResolver(pkg string) (*mirror, error) {
	if !cloudGoogleComRe.MatchString(pkg) {
		return nil, nil
	}
	var module string
	sm := cloudGoogleComRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) > 0 {
		module = sm[0][1]
	}
	return &mirror{GH, "googleapis", "google-cloud-go", module}, nil
}

// // code.cloudfoundry.org/gofileutils -> github.com/cloudfoundry/gofileutils
// var codeCloudfoundryOrgRe = regexp.MustCompile(`\Acode\.cloudfoundry\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

// func codeCloudfoundryOrgResolver(pkg string) (*mirror, error) {
// 	if !codeCloudfoundryOrgRe.MatchString(pkg) {
// 		return nil, nil
// 	}
// 	sm := codeCloudfoundryOrgRe.FindAllStringSubmatch(pkg, -1)
// 	if len(sm) == 0 {
// 		return nil, nil
// 	}
// 	return &mirror{GH, "cloudfoundry", sm[0][1], ""}, nil
// }

// go.elastic.co/apm -> github.com/elastic/apm-agent-go
// go.elastic.co/apm/module/apmhttp -> github.com/elastic/apm-agent-go/module/apmhttp
var goElasticCoModuleRe = regexp.MustCompile(`\Ago\.elastic\.co/apm(?:/(module/[0-9A-Za-z][-0-9A-Za-z]+))?\z`)

func goElasticCoResolver(pkg string) (*mirror, error) {
	if !goElasticCoModuleRe.MatchString(pkg) {
		return nil, nil
	}
	var module string
	sm := goElasticCoModuleRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) > 0 {
		module = sm[0][1]
	}
	return &mirror{GH, "elastic", "apm-agent-go", module}, nil
}

// go.etcd.io/etcd -> github.com/etcd-io/etcd
var goEtcdIoRe = regexp.MustCompile(`\Ago\.etcd\.io/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func goEtcdIoResolver(pkg string) (*mirror, error) {
	if !goEtcdIoRe.MatchString(pkg) {
		return nil, nil
	}
	sm := goEtcdIoRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil, nil
	}
	return &mirror{GH, "etcd-io", sm[0][1], ""}, nil
}

// go.mozilla.org/gopgagent -> github.com/mozilla-services/gopgagent
var goMozillaOrgRe = regexp.MustCompile(`\Ago\.mozilla\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func goMozillaOrgResolver(pkg string) (*mirror, error) {
	if !goMozillaOrgRe.MatchString(pkg) {
		return nil, nil
	}
	sm := goMozillaOrgRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil, nil
	}
	return &mirror{GH, "mozilla-services", sm[0][1], ""}, nil
}

// go.uber.org/zap -> github.com/uber-go/zap
var goUberOrgRe = regexp.MustCompile(`\Ago\.uber\.org/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func goUberOrgResolver(pkg string) (*mirror, error) {
	if !goUberOrgRe.MatchString(pkg) {
		return nil, nil
	}
	sm := goUberOrgRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil, nil
	}
	return &mirror{GH, "uber-go", sm[0][1], ""}, nil
}

// golang.org/x/pkg -> github.com/golang/pkg
var golangOrgRe = regexp.MustCompile(`\Agolang\.org/x/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func golangOrgResolver(pkg string) (*mirror, error) {
	if !golangOrgRe.MatchString(pkg) {
		return nil, nil
	}
	sm := golangOrgRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil, nil
	}
	return &mirror{GH, "golang", sm[0][1], ""}, nil
}

// gopkg.in/pkg.v3 -> github.com/go-pkg/pkg
// gopkg.in/user/pkg.v3 -> github.com/user/pkg
var gopkgInRe = regexp.MustCompile(`\Agopkg\.in/([0-9A-Za-z][-0-9A-Za-z]+)(?:\.v.+)?(?:/([0-9A-Za-z][-0-9A-Za-z]+)(?:\.v.+))?\z`)

func gopkgInResolver(pkg string) (*mirror, error) {
	if !gopkgInRe.MatchString(pkg) {
		return nil, nil
	}
	// fsnotify is a special case in gopkg.in
	if pkg == "gopkg.in/fsnotify.v1" {
		return &mirror{GH, "fsnotify", "fsnotify", ""}, nil
	}
	sm := gopkgInRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil, nil
	}
	var account, project string
	if sm[0][2] == "" {
		account = "go-" + sm[0][1]
		project = sm[0][1]
	} else {
		account = sm[0][1]
		project = sm[0][2]
	}
	return &mirror{GH, account, project, ""}, nil
}

// k8s.io/api -> github.com/kubernetes/api
var k8sIoRe = regexp.MustCompile(`\Ak8s\.io/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func k8sIoResolver(pkg string) (*mirror, error) {
	if !k8sIoRe.MatchString(pkg) {
		return nil, nil
	}
	sm := k8sIoRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil, nil
	}
	return &mirror{GH, "kubernetes", sm[0][1], ""}, nil
}

// mvdan.cc/editorconfig -> github.com/mvdan/editconfig
var mvdanCcRe = regexp.MustCompile(`\Amvdan\.cc/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func mvdanCcResolver(pkg string) (*mirror, error) {
	if !mvdanCcRe.MatchString(pkg) {
		return nil, nil
	}
	sm := mvdanCcRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil, nil
	}
	return &mirror{GH, "mvdan", sm[0][1], ""}, nil
}

// rsc.io/pdf -> github.com/rsc/pdf
var rscIoRe = regexp.MustCompile(`\Arsc\.io/([0-9A-Za-z][-0-9A-Za-z]+)\z`)

func rscIoResolver(pkg string) (*mirror, error) {
	if !rscIoRe.MatchString(pkg) {
		return nil, nil
	}
	sm := rscIoRe.FindAllStringSubmatch(pkg, -1)
	if len(sm) == 0 {
		return nil, nil
	}
	return &mirror{GH, "rsc", sm[0][1], ""}, nil
}

func gotestToolsResolver(pkg string) (*mirror, error) {
	switch pkg {
	case "gotest.tools":
		return &mirror{GH, "gotestyourself", "gotest.tools", ""}, nil
	case "gotest.tools/gotestsum":
		return &mirror{GH, "gotestyourself", "gotestsum", ""}, nil
	}
	return nil, nil
}
