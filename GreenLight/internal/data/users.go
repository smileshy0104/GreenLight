package data

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

var (
	// ErrDuplicateEmail 邮箱重复
	ErrDuplicateEmail = errors.New("duplicate email")
)

// AnonymousUser 匿名用户
var AnonymousUser = &User{}

// User 结构体
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// IsAnonymous 判断用户是否为匿名用户
func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// UserModel 结构体
type UserModel struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

// password 结构体
type password struct {
	plaintext *string
	hash      []byte
}
