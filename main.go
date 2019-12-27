package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/dmgk/modules2tuple/spack"
	"github.com/dmgk/modules2tuple/tuple"
)

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

	parser := tuple.NewParser(flagPackagePrefix, flagOffline)
	tuples, errors := parser.Load(args[0])
	if len(tuples) != 0 {
		if !flagSpack {
			fmt.Println(tuples)
		} else {
			resources, err := spack.ResourcesFromTuples(flagAppVersion, tuples)
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				os.Exit(1)
			}
			json, err := resources.ToJson()
			if err != nil {
				fmt.Fprintf(os.Stderr, err.Error())
				os.Exit(1)
			}
			fmt.Println(string(json))
		}
	}
	if errors != nil {
		fmt.Println(errors)
	}
}

var helpTemplate = template.Must(template.New("help").Parse(`
Vendor package dependencies and then run {{.Name}} on vendor/modules.txt:

  $ go mod vendor
  $ {{.Name}} vendor/modules.txt

By default, generated GH_TUPLE entries will place packages under "vendor".
This can be changed by passing different prefix using -prefix option (e.g.
-prefix src).

When generating GL_TUPLE entries, modules2tuple will attempt to use Gitlab
API to resolve short commit IDs and tags to the full 40-character IDs as
required by bsd.sites.mk. If network access is not available or not wanted,
this commit ID translation can be disabled with -offline flag.
`))

var (
	flagOffline       = false
	flagSpack         = false
	flagPackagePrefix = "vendor"
	flagAppVersion    = ""
	flagVersion       = false
)

var version = "devel"

func init() {
	basename := path.Base(os.Args[0])

	flag.BoolVar(&flagOffline, "offline", flagOffline, "disable network access")
	flag.BoolVar(&flagSpack, "spack", flagSpack, "print Spack resource stanzas")
	flag.StringVar(&flagPackagePrefix, "prefix", "vendor", "package prefix")
	flag.StringVar(&flagAppVersion, "app_version", "", "application version (for spack)")
	flag.BoolVar(&flagVersion, "v", flagVersion, "show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] modules.txt\n", basename)
		flag.PrintDefaults()
		helpTemplate.Execute(os.Stderr, map[string]string{
			"Name": basename,
		})
	}
}
