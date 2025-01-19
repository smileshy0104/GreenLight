package data

import (
	"DesignMode/GreenLight/internal/validator"
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

// ValidateMovie函数 （封装校验函数）
func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

type MovieModel struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}
