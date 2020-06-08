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
		//dbProdHost := os.Getenv("CLEARDB_DATABASE_URL")
		return []interface{}{os.Getenv("DBUSER"), os.Getenv("DBPASS"), os.Getenv("DBHOST"), os.Getenv("DBNAME")}
	case "stage":
		return []interface{}{dbStageUser, dbStagePassword, dbStageHost, dbStageName}
	default:
		return []interface{}{dbLocalUser, dbLocalPassword, dbLocalHost, dbLocalName}
	}
}
