package models

import (
	"time"
)

//PostRequestPayload represents the payload received in the POST request.
type PostRequestPayload struct {
	Repository struct {
		Name                *string  `json:"name"`
		Owner               *string  `json:"owner"`
		RequireStatusChecks []string `json:"required_status_checks"`
	} `json:"repository"`

	Workflow struct {
		Type *string `json:"type"`
	} `json:"workflow"`

	CodeCoverage struct {
		PullRequestThreshold *float64 `json:"pull_request_threshold"`
	} `json:"code_coverage"`
}

//PutRequestPayload represents the payload received in the PUT request.
type PutRequestPayload struct {
	Repository struct {
		Name                *string
		RequireStatusChecks []string `json:"required_status_checks"`
	} `json:"repository"`

	CodeCoverage struct {
		PullRequestThreshold *float64 `json:"pull_request_threshold"`
	} `json:"code_coverage"`
}

//Configuration represents the only business object of this API.
//Has all the information needed for a good release process execution.
type Configuration struct {
	ID                               *string `gorm:"primary_key"`
	RepositoryName                   *string
	RepositoryOwner                  *string
	RepositoryStatusChecks           []RequireStatusCheck
	WorkflowType                     *string
	CodeCoveragePullRequestThreshold *float64

	//GORM date attributes
	CreatedAt time.Time
	UpdatedAt time.Time
}

//RequireStatusCheck is a list of strings which represents all the status checks
//Required to be success before perform a 'git merge' in a protected branch.
type RequireStatusCheck struct {
	ID              *uint64 `gorm:"primary_key"`
	Check           string
	ConfigurationID *string
}

//NewConfiguration converts a PostRequestPayload into a Configuration.
func NewConfiguration(r *PostRequestPayload) *Configuration {
	var c Configuration

	c.ID = r.Repository.Name
	c.RepositoryName = r.Repository.Name
	c.RepositoryOwner = r.Repository.Owner
	c.WorkflowType = r.Workflow.Type
	c.CodeCoveragePullRequestThreshold = r.CodeCoverage.PullRequestThreshold

	reqChecks := make([]RequireStatusCheck, 0)
	for _, rq := range r.Repository.RequireStatusChecks {
		reqChecks = append(reqChecks, RequireStatusCheck{
			Check: rq,
		})
	}

	c.RepositoryStatusChecks = reqChecks

	return &c
}

//UpdateConfiguration updates a Configuration based on a PutRequestPayload.
func (c *Configuration) UpdateConfiguration(r *PutRequestPayload) {
	if r.CodeCoverage.PullRequestThreshold != nil {
		c.CodeCoveragePullRequestThreshold = r.CodeCoverage.PullRequestThreshold
	}

	if r.Repository.RequireStatusChecks != nil {
		reqChecks := make([]RequireStatusCheck, 0)
		for _, rq := range r.Repository.RequireStatusChecks {
			reqChecks = append(reqChecks, RequireStatusCheck{
				Check: rq,
			})
		}
		c.RepositoryStatusChecks = reqChecks
	}
}

//GetRequiredStatusCheck maps the RepositoryStatusChecks field in the Configuration struct into a string slice.
func (c *Configuration) GetRequiredStatusCheck() []string {
	var rsc []string
	for _, rc := range c.RepositoryStatusChecks {
		rsc = append(rsc, rc.Check)
	}
	return rsc
}

//Marshall converts the Configuration struct into a readable JSON interface.
func (c *Configuration) Marshall() interface{} {
	rsc := c.GetRequiredStatusCheck()
	return &struct {
		ID         string `json:"id"`
		Repository struct {
			Name                string   `json:"name"`
			Owner               string   `json:"owner"`
			RequiredStatusCheck []string `json:"required_status_check"`
		} `json:"repository"`
		CodeCoverage struct {
			PullRequestThreshold float64 `json:"pull_request_threshold"`
		} `json:"code_coverage"`
		Workflow struct {
			Type string `json:"type"`
		} `json:"workflow"`
	}{
		*c.ID,
		struct {
			Name                string   `json:"name"`
			Owner               string   `json:"owner"`
			RequiredStatusCheck []string `json:"required_status_check"`
		}{
			*c.RepositoryName,
			*c.RepositoryOwner,
			rsc,
		},
		struct {
			PullRequestThreshold float64 `json:"pull_request_threshold"`
		}{
			*c.CodeCoveragePullRequestThreshold,
		},
		struct {
			Type string `json:"type"`
		}{
			*c.WorkflowType,
		},
	}
}
