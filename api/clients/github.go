package clients

// Builder client, connects configurations API with the CI proxy API
// and implements the necessary functions to
// create and delete jobs necessary for the execution of release process

import (
	"encoding/json"
	"fmt"
	"github.com/hbalmes/ci_cd-api/api/configs"
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"github.com/mercadolibre/golang-restclient/rest"
	"net/http"
	"time"
)

type GithubClient interface {
	GetBranchInformation(config *models.Configuration, branchName string) (*models.GetBranchResponse, apierrors.ApiError)
	CreateGithubRef(config *models.Configuration, branchConfig *models.Branch, workflowConfig *models.WorkflowConfig) apierrors.ApiError
	ProtectBranch(config *models.Configuration, branchConfig *models.Branch) apierrors.ApiError
	SetDefaultBranch(config *models.Configuration, workflowConfig *models.WorkflowConfig) apierrors.ApiError
	CreateStatus(config *models.Configuration, statusWH *webhook.Status) apierrors.ApiError
	CreateBranch(config *models.Configuration, branchConfig *models.Branch, sha string) apierrors.ApiError
}

type githubClient struct {
	Client Client
}

func NewGithubClient() GithubClient {
	hs := make(http.Header)
	hs.Set("cache-control", "no-cache")
	hs.Set("Authorization", "token 79476df2e0c834810c237b3bda8e78ebc0bc7bca")
	hs.Set("Accept", "application/vnd.github.luke-cage-preview+json")

	return &githubClient{
		Client: &client{
			RestClient: &rest.RequestBuilder{
				BaseURL:        configs.GetGithubBaseURL(),
				Timeout:        2 * time.Second,
				Headers:        hs,
				ContentType:    rest.JSON,
				DisableCache:   true,
				DisableTimeout: false,
			},
		},
	}
}

type ghGetBranchResponse struct {
	Message  string `json:"message"`
	URL      string `json:"url"`
	Provider string `json:"provider"`
}

//Gets a repository branch info
//This perform a GET request to Github api using
func (c *githubClient) GetBranchInformation(config *models.Configuration, branchName string) (*models.GetBranchResponse, apierrors.ApiError) {

	if config.RepositoryName == nil || config.RepositoryOwner == nil || branchName == "" {
		return nil, apierrors.NewBadRequestApiError("invalid github body params")
	}

	response := c.Client.Get(fmt.Sprintf("/repos/%s/%s/branches/%s", *config.RepositoryOwner, *config.RepositoryName, branchName))

	if response.Err() != nil {
		return nil, apierrors.NewInternalServerApiError("Something went wrong getting branch information", response.Err())
	}

	if response.StatusCode() != http.StatusOK {
		return nil, apierrors.NewInternalServerApiError("error getting repository branch", response.Err())
	}

	var branchInfo models.GetBranchResponse
	if err := json.Unmarshal(response.Bytes(), &branchInfo); err != nil {
		return nil, apierrors.NewBadRequestApiError("error binding github branch response")
	}

	return &branchInfo, nil
}

//Protects the branch from pushs by following the workflow configuration
//This perform a PUT request to Github api
func (c *githubClient) ProtectBranch(config *models.Configuration, branchConfig *models.Branch) apierrors.ApiError {

	if branchConfig.Name == nil {
		return apierrors.NewBadRequestApiError("invalid branch protection body params")
	}

	body := map[string]interface{}{
		"enforce_admins":                true,
		"required_status_checks":        branchConfig.Requirements.RequiredStatusChecks,
		"required_pull_request_reviews": branchConfig.Requirements.RequiredPullRequestReviews,
		"restrictions":                  nil,
	}

	response := c.Client.Put(fmt.Sprintf("/repos/%s/%s/branches/%s/protection", *config.RepositoryOwner, *config.RepositoryName, *branchConfig.Name), body)

	if response.Err() != nil {
		return apierrors.NewInternalServerApiError("Something went wrong protecting branch", response.Err())
	}

	if response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusCreated {
		if response.StatusCode() == http.StatusNotFound {
			return apierrors.NewInternalServerApiError("branch not found", response.Err())
		}
		return apierrors.NewInternalServerApiError(fmt.Sprintf("error protecting branch - status: %d", response.StatusCode()), response.Err())
	}

	return nil
}

