package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/dmgk/modules2tuple/gitlab"
	"github.com/dmgk/modules2tuple/vanity"
)

type tupleKind int

const (
	kindGithub = iota
	kindGitlab
)

var varName = map[tupleKind]string{
	kindGithub: "GH_TUPLE",
	kindGitlab: "GL_TUPLE",
}

type tuple struct {
	kind    tupleKind
	pkg     string // Go package name
	account string // account
	project string // project
	tag     string // tag or commit ID
}

// v1.0.0
// v1.0.0+incompatible
// v1.2.3-pre-release-suffix
// v1.2.3-pre-release-suffix+incompatible
var versionRx = regexp.MustCompile(`\A(v\d+\.\d+\.\d+(?:-[0-9A-Za-z]+[0-9A-Za-z\.-]+)?)(?:\+incompatible)?\z`)

// v0.0.0-20181001143604-e0a95dfd547c
// v1.2.3-20181001143604-e0a95dfd547c
// v1.2.3-3.20181001143604-e0a95dfd547c
// v0.8.0-dev.2.0.20180608203834-19279f049241
var tagRx = regexp.MustCompile(`\Av\d+\.\d+\.\d+-(?:[0-9A-Za-z\.]+\.)?\d{14}-([0-9a-f]+)\z`)

func parseTuple(spec string) (*tuple, error) {
	const replaceOp = " => "

	// Replaced package spec
	if strings.Contains(spec, replaceOp) {
		parts := strings.Split(spec, replaceOp)
		if len(parts) != 2 {
			return nil, fmt.Errorf("unexpected number of packages in replace spec: %q", spec)
		}
		tOld, err := parseTuple(parts[0])
		if err != nil {
			return nil, err
		}
		t, err := parseTuple(parts[1])
		if err != nil {
			return nil, err
		}

		// Keep the old package name but with new account, project and tag
		t.pkg = tOld.pkg

		return t, nil
	}

	// Regular package spec
	fields := strings.Fields(spec)
	if len(fields) != 2 {
		return nil, fmt.Errorf("unexpected number of fields: %q", spec)
	}

	pkg := fields[0]
	version := fields[1]
	t := &tuple{pkg: pkg}

	// Parse package name
	if wk, ok := wellKnownPackages[pkg]; ok {
		t.account = wk.account
		t.project = wk.project
	} else {
		switch {
		case strings.HasPrefix(pkg, "github.com"):
			parts := strings.Split(pkg, "/")
			if len(parts) < 3 {
				return nil, fmt.Errorf("unexpected Github package name: %q", pkg)
			}
			t.account = parts[1]
			t.project = parts[2]
		case strings.HasPrefix(pkg, "gitlab.com"):
			nameParts := strings.Split(pkg, "/")
			if len(nameParts) < 3 {
				return nil, fmt.Errorf("unexpected Gitlab package name: %q", pkg)
			}
			t.kind = kindGitlab
			t.account = nameParts[1]
			t.project = nameParts[2]
		default:
			for _, vp := range vanity.Parsers {
				if vp.Match(pkg) {
					t.account, t.project = vp.Parse(pkg)
					break
				}
			}
		}
	}

	// Parse version
	switch {
	case tagRx.MatchString(version):
		sm := tagRx.FindAllStringSubmatch(version, -1)
		t.tag = sm[0][1]
	case versionRx.MatchString(version):
		sm := versionRx.FindAllStringSubmatch(version, -1)
		t.tag = sm[0][1]
	default:
		return nil, fmt.Errorf("unexpected version string: %q", version)
	}

	// Call Gitlab API to translate short commits IDs and tags
	// to the full 32 character commit ID as required by bsd.sites.mk
	if t.kind == kindGitlab {
		c, err := gitlab.GetCommit(t.account, t.project, t.tag)
		if err != nil {
			return nil, err
		}
		t.tag = c.ID
	}

	return t, nil
}

var groupRe = regexp.MustCompile(`[^\w]+`)

func (t *tuple) String() string {
	group := t.account + "_" + t.project
	group = groupRe.ReplaceAllString(group, "_")
	group = strings.ToLower(group)
	var comment string
	if t.account == "" || t.project == "" {
		comment = "# "
	}
	return fmt.Sprintf("%s%s:%s:%s:%s/%s/%s", comment, t.account, t.project, t.tag, group, flagPackagePrefix, t.pkg)
}

