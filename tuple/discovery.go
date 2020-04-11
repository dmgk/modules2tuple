package tuple

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

func discoverMirrors(pkg string) (string, error) {
	u, err := pkgURL(pkg)
	if err != nil {
		return "", err
	}
	resp, err := http.Get(u)
	if err != nil {
		return "", fmt.Errorf("tuple.discoverMirrors %s: %v", u, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return parseMetaGoImports(resp.Body)
	default:
		return "", fmt.Errorf("tuple.discoverMirrors %s: status %d", u, resp.StatusCode)
	}
}

// HTML meta imports discovery code adopted from
// https://github.com/golang/go/blob/master/src/cmd/go/internal/get/discovery.go
func parseMetaGoImports(r io.Reader) (string, error) {
	d := xml.NewDecoder(r)
	d.CharsetReader = charsetReader
	d.Strict = false

	for {
		t, err := d.RawToken()
		if err != nil {
			if err != io.EOF {
				return "", err
			}
			break
		}
		if e, ok := t.(xml.StartElement); ok && strings.EqualFold(e.Name.Local, "body") {
			break
		}
		if e, ok := t.(xml.EndElement); ok && strings.EqualFold(e.Name.Local, "head") {
			break
		}
		e, ok := t.(xml.StartElement)
		if !ok || !strings.EqualFold(e.Name.Local, "meta") {
			continue
		}
		if attrValue(e.Attr, "name") != "go-import" {
			continue
		}
		if f := strings.Fields(attrValue(e.Attr, "content")); len(f) == 3 {
			if f[1] == "git" {
				repo := schemeRe.ReplaceAllString(f[2], "")
				return suffixRe.ReplaceAllString(repo, ""), nil
			}
		}
	}
	return "", nil
}

var (
	schemeRe = regexp.MustCompile(`\Ahttps?://`)
	suffixRe = regexp.MustCompile(`\.git\z`)
)

func pkgURL(pkg string) (string, error) {
	u, err := url.Parse(pkg)
	if err != nil {
		return "", err
	}

	u.Scheme = "https"
	q := u.Query()
	q.Set("go-get", "1")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "utf-8", "ascii":
		return input, nil
	default:
		return nil, fmt.Errorf("can't decode XML document using charset %q", charset)
	}
}

func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}
