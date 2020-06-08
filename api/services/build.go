package services

import (
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/hbalmes/ci_cd-api/api/clients"
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/services/storage"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"time"
)

const (
	initialMajor       = 0
	initialMinor       = 0
	initialPatch       = 0
	initialBuildStatus = "finished"
	initialBuildType   = "productive"
	automaticBuildBody = "release created automatically by hbalmes/ci_cd-api"
)

type BuildService interface {
	ProcessBuild(config *models.Configuration, payload *webhook.Status) (*models.Build, apierrors.ApiError)
}

//Build represents the BuildService layer
//It has an instance of a DBClient layer and
//A Webhook service instance and
//A ConfigService instance
type Build struct {
	SQL          storage.SQLStorage
	GithubClient clients.GithubClient
}

//NewConfigurationSeNewWebhookServicervice initializes a WebhookService
func NewBuildService(sql storage.SQLStorage) *Build {
	return &Build{
		SQL: sql,
		GithubClient: clients.NewGithubClient(),
	}
}

//ProcessBuild checks if all quality checks passed, and creates a build
//For this checks quality checks like (coverage, ci, workflow and pull request review approved)
//Find the latest created build and increment the semver number following the configuration
func (s *Build) ProcessBuild(config *models.Configuration, payload *webhook.Status) (*models.Build, apierrors.ApiError) {

	//First check if all status checks configured pass
	buildeableSChecks := s.GetBuildeableStatusChecks(config)

	isBuildeable := s.CheckBuildability(buildeableSChecks, payload)

	if isBuildeable {
		//Gets the last craeted build
		//if the repo don't have builds created, we generate the default
		lastBuild := s.GetLatestBuild(config)

		//Busca a que PR pertenece el sha para luego saber que campo debo aumentar
		pRequest, err := s.GetPullRequestBySha(*payload.Sha)

		if err != nil {
			return nil, err
		}

		//traemos el incrementador y el tipo de build
		incrementer, buildType := s.GetIncrementerAndType(pRequest)
		newSemVer := s.IncrementSemVer(*lastBuild, incrementer)

		//Creates the build entity
		build:= s.CreateBuild(pRequest, newSemVer, buildType)

		//Creates the github release
		createGHReleaseErr := s.GithubClient.CreateRelease(config, build)

		if createGHReleaseErr != nil {
			return nil, createGHReleaseErr
		}

		//Release Tag Name
		tagName := fmt.Sprintf("v%d.%d.%d", build.Major, build.Minor, build.Patch)
		if build.Tag != nil {
			tagName = tagName + "-" + *build.Tag
		}

		build.GithubURL = utils.Stringify(tagName)

		//Save Build into db
		saveBuildErr := s.SaveBuild(build)

		if saveBuildErr != nil {
			return nil, saveBuildErr
		}

		//Save latest Build into db
		saveLatestErr := s.CreateAndSaveLatestBuild(build, lastBuild)

		if saveLatestErr != nil {
			return nil, err
		}

		//Send the issue comment to Pull request
		issueCommentBody := s.GetIssueCommentBody(build)
		sendIssuecommentErr := s.GithubClient.CreateIssueComment(config, pRequest, issueCommentBody)

		if sendIssuecommentErr != nil {
			log.Error().Err(sendIssuecommentErr).Str("sha", *payload.Sha).
				Str("repository", *payload.Repository.FullName).Msg("error creating new issue comment")
		}

		return build, nil
	}

	return nil, apierrors.NewApiError("They have not yet passed all the quality controls necessary to create a new version.", "error", 206, apierrors.CauseList{})
}

func (s *Build) CheckBuildability(reqSCConfigured []string, payload *webhook.Status) bool {

	//TODO: podriamos mejorar esto y buscar 1 todos los status

	for _, reqSCheck := range reqSCConfigured {
		var webhook webhook.Webhook

		if reqSCheck == "pull_request_review" {
			payload.State = utils.Stringify("approved")
		}
		//Build a ID to identify a unique webhook
		shBaseID := *payload.Repository.FullName + *payload.Sha + reqSCheck + *payload.State
		statusWebhookID := utils.Stringify(utils.GetMD5Hash(shBaseID))

		//Get from db all status from repository and sha
		if err := s.SQL.GetBy(&webhook, "id = ?", statusWebhookID); err != nil {
			if err != gorm.ErrRecordNotFound {
				return false
			}
			return false
		}
	}

	return true
}

func (s *Build) GetBuildeableStatusChecks(config *models.Configuration) []string {

	configuredReqStatusChecks := config.GetRequiredStatusCheck()
	reqSCWithoutCI := utils.Remove(configuredReqStatusChecks, "ci")
	//Add the webhook type pull_request_review
	reqSCWithPRReview := append(reqSCWithoutCI, "pull_request_review")

	return reqSCWithPRReview
}

//Gets the latest build created
//When the repository has no versions created, we create a default
//TODO: Add retries
func (s *Build) GetLatestBuild(config *models.Configuration) *semver.Version {

	var build models.Build
	var latestBuild models.LatestBuild
	var createInitialDefaultBuild bool
	var semverBuild semver.Version

	//get the last build generated for the repository
	err := s.SQL.GetBy(&latestBuild, "repository_name = ?", config.ID)
	if err == nil {
		if err := s.SQL.GetBy(&build, "id = ?", latestBuild.BuildID); err != nil {
			createInitialDefaultBuild = true
		}
	} else {
		createInitialDefaultBuild = true
	}

	if createInitialDefaultBuild {
		build = *s.CreateInitialBuild(config)
	}

	//Build Semver Version
	semverBuild.Major = int64(build.Major)
	semverBuild.Minor = int64(build.Minor)
	semverBuild.Patch = int64(build.Patch)
	semverBuild.Metadata = strconv.Itoa(int(latestBuild.ID))

	return &semverBuild
}

