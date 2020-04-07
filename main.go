package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dmgk/modules2tuple/flags"
	"github.com/dmgk/modules2tuple/tuple"
)

var version = "devel"

func main() {
	flag.Parse()

	if flags.ShowVersion {
		fmt.Fprintln(os.Stderr, version)
		os.Exit(0)
	}

	args := flag.Args()

	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var haveTuples bool
	parser := tuple.NewParser(flags.PackagePrefix, flags.Offline, flags.LookupGithubTags)
	tuples, errors := parser.Load(args[0])
	if len(tuples) != 0 {
		fmt.Print(tuples)
		haveTuples = true
	}
	if errors != nil {
		if haveTuples {
			fmt.Println()
		}
		fmt.Print(errors)
		fmt.Println()
	}
}
