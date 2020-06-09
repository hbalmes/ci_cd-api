package services

import (
	"errors"
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
		sqlGetByError   error
		sqlGetLatestErr error
		sqlGetBuildErr  error
		sqlGetPRErr     error
		sqlDeleteError  error
		config          *models.Configuration
		getConfig       apierrors.ApiError
		build           *models.Build
		buildErr        apierrors.ApiError
		pr              *models.PullRequest
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

	var pullr models.PullRequest
	pullr.ID = 0
	pullr.PullRequestNumber = 12345
	pullr.State = utils.Stringify("open")
	pullr.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	pullr.BaseRef = utils.Stringify("master")
	pullr.HeadRef = utils.Stringify("release/pepe")
	pullr.BaseSha = utils.Stringify("98765432ytrewqjhgfdsadsa")
	pullr.HeadSha = utils.Stringify("23456789qwertyuiasdfghjzxcvbn")
	pullr.CreatedAt = time.Now()
	pullr.UpdatedAt = time.Now()
	pullr.Body = utils.Stringify("pull request body test")
	pullr.Title = utils.Stringify("titulo")
	pullr.CreatedBy = utils.Stringify("hbalmes")

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

			checkBuildability := sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(e interface{}, qry ...interface{}) *webhook.Webhook {
					return &webhookOK
				}).
				Return(tt.expects.sqlGetByError).MaxTimes(tt.args.getSCTimes)

			getLatestBuilds := sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(e interface{}, qry ...interface{}) *models.LatestBuild {
					return &latestBuild
				}).
				Return(tt.expects.sqlGetLatestErr).
				After(checkBuildability).
				MaxTimes(tt.args.getLastBuildTimes)

			getBuilds := sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(e interface{}, qry ...interface{}) *models.Build {
					return &buildOK
				}).
				Return(tt.expects.sqlGetBuildErr).
				After(getLatestBuilds).
				MaxTimes(tt.args.getBuildsTimes)

			sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(e interface{}, qry ...interface{}) *models.PullRequest {
					return &pullr
				}).
				Return(tt.expects.sqlGetPRErr).
				After(getBuilds).
				MaxTimes(tt.args.getPRTimes)

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

	type args struct {
		pr *models.PullRequest
	}

	type expects struct {
		incrementer string
		buildType   string
	}

	var prBaseMasterHeadRelease models.PullRequest
	prBaseMasterHeadRelease.BaseRef = utils.Stringify("master")
	prBaseMasterHeadRelease.HeadRef = utils.Stringify("release/lala")

	var prBaseMasterHeadHotfix models.PullRequest
	prBaseMasterHeadHotfix.BaseRef = utils.Stringify("master")
	prBaseMasterHeadHotfix.HeadRef = utils.Stringify("hotfix/lala")

	var prBaseDevelopHeadFeature models.PullRequest
	prBaseDevelopHeadFeature.BaseRef = utils.Stringify("develop")
	prBaseDevelopHeadFeature.HeadRef = utils.Stringify("feature/lala")

	var prBaseDevelopHeadEnhancement models.PullRequest
	prBaseDevelopHeadEnhancement.BaseRef = utils.Stringify("develop")
	prBaseDevelopHeadEnhancement.HeadRef = utils.Stringify("enhancement/lala")

	var prBaseDevelopHeadFix models.PullRequest
	prBaseDevelopHeadFix.BaseRef = utils.Stringify("develop")
	prBaseDevelopHeadFix.HeadRef = utils.Stringify("fix/lala")

	var prBaseDevelopHeadBugFix models.PullRequest
	prBaseDevelopHeadBugFix.BaseRef = utils.Stringify("develop")
	prBaseDevelopHeadBugFix.HeadRef = utils.Stringify("bugfix/lala")

	var prBaseLalalaHeadLalala models.PullRequest
	prBaseLalalaHeadLalala.BaseRef = utils.Stringify("lalala")
	prBaseLalalaHeadLalala.HeadRef = utils.Stringify("lalalala2")

	tests := []struct {
		name    string
		args    args
		expects expects
	}{
		{
			name: "pr with base master head release, returns type productive and minor incrementer",
			args: args{
				pr: &prBaseMasterHeadRelease,
			},
			expects: expects{
				incrementer: "minor",
				buildType:   "productive",
			},
		},
		{
			name: "pr with base master head hotfix, returns type productive and patch incrementer",
			args: args{
				pr: &prBaseMasterHeadHotfix,
			},
			expects: expects{
				incrementer: "patch",
				buildType:   "productive",
			},
		},
		{
			name: "pr with base develop head feature, returns type test and patch minor",
			args: args{
				pr: &prBaseDevelopHeadFeature,
			},
			expects: expects{
				incrementer: "minor",
				buildType:   "test",
			},
		},
		{
			name: "pr with base develop head enhancement, returns type test and patch minor",
			args: args{
				pr: &prBaseDevelopHeadEnhancement,
			},
			expects: expects{
				incrementer: "minor",
				buildType:   "test",
			},
		},
		{
			name: "pr with base develop head bugfix, returns type test and patch minor",
			args: args{
				pr: &prBaseDevelopHeadBugFix,
			},
			expects: expects{
				incrementer: "patch",
				buildType:   "test",
			},
		},
		{
			name: "pr with base develop head fix, returns type test and patch minor",
			args: args{
				pr: &prBaseDevelopHeadFix,
			},
			expects: expects{
				incrementer: "patch",
				buildType:   "test",
			},
		},
		{
			name: "pr with base lalala head lalala, returns type test and patch minor",
			args: args{
				pr: &prBaseLalalaHeadLalala,
			},
			expects: expects{
				incrementer: "minor",
				buildType:   "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlStorage := interfaces.NewMockSQLStorage(ctrl)

			s := &Build{
				SQL: sqlStorage,
			}
			gotIncrementer, gotBuildType := s.GetIncrementerAndType(tt.args.pr)
			if gotIncrementer != tt.expects.incrementer {
				t.Errorf("GetIncrementerAndType() gotIncrementer = %v, want %v", gotIncrementer, tt.expects.incrementer)
			}
			if gotBuildType != tt.expects.buildType {
				t.Errorf("GetIncrementerAndType() gotBuildType = %v, want %v", gotBuildType, tt.expects.buildType)
			}
		})
	}
}

