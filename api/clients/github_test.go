package clients

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/hbalmes/ci_cd-api/api/configs"
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func Test_githubClient_GetBranchInformation(t *testing.T) {
	type restResponse struct {
		mockError      error
		mockStatusCode int
		mockBytes      []byte
	}

	type args struct {
		config     *models.Configuration
		body       map[string]interface{}
		branchName string
	}

	type expects struct {
		error apierrors.ApiError
		want  *models.GetBranchResponse
	}

	getBranchResp := models.GetBranchResponse{
		Name: "feature/pepe",
		Commit: struct {
			Sha string `json:"sha"`
		}{"1245678qwertyuasdfghzxcvb"},
		Protected: false,
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	codeCoverageThreadhold := 80.0

	var cicdConfigOK = models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
	}

	tests := []struct {
		name         string
		args         args
		restResponse restResponse
		wantErr      bool
		expects      expects
	}{
		{
			name: "bad request branch name nil (invalid github body params)",
			args: args{
				config: &cicdConfigOK,
			},
			expects: expects{
				error: apierrors.NewBadRequestApiError("invalid github body params"),
			},
			restResponse: restResponse{},
			wantErr:      true,
		},
		{
			name: "rest client error getting the branch information",
			restResponse: restResponse{
				mockError: errors.New("some error"),
			},
			args: args{
				config:     &cicdConfigOK,
				branchName: "feature/pepe",
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("Something went wrong getting branch information", errors.New("some error")),
			},
			wantErr: true,
		},
		{
			name: "error binding github branch respose",
			restResponse: restResponse{
				mockError:      nil,
				mockStatusCode: 200,
				mockBytes: utils.GetBytes(map[string]interface{}{
					"name": 1,
				}),
			},
			args: args{
				config:     &cicdConfigOK,
				branchName: "feature/pepe",
			},
			expects: expects{
				error: apierrors.NewBadRequestApiError("error binding github branch response"),
			},
			wantErr: true,
		},
		{
			name: "branch not found",
			restResponse: restResponse{
				mockStatusCode: 404,
			},
			args: args{
				config:     &cicdConfigOK,
				branchName: "feature/pepe",
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("error getting repository branch", nil),
			},
			wantErr: true,
		},
		{
			name: "branch information getted OK",
			restResponse: restResponse{
				mockStatusCode: 200,
				mockBytes: utils.GetBytes(map[string]interface{}{
					"name":      "feature/pepe",
					"commit":    map[string]interface{}{"sha": "1245678qwertyuasdfghzxcvb"},
					"protected": false,
				}),
			},
			args: args{
				config:     &cicdConfigOK,
				branchName: "feature/pepe",
			},
			expects: expects{
				error: nil,
				want:  &getBranchResp,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := NewMockClient(ctrl)
			response := NewMockResponse(ctrl)

			response.
				EXPECT().
				Err().
				Return(tt.restResponse.mockError).
				AnyTimes()

			response.
				EXPECT().
				StatusCode().
				Return(tt.restResponse.mockStatusCode).
				AnyTimes()

			response.
				EXPECT().
				Bytes().
				Return(tt.restResponse.mockBytes).
				AnyTimes()

			client.EXPECT().
				Get(gomock.Any()).
				Return(response).
				AnyTimes()

			c := &githubClient{
				Client: client,
			}
			branchResp, error := c.GetBranchInformation(tt.args.config, tt.args.branchName)
			if !reflect.DeepEqual(branchResp, tt.expects.want) {
				t.Errorf("GetBranchInformation() got = %v, want %v", tt.expects.want, tt.expects.error)
			}
			if !reflect.DeepEqual(error, tt.expects.error) {
				t.Errorf("GetBranchInformation() error = %v, want %v", error, tt.expects.error)
			}
		})
	}
}

