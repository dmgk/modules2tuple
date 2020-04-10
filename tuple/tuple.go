package tuple

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dmgk/modules2tuple/apis"
	"github.com/dmgk/modules2tuple/config"
)

// Parse parses a package spec into Tuple.
func Parse(spec string) (*Tuple, error) {
	const replaceSep = " => "

	if strings.Contains(spec, replaceSep) {
		// "replace" spec
		parts := strings.Split(spec, replaceSep)
		if len(parts) != 2 {
			return nil, fmt.Errorf("unexpected replace spec format: %q", spec)
		}

		leftPkg, leftVersion, err := parseSpec(parts[0])
		if err != nil {
			return nil, err
		}

		rightPkg, rightVersion, err := parseSpec(parts[1])
		if err != nil {
			return nil, err
		}

		// https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive
		if isFilesystemPath(rightPkg) {
			// get the left spec package and symlink it to the rightPkg path
			return Resolve(leftPkg, leftVersion, leftPkg, rightPkg)
		}
		// get the right spec package and put it under leftPkg path
		return Resolve(rightPkg, rightVersion, leftPkg, "")
	}

	// regular spec
	pkg, version, err := parseSpec(spec)
	if err != nil {
		return nil, err
	}
	return Resolve(pkg, version, pkg, "")
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

func parseSpec(spec string) (string, string, error) {
	parts := strings.Fields(spec)

	switch len(parts) {
	case 1:
		// must be a versionless local filesystem "replace" spec rhs
		if isFilesystemPath(parts[0]) {
			return parts[0], "", nil
		}
		return "", "", fmt.Errorf("unexpected spec format: %q", spec)
	case 2:
		// regular spec
		if tagRx.MatchString(parts[1]) {
			sm := tagRx.FindAllStringSubmatch(parts[1], -1)
			return parts[0], sm[0][1], nil
		}
		if versionRx.MatchString(parts[1]) {
			sm := versionRx.FindAllStringSubmatch(parts[1], -1)
			return parts[0], sm[0][1], nil
		}
		return "", "", fmt.Errorf("unexpected version string in spec: %q", spec)
	default:
		return "", "", fmt.Errorf("unexpected number of fields in spec: %q", spec)
	}
}

func isFilesystemPath(s string) bool {
	return s[0] == '.' || s[0] == '/'
}

type Tuple struct {
	Package   string // Go package name
	Version   string // tag or commit ID
	Subdir    string // GH_TUPLE subdir
	Link      string // path to symlink to
	Source    Source // tuple source
	Account   string // account
	Project   string // project
	Group     string // GH_TUPLE group
	Submodule string // submodule suffix if present

	// LinkSuffix string
}

var underscoreRe = regexp.MustCompile(`[^\w]+`)

func (t *Tuple) resolve(source Source, account, project, submodule string) {
	t.Source = source
	t.Account = account
	t.Project = project
	t.Submodule = submodule

	var submoduleBase string
	if t.Submodule != "" {
		submoduleBase = filepath.Base(t.Submodule)
	}
	// if t.Link != "" {
	//     linkBase = filepath.Base(t.Link)
	// }

	group := t.Account + "_" + t.Project
	if submoduleBase != "" {
		group = group + "_" + submoduleBase
	}
	group = underscoreRe.ReplaceAllString(group, "_")
	group = strings.Trim(group, "_")
	t.Group = strings.ToLower(group)

	// linkSuffix := underscoreRe.ReplaceAllString(linkBase, "_")
	// linkSuffix = strings.Trim(linkSuffix, "_")
	// t.LinkSuffix = strings.ToLower(linkSuffix)
}

func (t *Tuple) isResolved() bool {
	return t.Source != nil
}

func (t *Tuple) isLinked() bool {
	return t.Link != ""
}

func (t *Tuple) Postprocess() error {
	if config.Offline {
		return nil
	}

	// if isFilesystemPath(t.Subdir) {
	//     fmt.Printf("====> fs %s %s %s %s\n", t.Package, t.Subdir, t.Submodule, t.Link)
	// }
	// if t.Link != "" {
	//     fmt.Printf("====> %s %s %s %s\n", t.Package, t.Subdir, t.Submodule, t.Link)
	// }
	// if strings.Contains(t.Package, "azure/cli") {
	//     fmt.Printf("====> before t %#v\n", t)
	// }

	// if strings.Contains(t.Package, "autorest") || strings.Contains(t.Package, "vault") {
	//     fmt.Printf("====> before %s %s %s %s\n", t.Package, t.Subdir, t.Submodule, t.Link)
	// }

	switch t.Source.(type) {
	case GithubSource:
		// If package version is a tag and it's a submodule, call Gihub API to check tags.
		// Go seem to be able to magically translate tags like "v1.0.4" to the "api/v1.0.4",
		// lets try to do the same.
		if strings.HasPrefix(t.Version, "v") && t.Submodule != "" {
			tag, err := apis.GithubLookupTag(t.Account, t.Project, t.Submodule, t.Version)
			if err != nil {
				return err
			}
			t.Version = tag
		}
		// If package is a submodule, adjust GH_SUBDIR
		// NOTE: tag translation has to be done before this
		if t.Submodule != "" {
			hasContentAtSuffix, err := apis.GithubHasContentsAtPath(t.Account, t.Project, t.Submodule, t.Version)
			if err != nil {
				return err
			}
			if hasContentAtSuffix {
				// Trim suffix from GH_TUPLE subdir because repo already has contents and it'll
				// be extracted at the correct path.
				t.Subdir = strings.TrimSuffix(t.Subdir, "/"+t.Submodule)
			}
		}
	case GitlabSource:
		// Call Gitlab API to translate go.mod short commit IDs and tags
		// to the full 40-character commit IDs as required by bsd.sites.mk
		hash, err := apis.GitlabGetCommit(t.Source.String(), t.Account, t.Project, t.Version)
		if err != nil {
			return err
		}
		t.Version = hash
	}

	// if t.Link != "" {
	//     fmt.Printf("====> %s %s %s %s\n", t.Package, t.Subdir, t.Submodule, t.Link)
	// }
	// if strings.Contains(t.Package, "azure/cli") {
	//     fmt.Printf("====> after t %#v\n", t)
	// }

	// if strings.Contains(t.Package, "autorest") || strings.Contains(t.Package, "vault") {
	//     fmt.Printf("====> after %s subdir:%s submodule:%s link:%s\n", t.Package, t.Subdir, t.Submodule, t.Link)
	// }

	return nil
}

func (t *Tuple) subdirPath() string {
	if t.Subdir == "" {
		return ""
	}
	if t.isLinked() {
		return t.Subdir
	}
	return filepath.Join("vendor", t.Subdir)
}

// func (t *Tuple) linkPath() string {
//     if t.Link == "" {
//         return ""
//     }
//     fmt.Printf("====> link1 %#v\n", t.Link)
//     if isFilesystemPath(t.Link) {
//         fmt.Printf("====> link2 %#v\n", t.Link)
//         // return filepath.Join(t.Link, t.Submodule)
//         return t.Link
//     }
//     return filepath.Join("vendor", t.Link)
// }

func (t *Tuple) doLink() {
	if t.Link == "" {
		t.Link = filepath.Join("vendor", t.Subdir, t.Submodule)
	}
	// t.Subdir = t.Group
	t.Subdir = ""
}

// func (t *Tuple) LinkSuffix() string {
//     var linkBase string
//     if t.Link != "" {
//         linkBase = filepath.Base(t.Link)
//     }
//     res := underscoreRe.ReplaceAllString(linkBase, "_")
//     res = strings.Trim(res, "_")
//     return strings.ToLower(res)
// }

func (t *Tuple) String() string {
	var res string
	if t.Source != nil && t.Source.String() != "" {
		res = t.Source.String() + ":"
	}
	// res = fmt.Sprintf("%s%s:%s:%s:%s/%s", res, t.Account, t.Project, t.Version, t.Group, t.Subdir)
	res = fmt.Sprintf("%s%s:%s:%s:%s", res, t.Account, t.Project, t.Version, t.Group)
	if t.subdirPath() != "" {
		res = fmt.Sprintf("%s/%s", res, t.subdirPath())
	}
	if t.Link != "" {
		// res = fmt.Sprintf("%s ==> %s", res, t.linkPath())
		res = fmt.Sprintf("%s ==> %s", res, t.Link)
	}
	return res
}

func (t *Tuple) sortKey() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s", t.Source, t.Account, t.Project, t.Version, t.Submodule, t.Group, t.Link)
}