func TestBuild_GetPullRequestBySha(t *testing.T) {

	type args struct {
		sha string
	}

	type expects struct {
		wantPullRequestWebhook *models.PullRequest
		wantApiError           apierrors.ApiError
		sqlGetErr              error
	}

	var pullRequest models.PullRequest
	pullRequest.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequest.State = utils.Stringify("open")
	pullRequest.HeadSha = utils.Stringify("123456789asdfghjkqwertyu")
	pullRequest.CreatedBy = utils.Stringify("hbalmes")

	var emptyPr models.PullRequest

	tests := []struct {
		name    string
		expects expects
		args    args
	}{
		{
			name: "error getting pull request by sha",
			args: args{
				sha: "1234567wertyasdfghzxcvb",
			},
			expects: expects{
				wantPullRequestWebhook: &emptyPr,
				sqlGetErr:              gorm.ErrCantStartTransaction,
				wantApiError:           apierrors.NewInternalServerApiError("error getting pull request", errors.New("can't start transaction")),
			},
		},
		{
			name: "pull request not found for sha",
			args: args{
				sha: "1234567wertyasdfghzxcvb",
			},
			expects: expects{
				wantPullRequestWebhook: &emptyPr,
				sqlGetErr:              gorm.ErrRecordNotFound,
				wantApiError:           apierrors.NewNotFoundApiError("pull request not found for the sha"),
			},
		},
		{
			name: "pull request getted successfully",
			args: args{
				sha: "1234567wertyasdfghzxcvb",
			},
			expects: expects{
				wantPullRequestWebhook: &pullRequest,
				sqlGetErr:              nil,
				wantApiError:           nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlStorage := interfaces.NewMockSQLStorage(ctrl)

			s := &Build{
				SQL: sqlStorage,
			}

			sqlStorage.EXPECT().
				GetBy(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(e interface{}, qry ...interface{}) models.PullRequest {
					return *tt.expects.wantPullRequestWebhook
				}).
				Return(tt.expects.sqlGetErr).
				AnyTimes()

			_, gotApiError := s.GetPullRequestBySha(tt.args.sha)
			if !reflect.DeepEqual(gotApiError, tt.expects.wantApiError) {
				t.Errorf("GetPullRequestBySha() gotApiError = %v, want %v", gotApiError, tt.expects.wantApiError)
			}
		})
	}
}