func Test_githubClient_ProtectBranch(t *testing.T) {
	type restResponse struct {
		mockError      error
		mockStatusCode int
		mockBytes      []byte
	}

	type args struct {
		config       *models.Configuration
		body         map[string]interface{}
		branchConfig *models.Branch
	}

	type expects struct {
		error apierrors.ApiError
		want  *models.GetBranchResponse
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	codeCoverageThreadhold := 80.0

	var cicdConfigOK = models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
	}

	var masterBranchConfig models.Branch
	var branchConfig models.Branch

	workflowConfig := configs.GetGitflowConfig(&cicdConfigOK)

	branchesConfig := workflowConfig.Description.Branches

	for _, branch := range branchesConfig {
		if strings.HasPrefix("release/", *branch.Name) {
			//releaseBranchConfig = branch
		} else {
			switch *branch.Name {
			case "master":
				masterBranchConfig = branch
			case "develop":
				//developBranchConfig = branch
			}
		}
	}

	tests := []struct {
		name         string
		args         args
		restResponse restResponse
		wantErr      bool
		expects      expects
	}{
		{
			name: "bad request branch name nil (invalid branch protection body params)",
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &branchConfig,
			},
			expects: expects{
				error: apierrors.NewBadRequestApiError("invalid branch protection body params"),
			},
			restResponse: restResponse{},
			wantErr:      true,
		},
		{
			name: "update (put) branch protection fails",
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &masterBranchConfig,
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("branch not found", nil),
			},
			restResponse: restResponse{
				mockStatusCode: 404,
			},
			wantErr: true,
		},
		{
			name: "rest client error updating branch protection",
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &masterBranchConfig,
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("Something went wrong protecting branch", errors.New("some error")),
			},
			restResponse: restResponse{
				mockError: errors.New("some error"),
			},
			wantErr: true,
		},
		{
			name: "Bad request error updating branch protection",
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &masterBranchConfig,
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("error protecting branch - status: 400", nil),
			},
			restResponse: restResponse{
				mockStatusCode: 400,
			},
			wantErr: true,
		},
		{
			name: "Update Branch protection Success",
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &masterBranchConfig,
			},
			expects: expects{
				error: nil,
			},
			restResponse: restResponse{
				mockStatusCode: 200,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := NewMockClient(ctrl)
			response := NewMockResponse(ctrl)

			response.
				EXPECT().
				Err().
				Return(tt.restResponse.mockError).
				AnyTimes()

			response.
				EXPECT().
				StatusCode().
				Return(tt.restResponse.mockStatusCode).
				AnyTimes()

			response.
				EXPECT().
				Bytes().
				Return(tt.restResponse.mockBytes).
				AnyTimes()

			client.EXPECT().
				Put(gomock.Any(), gomock.Any()).
				Return(response).
				AnyTimes()

			c := &githubClient{
				Client: client,
			}

			if got := c.ProtectBranch(tt.args.config, tt.args.branchConfig); !reflect.DeepEqual(got, tt.expects.error) {
				t.Errorf("ProtectBranch() = %v, want %v", got, tt.expects.error)
			}
		})
	}
}

