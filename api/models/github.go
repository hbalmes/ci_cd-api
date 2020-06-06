package models

type BranchProtectionResponse struct {
	URL                  string `json:"url"`
	RequiredStatusChecks struct {
		URL         string   `json:"url"`
		Strict      bool     `json:"strict"`
		Contexts    []string `json:"contexts"`
		ContextsURL string   `json:"contexts_url"`
	} `json:"required_status_checks"`
	RequiredPullRequestReviews struct {
		URL                     string `json:"url"`
		DismissStaleReviews     bool   `json:"dismiss_stale_reviews"`
		RequireCodeOwnerReviews bool   `json:"require_code_owner_reviews"`
		DismissalRestrictions   struct {
			URL      string        `json:"url"`
			UsersURL string        `json:"users_url"`
			TeamsURL string        `json:"teams_url"`
			Users    []interface{} `json:"users"`
			Teams    []interface{} `json:"teams"`
		} `json:"dismissal_restrictions"`
	} `json:"required_pull_request_reviews"`
	EnforceAdmins struct {
		URL     string `json:"url"`
		Enabled bool   `json:"enabled"`
	} `json:"enforce_admins"`
	RequiredLinearHistory struct {
		Enabled bool `json:"enabled"`
	} `json:"required_linear_history"`
	AllowForcePushes struct {
		Enabled bool `json:"enabled"`
	} `json:"allow_force_pushes"`
	AllowDeletions struct {
		Enabled bool `json:"enabled"`
	} `json:"allow_deletions"`
}

type GetBranchResponse struct {
	Name   string `json:"name"`
	Commit struct {
		Sha string `json:"sha"`
	} `json:"commit"`
	Protected bool `json:"protected"`
}
