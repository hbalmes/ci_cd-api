package configs

import (
	"os"
)

const (
	dbStageUser     = ""
	dbStagePassword = ""
	dbStageHost     = ""
	dbStageName     = ""
)

const (
	dbLocalUser     = "root"
	dbLocalPassword = "123456"
	dbLocalHost     = "localhost:3306"
	dbLocalName     = "configurations"
)

func GetDBConnectionParams() []interface{} {
	switch scope := os.Getenv("SCOPE"); scope {
	case "production", "test":
		//return []interface{}{dbProductionUser, dbProductionPassword, dbProductionHost, dbProductionName}
		//TODO: Cambiar a los valores
		return []interface{}{dbStageUser, dbStagePassword, dbStageHost, dbStageName}
	case "stage":
		return []interface{}{dbStageUser, dbStagePassword, dbStageHost, dbStageName}
	default:
		return []interface{}{dbLocalUser, dbLocalPassword, dbLocalHost, dbLocalName}
	}
}