type ByAccountAndProject []*tuple

func (pp ByAccountAndProject) Len() int {
	return len(pp)
}

func (pp ByAccountAndProject) Swap(i, j int) {
	pp[i], pp[j] = pp[j], pp[i]
}

func (pp ByAccountAndProject) Less(i, j int) bool {
	// Commented lines sorted last
	si, sj := pp[i].String(), pp[j].String()
	if strings.HasPrefix(si, "#") {
		if strings.HasPrefix(sj, "#") {
			return pp[i].pkg < pp[j].pkg
		}
		return false
	}
	if strings.HasPrefix(sj, "#") {
		return true
	}
	return pp[i].String() < pp[j].String()
}

// List of well known Github mirrors
var wellKnownPackages = map[string]struct {
	account string // Github account
	project string // Github project
}{
	// Package name                              GH Account, GH Project
	"camlistore.org":                            {"perkeep", "perkeep"},
	"cloud.google.com/go":                       {"googleapis", "google-cloud-go"},
	"contrib.go.opencensus.io/exporter/ocagent": {"census-ecosystem", "opencensus-go-exporter-ocagent"},
	"docker.io/go-docker":                       {"docker", "go-docker"},
	"git.apache.org/thrift.git":                 {"apache", "thrift"},
	"go.opencensus.io":                          {"census-instrumentation", "opencensus-go"},
	"go4.org":                                   {"go4org", "go4"},
	"google.golang.org/api":                     {"googleapis", "google-api-go-client"},
	"google.golang.org/appengine":               {"golang", "appengine"},
	"google.golang.org/genproto":                {"google", "go-genproto"},
	"google.golang.org/grpc":                    {"grpc", "grpc-go"},
	"gopkg.in/fsnotify.v1":                      {"fsnotify", "fsnotify"}, // fsnotify is a special case in gopkg.in
	"sigs.k8s.io/yaml":                          {"kubernetes-sigs", "yaml"},
}

func main() {
	flag.Parse()

	if flagVersion {
		fmt.Fprintln(os.Stderr, version)
		os.Exit(0)
	}

	args := flag.Args()

	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	file, err := os.Open(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	const specPrefix = "# "
	var tuples []*tuple

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, specPrefix) {
			t, err := parseTuple(strings.TrimPrefix(line, specPrefix))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			tuples = append(tuples, t)
		}
	}

	sort.Sort(ByAccountAndProject(tuples))

	bufs := map[tupleKind]*bytes.Buffer{
		kindGithub: new(bytes.Buffer),
		kindGitlab: new(bytes.Buffer),
	}

	for _, t := range tuples {
		b, ok := bufs[t.kind]
		if !ok {
			panic(fmt.Sprintf("unknown tuple kind: %v", t.kind))
		}
		var eol string
		if b.Len() == 0 {
			b.WriteString(fmt.Sprintf("%s=\t", varName[t.kind]))
			eol = `\`
		}
		s := t.String()
		if strings.HasPrefix(s, "#") {
			eol = ""
		} else if eol == "" {
			eol = ` \`
		}
		b.WriteString(fmt.Sprintf("%s\n\t\t%s", eol, s))
	}
	for _, k := range []tupleKind{kindGithub, kindGitlab} {
		b := bufs[k]
		if b.Len() > 0 {
			fmt.Println(b.String())
		}
	}
}

var helpTemplate = template.Must(template.New("help").Parse(`
Vendor package dependencies and then run {{.Name}} on vendor/modules.txt:

	$ go mod vendor
	$ {{.Name}} vendor/modules.txt

By default, generated GH_TUPLE entries will place packages under "vendor".
This can be changed by passing different prefix using -prefix option (e.g. -prefix src).
`))

var (
	flagPackagePrefix string
	flagVersion       bool
)

var version = "devel"

func init() {
	basename := path.Base(os.Args[0])
	flag.StringVar(&flagPackagePrefix, "prefix", "vendor", "package prefix")
	flag.BoolVar(&flagVersion, "v", false, "show version")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] modules.txt\n", basename)
		flag.PrintDefaults()
		helpTemplate.Execute(os.Stderr, map[string]string{
			"Name": basename,
		})
	}
}
