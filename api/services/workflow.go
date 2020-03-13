package services

import (
	"github.com/herbal828/ci_cd-api/api/models"
)
import "github.com/herbal828/ci_cd-api/api/configs"

//ConfigurationService is an interface which represents the ConfigurationService for testing purpose.
type WorkflowService interface {
	SetWorkflow(config *models.WorkflowConfig) error
}

//SetWorkflow protects the necessary branches for the workflow selected by the user
//It performs all the actions needed to enabled successfuly Release Process.
func (c *Configuration) SetWorkflow(config *models.Configuration) error {

	//Get the selected workflow configuration
	wfc := configs.GetWorkflowConfiguration(config)

	workflowBranchesList := wfc.Description.Branches

	//Protect stable branches configured on the workflow
	for _, branch := range workflowBranchesList {
		if branch.Stable {
			//Protect the branch
			bpError := c.GithubClient.ProtectBranch(config, &branch)

			if bpError != nil {
				//the branch does not exist. We will create it.
				if bpError.Error() == "branch not found" {

					createBranchErr := c.GithubClient.CreateGithubRef(config, &branch, wfc)

					if createBranchErr != nil {
						return createBranchErr
					}
					//Adds to list the same branch to re-execute it
					workflowBranchesList = append(workflowBranchesList, branch)
					break
				}
			}
		}
	}

	//Update the default branch
	setDefaultBranchErr := c.GithubClient.SetDefaultBranch(config, wfc)

	if setDefaultBranchErr != nil {
		return setDefaultBranchErr
	}

	return nil
}
