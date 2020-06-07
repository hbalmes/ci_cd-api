package configs

import (
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetWorkflowConfiguration(t *testing.T) {
	type args struct {
		configuration *models.Configuration
	}

	type expects struct {
		name string
		defaultBranch string
		detail string
	}

	configWithGitflow := models.Configuration{
		WorkflowType:                     utils.Stringify("gitflow"),
	}

	var wfConfig models.WorkflowConfig
	wfConfig.Name = utils.Stringify("gitflow")
	wfConfig.DefaultBranch = utils.Stringify("develop")

	tests := []struct {
		name string
		args args
		expects expects
	}{
		{
			name: "gitflow wf config getted",
			args: args{configuration: &configWithGitflow},
			expects: expects{
				name:          "gitflow",
				defaultBranch: "develop",
				detail:        "Workflow Description",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetWorkflowConfiguration(tt.args.configuration)
			assert.NotNil(t, got)
			assert.Equal(t, got.DefaultBranch, &tt.expects.defaultBranch)
			assert.Equal(t, got.Name, &tt.expects.name)
			assert.Equal(t, got.Detail, &tt.expects.detail)

		})
	}
}