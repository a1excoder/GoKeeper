package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
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
		if len(os.Args) < 5 {
			ClientErrorMsg(fmt.Errorf("enter after bin name and mode flag auth mode and user name with password (./GoKeeper -c -a login password)"))
		}

		f, err := GetConfigFileData("config.json")
		if err != nil {
			ClientErrorMsg(err)
		}

		if os.Args[2] == "-a" {
			user := User{UserName: os.Args[3], Password: os.Args[4]}

			conn, err := user.ConnectToServer(f.Host, f.Port, AuthT)
			if err != nil {
				ClientErrorMsg(err)
			}

			MsgManager(conn)
		}

		if os.Args[2] == "-r" {
			user := User{UserName: os.Args[3], Password: os.Args[4]}

			conn, err := user.ConnectToServer(f.Host, f.Port, RegT)
			if err != nil {
				ClientErrorMsg(err)
			}

			MsgManager(conn)
		}

		ClientErrorMsg(fmt.Errorf("unknown flag of auth type"))
	case "--help":
		fmt.Println("enter after bin name and mode flag auth mode and user name with password (./GoKeeper -c -a login password)")
		os.Exit(1)
	default:
		ClientErrorMsg(fmt.Errorf("unknown flag"))
	}

}

func ClientErrorMsg(err error) {
	fmt.Println(err)
	os.Exit(1)
}

func ScanString(text string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(text)
	message, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return message[:len(message)-2], nil
}

func (note *Note) ViewNote() {
	fmt.Printf("id: %d\ntitle: %s\n", note.Id, note.Title)
	fmt.Printf("query: %s\n", note.Data)
}

func MsgManager(conn net.Conn) {
	var str string
	var err error
	var note Note

	for {
		str, err = ScanString(">>> ")
		if err != nil {
			ClientErrorMsg(err)
		}

		switch str {
		case "add":

			str, err = ScanString("enter title: ")
			if err != nil {
				ClientErrorMsg(err)
			}
			note.Title = str

			str, err = ScanString("enter query: ")
			if err != nil {
				ClientErrorMsg(err)
			}
			note.Data = str

			if err = CreateNote(conn, note); err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("note has been added")
		case "get":
			str, err = ScanString("enter note id: ")
			if err != nil {
				ClientErrorMsg(err)
			}

			if note.Id, err = strconv.Atoi(str); err != nil {
				ClientErrorMsg(err)
			}

			note_ptr, err := GetNote(conn, note)
			if err != nil {
				fmt.Println(err)
				continue
			}

			note_ptr.ViewNote()
		case "get all":
			notes, err := GetAllNotes(conn)
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, note := range notes {
				note.ViewNote()
				fmt.Println()
			}
		case "get by title":
			if note.Title, err = ScanString("Enter title: "); err != nil {
				ClientErrorMsg(err)
			}

			notes, err := GetAllNotesByTitle(conn, note)
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, note := range notes {
				note.ViewNote()
				fmt.Println()
			}
		case "delete":
			str, err = ScanString("enter note id: ")
			if err != nil {
				ClientErrorMsg(err)
			}

			if note.Id, err = strconv.Atoi(str); err != nil {
				ClientErrorMsg(err)
			}

			if err = DeleteNote(conn, note); err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("note was deleted")
		case "update":
			if str, err = ScanString("enter note id: "); err != nil {
				ClientErrorMsg(err)
			}
			if note.Id, err = strconv.Atoi(str); err != nil {
				ClientErrorMsg(err)
			}

			// what_update, err := ScanString("what do you update? 1(only title) | 2(only query) | 12(this and that): ")
			// if err != nil {
			// 	ClientErrorMsg(err)
			// }
			// status, err := strconv.Atoi(what_update)
			// if err != nil {
			// 	ClientErrorMsg(err)
			// }

			// switch status {
			// case 1:
			// 	if str, err = ScanString("enter new title: "); err != nil {
			// 		ClientErrorMsg(err)
			// 	}

			// }

			if note.Title, err = ScanString("enter new title: "); err != nil {
				ClientErrorMsg(err)
			}

			if note.Data, err = ScanString("enter new query: "); err != nil {
				ClientErrorMsg(err)
			}

			if err = UpdateNote(conn, note); err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("note was updated")
		case "help":
			fmt.Println("add(create new note)")
			fmt.Println("update(update note)")
			fmt.Println("delete(delete note)")
			fmt.Println("get(get note by id)")
			fmt.Println("get all(get all notes)")
			fmt.Println("get by title(get all notes by title)")
			fmt.Println("quit(quit from application)")
		case "quit":
			conn.Close()
			os.Exit(0)
		}
	}
}
