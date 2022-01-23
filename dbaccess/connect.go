package dbaccess

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func Connect() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "root@kus0c0de"
	dbServer := "192.168.71.36:3306"
	dbName := "picture"
	dbOption := "charset=utf8"
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", dbUser, dbPass, dbServer, dbName, dbOption)
	db, err := sql.Open(dbDriver, dataSourceName)
	if err != nil {
		fmt.Println(err)
	}
	return db
}
