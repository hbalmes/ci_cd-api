package models

import "time"

type PullRequest struct {
	ID                int64     `gorm:"primary_key"`
	PullRequestNumber int
	State             *string
	RepositoryName    *string
	BaseRef           *string
	HeadRef           *string
	BaseSha           *string
	HeadSha           *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Body              *string
	Title             *string
	CreatedBy         *string
}
