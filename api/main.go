package main

import (
	"fmt"
	"github.com/herbal828/ci_cd-api/api/controllers/routers"
	"github.com/herbal828/ci_cd-api/api/models"
	"github.com/herbal828/ci_cd-api/api/models/webhook"
	"github.com/herbal828/ci_cd-api/api/services/storage"
	_ "github.com/jinzhu/gorm/dialects/mysql"
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

func main() {
	sql, err := storage.NewMySQL()
	defer sql.Client.Close()
	//Something was wrong stablishing the database connection
	if err != nil {
		fmt.Println("There was an error stablishing the MySQL connection")
	}

	sql.Client.AutoMigrate(&models.Configuration{}, &models.RequireStatusCheck{}, &webhook.Webhook{}, &webhook.PullRequest{})

	routers.SQLConnection = sql

	router := routers.Route()
	//Init GinGonic server
	router.Run(":8080")
}
