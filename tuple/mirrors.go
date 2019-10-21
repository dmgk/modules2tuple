package tuple

var mirrors = map[string]struct {
	source  Source
	account string
	project string
}{
	"camlistore.org":                            {source: GH{}, account: "perkeep", project: "perkeep"},
	"cloud.google.com/go":                       {source: GH{}, account: "googleapis", project: "google-cloud-go"},
	"contrib.go.opencensus.io/exporter/ocagent": {source: GH{}, account: "census-ecosystem", project: "opencensus-go-exporter-ocagent"},
	"docker.io/go-docker":                       {source: GH{}, account: "docker", project: "go-docker"},
	"git.apache.org/thrift.git":                 {source: GH{}, account: "apache", project: "thrift"},
	"go.opencensus.io":                          {source: GH{}, account: "census-instrumentation", project: "opencensus-go"},
	"go.elastic.co/apm":                         {source: GH{}, account: "elastic", project: "apm-agent-go"},
	"go.elastic.co/fastjson":                    {source: GH{}, account: "elastic", project: "go-fastjson"},
	"go4.org":                                   {source: GH{}, account: "go4org", project: "go4"},
	"gocloud.dev":                               {source: GH{}, account: "google", project: "go-cloud"},
	"google.golang.org/api":                     {source: GH{}, account: "googleapis", project: "google-api-go-client"},
	"google.golang.org/appengine":               {source: GH{}, account: "golang", project: "appengine"},
	"google.golang.org/genproto":                {source: GH{}, account: "google", project: "go-genproto"},
	"google.golang.org/grpc":                    {source: GH{}, account: "grpc", project: "grpc-go"},
	"gopkg.in/fsnotify.v1":                      {source: GH{}, account: "fsnotify", project: "fsnotify"}, // fsnotify is a special case in gopkg.in
	"sigs.k8s.io/yaml":                          {source: GH{}, account: "kubernetes-sigs", project: "yaml"},
	"go.mongodb.org/mongo-driver":               {source: GH{}, account: "mongodb", project: "mongo-go-driver"},
	"gotest.tools":                              {source: GH{}, account: "gotestyourself", project: "gotest.tools"},
	"howett.net/plist":                          {source: GL{"https://gitlab.howett.net"}, account: "go", project: "plist"},
}

func (t *Tuple) fromMirror() bool {
	if m, ok := mirrors[t.Package]; ok {
		t.setSource(m.source, m.account, m.project)
		return true
	}
	return false
}
