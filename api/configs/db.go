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

const (
	dbProductionUser     = "root"
	dbProductionPassword = "123456"
	dbProductionName     = "configurations"
)

func GetDBConnectionParams() []interface{} {
	switch scope := os.Getenv("SCOPE"); scope {
	case "production", "test":
		dbProdHost := os.Getenv("CLEARDB_DATABASE_URL")
		return []interface{}{dbProductionUser, dbProductionPassword, dbProdHost, dbProductionName}
	case "stage":
		return []interface{}{dbStageUser, dbStagePassword, dbStageHost, dbStageName}
	default:
		return []interface{}{dbLocalUser, dbLocalPassword, dbLocalHost, dbLocalName}
	}
}