//Create a new reference, in this case a branch
//This perform a POST request to Github api
func (c *githubClient) CreateBranch(config *models.Configuration, branchConfig *models.Branch, sha string) apierrors.ApiError {

	if branchConfig.Name == nil || config.RepositoryOwner == nil || config.RepositoryName == nil || sha == "" {
		return apierrors.NewBadRequestApiError("invalid body params")
	}

	ref := utils.Stringify(fmt.Sprintf("refs/heads/%s", *branchConfig.Name))

	body := map[string]interface{}{
		"ref": ref,
		"sha": sha,
	}

	url := fmt.Sprintf("/repos/%s/%s/git/refs", *config.RepositoryOwner, *config.RepositoryName)

	response := c.Client.Post(url, body)

	if response.Err() != nil {
		return apierrors.NewInternalServerApiError("Something went wrong creating a branch", response.Err())
	}

	if response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusCreated {
		return apierrors.NewInternalServerApiError(fmt.Sprintf("error creating a branch - status: %d", response.StatusCode()), response.Err())
	}

	return nil
}

//Create a new reference on github. First we get the information needed to make the creation and then the creation itself.
//This perform a GetBranchInformation and CreateBranch
func (c *githubClient) CreateGithubRef(config *models.Configuration, branchConfig *models.Branch, workflowConfig *models.WorkflowConfig) apierrors.ApiError {

	if branchConfig.Name == nil || config.RepositoryOwner == nil || config.RepositoryName == nil {
		return apierrors.NewBadRequestApiError("invalid body params")
	}

	//First gets SHA necessary to initialise the new branch or reference
	initialBranch := workflowConfig.DefaultBranch

	if branchConfig.Name == workflowConfig.DefaultBranch {
		initialBranch = utils.Stringify("master")
	}

	branchInfo, getBranchError := c.GetBranchInformation(config, *initialBranch)

	if getBranchError != nil {
		return getBranchError
	}

	createRefErr := c.CreateBranch(config, branchConfig, branchInfo.Commit.Sha)

	if createRefErr != nil {
		return createRefErr
	}

	return nil
}

//SetDefaultBranch updates the default branch of repository.
//This is the branch from which new branches should start
func (c *githubClient) SetDefaultBranch(config *models.Configuration, workflowConfig *models.WorkflowConfig) apierrors.ApiError {

	if config.RepositoryOwner == nil || config.RepositoryName == nil || *workflowConfig.DefaultBranch == "" {
		return apierrors.NewBadRequestApiError("invalid body params")
	}

	body := map[string]interface{}{
		"name":           *config.RepositoryName,
		"default_branch": workflowConfig.DefaultBranch,
	}

	response := c.Client.Post(fmt.Sprintf("/repos/%s/%s", *config.RepositoryOwner, *config.RepositoryName), body)

	if response.Err() != nil {
		return apierrors.NewInternalServerApiError("Something went wrong setting default branch", response.Err())
	}

	if response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusCreated {
		return apierrors.NewInternalServerApiError(fmt.Sprintf("error updating default branch - status: %d", response.StatusCode()), response.Err())
	}

	return nil
}

//CreateStatus create commit statuses for a given SHA.
//This perform a POST request
func (c *githubClient) CreateStatus(config *models.Configuration, statusWH *webhook.Status) apierrors.ApiError {

	if config.RepositoryOwner == nil || config.RepositoryName == nil || statusWH.Sha == nil || statusWH.Context == nil {
		return apierrors.NewBadRequestApiError("invalid body params")
	}

	body := map[string]interface{}{
		"state":       statusWH.State,
		"target_url":  statusWH.TargetURL,
		"description": statusWH.Description,
		"context":     statusWH.Context,
	}

	response := c.Client.Post(fmt.Sprintf("/repos/%s/%s/statuses/%p", *config.RepositoryOwner, *config.RepositoryName, statusWH.Sha), body)

	if response.Err() != nil {
		return apierrors.NewInternalServerApiError("RestClient Error creating new status", response.Err())
	}

	if response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusCreated {
		return apierrors.NewInternalServerApiError("Error creating new status", response.Err())
	}

	return nil
}
