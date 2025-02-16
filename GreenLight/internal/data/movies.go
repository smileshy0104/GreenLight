package data

import (
	"DesignMode/GreenLight/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

// 定义Movie结构体（必须使用大写字母开头向外暴露）
type Movie struct {
	ID        int64  `json:"id"` // Unique integer ID for the movie
	CreatedAt int32  `json:"-"`  // Use the - directive to never export in JSON output
	Title     string `json:"title"`
	Year      int32  `json:"year,omitempty"` // Movie release year0
	//Runtime   Runtime   `json:"runtime,omitempty"`
	Runtime Runtime `json:"runtime,omitempty,string"` // 增加string directive后，该字段在respond中会以string类型输出
	Genres  string  `json:"genres,omitempty"`
	Version int32   `json:"version"` // The version number starts at 1 and is incremented each
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
	v.Check(movie.Genres != "", "genres", "must be provided")
	v.Check(len(strings.Split(movie.Genres, ",")) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(strings.Split(movie.Genres, ",")) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(strings.Split(movie.Genres, ",")), "genres", "must not contain duplicate values")
}

// MovieModel 结构体
type MovieModel struct {
	DB       *sql.DB
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

// 创建一个movie
func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (created_at, title, year, runtime, genres) 
		VALUES (?,?,?,?,?) 
		`
	// 通过context上下文的延时函数，超时则自动cancel
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 执行查询
	args := []interface{}{time.Now().Unix(), movie.Title, movie.Year, movie.Runtime, movie.Genres}

	return m.DB.QueryRowContext(ctx, query, args...).Err()
}

// 获取一个movie
func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, year, runtime, genres, version
        FROM movies
 		WHERE id = ?
 		`

	var movie Movie

	// 通过context上下文的延时函数，超时则自动cancel
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 执行查询
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		&movie.Genres,
		&movie.Version)

	// 处理错误
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

// 更新一个movie
func (m MovieModel) Update(movie *Movie) error {
	// 增加了个version条件，可以防止修改冲突的问题！！
	// 因为version变成了个更新的添加，所以第二次更新不会成功！
	query := `
		UPDATE movies
		SET title = ?, year = ?, runtime = ?, genres = ?, version = version + 1
		WHERE id = ? AND version = ?
		`

	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		movie.Genres,
		movie.ID,
		movie.Version, // 增加了个version字段，可以防止修改冲突的问题！！
	}

	// 使用context上下文的延时函数，超时则自动cancel
	// 当对应的上下文context超时了，PostgreSql driver会发送对应的取消信号给数据库，程序会自动中断对应的查询！
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 执行查询
	err := m.DB.QueryRowContext(ctx, query, args...).Err()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict // 返回字段版本冲突
		default:
			return err
		}
	}

	return nil
}

// 删除一个movie
func (m MovieModel) Delete(id int64) error {
	// 检查id是否小于1
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM movies
		WHERE id = ?
		`

	// 使用context上下文的延时函数，超时则自动cancel
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// 判断删除的行数是否为0
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// 获取所有movie
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		`)
	// 添加title条件
	if title != "" {
		query += fmt.Sprintf(" WHERE title LIKE \"%s\"", "%"+title+"%")
	}

	// 添加genres条件
	if len(genres) > 0 && title != "" {
		query += fmt.Sprintf(" AND genres LIKE \"%s\"", "%"+strings.Join(genres, ",")+"%")
	} else if len(genres) > 0 && title == "" {
		query += fmt.Sprintf(" WHERE genres LIKE \"%s\"", "%"+strings.Join(genres, ",")+"%")
	}

	// 添加排序
	query += fmt.Sprintf(" ORDER BY %s %s", filters.sortColumn(), filters.sortDirection())

	// 添加分页
	query += fmt.Sprintf(" LIMIT ? OFFSET ?")

	// 通过context上下文的延时函数，超时则自动cancel
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 执行查询
	args := []interface{}{filters.limit(), filters.offset()}

	// 获取所有movie
	rows, err := m.DB.QueryContext(ctx, query, args...)

	if err != nil {
		return nil, Metadata{}, err
	}

	// 延迟关闭rows
	defer func() {
		if err := rows.Close(); err != nil {
			m.ErrorLog.Println(err)
		}
	}()

	// 初始化一个totalRecords变量，用于存储查询结果中的总记录数
	totalRecords := 0

	// 初始化一个movies变量，用于存储查询结果
	movies := []*Movie{}

	// 遍历rows
	for rows.Next() {
		// 创建一个Movie变量，用于存储查询结果
		var movie Movie

		// 使用rows.Scan()方法将查询结果中的列值赋值给movie变量
		err := rows.Scan(
			//&totalRecords, // Scan the count from the window function into totalRecords.
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			&movie.Genres,
			&movie.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		// 将movie变量添加到movies切片中
		movies = append(movies, &movie)
	}

	// 检查rows.Err()
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// 计算分页信息
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil
}
