package models

type Build struct {
	ID             uint32  `json:"id" gorm:"primary_key;AUTO_INCREMENT"`
	Sha            *string `json:"sha"`
	Major          uint8   `json:"major"`
	Minor          uint16  `json:"minor"`
	Patch          uint16  `json:"patch"`
	Tag            *string `json:"tag"`
	Status         *string `json:"status"`
	Branch         *string `json:"branch"`
	Username       *string `json:"username"`
	UpdatedAt      *string `json:"updated_at"`
	CreatedAt      *string `json:"created_at"`
	RepositoryName *string `json:"repository_name"`
	Type           *string `json:"type"`
	Body           *string `json:"body"`
	GithubID       *string `json:"github_id"`
	GithubURL      *string `json:"github_url"`
}
