package flags

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"path"
)

const (
	GithubCredentialsKey = "M2T_GITHUB"
	// GitlabsCredentialsKey = "M2T_GITLAB"
	LookupGithubTagsKey = "M2T_GHTAGS"
	OfflineKey          = "M2T_OFFLINE"
	PrefixKey           = "M2T_PREFIX"
)

var (
	LookupGithubTags = os.Getenv(LookupGithubTagsKey) == "1"
	Offline          = os.Getenv(OfflineKey) == "1"
	PackagePrefix    = os.Getenv(PrefixKey)
	ShowVersion      = false
)

var helpTemplate = template.Must(template.New("help").Parse(`
Vendor package dependencies and then run {{.Name}} on vendor/modules.txt:

  $ go mod vendor
  $ {{.Name}} vendor/modules.txt

By default, generated GH_TUPLE entries will place package directories under
"vendor". This can be changed by passing different prefix using -prefix option
(e.g. -prefix src).

When generating GL_TUPLE entries, modules2tuple will attempt to use Gitlab
API to resolve short commit IDs and tags to the full 40-character IDs as
required by bsd.sites.mk. If network access is not available or not wanted,
this commit ID translation can be disabled with -offline flag.
`))

func init() {
	basename := path.Base(os.Args[0])
	if PackagePrefix == "" {
		PackagePrefix = "vendor"
	}

	flag.BoolVar(&LookupGithubTags, "ghtags", LookupGithubTags, fmt.Sprintf("lookup tags with Github API (env %s)", LookupGithubTagsKey))
	flag.BoolVar(&Offline, "offline", Offline, fmt.Sprintf("disable all network access (env %s)", OfflineKey))
	flag.StringVar(&PackagePrefix, "prefix", PackagePrefix, fmt.Sprintf("package dir prefix (env %s)", PrefixKey))
	flag.BoolVar(&ShowVersion, "v", false, "show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] modules.txt\n", basename)
		flag.PrintDefaults()
		_ = helpTemplate.Execute(os.Stderr, map[string]string{
			"Name": basename,
		})
	}
}
