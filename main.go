package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/dmgk/modules2tuple/config"
	"github.com/dmgk/modules2tuple/parser"
)

var version = "devel"

func main() {
	flag.Parse()

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
