package helper

import (
	"database/sql"
	"fmt"
)

func connect() *sql.DB {
	user := env("DB_USERNAME", "root")
	password := env("DB_PASSWORD", "secret")
	host := env("DB_HOST", "localhost")
	port := env("DB_PORT", "3306")
	database := env("DB_DATABASE", "")

	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, database)
	db, err := sql.Open("mysql", connection)
	checkError(err)
	return db
}
