package services

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/hbalmes/ci_cd-api/api/mocks/interfaces"
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"reflect"
	"testing"
	"time"
)

func TestConfiguration_SetWorkflow(t *testing.T) {

	type args struct {
		config *models.Configuration
	}

	type expects struct {
		protectBranchErr    apierrors.ApiError
		createGitRefErr     apierrors.ApiError
		setDefaultBranchErr apierrors.ApiError
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
		name    string
		args    args
		expects expects
		wantErr bool
	}{
		{
			name: "branch protection fails",
			args: args{
				config: &cicdConfigOK,
			},
			expects: expects{
				protectBranchErr: apierrors.NewNotFoundApiError("Some Error"),
			},
			wantErr: true,
		},
		{
			name: "branch protection fails (not found) - create branch fails",
			args: args{
				config: &cicdConfigOK,
			},
			expects: expects{
				protectBranchErr: apierrors.NewInternalServerApiError("branch not found", errors.New("branch not found")),
				createGitRefErr:  apierrors.NewBadRequestApiError("Bad request"),
			},
			wantErr: true,
		},
		{
			name: "branch protection fails (not found) - create branch OK - set default branch OK",
			args: args{
				config: &cicdConfigOK,
			},
			expects: expects{
				protectBranchErr: apierrors.NewInternalServerApiError("branch not found", errors.New("branch not found")),
			},
			wantErr: false,
		},
		{
			name: "branch protection fails (not found) - create branch OK - set default branch FAIL",
			args: args{
				config: &cicdConfigOK,
			},
			expects: expects{
				protectBranchErr:    apierrors.NewInternalServerApiError("branch not found", errors.New("branch not found")),
				setDefaultBranchErr: apierrors.NewNotFoundApiError("Some Error"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlStorage := interfaces.NewMockSQLStorage(ctrl)
			githubClient := interfaces.NewMockGithubClient(ctrl)

			githubClient.EXPECT().
				ProtectBranch(gomock.Any(), gomock.Any()).
				Return(tt.expects.protectBranchErr).
				AnyTimes()

			githubClient.EXPECT().
				CreateGithubRef(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.expects.createGitRefErr).
				AnyTimes()

			githubClient.EXPECT().
				SetDefaultBranch(gomock.Any(), gomock.Any()).
				Return(tt.expects.setDefaultBranchErr).
				AnyTimes()

			c := &Configuration{
				SQL:          sqlStorage,
				GithubClient: githubClient,
			}
			if err := c.SetWorkflow(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("SetWorkflow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfiguration_CheckWorkflow(t *testing.T) {

	type args struct {
		config    *models.Configuration
		prWebhook *webhook.PullRequestWebhook
		headRef   string
		baseRef   string
	}

	type expects struct {
		statusWebhook *webhook.Status
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

	var cicdConfigWithoutGitflow = models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("feature"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
	}

	var pullRequestWebhook webhook.PullRequestWebhook
	pullRequestWebhook.PullRequest.Base.Sha = utils.Stringify("lkjhgfdsoiuytrewqmnbvcxz12345")
	pullRequestWebhook.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequestWebhook.PullRequest.Head.Sha = utils.Stringify("123456789qwertyuasdfghjzxcvbn")
	pullRequestWebhook.PullRequest.Body = utils.Stringify("Pull request Body")
	pullRequestWebhook.Sender.Login = utils.Stringify("hbalmes")

	var statusWebhookOK = webhook.Status{
		ID:          0,
		Sha:         utils.Stringify("123456789qwertyuasdfghjzxcvbn"),
		Context:     utils.Stringify("workflow"),
		State:       utils.Stringify("success"),
		TargetURL:   utils.Stringify("http://www.url_de_wiki"),
		Description: utils.Stringify("Great! You comply with the workflow"),
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
		Repository: struct {
			ID       int     `json:"id"`
			FullName *string `json:"full_name"`
		}{
			ID:       0,
			FullName: utils.Stringify("hbalmes/ci-cd_api"),
		},
	}

	var statusWebhookFail = webhook.Status{
		ID:          0,
		Sha:         utils.Stringify("123456789qwertyuasdfghjzxcvbn"),
		Context:     utils.Stringify("workflow"),
		State:       utils.Stringify("error"),
		TargetURL:   utils.Stringify("http://www.url_de_wiki"),
		Description: utils.Stringify("Oops! You're not complying with the workflow."),
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
		Repository: struct {
			ID       int     `json:"id"`
			FullName *string `json:"full_name"`
		}{
			ID:       0,
			FullName: utils.Stringify("hbalmes/ci-cd_api"),
		},
	}

	tests := []struct {
		name    string
		expects expects
		args    args
		want    *webhook.Status
		baseRef string
		headRef string
	}{
		{
			name: "base: develop - head: feature/pepe - Workflow FAIL",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "feature-pepe",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base: develop - head: feature - Workflow FAIL",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "feature",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base: develop - head: master - Workflow FAIL",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "master",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base: develop - head: release - Workflow FAIL",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "release",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base: develop - head: release/pepe - Workflow FAIL",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "release/pepe",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base: develop - head: release- - Workflow FAIL",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "release/pepe",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base: develop - head: branch-name - Workflow FAIL",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "branch-name",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base: develop - head: bugfix- - Workflow FAIL",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "bugfix-",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base: develop - head: fix- - Workflow FAIL",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "fix-",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base: develop - head: enhancement- - Workflow FAIL",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "enhancement-",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base: develop - head: feature/pepe - Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "feature/pepe",
			},
			want: &statusWebhookOK,
		},
		{
			name: "base:develop head:bugfix/ Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "bugfix/",
			},
			want: &statusWebhookOK,
		},
		{
			name: "base:develop head:fix/ Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "fix/",
			},
			want: &statusWebhookOK,
		},
		{
			name: "base:develop head:enhancement/ Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "develop",
				headRef:   "enhancement/",
			},
			want: &statusWebhookOK,
		},
		{
			name: "base:master head:release/ Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "master",
				headRef:   "release/",
			},
			want: &statusWebhookOK,
		},
		{
			name: "base:master head:hotfix/ Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "master",
				headRef:   "hotfix/",
			},
			want: &statusWebhookOK,
		},
		{
			name: "base:master head:develop Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "master",
				headRef:   "develop/",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base:release head:master Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "release",
				headRef:   "master",
			},
			want: &statusWebhookOK,
		},
		{
			name: "base:release head:develop Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "release",
				headRef:   "develop",
			},
			want: &statusWebhookOK,
		},
		{
			name: "base:release head:lalala Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "release",
				headRef:   "lalala",
			},
			want: &statusWebhookOK,
		},
		{
			name: "base:release/ head:bugfix- Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "release/",
				headRef:   "bugfix-",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base:release/ head:hotfix- Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "release/",
				headRef:   "hotfix-",
			},
			want: &statusWebhookFail,
		},
		{
			name: "base:release/ head:hotfix/ Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "release/",
				headRef:   "hotfix/",
			},
			want: &statusWebhookOK,
		},
		{
			name: "base:branch_base head:branch_head Workflow OK",
			args: args{
				config:    &cicdConfigOK,
				prWebhook: &pullRequestWebhook,
				baseRef:   "branch_base",
				headRef:   "branch_head",
			},
			want: &statusWebhookOK,
		},
		{
			name: "workflow != gitflow - gets gitflow config (for now) - Workflow OK",
			args: args{
				config:    &cicdConfigWithoutGitflow,
				prWebhook: &pullRequestWebhook,
				baseRef:   "release",
				headRef:   "develop",
			},
			want: &statusWebhookOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlStorage := interfaces.NewMockSQLStorage(ctrl)
			githubClient := interfaces.NewMockGithubClient(ctrl)

			prWebhook := tt.args.prWebhook
			prWebhook.PullRequest.Head.Ref = utils.Stringify(tt.args.headRef)
			prWebhook.PullRequest.Base.Ref = utils.Stringify(tt.args.baseRef)

			c := &Configuration{
				SQL:          sqlStorage,
				GithubClient: githubClient,
			}
			if got := c.CheckWorkflow(tt.args.config, prWebhook); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckWorkflow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfiguration_UnsetWorkflow(t *testing.T) {
	type args struct {
		config *models.Configuration
	}

	type expects struct {
		unprotectBranchErr apierrors.ApiError
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
		name    string
		args    args
		expects expects
		wantErr bool
	}{
		{
			name: "unprotect branch fails",
			args: args{
				config: &cicdConfigOK,
			},
			expects: expects{
				unprotectBranchErr: apierrors.NewNotFoundApiError("Some Error"),
			},
			wantErr: true,
		},
		{
			name: "unprotect branch fails (branch not found)",
			args: args{
				config: &cicdConfigOK,
			},
			expects: expects{
				unprotectBranchErr: apierrors.NewInternalServerApiError("branch not found", errors.New("branch not found")),
			},
			wantErr: true,
		},
		{
			name: "unprotect branch success",
			args: args{
				config: &cicdConfigOK,
			},
			expects: expects{
				unprotectBranchErr: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlStorage := interfaces.NewMockSQLStorage(ctrl)
			githubClient := interfaces.NewMockGithubClient(ctrl)

			githubClient.EXPECT().
				UnprotectBranch(gomock.Any(), gomock.Any()).
				Return(tt.expects.unprotectBranchErr).
				AnyTimes()

			c := &Configuration{
				SQL:          sqlStorage,
				GithubClient: githubClient,
			}
			if err := c.UnsetWorkflow(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("SetWorkflow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