func TestBuild_CreateBuild(t *testing.T) {

	type args struct {
		pullRequest *models.PullRequest
		newSemVer   semver.Version
		buildType   string
	}

	type expects struct {
		wantBuild      *models.Build
		wantApiErr     apierrors.ApiError
		sqlInsertError error
	}

	var pullRequest models.PullRequest
	pullRequest.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	pullRequest.State = utils.Stringify("open")
	pullRequest.HeadSha = utils.Stringify("123456789asdfghjkqwertyu")
	pullRequest.CreatedBy = utils.Stringify("hbalmes")
	pullRequest.HeadRef = utils.Stringify("release/lalala")
	pullRequest.BaseRef = utils.Stringify("master")

	var initialSemVer semver.Version
	initialSemVer.Major = 0
	initialSemVer.Minor = 1
	initialSemVer.Patch = 0
	initialSemVer.Metadata = ""

	var buildOK models.Build
	buildOK.Sha = utils.Stringify("123456789asdfghjkqwertyu")
	buildOK.Status = utils.Stringify("pending")
	buildOK.Username = utils.Stringify("hbalmes")
	buildOK.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	buildOK.Major = 0
	buildOK.Minor = 1
	buildOK.Patch = 0
	buildOK.Branch = utils.Stringify("release/lalala")
	buildOK.Type = utils.Stringify("productive")
	buildOK.ID = 0

	tests := []struct {
		name    string
		args    args
		expects expects
	}{
		{
			name: "build created and saved successfully",
			args: args{
				pullRequest: &pullRequest,
				newSemVer:   initialSemVer,
				buildType:   "productive",
			},
			expects: expects{
				wantBuild: &buildOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlStorage := interfaces.NewMockSQLStorage(ctrl)

			s := &Build{
				SQL: sqlStorage,
			}

			got := s.CreateBuild(tt.args.pullRequest, tt.args.newSemVer, tt.args.buildType)

			if got != nil {
				assert.Equal(t, utils.Stringify("123456789asdfghjkqwertyu"), tt.expects.wantBuild.Sha)
				assert.Equal(t, utils.Stringify("pending"), tt.expects.wantBuild.Status)
				assert.Equal(t, utils.Stringify("hbalmes"), tt.expects.wantBuild.Username)
				assert.Equal(t, utils.Stringify("hbalmes/ci-cd_api"), tt.expects.wantBuild.RepositoryName)
				assert.Equal(t, uint8(0), tt.expects.wantBuild.Major)
				assert.Equal(t, uint16(1), tt.expects.wantBuild.Minor)
				assert.Equal(t, uint16(0), tt.expects.wantBuild.Patch)
				assert.Equal(t, utils.Stringify("release/lalala"), tt.expects.wantBuild.Branch)
				assert.Equal(t, utils.Stringify("productive"), tt.expects.wantBuild.Type)
			}
		})
	}
}


