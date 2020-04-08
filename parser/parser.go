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

	"github.com/dmgk/modules2tuple/tuple"
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
					err = t.Postprocess()
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

	res.Postprocess()

	// if !config.Offline {
	//     if err := tuples.EnsureUniqueGithubProjectAndTag(); err != nil {
	//         switch err := err.(type) {
	//         case tuple.DuplicateProjectAndTag:
	//             errors.DuplicateProjectAndTag = append(errors.DuplicateProjectAndTag, err)
	//         default:
	//             return nil, err
	//         }
	//     }
	// }

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

func (r *Result) Postprocess() {
	if err := r.tuples.Postprocess(); err != nil {
		r.AddError(err)
	}
}

type errSlice []error

func (ee errSlice) String() string {
	var ss []string
	for _, e := range ee {
		ss = append(ss, fmt.Sprintf("\t\t#\t%s", e))
	}
	return strings.Join(ss, "\n")
}

func (r *Result) String() string {
	var ss []string

	if len(r.tuples) > 0 {
		var b bytes.Buffer
		b.WriteString(r.tuples.String())
		ss = append(ss, b.String())
	}

	if len(r.errSource) > 0 {
		var b bytes.Buffer
		b.WriteString("\t\t# Mirrors for the following packages are not currently known, please look them up and handle these tuples manually:\n")
		sort.Slice(r.errSource, func(i, j int) bool {
			return r.errSource[i].Error() < r.errSource[j].Error()
		})
		b.WriteString(errSlice(r.errSource).String())
		ss = append(ss, b.String())
	}

	if len(r.errOther) > 0 {
		var b bytes.Buffer
		b.WriteString("\t\t# Other errors found during processing:\n")
		b.WriteString(errSlice(r.errOther).String())
		ss = append(ss, b.String())
	}

	return strings.Join(ss, "\n\n")
}
