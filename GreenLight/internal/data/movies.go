package data

import (
	"database/sql"
	"log"
	"time"
)

// 定义Movie结构体（必须使用大写字母开头向外暴露）
type Movie struct {
	ID        int64     `json:"id"` // Unique integer ID for the movie
	CreatedAt time.Time `json:"-"`  // Use the - directive to never export in JSON output
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"` // Movie release year0
	//Runtime   Runtime   `json:"runtime,omitempty"`
	Runtime Runtime  `json:"runtime,omitempty,string"` // 增加string directive后，该字段在respond中会以string类型输出
	Genres  []string `json:"genres,omitempty"`
	Version int32    `json:"version"` // The version number starts at 1 and is incremented each
	// time the movie information is updated.
}

type MovieModel struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}
