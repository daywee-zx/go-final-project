package main

import (
	"database/sql"
	"go_final/internal/scheduler_db"
	"go_final/internal/server"
	"log"

	"os"
)

var (
	webDir = "./web"
)

const (
	portEnvVar     = "TODO_PORT"
	dbPathEnvVar   = "TODO_DBFILE"
	passwordEnvVar = "TODO_PASSWORD"
)

// InitEnv returns port, path to db file and password from environment.
// If not set, port = "7540"; dbPath = "scheduler.db"; password = "123456"
func InitEnv() (string, string, string) {
	port := os.Getenv(portEnvVar)
	if port == "" {
		port = "7540"
	}

	dbPath := os.Getenv(dbPathEnvVar)
	if dbPath == "" {
		dbPath = "scheduler.db"
	}

	password := os.Getenv(passwordEnvVar)
	if password == "" {
		password = "123456"
	}

	return port, dbPath, password
}

func main() {
	port, dbPath, password := InitEnv()

	logPath, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		logPath = os.Stdout
	}
	logs := log.New(logPath, "", log.LstdFlags)

	install := false
	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		install = true
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		logs.Fatal(err)
	}
	defer db.Close()

	schedulerDB := scheduler_db.New(db)
	if install {
		err = schedulerDB.Init()
		if err != nil {
			logs.Fatal(err)
		}
	}

	s := server.New(port, logs, schedulerDB, password)

	logs.Print("Starting server on port ", port)
	logs.Fatal(s.Init(webDir))
}
