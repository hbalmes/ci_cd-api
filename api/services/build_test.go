package services

import (
	"github.com/coreos/go-semver/semver"
	"github.com/golang/mock/gomock"
	"github.com/hbalmes/ci_cd-api/api/mocks/interfaces"
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/services/storage"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestBuild_ProcessBuild(t *testing.T) {
	type args struct {
		payload           *webhook.Status
		config            *models.Configuration
		getSCTimes        int
		getLastBuildTimes int
		getBuildsTimes    int
		getPRTimes        int
	}

	type expects struct {
		sqlGetByError      error
		sqlGetLatestErr    error
		sqlGetBuildErr     error
		sqlGetPRErr        error
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
			name: "not passed all the status checks yet - build not created",
			args: args{
				payload:           &allowedStatusWebhookSuccess,
				config:            &cicdConfigOK,
				getSCTimes:        1,
				getBuildsTimes:    0,
				getLastBuildTimes: 0,
				getPRTimes:        0,
			},

			expects: expects{
				config:        &cicdConfigOK,
				sqlGetByError: gorm.ErrRecordNotFound,
				buildErr:      apierrors.NewApiError("They have not yet passed all the quality controls necessary to create a new version.", "error", 206, apierrors.CauseList{}),
			},
			wantErr: true,
		},
		{
			name: "not passed all the status checks yet, db error, build not created",
			args: args{
				payload:           &allowedStatusWebhookSuccess,
				config:            &cicdConfigOK,
				getSCTimes:        1,
				getBuildsTimes:    0,
				getLastBuildTimes: 0,
				getPRTimes:        0,
			},

			expects: expects{
				config:        &cicdConfigOK,
				sqlGetByError: gorm.ErrCantStartTransaction,
				buildErr:      apierrors.NewApiError("They have not yet passed all the quality controls necessary to create a new version.", "error", 206, apierrors.CauseList{}),
			},
			wantErr: true,
		},
		{
			name: "is buildable, repo without builds, error getting pull request",
			args: args{
				payload:           &allowedStatusWebhookSuccess,
				config:            &cicdConfigOK,
				getSCTimes:        5,
				getBuildsTimes:    1,
				getLastBuildTimes: 1,
				getPRTimes:        1,
			},

			expects: expects{
				config:          &cicdConfigOK,
				sqlGetByError:   nil,
				sqlGetLatestErr: nil,
				sqlGetPRErr:     gorm.ErrRecordNotFound,
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
					Times(tt.args.getSCTimes),

				sqlStorage.EXPECT().
					GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(e interface{}, qry ...interface{}) *models.LatestBuild {
						return &latestBuild
					}).
					Return(tt.expects.sqlGetLatestErr).
					Times(tt.args.getLastBuildTimes),

				sqlStorage.EXPECT().
					GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(e interface{}, qry ...interface{}) *models.Build {
						return &buildOK
					}).
					Return(tt.expects.sqlGetBuildErr).
					Times(tt.args.getBuildsTimes),

				sqlStorage.EXPECT().
					GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(e interface{}, qry ...interface{}) *webhook.PullRequest {
						return tt.expects.pullRequestWebhook
					}).
					Return(tt.expects.sqlGetPRErr).
					Times(tt.args.getPRTimes),
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

