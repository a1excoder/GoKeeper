package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/jmoiron/sqlx"
)

const (
	MinConn = 1
	MaxConn = 8
)

const (
	ErrorT             = 0
	SuccessT           = 1
	AuthT              = 2
	RegT               = 3
	NewNoteT           = 4
	GetNoteT           = 5
	UpdateNoteT        = 6
	DeleteNoteT        = 7
	GetAllMyNotesT     = 8
	GetLikeTitleNotesT = 9
	LogoutT            = 10
	GetCountAllMyNotes = 11
)

type MessageData struct {
	MessageTypeStatus int `json:"message_type_status"`
	Data              []byte
}

type ErrorMessageData struct {
	ErrorText string `json:"error_text"`
}

type NoteSliceData struct {
	Count int
	Notes []Note
}

func ClientMsgWorker(connection net.Conn, db *sqlx.DB, user *User) (bool, error) {
	msg := new(MessageData)
	note := new(Note)

	for {
		bytes, err := GetMessageData(connection)
		if err != nil {
			return true, err
		}

		if err = json.Unmarshal(bytes, &msg); err != nil {
			return true, err
		}

		switch msg.MessageTypeStatus {
		case NewNoteT:
			if err = json.Unmarshal(msg.Data, &note); err != nil {
				return true, err
			}

			if err = note.CreateNote(db, user); err != nil {
				return false, err
			}

			log.Printf("client(%s) note has been created\n", connection.RemoteAddr().String())
			if err = SendStatus(connection, SuccessT); err != nil {
				return true, err
			}
		case GetNoteT:
			if err = json.Unmarshal(msg.Data, &note); err != nil {
				return true, err
			}

			note, err = user.GetNoteById(db, note.Id)
			if err != nil {
				return false, err
			}

			note_data, err := json.Marshal(note)
			if err != nil {
				return true, err
			}

			msg.MessageTypeStatus = SuccessT
			msg.Data = note_data

			msg_data, err := json.Marshal(msg)
			if err != nil {
				return true, err
			}

			_, err = connection.Write(msg_data)
			if err != nil {
				return true, err
			}

			log.Printf("client(%s) note has been sent\n", connection.RemoteAddr().String())
		case UpdateNoteT:
			if err = json.Unmarshal(msg.Data, &note); err != nil {
				return true, err
			}

			if err = user.EditNoteById(db, *note); err != nil {
				return false, err
			}

			log.Printf("client(%s) note has been updated\n", connection.RemoteAddr().String())
			if err = SendStatus(connection, SuccessT); err != nil {
				return true, err
			}
		case DeleteNoteT:
			if err = json.Unmarshal(msg.Data, &note); err != nil {
				return true, err
			}

			if err = user.DeleteNoteById(db, *note); err != nil {
				return false, err
			}

			log.Printf("client(%s) note has been deleted\n", connection.RemoteAddr().String())
			if err = SendStatus(connection, SuccessT); err != nil {
				return true, err
			}
		case GetAllMyNotesT:
			note_slice := NoteSliceData{}

			notes, err := user.GetNotesByUser(db)
			if err != nil {
				return false, err
			}

			note_slice.Notes = notes
			note_slice.Count, err = user.GetNotesNumberByUser(db)
			if err != nil {
				return false, err
			}

			note_slice_data, err := json.Marshal(note_slice)
			if err != nil {
				return true, err
			}

			msg.MessageTypeStatus = SuccessT
			msg.Data = note_slice_data

			msg_data, err := json.Marshal(msg)
			if err != nil {
				return true, err
			}

			if _, err = connection.Write(msg_data); err != nil {
				return true, err
			}

			log.Printf("client(%s) notes has been sent\n", connection.RemoteAddr().String())
		case GetLikeTitleNotesT:
			note_slice := NoteSliceData{}

			if err = json.Unmarshal(msg.Data, &note); err != nil {
				return true, err
			}

			note_slice.Notes, err = user.GetNotesByTitle(db, note.Title)
			if err != nil {
				return false, err
			}

			note_slice.Count, err = user.GetNotesNumberByTitle(db, note.Title)
			if err != nil {
				return false, err
			}

			note_slice_data, err := json.Marshal(note_slice)
			if err != nil {
				return true, err
			}

			msg.MessageTypeStatus = SuccessT
			msg.Data = note_slice_data

			msg_data, err := json.Marshal(msg)
			if err != nil {
				return true, err
			}

			_, err = connection.Write(msg_data)
			if err != nil {
				return true, err
			}

			log.Printf("client(%s) notes has been sent\n", connection.RemoteAddr().String())
		}

	}
}

