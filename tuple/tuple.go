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

type Source int

const (
	SourceUnknown = iota
	SourceGithub
	SourceGitlab
)

type Tuple struct {
	Source  Source // tuple source
	Package string // Go package name
	Account string // account
	Project string // project
	Tag     string // tag or commit ID
	Group   string // GH_TUPLE group
	Prefix  string // package prefix
}

var groupRe = regexp.MustCompile(`[^\w]+`)

func (t *Tuple) SetSource(source Source, account, project string) {
	t.Source = source
	t.Account = account
	t.Project = project
	group := account + "_" + project
	group = groupRe.ReplaceAllString(group, "_")
	t.Group = strings.ToLower(group)
}

func (t *Tuple) String() string {
	return fmt.Sprintf("%s:%s:%s:%s/%s/%s", t.Account, t.Project, t.Tag, t.Group, t.Prefix, t.Package)
}

type Tuples []*Tuple

// EnsureUniqueGroups makes sure all Group names are unique.
// This function assumes that tt is pre-sorted in ByAccountAndProject order.
func (tt Tuples) EnsureUniqueGroups() {
	if len(tt) < 2 {
		return
	}

	var prevGroup string
	suffix := 1

	for i, t := range tt {
		if i == 0 {
			prevGroup = t.Group
			continue
		}
		if t.Group == prevGroup {
			t.Group = fmt.Sprintf("%s_%d", t.Group, suffix)
			suffix++
		} else {
			prevGroup = t.Group
			suffix = 1
		}
	}
}

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
	if strings.HasPrefix(si, "#") {
		if strings.HasPrefix(sj, "#") {
			return tt[i].Package < tt[j].Package
		}
		return false
	}
	if strings.HasPrefix(sj, "#") {
		return true
	}
	return tt[i].String() < tt[j].String()
}

var varName = map[Source]string{
	SourceGithub: "GH_TUPLE",
	SourceGitlab: "GL_TUPLE",
}

func (tt Tuples) String() string {
	bufs := map[Source]*bytes.Buffer{
		SourceGithub: new(bytes.Buffer),
		SourceGitlab: new(bytes.Buffer),
	}

	for _, t := range tt {
		b, ok := bufs[t.Source]
		if !ok {
			panic(fmt.Sprintf("unknown tuple source: %v", t.Source))
		}
		var eol string
		if b.Len() == 0 {
			b.WriteString(fmt.Sprintf("%s=\t", varName[t.Source]))
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
	for _, k := range []Source{SourceGithub, SourceGitlab} {
		b := bufs[k].Bytes()
		if len(b) > 0 {
			sb.Write(bufs[k].Bytes())
			sb.WriteRune('\n')
		}
	}

	return sb.String()
}

type Errors struct {
	SourceErrors      []SourceError
	ReplacementErrors []ReplacementError
}

func (ee Errors) IsEmpty() bool {
	return len(ee.SourceErrors) == 0 && len(ee.ReplacementErrors) == 0
}

func (ee Errors) Error() string {
	var buf bytes.Buffer

	if len(ee.SourceErrors) > 0 {
		buf.WriteString("\t\t# Mirrors for the following packages are not currently known, please look them up and handle these tuples manually:\n")
		sort.Slice(ee.SourceErrors, func(i, j int) bool {
			return ee.SourceErrors[i] < ee.SourceErrors[j]
		})
		for _, err := range ee.SourceErrors {
			buf.WriteString(fmt.Sprintf("\t\t#\t%s\n", string(err)))
		}
	}

	if len(ee.ReplacementErrors) > 0 {
		buf.WriteString("\t\t# The following replacement packages are missing version/commit ID, you may need to symlink them in post-patch:\n")
		sort.Slice(ee.ReplacementErrors, func(i, j int) bool {
			return ee.ReplacementErrors[i] < ee.ReplacementErrors[j]
		})
		for _, err := range ee.ReplacementErrors {
			buf.WriteString(fmt.Sprintf("\t\t#\t%s\n", string(err)))
		}
	}

	return buf.String()
}

type ReplacementError string

func (err ReplacementError) Error() string {
	return string(err)
}

type SourceError string

func (err SourceError) Error() string {
	return string(err)
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
		replace := strings.Split(spec, replaceOp)
		if len(replace) != 2 {
			return nil, fmt.Errorf("unexpected number of packages in replace spec: %q", spec)
		}

		var srcPkg string
		srcParts := strings.Fields(replace[0])
		switch len(srcParts) {
		case 1, 2:
			srcPkg = srcParts[0]
		default:
			return nil, fmt.Errorf("unexpected number of fields in the replace spec source: %q", spec)
		}

		tgtParts := strings.Fields(replace[1])
		switch len(tgtParts) {
		case 1:
			return nil, ReplacementError(spec)
		case 2:
			// OK
		default:
			return nil, fmt.Errorf("unexpected number of fields in the replace spec target: %q", spec)
		}

		t, err := New(replace[1], packagePrefix)
		if err != nil {
			return nil, err
		}

		// Keep the target package's account, project and tag but source package name
		t.Package = srcPkg

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
		t.SetSource(SourceGithub, m.Account, m.Project)
	} else {
		switch {
		case strings.HasPrefix(pkg, "github.com"):
			parts := strings.Split(pkg, "/")
			if len(parts) < 3 {
				return nil, fmt.Errorf("unexpected Github package name: %q", pkg)
			}
			t.SetSource(SourceGithub, parts[1], parts[2])
		case strings.HasPrefix(pkg, "gitlab.com"):
			nameParts := strings.Split(pkg, "/")
			if len(nameParts) < 3 {
				return nil, fmt.Errorf("unexpected Gitlab package name: %q", pkg)
			}
			t.SetSource(SourceGitlab, nameParts[1], nameParts[2])
		default:
			for _, vp := range vanity.Parsers {
				if vp.Match(pkg) {
					account, project := vp.Parse(pkg)
					t.SetSource(SourceGithub, account, project)
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

	if t.Source == SourceUnknown {
		return nil, SourceError(t.String())
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
					if !p.offline && t.Source == SourceGitlab {
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

	var tuples Tuples
	var errors Errors

	for res := range ch {
		if err, ok := res.(error); ok {
			switch err := err.(type) {
			case SourceError:
				errors.SourceErrors = append(errors.SourceErrors, err)
			case ReplacementError:
				errors.ReplacementErrors = append(errors.ReplacementErrors, err)
			default:
				return nil, err
			}
		} else {
			tuples = append(tuples, res.(*Tuple))
		}
	}
	sort.Sort(ByAccountAndProject(tuples))
	tuples.EnsureUniqueGroups()

	if !errors.IsEmpty() {
		return tuples, errors
	}

	return tuples, nil
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
