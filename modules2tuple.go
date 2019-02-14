package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type Package struct {
	Name    string // full package name
	Account string // Github account
	Project string // Github project
	Tag     string // tag or commit ID
}

// v1.0.0
// v1.0.0+incompatible
// v1.2.3-pre-release-suffix
// v1.2.3-pre-release-suffix+incompatible
var versionRx = regexp.MustCompile(`\A(v\d+\.\d+\.\d+(?:-[0-9A-Za-z]+[0-9A-Za-z\.-]+)?)(?:\+incompatible)?\z`)

// v0.0.0-20181001143604-e0a95dfd547c
// v1.2.3-20181001143604-e0a95dfd547c
// v1.2.3-3.20181001143604-e0a95dfd547c
var tagRx = regexp.MustCompile(`\Av\d+\.\d+\.\d+-(?:\d+\.)?\d{14}-([0-9a-f]{7})[0-9a-f]+\z`)

func ParsePackage(spec string) (*Package, error) {
	const replaceOp = " => "

	if strings.Contains(spec, replaceOp) {
		// Replaced package spec

		pkgs := strings.Split(spec, replaceOp)
		if len(pkgs) != 2 {
			return nil, fmt.Errorf("unexpected number of packages in replace spec: %q", spec)
		}
		pOld, err := ParsePackage(pkgs[0])
		if err != nil {
			return nil, err
		}
		p, err := ParsePackage(pkgs[1])
		if err != nil {
			return nil, err
		}

		// Keep the old package Name but with new Account, Project and Tag
		p.Name = pOld.Name

		return p, nil
	} else {
		// Regular package spec

		fields := strings.Fields(spec)
		if len(fields) != 2 {
			return nil, fmt.Errorf("unexpected number of fileds: %q", spec)
		}

		name := fields[0]
		version := fields[1]

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
	return fmt.Sprintf("%s:%s:%s:%s/%s/%s", p.Account, p.Project, p.Tag, p.Group(), packagePrefix, p.Name)
}

type PackagesByAccountAndProject []*Package

func (pp PackagesByAccountAndProject) Len() int {
	return len(pp)
}

func (pp PackagesByAccountAndProject) Swap(i, j int) {
	pp[i], pp[j] = pp[j], pp[i]
}

func (pp PackagesByAccountAndProject) Less(i, j int) bool {
	return pp[i].Account+"/"+pp[i].Project < pp[j].Account+"/"+pp[j].Project
}

type WellKnown struct {
	Account string // Github account
	Project string // Github project
}

// List of well-known Github mirrors
var wellKnownPackages = map[string]WellKnown{
	// Package name                          GH Account, GH Project
	"golang.org/x/crypto":                    {"golang", "crypto"},
	"golang.org/x/net":                       {"golang", "net"},
	"golang.org/x/oauth2":                    {"golang", "oauth2"},
	"golang.org/x/sync":                      {"golang", "sync"},
	"golang.org/x/sys":                       {"golang", "sys"},
	"golang.org/x/text":                      {"golang", "text"},
	"golang.org/x/tools":                     {"golang", "tools"},
	"google.golang.org/appengine":            {"golang", "appengine"},
	"gopkg.in/alexcesaro/quotedprintable.v3": {"alexcesaro", "quotedprintable"},
	"gopkg.in/yaml.v2":                       {"go-yaml", "yaml"},
}

var packagePrefix string

func main() {
	flag.Parse()
	args := flag.Args()
	file := os.Stdin

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
	const specPrefix = "# "

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, specPrefix) {
			pkg, err := ParsePackage(strings.TrimPrefix(line, specPrefix))
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

	sort.Sort(PackagesByAccountAndProject(parsedPackages))

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
	flag.StringVar(&packagePrefix, "prefix", "src", "package prefix")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] [modules.txt]\n", os.Args[0])
		flag.PrintDefaults()
	}
}
