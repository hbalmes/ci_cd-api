package clients

// Builder client, connects configurations API with the CI proxy API
// and implements the necessary functions to
// create and delete jobs necessary for the execution of release process

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/herbal828/ci_cd-api/api/configs"
	"github.com/herbal828/ci_cd-api/api/models"
	"github.com/mercadolibre/golang-restclient/rest"
	"net/http"
	"time"
)

type GithubClient interface {
	GetBranchInformation(config *models.Configuration, branchName string) (*models.GetBranchResponse, error)
	CreateGithubRef(config *models.Configuration, branchConfig *models.Branch, workflowConfig *models.WorkflowConfig) error
	ProtectBranch(config *models.Configuration, branchConfig *models.Branch) error
	SetDefaultBranch(config *models.Configuration, workflowConfig *models.WorkflowConfig) error
}

type githubClient struct {
	Client Client
}

func NewGithubClient() GithubClient {
	hs := make(http.Header)
	hs.Set("cache-control", "no-cache")
	hs.Set("Authorization", "token <<TOKEN>>")
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
func (c *githubClient) GetBranchInformation(config *models.Configuration, branchName string) (*models.GetBranchResponse, error) {

	if config.RepositoryName == nil || config.RepositoryOwner == nil || branchName == "" {
		err := errors.New("invalid github body params")
		return nil, err
	}

	response := c.Client.Get(fmt.Sprintf("/repos/%s/%s/branches/%s", *config.RepositoryOwner, *config.RepositoryName, branchName))

	if response.Err() != nil {
		return nil, response.Err()
	}

	var branchInfo models.GetBranchResponse
	if err := json.Unmarshal(response.Bytes(), &branchInfo); err != nil {
		return nil, errors.New("error binding github branch response")
	}

	if response.StatusCode() != http.StatusOK {
		return nil, errors.New("error getting repository branch")
	}

	return &branchInfo, nil
}

//Protects the branch from pushs by following the workflow configuration
//This perform a PUT request to Github api
func (c *githubClient) ProtectBranch(config *models.Configuration, branchConfig *models.Branch) error {

	if branchConfig.Name == "" {
		err := errors.New("invalid branch protection body params")
		return err
	}

	body := map[string]interface{}{
		"enforce_admins":                true,
		"required_status_checks":        branchConfig.Requirements.RequiredStatusChecks,
		"required_pull_request_reviews": branchConfig.Requirements.RequiredPullRequestReviews,
		"restrictions":                  nil,
	}

	response := c.Client.Put(fmt.Sprintf("/repos/%s/%s/branches/%s/protection", *config.RepositoryOwner, *config.RepositoryName, branchConfig.Name), body)

	if response.Err() != nil {
		return response.Err()
	}

	if response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusCreated {
		if response.StatusCode() == http.StatusNotFound {
			return errors.New("branch not found")
		}
		return errors.New(fmt.Sprintf("error protecting branch - status: %d", response.StatusCode()))
	}

	return nil
}

//Create a new reference, in this case a branch
//This perform a POST request to Github api
func (c *githubClient) CreateBranch(config *models.Configuration, branchConfig *models.Branch, sha string) error {

	if branchConfig.Name == "" || config.RepositoryOwner == nil || config.RepositoryName == nil || sha == "" {
		err := errors.New("invalid body params")
		return err
	}

	ref := fmt.Sprintf("refs/heads/%s", branchConfig.Name)

	body := map[string]interface{}{
		"ref": ref,
		"sha": sha,
	}

	response := c.Client.Post(fmt.Sprintf("/repos/%s/%s/git/refs", *config.RepositoryOwner, *config.RepositoryName), body)

	if response.Err() != nil {
		return response.Err()
	}

	if response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusCreated {
		return errors.New(fmt.Sprintf("error creating a branch - status: %d", response.StatusCode()))
	}

	return nil
}

//Create a new reference on github. First we get the information needed to make the creation and then the creation itself.
//This perform a GetBranchInformation and CreateBranch
func (c *githubClient) CreateGithubRef(config *models.Configuration, branchConfig *models.Branch, workflowConfig *models.WorkflowConfig) error {

	if branchConfig.Name == "" || config.RepositoryOwner == nil || config.RepositoryName == nil {
		err := errors.New("invalid body params")
		return err
	}

	//First gets SHA necessary to initialise the new branch or reference
	initialBranch := workflowConfig.DefaultBranch

	if branchConfig.Name == workflowConfig.DefaultBranch {
		initialBranch = "master"
	}

	branchInfo, getBranchError := c.GetBranchInformation(config, initialBranch)

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
func (c *githubClient) SetDefaultBranch(config *models.Configuration, workflowConfig *models.WorkflowConfig) error {

	if config.RepositoryOwner == nil || config.RepositoryName == nil || workflowConfig.DefaultBranch == "" {
		err := errors.New("invalid body params")
		return err
	}

	body := map[string]interface{}{
		"name":           *config.RepositoryName,
		"default_branch": workflowConfig.DefaultBranch,
	}

	response := c.Client.Post(fmt.Sprintf("/repos/%s/%s", *config.RepositoryOwner, *config.RepositoryName), body)

	if response.Err() != nil {
		return response.Err()
	}

	if response.StatusCode() != http.StatusOK && response.StatusCode() != http.StatusCreated {
		return errors.New(fmt.Sprintf("error updating default branch - status: %d", response.StatusCode()))
	}

	return nil
}
