package spack

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"text/template"

	"github.com/dmgk/modules2tuple/tuple"
)

type resource struct {
	AppVersion   string
	Name         string
	RepoURL      string
	Fetcher      string
	CommitIdType string
	CommitId     string
	Placement    string
}

func Resources(appVersion string, tt tuple.Tuples) (string, error) {
	buf := bytes.Buffer{}
	for _, t := range tt {
		r, err := Resource(appVersion, t)
		if err != nil {
			return "", err
		}
		buf.WriteString(r)
	}
	return buf.String(), nil
}

func Resource(appVersion string, t *tuple.Tuple) (string, error) {
	var repoSite string
	var repoURL string
	var fetcher = "git"

	st := reflect.TypeOf(t.Source)
	switch st.String() {
	case "tuple.GH":
		repoSite = "https://github.com" // tuple.GH.Site() returns ""...
		repoURL = fmt.Sprintf("%s/%s/%s", repoSite, t.Account, t.Project)
	case "tuple.GL":
		repoSite = t.Source.Site()
		repoURL = fmt.Sprintf("%s/%s/%s", repoSite, t.Account, t.Project)
	default:
		return "", fmt.Errorf("Unknown site type: %s", st.String())
	}

	commitIdType := "tag"
	commitId := t.Tag
	matched, err := regexp.MatchString("[0-9a-f]{12}", t.Tag)
	if err != nil {
		return "", err
	}
	if matched {
		commitIdType = "commit"
		// commitId = commitId[:7]
	}

	var buf bytes.Buffer
	r := resource{
		AppVersion:   appVersion,
		Name:         t.Package,
		RepoURL:      repoURL,
		Fetcher:      fetcher,
		CommitIdType: commitIdType,
		CommitId:     commitId,
		Placement:    fmt.Sprintf("%s/%s", t.Prefix, t.Package),
	}
	resource_template := `
    resource(name="{{.Name}}",
             {{.Fetcher}}="{{.RepoURL}}",
             {{.CommitIdType}}="{{.CommitId}}",
             destination=".",{{if .AppVersion}}
             when="@{{.AppVersion}}",{{end}}
             placement="{{.Placement}}")`
	templ := template.Must(template.New("resource").Parse(resource_template))
	err = templ.Execute(&buf, r)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
