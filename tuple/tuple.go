package tuple

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/dmgk/modules2tuple/gitlab"
	"github.com/dmgk/modules2tuple/vanity"
)

type Kind int

const (
	KindGithub = iota
	KindGitlab
)

type Tuple struct {
	Kind    Kind   // tuple kind
	Package string // Go package name
	Account string // account
	Project string // project
	Tag     string // tag or commit ID
	Prefix  string // package prefix
}

const comment = "#"

var groupRe = regexp.MustCompile(`[^\w]+`)

func (t *Tuple) String() string {
	group := t.Account + "_" + t.Project
	group = groupRe.ReplaceAllString(group, "_")
	group = strings.ToLower(group)
	var bol string
	if t.Account == "" || t.Project == "" {
		bol = comment
	}
	return fmt.Sprintf("%s%s:%s:%s:%s/%s/%s", bol, t.Account, t.Project, t.Tag, group, t.Prefix, t.Package)
}

type Tuples []*Tuple

type ByAccountAndProject Tuples

func (tt ByAccountAndProject) Len() int {
	return len(tt)
}

func (tt ByAccountAndProject) Swap(i, j int) {
	tt[i], tt[j] = tt[j], tt[i]
}

func (tt ByAccountAndProject) Less(i, j int) bool {
	// Commented lines sorted last
	si, sj := tt[i].String(), tt[j].String()
	if strings.HasPrefix(si, comment) {
		if strings.HasPrefix(sj, comment) {
			return tt[i].Package < tt[j].Package
		}
		return false
	}
	if strings.HasPrefix(sj, comment) {
		return true
	}
	return tt[i].String() < tt[j].String()
}

var varName = map[Kind]string{
	KindGithub: "GH_TUPLE",
	KindGitlab: "GL_TUPLE",
}

func (tt Tuples) String() string {
	bufs := map[Kind]*bytes.Buffer{
		KindGithub: new(bytes.Buffer),
		KindGitlab: new(bytes.Buffer),
	}

	for _, t := range tt {
		b, ok := bufs[t.Kind]
		if !ok {
			panic(fmt.Sprintf("unknown tuple kind: %v", t.Kind))
		}
		var eol string
		if b.Len() == 0 {
			b.WriteString(fmt.Sprintf("%s=\t", varName[t.Kind]))
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

	var sb strings.Builder
	for _, k := range []Kind{KindGithub, KindGitlab} {
		b := bufs[k].Bytes()
		if len(b) > 0 {
			sb.WriteRune('\n')
			sb.Write(bufs[k].Bytes())
		}
	}

	return sb.String()
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

func New(spec, packagePrefix string) (*Tuple, error) {
	const replaceOp = " => "

	// Replaced package spec
	if strings.Contains(spec, replaceOp) {
		parts := strings.Split(spec, replaceOp)
		if len(parts) != 2 {
			return nil, fmt.Errorf("unexpected number of packages in replace spec: %q", spec)
		}
		tOld, err := New(parts[0], packagePrefix)
		if err != nil {
			return nil, err
		}
		t, err := New(parts[1], packagePrefix)
		if err != nil {
			return nil, err
		}

		// Keep the old package name but with new account, project and tag
		t.Package = tOld.Package

		return t, nil
	}

	// Regular package spec
	fields := strings.Fields(spec)
	if len(fields) != 2 {
		return nil, fmt.Errorf("unexpected number of fields: %q", spec)
	}

	pkg := fields[0]
	version := fields[1]
	t := &Tuple{Package: pkg, Prefix: packagePrefix}

	// Parse package name
	if m, ok := mirrors[pkg]; ok {
		t.Account = m.Account
		t.Project = m.Project
	} else {
		switch {
		case strings.HasPrefix(pkg, "github.com"):
			parts := strings.Split(pkg, "/")
			if len(parts) < 3 {
				return nil, fmt.Errorf("unexpected Github package name: %q", pkg)
			}
			t.Account = parts[1]
			t.Project = parts[2]
		case strings.HasPrefix(pkg, "gitlab.com"):
			nameParts := strings.Split(pkg, "/")
			if len(nameParts) < 3 {
				return nil, fmt.Errorf("unexpected Gitlab package name: %q", pkg)
			}
			t.Kind = KindGitlab
			t.Account = nameParts[1]
			t.Project = nameParts[2]
		default:
			for _, vp := range vanity.Parsers {
				if vp.Match(pkg) {
					t.Account, t.Project = vp.Parse(pkg)
					break
				}
			}
		}
	}

	// Parse version
	switch {
	case tagRx.MatchString(version):
		sm := tagRx.FindAllStringSubmatch(version, -1)
		t.Tag = sm[0][1]
	case versionRx.MatchString(version):
		sm := versionRx.FindAllStringSubmatch(version, -1)
		t.Tag = sm[0][1]
	default:
		return nil, fmt.Errorf("unexpected version string: %q", version)
	}

	return t, nil
}

type Parser struct {
	packagePrefix string
	offline       bool
}

// NewParser creates a new modules.txt parser with given options.
func NewParser(packagePrefix string, offline bool) *Parser {
	return &Parser{packagePrefix, offline}
}

// Read parses tuples from modules.txt contents provided as io.Reader.
func (p *Parser) Read(r io.Reader) (Tuples, error) {
	ch := make(chan interface{})

	go func() {
		defer close(ch)

		const specPrefix = "# "
		scanner := bufio.NewScanner(r)
		sem := make(chan int, runtime.NumCPU())
		var wg sync.WaitGroup

		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, specPrefix) {
				sem <- 1
				wg.Add(1)
				go func() {
					defer func() {
						<-sem
						wg.Done()
					}()
					t, err := New(strings.TrimPrefix(line, specPrefix), p.packagePrefix)
					if err != nil {
						ch <- err
						return
					}
					// Call Gitlab API to translate go.mod short commit IDs and tags
					// to the full 40-character commit IDs as required by bsd.sites.mk
					if !p.offline && t.Kind == KindGitlab {
						c, err := gitlab.GetCommit(t.Account, t.Project, t.Tag)
						if err != nil {
							ch <- err
							return
						}
						t.Tag = c.ID
					}
					ch <- t
				}()
			}
		}
		wg.Wait()
	}()

	var tt Tuples
	for res := range ch {
		if err, ok := res.(error); ok {
			return nil, err
		}
		tt = append(tt, res.(*Tuple))
	}
	sort.Sort(ByAccountAndProject(tt))

	return tt, nil
}

// Load parses tuples from modules.txt found at path.
func (p *Parser) Load(path string) (Tuples, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return p.Read(f)
}