func Test_githubClient_CreateBranch(t *testing.T) {
	type restResponse struct {
		mockError      error
		mockStatusCode int
		mockBytes      []byte
	}

	type args struct {
		config       *models.Configuration
		body         map[string]interface{}
		branchConfig *models.Branch
		sha          string
	}

	type expects struct {
		error apierrors.ApiError
		want  *models.GetBranchResponse
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	codeCoverageThreadhold := 80.0

	var cicdConfigOK = models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
	}

	var masterBranchConfig models.Branch
	var branchConfig models.Branch

	workflowConfig := configs.GetGitflowConfig(&cicdConfigOK)

	branchesConfig := workflowConfig.Description.Branches

	for _, branch := range branchesConfig {
		if strings.HasPrefix("release/", *branch.Name) {
			//releaseBranchConfig = branch
		} else {
			switch *branch.Name {
			case "master":
				masterBranchConfig = branch
			case "develop":
				//developBranchConfig = branch
			}
		}
	}

	tests := []struct {
		name         string
		args         args
		restResponse restResponse
		wantErr      bool
		expects      expects
	}{
		{
			name: "bad request branch name nil (invalid branch protection body params)",
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &branchConfig,
			},
			expects: expects{
				error: apierrors.NewBadRequestApiError("invalid body params"),
			},
			restResponse: restResponse{},
			wantErr:      true,
		},
		{
			name: "rest client error getting the branch information",
			restResponse: restResponse{
				mockError: errors.New("some error"),
			},
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &masterBranchConfig,
				sha:          "123456qwertyasdfgzxcvb",
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("Something went wrong creating a branch", errors.New("some error")),
			},
			wantErr: true,
		},
		{
			name: "Bad request creating branch",
			restResponse: restResponse{
				mockStatusCode: 400,
			},
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &masterBranchConfig,
				sha:          "123456qwertyasdfgzxcvb",
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("error creating a branch - status: 400", nil),
			},
			wantErr: true,
		},
		{
			name: "branch created successfully",
			restResponse: restResponse{
				mockStatusCode: 201,
			},
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &masterBranchConfig,
				sha:          "123456qwertyasdfgzxcvb",
			},
			expects: expects{
				error: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := NewMockClient(ctrl)
			response := NewMockResponse(ctrl)

			response.
				EXPECT().
				Err().
				Return(tt.restResponse.mockError).
				AnyTimes()

			response.
				EXPECT().
				StatusCode().
				Return(tt.restResponse.mockStatusCode).
				AnyTimes()

			response.
				EXPECT().
				Bytes().
				Return(tt.restResponse.mockBytes).
				AnyTimes()

			client.EXPECT().
				Post(gomock.Any(), gomock.Any()).
				Return(response).
				AnyTimes()

			c := &githubClient{
				Client: client,
			}

			if got := c.CreateBranch(tt.args.config, tt.args.branchConfig, tt.args.sha); !reflect.DeepEqual(got, tt.expects.error) {
				t.Errorf("ProtectBranch() = %v, want %v", got, tt.expects.error)
			}
		})
	}
}

