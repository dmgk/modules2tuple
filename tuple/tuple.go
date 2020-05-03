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
	"github.com/dmgk/modules2tuple/debug"
)

// Parse parses a package spec into Tuple.
func Parse(spec string) (*Tuple, error) {
	const replaceSep = " => "

	// "replace" spec
	if strings.Contains(spec, replaceSep) {
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
		return parts[0], "", nil
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
	pkg      string // Go package name
	version  string // tag or commit ID
	subdir   string // GH_TUPLE subdir
	group    string // GH_TUPLE group
	module   string // module, if any
	link_src *Tuple // symlink source tuple, if any
	link_tgt string // symlink target, if any
	source   Source // tuple source (Github ot Gitlab)
	account  string // source account
	project  string // source project
	hidden   bool   // if true, tuple will be excluded from G{H,L}_TUPLE
}

var underscoreRe = regexp.MustCompile(`[^\w]+`)

func (t *Tuple) makeResolved(source Source, account, project, module string) {
	t.source = source
	t.account = account
	t.project = project
	t.module = module

	var moduleBase string
	if t.module != "" {
		moduleBase = filepath.Base(t.module)
	}

	group := t.account + "_" + t.project
	if moduleBase != "" {
		group = group + "_" + moduleBase
	}
	group = underscoreRe.ReplaceAllString(group, "_")
	group = strings.Trim(group, "_")
	t.group = strings.ToLower(group)
}

func (t *Tuple) isResolved() bool {
	return t.source != nil
}

func (t *Tuple) Fix() error {
	if config.Offline {
		return nil
	}

	switch t.source.(type) {
	case GithubSource:
		// If package version is a tag and it's a module in a multi-module repo, call Gihub API
		// to check tags. Go seem to be able to magically translate tags like "v1.0.4" to the
		// "api/v1.0.4", lets try to do the same.
		if strings.HasPrefix(t.version, "v") && t.module != "" {
			tag, err := apis.GithubLookupTag(t.account, t.project, t.module, t.version)
			if err != nil {
				return err
			}
			if t.version != tag {
				debug.Printf("[Tuple.Fix] translated Github tag %q to %q\n", t.version, tag)
				t.version = tag
			}
		}
		// If package is a module in a multi-module repo, adjust GH_SUBDIR
		// NOTE: tag translation has to be done before this
		if t.module != "" {
			hasContentAtSuffix, err := apis.GithubHasContentsAtPath(t.account, t.project, t.module, t.version)
			if err != nil {
				return err
			}
			if hasContentAtSuffix {
				// Trim suffix from GH_TUPLE subdir because repo already has contents and it'll
				// be extracted at the correct path.
				debug.Printf("[Tuple.Fix] trimmed module suffix %q from %q\n", t.module, t.subdir)
				t.subdir = strings.TrimSuffix(t.subdir, "/"+t.module)
			}
		}
		// Ports framework doesn't understand tags that have more than 2 path separators in it,
		// replace by commit ID
		if len(strings.Split(t.version, "/")) > 2 {
			hash, err := apis.GithubGetCommit(t.account, t.project, t.version)
			if err != nil {
				return err
			}
			if len(hash) < 12 {
				return errors.New("unexpectedly short Githib commit hash")
			}
			debug.Printf("[Tuple.Fix] translated Github tag %q to %q\n", t.version, hash[:12])
			t.version = hash[:12]
		}
	case GitlabSource:
		// Call Gitlab API to translate go.mod short commit IDs and tags
		// to the full 40-character commit IDs as required by bsd.sites.mk
		hash, err := apis.GitlabGetCommit(t.source.String(), t.account, t.project, t.version)
		if err != nil {
			return err
		}
		debug.Printf("[Tuple.Fix] translated Gitlab tag %q to %q\n", t.version, hash)
		t.version = hash
	}

	return nil
}

func (t *Tuple) subdirPath() string {
	if t.subdir == "" {
		return ""
	}
	if t.isLinked() {
		return t.subdir
	}
	return filepath.Join("vendor", t.subdir)
}

func (t *Tuple) isLinked() bool {
	return t.link_tgt != ""
}

func (t *Tuple) makeLinked() {
	if !t.isLinked() {
		t.link_tgt = filepath.Join("vendor", t.subdir, t.module)
	}
	t.subdir = ""
}

func (t *Tuple) makeLinkedAs(src *Tuple) {
	if !t.isLinked() {
		t.link_src = src
		t.link_tgt = filepath.Join("vendor", t.subdir, t.module)
	}
	t.subdir = ""
}

func (t *Tuple) String() string {
	var res string

	if t.source != nil && t.source.String() != "" {
		res = t.source.String() + ":"
	}
	res = fmt.Sprintf("%s%s:%s:%s:%s", res, t.account, t.project, t.version, t.group)
	if t.subdirPath() != "" {
		res = fmt.Sprintf("%s/%s", res, t.subdirPath())
	}
	// if t.isLinked() {
	//     res = fmt.Sprintf("%s ==> %s", res, t.Link)
	// }

	return res
}

func (t *Tuple) defaultSortKey() string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s", t.source, t.account, t.project, t.version, t.module, t.group, t.link_tgt)
}

