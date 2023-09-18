package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq" // PostgreSQL driver
)

type DatabaseConnector struct {
	Username string
	Password string
	Host     string
	Port     int
	DBName   string
}

func (dbConnector DatabaseConnector) toConnectString() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		dbConnector.Username,
		dbConnector.Password,
		dbConnector.Host,
		dbConnector.Port,
		dbConnector.DBName,
	)
}
func (dbConnector DatabaseConnector) Connect() (*sql.DB, error) {
	dbConfig := dbConnector.toConnectString()

	db, err := sql.Open("postgres", dbConfig)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