func Test_githubClient_CreateGithubRef(t *testing.T) {

	type restPostResponse struct {
		mockError      error
		mockStatusCode int
		mockBytes      []byte
	}

	type restGetResponse struct {
		mockError      error
		mockStatusCode int
		mockBytes      []byte
	}

	type args struct {
		config         *models.Configuration
		body           map[string]interface{}
		branchConfig   *models.Branch
		workflowConfig *models.WorkflowConfig
	}

	type expects struct {
		error         apierrors.ApiError
		want          *models.GetBranchResponse
		getBranchInfo *models.GetBranchResponse
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	codeCoverageThreadhold := 80.0

	var cicdConfigOK = models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
	}

	var masterBranchConfig models.Branch
	var branchConfig models.Branch

	workflowConfig := configs.GetGitflowConfig(&cicdConfigOK)

	branchesConfig := workflowConfig.Description.Branches

	for _, branch := range branchesConfig {
		if strings.HasPrefix("release/", *branch.Name) {
			//releaseBranchConfig = branch
		} else {
			switch *branch.Name {
			case "master":
				masterBranchConfig = branch
			case "develop":
				//developBranchConfig = branch
			}
		}
	}

	featureWorkflowConfig := models.WorkflowConfig{
		Name:          utils.Stringify("feature"),
		Description:   models.Description{},
		Detail:        utils.Stringify("feature workflow config"),
		DefaultBranch: utils.Stringify("master"),
	}

	branchInfo := models.GetBranchResponse{
		Name: "master",
		Commit: struct {
			Sha string `json:"sha"`
		}{Sha: "234567qwertasdfghzxcvb"},
		Protected: false,
	}

	tests := []struct {
		name             string
		args             args
		restGetResponse  restGetResponse
		restPostResponse restPostResponse
		wantErr          bool
		expects          expects
	}{
		{
			name: "bad request branch name nil (invalid branch protection body params)",
			args: args{
				config:         &cicdConfigOK,
				branchConfig:   &branchConfig,
				workflowConfig: workflowConfig,
			},
			expects: expects{
				error: apierrors.NewBadRequestApiError("invalid body params"),
			},
			wantErr: true,
		},
		{
			name: "feature workflow - default branch master - failure getting branch info",
			args: args{
				config:         &cicdConfigOK,
				branchConfig:   &masterBranchConfig,
				workflowConfig: &featureWorkflowConfig,
			},
			restGetResponse: restGetResponse{
				mockError:      nil,
				mockStatusCode: 200,
				mockBytes: utils.GetBytes(map[string]interface{}{
					"name": 1,
				}),
			},
			expects: expects{
				error: apierrors.NewBadRequestApiError("error binding github branch response"),
			},
			wantErr: true,
		},
		{
			name: "failure creating branch",
			args: args{
				config:         &cicdConfigOK,
				branchConfig:   &masterBranchConfig,
				workflowConfig: &featureWorkflowConfig,
			},
			restGetResponse: restGetResponse{
				mockError:      nil,
				mockStatusCode: 200,
				mockBytes:      utils.GetBytes(branchInfo),
			},
			restPostResponse: restPostResponse{
				mockError:      nil,
				mockStatusCode: 400,
				mockBytes:      utils.GetBytes(branchInfo),
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("error creating a branch - status: 400", nil),
			},
			wantErr: true,
		},
		{
			name: "ref created successfully",
			args: args{
				config:         &cicdConfigOK,
				branchConfig:   &masterBranchConfig,
				workflowConfig: &featureWorkflowConfig,
			},
			restGetResponse: restGetResponse{
				mockError:      nil,
				mockStatusCode: 200,
				mockBytes:      utils.GetBytes(branchInfo),
			},
			restPostResponse: restPostResponse{
				mockError:      nil,
				mockStatusCode: 201,
				mockBytes:      utils.GetBytes(branchInfo),
			},
			expects: expects{
				error: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := NewMockClient(ctrl)
			postResponse := NewMockResponse(ctrl)
			getResponse := NewMockResponse(ctrl)

			getResponse.
				EXPECT().
				Err().
				Return(tt.restGetResponse.mockError).
				AnyTimes()

			getResponse.
				EXPECT().
				StatusCode().
				Return(tt.restGetResponse.mockStatusCode).
				AnyTimes()

			getResponse.
				EXPECT().
				Bytes().
				Return(tt.restGetResponse.mockBytes).
				AnyTimes()

			postResponse.
				EXPECT().
				Err().
				Return(tt.restPostResponse.mockError).
				AnyTimes()

			postResponse.
				EXPECT().
				StatusCode().
				Return(tt.restPostResponse.mockStatusCode).
				AnyTimes()

			postResponse.
				EXPECT().
				Bytes().
				Return(tt.restPostResponse.mockBytes).
				AnyTimes()

			client.EXPECT().
				Get(gomock.Any()).
				Return(getResponse).
				AnyTimes()

			client.EXPECT().
				Post(gomock.Any(), gomock.Any()).
				Return(postResponse).
				AnyTimes()

			c := &githubClient{
				Client: client,
			}

			if got := c.CreateGithubRef(tt.args.config, tt.args.branchConfig, tt.args.workflowConfig); !reflect.DeepEqual(got, tt.expects.error) {
				t.Errorf("ProtectBranch() = %v, want %v", got, tt.expects.error)
			}
		})
	}
}

