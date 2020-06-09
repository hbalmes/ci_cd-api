package configs

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetDBConnectionParams(t *testing.T) {

	os.Setenv("DBHOST", "host")
	os.Setenv("DBPASS", "pass")
	os.Setenv("DBNAME", "name")
	os.Setenv("DBUSER", "user")

	type args struct {
		scope string
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "test scope productive",
			args: args{
				scope: "production",
			},
		},
		{
			name: "test scope beta",
			args: args{
				scope: "beta",
			},
		},
		{
			name: "test scope stage",
			args: args{
				scope: "stage",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("SCOPE")
			os.Setenv("SCOPE", tt.args.scope)
			got := GetDBConnectionParams()
			assert.NotNil(t, got)
		})
	}
}
