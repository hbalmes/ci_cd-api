package services

import (
	"github.com/hbalmes/ci_cd-api/api/clients"
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/services/storage"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog/log"
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
	SQL           storage.SQLStorage
	GithubClient  clients.GithubClient
	ConfigService ConfigurationService
	BuildService  BuildService
}

//NewConfigurationSeNewWebhookServicervice initializes a WebhookService
func NewWebhookService(sql storage.SQLStorage) *Webhook {
	return &Webhook{
		SQL:           sql,
		GithubClient:  clients.NewGithubClient(),
		ConfigService: NewConfigurationService(sql),
		BuildService:  NewBuildService(sql),
	}
}

//ProcessStatusWebhook process
func (s *Webhook) ProcessStatusWebhook(payload *webhook.Status) (*webhook.Webhook, apierrors.ApiError) {

	var wh webhook.Webhook

	webhookType := "status"

	//Validates that the repository has a ci cd configuration
	conf, err := s.ConfigService.Get(*payload.Repository.FullName)

	if err != nil {
		return nil, apierrors.NewInternalServerApiError("error checking configuration existance", err)
	}

	if conf == nil {
		return nil, apierrors.NewNotFoundApiError("error getting application ci_cd configuration")
	}

	contextAllowed := utils.ContainsStatusChecks(conf.RepositoryStatusChecks, *payload.Context)

	if !contextAllowed {
		return nil, apierrors.NewBadRequestApiError("Context not configured for the repository")
	}

	//Build a ID to identify a unique webhook
	shBaseID := *payload.Repository.FullName + *payload.Sha + *payload.Context + *payload.State
	statusWebhookID := utils.Stringify(utils.GetMD5Hash(shBaseID))

	//Search the status webhook into database
	if err := s.SQL.GetBy(&wh, "id = ?", *statusWebhookID); err != nil {

		//If the error is not a not found error, then there is a problem
		if err != gorm.ErrRecordNotFound {
			return nil, apierrors.NewNotFoundApiError("error checking status webhook existence")
		}

		//Fill every field in the webhook
		wh.ID = statusWebhookID
		//wh.GithubDeliveryID = utils.Stringify(ctx.GetHeader("X-GitHub-Delivery"))
		wh.Type = utils.Stringify(webhookType)
		wh.GithubRepositoryName = payload.Repository.FullName
		wh.SenderName = payload.Sender.Login
		wh.WebhookCreateAt = payload.CreatedAt
		wh.WebhookUpdated = payload.UpdatedAt
		wh.State = payload.State
		wh.Context = payload.Context
		wh.Sha = payload.Sha
		wh.Description = payload.Description

		//Save it into database
		if err := s.SQL.Insert(&wh); err != nil {
			return nil, apierrors.NewInternalServerApiError("error saving new status webhook", err)
		}

		//TODO: Logear el build generado o insertarlo en el apartado build
		build, _ := s.BuildService.ProcessBuild(conf, payload)

		if build != nil {
			//TODO: Logear
		}

	} else { //If webhook already exists then return it
		return nil, apierrors.NewConflictApiError("Resource Already exists")
	}

	return &wh, nil
}

//ProcessPullRequestWebhook process
func (s *Webhook) ProcessPullRequestWebhook(payload *webhook.PullRequestWebhook) (*webhook.Webhook, apierrors.ApiError) {

	var prWH models.PullRequest
	var wh webhook.Webhook
	var cf Configuration

	//Validates that the repository has a ci cd configuration
	config, err := s.ConfigService.Get(*payload.Repository.FullName)

	if err != nil {
		return nil, apierrors.NewInternalServerApiError("error checking configuration existance", err)
	}

	if config == nil {
		return nil, apierrors.NewNotFoundApiError("configuration not found for the repository")
	}

	if payload.PullRequest.Base.Ref == nil || payload.PullRequest.Head.Ref == nil {
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
		whBaseID := *payload.Repository.FullName + *payload.PullRequest.Head.Sha + string(payload.PullRequest.ID) + *payload.PullRequest.State
		prWebhookID := utils.Stringify(utils.GetMD5Hash(whBaseID))

		//Fill every field in the webhook
		wh.ID = prWebhookID
		//wh.GithubDeliveryID = utils.Stringify(ctx.GetHeader("X-GitHub-Delivery"))
		wh.Type = utils.Stringify(webhookType)
		wh.GithubRepositoryName = payload.Repository.FullName
		wh.SenderName = payload.Sender.Login
		wh.WebhookCreateAt = payload.PullRequest.CreatedAt
		wh.WebhookUpdated = payload.PullRequest.UpdatedAt
		wh.State = payload.PullRequest.State
		wh.Sha = payload.PullRequest.Head.Sha
		wh.Description = payload.PullRequest.Body
		wh.GithubPullRequestNumber = &payload.PullRequest.Number

		//Save it into database
		if err := s.SQL.Insert(&wh); err != nil {
			return nil, apierrors.NewInternalServerApiError("error saving new pull request webhook", err)
		}

		switch *payload.Action {
		case "opened", "synchronize":

			statusWH := cf.CheckWorkflow(config, payload)

			notifyStatusErr := s.GithubClient.CreateStatus(config, statusWH)

			if notifyStatusErr != nil {
				return nil, apierrors.NewInternalServerApiError(notifyStatusErr.Message(), notifyStatusErr)
			}

		default:
			return nil, apierrors.NewBadRequestApiError("Action not supported yet")
		}

	} else {

		//TODO: mejorar este codigo.
		switch *payload.Action {
		case "synchronize":

			//Update the Pull request
			updateErr := s.UpdatePullRequestWebhook(*payload)

			if updateErr != nil {
				return nil, apierrors.NewInternalServerApiError(updateErr.Error(), updateErr)
			}

			statusWH := cf.CheckWorkflow(config, payload)

			notifyStatusErr := s.GithubClient.CreateStatus(config, statusWH)

			if notifyStatusErr != nil {
				return nil, apierrors.NewInternalServerApiError(notifyStatusErr.Message(), notifyStatusErr)
			}
		case "closed", "reopened":

			//Update the Pull request
			updateErr := s.UpdatePullRequestWebhook(*payload)

			if updateErr != nil {
				return nil, apierrors.NewInternalServerApiError(updateErr.Error(), updateErr)
			}

		default:
			return nil, apierrors.NewConflictApiError("Resource Already exists")
		}
	}

	return &wh, nil
}

