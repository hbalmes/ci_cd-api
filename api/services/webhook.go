package services

import (
	"github.com/hbalmes/ci_cd-api/api/clients"
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/services/storage"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"github.com/jinzhu/gorm"
)

const (
	pullRequestReviewSubmittedAction = "submitted"
	pullRequestReviewEditedAction    = "edited"
	pullRequestReviewDismissedAction = "dismissed"
	approvedPullRequestReviewState   = "approved"
)

type WebhookService interface {
	ProcessStatusWebhook(payload *webhook.Status) (*webhook.Webhook, apierrors.ApiError)
	ProcessPullRequestWebhook(payload *webhook.PullRequestWebhook) (*webhook.Webhook, apierrors.ApiError)
	ProcessPullRequestReviewWebhook(payload *webhook.PullRequestReviewWebhook) (*webhook.Webhook, apierrors.ApiError)
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

//ProcessStatusWebhook process
func (s *Webhook) ProcessStatusWebhook(payload *webhook.Status) (*webhook.Webhook, apierrors.ApiError) {

	var wh webhook.Webhook

	webhookType := "status"

	repository := utils.Stringify(payload.Repository.FullName)

	var conf models.Configuration

	//Validates that the repository has a ci cd configuration
	if err := s.SQL.GetBy(&conf, "id = ?", *repository); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apierrors.NewNotFoundApiError("repository dons't have a ci-cd configuration")
		} else {
			return nil, apierrors.NewInternalServerApiError("error checking configuration existence", err)
		}
	}

	contextAllowed := utils.ContainsStatusChecks(conf.RepositoryStatusChecks, payload.Context)

	if !contextAllowed {
		return nil, apierrors.NewBadRequestApiError("Context not configured for the repository")
	}

	//Build a ID to identify a unique webhook
	shBaseID := payload.Repository.FullName + payload.Sha + payload.Context + payload.State
	statusWebhookID := utils.Stringify(utils.GetMD5Hash(shBaseID))

	//Search the status webhook into database
	if err := s.SQL.GetBy(&wh, "id = ?", &statusWebhookID); err != nil {

		//If the error is not a not found error, then there is a problem
		if err != gorm.ErrRecordNotFound {
			return nil, apierrors.NewNotFoundApiError("error checking status webhook existence")
		}

		//Fill every field in the webhook
		wh.ID = statusWebhookID
		//wh.GithubDeliveryID = utils.Stringify(ctx.GetHeader("X-GitHub-Delivery"))
		wh.Type = utils.Stringify(webhookType)
		wh.GithubRepositoryName = utils.Stringify(payload.Repository.FullName)
		wh.SenderName = utils.Stringify(payload.Sender.Login)
		wh.WebhookCreateAt = payload.CreatedAt
		wh.WebhookUpdated = payload.UpdatedAt
		wh.State = utils.Stringify(payload.State)
		wh.Context = utils.Stringify(payload.Context)
		wh.Sha = utils.Stringify(payload.Sha)
		wh.Description = utils.Stringify(payload.Description)

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
func (s *Webhook) ProcessPullRequestWebhook(payload *webhook.PullRequestWebhook) (*webhook.Webhook, apierrors.ApiError) {

	var prWH webhook.PullRequest
	var wh webhook.Webhook
	var cf Configuration

	repository := utils.Stringify(payload.Repository.FullName)

	var config models.Configuration

	//Validates that the repository has a ci cd configuration
	if err := s.SQL.GetBy(&config, "id = ?", *repository); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apierrors.NewNotFoundApiError("repository dons't have a ci-cd configuration")
		} else {
			return nil, apierrors.NewInternalServerApiError("error checking configuration existence", err)
		}
	}

	if payload.PullRequest.Base.Ref == "" || payload.PullRequest.Head.Ref == "" {
		return nil, apierrors.NewBadRequestApiError("Base or Head Ref cant be null")
	}

	webhookType := "pull_request"

	//Search the pull request webhook in database
	if err := s.SQL.GetBy(&prWH, "id = ?", &payload.PullRequest.ID); err != nil {

		//If the error is not a not found error, then there is a problem
		if err != gorm.ErrRecordNotFound {
			return nil, apierrors.NewNotFoundApiError("error checking pull request existence")
		}

		//Save the Pull request
		saveErr := s.SavePullRequestWebhook(*payload)

		if saveErr != nil {
			return nil, saveErr
		}

		//Build a ID to identify a unique webhook
		whBaseID := payload.Repository.FullName + payload.PullRequest.Head.Sha + string(payload.PullRequest.ID) + payload.PullRequest.State
		prWebhookID := utils.Stringify(utils.GetMD5Hash(whBaseID))

		//Fill every field in the webhook
		wh.ID = prWebhookID
		//wh.GithubDeliveryID = utils.Stringify(ctx.GetHeader("X-GitHub-Delivery"))
		wh.Type = utils.Stringify(webhookType)
		wh.GithubRepositoryName = utils.Stringify(payload.Repository.FullName)
		wh.SenderName = utils.Stringify(payload.Sender.Login)
		wh.WebhookCreateAt = payload.PullRequest.CreatedAt
		wh.WebhookUpdated = payload.PullRequest.UpdatedAt
		wh.State = utils.Stringify(payload.PullRequest.State)
		wh.Sha = utils.Stringify(payload.PullRequest.Head.Sha)
		wh.Description = utils.Stringify(payload.PullRequest.Body)
		wh.GithubPullRequestNumber = &payload.PullRequest.Number

		//Save it into database
		if err := s.SQL.Insert(&wh); err != nil {
			return nil, apierrors.NewInternalServerApiError("error saving new pull request webhook", err)
		}

		switch payload.Action {
		case "opened":

			statusWH := cf.CheckWorkflow(&config, payload)

			notifyStatusErr := s.GithubClient.CreateStatus(&config, statusWH)

			if notifyStatusErr != nil {
				return nil, apierrors.NewInternalServerApiError(notifyStatusErr.Message(), notifyStatusErr)
			}

		default:
			return nil, apierrors.NewBadRequestApiError("Action not supported yet")
		}

	} else { //If webhook already exists then return it
		return nil, apierrors.NewConflictApiError("Resource Already exists")
	}

	return &wh, nil
}