func Test_githubClient_SetDefaultBranch(t *testing.T) {
	type restResponse struct {
		mockError      error
		mockStatusCode int
		mockBytes      []byte
	}

	type args struct {
		config         *models.Configuration
		workflowConfig *models.WorkflowConfig
		body           map[string]interface{}
	}

	type expects struct {
		error apierrors.ApiError
		want  *models.GetBranchResponse
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	codeCoverageThreadhold := 80.0

	var cicdConfigOK = models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
	}

	var cicdConfigError = models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
	}

	workflowConfig := configs.GetGitflowConfig(&cicdConfigOK)

	tests := []struct {
		name         string
		args         args
		restResponse restResponse
		wantErr      bool
		expects      expects
	}{
		{
			name: "bad request branch name nil (invalid branch protection body params)",
			args: args{
				workflowConfig: workflowConfig,
				config:         &cicdConfigError,
			},
			expects: expects{
				error: apierrors.NewBadRequestApiError("invalid body params"),
			},
			restResponse: restResponse{},
			wantErr:      true,
		},
		{
			name: "rest client error posting",
			restResponse: restResponse{
				mockError: errors.New("some error"),
			},
			args: args{
				workflowConfig: workflowConfig,
				config:         &cicdConfigOK,
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("Something went wrong setting default branch", errors.New("some error")),
			},
			wantErr: true,
		},
		{
			name: "Bad request setting default branch",
			restResponse: restResponse{
				mockStatusCode: 400,
			},
			args: args{
				workflowConfig: workflowConfig,
				config:         &cicdConfigOK,
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("error updating default branch - status: 400", nil),
			},
			wantErr: true,
		},
		{
			name: "default branch setted successfully",
			restResponse: restResponse{
				mockStatusCode: 201,
			},
			args: args{
				workflowConfig: workflowConfig,
				config:         &cicdConfigOK,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := NewMockClient(ctrl)
			response := NewMockResponse(ctrl)

			response.
				EXPECT().
				Err().
				Return(tt.restResponse.mockError).
				AnyTimes()

			response.
				EXPECT().
				StatusCode().
				Return(tt.restResponse.mockStatusCode).
				AnyTimes()

			response.
				EXPECT().
				Bytes().
				Return(tt.restResponse.mockBytes).
				AnyTimes()

			client.EXPECT().
				Post(gomock.Any(), gomock.Any()).
				Return(response).
				AnyTimes()

			c := &githubClient{
				Client: client,
			}

			if got := c.SetDefaultBranch(tt.args.config, tt.args.workflowConfig); !reflect.DeepEqual(got, tt.expects.error) {
				t.Errorf("ProtectBranch() = %v, want %v", got, tt.expects.error)
			}
		})
	}
}

func Test_githubClient_CreateStatus(t *testing.T) {
	type restResponse struct {
		mockError      error
		mockStatusCode int
		mockBytes      []byte
	}

	type args struct {
		config *models.Configuration
		status *webhook.Status
	}

	type expects struct {
		error apierrors.ApiError
		want  *models.GetBranchResponse
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	codeCoverageThreadhold := 80.0

	var cicdConfigOK = models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
	}

	var cicdConfigError = models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
	}

	var statusOK = webhook.Status{
		ID:          1,
		Sha:         utils.Stringify("234567qwertasdfghzxcvb"),
		Name:        utils.Stringify("workflow"),
		Context:     utils.Stringify("workflow"),
		Description: utils.Stringify("Great! You comply with the workflow"),
		State:       utils.Stringify("success"),
		TargetURL:   utils.Stringify("http://www.url_de_wiki"),
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
		Repository: struct {
			ID       int     `json:"id"`
			FullName *string `json:"full_name"`
		}{FullName: utils.Stringify("hbalmes/ci-cd_api")},
		Sender: struct {
			Login *string `json:"login"`
		}{Login: utils.Stringify("hbalmes")},
	}

	tests := []struct {
		name         string
		args         args
		restResponse restResponse
		wantErr      bool
		expects      expects
	}{
		{
			name: "bad request branch name nil (invalid branch protection body params)",
			args: args{
				config: &cicdConfigError,
			},
			expects: expects{
				error: apierrors.NewBadRequestApiError("invalid body params"),
			},
			restResponse: restResponse{},
			wantErr:      true,
		},
		{
			name: "rest client error posting",
			restResponse: restResponse{
				mockError: errors.New("some error"),
			},
			args: args{
				config: &cicdConfigOK,
				status: &statusOK,
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("RestClient Error creating new status", errors.New("some error")),
			},
			wantErr: true,
		},
		{
			name: "Bad request setting default branch",
			restResponse: restResponse{
				mockStatusCode: 400,
			},
			args: args{
				config: &cicdConfigOK,
				status: &statusOK,
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("Error creating new status", nil),
			},
			wantErr: true,
		},
		{
			name: "default branch setted successfully",
			restResponse: restResponse{
				mockStatusCode: 201,
			},
			args: args{
				config: &cicdConfigOK,
				status: &statusOK,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := NewMockClient(ctrl)
			response := NewMockResponse(ctrl)

			response.
				EXPECT().
				Err().
				Return(tt.restResponse.mockError).
				AnyTimes()

			response.
				EXPECT().
				StatusCode().
				Return(tt.restResponse.mockStatusCode).
				AnyTimes()

			response.
				EXPECT().
				Bytes().
				Return(tt.restResponse.mockBytes).
				AnyTimes()

			client.EXPECT().
				Post(gomock.Any(), gomock.Any()).
				Return(response).
				AnyTimes()

			c := &githubClient{
				Client: client,
			}

			if got := c.CreateStatus(tt.args.config, tt.args.status); !reflect.DeepEqual(got, tt.expects.error) {
				t.Errorf("ProtectBranch() = %v, want %v", got, tt.expects.error)
			}
		})
	}
}