func TestBuild_CreateInitialBuild(t *testing.T) {
	type fields struct {
		SQL storage.SQLStorage
	}

	type args struct {
		config *models.Configuration
	}

	type expects struct {
		buildResult *models.Build
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

	var buildOK models.Build
	buildOK.Sha = utils.Stringify("123456789asdfghjkqwertyu")
	buildOK.Status = utils.Stringify("pending")
	buildOK.Username = utils.Stringify("hbalmes")
	buildOK.RepositoryName = utils.Stringify("ci-cd_api")
	buildOK.Major = 0
	buildOK.Minor = 0
	buildOK.Patch = 0
	buildOK.Branch = utils.Stringify("feature/lalala")
	buildOK.Type = utils.Stringify("productive")

	tests := []struct {
		name    string
		fields  fields
		args    args
		expects expects
	}{
		{
			name: "create a initial build ok",
			args: args{
				config: &cicdConfigOK,
			},
			expects: expects{
				buildResult: &buildOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Build{
				SQL: tt.fields.SQL,
			}

			got := s.CreateInitialBuild(tt.args.config)

			if got != nil {
				assert.Equal(t, uint8(0), tt.expects.buildResult.Major, "the Major should be equals")
				assert.Equal(t, uint16(0), tt.expects.buildResult.Minor, "the Minor should be equals")
				assert.Equal(t, uint16(0), tt.expects.buildResult.Patch, "the Patch should be equals")
				assert.Equal(t, "pending", *tt.expects.buildResult.Status, "the apps should be equals")
				assert.Equal(t, "ci-cd_api", *tt.expects.buildResult.RepositoryName, "the repo URLs should be equals")
				assert.Equal(t, "productive", *tt.expects.buildResult.Type, "the repo URLs should be equals")
			}

		})
	}
}

func TestBuild_IncrementSemVer(t *testing.T) {
	type fields struct {
		SQL storage.SQLStorage
	}
	type args struct {
		version     semver.Version
		incrementer string
	}

	type expects struct {
		versionRes semver.Version
	}

	var initialSemVer semver.Version
	initialSemVer.Major = 0
	initialSemVer.Minor = 1
	initialSemVer.Patch = 0
	initialSemVer.Metadata = ""

	var majorSemVer semver.Version
	majorSemVer.Major = 1
	majorSemVer.Minor = 0
	majorSemVer.Patch = 0
	majorSemVer.Metadata = ""

	var minorSemVer semver.Version
	minorSemVer.Major = 0
	minorSemVer.Minor = 2
	minorSemVer.Patch = 0
	minorSemVer.Metadata = ""

	var patchSemVer semver.Version
	patchSemVer.Major = 0
	patchSemVer.Minor = 1
	patchSemVer.Patch = 1
	patchSemVer.Metadata = ""

	var complexSemVer semver.Version
	complexSemVer.Major = 4
	complexSemVer.Minor = 7
	complexSemVer.Patch = 8
	complexSemVer.Metadata = ""

	tests := []struct {
		name    string
		fields  fields
		args    args
		expects expects
	}{
		{
			name: "create a version incrementing major",
			args: args{
				version:     initialSemVer,
				incrementer: "major",
			},
			expects: expects{
				versionRes: majorSemVer,
			},
		},
		{
			name: "create a version incrementing minor",
			args: args{
				version:     initialSemVer,
				incrementer: "minor",
			},
			expects: expects{
				versionRes: minorSemVer,
			},
		},
		{
			name: "create a version incrementing patch",
			args: args{
				version:     initialSemVer,
				incrementer: "patch",
			},
			expects: expects{
				versionRes: patchSemVer,
			},
		},
		{
			name: "create a version incrementing major on a complex version",
			args: args{
				version:     complexSemVer,
				incrementer: "major",
			},
			expects: expects{
				versionRes: semver.Version{
					Major: 5,
					Minor: 0,
					Patch: 0,
				},
			},
		},
		{
			name: "create a version incrementing minor",
			args: args{
				version:     initialSemVer,
				incrementer: "lalala",
			},
			expects: expects{
				versionRes: minorSemVer,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Build{
				SQL: tt.fields.SQL,
			}
			if got := s.IncrementSemVer(tt.args.version, tt.args.incrementer); !reflect.DeepEqual(got, tt.expects.versionRes) {
				t.Errorf("IncrementSemVer() = %v, want %v", got, tt.expects.versionRes)
			}
		})
	}
}

func TestBuild_GetLatestBuild(t *testing.T) {

	type args struct {
		config   *models.Configuration
		getTimes int
	}

	type expects struct {
		versionRes      *semver.Version
		sqlGetLatestErr error
		sqlGetBuildErr  error
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

	var latestBuild models.LatestBuild
	latestBuild.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	latestBuild.BuildID = 1

	var initialSemVer semver.Version
	initialSemVer.Major = 0
	initialSemVer.Minor = 0
	initialSemVer.Patch = 0
	initialSemVer.Metadata = "0"

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

	tests := []struct {
		name    string
		args    args
		expects expects
	}{
		{
			name: "latest build getted ok, but fail getting build, create initial build",
			args: args{
				config:   &cicdConfigOK,
				getTimes: 1,
			},
			expects: expects{
				versionRes:     &initialSemVer,
				sqlGetBuildErr: gorm.ErrRecordNotFound,
			},
		},
		{
			name: "latest build fails, create initial build",
			args: args{
				config:   &cicdConfigOK,
				getTimes: 0,
			},
			expects: expects{
				versionRes:      &initialSemVer,
				sqlGetLatestErr: gorm.ErrCantStartTransaction,
			},
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
					Times(tt.args.getTimes),
			)

			s := &Build{
				SQL: sqlStorage,
			}
			if got := s.GetLatestBuild(tt.args.config); !reflect.DeepEqual(got, tt.expects.versionRes) {
				t.Errorf("GetLatestBuild() = %v, want %v", got, tt.expects.versionRes)
			}
		})
	}
}

func TestBuild_GetIncrementerAndType(t *testing.T) {
	type fields struct {
		SQL storage.SQLStorage
	}
	type args struct {
		pr *webhook.PullRequest
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		wantIncrementer string
		wantBuildType   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Build{
				SQL: tt.fields.SQL,
			}
			gotIncrementer, gotBuildType := s.GetIncrementerAndType(tt.args.pr)
			if gotIncrementer != tt.wantIncrementer {
				t.Errorf("GetIncrementerAndType() gotIncrementer = %v, want %v", gotIncrementer, tt.wantIncrementer)
			}
			if gotBuildType != tt.wantBuildType {
				t.Errorf("GetIncrementerAndType() gotBuildType = %v, want %v", gotBuildType, tt.wantBuildType)
			}
		})
	}
}