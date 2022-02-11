package main

import (
	"log"
)

func main() {
	db, err := CreateConn("sqlite3", "notes.db")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	f, err := GetConfigFileData("config.json")
	if err != nil {
		log.Fatalln(err)
	}

	StartRoutineServer(f.Host, f.Port, int(f.MaxConn), db)
}
