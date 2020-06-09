package configs

import (
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/utils"
)

func GetWorkflowConfiguration(configuration *models.Configuration) *models.WorkflowConfig {
	var workflowConfig models.WorkflowConfig

	switch *configuration.WorkflowType {
	case "gitflow":
		workflowConfig = *GetGitflowConfig(configuration)
	default:
		workflowConfig = *GetGitflowConfig(configuration)
	}

	return &workflowConfig
}

func GetGitflowConfig(configuration *models.Configuration) *models.WorkflowConfig {

	var masterRequirements models.Requirements
	var masterWorkflowRequiredStatusChecks models.RequiredStatusChecks
	var defaultBranch = "develop"

	//Branch Master

	masterWorkflowRequiredStatusChecks.IncludeAdmins = true
	masterWorkflowRequiredStatusChecks.Strict = true
	masterWorkflowRequiredStatusChecks.Contexts = GetRequiredStatusCheck(configuration)

	masterRequirements.EnforceAdmins = true
	masterRequirements.AcceptPrFrom = []string{"release/", "hotfix/"}
	masterRequirements.RequiredStatusChecks = masterWorkflowRequiredStatusChecks
	masterRequirements.ProtectAtStartup = true

	masterBranchConfig := models.Branch{
		Requirements: masterRequirements,
		Stable:       true,
		Name:         utils.Stringify("master"),
		Releasable:   true,
		StartWith:    false,
	}

	//Develop Branch

	var developRequirements models.Requirements
	var developWorkflowRequiredStatusChecks models.RequiredStatusChecks

	developWorkflowRequiredStatusChecks.IncludeAdmins = true
	developWorkflowRequiredStatusChecks.Strict = true
	developWorkflowRequiredStatusChecks.Contexts = GetRequiredStatusCheck(configuration)

	developRequirements.EnforceAdmins = true
	developRequirements.AcceptPrFrom = []string{"feature/", "fix/", "enhancement/", "bugfix/"}
	developRequirements.RequiredStatusChecks = developWorkflowRequiredStatusChecks
	developRequirements.ProtectAtStartup = true

	developBranchConfig := models.Branch{
		Requirements: developRequirements,
		Stable:       true,
		Name:         utils.Stringify("develop"),
		Releasable:   false,
		StartWith:    false,
	}

	//Release Branch

	var releaseRequirements models.Requirements
	var releaseWorkflowRequiredStatusChecks models.RequiredStatusChecks

	releaseWorkflowRequiredStatusChecks.IncludeAdmins = true
	releaseWorkflowRequiredStatusChecks.Strict = true
	releaseWorkflowRequiredStatusChecks.Contexts = GetRequiredStatusCheck(configuration)

	releaseRequirements.EnforceAdmins = true
	releaseRequirements.AcceptPrFrom = []string{"hotfix/"}
	releaseRequirements.RequiredStatusChecks = releaseWorkflowRequiredStatusChecks
	releaseRequirements.ProtectAtStartup = false

	releaseBranchConfig := models.Branch{
		Requirements: releaseRequirements,
		Stable:       true,
		Name:         utils.Stringify("release/"),
		Releasable:   false,
		StartWith:    true,
	}

	//Build the gitflow configuration

	gfConfig := models.WorkflowConfig{
		Name:          utils.Stringify("gitflow"),
		DefaultBranch: utils.Stringify(defaultBranch),
		Description: models.Description{
			Branches: []models.Branch{
				masterBranchConfig,
				developBranchConfig,
				releaseBranchConfig,
			},
		},
		Detail: utils.Stringify("Workflow Description"),
	}

	return &gfConfig
}

//GetRequiredStatusCheck maps the RepositoryStatusChecks field in the Configuration struct into a string slice.
func GetRequiredStatusCheck(c *models.Configuration) []string {
	var rsc []string
	for _, rc := range c.RepositoryStatusChecks {
		rsc = append(rsc, rc.Check)
	}
	return rsc
}
