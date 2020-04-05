package services

import (
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"strings"
)
import "github.com/hbalmes/ci_cd-api/api/configs"

//ConfigurationService is an interface which represents the ConfigurationService for testing purpose.
type WorkflowService interface {
	SetWorkflow(config *models.WorkflowConfig) error
	CheckWorkflow(config *models.Configuration, prWebhook *webhook.PullRequestWebhook) *webhook.Status
}

const (
	statusWebhookTargetURL          = "http://www.url_de_wiki"
	statusWebhookSuccessState       = "success"
	statusWebhookSuccessDescription = "Great! You comply with the workflow"
	statusWebhookFailureState       = "error"
	statusWebhookFailureDescription = "Oops! You're not complying with the workflow."
)

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
				} else {
					return bpError
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

//CheckWorkflow check the workflow
//This controls that the head branch and the base branch convination, follow the configured workflow.
func (c *Configuration) CheckWorkflow(config *models.Configuration, prWebhook *webhook.PullRequestWebhook) *webhook.Status {

	var stWebhook webhook.Status
	workflowOk := false

	//Get the selected workflow configuration
	wfc := configs.GetWorkflowConfiguration(config)

	workflowBranchesList := wfc.Description.Branches

	//Check if the base branch are in the stable branch list
	for _, branch := range workflowBranchesList {
		if branch.Stable {
			if branch.Name == prWebhook.PullRequest.Base.Ref {

				//Check if base branch accepts PR from the head branch in the configured workflow
				for _, acceptedBranch := range branch.Requirements.AcceptPrFrom {
					if strings.HasPrefix(prWebhook.PullRequest.Head.Ref, acceptedBranch) {
						workflowOk = true
						continue
					}
				}
			} else if branch.StartWith { //check if the base branch start with the stable branch name i.e release/
				if strings.HasPrefix(prWebhook.PullRequest.Base.Ref, branch.Name) {
					//Check if base branch accepts PR from the head branch in the configured workflow
					acceptedPullRequest := utils.StringContains(branch.Requirements.AcceptPrFrom, prWebhook.PullRequest.Head.Ref)

					if acceptedPullRequest {
						workflowOk = true
						continue
					}
				}
			}
		} else {
			workflowOk = true
		}
	}

	if workflowOk {
		stWebhook.State = statusWebhookSuccessState
		stWebhook.Description = statusWebhookSuccessDescription
	} else {
		stWebhook.State = statusWebhookFailureState
		stWebhook.Description = statusWebhookFailureDescription
	}

	stWebhook.Context = "workflow"
	stWebhook.TargetURL = statusWebhookTargetURL
	stWebhook.Sha = prWebhook.PullRequest.Head.Sha

	return &stWebhook
}