func ClientWorker(connection net.Conn, db *sqlx.DB, ch chan struct{}) {
	defer func() {
		connection.Close()
		log.Printf("client(%s) disconnected\n", connection.RemoteAddr().String())
		log.Printf("max: %d / now: %d\n", cap(ch), len(ch))
	}()

	user, err := Validate(connection, db)
	if err != nil {
		log.Printf("client(%s) %s\n", connection.RemoteAddr().String(), err)
		if serr := SendErrorMsg(connection, err.Error()); serr != nil {
			log.Printf("client(%s) %s\n", connection.RemoteAddr().String(), serr)
		}

	} else {
		msg := MessageData{MessageTypeStatus: SuccessT}
		msg_data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("client(%s) %s\n", connection.RemoteAddr().String(), err)
			goto End
		}
		log.Printf("client(%s) authorized\n", connection.RemoteAddr().String())

		_, err = connection.Write(msg_data)
		if err != nil {
			log.Printf("client(%s) %s\n", connection.RemoteAddr().String(), err)
			goto End
		}

		for {
			status, err := ClientMsgWorker(connection, db, user)
			if status {
				if err != nil {
					log.Printf("client(%s) %s\n", connection.RemoteAddr().String(), err)
					if err = SendErrorMsg(connection, err.Error()); err != nil {
						log.Printf("client(%s) %s\n", connection.RemoteAddr().String(), err)
					}
				}
				break
			}
			if err != nil {
				if err = SendErrorMsg(connection, err.Error()); err != nil {
					log.Println(err)
					break
				}

			}
		}

	}

End:
	<-ch
}

func Validate(connection net.Conn, db *sqlx.DB) (*User, error) {
	msg_data := MessageData{}
	user_data := User{}

	data, err := GetMessageData(connection)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(data, &msg_data); err != nil {
		return nil, err
	}

	switch msg_data.MessageTypeStatus {
	case 2:
		if err = json.Unmarshal(msg_data.Data, &user_data); err != nil {
			return nil, err
		}

		user, err := GetUser(db, user_data.UserName)
		if err != nil {
			return nil, err
		}

		status, err := CheckUserPassword(db, user.UserName, user_data.Password)
		if err != nil {
			return nil, err
		}

		if !status {
			return nil, fmt.Errorf("wrong password")
		}

		return user, nil
	case 3:
		if err = json.Unmarshal(msg_data.Data, &user_data); err != nil {
			return nil, err
		}

		if err = user_data.CreateUser(db); err != nil {
			return nil, err
		}

		return &user_data, nil
	default:
		return nil, fmt.Errorf("message type is not 2 or 3")
	}
}

func SendErrorMsg(connection net.Conn, error_text string) error {
	err_msg := ErrorMessageData{ErrorText: error_text}
	err_msg_data, err := json.Marshal(err_msg)
	if err != nil {
		return err
	}

	msg := MessageData{MessageTypeStatus: ErrorT, Data: err_msg_data}
	msg_data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = connection.Write(msg_data)

	return err
}

func SendStatus(connection net.Conn, status int) error {
	msg := MessageData{MessageTypeStatus: status}
	msg_data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = connection.Write(msg_data)

	return err
}

func StartRoutineServer(host, port string, max_conn int, db *sqlx.DB) error {
	if max_conn > 8 || max_conn < 1 {
		return fmt.Errorf("max 8 / min 1")
	}

	channels := make(chan struct{}, max_conn)
	listener, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("Server is listening [%s:%s]\n", host, port)
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		channels <- struct{}{}
		go ClientWorker(connection, db, channels)

		log.Printf("client(%s) connected\n", connection.RemoteAddr().String())
		log.Printf("max: %d / now: %d\n", cap(channels), len(channels))
	}
}

func GetMessageData(connection net.Conn) ([]byte, error) {
	data := make([]byte, 8192)
	n, err := connection.Read(data)
	if err != nil {
		return nil, err
	}

	return data[:n], nil
}
