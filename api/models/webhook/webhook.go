package webhook

import (
	"time"
)

//Webhook represents a Github Webhook
//The ID is only for our internal systems. It's not the webhook ID given by GitHub
type Webhook struct {
	ID                      *string `gorm:"primary_key"`
	Type                    *string
	GithubDeliveryID        *string
	GithubRepositoryName    *string `gorm:"index:repository"`
	GithubPullRequestNumber *int
	Sha                     *string
	Context                 *string
	State                   *string
	Description             *string
	SenderName              *string
	WebhookCreateAt         time.Time
	WebhookUpdated          time.Time

	//GORM date attributes
	CreatedAt time.Time
	UpdatedAt time.Time
}

//GithubPayload represents the content of a Github Webhook
type GithubWebhookStandardPayload struct {

	//Repository represents the application repository info
	Repository *struct {
		FullName *string `json:"full_name"`
		Name     *string `json:"name"`
	} `json:"repository"`
}

//Marshall converts the Configuration struct into a readable JSON interface.
func (c *Webhook) Marshall() interface{} {
	return &struct {
		ID                      *string   `json:"id"`
		Type                    *string   `json:"type"`
		GithubDeliveryID        *string   `json:"github_delivery_id"`
		GithubRepositoryName    *string   `json:"github_repository_name"`
		GithubPullRequestNumber *int      `json:"github_pull_request_number"`
		Sha                     *string   `json:"sha"`
		Context                 *string   `json:"context"`
		State                   *string   `json:"state"`
		Description             *string   `json:"description"`
		SenderName              *string   `json:"sender_name"`
		WebhookCreateAt         time.Time `json:"webhook_create_at"`
		WebhookUpdated          time.Time `json:"webhook_updated"`
		CreatedAt               time.Time `json:"created_at"`
		UpdatedAt               time.Time `json:"update_at"`
	}{
		c.ID,
		c.Type,
		c.GithubDeliveryID,
		c.GithubRepositoryName,
		c.GithubPullRequestNumber,
		c.Sha,
		c.Context,
		c.State,
		c.Description,
		c.SenderName,
		c.WebhookCreateAt,
		c.WebhookUpdated,
		c.CreatedAt,
		c.UpdatedAt,
	}
}