type Slice []*Tuple

func (s Slice) Postprocess() error {
	if len(s) < 2 {
		return nil
	}

	ensureUniqueGroups(s)
	if err := ensureUniqueGithubProjectAndTag(s); err != nil {
		return err
	}
	ensureUniqueSubdirs(s)

	return nil
}

// ensureUniqueGroups makes sure there are no duplicate group names.
// This function assumes that s is pre-sorted in sortKey() order.
func ensureUniqueGroups(s Slice) {
	var prevGroup string
	suffix := 1

	key := func(i int) string {
		return s[i].Group
	}
	sort.Slice(s, func(i, j int) bool {
		return key(i) < key(j)
	})

	for _, t := range s {
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

type DuplicateProjectAndTag string

func (err DuplicateProjectAndTag) Error() string {
	return string(err)
}

// ensureUniqueGithubProjectAndTag checks that tuples have a unique GH_PROJECT/GH_TAGNAME
// combination. Due to the way Github prepares release tarballs and the way port framework
// works, tuples sharing GH_PROJECT/GH_TAGNAME pair will be extracted into the same directory.
// Try avoiding this mess by switching one of the conflicting tuple's GH_TAGNAME from git tag
// to git commit ID.
// This function assumes that s is pre-sorted in sortKey() order.
func ensureUniqueGithubProjectAndTag(s Slice) error {
	if config.Offline {
		return nil
	}

	key := func(i int) string {
		return fmt.Sprintf("%T:%s:%s:%s", s[i].Source, s[i].Account, s[i].Project, s[i].Version)
	}
	sort.Slice(s, func(i, j int) bool {
		return key(i) < key(j)
	})

	var prevTuple *Tuple

	for _, t := range s {
		if t.Source != GH {
			continue // not a Github tuple, skip
		}

		if prevTuple == nil {
			prevTuple = t
			continue
		}

		if t.Account != prevTuple.Account {
			// different Account, same Project and Version
			if t.Project == prevTuple.Project && t.Version == prevTuple.Version {
				hash, err := apis.GithubGetCommit(t.Account, t.Project, t.Version)
				if err != nil {
					return DuplicateProjectAndTag(t.String())
				}
				if len(hash) < 12 {
					return errors.New("unexpectedly short Githib commit hash")
				}
				t.Version = hash[:12]
			}
		}
	}

	return nil
}

func ensureUniqueSubdirs(s Slice) error {
	key := func(i int) string {
		return s[i].Subdir
	}
	sort.Slice(s, func(i, j int) bool {
		return key(i) < key(j)
	})

	var (
		prevSubdir string
		prevTuple  *Tuple
	)

	for _, t := range s {
		if prevSubdir == "" {
			prevSubdir = t.Subdir
			prevTuple = t
			continue
		}

		if t.Subdir == prevSubdir {
			if prevTuple != nil {
				// prevTuple.Link = filepath.Join(prevTuple.Subdir, prevTuple.Submodule)
				// prevTuple.Subdir = prevTuple.Group // prevTuple.Subdir + "_" + prevTuple.LinkSuffix()
				prevTuple.doLink()
				prevTuple = nil
			}
			// t.Link = filepath.Join(t.Subdir, t.Submodule)
			// t.Subdir = t.Group // t.Subdir + "_" + t.LinkSuffix()
			t.doLink()
		} else {
			prevSubdir = t.Subdir
			prevTuple = t
		}
	}

	return nil
}

// If tuple slice contains more than largeLimit entries, start tuple list on the new line for easier sorting/editing.
// Otherwise omit the first line continuation for more compact representation.
const largeLimit = 3

func (s Slice) String() string {
	sort.Slice(s, func(i, j int) bool {
		return s[i].sortKey() < s[j].sortKey()
	})

	mapBySource := make(map[Source][]string)
	for _, t := range s {
		mapBySource[t.Source] = append(mapBySource[t.Source], t.String())
	}

	var lines []string
	for s, tt := range mapBySource {
		buf := bytes.NewBufferString(fmt.Sprintf("%s=\t", sourceVarName(s)))
		large := len(tt) > largeLimit
		if large {
			buf.WriteString("\\\n")
		}
		for i := 0; i < len(tt); i += 1 {
			if i > 0 || large {
				buf.WriteString("\t\t")
			}
			buf.WriteString(tt[i])
			if i < len(tt)-1 {
				buf.WriteString(" \\\n")
			}
		}
		lines = append(lines, buf.String())
	}

	return strings.Join(lines, "\n\n")
}

func (s Slice) PostExtract() string {
	var links Slice
	for _, t := range s {
		if t.isLinked() {
			links = append(links, t)
		}
	}
	if len(links) == 0 {
		return ""
	}

	key := func(i int) string {
		return links[i].Link
	}
	sort.Slice(links, func(i, j int) bool {
		return key(i) < key(j)
	})

	var lines []string
	dirs := map[string]struct{}{}

	for _, l := range links {
		var b bytes.Buffer

		dir := filepath.Dir(l.Link)
		if _, ok := dirs[dir]; !ok {
			b.WriteString(fmt.Sprintf("\t@${MKDIR} %s\n", filepath.Join("${WRKSRC}", dir)))
			dirs[dir] = struct{}{}
		}
		b.WriteString(fmt.Sprintf("\t@${RLN} ${WRKSRC_%s} %s", l.Group, filepath.Join("${WRKSRC}", l.Link)))
		dirs[l.Link] = struct{}{}

		lines = append(lines, b.String())
	}

	var b bytes.Buffer
	b.WriteString("post-extract:\n")
	b.WriteString(strings.Join(lines, "\n"))

	return b.String()
}
