// +build online

package parser

import (
	"strings"
	"testing"

	"github.com/dmgk/modules2tuple/config"
)

func init() {
	config.Offline = false
}

func TestUniqueProjectAndTag(t *testing.T) {
	given := `
# github.com/json-iterator/go v1.1.7
# github.com/ugorji/go v1.1.7`

	expected := `GH_TUPLE=	json-iterator:go:v1.1.7:json_iterator_go/vendor/github.com/json-iterator/go \
		ugorji:go:v1.1.7:ugorji_go

post-extract:
	@${MKDIR} ${WRKSRC}/vendor/github.com/ugorji
	@${RLN} ${WRKSRC_json_iterator_go} ${WRKSRC}/vendor/github.com/ugorji/go`

	tt, err := Read(strings.NewReader(given))
	if err != nil {
		t.Fatal(err)
	}
	out := tt.String()
	if out != expected {
		t.Errorf("expected output\n%s\n, got\n%s", expected, out)
	}
}
