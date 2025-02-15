package data

import (
	"database/sql"
	"errors"
	"log"
	"os"
)

var (
	// 创建一个 ErrRecordNotFound 变量，并赋值给它一个错误对象（记录未找到）。
	ErrRecordNotFound = errors.New("record not found")
	// 创建一个 ErrEditConflict 变量，并赋值给它一个错误对象（修改冲突）。
	ErrEditConflict = errors.New("edit conflict")
)

// Models 结构体，用于封装数据库模型。
type Models struct {
	Movies      MovieModel
	Users       UserModel
	Tokens      TokenModel
	Permissions PermissionModel
}

// 创建一个Models结构体，并初始化其中的各个字段。
func NewModels(db *sql.DB) Models {
	// 创建一个日志记录器，并记录一条消息，表示数据库连接池已建立。
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return Models{
		Movies: MovieModel{
			DB:       db,
			InfoLog:  infoLog,
			ErrorLog: errorLog,
		},
		Users: UserModel{
			DB:       db,
			InfoLog:  infoLog,
			ErrorLog: errorLog,
		},
		Tokens: TokenModel{
			DB:       db,
			InfoLog:  infoLog,
			ErrorLog: errorLog,
		},
		Permissions: PermissionModel{
			DB:       db,
			InfoLog:  infoLog,
			ErrorLog: errorLog,
		},
	}
}
