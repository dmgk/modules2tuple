package tuple

import (
	"bufio"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/dmgk/modules2tuple/gitlab"
)

type Parser struct {
	packagePrefix string
	offline       bool
}

// NewParser creates a new modules.txt parser with given options.
func NewParser(packagePrefix string, offline bool) *Parser {
	return &Parser{packagePrefix, offline}
}

// Read parses tuples from modules.txt contents provided as io.Reader.
func (p *Parser) Read(r io.Reader) (Tuples, error) {
	ch := make(chan interface{})

	go func() {
		defer close(ch)

		const specPrefix = "# "
		scanner := bufio.NewScanner(r)
		sem := make(chan int, runtime.NumCPU())
		var wg sync.WaitGroup

		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, specPrefix) {
				sem <- 1
				wg.Add(1)
				go func() {
					defer func() {
						<-sem
						wg.Done()
					}()
					t, err := Parse(strings.TrimPrefix(line, specPrefix), p.packagePrefix)
					if err != nil {
						ch <- err
						return
					}
					// Call Gitlab API to translate go.mod short commit IDs and tags
					// to the full 40-character commit IDs as required by bsd.sites.mk
					if !p.offline && t.Source == SourceGitlab {
						c, err := gitlab.GetCommit(t.Account, t.Project, t.Tag)
						if err != nil {
							ch <- err
							return
						}
						t.Tag = c.ID
					}
					ch <- t
				}()
			}
		}
		wg.Wait()
	}()

	var tuples Tuples
	var errors Errors

	for res := range ch {
		if err, ok := res.(error); ok {
			switch err := err.(type) {
			case SourceError:
				errors.SourceErrors = append(errors.SourceErrors, err)
			case ReplacementError:
				errors.ReplacementErrors = append(errors.ReplacementErrors, err)
			default:
				return nil, err
			}
		} else {
			tuples = append(tuples, res.(*Tuple))
		}
	}
	sort.Sort(ByAccountAndProject(tuples))
	tuples.EnsureUniqueGroups()

	if errors.Any() {
		return tuples, errors
	}

	return tuples, nil
}

// Load parses tuples from modules.txt found at path.
func (p *Parser) Load(path string) (Tuples, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return p.Read(f)
}
