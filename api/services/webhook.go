package services

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/herbal828/ci_cd-api/api/clients"
	"github.com/herbal828/ci_cd-api/api/models"
	"github.com/herbal828/ci_cd-api/api/models/webhook"
	"github.com/herbal828/ci_cd-api/api/services/storage"
	"github.com/herbal828/ci_cd-api/api/utils"
	"github.com/herbal828/ci_cd-api/api/utils/apierrors"
	"github.com/jinzhu/gorm"
	"io/ioutil"
)

type WebhookService interface {
	CreateWebhook(ctx *gin.Context, webhookEvent string) (*webhook.Webhook, apierrors.ApiError)
	ProcessStatusWebhook(ctx utils.HTTPContext, conf *models.Configuration) (*webhook.Webhook, apierrors.ApiError)
	ProcessPullRequestWebhook(ctx utils.HTTPContext) (*webhook.Webhook, apierrors.ApiError)
	SavePullRequestWebhook(pullRequestWH webhook.PullRequestWebhook) apierrors.ApiError
}

//Webhook represents the WebhookService layer
//It has an instance of a DBClient layer and
//A github client instance
type Webhook struct {
	SQL          storage.SQLStorage
	GithubClient clients.GithubClient
}

//NewConfigurationSeNewWebhookServicervice initializes a WebhookService
func NewWebhookService(sql storage.SQLStorage) *Webhook {
	return &Webhook{
		SQL:          sql,
		GithubClient: clients.NewGithubClient(),
	}
}

//CreateWebhook creates a new webhook for the given repository
//It could returns
//	200OK in case of a success processing the creation
//	400BadRequest in case of an error parsing the request payload
//	500InternalServerError in case of an internal error procesing the creation
func (s *Webhook) CreateWebhook(ctx *gin.Context, webhookEvent string) (*webhook.Webhook, apierrors.ApiError) {

	// Read the content
	var bodyBytes []byte
	if ctx.Request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(ctx.Request.Body)
	}
	// Restore the io.ReadCloser to its original state
	ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	//validate that the repository comes in the payload
	var ghPayload webhook.GithubWebhookStandardPayload
	if err := json.Unmarshal(bodyBytes, &ghPayload); err != nil {
		return nil, apierrors.NewBadRequestApiError("invalid github webhook payload")
	}

	repository := ghPayload.Repository.Name

	var config models.Configuration

	//Validates that the repository has a ci cd configuration
	if err := s.SQL.GetBy(&config, "id = ?", *repository); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apierrors.NewNotFoundApiError("repository dosn't have a ci-cd configuration")
		} else {
			return nil, apierrors.NewInternalServerApiError("error checking configuration existence", err)
		}
	}

	var wh webhook.Webhook

	if webhookEvent == "" {
		return nil, apierrors.NewBadRequestApiError("x-github-event is null")
	}

	switch webhookEvent {
	case "status":
		wh, err := s.ProcessStatusWebhook(ctx, &config)
		if err != nil {
			return nil, err
		}
		return wh, nil
	case "pull_request_review":
	case "issue_comment":
	case "pull_request":
		wh, err := s.ProcessPullRequestWebhook(ctx)
		if err != nil {
			return nil, err
		}
		return wh, nil
	case "create":
	default:
		return nil, apierrors.NewBadRequestApiError("Event not supported yet")
	}
	return &wh, nil
}

//ProcessStatusWebhook process
func (s *Webhook) ProcessStatusWebhook(ctx utils.HTTPContext, conf *models.Configuration) (*webhook.Webhook, apierrors.ApiError) {

	var statusWH webhook.Status
	var wh webhook.Webhook

	if err := ctx.BindJSON(&statusWH); err != nil {
		return nil, apierrors.NewBadRequestApiError("invalid status webhook payload")
	}

	contextAllowed := utils.Contains(conf.RepositoryStatusChecks, statusWH.Context)
	if !contextAllowed {
		return nil, apierrors.NewBadRequestApiError("Context not configured for the repository")
	}

	//Build a ID to identify a unique webhook
	shBaseID := statusWH.Repository.FullName + statusWH.Sha + statusWH.Context + statusWH.State
	statusWebhookID := utils.Stringify(utils.GetMD5Hash(shBaseID))

	//Search the status webhook into database
	if err := s.SQL.GetBy(&wh, "id = ?", &statusWebhookID); err != nil {

		//If the error is not a not found error, then there is a problem
		if err != gorm.ErrRecordNotFound {
			return nil, apierrors.NewNotFoundApiError("error checking status webhook existence")
		}

		//Fill every field in the webhook
		wh.ID = statusWebhookID
		wh.GithubDeliveryID = utils.Stringify(ctx.GetHeader("X-GitHub-Delivery"))
		wh.Type = utils.Stringify(ctx.GetHeader("X-GitHub-Event"))
		wh.GithubRepositoryName = utils.Stringify(statusWH.Repository.FullName)
		wh.SenderName = utils.Stringify(statusWH.Sender.Login)
		wh.WebhookCreateAt = statusWH.CreatedAt
		wh.WebhookUpdated = statusWH.UpdatedAt
		wh.State = utils.Stringify(statusWH.State)
		wh.Context = utils.Stringify(statusWH.Context)
		wh.Sha = utils.Stringify(statusWH.Sha)
		wh.Description = utils.Stringify(statusWH.Description)

		//Save it into database
		if err := s.SQL.Insert(&wh); err != nil {
			return nil, apierrors.NewInternalServerApiError("error saving new status webhook", err)
		}

	} else { //If webhook already exists then return it
		return nil, apierrors.NewConflictApiError("Resource Already exists")
	}

	return &wh, nil
}

