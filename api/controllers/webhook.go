package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/herbal828/ci_cd-api/api/services"
	"github.com/herbal828/ci_cd-api/api/services/storage"
	"github.com/herbal828/ci_cd-api/api/utils"
	"github.com/herbal828/ci_cd-api/api/utils/apierrors"
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
	if event, deliveryID := getGetGithubHeaders(ginContext); event != "" && deliveryID != "" {

		whook, createWHErr := c.Service.CreateWebhook(ginContext, event)
		if createWHErr != nil {
			ginContext.JSON(
				createWHErr.Status(),
				createWHErr,
			)
			return
		}

		ginContext.JSON(http.StatusOK, whook.Marshall())

	} else {
		ginContext.JSON(
			http.StatusBadRequest,
			apierrors.NewBadRequestApiError("invalid headers"),
		)
	}

}

func getGetGithubHeaders(context utils.HTTPContext) (string, string) {
	ghEvent := context.GetHeader(ghEventHeader)
	ghDeliveryID := context.GetHeader(ghDeliveryIDHeader)

	return ghEvent, ghDeliveryID
}
