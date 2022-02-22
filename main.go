package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("enter after bin name flag < -s(server) | -c(client) >")
	}

	switch os.Args[1] {
	case "-s":
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
	case "-c":
		user := User{UserName: "Tester", Password: "5345234"}

		conn, err := user.ConnectToServer("127.0.0.1", "4444", AuthT)
		if err != nil {
			log.Fatalln(err)
		}

		notes, err := GetAllNotes(conn)
		if err != nil {
			log.Fatalln(err)
		}

		for key, note := range notes {
			fmt.Printf("%d) id(%d) title(%s) data(%s)\n", key, note.Id, note.Title, note.Data)
		}
	default:
		log.Fatalln("unknown flag")
	}

}
