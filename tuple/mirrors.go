package tuple

var mirrors = map[string]struct {
	account string
	project string
}{
	// Package name                              GH Account, GH Project
	"camlistore.org":                            {"perkeep", "perkeep"},
	"cloud.google.com/go":                       {"googleapis", "google-cloud-go"},
	"contrib.go.opencensus.io/exporter/ocagent": {"census-ecosystem", "opencensus-go-exporter-ocagent"},
	"docker.io/go-docker":                       {"docker", "go-docker"},
	"git.apache.org/thrift.git":                 {"apache", "thrift"},
	"go.opencensus.io":                          {"census-instrumentation", "opencensus-go"},
	"go4.org":                                   {"go4org", "go4"},
	"gocloud.dev":                               {"google", "go-cloud"},
	"google.golang.org/api":                     {"googleapis", "google-api-go-client"},
	"google.golang.org/appengine":               {"golang", "appengine"},
	"google.golang.org/genproto":                {"google", "go-genproto"},
	"google.golang.org/grpc":                    {"grpc", "grpc-go"},
	"gopkg.in/fsnotify.v1":                      {"fsnotify", "fsnotify"}, // fsnotify is a special case in gopkg.in
	"sigs.k8s.io/yaml":                          {"kubernetes-sigs", "yaml"},
	"go.mongodb.org/mongo-driver":               {"mongodb", "mongo-go-driver"},
	"gotest.tools":                              {"gotestyourself", "gotest.tools"},
}

func (t *Tuple) fromMirror() bool {
	if m, ok := mirrors[t.Package]; ok {
		t.setSource(SourceGithub, m.account, m.project)
		return true
	}
	return false
}
