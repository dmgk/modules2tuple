package main

import (
	"flag"
	"fmt"
	"html/template"
	"os"
	"path"

	"github.com/dmgk/modules2tuple/config"
	"github.com/dmgk/modules2tuple/parser"
)

var version = "devel"

func main() {
	if config.ShowVersion {
		fmt.Fprintln(os.Stderr, version)
		os.Exit(0)
	}

	args := flag.Args()

	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	res, err := parser.Load(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(res)
}

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
`))

func init() {
	basename := path.Base(os.Args[0])

	flag.BoolVar(&config.Offline, "offline", config.Offline, "")
	flag.BoolVar(&config.Debug, "debug", config.Debug, "")
	flag.BoolVar(&config.ShowVersion, "v", false, "")

	flag.Usage = func() {
		err := usageTemplate.Execute(os.Stderr, map[string]interface{}{
			"basename": basename,
			"offline":  config.Offline,
			"debug":    config.Debug,
		})
		if err != nil {
			panic(err)
		}

	}

	flag.Parse()
}
