package main

import (
	"encoding/json"
	"fmt"
	"net"
)

func (user User) ConnectToServer(host, port string, Type int) (net.Conn, error) {
	connection, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		return nil, err
	}

	user_data, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	msg := MessageData{MessageTypeStatus: Type, Data: user_data}
	msg_data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	_, err = connection.Write(msg_data)
	if err != nil {
		return nil, err
	}

	b, err := GetMessageData(connection)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(b, &msg); err != nil {
		return nil, err
	}

	if msg.MessageTypeStatus == SuccessT {
		return connection, nil
	} else if msg.MessageTypeStatus == ErrorT {
		err_data := ErrorMessageData{}

		if err := json.Unmarshal(msg.Data, &err_data); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf(err_data.ErrorText)
	}

	return nil, fmt.Errorf("unknown error")
}

func CreateNote(connection net.Conn, note Note) error {
	note_data, err := json.Marshal(note)
	if err != nil {
		return err
	}

	msg := MessageData{MessageTypeStatus: NewNoteT, Data: note_data}
	msg_data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err = connection.Write(msg_data); err != nil {
		return err
	}

	bytes, err := GetMessageData(connection)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(bytes, &msg); err != nil {
		return err
	}

	switch msg.MessageTypeStatus {
	case SuccessT:
		return nil
	case ErrorT:
		err_msg := ErrorMessageData{}
		if err = json.Unmarshal(msg.Data, &err_msg); err != nil {
			return err
		}

		return fmt.Errorf(err_msg.ErrorText)
	default:
		return fmt.Errorf("unknown server message code")
	}
}

func GetNote(connection net.Conn, note Note) (*Note, error) {
	note_data, err := json.Marshal(note)
	if err != nil {
		return nil, err
	}

	msg := MessageData{MessageTypeStatus: GetNoteT, Data: note_data}
	msg_data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	if _, err = connection.Write(msg_data); err != nil {
		return nil, err
	}

	bytes, err := GetMessageData(connection)

	if err = json.Unmarshal(bytes, &msg); err != nil {
		return nil, err
	}

	switch msg.MessageTypeStatus {
	case SuccessT:
		_note := Note{}
		if err = json.Unmarshal(msg.Data, &_note); err != nil {
			return nil, err
		}

		return &_note, nil
	case ErrorT:
		err_msg := ErrorMessageData{}
		if err = json.Unmarshal(msg.Data, &err_msg); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf(err_msg.ErrorText)
	default:
		return nil, fmt.Errorf("unknown server message code")
	}
}

func UpdateNote(connection net.Conn, note Note) error {
	note_data, err := json.Marshal(note)
	if err != nil {
		return err
	}

	msg := MessageData{MessageTypeStatus: UpdateNoteT, Data: note_data}
	msg_data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err = connection.Write(msg_data); err != nil {
		return err
	}

	bytes, err := GetMessageData(connection)

	if err = json.Unmarshal(bytes, &msg); err != nil {
		return err
	}

	switch msg.MessageTypeStatus {
	case SuccessT:
		return nil
	case ErrorT:
		err_msg := ErrorMessageData{}
		if err = json.Unmarshal(msg.Data, &err_msg); err != nil {
			return err
		}

		return fmt.Errorf(err_msg.ErrorText)
	default:
		return fmt.Errorf("unknown server message code")
	}
}

func DeleteNote(connection net.Conn, note Note) error {
	note_data, err := json.Marshal(note)
	if err != nil {
		return err
	}

	msg := MessageData{MessageTypeStatus: DeleteNoteT, Data: note_data}
	msg_data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err = connection.Write(msg_data); err != nil {
		return err
	}

	bytes, err := GetMessageData(connection)

	if err = json.Unmarshal(bytes, &msg); err != nil {
		return err
	}

	switch msg.MessageTypeStatus {
	case SuccessT:
		return nil
	case ErrorT:
		err_msg := ErrorMessageData{}
		if err = json.Unmarshal(msg.Data, &err_msg); err != nil {
			return err
		}

		return fmt.Errorf(err_msg.ErrorText)
	default:
		return fmt.Errorf("unknown server message code")
	}
}

func GetAllNotes(connection net.Conn) ([]Note, error) {
	msg := MessageData{MessageTypeStatus: GetAllMyNotesT}
	msg_data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	if _, err = connection.Write(msg_data); err != nil {
		return nil, err
	}

	bytes, err := GetMessageData(connection)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(bytes, &msg); err != nil {
		return nil, err
	}

	switch msg.MessageTypeStatus {
	case SuccessT:
		note_slice := NoteSliceData{}
		if err = json.Unmarshal(msg.Data, &note_slice); err != nil {
			return nil, err
		}

		notes := make([]Note, 0, note_slice.Count)
		notes = append(notes, note_slice.Notes...)

		return notes, nil
	case ErrorT:
		err_msg := ErrorMessageData{}
		if err = json.Unmarshal(msg.Data, &err_msg); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf(err_msg.ErrorText)
	default:
		return nil, fmt.Errorf("unknown server message code")
	}
}

func GetAllNotesByTitle(connection net.Conn, note Note) ([]Note, error) {
	note_data, err := json.Marshal(note)
	if err != nil {
		return nil, err
	}

	msg := MessageData{MessageTypeStatus: GetLikeTitleNotesT, Data: note_data}
	msg_data, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	if _, err = connection.Write(msg_data); err != nil {
		return nil, err
	}

	bytes, err := GetMessageData(connection)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(bytes, &msg); err != nil {
		return nil, err
	}

	switch msg.MessageTypeStatus {
	case SuccessT:
		note_slice := NoteSliceData{}
		if err = json.Unmarshal(msg.Data, &note_slice); err != nil {
			return nil, err
		}

		notes := make([]Note, 0, note_slice.Count)
		notes = append(notes, note_slice.Notes...)

		return notes, nil
	case ErrorT:
		err_msg := ErrorMessageData{}
		if err = json.Unmarshal(msg.Data, &err_msg); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf(err_msg.ErrorText)
	default:
		return nil, fmt.Errorf("unknown server message code")
	}
}
