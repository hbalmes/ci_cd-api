package configs

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetGithubBaseURL(t *testing.T) {

	os.Setenv("DBHOST", "host")
	os.Setenv("DBPASS", "pass")
	os.Setenv("DBNAME", "name")
	os.Setenv("DBUSER", "user")

	type args struct {
		scope string
	}

	type expects struct {
		url string
	}

	tests := []struct {
		name    string
		args    args
		expects expects
	}{
		{
			name: "test scope productive",
			args: args{
				scope: "production",
			},
			expects: expects{url: "https://api.github.com"},
		},
		{
			name: "test scope beta",
			args: args{
				scope: "test",
			},
			expects: expects{url: "http://test.rp-ci-proxy.melifrontends.com"},
		},
		{
			name: "test scope stage",
			args: args{
				scope: "stage",
			},
			expects: expects{url: "https://api.github.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("SCOPE")
			os.Setenv("SCOPE", tt.args.scope)
			got := GetGithubBaseURL()
			assert.NotNil(t, got)
			assert.Equal(t, got, tt.expects.url)
		})
	}
}