func Test_githubClient_UnprotectBranch(t *testing.T) {
	type restResponse struct {
		mockError      error
		mockStatusCode int
		mockBytes      []byte
	}

	type args struct {
		config       *models.Configuration
		body         map[string]interface{}
		branchConfig *models.Branch
	}

	type expects struct {
		error apierrors.ApiError
		want  *models.GetBranchResponse
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	codeCoverageThreadhold := 80.0

	var cicdConfigOK = models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
	}

	var masterBranchConfig models.Branch
	var branchConfig models.Branch

	workflowConfig := configs.GetGitflowConfig(&cicdConfigOK)

	branchesConfig := workflowConfig.Description.Branches

	for _, branch := range branchesConfig {
		if strings.HasPrefix("release/", *branch.Name) {
			//releaseBranchConfig = branch
		} else {
			switch *branch.Name {
			case "master":
				masterBranchConfig = branch
			case "develop":
				//developBranchConfig = branch
			}
		}
	}

	tests := []struct {
		name         string
		args         args
		restResponse restResponse
		wantErr      bool
		expects      expects
	}{
		{
			name: "bad request branch name nil (invalid branch protection body params)",
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &branchConfig,
			},
			expects: expects{
				error: apierrors.NewBadRequestApiError("invalid branch body params"),
			},
			restResponse: restResponse{},
			wantErr:      true,
		},
		{
			name: "branch not found",
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &masterBranchConfig,
			},
			expects: expects{
				error: nil,
			},
			restResponse: restResponse{
				mockStatusCode: 404,
			},
			wantErr: false,
		},
		{
			name: "rest client error updating branch protection",
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &masterBranchConfig,
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("Something went wrong deleting branch protection", errors.New("some error")),
			},
			restResponse: restResponse{
				mockError: errors.New("some error"),
			},
			wantErr: true,
		},
		{
			name: "Bad request error deleting branch protection",
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &masterBranchConfig,
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("error deleting branch protection- status: 400", nil),
			},
			restResponse: restResponse{
				mockStatusCode: 400,
			},
			wantErr: true,
		},
		{
			name: "Delete Branch protection Success",
			args: args{
				config:       &cicdConfigOK,
				branchConfig: &masterBranchConfig,
			},
			expects: expects{
				error: nil,
			},
			restResponse: restResponse{
				mockStatusCode: 204,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := NewMockClient(ctrl)
			response := NewMockResponse(ctrl)

			response.
				EXPECT().
				Err().
				Return(tt.restResponse.mockError).
				AnyTimes()

			response.
				EXPECT().
				StatusCode().
				Return(tt.restResponse.mockStatusCode).
				AnyTimes()

			response.
				EXPECT().
				Bytes().
				Return(tt.restResponse.mockBytes).
				AnyTimes()

			client.EXPECT().
				Delete(gomock.Any()).
				Return(response).
				AnyTimes()

			c := &githubClient{
				Client: client,
			}

			if got := c.UnprotectBranch(tt.args.config, tt.args.branchConfig); !reflect.DeepEqual(got, tt.expects.error) {
				t.Errorf("ProtectBranch() = %v, want %v", got, tt.expects.error)
			}
		})
	}
}

