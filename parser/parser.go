package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/dmgk/modules2tuple/v2/tuple"
)

// Load parses tuples from vendor/modules.txt at path.
func Load(path string) (*Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Read(f)
}

// Read parses tuples from modules.txt contents provided as io.Reader.
func Read(r io.Reader) (*Result, error) {
	ch := make(chan interface{})

	go func() {
		defer close(ch)

		const specPrefix = "# "

		scanner := bufio.NewScanner(r)
		sem := make(chan int, runtime.NumCPU())
		// sem := make(chan int, 1)
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
					t, err := tuple.Parse(strings.TrimPrefix(line, specPrefix))
					if err != nil {
						ch <- err
						return
					}
					err = t.Fix()
					if err != nil {
						ch <- err
						return
					}
					ch <- t
				}()
			}
		}
		wg.Wait()
	}()

	res := &Result{}

	for v := range ch {
		if t, ok := v.(*tuple.Tuple); ok {
			res.AddTuple(t)
		} else if err, ok := v.(error); ok {
			res.AddError(err)
		} else {
			panic("unknown value type")
		}
	}

	res.Fix()
	return res, nil
}

type Result struct {
	tuples    tuple.Slice
	errSource []error
	errOther  []error
}

func (r *Result) AddTuple(t *tuple.Tuple) {
	r.tuples = append(r.tuples, t)
}

func (r *Result) AddError(err error) {
	switch err := err.(type) {
	case tuple.SourceError:
		r.errSource = append(r.errSource, err)
	default:
		r.errOther = append(r.errOther, err)
	}
}

func (r *Result) Fix() {
	if err := r.tuples.Fix(); err != nil {
		r.AddError(err)
	}
}

type errSlice []error

func (errs errSlice) String() string {
	var lines []string
	for _, err := range errs {
		lines = append(lines, fmt.Sprintf("\t\t#\t%s", err))
	}
	return strings.Join(lines, "\n")
}

func (r *Result) String() string {
	var lines []string

	if len(r.tuples) > 0 {
		var b bytes.Buffer
		b.WriteString(r.tuples.String())
		lines = append(lines, b.String())
	}

	if len(r.errSource) > 0 {
		var b bytes.Buffer
		b.WriteString("\t\t# Mirrors for the following packages are not currently known, please look them up and handle these tuples manually:\n")
		sort.Slice(r.errSource, func(i, j int) bool {
			return r.errSource[i].Error() < r.errSource[j].Error()
		})
		b.WriteString(errSlice(r.errSource).String())
		lines = append(lines, b.String())
	}

	if len(r.errOther) > 0 {
		var b bytes.Buffer
		b.WriteString("\t\t# Errors found during processing:\n")
		b.WriteString(errSlice(r.errOther).String())
		lines = append(lines, b.String())
	}

	links := r.tuples.Links()
	if len(links) > 0 {
		lines = append(lines, links.String())
	}

	return strings.Join(lines, "\n\n")
}
