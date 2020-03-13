package services

import (
	"errors"
	"github.com/herbal828/ci_cd-api/api/clients"
	"github.com/herbal828/ci_cd-api/api/models"
	"github.com/herbal828/ci_cd-api/api/services/storage"
	"github.com/jinzhu/gorm"
)

//ConfigurationService is an interface which represents the ConfigurationService for testing purpose.
type ConfigurationService interface {
	Create(*models.PostRequestPayload) (*models.Configuration, error)
	Get(string) (*models.Configuration, error)
	Update(r *models.PutRequestPayload) (*models.Configuration, error)
	Delete(id string) error
}

//Configuration represents the ConfigurationService layer
//It has an instance of a DBClient layer and
//A github client instance
type Configuration struct {
	SQL          storage.SQLStorage
	GithubClient clients.GithubClient
}

//NewConfigurationService initializes a ConfigurationService
func NewConfigurationService(sql storage.SQLStorage) *Configuration {
	return &Configuration{
		SQL:          sql,
		GithubClient: clients.NewGithubClient(),
	}
}

//Create creates a Release Process valid configuration.
//It performs all the actions needed to enabled successfuly Release Process.
func (s *Configuration) Create(r *models.PostRequestPayload) (*models.Configuration, error) {

	config := *models.NewConfiguration(r)

	repoName := config.ID
	var cf models.Configuration

	//Search the configuration into database
	if err := s.SQL.GetBy(&cf, "id = ?", *repoName); err != nil {

		//If the error is not a not found error, then there is a problem
		if err != gorm.ErrRecordNotFound {
			return nil, errors.New("error checking configuration existence")
		}

		setWorkflowError := s.SetWorkflow(&config)

		if setWorkflowError != nil {
			return nil, setWorkflowError
		}

		//Save it into database
		if err := s.SQL.Insert(&config); err != nil {
			return nil, errors.New("error saving new configuration")
		}
		return &config, nil

	} else { //If configuration already exists then return it
		return &cf, nil
	}
}

//Get searches a configuration into database.
//Returns an error if the config is not found.
func (s *Configuration) Get(id string) (*models.Configuration, error) {
	var cf models.Configuration
	if err := s.SQL.GetBy(&cf, "id = ?", id); err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, errors.New("error checking configuration existance")
		}
		return nil, err
	}
	return &cf, nil
}

//Update modifies a configuration.
//It receives a PutRequestPayload.
//Returns an error if the config is not found or if it some problem updating the config.
func (s *Configuration) Update(r *models.PutRequestPayload) (*models.Configuration, error) {

	oldConfig, err := s.Get(*r.Repository.Name)

	if err != nil {
		return nil, err
	}

	newConfig := *oldConfig
	newConfig.UpdateConfiguration(r)

	//Update the repository status checks
	if r.Repository.RequireStatusChecks != nil {
		//TODO: Cambiar la proteccion con los nuevos status

		//TODO: Change this, because it is a change made in order to be able to update the required status checks
		//we did this because when we updated the fields, it doesn't update them in the require_status_check
		// child table, so we removed them and then saved the new ones.

		//Delete from configurations DB
		if sqlErr := s.SQL.DeleteFromRequireStatusChecksByConfigurationID(oldConfig.ID); sqlErr != nil {
			return nil, sqlErr
		}

	}

	//Save the new config into database
	if err := s.SQL.Update(&newConfig); err != nil {
		return nil, errors.New("error updating repository configuration")
	}
	return &newConfig, nil
}

//Delete erase the configuration.
//It makes a sof delete.
//Receives the configuration id (repoName) and returns an error it it occurs.
func (s *Configuration) Delete(id string) error {

	cf, err := s.Get(id)

	if err != nil {
		return err
	}

	//Unset Workflow
	//TODO: Desproteger de acuerdo al wf que tiene configurado

	//Delete from configurations DB
	if sqlErr := s.SQL.Delete(cf); sqlErr != nil {
		return sqlErr
	}

	return nil
}