func Test_githubClient_CreateIssueComment(t *testing.T) {

	type restResponse struct {
		mockError      error
		mockStatusCode int
		mockBytes      []byte
	}

	type args struct {
		config           *models.Configuration
		issueCommentBody string
		pullRequest      *models.PullRequest
	}

	type expects struct {
		error apierrors.ApiError
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	codeCoverageThreadhold := 80.0

	var cicdConfigOK = models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
	}

	var cicdConfigBadRequest = models.Configuration{
		RepositoryName:         nil,
		RepositoryOwner:        nil,
		RepositoryStatusChecks: nil,
	}

	var pullr models.PullRequest
	pullr.ID = 0
	pullr.PullRequestNumber = 12345
	pullr.State = utils.Stringify("open")
	pullr.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	pullr.BaseRef = utils.Stringify("master")
	pullr.HeadRef = utils.Stringify("release/pepe")
	pullr.BaseSha = utils.Stringify("98765432ytrewqjhgfdsadsa")
	pullr.HeadSha = utils.Stringify("123456789asdfghjkqwertyu")
	pullr.CreatedAt = time.Now()
	pullr.UpdatedAt = time.Now()
	pullr.Body = utils.Stringify("pull request body test")
	pullr.Title = utils.Stringify("titulo")
	pullr.CreatedBy = utils.Stringify("hbalmes")

	tests := []struct {
		name         string
		args         args
		restResponse restResponse
		wantErr      bool
		expects      expects
	}{
		{
			name: "bad request branch name nil (invalid github body params)",
			args: args{
				config:      &cicdConfigBadRequest,
				pullRequest: &pullr,
				issueCommentBody: "lalalala",
			},
			expects: expects{
				error: apierrors.NewBadRequestApiError("invalid body params"),
			},
			restResponse: restResponse{},
			wantErr:      true,
		},
		{
			name: "rest client error creating issue comment",
			restResponse: restResponse{
				mockError: errors.New("some error"),
			},
			args: args{
				config: &cicdConfigOK,
				pullRequest: &pullr,
				issueCommentBody: "lalalala",
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("restClient Error creating new issue comment", errors.New("some error")),
			},
			wantErr: true,
		},
		{
			name: "rest client error , url not found",
			restResponse: restResponse{
				mockStatusCode: 404,
			},
			args: args{
				config: &cicdConfigOK,
				pullRequest: &pullr,
				issueCommentBody: "lalalala",
			},
			expects: expects{
				error: apierrors.NewInternalServerApiError("error creating new issue comment", nil),
			},
			wantErr: true,
		},
		{
			name: "issue comment created successfully",
			restResponse: restResponse{
				mockStatusCode: 200,
				mockBytes: utils.GetBytes(map[string]interface{}{
					"name":      "feature/pepe",
					"commit":    map[string]interface{}{"sha": "1245678qwertyuasdfghzxcvb"},
					"protected": false,
				}),
			},
			args: args{
				config: &cicdConfigOK,
				pullRequest: &pullr,
				issueCommentBody: "lalalala",
			},
			expects: expects{
				error: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client := NewMockClient(ctrl)
			response := NewMockResponse(ctrl)

			response.
				EXPECT().
				Err().
				Return(tt.restResponse.mockError).
				AnyTimes()

			response.
				EXPECT().
				StatusCode().
				Return(tt.restResponse.mockStatusCode).
				AnyTimes()

			response.
				EXPECT().
				Bytes().
				Return(tt.restResponse.mockBytes).
				AnyTimes()

			client.EXPECT().
				Post(gomock.Any(), gomock.Any()).
				Return(response).
				AnyTimes()

			c := &githubClient{
				Client: client,
			}

			if got := c.CreateIssueComment(tt.args.config, tt.args.pullRequest, tt.args.issueCommentBody); !reflect.DeepEqual(got, tt.expects.error) {
				t.Errorf("CreateIssueComment() = %v, want %v", got, tt.expects.error)
			}
		})
	}
}
