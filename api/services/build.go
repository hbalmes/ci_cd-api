package services

import (
	"github.com/hbalmes/ci_cd-api/api/services/storage"
)

type buildService interface {
	//CheckBuildability
}

//Build represents the BuildService layer
//It has an instance of a DBClient layer and
//A Webhook service instance and
//A ConfigService instance
type Build struct {
	SQL           storage.SQLStorage
	Webhook       WebhookService
	ConfigService ConfigurationService
}

//NewConfigurationSeNewWebhookServicervice initializes a WebhookService
func NewBuildService(sql storage.SQLStorage) *Build {
	return &Build{
		SQL:           sql,
		Webhook: NewWebhookService(sql),
		ConfigService: NewConfigurationService(sql),
	}
}

/*func (s *Build) CheckBuildability() (*webhook.Webhook, apierrors.ApiError) {


}*/