//ProcessPullRequestWebhook process
func (s *Webhook) ProcessPullRequestReviewWebhook(payload *webhook.PullRequestReviewWebhook) (*webhook.Webhook, apierrors.ApiError) {
	log.Info().Str("action", *payload.Action).Str("state", *payload.Review.State).
		Str("repository", *payload.Repository.FullName).Msg("processing pull request review webhook")

	var wh webhook.Webhook

	//Validates that the repository has a ci cd configuration
	config, err := s.ConfigService.Get(*payload.Repository.FullName)

	if err != nil {
		log.Error().Err(err).Str("action", *payload.Action).Str("state", *payload.Review.State).
			Str("repository", *payload.Repository.FullName).Msg("error checking status webhook existence")
		return nil, apierrors.NewInternalServerApiError("error checking configuration existance", err)
	}

	if config == nil {
		log.Error().Err(err).Str("action", *payload.Action).Str("state", *payload.Review.State).
			Str("repository", *payload.Repository.FullName).Msg("configuration not found for the repository")
		return nil, apierrors.NewNotFoundApiError("configuration not found for the repository")
	}

	log.Info().Msgf("config getted successfully for %s", *payload.Repository.FullName)

	webhookType := utils.Stringify("pull_request_review")

	//Build a ID to identify a unique webhook
	prWHBaseID := *payload.Repository.FullName + *payload.PullRequest.Head.Sha + *webhookType + *payload.Review.State
	prWebhookID := utils.Stringify(utils.GetMD5Hash(prWHBaseID))

	switch *payload.Action {
	case pullRequestReviewSubmittedAction:
		//If the revision was approved. We must keep in the database
		if *payload.Review.State == approvedPullRequestReviewState {

			//Search the status webhook into database
			if err := s.SQL.GetBy(&wh, "id = ?", &prWebhookID); err != nil {

				//If the error is not a not found error, then there is a problem
				if err != gorm.ErrRecordNotFound {
					log.Error().Err(err).Str("action", *payload.Action).Str("state", *payload.Review.State).
						Str("repository", *payload.Repository.FullName).Msg("error checking status webhook existence")
					return nil, apierrors.NewNotFoundApiError("error checking status webhook existence")
				}

				//Fill every field in the webhook
				wh.ID = prWebhookID
				//wh.GithubDeliveryID = utils.Stringify(ctx.GetHeader("X-GitHub-Delivery"))
				wh.Type = webhookType
				wh.GithubRepositoryName = payload.Repository.FullName
				wh.SenderName = payload.Sender.Login
				wh.State = payload.Review.State
				wh.Sha = payload.PullRequest.Head.Sha
				wh.Description = payload.Review.Body
				wh.WebhookCreateAt = payload.PullRequest.CreatedAt
				wh.WebhookUpdated = payload.PullRequest.UpdatedAt

				//Save it into database
				if err := s.SQL.Insert(&wh); err != nil {
					log.Error().Err(err).Str("action", *payload.Action).Str("state", *payload.Review.State).
						Str("repository", *payload.Repository.FullName).Msg("error saving new pull request review webhook")
					return nil, apierrors.NewInternalServerApiError("error saving new pull request review webhook", err)
				}
			} else { //If webhook already exists then return it
				//Returns the saved webhook
				return &wh, nil
			}
		} else {
			log.Info().Str("action", *payload.Action).Str("state", *payload.Review.State).
				Str("repository", *payload.Repository.FullName).Msg("pull request review state not supported yet")
			return nil, apierrors.NewBadRequestApiError("pull request review state not supported yet")
		}

	case pullRequestReviewDismissedAction:
		//Search the status webhook into database
		if err := s.SQL.GetBy(&wh, "id = ?", &prWebhookID); err != nil {
			//If the error is not a not found error, then there is a problem
			if err == gorm.ErrRecordNotFound {
				log.Error().Err(err).Str("action", *payload.Action).Str("state", *payload.Review.State).
					Str("repository", *payload.Repository.FullName).Msg("webhook not found")
			} else {
				log.Error().Err(err).Str("action", *payload.Action).Str("state", *payload.Review.State).
					Str("repository", *payload.Repository.FullName).
					Msg("error checking status webhook existence")
				return nil, apierrors.NewNotFoundApiError("error checking status webhook existence")
			}

		} else {
			//Delete the value from DB
			if err := s.SQL.Delete(&wh); err != nil {
				log.Error().Err(err).Str("action", *payload.Action).Str("state", *payload.Review.State).
					Str("repository", *payload.Repository.FullName).
					Msg("error deleting pull request review webhook")
				return nil, apierrors.NewInternalServerApiError("error deleting pull request review webhook", err)
			}
		}
	default:
		log.Info().Msgf("pull request review webhook - action: %s", *payload.Action)
		return nil, apierrors.NewBadRequestApiError("action not supported yet")
	}

	//We create the payload necessary to process the build
	buildPayload := s.BuildStatusWebhookPayload(*payload)
	build, buildErr := s.BuildService.ProcessBuild(config, buildPayload)

	if buildErr != nil {
		log.Error().Err(err).Str("action", *payload.Action).Str("state", *payload.Review.State).
			Str("repository", *payload.Repository.FullName).Msg("error checking status webhook existence")
	}

	if build != nil {
		log.Info().Str("sha", *build.Sha).Str("type", *build.Type).Str("build", *build.GithubURL).
			Str("repository", *payload.Repository.FullName).Msg("build created successfully")
	}

	return &wh, nil
}

