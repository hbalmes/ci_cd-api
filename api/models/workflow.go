package models

type WorkflowConfig struct {
	Name          string `json:"name"`
	Description   Description
	Detail        string `json:"detail"`
	DefaultBranch string `json:"default_branch"`
}

type Description struct {
	Branches []Branch
}

type Branch struct {
	Requirements Requirements
	Stable       bool   `json:"stable"`
	Name         string `json:"name"`
	Releasable   bool   `json:"releaseable"`
	StartWith    bool   `json:"start_with"`
}

type Requirements struct {
	RequiredPullRequestReviews RequiredPullRequestReviews
	AcceptPrFrom               []string `json:"accept_pr_from"`
	RequiredStatusChecks       RequiredStatusChecks
	Restriction                interface{} `json:"restriction"`
	EnforceAdmins              bool        `json:"enforce_admins"`
}

type RequiredPullRequestReviews struct {
	DismissStaleReviews bool `json:"dismiss_stale_reviews"`
}

type RequiredStatusChecks struct {
	Contexts      []string `json:"contexts"`
	IncludeAdmins bool     `json:"include_admins"`
	Strict        bool     `json:"strict"`
}
