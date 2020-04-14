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
		config        models.Configuration
		clientsResult clientsResult
	}

	var pullRequestReviewPayloadOK webhook.PullRequestReviewWebhook
	pullRequestReviewPayloadOK.Action = "submitted"
	pullRequestReviewPayloadOK.Sender.Login = "hbalmes"
	pullRequestReviewPayloadOK.Repository.FullName = "hbalmes/ci-cd_api"
	pullRequestReviewPayloadOK.PullRequest.Head.Sha = "123456789asdfghjkqwertyu"
	pullRequestReviewPayloadOK.Review.State = "approved"
	pullRequestReviewPayloadOK.PullRequest.CreatedAt = "2019-05-15T15:20:33Z"
	pullRequestReviewPayloadOK.PullRequest.UpdatedAt = "2019-05-15T15:20:38Z"
	pullRequestReviewPayloadOK.Review.Body = "Aprobado"

	var pullRequestReviewPayloadAlreadyExists webhook.PullRequestReviewWebhook
	pullRequestReviewPayloadAlreadyExists.Action = "edited"
	pullRequestReviewPayloadAlreadyExists.Sender.Login = "hbalmes"
	pullRequestReviewPayloadAlreadyExists.Repository.FullName = "hbalmes/ci-cd_api"

	var pullRequestReviewPayloadReviewNotApproved webhook.PullRequestReviewWebhook
	pullRequestReviewPayloadReviewNotApproved.Action = "submitted"
	pullRequestReviewPayloadReviewNotApproved.Sender.Login = "hbalmes"
	pullRequestReviewPayloadReviewNotApproved.Repository.FullName = "hbalmes/ci-cd_api"
	pullRequestReviewPayloadReviewNotApproved.Review.State = "edited"

	var pullRequestReviewPayloadReviewDismissed webhook.PullRequestReviewWebhook
	pullRequestReviewPayloadReviewDismissed.Action = "dismissed"
	pullRequestReviewPayloadReviewDismissed.Sender.Login = "hbalmes"
	pullRequestReviewPayloadReviewDismissed.Repository.FullName = "hbalmes/ci-cd_api"
	pullRequestReviewPayloadReviewDismissed.Review.State = "edited"

	var pullRequestReviewPayloadReviewActionNotSupported webhook.PullRequestReviewWebhook
	pullRequestReviewPayloadReviewActionNotSupported.Action = "lalala"

	var webhookOK webhook.Webhook
	webhookOK.Type = utils.Stringify("pull_request_review")

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
				error: gorm.ErrRecordNotFound,
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
				error: gorm.ErrInvalidTransaction,
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
				error: gorm.ErrRecordNotFound,
			},
			wantErr: true,
		},
		{
			name: "test - action: submitted, review approved - Webhook already exists",
			args: args{
				payload: &pullRequestReviewPayloadOK,
			},
			expects: expects{
				error: nil,
			},
			wantErr: false,
		},
		{
			name: "test - action: submitted, review edited - Review State not supported",
			args: args{
				payload: &pullRequestReviewPayloadReviewNotApproved,
			},
			expects: expects{
				error: nil,
			},
			wantErr: true,
		},
		{
			name: "test - action: edited- Action not supported",
			args: args{
				payload: &pullRequestReviewPayloadAlreadyExists,
			},
			expects: expects{
				error: nil,
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
				SQL:          sqlStorage,
				GithubClient: githubClient,
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
		config  *models.Configuration
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
		config              models.Configuration
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
			name: "test - Bad Request - Pull Request Payload",
			args: args{
				payload: &pullRequestWebhookBadRequest,
				config:  &cicdConfigOK,
			},

			expects: expects{
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: apierrors.NewNotFoundApiError("some error"),
				},
				sqlGetByError: gorm.ErrRecordNotFound,
			},
			wantErr: true,
		},
		{
			name: "test - Pull Request Already exists",
			args: args{
				payload: &pullRequestWebhook,
				config:  &cicdConfigOK,
			},

			expects: expects{
				clientsResult: clientsResult{
					sqlClient: nil,
				},
				sqlGetByError: nil,
			},
			wantErr: true,
		},
		{
			name: "test - Error getting values from DB",
			args: args{
				payload: &pullRequestWebhook,
				config:  &cicdConfigOK,
			},

			expects: expects{
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
				config:  &cicdConfigOK,
			},

			expects: expects{
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
				config:  &cicdConfigOK,
			},

			expects: expects{
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
				config:  &cicdConfigOK,
			},

			expects: expects{
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: nil,
				},
				sqlGetByError: gorm.ErrRecordNotFound,
			},
			wantErr: false,
		},
		{
			name: "test - Action Not supported yet",
			args: args{
				payload: &pullRequestWebhookActionNotSupported,
				config:  &cicdConfigOK,
			},

			expects: expects{
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: nil,
				},
				sqlGetByError: gorm.ErrRecordNotFound,
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
			//workflowService := interfaces.NewMockWorkflowService(ctrl)

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

			githubClient.EXPECT().
				CreateStatus(gomock.Any(), gomock.Any()).
				Return(tt.expects.clientsResult.githubClient).
				AnyTimes()

			s := &Webhook{
				SQL:          sqlStorage,
				GithubClient: githubClient,
			}
			_, err := s.ProcessPullRequestWebhook(tt.args.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Webhook.ProcessPullRequestReviewWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestWebhook_ProcessPullRequestWebhookErrorSavingOnDB(t *testing.T) {

	type args struct {
		payload *webhook.PullRequestWebhook
		config  *models.Configuration
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
		config              models.Configuration
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
				config:  &cicdConfigOK,
			},

			expects: expects{
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

			gomock.InOrder(
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
				SQL:          sqlStorage,
				GithubClient: githubClient,
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
		config  *models.Configuration
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
		sqlConfigGetByError error
		sqlInsertError      error
		sqlInsertWHError    error
		getConfigError      error
		config              *models.Configuration
		clientsResult       clientsResult
		workflowCheckResult workflowCheckResult
		savePullRequest     apierrors.ApiError
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

	var notAllowedStatusWebhookSuccess webhook.Status
	notAllowedStatusWebhookSuccess.Sha = utils.Stringify("23456789qwertyuiasdfghjzxcvbn")
	notAllowedStatusWebhookSuccess.State = utils.Stringify("success")
	notAllowedStatusWebhookSuccess.Sender.Login = utils.Stringify("hbalmes")
	notAllowedStatusWebhookSuccess.Repository.FullName = utils.Stringify("hbalmes/ci-cd_api")
	notAllowedStatusWebhookSuccess.Description = utils.Stringify("Webhook description")
	notAllowedStatusWebhookSuccess.TargetURL = utils.Stringify("http://url-api.com")
	notAllowedStatusWebhookSuccess.Name = utils.Stringify("lalalala")
	notAllowedStatusWebhookSuccess.Context = utils.Stringify("lalalala")

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
		wantErr bool
		expects expects
	}{
		{
			name: "test - Not allowed status webhook",
			args: args{
				payload: &notAllowedStatusWebhookSuccess,
				config:  &cicdConfigOK,
			},

			expects: expects{
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: nil,
				},
				sqlGetByError:    nil,
				sqlInsertWHError: nil,
				sqlInsertError:   nil,
			},
			wantErr: true,
		},
		{
			name: "test - Status webhook allowed and already exists on DB",
			args: args{
				payload: &allowedStatusWebhookSuccess,
				config:  &cicdConfigOK,
			},

			expects: expects{
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: nil,
				},
				sqlGetByError:    nil,
				sqlInsertWHError: nil,
				sqlInsertError:   nil,
			},
			wantErr: true,
		},
		{
			name: "test - Error getting webhook from DB",
			args: args{
				payload: &allowedStatusWebhookSuccess,
				config:  &cicdConfigOK,
			},

			expects: expects{
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: nil,
				},
				sqlGetByError:    gorm.ErrInvalidSQL,
				sqlInsertWHError: nil,
				sqlInsertError:   nil,
			},
			wantErr: true,
		},
		{
			name: "test - Error inserting webhook to DB",
			args: args{
				payload: &allowedStatusWebhookSuccess,
				config:  &cicdConfigOK,
			},

			expects: expects{
				clientsResult: clientsResult{
					sqlClient:    nil,
					githubClient: nil,
				},
				sqlGetByError:  gorm.ErrRecordNotFound,
				sqlInsertError: gorm.ErrInvalidTransaction,
			},
			wantErr: true,
		},
		{
			name: "test - Webhook save OK",
			args: args{
				payload: &allowedStatusWebhookSuccess,
				config:  &cicdConfigOK,
			},
			expects: expects{
				sqlGetByError:    gorm.ErrRecordNotFound,
				sqlInsertWHError: nil,
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
				SQL:          sqlStorage,
				GithubClient: githubClient,
			}
			_, err := s.ProcessStatusWebhook(tt.args.payload, tt.args.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("Webhook.ProcessStatusWebhook() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}
