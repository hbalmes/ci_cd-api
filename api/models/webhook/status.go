package webhook

import "time"

type Status struct {
	ID          int64     `json:"id"`
	Sha         string    `json:"sha"`
	Name        string    `json:"name"`
	Context     string    `json:"context"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Repository  struct {
		ID       int    `json:"id"`
		FullName string `json:"full_name"`
	} `json:"repository"`
	Sender struct {
		Login string `json:"login"`
	} `json:"sender"`
}
