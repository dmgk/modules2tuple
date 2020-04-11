package config

import (
	"flag"
	"html/template"
	"os"
	"path"
	"strings"
)

const (
	GithubCredentialsKey = "M2T_GITHUB"
	OfflineKey           = "M2T_OFFLINE"
	DebugKey             = "M2T_DEBUG"
)

var (
	GithubUsername string
	GithubToken    string
	Offline        = os.Getenv(OfflineKey) != ""
	Debug          = os.Getenv(DebugKey) != ""
	ShowVersion    = false
)

var usageTemplate = template.Must(template.New("Usage").Parse(`usage: {{.basename}} [options] modules.txt

Options:
    -offline  disable all network access (env M2T_OFFLINE, default {{.offline}})
    -debug    print debug info (env M2T_DEBUG, default {{.debug}})
    -v        show version

Usage:
    Vendor package dependencies and then run {{.basename}} with vendor/modules.txt:

    $ go mod vendor
    $ {{.basename}} vendor/modules.txt

When running in offline mode:
    - mirrors are looked up using static list and some may not be resolved
    - milti-module repos and version suffixes ("/v2") are not automatically handled
    - Github tags for modules ("v1.2.3" vs "api/v1.2.3") are not automatically resolved
    - Gitlab commit IDs are not resolved to the full 40-char IDs
    - post-extract target is not generated
`))

func init() {
	basename := path.Base(os.Args[0])

	githubCredentials := os.Getenv(GithubCredentialsKey)
	if githubCredentials != "" {
		parts := strings.Split(githubCredentials, ":")
		if len(parts) == 2 {
			GithubUsername = parts[0]
			GithubToken = parts[1]
		}
	}

	flag.BoolVar(&Offline, "offline", Offline, "")
	flag.BoolVar(&Debug, "debug", Debug, "")
	flag.BoolVar(&ShowVersion, "v", false, "")

	flag.Usage = func() {
		err := usageTemplate.Execute(os.Stderr, map[string]interface{}{
			"basename": basename,
			"offline":  Offline,
			"debug":    Debug,
		})
		if err != nil {
			panic(err)
		}

	}

	// flag.Parse()
}