func (s *Build) CreateInitialBuild(config *models.Configuration) *models.Build {

	now := time.Now()

	build := models.Build{
		Major:          initialMajor,
		Minor:          initialMinor,
		Patch:          initialPatch,
		Status:         utils.Stringify(initialBuildStatus),
		UpdatedAt:      utils.Stringify(now.Format("2006-01-02 15:04:05")),
		CreatedAt:      utils.Stringify(now.Format("2006-01-02 15:04:05")),
		RepositoryName: config.ID,
		Type:           utils.Stringify(initialBuildType),
	}

	return &build
}

func (s *Build) GetPullRequestBySha(sha string) (*models.PullRequest, apierrors.ApiError) {

	var pr models.PullRequest
	//Get from db all status from repository and sha
	if err := s.SQL.GetBy(&pr, "head_sha = ?", sha); err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, apierrors.NewInternalServerApiError("error getting pull request", err)
		}
		return nil, apierrors.NewNotFoundApiError("pull request not found for the sha")
	}

	return &pr, nil
}

func (s *Build) GetIncrementerAndType(pr *models.PullRequest) (incrementer string, buildType string) {

	switch *pr.BaseRef {
	case "master":
		buildType = "productive"

		if strings.HasPrefix(*pr.HeadRef, "release/") {
			return "minor", buildType
		}

		if strings.HasPrefix(*pr.HeadRef, "hotfix/") {
			return "patch", buildType
		}
	case "develop":
		buildType = "test"

		developMinorHeadList := []string{"feature/", "enhancement/"}
		developPatchHeadList := []string{"fix/", "bugfix/"}

		for _, headBranchName := range developMinorHeadList {
			if strings.HasPrefix(*pr.HeadRef, headBranchName) {
				return "minor", buildType
			}
		}

		for _, headBranchName := range developPatchHeadList {
			if strings.HasPrefix(*pr.HeadRef, headBranchName) {
				return "patch", buildType
			}
		}
	}

	//if something was wrong
	return "minor", "test"
}

func (s *Build) IncrementSemVer(version semver.Version, incrementer string) semver.Version {

	newVersion := version

	switch incrementer {
	case "major":
		newVersion.BumpMajor()
	case "minor":
		newVersion.BumpMinor()
	case "patch":
		newVersion.BumpPatch()
	default:
		newVersion.BumpMinor()
	}

	return newVersion
}

func (s *Build) CreateBuild(pullRequest *models.PullRequest, newSemVer semver.Version, buildType string) *models.Build {
	var build models.Build

	build.Major = uint8(newSemVer.Major)
	build.Minor = uint16(newSemVer.Minor)
	build.Patch = uint16(newSemVer.Patch)
	build.Status = utils.Stringify(initialBuildStatus)
	build.Sha = pullRequest.HeadSha
	build.Type = utils.Stringify(buildType)
	build.RepositoryName = pullRequest.RepositoryName
	build.UpdatedAt = utils.Stringify(time.Now().Format("2006-01-02 15:04:05"))
	build.CreatedAt = utils.Stringify(time.Now().Format("2006-01-02 15:04:05"))
	build.Branch = pullRequest.HeadRef
	build.Username = pullRequest.CreatedBy
	build.Body = utils.Stringify(automaticBuildBody)

	return &build
}

func (s *Build) SaveBuild(build *models.Build) apierrors.ApiError {
	//Save it into build table
	if err := s.SQL.Insert(&build); err != nil {
		return apierrors.NewInternalServerApiError("something was wrong inserting new build", err)
	}
	return nil
}




func (s *Build) CreateAndSaveLatestBuild(build *models.Build, lastBuild *semver.Version) apierrors.ApiError {

	var latestBuild models.LatestBuild
	latestBuildID, _ := strconv.Atoi(lastBuild.Metadata)
	latestBuild.ID = uint16(latestBuildID)
	latestBuild.BuildID = build.ID
	latestBuild.RepositoryName = build.RepositoryName

	//TODO: FIX this please
	//Delete from configurations DB
	if sqlErr := s.SQL.Delete(&latestBuild); sqlErr != nil {
		return apierrors.NewInternalServerApiError("something was wrong deleting repo latest build", sqlErr)
	}

	//Save it into latestBuild Table
	if err := s.SQL.Update(&latestBuild); err != nil {
		return apierrors.NewInternalServerApiError("something was wrong updating repo latest build", err)
	}

	return nil
}

func (s *Build) GetIssueCommentBody(build *models.Build) string {
	var body string
	var emoji string

	switch *build.Status {
	case "pending":
		emoji = ":clock8:"
		break
	case "finished":
		emoji = ":white_check_mark:"
		break
	case "error":
		emoji = ":red_circle:"
		break
	}

	body = "# Build report \n" + "\n" +
		"> **Status:** _" + fmt.Sprintf("**%s** %s", *build.Status, emoji) + "\n" +
		"**Version:**" + fmt.Sprintf("[%s](https://github.com/%s/releases/tag/%s)", *build.GithubURL, *build.RepositoryName, *build.GithubURL) +
		"> **ID:**" + fmt.Sprintf("%s", *build.GithubID)
	return body
}

