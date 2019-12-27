package spack

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"

	"github.com/dmgk/modules2tuple/tuple"
)

type Resource struct {
	Name        string `json:"name,omitempty"`
	Git         string `json:"git,omitempty"`
	Tag         string `json:"tag,omitempty"`
	Commit      string `json:"commit,omitempty"`
	Placement   string `json:"placement,omitempty"`
	When        string `json:"when,omitempty"`
	Destination string `json:"destination,omitempty"`
}

type Resources []*Resource

func (r Resources) ToJson() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r Resource) ToJson() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func ResourcesFromTuples(appVersion string, tt tuple.Tuples) (Resources, error) {
	resources := make(Resources, 0, len(tt))

	for _, t := range tt {
		r, err := resourceFromTuple(appVersion, t)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func resourceFromTuple(appVersion string, t *tuple.Tuple) (*Resource, error) {
	r := Resource{
		Name:        t.Package,
		Destination: ".",
		Placement:   fmt.Sprintf("%s/%s", t.Prefix, t.Package),
	}

	st := reflect.TypeOf(t.Source)
	switch st.String() {
	case "tuple.GH":
		repoSite := "https://github.com" // tuple.GH.Site() returns ""...
		r.Git = fmt.Sprintf("%s/%s/%s", repoSite, t.Account, t.Project)
	case "tuple.GL":
		repoSite := t.Source.Site()
		r.Git = fmt.Sprintf("%s/%s/%s", repoSite, t.Account, t.Project)
	default:
		return nil, fmt.Errorf("Unknown site type: %s", st.String())
	}

	matched, err := regexp.MatchString("[0-9a-f]{12}", t.Tag)
	if err != nil {
		return nil, err
	}
	if matched {
		r.Commit = t.Tag
	} else {
		r.Tag = t.Tag
	}

	if appVersion != "" {
		r.When = "@" + appVersion
	}

	return &r, nil
}
