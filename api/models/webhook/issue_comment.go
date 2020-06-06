package webhook

//IssueComment represents an issue comment Github Webhook
type IssueComment struct {
	Action  *string `json:"action"`
	Comment *struct {
		Body *string `json:"body"`
	} `json:"comment"`
}
