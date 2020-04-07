package tuple

import (
	"bufio"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type Parser struct {
	packagePrefix    string
	offline          bool
	lookupGithubTags bool
}

// NewParser creates a new modules.txt parser with given options.
func NewParser(packagePrefix string, offline bool, lookupGithubTags bool) *Parser {
	return &Parser{packagePrefix, offline, lookupGithubTags}
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
					if !p.offline {
						if err = t.PostProcessTag(p.lookupGithubTags); err != nil {
							ch <- err
							return
						}
					}
					t.PostProcessSubdir()
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
				errors.Source = append(errors.Source, err)
			case ReplacementMissingCommitError:
				errors.ReplacementMissingCommit = append(errors.ReplacementMissingCommit, err)
			case ReplacementLocalFilesystemError:
				errors.ReplacementLocalFilesystem = append(errors.ReplacementLocalFilesystem, err)
			default:
				return nil, err
			}
		} else {
			tuples = append(tuples, res.(*Tuple))
		}
	}
	sort.Sort(ByAccountAndProject(tuples))

	tuples.EnsureUniqueGroups()
	if !p.offline {
		if err := tuples.EnsureUniqueGithubProjectAndTag(); err != nil {
			switch err := err.(type) {
			case DuplicateProjectAndTag:
				errors.DuplicateProjectAndTag = append(errors.DuplicateProjectAndTag, err)
			default:
				return nil, err
			}
		}
	}

	if errors.Any() {
		return tuples, errors
	}

	return tuples, nil
}

// Load parses tuples from vendor/modules.txt at path.
func (p *Parser) Load(path string) (Tuples, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return p.Read(f)
}