type Slice []*Tuple

func (s Slice) Fix() error {
	if len(s) < 2 {
		return nil
	}

	fixGroups(s)
	if err := fixGithubProjectsAndTags(s); err != nil {
		return err
	}
	fixSubdirs(s)
	fixFsNotify(s)

	return nil
}

// fixGroups makes sure there are no duplicate group names.
func fixGroups(s Slice) {
	var prevGroup string
	suffix := 1

	var maxGroup, maxPkg int
	for _, t := range s {
		if len(t.group) > maxGroup {
			maxGroup = len(t.group)
		}
		if len(t.pkg) > maxPkg {
			maxPkg = len(t.pkg)
		}
	}

	key := func(i int) string {
		return fmt.Sprintf("%*s %*s", -(maxGroup + 1), s[i].group, -(maxPkg + 1), s[i].pkg)
	}
	sort.Slice(s, func(i, j int) bool {
		return key(i) < key(j)
	})

	if config.Debug {
		debug.Print("[fixGroups] looking at slice:\n")
		for i := range s {
			debug.Printf("[fixGroups]      %s\n", key(i))
		}
	}

	for _, t := range s {
		if prevGroup == "" {
			prevGroup = t.group
			continue
		}
		if t.group == prevGroup {
			oldGroup := t.group
			t.group = fmt.Sprintf("%s_%d", t.group, suffix)
			debug.Printf("[fixGroups] deduped group %q as %q\n", oldGroup, t.group)
			suffix++
		} else {
			prevGroup = t.group
			suffix = 1
		}
	}
}

type DuplicateProjectAndTag string

func (err DuplicateProjectAndTag) Error() string {
	return string(err)
}

// fixGithubProjectsAndTags checks that tuples have a unique GH_PROJECT/GH_TAGNAME
// combination. Due to the way Github prepares release tarballs and the way port framework
// works, tuples sharing GH_PROJECT/GH_TAGNAME pair will be extracted into the same directory.
// Try avoiding this mess by switching one of the conflicting tuple's GH_TAGNAME from git tag
// to git commit ID.
func fixGithubProjectsAndTags(s Slice) error {
	if config.Offline {
		return nil
	}

	key := func(i int) string {
		return fmt.Sprintf("%T:%s:%s:%s", s[i].source, s[i].account, s[i].project, s[i].version)
	}
	sort.Slice(s, func(i, j int) bool {
		return key(i) < key(j)
	})

	var prevTuple *Tuple

	for _, t := range s {
		if t.source != GH {
			continue // not a Github tuple, skip
		}

		if prevTuple == nil {
			prevTuple = t
			continue
		}

		if t.account != prevTuple.account {
			// different Account, same Project and Version
			if t.project == prevTuple.project && t.version == prevTuple.version {
				hash, err := apis.GithubGetCommit(t.account, t.project, t.version)
				if err != nil {
					return DuplicateProjectAndTag(t.String())
				}
				if len(hash) < 12 {
					return errors.New("unexpectedly short Githib commit hash")
				}
				debug.Printf("[fixGithubProjectsAndTags] translated Github tag %q to %q\n", t.version, hash[:12])
				t.version = hash[:12]
			}
		}
	}

	return nil
}

// fixSubdirs ensures that all subdirs are unique and makes symlinks as needed.
func fixSubdirs(s Slice) {
	var maxSubdir, maxVersion, maxModule int
	for _, t := range s {
		if len(t.subdir) > maxSubdir {
			maxSubdir = len(t.subdir)
		}
		if len(t.version) > maxVersion {
			maxVersion = len(t.version)
		}
		if len(t.module) > maxModule {
			maxModule = len(t.module)
		}
	}

	key := func(i int) string {
		return fmt.Sprintf("%*s %*s %*s", -(maxSubdir + 1), s[i].subdir, maxVersion+1, s[i].version, -(maxModule + 1), s[i].module)
	}
	sort.Slice(s, func(i, j int) bool {
		return key(i) < key(j)
	})

	if config.Debug {
		debug.Print("[fixSubdirs] looking at slice:\n")
		for i := range s {
			debug.Printf("[fixSubdirs]     %s\n", key(i))
		}
	}

	var (
		prevSubdir, prevVersion       string
		currentSubdir, currentVersion string
	)

	for _, t := range s {
		if prevSubdir == "" {
			prevSubdir, prevVersion = t.subdir, t.version
			continue
		}

		currentSubdir, currentVersion = t.subdir, t.version
		if prevSubdir == t.subdir {
			if t.version != prevVersion {
				debug.Printf("[fixSubdirs] linking %s/%s@%s (parent %s@%s)\n", t.pkg, t.module, t.version, prevSubdir, prevVersion)
				t.makeLinked()
			} else {
				debug.Printf("[fixSubdirs] hiding %s/%s@%s (parent %s@%s)\n", t.pkg, t.module, t.version, prevSubdir, prevVersion)
				t.hidden = true
			}
		} else {
			prevSubdir = t.subdir
		}
		prevSubdir, prevVersion = currentSubdir, currentVersion
	}
}

