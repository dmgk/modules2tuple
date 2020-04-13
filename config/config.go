package config

import (
	"os"
	"strings"
)

const (
	GithubCredentialsKey = "M2T_GITHUB"
	OfflineKey           = "M2T_OFFLINE"
	DebugKey             = "M2T_DEBUG"
)

var (
	GithubToken    string
	GithubUsername string
	Offline        bool
	Debug          bool
	ShowVersion    bool
)

func init() {
	Offline = os.Getenv(OfflineKey) != ""
	Debug = os.Getenv(DebugKey) != ""

	githubCredentials := os.Getenv(GithubCredentialsKey)
	if githubCredentials != "" {
		parts := strings.Split(githubCredentials, ":")
		if len(parts) == 2 {
			GithubUsername = parts[0]
			GithubToken = parts[1]
		}
	}
}
