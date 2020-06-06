package webhook

import "time"

type PullRequestReviewWebhook struct {
	Action *string `json:"action"`
	Review struct {
		ID   int `json:"id"`
		User struct {
			Login *string `json:"login"`
			ID    int     `json:"id"`
		} `json:"user"`
		Body        *string `json:"body"`
		CommitID    *string `json:"commit_id"`
		SubmittedAt *string `json:"submitted_at"`
		State       *string `json:"state"`
	} `json:"review"`
	PullRequest struct {
		ID                 int           `json:"id"`
		Number             int           `json:"number"`
		State              *string       `json:"state"`
		Title              *string       `json:"title"`
		Body               *string       `json:"body"`
		CreatedAt          time.Time     `json:"created_at"`
		UpdatedAt          time.Time     `json:"updated_at"`
		ClosedAt           interface{}   `json:"closed_at"`
		MergedAt           interface{}   `json:"merged_at"`
		MergeCommitSha     *string       `json:"merge_commit_sha"`
		Assignee           interface{}   `json:"assignee"`
		Assignees          []interface{} `json:"assignees"`
		RequestedReviewers []interface{} `json:"requested_reviewers"`
		RequestedTeams     []interface{} `json:"requested_teams"`
		Head               struct {
			Label *string `json:"label"`
			Ref   *string `json:"ref"`
			Sha   *string `json:"sha"`
			User  struct {
				Login *string `json:"login"`
			} `json:"user"`
		} `json:"head"`
	} `json:"pull_request"`
	Sender struct {
		Login *string `json:"login"`
	} `json:"sender"`
	Repository struct {
		ID       int     `json:"id"`
		Name     *string `json:"name"`
		FullName *string `json:"full_name"`
		Owner    struct {
			Login *string `json:"login"`
			ID    int     `json:"id"`
		} `json:"owner"`
	} `json:"repository"`
}