func TestBuild_SaveBuild(t *testing.T) {

	type args struct {
		build *models.Build
	}

	type expects struct {
		wantBuild      *models.Build
		wantApiErr     apierrors.ApiError
		sqlInsertError error
	}

	var buildOK models.Build
	buildOK.Sha = utils.Stringify("123456789asdfghjkqwertyu")
	buildOK.Status = utils.Stringify("pending")
	buildOK.Username = utils.Stringify("hbalmes")
	buildOK.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	buildOK.Major = 0
	buildOK.Minor = 1
	buildOK.Patch = 0
	buildOK.Branch = utils.Stringify("release/lalala")
	buildOK.Type = utils.Stringify("productive")
	buildOK.ID = 0

	tests := []struct {
		name    string
		args    args
		expects expects
	}{
		{
			name: "build created and saved successfully",
			args: args{
				build: &buildOK,
			},
			expects: expects{
				wantBuild: &buildOK,
			},
		},
		{
			name: "error creating build, error inserting build to db",
			args: args{
				build: &buildOK,
			},
			expects: expects{
				wantBuild:      nil,
				wantApiErr:     apierrors.NewInternalServerApiError("something was wrong inserting new build", errors.New("record not found")),
				sqlInsertError: gorm.ErrRecordNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlStorage := interfaces.NewMockSQLStorage(ctrl)

			sqlStorage.EXPECT().
				Insert(gomock.Any()).
				Return(tt.expects.sqlInsertError).
				AnyTimes()

			s := &Build{
				SQL: sqlStorage,
			}

			got := s.SaveBuild(tt.args.build)

			if !reflect.DeepEqual(got, tt.expects.wantApiErr) {
				t.Errorf("CreateAndSaveBuild() got1 = %v, want %v", got, tt.expects.wantApiErr)
			}
		})
	}
}

func TestBuild_CreateAndSaveLatestBuild(t *testing.T) {
	type args struct {
		build     *models.Build
		lastBuild semver.Version
	}

	type expects struct {
		wantApiErr     apierrors.ApiError
		sqlDeleteError error
		sqlUpdateError error
	}

	var lastSemVer semver.Version
	lastSemVer.Major = 0
	lastSemVer.Minor = 1
	lastSemVer.Patch = 0
	lastSemVer.Metadata = ""

	var buildOK models.Build
	buildOK.Sha = utils.Stringify("123456789asdfghjkqwertyu")
	buildOK.Status = utils.Stringify("pending")
	buildOK.Username = utils.Stringify("hbalmes")
	buildOK.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	buildOK.Major = 0
	buildOK.Minor = 1
	buildOK.Patch = 0
	buildOK.Branch = utils.Stringify("release/lalala")
	buildOK.Type = utils.Stringify("productive")
	buildOK.ID = 0

	tests := []struct {
		name    string
		args    args
		expects expects
	}{
		{
			name: "error deleting lastBuild",
			args: args{
				build:     &buildOK,
				lastBuild: lastSemVer,
			},
			expects: expects{
				wantApiErr:     apierrors.NewInternalServerApiError("something was wrong deleting repo latest build", errors.New("can't start transaction")),
				sqlDeleteError: gorm.ErrCantStartTransaction,
				sqlUpdateError: nil,
			},
		},
		{
			name: "error updating lastBuild",
			args: args{
				build:     &buildOK,
				lastBuild: lastSemVer,
			},
			expects: expects{
				wantApiErr:     apierrors.NewInternalServerApiError("something was wrong updating repo latest build", errors.New("can't start transaction")),
				sqlDeleteError: nil,
				sqlUpdateError: gorm.ErrCantStartTransaction,
			},
		},
		{
			name: "last build created successfully",
			args: args{
				build:     &buildOK,
				lastBuild: lastSemVer,
			},
			expects: expects{
				wantApiErr:     nil,
				sqlDeleteError: nil,
				sqlUpdateError: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlStorage := interfaces.NewMockSQLStorage(ctrl)

			sqlStorage.EXPECT().
				Delete(gomock.Any()).
				Return(tt.expects.sqlDeleteError).
				AnyTimes()

			sqlStorage.EXPECT().
				Update(gomock.Any()).
				Return(tt.expects.sqlUpdateError).
				AnyTimes()

			s := &Build{
				SQL: sqlStorage,
			}

			if got := s.CreateAndSaveLatestBuild(tt.args.build, &tt.args.lastBuild); !reflect.DeepEqual(got, tt.expects.wantApiErr) {
				t.Errorf("CreateAndSaveLatestBuild() = %v, want %v", got, tt.expects.wantApiErr)
			}
		})
	}
}

