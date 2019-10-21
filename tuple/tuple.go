package tuple

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
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

func (t *Tuple) String() string {
	if t.Source != nil && !t.Source.IsDefaultSite() {
		return fmt.Sprintf("%s:%s:%s:%s:%s/%s/%s", t.Source.Site(), t.Account, t.Project, t.Tag, t.Group, t.Prefix, t.Package)
	}
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
	return tt[i].String() < tt[j].String()
}

func (tt Tuples) String() string {
	if len(tt) == 0 {
		return ""
	}

	bufs := make(map[reflect.Type]*bytes.Buffer)

	for _, t := range tt {
		st := reflect.TypeOf(t.Source)
		if st == nil {
			panic(fmt.Sprintf("unknown source in tuple: %v", t))
		}

		buf, ok := bufs[st]
		if !ok {
			buf = &bytes.Buffer{}
			bufs[st] = buf
		}

		var eol string
		if buf.Len() == 0 {
			buf.WriteString(fmt.Sprintf("%s=\t", t.Source.VarName()))
			eol = `\`
		}
		s := t.String()
		if strings.HasPrefix(s, "#") {
			eol = ""
		} else if eol == "" {
			eol = ` \`
		}
		buf.WriteString(fmt.Sprintf("%s\n\t\t%s", eol, s))
	}

	var ss []string
	for _, buf := range bufs {
		s := buf.String()
		if len(s) > 0 {
			ss = append(ss, s)
		}
	}
	sort.Sort(sort.StringSlice(ss))

	return fmt.Sprintf("%s\n", strings.Join(ss, "\n\n"))
}

type Errors struct {
	Source                     []SourceError
	ReplacementLocalFilesystem []ReplacementLocalFilesystemError
	ReplacementMissingCommit   []ReplacementMissingCommitError
}

func (ee Errors) Any() bool {
	return len(ee.Source) > 0 ||
		len(ee.ReplacementLocalFilesystem) > 0 ||
		len(ee.ReplacementMissingCommit) > 0
}

func (ee Errors) Error() string {
	var buf bytes.Buffer

	if len(ee.Source) > 0 {
		buf.WriteString("\t\t# Mirrors for the following packages are not currently known, please look them up and handle these tuples manually:\n")
		sort.Slice(ee.Source, func(i, j int) bool {
			return ee.Source[i] < ee.Source[j]
		})
		for _, err := range ee.Source {
			buf.WriteString(fmt.Sprintf("\t\t#\t%s\n", string(err)))
		}
	}

	if len(ee.ReplacementMissingCommit) > 0 {
		buf.WriteString("\t\t# The following replacement packages are missing version/commit ID, you may need to symlink them in post-patch:\n")
		sort.Slice(ee.ReplacementMissingCommit, func(i, j int) bool {
			return ee.ReplacementMissingCommit[i] < ee.ReplacementMissingCommit[j]
		})
		for _, err := range ee.ReplacementMissingCommit {
			buf.WriteString(fmt.Sprintf("\t\t#\t%s\n", string(err)))
		}
	}

	if len(ee.ReplacementLocalFilesystem) > 0 {
		buf.WriteString("\t\t# The following replacement packages are referencing a local filesystem path, you may need to symlink them in post-patch:\n")
		sort.Slice(ee.ReplacementLocalFilesystem, func(i, j int) bool {
			return ee.ReplacementLocalFilesystem[i] < ee.ReplacementLocalFilesystem[j]
		})
		for _, err := range ee.ReplacementLocalFilesystem {
			buf.WriteString(fmt.Sprintf("\t\t#\t%s\n", string(err)))
		}
	}

	return buf.String()
}

type ReplacementLocalFilesystemError string

func (err ReplacementLocalFilesystemError) Error() string {
	return string(err)
}

type ReplacementMissingCommitError string

func (err ReplacementMissingCommitError) Error() string {
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

func Parse(spec, packagePrefix string) (*Tuple, error) {
	const replaceOp = " => "

	// Replaced package spec
	if strings.Contains(spec, replaceOp) {
		replaceSpecs := strings.Split(spec, replaceOp)
		if len(replaceSpecs) != 2 {
			return nil, fmt.Errorf("unexpected number of packages in replace spec: %q", spec)
		}

		var sourcePkg string
		sourceParts := strings.Fields(replaceSpecs[0])
		switch len(sourceParts) {
		case 1, 2:
			sourcePkg = sourceParts[0]
		default:
			return nil, fmt.Errorf("unexpected number of fields in the replace spec source: %q", spec)
		}

		targetParts := strings.Fields(replaceSpecs[1])
		switch len(targetParts) {
		case 1:
			if targetParts[0][0] == '.' || targetParts[0][0] == '/' {
				// Target package spec is local filesystem path
				return nil, ReplacementLocalFilesystemError(spec)
			}
			// Target package spec is missing commit ID/tag
			return nil, ReplacementMissingCommitError(spec)
		case 2:
			// OK
		default:
			return nil, fmt.Errorf("unexpected number of fields in the replace spec target: %q", spec)
		}

		tuple, err := Parse(replaceSpecs[1], packagePrefix)
		if err != nil {
			return nil, err
		}

		// Keep the target package's account/project/tag but set the source package name
		tuple.Package = sourcePkg

		return tuple, nil
	}

	// Regular package spec
	fields := strings.Fields(spec)
	if len(fields) != 2 {
		return nil, fmt.Errorf("unexpected number of fields: %q", spec)
	}

	pkg := fields[0]
	version := fields[1]
	tuple := &Tuple{Package: pkg, Prefix: packagePrefix}

	if !tuple.fromMirror() {
		switch {
		case strings.HasPrefix(pkg, "github.com"):
			if err := tuple.fromGithub(); err != nil {
				return nil, err
			}
		case strings.HasPrefix(pkg, "gitlab.com"):
			if err := tuple.fromGitlab(); err != nil {
				return nil, err
			}
		default:
			tuple.fromVanity()
		}
	}

	switch {
	case tagRx.MatchString(version):
		sm := tagRx.FindAllStringSubmatch(version, -1)
		tuple.Tag = sm[0][1]
	case versionRx.MatchString(version):
		sm := versionRx.FindAllStringSubmatch(version, -1)
		tuple.Tag = sm[0][1]
	default:
		return nil, fmt.Errorf("unexpected version string: %q", version)
	}

	if tuple.Source == nil {
		return nil, SourceError(tuple.String())
	}

	return tuple, nil
}

var groupRe = regexp.MustCompile(`[^\w]+`)

func (t *Tuple) setSource(source Source, account, project string) {
	t.Source = source
	t.Account = account
	t.Project = project
	group := account + "_" + project
	group = groupRe.ReplaceAllString(group, "_")
	t.Group = strings.ToLower(group)
}

func (t *Tuple) fromGithub() error {
	parts := strings.Split(t.Package, "/")
	if len(parts) < 3 {
		return fmt.Errorf("unexpected Github package name: %q", t.Package)
	}
	t.setSource(GH{}, parts[1], parts[2])
	return nil
}

func (t *Tuple) fromGitlab() error {
	parts := strings.Split(t.Package, "/")
	if len(parts) < 3 {
		return fmt.Errorf("unexpected Gitlab package name: %q", t.Package)
	}
	t.setSource(GL{}, parts[1], parts[2])
	return nil
}
