package services

import (
	"github.com/golang/mock/gomock"
	"github.com/hbalmes/ci_cd-api/api/mocks/interfaces"
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"github.com/jinzhu/gorm"
	"reflect"
	"testing"
	"time"
)

func TestBuild_ProcessBuild(t *testing.T) {
	type args struct {
		payload  *webhook.Status
		config   *models.Configuration
		getTimes int
	}

	type expects struct {
		sqlGetByError      error
		sqlGetLatestErr    error
		sqlGetBuildErr     error
		sqlGetPRErr        error
		sqlInsertPRError   error
		sqlInsertWHError   error
		sqlDeleteError     error
		config             *models.Configuration
		getConfig          apierrors.ApiError
		build              *models.Build
		buildErr           apierrors.ApiError
		pullRequestWebhook *webhook.PullRequest
	}

	var webhookOK webhook.Webhook
	webhookOK.Type = utils.Stringify("status")

	var latestBuild models.LatestBuild
	latestBuild.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	latestBuild.BuildID = 1

	var buildOK models.Build
	buildOK.Sha = utils.Stringify("123456789asdfghjkqwertyu")
	buildOK.Status = utils.Stringify("pending")
	buildOK.Username = utils.Stringify("hbalmes")
	buildOK.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	buildOK.Major = 0
	buildOK.Minor = 1
	buildOK.Patch = 0
	buildOK.Branch = utils.Stringify("feature/lalala")
	buildOK.Type = utils.Stringify("test")

	var pullRequest webhook.PullRequest
	pullRequest.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequest.State = utils.Stringify("open")
	pullRequest.HeadSha = utils.Stringify("123456789asdfghjkqwertyu")
	pullRequest.CreatedBy = utils.Stringify("hbalmes")

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	codeCoverageThreadhold := 80.0

	cicdConfigOK := models.Configuration{
		ID:                               utils.Stringify("ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		CreatedAt:                        time.Time{},
		UpdatedAt:                        time.Time{},
	}

	var allowedStatusWebhookSuccess webhook.Status
	allowedStatusWebhookSuccess.Context = utils.Stringify("workflow")
	allowedStatusWebhookSuccess.Sha = utils.Stringify("23456789qwertyuiasdfghjzxcvbn")
	allowedStatusWebhookSuccess.State = utils.Stringify("success")
	allowedStatusWebhookSuccess.Sender.Login = utils.Stringify("hbalmes")
	allowedStatusWebhookSuccess.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	allowedStatusWebhookSuccess.Description = utils.Stringify("Webhook description")
	allowedStatusWebhookSuccess.TargetURL = utils.Stringify("http://url-api.com")
	allowedStatusWebhookSuccess.Name = utils.Stringify("workflow")

	tests := []struct {
		name    string
		args    args
		wantErr bool
		expects expects
	}{
		{
			name: "not yet passed all the quality checks - build not created",
			args: args{
				payload: &allowedStatusWebhookSuccess,
				config:  &cicdConfigOK,
				getTimes: 5,
			},

			expects: expects{
				config:           &cicdConfigOK,
				sqlGetByError:    gorm.ErrRecordNotFound,
				sqlInsertPRError: nil,
				sqlInsertWHError: gorm.ErrCantStartTransaction,
				buildErr: apierrors.NewApiError("They have not yet passed all the quality controls necessary to create a new version.", "error", 206, apierrors.CauseList{}),
			},
			wantErr: true,
		},
		{
			name: "is buildable, repo without builds, error getting pull request",
			args: args{
				payload: &allowedStatusWebhookSuccess,
				config:  &cicdConfigOK,
				getTimes: 5, // Para pasar el get the status success
			},

			expects: expects{
				config:           &cicdConfigOK,
				sqlGetByError:    nil,
				sqlGetLatestErr:  gorm.ErrRecordNotFound,
				sqlInsertPRError: nil,
				sqlInsertWHError: gorm.ErrCantStartTransaction,
				buildErr:         apierrors.NewNotFoundApiError("pull request not found for the sha"),
			},
			wantErr: true,
		},
		{
			name: "is buildable, repo with builds, error getting pull request",
			args: args{
				payload:  &allowedStatusWebhookSuccess,
				config:   &cicdConfigOK,
				getTimes: 5, // Para pasar el get the status success
			},

			expects: expects{
				config:          &cicdConfigOK,
				sqlGetByError:   nil,
				sqlGetLatestErr: nil,
				sqlGetBuildErr:  nil,
				buildErr:        apierrors.NewNotFoundApiError("pull request not found for the sha"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlStorage := interfaces.NewMockSQLStorage(ctrl)

			gomock.InOrder(
				sqlStorage.EXPECT().
					GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(e interface{}, qry ...interface{}) *webhook.Webhook {
						return &webhookOK
					}).
					Return(tt.expects.sqlGetByError).
					Times(tt.args.getTimes),

				sqlStorage.EXPECT().
					GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(e interface{}, qry ...interface{}) *models.LatestBuild {
						return &latestBuild
					}).
					Return(tt.expects.sqlGetLatestErr).
					Times(1),

				sqlStorage.EXPECT().
					GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(e interface{}, qry ...interface{}) *models.Build {
						return &buildOK
					}).
					Return(tt.expects.sqlGetBuildErr).
					Times(1),

				sqlStorage.EXPECT().
					GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(e interface{}, qry ...interface{}) *webhook.PullRequest {
						return &pullRequest
					}).
					Return(tt.expects.sqlGetLatestErr).
					Times(1),

				sqlStorage.EXPECT().
					Insert(gomock.Any()).
					Return(tt.expects.sqlInsertPRError).
					AnyTimes(),

				sqlStorage.EXPECT().
					Insert(gomock.Any()).
					Return(tt.expects.sqlInsertWHError).
					AnyTimes(),
			)

			s := &Build{
				SQL: sqlStorage,
			}
			build, buildErr := s.ProcessBuild(tt.args.config, tt.args.payload)
			if !reflect.DeepEqual(build, tt.expects.build) {
				t.Errorf("ProcessBuild() got = %v, want %v", build, tt.expects.build)
			}
			if !reflect.DeepEqual(buildErr, tt.expects.buildErr) {
				t.Errorf("ProcessBuild() got1 = %v, want %v", buildErr, tt.expects.buildErr)
			}
		})
	}
}
