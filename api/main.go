package main

import (
	"fmt"
	"github.com/hbalmes/ci_cd-api/api/controllers/routers"
	"github.com/hbalmes/ci_cd-api/api/models"
	"github.com/hbalmes/ci_cd-api/api/models/webhook"
	"github.com/hbalmes/ci_cd-api/api/services/storage"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"os"
)

func init() {
	// We check if we're running in a TTY terminal to enable/disable output colors
	// This helps to avoid log pollution in non-interactive outputs such as
	// Jenkins or files
	//if !terminal.IsTerminal(int(os.Stdout.Fd())) {
	//If we're not running in a TTY terminal, we disable output colors entirely
	//	gin.DisableConsoleColor()
	//}
}

const defaultPort = ":8080"

func main() {
	sql, err := storage.NewMySQL()
	defer sql.Client.Close()
	//Something was wrong stablishing the database connection
	if err != nil {
		fmt.Println("There was an error stablishing the MySQL connection")
	}

	sql.Client.AutoMigrate(&models.Configuration{}, &models.RequireStatusCheck{}, &webhook.Webhook{}, &models.PullRequest{}, &models.Build{}, &models.LatestBuild{})

	routers.SQLConnection = sql

	router := routers.Route()
	//Init GinGonic server

	serverPort := os.Getenv("PORT")

	//Hack to heroku
	if serverPort == "" {
		serverPort = defaultPort
	}
	router.Run(":" + serverPort)
}
