package configs

import "os"

const (
	githubProductionBaseURL  = "https://api.github.com"
)

const (
	githubTestBaseURL  = "http://test.rp-ci-proxy.melifrontends.com"
)

const (
	githubLocalBaseURL  = "http://localhost:8888"
)

func GetGithubBaseURL() string {
	switch scope := os.Getenv("SCOPE"); scope {
	case "production":
		return githubProductionBaseURL
	case "test":
		return githubTestBaseURL
	default:
		return githubProductionBaseURL
	}
}