package tuple

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dmgk/modules2tuple/apis"
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

	for _, t := range tt {
		if prevGroup == "" {
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

// EnsureUniqueGithubProjectAndTag checks that tuples have a unique GH_PROJECT/GH_TAGNAME
// combination. Due to the way Github prepares release tarballs and th way ports framework
// works, tuples sharing GH_PROJECT/GH_TAGNAME pair will be extracted to the same directory.
// Try avoiding this mess by switching one of the conflicting tuple's GH_TAGNAME from git tag
// to git commit ID.
// This function assumes that tt is pre-sorted in ByAccountAndProject order.
func (tt Tuples) EnsureUniqueGithubProjectAndTag() error {
	if len(tt) < 2 {
		return nil
	}

	var prevTuple *Tuple

	for _, t := range tt {
		if _, ok := t.Source.(GH); !ok {
			// not a Github tuple, skip
			continue
		}

		if prevTuple == nil {
			prevTuple = t
			continue
		}

		if t.Account != prevTuple.Account {
			// different Account, but the same Project and Tag
			if t.Project == prevTuple.Project && t.Tag == prevTuple.Tag {
				c, err := apis.GetGithubCommit(t.Account, t.Project, t.Tag)
				if err != nil {
					return DuplicateProjectAndTag(t.String())
				}
				if len(c.ID) < 12 {
					return errors.New("unexpectedly short commit ID")
				}
				t.Tag = c.ID[:12]
			}
		}
	}

	return nil
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

// If tuple contains more than largeLimit entries, start tuple list on the new line for easier sorting/editing.
// Otherwise omit the first line continuation for more compact representation.
const largeLimit = 3

func (tt Tuples) String() string {
	if len(tt) == 0 {
		return ""
	}

	tm := make(map[Source][]string)
	for _, t := range tt {
		tm[t.Source] = append(tm[t.Source], t.String())
	}

	var ss []string
	for s, ee := range tm {
		buf := bytes.NewBufferString(fmt.Sprintf("%s=\t", s.VarName()))
		large := len(ee) > largeLimit
		if large {
			buf.WriteString("\\\n")
		}
		for i := 0; i < len(ee); i += 1 {
			if i > 0 || large {
				buf.WriteString("\t\t")
			}
			buf.WriteString(ee[i])
			if i < len(ee)-1 {
				buf.WriteString(" \\\n")
			}
		}
		ss = append(ss, buf.String())
	}
	sort.Sort(sort.StringSlice(ss))

	return fmt.Sprintf("%s\n", strings.Join(ss, "\n\n"))
}

type Errors struct {
	Source                     []SourceError
	ReplacementLocalFilesystem []ReplacementLocalFilesystemError
	ReplacementMissingCommit   []ReplacementMissingCommitError
	DuplicateProjectAndTag     []DuplicateProjectAndTag
}

func (ee Errors) Any() bool {
	return len(ee.Source) > 0 ||
		len(ee.ReplacementLocalFilesystem) > 0 ||
		len(ee.ReplacementMissingCommit) > 0 ||
		len(ee.DuplicateProjectAndTag) > 0
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

	if len(ee.DuplicateProjectAndTag) > 0 {
		buf.WriteString("\t\t# The following tuple has duplicate GH_PROJECT/GH_TAGNAME combinations and an attempt to fix it using Github API failed:\n")
		sort.Slice(ee.DuplicateProjectAndTag, func(i, j int) bool {
			return ee.DuplicateProjectAndTag[i] < ee.DuplicateProjectAndTag[j]
		})
		for _, err := range ee.DuplicateProjectAndTag {
			buf.WriteString(fmt.Sprintf("\t\t#\t%s\n", string(err)))
		}
	}

	return buf.String()
}

type SourceError string

func (err SourceError) Error() string {
	return string(err)
}

type ReplacementLocalFilesystemError string

func (err ReplacementLocalFilesystemError) Error() string {
	return string(err)
}

type ReplacementMissingCommitError string

func (err ReplacementMissingCommitError) Error() string {
	return string(err)
}

type DuplicateProjectAndTag string

func (err DuplicateProjectAndTag) Error() string {
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
// v3.0.1-0.20190209023717-9147687966d9+incompatible
var tagRx = regexp.MustCompile(`\Av\d+\.\d+\.\d+-(?:[0-9A-Za-z\.]+\.)?\d{14}-([0-9a-f]+)(?:\+incompatible)?\z`)

// Parse parses a package spec into Tuple.
func Parse(spec, packagePrefix string) (*Tuple, error) {
	t, err := parseReplaceSpec(spec, packagePrefix)
	if err != nil {
		return nil, err
	}
	if t != nil {
		// Tuple was parsed from a replace spec
		return t, nil
	}
	return parseSpec(spec, packagePrefix)
}

func parseReplaceSpec(spec, packagePrefix string) (*Tuple, error) {
	const replaceOp = " => "

	if !strings.Contains(spec, replaceOp) {
		// Not a replace spec
		return nil, nil
	}

	replaceSpecs := strings.Split(spec, replaceOp)
	if len(replaceSpecs) != 2 {
		return nil, fmt.Errorf("unexpected number of packages in replace spec: %q", spec)
	}

	var sourcePkg string

	sourceFields := strings.Fields(replaceSpecs[0])
	switch len(sourceFields) {
	case 1, 2:
		sourcePkg = sourceFields[0]
	default:
		return nil, fmt.Errorf("unexpected number of fields in the replace spec source: %q", spec)
	}

	targetFields := strings.Fields(replaceSpecs[1])
	switch len(targetFields) {
	case 1:
		if targetFields[0][0] == '.' || targetFields[0][0] == '/' {
			// Target package spec is a local filesystem path
			return nil, ReplacementLocalFilesystemError(spec)
		}
		// Target package spec is missing commit ID/tag
		return nil, ReplacementMissingCommitError(spec)
	case 2:
		// OK
	default:
		return nil, fmt.Errorf("unexpected number of fields in the replace spec target: %q", spec)
	}

	t, err := parseSpec(replaceSpecs[1], packagePrefix)
	if err != nil {
		return nil, err
	}

	// Keep the target package's account/project/tag but set the source package name
	t.Package = sourcePkg

	return t, nil
}

func parseSpec(spec, packagePrefix string) (*Tuple, error) {
	fields := strings.Fields(spec)
	if len(fields) != 2 {
		return nil, fmt.Errorf("unexpected number of fields: %q", spec)
	}
	pkg := fields[0]
	version := fields[1]

	pkgParsers := []func(string, string) (*Tuple, error){
		tryMirror,
		tryGithub,
		tryGitlab,
		tryVanity,
		noMatch,
	}

	var t *Tuple
	for _, fn := range pkgParsers {
		var err error
		t, err = fn(pkg, packagePrefix)
		if err != nil {
			return nil, err
		}
		if t != nil {
			break
		}
	}

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

	if t.Source == nil {
		return nil, SourceError(t.String())
	}

	return t, nil
}

func tryGithub(pkg, packagePrefix string) (*Tuple, error) {
	if !strings.HasPrefix(pkg, "github.com") {
		return nil, nil
	}
	parts := strings.Split(pkg, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected Github package name: %q", pkg)
	}
	return newTuple(GH{}, pkg, parts[1], parts[2], packagePrefix), nil
}

func tryGitlab(pkg, packagePrefix string) (*Tuple, error) {
	if !strings.HasPrefix(pkg, "gitlab.com") {
		return nil, nil
	}
	parts := strings.Split(pkg, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("unexpected Gitlab package name: %q", pkg)
	}
	return newTuple(GL{}, pkg, parts[1], parts[2], packagePrefix), nil
}

// noMatch returns "unparsed" tuple as a fallback.
func noMatch(pkg, packagePrefix string) (*Tuple, error) {
	return newTuple(nil, pkg, "", "", packagePrefix), nil
}

var groupRe = regexp.MustCompile(`[^\w]+`)

func newTuple(source Source, pkg, account, project, packagePrefix string) *Tuple {
	t := &Tuple{
		Source:  source,
		Package: pkg,
		Account: account,
		Project: project,
		Prefix:  packagePrefix,
	}
	if t.Account != "" && t.Project != "" {
		group := t.Account + "_" + t.Project
		group = groupRe.ReplaceAllString(group, "_")
		t.Group = strings.ToLower(group)
	} else {
		t.Group = "group_name"
	}
	return t
}