func TestBuild_getIssueCommentBody(t *testing.T) {

	type args struct {
		build *models.Build
	}

	type expects struct {
		bodyResult string
	}

	var pendingBuild models.Build
	pendingBuild.Sha = utils.Stringify("123456789asdfghjkqwertyu")
	pendingBuild.Status = utils.Stringify("pending")
	pendingBuild.Username = utils.Stringify("hbalmes")
	pendingBuild.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	pendingBuild.Major = 0
	pendingBuild.Minor = 1
	pendingBuild.Patch = 0
	pendingBuild.ID = 0
	pendingBuild.GithubURL = utils.Stringify("v0.1.0")
	pendingBuild.GithubID = utils.Stringify("123456")

	var finishedBuild models.Build
	finishedBuild.Sha = utils.Stringify("123456789asdfghjkqwertyu")
	finishedBuild.Status = utils.Stringify("finished")
	finishedBuild.Username = utils.Stringify("hbalmes")
	finishedBuild.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	finishedBuild.Major = 0
	finishedBuild.Minor = 1
	finishedBuild.Patch = 0
	finishedBuild.ID = 0
	finishedBuild.GithubURL = utils.Stringify("v0.1.0")
	finishedBuild.GithubID = utils.Stringify("123456")

	var errorBuild models.Build
	errorBuild.Sha = utils.Stringify("123456789asdfghjkqwertyu")
	errorBuild.Status = utils.Stringify("error")
	errorBuild.Username = utils.Stringify("hbalmes")
	errorBuild.RepositoryName = utils.Stringify("hbalmes/ci-cd_api")
	errorBuild.Major = 0
	errorBuild.Minor = 1
	errorBuild.Patch = 0
	errorBuild.ID = 0
	errorBuild.GithubURL = utils.Stringify("v0.1.0")
	errorBuild.GithubID = utils.Stringify("123456")

	tests := []struct {
		name    string
		args    args
		expects expects
	}{
		{
			name: "pending issue comment body",
			args: args{
				build: &pendingBuild,
			},
			expects: expects{
				bodyResult: "# Build report \n\n> **Status:** _**pending** :clock8:\n**Version:**[v0.1.0](https://github.com/hbalmes/ci-cd_api/releases/tag/v0.1.0)> **ID:**v0.1.0",
			},
		},
		{
			name: "finished issue comment body",
			args: args{
				build: &finishedBuild,
			},
			expects: expects{
				bodyResult: "# Build report \n\n> **Status:** _**finished** :white_check_mark:\n**Version:**[v0.1.0](https://github.com/hbalmes/ci-cd_api/releases/tag/v0.1.0)> **ID:**v0.1.0",
			},
		},
		{
			name: "error issue comment body",
			args: args{
				build: &errorBuild,
			},
			expects: expects{
				bodyResult: "# Build report \n\n> **Status:** _**error** :red_circle:\n**Version:**[v0.1.0](https://github.com/hbalmes/ci-cd_api/releases/tag/v0.1.0)> **ID:**v0.1.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			sqlStorage := interfaces.NewMockSQLStorage(ctrl)
			ghClient := interfaces.NewMockGithubClient(ctrl)

			s := &Build{
				SQL:          sqlStorage,
				GithubClient: ghClient,
			}
			if got := s.GetIssueCommentBody(tt.args.build); got != tt.expects.bodyResult {
				t.Errorf("getIssueCommentBody() = %v, want %v", got, tt.expects.bodyResult)
			}
		})
	}
}
