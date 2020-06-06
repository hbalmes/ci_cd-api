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
	dbProductionUser     = "bde563f654d6b4"
	dbProductionPassword = "dc427777"
	dbProductionName     = "heroku_56d233420e055a9"
	dbProductionHost     = "us-cdbr-east-05.cleardb.net"
)

func GetDBConnectionParams() []interface{} {
	switch scope := os.Getenv("SCOPE"); scope {
	case "production", "test":
		//dbProdHost := os.Getenv("CLEARDB_DATABASE_URL")
		return []interface{}{dbProductionUser, dbProductionPassword, dbProductionHost, dbProductionName}
	case "stage":
		return []interface{}{dbStageUser, dbStagePassword, dbStageHost, dbStageName}
	default:
		return []interface{}{dbLocalUser, dbLocalPassword, dbLocalHost, dbLocalName}
	}
}