func (s *Webhook) SavePullRequestWebhook(pullRequestWH webhook.PullRequestWebhook) apierrors.ApiError {

	var prWH models.PullRequest

	//Fill every field in the pull request
	prWH.ID = pullRequestWH.PullRequest.ID
	prWH.PullRequestNumber = pullRequestWH.PullRequest.Number
	prWH.Body = pullRequestWH.PullRequest.Body
	prWH.State = pullRequestWH.PullRequest.State
	prWH.RepositoryName = pullRequestWH.Repository.FullName
	prWH.Title = pullRequestWH.PullRequest.Title
	prWH.BaseRef = pullRequestWH.PullRequest.Base.Ref
	prWH.BaseSha = pullRequestWH.PullRequest.Base.Sha
	prWH.HeadRef = pullRequestWH.PullRequest.Head.Ref
	prWH.HeadSha = pullRequestWH.PullRequest.Head.Sha
	prWH.CreatedAt = pullRequestWH.PullRequest.CreatedAt
	prWH.UpdatedAt = pullRequestWH.PullRequest.UpdatedAt
	prWH.CreatedBy = pullRequestWH.PullRequest.User.Login

	//Save it into database
	if err := s.SQL.Insert(&prWH); err != nil {
		return apierrors.NewInternalServerApiError("error saving new status webhook", err)
	}

	return nil
}

func (s *Webhook) UpdatePullRequestWebhook(pullRequestWH webhook.PullRequestWebhook) apierrors.ApiError {

	var prWH models.PullRequest

	//Fill every field in the pull request
	prWH.ID = pullRequestWH.PullRequest.ID
	prWH.PullRequestNumber = pullRequestWH.PullRequest.Number
	prWH.Body = pullRequestWH.PullRequest.Body
	prWH.State = pullRequestWH.Action
	prWH.RepositoryName = pullRequestWH.Repository.FullName
	prWH.Title = pullRequestWH.PullRequest.Title
	prWH.BaseRef = pullRequestWH.PullRequest.Base.Ref
	prWH.BaseSha = pullRequestWH.PullRequest.Base.Sha
	prWH.HeadRef = pullRequestWH.PullRequest.Head.Ref
	prWH.HeadSha = pullRequestWH.PullRequest.Head.Sha
	prWH.CreatedAt = pullRequestWH.PullRequest.CreatedAt
	prWH.UpdatedAt = pullRequestWH.PullRequest.UpdatedAt
	prWH.CreatedBy = pullRequestWH.PullRequest.User.Login

	//Save it into database
	if err := s.SQL.Update(&prWH); err != nil {
		return apierrors.NewInternalServerApiError("error updating new status webhook", err)
	}

	return nil
}

func (s *Webhook) BuildStatusWebhookPayload(pullRequestReviewWH webhook.PullRequestReviewWebhook) *webhook.Status {
	var statusWebhook webhook.Status

	//Fill every field in the status webhook with pullRequest Webhook
	statusWebhook.Sha = pullRequestReviewWH.PullRequest.Head.Sha
	statusWebhook.State = utils.Stringify("success")
	statusWebhook.Repository.FullName = pullRequestReviewWH.Repository.FullName
	statusWebhook.Sender.Login = pullRequestReviewWH.Sender.Login

	return &statusWebhook
}
