package main

import (
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/jmoiron/sqlx"
	_ "github.com/jmoiron/sqlx"

	"golang.org/x/crypto/bcrypt"
	_ "golang.org/x/crypto/bcrypt"
)

const schema = `CREATE TABLE "users" (
	"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
	"user_name"	TEXT NOT NULL,
	"password"	TEXT NOT NULL
);

CREATE TABLE "notes" (
	"id"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
	"user_id"	INTEGER NOT NULL,
	"title"	TEXT NOT NULL,
	"data_text"	TEXT NOT NULL
)`

type User struct {
	Id       int
	UserName string `db:"user_name" json:"user_name"`
	Password string `json:"password"`
}

type Note struct {
	Id     int
	UserId int    `db:"user_id" json:"user_id"`
	Title  string `db:"title" json:"title"`
	Data   string `db:"data_text" json:"data_text"`
}

func CreateConn(driverName, dataSourceName string) (*sqlx.DB, error) {
	file, err := os.Open(dataSourceName)
	file.Close()

	if err != nil {
		db, err := sqlx.Connect(driverName, dataSourceName)
		if err != nil {
			return nil, err
		}
		defer db.Close()

		db.MustExec(schema)
		db.MustBegin()
	}

	return sqlx.Connect(driverName, dataSourceName)
}

func (data *User) CreateUser(db *sqlx.DB) error {
	if data.Password == "" || data.UserName == "" {
		return fmt.Errorf("password is null")
	}

	_, err := GetUser(db, data.UserName)
	if err == nil {
		return fmt.Errorf("user with \"%s\" nickname has been registered", data.UserName)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	data.Password = string(hashedPassword)

	tx := db.MustBegin()
	_, err = tx.NamedExec("insert into users (user_name, password) values (:user_name, :password)", data)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetUser(db *sqlx.DB, user_name string) (*User, error) {
	user := new(User)

	err := db.Get(user, "select * from users where user_name=$1", user_name)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func CheckUserPassword(db *sqlx.DB, user_name, password string) (bool, error) {
	user, err := GetUser(db, user_name)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil, nil
}

func (user *User) GetNotesNumberByUserId(db *sqlx.DB) (count int, err error) {
	row := db.QueryRow("select count(*) from notes where user_id=$1", user.Id)
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (user *User) GetNotesByTitle(db *sqlx.DB, title string) ([]Note, error) {
	var count int

	row := db.QueryRow("select count(*) from notes where user_id=? and title like ?", user.Id, "%"+title+"%")
	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	notes := make([]Note, 0, count)
	err = db.Select(&notes, "select * from notes where user_id=? and title like ?", user.Id, "%"+title+"%")
	if err != nil {
		return nil, err
	}

	return notes, nil
}

func (data *Note) CreateNote(db *sqlx.DB, user *User) error {
	notes, err := user.GetNotesByTitle(db, data.Title)
	if err != nil {
		return err
	}

	for _, note := range notes {
		if note.Title == data.Title {
			return fmt.Errorf("note with \"%s\" name already exists", data.Title)
		}
	}

	data.UserId = user.Id
	tx := db.MustBegin()
	_, err = tx.NamedExec("insert into notes (user_id, title, data_text) values (:user_id, :title, :data_text)", data)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (user *User) GetNotesByUser(db *sqlx.DB) ([]Note, error) {
	var count int

	row := db.QueryRow("select count(*) from notes where user_id=?", user.Id)
	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	notes := make([]Note, 0, count)
	err = db.Select(&notes, "select * from notes where user_id=?", user.Id)
	if err != nil {
		return nil, err
	}

	return notes, nil
}

func (user *User) GetNoteById(db *sqlx.DB, note_id int) (*Note, error) {
	var note Note

	err := db.Get(&note, "select * from notes where id=$1 and user_id=$2", note_id, user.Id)
	if err != nil {
		return nil, err
	}

	return &note, nil
}

func (user *User) EditNoteById(db *sqlx.DB, new_note Note) error {
	note, err := user.GetNoteById(db, new_note.Id)
	if err != nil {
		return err
	}

	note.Title = new_note.Title
	note.Data = new_note.Data

	tx := db.MustBegin()
	_, err = tx.NamedExec("update notes set title=:title, data_text=:data_text where id=:id", note)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (user *User) DeleteNoteById(db *sqlx.DB, new_note Note) error {
	note, err := user.GetNoteById(db, new_note.Id)
	if err != nil {
		return err
	}

	tx := db.MustBegin()
	_, err = tx.NamedExec("delete from notes where id=:id", note)
	if err != nil {
		return err
	}

	return tx.Commit()
}
