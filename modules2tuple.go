package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Package struct {
	Name    string // full package name
	Account string // Guthub account
	Project string // Guthub project
	Tag     string // Tag or commit ID
}

var versionRx = regexp.MustCompile(`\A(v\d+\.\d+\.\d+(-[0-9A-Za-z]+[0-9A-Za-z\.-]+)?)(\+incompatible)?\z`)
var tagRx = regexp.MustCompile(`\Av\d+\.\d+\.\d+-\d{14}-([0-9a-f]{7})[0-9a-f]+\z`)

func ParsePackage(spec string) (*Package, error) {
	fields := strings.Fields(spec)
	if len(fields) != 3 {
		return nil, fmt.Errorf("unexpected number of fileds: %q", spec)
	}

	name := fields[1]
	version := fields[2]

	p := &Package{Name: name}

	// Parse package name
	if strings.HasPrefix(name, "github.com") {
		nameParts := strings.Split(name, "/")
		if len(nameParts) < 3 {
			return nil, fmt.Errorf("unexpected Github package name: %q", name)
		}
		p.Account = nameParts[1]
		p.Project = nameParts[2]
	} else if wk, ok := wellKnownPackages[name]; ok {
		p.Account = wk.Account
		p.Project = wk.Project
	}

	// Parse version
	if tagRx.MatchString(version) {
		sm := tagRx.FindAllStringSubmatch(version, -1)
		p.Tag = sm[0][1]
	} else if versionRx.MatchString(version) {
		sm := versionRx.FindAllStringSubmatch(version, -1)
		p.Tag = sm[0][1]
	} else {
		return nil, fmt.Errorf("unexpected version string: %q", version)
	}

	return p, nil
}

func (p *Package) Parsed() bool {
	return p.Account != "" && p.Project != ""
}

func (p *Package) Group() string {
	g := p.Account + "_" + p.Project
	g = strings.Replace(g, "-", "_", -1)
	return strings.ToLower(g)
}

func (p *Package) String() string {
	return fmt.Sprintf("%s:%s:%s:%s/%s/%s", p.Account, p.Project, p.Tag, p.Group(), prefix, p.Name)
}

type WellKnown struct {
	Account string // Github account
	Project string // Github project
}

// List of well-known Github mirrors
var wellKnownPackages = map[string]WellKnown{
	"golang.org/x/crypto":         {"golang", "crypto"},
	"golang.org/x/net":            {"golang", "net"},
	"golang.org/x/sync":           {"golang", "sync"},
	"golang.org/x/sys":            {"golang", "sys"},
	"golang.org/x/text":           {"golang", "text"},
	"golang.org/x/tools":          {"golang", "tools"},
	"google.golang.org/appengine": {"golang", "appengine"},
	"gopkg.in/yaml.v2":            {"go-yaml", "yaml"},
}

var prefix string

func main() {
	flag.Parse()

	file := os.Stdin

	args := flag.Args()
	if len(args) > 0 {
		var err error
		file, err = os.Open(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	var parsedPackages []*Package
	var unparsedPackages []*Package

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			pkg, err := ParsePackage(line)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if pkg.Parsed() {
				parsedPackages = append(parsedPackages, pkg)
			} else {
				unparsedPackages = append(unparsedPackages, pkg)
			}
		}
	}

	fmt.Println("GH_TUPLE=\t\\")
	for i, p := range parsedPackages {
		fmt.Printf("\t\t%s", p)
		if i < len(parsedPackages)-1 {
			fmt.Print(" \\")
		}
		fmt.Println("")
	}

	for _, p := range unparsedPackages {
		fmt.Printf("#\t\t%s\n", p)
	}
}

func init() {
	flag.StringVar(&prefix, "prefix", "src", "package prefix")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] [modules.txt]\n", os.Args[0])
		flag.PrintDefaults()
	}
}