// fixFsNotify takes care of fsnotify annoying ability to appear under multiple different import paths.
// github.com/fsnotify/fsnotify is the canonical package name.
func fixFsNotify(s Slice) {
	key := func(i int) string {
		return fmt.Sprintf("%s:%s:%s:%s", s[i].account, s[i].project, s[i].version, s[i].group)
	}
	sort.Slice(s, func(i, j int) bool {
		return key(i) < key(j)
	})

	const (
		fsnotifyAccount = "fsnotify"
		fsnotifyProject = "fsnotify"
	)
	var fsnotifyTuple *Tuple

	for _, t := range s {
		if t.account == fsnotifyAccount && t.project == fsnotifyProject && fsnotifyTuple == nil {
			fsnotifyTuple = t
			continue
		}
		if fsnotifyTuple != nil {
			if t.version == fsnotifyTuple.version {
				t.makeLinkedAs(fsnotifyTuple)
				t.hidden = true
				debug.Printf("[fixFsNotify] linking fnotify %s@%s => %s@%s\n", fsnotifyTuple.pkg, fsnotifyTuple.version, t.pkg, t.version)
			}
		}
	}
}

// If tuple slice contains more than largeLimit entries, start tuple list on the new line for easier sorting/editing.
// Otherwise omit the first line continuation for more compact representation.
const largeLimit = 3

// String returns G{H,L}_TUPLE variables contents.
func (s Slice) String() string {
	sort.Slice(s, func(i, j int) bool {
		return s[i].defaultSortKey() < s[j].defaultSortKey()
	})

	var githubTuples, gitlabTuples Slice
	for _, t := range s {
		switch t.source.(type) {
		case GithubSource:
			githubTuples = append(githubTuples, t)
		case GitlabSource:
			gitlabTuples = append(gitlabTuples, t)
		default:
			panic("unknown source type")
		}
	}

	var lines []string
	for _, tt := range []Slice{githubTuples, gitlabTuples} {
		if len(tt) == 0 {
			continue
		}
		buf := bytes.NewBufferString(fmt.Sprintf("%s=\t", sourceVarName(tt[0].source)))
		large := len(tt) > largeLimit
		if large {
			buf.WriteString("\\\n")
		}
		for i := 0; i < len(tt); i += 1 {
			if tt[i].hidden {
				continue
			}
			if i > 0 || large {
				buf.WriteString("\t\t")
			}
			buf.WriteString(tt[i].String())
			if i < len(tt)-1 {
				buf.WriteString(" \\\n")
			}
		}
		lines = append(lines, buf.String())
	}

	return strings.Join(lines, "\n\n")
}

type Links []*Tuple

// Links returns a slice of tuples that require symlinking.
func (s Slice) Links() Links {
	var res Links

	for _, t := range s {
		if t.isLinked() {
			res = append(res, t)
		}
	}
	key := func(i int) string {
		return res[i].link_tgt
	}
	sort.Slice(res, func(i, j int) bool {
		return key(i) < key(j)
	})

	return res
}

// String returns "post-extract" target contents.
func (l Links) String() string {
	if len(l) == 0 {
		return ""
	}

	var lines []string
	dirs := map[string]struct{}{}

	for _, t := range l {
		var b bytes.Buffer

		var src string
		var need_target_mkdir bool

		if t.link_src != nil {
			src = filepath.Join(fmt.Sprintf("${WRKSRC_%s}", t.link_src.group), t.link_src.module)
			// symlinking other package under different name, target dir is not guaranteed to exist
			need_target_mkdir = true
		} else {
			src = filepath.Join(fmt.Sprintf("${WRKSRC_%s}", t.group), t.module)
			// symlinking module under another module, target dir already exists
			need_target_mkdir = false
		}
		tgt := filepath.Join("${WRKSRC}", t.link_tgt)

		if need_target_mkdir {
			dir := filepath.Dir(t.link_tgt)
			if dir != "" && dir != "." {
				if _, ok := dirs[dir]; !ok {
					b.WriteString(fmt.Sprintf("\t@${MKDIR} %s\n", filepath.Join("${WRKSRC}", dir)))
					dirs[dir] = struct{}{}
				}
			}
		}

		// symlinking over another module, rm -rf first
		if t.module != "" {
			debug.Printf("[Links.String] rm %s\n", tgt)
			b.WriteString(fmt.Sprintf("\t@${RM} -r %s\n", tgt))
		}

		debug.Printf("[Links.String] ln %s => %s\n", src, tgt)
		b.WriteString(fmt.Sprintf("\t@${RLN} %s %s", src, tgt))

		lines = append(lines, b.String())
	}

	var b bytes.Buffer
	b.WriteString("post-extract:\n")
	b.WriteString(strings.Join(lines, "\n"))

	return b.String()
}
