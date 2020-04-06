package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/services"
	"github.com/hbalmes/ci_cd-api/api/services/storage"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"net/http"
)

const (
	ghEventHeader      = "X-Github-Event"
	ghDeliveryIDHeader = "X-GitHub-Delivery"
)

type Webhook struct {
	Service services.WebhookService
}

//NewWebhookController initializes a WebhookController
func NewWebhookController(sql storage.SQLStorage) *Webhook {
	return &Webhook{
		Service: services.NewWebhookService(sql),
	}
}

//Create creates a new github webhook for the given repository
//It could returns
//	201Created in case of a success processing the creation
//	400BadRequest in case of an error parsing the request payload
//	500InternalServerError in case of an internal error procesing the creation
func (c *Webhook) CreateWebhook(ginContext *gin.Context) {

	//Check if 'X-Github-Event' header is present
	if webhookEvent, deliveryID := getGetGithubHeaders(ginContext); webhookEvent != "" && deliveryID != "" {

		switch webhookEvent {
		case "status":
			var statusWH webhook.Status

			if err := ginContext.BindJSON(&statusWH); err != nil {
				ginContext.JSON(
					http.StatusBadRequest,
					apierrors.NewBadRequestApiError("invalid status webhook payload"),
				)
				return
			}

			whook, err := c.Service.ProcessStatusWebhook(&statusWH)

			if err != nil {
				ginContext.JSON(
					http.StatusInternalServerError,
					err,
				)
				return
			}

			ginContext.JSON(http.StatusOK, whook.Marshall())
			return

		case "pull_request_review":

			var pullRequestReviewWH webhook.PullRequestReviewWebhook

			if err := ginContext.BindJSON(&pullRequestReviewWH); err != nil {
				ginContext.JSON(
					http.StatusBadRequest,
					apierrors.NewBadRequestApiError("invalid pull request review webhook payload"),
				)
				return
			}

			whook, err := c.Service.ProcessPullRequestReviewWebhook(&pullRequestReviewWH)

			if err != nil {
				ginContext.JSON(
					http.StatusInternalServerError,
					err,
				)
				return
			}

			ginContext.JSON(http.StatusOK, whook.Marshall())
			return

		case "pull_request":

			var pullRequestWH webhook.PullRequestWebhook
			if err := ginContext.BindJSON(&pullRequestWH); err != nil {
				ginContext.JSON(
					http.StatusBadRequest,
					apierrors.NewBadRequestApiError("invalid pull_request webhook payload"),
				)
				return
			}

			whook, err := c.Service.ProcessPullRequestWebhook(&pullRequestWH)

			if err != nil {
				ginContext.JSON(
					http.StatusInternalServerError,
					err,
				)
				return
			}

			ginContext.JSON(http.StatusOK, whook.Marshall())
			return

		default:
			ginContext.JSON(
				http.StatusBadRequest,
				apierrors.NewBadRequestApiError("Event not supported yet"),
			)
			return
		}

	} else {
		ginContext.JSON(
			http.StatusBadRequest,
			apierrors.NewBadRequestApiError("invalid headers"),
		)
		return
	}
}

func getGetGithubHeaders(context utils.HTTPContext) (string, string) {
	ghEvent := context.GetHeader(ghEventHeader)
	ghDeliveryID := context.GetHeader(ghDeliveryIDHeader)

	return ghEvent, ghDeliveryID
}
