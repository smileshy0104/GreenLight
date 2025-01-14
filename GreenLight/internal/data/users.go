package data

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var AnonymousUser = &User{}

// User type whose fields describe a user. Note, that we use the json:"-" struct tag to prevent
// the Password and Version fields from appearing in any output when we encode it to JSON.
// Also, notice that the Password field uses the custom password type defined below.
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// UserModel struct wraps a sql.DB connection pool and allows us to work with the User struct type
// and the users table in our database.
type UserModel struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

// password tyep is a struct containing the plaintext and hashed version of a password for a User.
// The plaintext field is a *pointer* to a string, so that we're able to distinguish between a
// plaintext password not being present in the struct at all, versus a plaintext password which
// is the empty string "".
type password struct {
	plaintext *string
	hash      []byte
}
