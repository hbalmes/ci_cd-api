package services

import (
	"errors"
	"fmt"
	"github.com/hbalmes/ci_cd-api/api/clients"
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/services/storage"
	"github.com/hbalmes/ci_cd-api/api/utils"
	"github.com/hbalmes/ci_cd-api/api/utils/apierrors"
	"github.com/jinzhu/gorm"
	"os"
)

//ConfigurationService is an interface which represents the ConfigurationService for testing purpose.
type ConfigurationService interface {
	Create(r *models.PostRequestPayload) (*models.Configuration, apierrors.ApiError)
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
func (s *Configuration) Create(r *models.PostRequestPayload) (*models.Configuration, apierrors.ApiError) {

	config := *models.NewConfiguration(r)
	config.ID = utils.Stringify(fmt.Sprintf("%s/%s", *r.Repository.Owner, *r.Repository.Name))

	var cf models.Configuration

	//Search the configuration into database
	if err := s.SQL.GetBy(&cf, "id = ?", fmt.Sprintf("%s/%s", *config.RepositoryOwner, *config.RepositoryName)); err != nil {

		//If the error is not a not found error, then there is a problem
		if err != gorm.ErrRecordNotFound {
			return nil, apierrors.NewInternalServerApiError("error checking configuration existence", err)
		}

		setWorkflowError := s.SetWorkflow(&config)

		if setWorkflowError != nil {
			return nil, setWorkflowError
		}

		//Save it into database
		if err := s.SQL.Insert(&config); err != nil {
			return nil, apierrors.NewInternalServerApiError("error saving new configuration", err)
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
	//To wakeup the db.
	scope := os.Getenv("SCOPE")
	if scope == "production" {
		for i := 1; i < 15; i++ {
			if err := s.SQL.GetBy(&cf, "id = ?", "hbalmes/ci_cd-api"); err != nil {
				if err != gorm.ErrRecordNotFound {
					continue
				}
			}
			break
		}
	}

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

	unsetWorkflowError := s.UnsetWorkflow(cf)

	if unsetWorkflowError != nil {
		return unsetWorkflowError
	}

	//Delete from configurations DB
	if sqlErr := s.SQL.Delete(cf); sqlErr != nil {
		return sqlErr
	}

	//TODO:Descomentar esto y probar.
	//Delete from configurations DB
	/*if sqlErr := s.SQL.DeleteFromRequireStatusChecksByConfigurationID(cf.ID); sqlErr != nil {
		return sqlErr
	}*/

	return nil
}
