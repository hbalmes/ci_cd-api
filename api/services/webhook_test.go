package services

import (
	"github.com/golang/mock/gomock"
	"github.com/hbalmes/ci_cd-api/api/mocks/interfaces"
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"github.com/jinzhu/gorm"
	"testing"
	"time"
)

func TestWebhook_ProcessPullRequestReviewWebhook(t *testing.T) {

	type args struct {
		payload *webhook.PullRequestReviewWebhook
	}

	type clientsResult struct {
		sqlClient    apierrors.ApiError
		githubClient apierrors.ApiError
	}

	type expects struct {
		error         error
		errorDelete   error
		config        *models.Configuration
		clientsResult clientsResult
		getConfig     apierrors.ApiError
		build         *models.Build
		buildErr      apierrors.ApiError
	}

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	codeCoverageThreadhold := 80.0

	cicdConfigOK := models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		CreatedAt:                        time.Time{},
		UpdatedAt:                        time.Time{},
	}

	var pullRequestReviewPayloadOK webhook.PullRequestReviewWebhook
	pullRequestReviewPayloadOK.Action = utils.Stringify("submitted")
	pullRequestReviewPayloadOK.Sender.Login = utils.Stringify("hbalmes")
	pullRequestReviewPayloadOK.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequestReviewPayloadOK.PullRequest.Head.Sha = utils.Stringify("123456789asdfghjkqwertyu")
	pullRequestReviewPayloadOK.Review.State = utils.Stringify("approved")
	pullRequestReviewPayloadOK.Review.Body = utils.Stringify("Aprobado")

	var pullRequestReviewPayloadAlreadyExists webhook.PullRequestReviewWebhook
	pullRequestReviewPayloadAlreadyExists.Action = utils.Stringify("edited")
	pullRequestReviewPayloadAlreadyExists.Sender.Login = utils.Stringify("hbalmes")
	pullRequestReviewPayloadAlreadyExists.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")

	var pullRequestReviewPayloadReviewNotApproved webhook.PullRequestReviewWebhook
	pullRequestReviewPayloadReviewNotApproved.Action = utils.Stringify("submitted")
	pullRequestReviewPayloadReviewNotApproved.Sender.Login = utils.Stringify("hbalmes")
	pullRequestReviewPayloadReviewNotApproved.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequestReviewPayloadReviewNotApproved.Review.State = utils.Stringify("edited")

	var pullRequestReviewPayloadReviewDismissed webhook.PullRequestReviewWebhook
	pullRequestReviewPayloadReviewDismissed.Action = utils.Stringify("dismissed")
	pullRequestReviewPayloadReviewDismissed.Sender.Login = utils.Stringify("hbalmes")
	pullRequestReviewPayloadReviewDismissed.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequestReviewPayloadReviewDismissed.Review.State = utils.Stringify("edited")
	pullRequestReviewPayloadReviewDismissed.PullRequest.Head.Sha = utils.Stringify("123456789asdfghjkqwertyu")

	var pullRequestReviewPayloadReviewActionNotSupported webhook.PullRequestReviewWebhook
	pullRequestReviewPayloadReviewActionNotSupported.Action = utils.Stringify("lalala")
	pullRequestReviewPayloadReviewActionNotSupported.Sender.Login = utils.Stringify("hbalmes")
	pullRequestReviewPayloadReviewActionNotSupported.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequestReviewPayloadReviewActionNotSupported.Review.State = utils.Stringify("edited")
	pullRequestReviewPayloadReviewActionNotSupported.PullRequest.Head.Sha = utils.Stringify("123456789asdfghjkqwertyu")

	var pullRequestReviewPayloadReviewStateNotSupported webhook.PullRequestReviewWebhook
	pullRequestReviewPayloadReviewStateNotSupported.Action = utils.Stringify("submitted")
	pullRequestReviewPayloadReviewStateNotSupported.Sender.Login = utils.Stringify("hbalmes")
	pullRequestReviewPayloadReviewStateNotSupported.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequestReviewPayloadReviewStateNotSupported.Review.State = utils.Stringify("lalala")
	pullRequestReviewPayloadReviewStateNotSupported.PullRequest.Head.Sha = utils.Stringify("123456789asdfghjkqwertyu")

	var webhookOK webhook.Webhook
	webhookOK.Type = utils.Stringify("pull_request_review")

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
	buildOK.GithubURL = utils.Stringify("v0.1.0")

	tests := []struct {
		name    string
		args    args
		wantErr bool
		expects expects
	}{
		{
			name: "test - action: submitted, review approved - save OK",
			args: args{
				payload: &pullRequestReviewPayloadOK,
			},

			expects: expects{
				clientsResult: clientsResult{
					sqlClient: nil,
				},
				error:  gorm.ErrRecordNotFound,
				config: &cicdConfigOK,
				build:  &buildOK,
			},
			wantErr: false,
		},
		{
			name: "test - action: submitted, review approved - error Getting webhook from DB",
			args: args{
				payload: &pullRequestReviewPayloadOK,
			},

			expects: expects{
				clientsResult: clientsResult{
					sqlClient: nil,
				},
				error:  gorm.ErrInvalidTransaction,
				config: &cicdConfigOK,
			},
			wantErr: true,
		},
		{
			name: "test - action: submitted, review approved - error inserting webhook to DB",
			args: args{
				payload: &pullRequestReviewPayloadOK,
			},
			expects: expects{
				clientsResult: clientsResult{
					sqlClient: apierrors.NewBadRequestApiError("error al guardar papu"),
				},
				error:  gorm.ErrRecordNotFound,
				config: &cicdConfigOK,
			},
			wantErr: true,
		},
		{
			name: "test - action: submitted, review approved - Webhook already exists",
			args: args{
				payload: &pullRequestReviewPayloadOK,
			},
			expects: expects{
				error:  nil,
				config: &cicdConfigOK,
			},
			wantErr: false,
		},
		{
			name: "test - action: submitted, review edited - Review State not supported",
			args: args{
				payload: &pullRequestReviewPayloadReviewStateNotSupported,
			},
			expects: expects{
				error:  nil,
				config: &cicdConfigOK,
			},
			wantErr: true,
		},
		{
			name: "test - action: edited- Action not supported",
			args: args{
				payload: &pullRequestReviewPayloadReviewActionNotSupported,
			},
			expects: expects{
				error:  nil,
				config: &cicdConfigOK,
			},
			wantErr: true,
		},
		{
			name: "test - action: dismissed - value not found - not delete",
			args: args{
				payload: &pullRequestReviewPayloadReviewDismissed,
			},
			expects: expects{
				error:       gorm.ErrRecordNotFound,
				errorDelete: nil,
				config:      &cicdConfigOK,
			},
			wantErr: true,
		},
		{
			name: "test - action: dismissed - error getting webhook from db - not delete",
			args: args{
				payload: &pullRequestReviewPayloadReviewDismissed,
			},
			expects: expects{
				error:       gorm.ErrUnaddressable,
				errorDelete: nil,
				config:      &cicdConfigOK,
			},
			wantErr: true,
		},
		{
			name: "test - action: dismissed - delete Ok from db",
			args: args{
				payload: &pullRequestReviewPayloadReviewDismissed,
			},
			expects: expects{
				error:       nil,
				errorDelete: nil,
				config:      &cicdConfigOK,
			},
			wantErr: false,
		},
		{
			name: "test - action: dismissed - delete Fail from db",
			args: args{
				payload: &pullRequestReviewPayloadReviewDismissed,
			},
			expects: expects{
				error:       nil,
				errorDelete: gorm.ErrInvalidSQL,
				config:      &cicdConfigOK,
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
			configService := interfaces.NewMockConfigurationService(ctrl)
			buildService := interfaces.NewMockBuildService(ctrl)

			configService.EXPECT().
				Get(gomock.Any()).
				Return(tt.expects.config, tt.expects.getConfig).
				AnyTimes()

			buildService.EXPECT().
				ProcessBuild(gomock.Any(), gomock.Any()).Return(tt.expects.build, tt.expects.buildErr).
				AnyTimes()

			sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(e interface{}, qry ...interface{}) *webhook.Webhook {
					return &webhookOK
				}).
				Return(tt.expects.error).
				AnyTimes()

			sqlStorage.EXPECT().
				Insert(gomock.Any()).
				Return(tt.expects.clientsResult.sqlClient).
				AnyTimes()

			sqlStorage.EXPECT().
				Delete(gomock.Any()).
				Return(tt.expects.errorDelete).
				AnyTimes()

			s := &Webhook{
				SQL:           sqlStorage,
				GithubClient:  githubClient,
				ConfigService: configService,
				BuildService:  buildService,
			}
			_, err := s.ProcessPullRequestReviewWebhook(tt.args.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Webhook.ProcessPullRequestReviewWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestWebhook_SavePullRequestWebhook(t *testing.T) {
	type args struct {
		prWebhook webhook.PullRequestWebhook
	}

	type clientsResult struct {
		sqlClient    error
		githubClient apierrors.ApiError
	}

	type expects struct {
		error         error
		errorDelete   error
		config        models.Configuration
		clientsResult clientsResult
	}

	var pullRequestWebhook webhook.PullRequestWebhook
	pullRequestWebhook.Number = 123

	tests := []struct {
		name    string
		args    args
		wantErr bool
		expects expects
	}{{
		name: "test - save pull request -  OK",
		args: args{
			prWebhook: pullRequestWebhook,
		},

		expects: expects{
			clientsResult: clientsResult{
				sqlClient: nil,
			},
			error: gorm.ErrRecordNotFound,
		},
		wantErr: false,
	},
		{
			name: "test - save pull request - FAIL",
			args: args{
				prWebhook: pullRequestWebhook,
			},

			expects: expects{
				clientsResult: clientsResult{
					sqlClient: gorm.ErrUnaddressable,
				},
				error: gorm.ErrRecordNotFound,
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

			sqlStorage.EXPECT().
				Insert(gomock.Any()).
				Return(tt.expects.clientsResult.sqlClient).
				AnyTimes()

			s := &Webhook{
				SQL:          sqlStorage,
				GithubClient: githubClient,
			}

			err := s.SavePullRequestWebhook(tt.args.prWebhook)

			if (err != nil) != tt.wantErr {
				t.Errorf("Webhook.ProcessPullRequestReviewWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestWebhook_ProcessPullRequestWebhook(t *testing.T) {

	type args struct {
		payload *webhook.PullRequestWebhook
	}

	type clientsResult struct {
		sqlClient    apierrors.ApiError
		githubClient apierrors.ApiError
	}

	type workflowCheckResult struct {
		webhookStatus *webhook.Status
		error         apierrors.ApiError
	}

	type expects struct {
		sqlGetByError       error
		sqlInsertError      error
		sqlDeleteError      error
		sqlUpdateError      error
		getConfig           apierrors.ApiError
		config              *models.Configuration
		clientsResult       clientsResult
		workflowCheckResult workflowCheckResult
		savePullRequest     apierrors.ApiError
	}

	var pullRequestWebhook webhook.PullRequestWebhook
	pullRequestWebhook.Number = 12345
	pullRequestWebhook.Action = utils.Stringify("synchronize")
	pullRequestWebhook.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequestWebhook.Sender.Login = utils.Stringify("hbalmes")
	pullRequestWebhook.PullRequest.State = utils.Stringify("open")
	pullRequestWebhook.PullRequest.Head.Sha = utils.Stringify("123456789qwertyuasdfghjzxcvbn")
	pullRequestWebhook.PullRequest.Head.Ref = utils.Stringify("feature/test")
	pullRequestWebhook.PullRequest.Base.Sha = utils.Stringify("lkjhgfdsoiuytrewqmnbvcxz12345")
	pullRequestWebhook.PullRequest.Base.Ref = utils.Stringify("develop")
	pullRequestWebhook.PullRequest.Body = utils.Stringify("Pull request Body")

	var prAlreadyExistsDeleteWebhook webhook.PullRequestWebhook
	prAlreadyExistsDeleteWebhook.Number = 12345
	prAlreadyExistsDeleteWebhook.Action = utils.Stringify("synchronize")
	prAlreadyExistsDeleteWebhook.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	prAlreadyExistsDeleteWebhook.Sender.Login = utils.Stringify("hbalmes")
	prAlreadyExistsDeleteWebhook.PullRequest.State = utils.Stringify("open")
	prAlreadyExistsDeleteWebhook.PullRequest.Head.Sha = utils.Stringify("123456789qwertyuasdfghjzxcvbn")
	prAlreadyExistsDeleteWebhook.PullRequest.Head.Ref = utils.Stringify("feature/test")
	prAlreadyExistsDeleteWebhook.PullRequest.Base.Sha = utils.Stringify("lkjhgfdsoiuytrewqmnbvcxz12345")
	prAlreadyExistsDeleteWebhook.PullRequest.Base.Ref = utils.Stringify("develop")
	prAlreadyExistsDeleteWebhook.PullRequest.Body = utils.Stringify("Pull request Body")

	var pullRequestWebhookClosed webhook.PullRequestWebhook
	pullRequestWebhookClosed.Number = 12345
	pullRequestWebhookClosed.Action = utils.Stringify("closed")
	pullRequestWebhookClosed.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequestWebhookClosed.Sender.Login = utils.Stringify("hbalmes")
	pullRequestWebhookClosed.PullRequest.State = utils.Stringify("open")
	pullRequestWebhookClosed.PullRequest.Head.Sha = utils.Stringify("123456789qwertyuasdfghjzxcvbn")
	pullRequestWebhookClosed.PullRequest.Head.Ref = utils.Stringify("feature/test")
	pullRequestWebhookClosed.PullRequest.Base.Sha = utils.Stringify("lkjhgfdsoiuytrewqmnbvcxz12345")
	pullRequestWebhookClosed.PullRequest.Base.Ref = utils.Stringify("develop")
	pullRequestWebhookClosed.PullRequest.Body = utils.Stringify("Pull request Body")

	var pullRequestWebhookActionNotSupported webhook.PullRequestWebhook
	pullRequestWebhookActionNotSupported.Number = 12345
	pullRequestWebhookActionNotSupported.Action = utils.Stringify("lalalala")
	pullRequestWebhookActionNotSupported.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequestWebhookActionNotSupported.Sender.Login = utils.Stringify("hbalmes")
	pullRequestWebhookActionNotSupported.PullRequest.State = utils.Stringify("open")
	pullRequestWebhookActionNotSupported.PullRequest.Head.Sha = utils.Stringify("123456789qwertyuasdfghjzxcvbn")
	pullRequestWebhookActionNotSupported.PullRequest.Head.Ref = utils.Stringify("feature/test")
	pullRequestWebhookActionNotSupported.PullRequest.Base.Sha = utils.Stringify("lkjhgfdsoiuytrewqmnbvcxz12345")
	pullRequestWebhookActionNotSupported.PullRequest.Base.Ref = utils.Stringify("develop")
	pullRequestWebhookActionNotSupported.PullRequest.Body = utils.Stringify("Pull request Body")

	var pullRequestWebhookBadRequest webhook.PullRequestWebhook
	pullRequestWebhookBadRequest.Number = 12345
	pullRequestWebhookBadRequest.Action = utils.Stringify("opened")
	pullRequestWebhookBadRequest.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequestWebhookBadRequest.Sender.Login = utils.Stringify("hbalmes")
	pullRequestWebhookBadRequest.PullRequest.State = utils.Stringify("open")
	pullRequestWebhookBadRequest.PullRequest.Head.Sha = utils.Stringify("123456789qwertyuasdfghjzxcvbn")
	pullRequestWebhookBadRequest.PullRequest.Base.Sha = utils.Stringify("lkjhgfdsoiuytrewqmnbvcxz12345")
	pullRequestWebhookBadRequest.PullRequest.Base.Ref = utils.Stringify("develop")
	pullRequestWebhookBadRequest.PullRequest.Body = utils.Stringify("Pull request Body")

	var webhookOK webhook.Webhook
	webhookOK.Type = utils.Stringify("pull_request")

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	codeCoverageThreadhold := 80.0

	cicdConfigOK := models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		CreatedAt:                        time.Time{},
		UpdatedAt:                        time.Time{},
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		expects expects
	}{
		{
			name: "test - error getting config",
			args: args{
				payload: &pullRequestWebhookBadRequest,
			},
			expects: expects{
				getConfig: apierrors.NewNotFoundApiError("configuration not found for the repository"),
			},
			wantErr: true,
		},
		{
			name: "test - config not found",
			args: args{
				payload: &pullRequestWebhookBadRequest,
			},
			expects: expects{
				getConfig: nil,
				config:    nil,
			},
			wantErr: true,
		},
		{
			name: "test - Bad Request - Pull Request Payload",
			args: args{
				payload: &pullRequestWebhookBadRequest,
			},

			expects: expects{
				getConfig: nil,
				config:    &cicdConfigOK,
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: apierrors.NewNotFoundApiError("some error"),
				},
				sqlGetByError: gorm.ErrRecordNotFound,
			},
			wantErr: true,
		},
		{
			name: "test - Pull Request Already exists (synchronize)",
			args: args{
				payload: &pullRequestWebhook,
			},
			expects: expects{
				getConfig: nil,
				config:    &cicdConfigOK,
				clientsResult: clientsResult{
					sqlClient: nil,
				},
				sqlGetByError: nil,
			},
			wantErr: false,
		},
		{
			name: "pull request already exists (synchronize), update fails",
			args: args{
				payload: &pullRequestWebhook,
			},
			expects: expects{
				getConfig: nil,
				config:    &cicdConfigOK,
				clientsResult: clientsResult{
					sqlClient: nil,
				},
				sqlGetByError: nil,
				sqlUpdateError: gorm.ErrCantStartTransaction,
			},
			wantErr: true,
		},
		{
			name: "test - Pull Request Already exists (synchronize) - failure sending status",
			args: args{
				payload: &pullRequestWebhook,
			},
			expects: expects{
				getConfig: nil,
				config:    &cicdConfigOK,
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: apierrors.NewNotFoundApiError("some error"),
				},
				sqlGetByError: nil,
			},
			wantErr: true,
		},
		{
			name: "pull request already exists (deleted), default response",
			args: args{
				payload: &prAlreadyExistsDeleteWebhook,
			},
			expects: expects{
				getConfig: nil,
				config:    &cicdConfigOK,
				clientsResult: clientsResult{
					sqlClient: nil,
				},
				sqlGetByError: nil,
				sqlUpdateError: gorm.ErrCantStartTransaction,
			},
			wantErr: true,
		},
		{
			name: "test - Error getting values from DB",
			args: args{
				payload: &pullRequestWebhook,
			},
			expects: expects{
				getConfig: nil,
				config:    &cicdConfigOK,
				clientsResult: clientsResult{
					sqlClient: nil,
				},
				sqlGetByError: gorm.ErrUnaddressable,
			},
			wantErr: true,
		},
		{
			name: "test - Error saving pull request webhook",
			args: args{
				payload: &pullRequestWebhook,
			},
			expects: expects{
				getConfig: nil,
				config:    &cicdConfigOK,
				clientsResult: clientsResult{
					sqlClient: nil,
				},
				sqlGetByError:  gorm.ErrRecordNotFound,
				sqlInsertError: gorm.ErrUnaddressable,
			},
			wantErr: true,
		},
		{
			name: "test - Error creating github Status",
			args: args{
				payload: &pullRequestWebhook,
			},
			expects: expects{
				getConfig: nil,
				config:    &cicdConfigOK,
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: apierrors.NewNotFoundApiError("some error"),
				},
				sqlGetByError: gorm.ErrRecordNotFound,
			},
			wantErr: true,
		},
		{
			name: "test - Pull request Webhook created OK",
			args: args{
				payload: &pullRequestWebhook,
			},
			expects: expects{
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: nil,
				},
				sqlGetByError: gorm.ErrRecordNotFound,
				config:        &cicdConfigOK,
			},
			wantErr: false,
		},
		{
			name: "test - Action Not supported yet",
			args: args{
				payload: &pullRequestWebhookActionNotSupported,
			},
			expects: expects{
				getConfig: nil,
				config:    &cicdConfigOK,
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: nil,
				},
				sqlGetByError: gorm.ErrRecordNotFound,
			},
			wantErr: true,
		},
		{
			name: "test - Pull Request Already exists (closed) - update Ok",
			args: args{
				payload: &pullRequestWebhookClosed,
			},
			expects: expects{
				getConfig: nil,
				config:    &cicdConfigOK,
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: apierrors.NewNotFoundApiError("some error"),
				},
				sqlGetByError: nil,
			},
			wantErr: false,
		},
		{
			name: "test - Pull Request Already exists (closed) - update fail",
			args: args{
				payload: &pullRequestWebhookClosed,
			},
			expects: expects{
				getConfig: nil,
				config:    &cicdConfigOK,
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: apierrors.NewNotFoundApiError("some error"),
				},
				sqlGetByError:  nil,
				sqlUpdateError: gorm.ErrCantStartTransaction,
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
			configService := interfaces.NewMockConfigurationService(ctrl)

			configService.EXPECT().
				Get(gomock.Any()).
				Return(tt.expects.config, tt.expects.getConfig).
				AnyTimes()

			sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(e interface{}, qry ...interface{}) *webhook.Webhook {
					return &webhookOK
				}).
				Return(tt.expects.sqlGetByError).
				AnyTimes()

			sqlStorage.EXPECT().
				Insert(gomock.Any()).
				Return(tt.expects.sqlInsertError).
				AnyTimes()

			sqlStorage.EXPECT().
				Delete(gomock.Any()).
				Return(tt.expects.sqlDeleteError).
				AnyTimes()

			sqlStorage.EXPECT().
				Update(gomock.Any()).
				Return(tt.expects.sqlUpdateError).
				AnyTimes()

			githubClient.EXPECT().
				CreateStatus(gomock.Any(), gomock.Any()).
				Return(tt.expects.clientsResult.githubClient).
				AnyTimes()

			s := &Webhook{
				SQL:           sqlStorage,
				GithubClient:  githubClient,
				ConfigService: configService,
			}
			_, err := s.ProcessPullRequestWebhook(tt.args.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Webhook.ProcessPullRequestWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestWebhook_ProcessPullRequestWebhookErrorSavingOnDB(t *testing.T) {

	type args struct {
		payload *webhook.PullRequestWebhook
	}

	type clientsResult struct {
		sqlClient    apierrors.ApiError
		githubClient apierrors.ApiError
	}

	type workflowCheckResult struct {
		webhookStatus *webhook.Status
		error         apierrors.ApiError
	}

	type expects struct {
		sqlGetByError       error
		sqlInsertPRError    error
		sqlInsertWHError    error
		sqlDeleteError      error
		getConfig           apierrors.ApiError
		config              *models.Configuration
		clientsResult       clientsResult
		workflowCheckResult workflowCheckResult
		savePullRequest     apierrors.ApiError
	}

	var pullRequestWebhook webhook.PullRequestWebhook
	pullRequestWebhook.Number = 12345
	pullRequestWebhook.Action = utils.Stringify("opened")
	pullRequestWebhook.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequestWebhook.Sender.Login = utils.Stringify("hbalmes")
	pullRequestWebhook.PullRequest.State = utils.Stringify("open")
	pullRequestWebhook.PullRequest.Head.Sha = utils.Stringify("123456789qwertyuasdfghjzxcvbn")
	pullRequestWebhook.PullRequest.Head.Ref = utils.Stringify("feature/test")
	pullRequestWebhook.PullRequest.Base.Sha = utils.Stringify("lkjhgfdsoiuytrewqmnbvcxz12345")
	pullRequestWebhook.PullRequest.Base.Ref = utils.Stringify("develop")
	pullRequestWebhook.PullRequest.Body = utils.Stringify("Pull request Body")

	var webhookOK webhook.Webhook
	webhookOK.Type = utils.Stringify("pull_request")

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

	tests := []struct {
		name    string
		args    args
		wantErr bool
		expects expects
	}{
		{
			name: "test - Fail inserting new webhook",
			args: args{
				payload: &pullRequestWebhook,
			},

			expects: expects{
				config: &cicdConfigOK,
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: nil,
				},
				sqlGetByError:    gorm.ErrRecordNotFound,
				sqlInsertPRError: nil,
				sqlInsertWHError: gorm.ErrCantStartTransaction,
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
			configService := interfaces.NewMockConfigurationService(ctrl)

			gomock.InOrder(

				configService.EXPECT().
					Get(gomock.Any()).
					Return(tt.expects.config, tt.expects.getConfig).
					AnyTimes(),

				sqlStorage.EXPECT().
					GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(e interface{}, qry ...interface{}) *webhook.Webhook {
						return &webhookOK
					}).
					Return(tt.expects.sqlGetByError).
					AnyTimes(),

				sqlStorage.EXPECT().
					Insert(gomock.Any()).
					Return(tt.expects.sqlInsertPRError).
					Times(1),

				sqlStorage.EXPECT().
					Insert(gomock.Any()).
					Return(tt.expects.sqlInsertWHError).
					Times(1),
			)

			s := &Webhook{
				SQL:           sqlStorage,
				GithubClient:  githubClient,
				ConfigService: configService,
			}
			_, err := s.ProcessPullRequestWebhook(tt.args.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Webhook.ProcessPullRequestReviewWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestWebhook_ProcessStatusWebhook(t *testing.T) {

	type args struct {
		payload *webhook.Status
	}

	type expects struct {
		sqlGetByError  error
		sqlInsertError error
		getConfig      apierrors.ApiError
		config         *models.Configuration
		build          *models.Build
		buildErr       apierrors.ApiError
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

	notAllowedStatusWebhookSuccess := webhook.Status{
		ID:          0,
		Sha:         utils.Stringify("23456789qwertyuiasdfghjzxcvbn"),
		Name:        utils.Stringify("lalalala"),
		Context:     utils.Stringify("lalalala"),
		Description: utils.Stringify("Webhook description"),
		State:       utils.Stringify("success"),
		TargetURL:   utils.Stringify("http://url-api.com"),
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
		Repository: struct {
			ID       int     `json:"id"`
			FullName *string `json:"full_name"`
		}{
			FullName: utils.Stringify("hbalmes/ci-cd_api"),
		},
		Sender: struct {
			Login *string `json:"login"`
		}{Login: utils.Stringify("hbalmes")},
	}

	var webhookOK webhook.Webhook
	webhookOK.Type = utils.Stringify("pull_request")

	statusList := []string{"workflow", "continuous-integration", "minimum-coverage", "pull-request-coverage"}

	reqChecks := make([]models.RequireStatusCheck, 0)
	for _, rq := range statusList {
		reqChecks = append(reqChecks, models.RequireStatusCheck{
			Check: rq,
		})
	}

	codeCoverageThreadhold := 80.0

	cicdConfigOK := models.Configuration{
		ID:                               utils.Stringify("hbalmes/ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		CreatedAt:                        time.Time{},
		UpdatedAt:                        time.Time{},
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		expects expects
	}{
		{
			name: "test - Not allowed status webhook",
			args: args{
				payload: &notAllowedStatusWebhookSuccess,
			},
			expects: expects{
				getConfig:      nil,
				config:         &cicdConfigOK,
				sqlGetByError:  nil,
				sqlInsertError: nil,
			},
			wantErr: true,
		},
		{
			name: "test - config not found",
			args: args{
				payload: &notAllowedStatusWebhookSuccess,
			},
			expects: expects{
				getConfig:      nil,
				config:         nil,
				sqlGetByError:  nil,
				sqlInsertError: nil,
			},
			wantErr: true,
		},
		{
			name: "test - Error Getting config",
			args: args{
				payload: &notAllowedStatusWebhookSuccess,
			},
			expects: expects{
				getConfig:      apierrors.NewNotFoundApiError("config not found"),
				config:         &cicdConfigOK,
				sqlGetByError:  nil,
				sqlInsertError: nil,
			},
			wantErr: true,
		},
		{
			name: "test - Status webhook allowed and already exists on DB",
			args: args{
				payload: &allowedStatusWebhookSuccess,
			},
			expects: expects{
				getConfig:      nil,
				config:         &cicdConfigOK,
				sqlGetByError:  nil,
				sqlInsertError: nil,
			},
			wantErr: true,
		},
		{
			name: "test - Error getting webhook from DB",
			args: args{
				payload: &allowedStatusWebhookSuccess,
			},
			expects: expects{
				getConfig:      nil,
				config:         &cicdConfigOK,
				sqlGetByError:  gorm.ErrInvalidSQL,
				sqlInsertError: nil,
			},
			wantErr: true,
		},
		{
			name: "test - Error inserting webhook to DB",
			args: args{
				payload: &allowedStatusWebhookSuccess,
			},
			expects: expects{
				getConfig:      nil,
				config:         &cicdConfigOK,
				sqlGetByError:  gorm.ErrRecordNotFound,
				sqlInsertError: gorm.ErrInvalidTransaction,
			},
			wantErr: true,
		},
		{
			name: "test - Webhook save OK",
			args: args{
				payload: &allowedStatusWebhookSuccess,
			},
			expects: expects{
				config:        &cicdConfigOK,
				sqlGetByError: gorm.ErrRecordNotFound,
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
			configService := interfaces.NewMockConfigurationService(ctrl)
			buildService := interfaces.NewMockBuildService(ctrl)

			buildService.EXPECT().
				ProcessBuild(gomock.Any(), gomock.Any()).Return(tt.expects.build, tt.expects.buildErr).
				AnyTimes()

			configService.EXPECT().
				Get(gomock.Any()).
				Return(tt.expects.config, tt.expects.getConfig).
				AnyTimes()

			sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(e interface{}, qry ...interface{}) *webhook.Webhook {
					return &webhookOK
				}).
				Return(tt.expects.sqlGetByError).
				AnyTimes()

			sqlStorage.EXPECT().
				Insert(gomock.Any()).
				Return(tt.expects.sqlInsertError).
				AnyTimes()

			s := &Webhook{
				SQL:           sqlStorage,
				GithubClient:  githubClient,
				ConfigService: configService,
				BuildService:  buildService,
			}
			_, err := s.ProcessStatusWebhook(tt.args.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Webhook.ProcessStatusWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
