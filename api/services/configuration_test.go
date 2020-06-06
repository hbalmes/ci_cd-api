package services

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/hbalmes/ci_cd-api/api/mocks/interfaces"
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConfiguration_Create(t *testing.T) {
	type args struct {
		payload *models.PostRequestPayload
	}

	type expects struct {
		sqlGetByError       error
		sqlInsertError      error
		setWorkflowError 	apierrors.ApiError
		getConfigError      error
		config              *models.Configuration
		error               apierrors.ApiError
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
		ID:                               utils.Stringify("ci-cd_api"),
		RepositoryName:                   utils.Stringify("ci-cd_api"),
		RepositoryOwner:                  utils.Stringify("hbalmes"),
		RepositoryStatusChecks:           reqChecks,
		WorkflowType:                     utils.Stringify("gitflow"),
		CodeCoveragePullRequestThreshold: &codeCoverageThreadhold,
		CreatedAt:                        time.Time{},
		UpdatedAt:                        time.Time{},
	}

	var postRequestPayloadOK models.PostRequestPayload
	postRequestPayloadOK.Repository.Name = utils.Stringify("ci-cd_api")
	postRequestPayloadOK.Repository.Owner = utils.Stringify("hbalmes")
	postRequestPayloadOK.Repository.RequireStatusChecks = statusList
	postRequestPayloadOK.CodeCoverage.PullRequestThreshold = &codeCoverageThreadhold
	postRequestPayloadOK.Workflow.Type = utils.Stringify("gitflow")
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expects expects
	}{
		{
			name: "error checking configuration existence",
			args: args{
				payload: &postRequestPayloadOK,
			},
			expects: expects{
				sqlGetByError:    gorm.ErrInvalidSQL,
				sqlInsertError:   nil,
				error:            apierrors.NewInternalServerApiError("error checking configuration existence", nil),
			},
			wantErr: true,
		},
		{
			name: "error setting workflow",
			args: args{
				payload: &postRequestPayloadOK,
			},
			expects: expects{
				sqlGetByError:    gorm.ErrRecordNotFound,
				sqlInsertError:   nil,
				setWorkflowError: apierrors.NewInternalServerApiError("some error", nil),
				error:            apierrors.NewInternalServerApiError("error checking configuration existence", nil),
			},
			wantErr: true,
		},
		{
			name: "error inserting in db",
			args: args{
				payload: &postRequestPayloadOK,
			},
			expects: expects{
				sqlGetByError:    gorm.ErrRecordNotFound,
				sqlInsertError:   gorm.ErrInvalidSQL,
				setWorkflowError: nil,
				error:            apierrors.NewInternalServerApiError("error checking configuration existence", nil),
			},
			wantErr: true,
		},
		{
			name: "Config created successfully",
			args: args{
				payload: &postRequestPayloadOK,
			},
			expects: expects{
				sqlGetByError:    gorm.ErrRecordNotFound,
				config: &cicdConfigOK,
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

			s := &Configuration{
				SQL:          sqlStorage,
				GithubClient: githubClient,
			}

			sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(e interface{}, qry ...interface{}) *models.Configuration {
					return &cicdConfigOK
				}).
				Return(tt.expects.sqlGetByError).
				AnyTimes()

			sqlStorage.EXPECT().
				Insert(gomock.Any()).
				Return(tt.expects.sqlInsertError).
				AnyTimes()

			githubClient.EXPECT().
				ProtectBranch(gomock.Any(), gomock.Any()).
				Return(tt.expects.setWorkflowError).
				AnyTimes()

			githubClient.EXPECT().
				SetDefaultBranch(gomock.Any(), gomock.Any()).
				Return(tt.expects.setWorkflowError).
				AnyTimes()

			conf, err := s.Create(tt.args.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Configuration.Create() error = %v, wantErr %v", tt.expects.error, tt.wantErr)
				return
			}

			if !tt.wantErr {
				assert.Equal(t, *tt.expects.config.RepositoryName, *conf.RepositoryName, "Repository not matches")
			}

		})
	}
}

func TestConfiguration_Get(t *testing.T) {
	type args struct {
		id string
	}

	type expects struct {
		errorLog string
		infoLog  string
		error    error
		config   models.Configuration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expects expects
	}{
		{
			name: "test Get - Config obtained successfully, should return it",
			args: args{
				id: "fury_repo-name",
			},
			wantErr: false,
		},
		{
			name: "test Get - Error checking configuration existence, should return an Error",
			args: args{
				id: "fury_repo-name",
			},
			expects: expects{
				error: errors.New("record not found"),
			},
			wantErr: true,
		},
		{
			name: "test Get - gorm Error record Not Found",
			args: args{
				id: "fury_repo-name",
			},
			expects: expects{
				error: gorm.ErrRecordNotFound,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			sqlStorage := interfaces.NewMockSQLStorage(ctl)
			sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(tt.expects.error).
				AnyTimes()

			s := &Configuration{
				SQL:    sqlStorage,
			}

			_, err := s.Get(tt.args.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("Configuration.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
