package routers

import (
	"github.com/herbal828/ci_cd-api/api/controllers"
	"github.com/herbal828/ci_cd-api/api/services/storage"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	//SQLConnection is a stablished connection with the relational database
	SQLConnection *storage.SQL
)

//Route defines all the endpoints of this API.
func Route() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	ct := controllers.NewConfigurationController(SQLConnection)
	whct := controllers.NewWebhookController(SQLConnection)

	//POST to /configurations performs a release process configuration create
	r.POST("/configurations", func(c *gin.Context) {
		ct.Create(c)
	})

	//GET to /configurations/:repoName performs a release process configuration get
	r.GET("/configurations/:repoName", func(c *gin.Context) {
		ct.Show(c)
	})

	//PUT to /configurations/:repoName performs a release process configuration update
	r.PUT("/configurations/:repoName", func(c *gin.Context) {
		ct.Update(c)
	})

	//DELETE to /configurations/:repoName performs a release process configuration delete
	r.DELETE("/configurations/:repoName", func(c *gin.Context) {
		ct.Delete(c)
	})

	//POST to /webhooks get all the github webhooks
	r.POST("/webhooks", func(c *gin.Context) {
		whct.CreateWebhook(c)
	})

	return r
}
