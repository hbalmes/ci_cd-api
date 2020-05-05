package models

type Build struct {
	ID             int     `json:"id" gorm:"primary_key"`
	Version        *string `json:"version"`
	Status         *string `json:"status"`
	Branch         *string `json:"branch"`
	Username       *string `json:"username"`
	UpdatedAt      *string `json:"updated_at"`
	CreatedAt      *string `json:"created_at"`
	RepositoryName *string `json:"repository_name"`
	Type           *string  `json:"type"`
}
