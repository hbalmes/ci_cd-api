package webhook

import "time"

type PullRequestWebhook struct {
	Action      string `json:"action"`
	Number      int    `json:"number"`
	PullRequest struct {
		URL    string `json:"url"`
		ID     int64  `json:"id"`
		Number int    `json:"number"`
		State  string `json:"state"`
		Locked bool   `json:"locked"`
		Title  string `json:"title"`
		User   struct {
			Login string `json:"login"`
		} `json:"user"`
		Body               string        `json:"body"`
		CreatedAt          time.Time     `json:"created_at"`
		UpdatedAt          time.Time     `json:"updated_at"`
		ClosedAt           interface{}   `json:"closed_at"`
		MergedAt           interface{}   `json:"merged_at"`
		MergeCommitSha     interface{}   `json:"merge_commit_sha"`
		Assignee           interface{}   `json:"assignee"`
		Assignees          []interface{} `json:"assignees"`
		RequestedReviewers []interface{} `json:"requested_reviewers"`
		RequestedTeams     []interface{} `json:"requested_teams"`
		Labels             []interface{} `json:"labels"`
		Milestone          interface{}   `json:"milestone"`
		Head               struct {
			Label string `json:"label"`
			Ref   string `json:"ref"`
			Sha   string `json:"sha"`
			User  struct {
				Login string `json:"login"`
				ID    int    `json:"id"`
			} `json:"user"`
			Repo struct {
				ID       int    `json:"id"`
				NodeID   string `json:"node_id"`
				Name     string `json:"name"`
				FullName string `json:"full_name"`
			} `json:"repo"`
		} `json:"head"`
		Base struct {
			Label string `json:"label"`
			Ref   string `json:"ref"`
			Sha   string `json:"sha"`
			User  struct {
				Login string `json:"login"`
				ID    int    `json:"id"`
			} `json:"user"`
			Repo struct {
				ID       int    `json:"id"`
				NodeID   string `json:"node_id"`
				Name     string `json:"name"`
				FullName string `json:"full_name"`
			} `json:"repo"`
			CreatedAt       time.Time   `json:"created_at"`
			UpdatedAt       time.Time   `json:"updated_at"`
			PushedAt        time.Time   `json:"pushed_at"`
			GitURL          string      `json:"git_url"`
			SSHURL          string      `json:"ssh_url"`
			CloneURL        string      `json:"clone_url"`
			SvnURL          string      `json:"svn_url"`
			Homepage        interface{} `json:"homepage"`
			Size            int         `json:"size"`
			StargazersCount int         `json:"stargazers_count"`
			WatchersCount   int         `json:"watchers_count"`
			Language        interface{} `json:"language"`
			HasIssues       bool        `json:"has_issues"`
			HasProjects     bool        `json:"has_projects"`
			HasDownloads    bool        `json:"has_downloads"`
			HasWiki         bool        `json:"has_wiki"`
			HasPages        bool        `json:"has_pages"`
			ForksCount      int         `json:"forks_count"`
			MirrorURL       interface{} `json:"mirror_url"`
			Archived        bool        `json:"archived"`
			Disabled        bool        `json:"disabled"`
			OpenIssuesCount int         `json:"open_issues_count"`
			License         interface{} `json:"license"`
			Forks           int         `json:"forks"`
			OpenIssues      int         `json:"open_issues"`
			Watchers        int         `json:"watchers"`
			DefaultBranch   string      `json:"default_branch"`
		} `json:"base"`
	} `json:"pull_request"`
	AuthorAssociation   string      `json:"author_association"`
	Draft               bool        `json:"draft"`
	Merged              bool        `json:"merged"`
	Mergeable           interface{} `json:"mergeable"`
	Rebaseable          interface{} `json:"rebaseable"`
	MergeableState      string      `json:"mergeable_state"`
	MergedBy            interface{} `json:"merged_by"`
	Comments            int         `json:"comments"`
	ReviewComments      int         `json:"review_comments"`
	MaintainerCanModify bool        `json:"maintainer_can_modify"`
	Commits             int         `json:"commits"`
	Additions           int         `json:"additions"`
	Deletions           int         `json:"deletions"`
	ChangedFiles        int         `json:"changed_files"`
	Repository          struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
		ID    int    `json:"id"`
	} `json:"sender"`
}

type PullRequest struct {
	ID                *int64     `gorm:"primary_key"`
	PullRequestNumber *int       `json:"pull_request_number"`
	State             *string    `json:"state"`
	RepositoryName    *string    `json:"repository_name"`
	BaseRef           *string    `json:"base_ref"`
	HeadRef           *string    `json:"head_ref"`
	BaseSha           *string    `json:"base_sha"`
	HeadSha           *string    `json:"head_sha"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Body              *string    `json:"body"`
	Title             *string    `json:"title"`
	CreatedBy         *string    `json:"created_by"`
}
