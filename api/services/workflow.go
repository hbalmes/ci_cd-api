package services

import (
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"strings"
)
import "github.com/hbalmes/ci_cd-api/api/configs"

//ConfigurationService is an interface which represents the ConfigurationService for testing purpose.
type WorkflowService interface {
	SetWorkflow(config *models.Configuration) apierrors.ApiError
	UnsetWorkflow(config *models.Configuration) apierrors.ApiError
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
func (c *Configuration) SetWorkflow(config *models.Configuration) apierrors.ApiError {

	//Get the selected workflow configuration
	wfc := configs.GetWorkflowConfiguration(config)

	workflowBranchesList := wfc.Description.Branches

	//Protect stable branches configured on the workflow
	for _, branch := range workflowBranchesList {
		if branch.Stable && branch.Requirements.ProtectAtStartup {
			//Protect the branch
			bpError := c.GithubClient.ProtectBranch(config, &branch)

			if bpError != nil {
				//the branch does not exist. We will create it.
				if bpError.Message() == "branch not found" {

					createBranchErr := c.GithubClient.CreateGithubRef(config, &branch, wfc)

					if createBranchErr != nil {
						return createBranchErr
					}
					//Adds to list the same branch to re-execute it
					workflowBranchesList = append(workflowBranchesList, branch)
					continue
					//break
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
//This controls that the head branch and the base branch combination, follow the configured workflow.
//Returns the body of the status webhook to be sent to the pull request
func (c *Configuration) CheckWorkflow(config *models.Configuration, prWebhook *webhook.PullRequestWebhook) *webhook.Status {

	var stWebhook webhook.Status
	workflowOk := false
	var isAllowedPullRequestBaseBranch bool
	var baseBranchConfig models.Branch

	//Get the selected workflow configuration
	wfc := configs.GetWorkflowConfiguration(config)

	workflowBranchesList := wfc.Description.Branches

	//Check if the base branch are in the stable branch list
	for _, branch := range workflowBranchesList {
		if branch.Name == prWebhook.PullRequest.Base.Ref || strings.HasPrefix(*prWebhook.PullRequest.Base.Ref, *branch.Name) {
			isAllowedPullRequestBaseBranch = true
			baseBranchConfig = branch
			break
		} else {
			isAllowedPullRequestBaseBranch = false
			continue
		}
	}

	if isAllowedPullRequestBaseBranch {
		//Check if base branch accepts PR from the head branch in the configured workflow
		for _, acceptedBranch := range baseBranchConfig.Requirements.AcceptPrFrom {
			if strings.HasPrefix(*prWebhook.PullRequest.Head.Ref, acceptedBranch) {
				workflowOk = true
				continue
			}
		}
	} else {
		workflowOk = true
	}

	if workflowOk {
		stWebhook.State = utils.Stringify(statusWebhookSuccessState)
		stWebhook.Description = utils.Stringify(statusWebhookSuccessDescription)
	} else {
		stWebhook.State = utils.Stringify(statusWebhookFailureState)
		stWebhook.Description = utils.Stringify(statusWebhookFailureDescription)
	}

	stWebhook.Repository.FullName = prWebhook.Repository.FullName
	stWebhook.Context = utils.Stringify("workflow")
	stWebhook.TargetURL = utils.Stringify(statusWebhookTargetURL)
	stWebhook.Sha = prWebhook.PullRequest.Head.Sha

	return &stWebhook
}

//UnsetWorkflow unprotect the necessary branches for the workflow configured
func (c *Configuration) UnsetWorkflow(config *models.Configuration) apierrors.ApiError {

	//Get the selected workflow configuration
	wfc := configs.GetWorkflowConfiguration(config)

	workflowBranchesList := wfc.Description.Branches

	//Unprotect stable branches configured on the workflow
	for _, branch := range workflowBranchesList {
		if branch.Stable && branch.Requirements.ProtectAtStartup {
			//delete branch protection
			bpError := c.GithubClient.UnprotectBranch(config, &branch)

			if bpError != nil {
				return bpError
			}
		}
	}

	return nil
}