//ProcessPullRequestWebhook process
func (s *Webhook) ProcessPullRequestReviewWebhook(payload *webhook.PullRequestReviewWebhook) (*webhook.Webhook, apierrors.ApiError) {

	var wh webhook.Webhook

	webhookType := "pull_request_review"

	//Build a ID to identify a unique webhook
	prWHBaseID := payload.Repository.FullName + payload.PullRequest.Head.Sha + webhookType + payload.Review.State
	prWebhookID := utils.Stringify(utils.GetMD5Hash(prWHBaseID))

	switch payload.Action {
	case pullRequestReviewSubmittedAction:
		//If the revision was approved. We must keep in the database
		if payload.Review.State == approvedPullRequestReviewState {

			//Search the status webhook into database
			if err := s.SQL.GetBy(&wh, "id = ?", &prWebhookID); err != nil {

				//If the error is not a not found error, then there is a problem
				if err != gorm.ErrRecordNotFound {
					return nil, apierrors.NewNotFoundApiError("error checking status webhook existence")
				}

				//Fill every field in the webhook
				wh.ID = prWebhookID
				//wh.GithubDeliveryID = utils.Stringify(ctx.GetHeader("X-GitHub-Delivery"))
				wh.Type = utils.Stringify(webhookType)
				wh.GithubRepositoryName = utils.Stringify(payload.Repository.FullName)
				wh.SenderName = utils.Stringify(payload.Sender.Login)
				wh.State = utils.Stringify(payload.Review.State)
				wh.Sha = utils.Stringify(payload.PullRequest.Head.Sha)
				wh.Description = utils.Stringify(payload.Review.Body)

				//Save it into database
				if err := s.SQL.Insert(&wh); err != nil {
					return nil, apierrors.NewInternalServerApiError("error saving new pull request review webhook", err)
				}

				return &wh, nil

			} else { //If webhook already exists then return it
				//Returns the saved webhook
				return &wh, nil
			}
		}

		return nil, apierrors.NewBadRequestApiError("pull request review state not supported yet")

	case pullRequestReviewDismissedAction:
		//Search the status webhook into database
		if err := s.SQL.GetBy(&wh, "id = ?", &prWebhookID); err != nil {
			//If the error is not a not found error, then there is a problem
			if err == gorm.ErrRecordNotFound {
				return nil, apierrors.NewNotFoundApiError("webhook not found")
			} else {
				return nil, apierrors.NewNotFoundApiError("error checking status webhook existence")
			}

		} else {
			//Delete the value from DB
			if err := s.SQL.Delete(&wh); err != nil {
				return nil, apierrors.NewInternalServerApiError("error saving new pull request review webhook", err)
			}
		}
	default:
		return nil, apierrors.NewBadRequestApiError("action not supported yet")
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