//ProcessPullRequestWebhook process
func (s *Webhook) ProcessPullRequestWebhook(ctx utils.HTTPContext) (*webhook.Webhook, apierrors.ApiError) {

	var pullRequestWH webhook.PullRequestWebhook
	var prWH webhook.PullRequest
	var wh webhook.Webhook

	if err := ctx.BindJSON(&pullRequestWH); err != nil {
		return nil, apierrors.NewBadRequestApiError("invalid pull_request webhook payload")
	}

	//Search the pull request webhook in database
	if err := s.SQL.GetBy(&prWH, "id = ?", &pullRequestWH.PullRequest.ID); err != nil {

		//If the error is not a not found error, then there is a problem
		if err != gorm.ErrRecordNotFound {
			return nil, apierrors.NewNotFoundApiError("error checking pull request existence")
		}

		switch pullRequestWH.Action {
		case "opened":
			wh, err := s.ProcessStatusWebhook(ctx, &config)
			if err != nil {
				return nil, err
			}
			return wh, nil
		case "pull_request_review":
		case "issue_comment":
		case "pull_request":
			wh, err := s.ProcessPullRequestWebhook(ctx)
			if err != nil {
				return nil, err
			}
			return wh, nil
		case "create":
		default:
			return nil, apierrors.NewBadRequestApiError("Event not supported yet")
		}

		saveErr := s.SavePullRequestWebhook(pullRequestWH)

		if saveErr != nil {
			return nil, saveErr
		}

		//Build a ID to identify a unique webhook
		whBaseID := pullRequestWH.Repository.FullName + pullRequestWH.PullRequest.Head.Sha + string(pullRequestWH.PullRequest.ID) + pullRequestWH.PullRequest.State
		prWebhookID := utils.Stringify(utils.GetMD5Hash(whBaseID))

		//Fill every field in the webhook
		wh.ID = prWebhookID
		wh.GithubDeliveryID = utils.Stringify(ctx.GetHeader("X-GitHub-Delivery"))
		wh.Type = utils.Stringify(ctx.GetHeader("X-GitHub-Event"))
		wh.GithubRepositoryName = utils.Stringify(pullRequestWH.Repository.FullName)
		wh.SenderName = utils.Stringify(pullRequestWH.Sender.Login)
		wh.WebhookCreateAt = pullRequestWH.PullRequest.CreatedAt
		wh.WebhookUpdated = pullRequestWH.PullRequest.UpdatedAt
		wh.State = utils.Stringify(pullRequestWH.PullRequest.State)
		wh.Sha = utils.Stringify(pullRequestWH.PullRequest.Head.Sha)
		wh.Description = utils.Stringify(pullRequestWH.PullRequest.Body)
		wh.GithubPullRequestNumber = &pullRequestWH.PullRequest.Number

		//Save it into database
		if err := s.SQL.Insert(&wh); err != nil {
			return nil, apierrors.NewInternalServerApiError("error saving new pull request webhook", err)
		}

	} else { //If webhook already exists then return it
		return nil, apierrors.NewConflictApiError("Resource Already exists")
	}

	return &wh, nil
}

func (s *Webhook) SavePullRequestWebhook(pullRequestWH webhook.PullRequestWebhook) apierrors.ApiError {

	var prWH webhook.PullRequest

	//Fill every field in the pull request
	prWH.ID = &pullRequestWH.PullRequest.ID
	prWH.PullRequestNumber = &pullRequestWH.PullRequest.Number
	prWH.Body = utils.Stringify(pullRequestWH.PullRequest.Body)
	prWH.State = utils.Stringify(pullRequestWH.PullRequest.State)
	prWH.RepositoryName = utils.Stringify(pullRequestWH.Repository.FullName)
	prWH.Title = utils.Stringify(pullRequestWH.PullRequest.Title)
	prWH.BaseRef = utils.Stringify(pullRequestWH.PullRequest.Base.Ref)
	prWH.BaseSha = utils.Stringify(pullRequestWH.PullRequest.Base.Sha)
	prWH.HeadRef = utils.Stringify(pullRequestWH.PullRequest.Head.Ref)
	prWH.HeadSha = utils.Stringify(pullRequestWH.PullRequest.Head.Sha)
	prWH.CreatedAt = pullRequestWH.PullRequest.CreatedAt
	prWH.UpdatedAt = pullRequestWH.PullRequest.UpdatedAt
	prWH.CreatedBy = utils.Stringify(pullRequestWH.PullRequest.User.Login)

	//Save it into database
	if err := s.SQL.Insert(&prWH); err != nil {
		return apierrors.NewInternalServerApiError("error saving new status webhook", err)
	}

	return nil
}
