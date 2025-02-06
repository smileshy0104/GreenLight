package data

import (
	"DesignMode/GreenLight/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
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

// Insert accepts a pointer to a movie struct, which should contain the data for the
// new record and inserts the record into the movies table.
func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title, year, runtime, genres) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at, version
		`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create an args slice containing the values for the placeholder parameters from the movie
	// struct. Declaring this slice immediately next to our SQL query helps to make it nice and
	// clear *what values are being user where* in the query
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Get fetches a record from the movies table and returns the corresponding Movie struct.
// It cancels the query call if the SQL query does not finish within 3 seconds.
func (m MovieModel) Get(id int64) (*Movie, error) {
	// The PostgreSQL bigserial type that we're using for the movie ID starts auto-incrementing
	// at 1 by default, so we know that no movies will have ID values less tan that.
	// To avoid making an unnecessary database call,
	// we take a shortcut and return an ErrRecordNotFound error straight away.
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, year, runtime, genres, version
        FROM movies
 		WHERE id = $1
 		`

	var movie Movie

	// Use the context.WithTimeout() function to create a context.Context which carries a 3-second
	// timeout deadline. Note, that we're using the empty context.Background() as the
	// 'parent' context.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Defer cancel to make sure that we cancel the context before the Get() method returns
	defer cancel()

	// Use the QueryRowContext() method to execute the query, passing in the context with the
	// deadline ctx as the first argument.
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version)

	// Handle any errors. If there was no matching movie found, Scan() will return a sql.ErrNoRows
	// error. We check for this and return our custom ErrRecordNotFound error instead.
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

// Update updates a specific movie in the movies table.
func (m MovieModel) Update(movie *Movie) error {
	// 增加了个version条件，可以防止修改冲突的问题！！
	// 因为version变成了个更新的添加，所以第二次更新不会成功！
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version
		`

	// Create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version, // 增加了个version字段，可以防止修改冲突的问题！！
	}

	// Create a context with a 3-second timeout.
	// 使用context上下文的延时函数，超时则自动cancel
	// 当对应的上下文context超时了，PostgreSql driver会发送对应的取消信号给数据库，程序会自动中断对应的查询！
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the SQL query. If no matching row could be found, we know the movie version
	// has changed (or the record has been deleted) and we return ErrEditConflict.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
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

// Delete is a placeholder method for deleting a specific record in the movies table.
func (m MovieModel) Delete(id int64) error {
	// Return an ErrRecordNotFound error if the movie ID is less than 1
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM movies
		WHERE id = $1
		`

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the SQL query using the Exec() method,
	// passing in the id variable as the value for the placeholder parameter. The Exec(
	// ) method returns a sql.Result object.
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result
	// object to get the number of rows affected by the query.
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no rows were affected,
	// we know that the movies table didn't contain a record with the provided ID at the moment
	// we tried to delete it. In that case we return an ErrRecordNotFound error.
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// GetAll returns a list of movies in the form of a string of Movie type based on a set of
// provided filters.
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	// Add an ORDER BY clause and interpolate the sort column and direction using fmt.Sprintf.
	// Importantly, notice that we also include a secondary sort on the movie ID to ensure
	// a consistent ordering. Furthermore, we include LIMIT and OFFSET clauses with placeholder
	// parameter values for pagination implementation. The window function is used to calculate
	// the total filtered rows which will be used in our pagination metadata.
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`,
		filters.sortColumn(), filters.sortDirection())

	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Organize our four placeholder parameter values in a slice.
	args := []interface{}{title, pq.Array(genres), filters.limit(), filters.offset()}

	// Use QueryContext to execute the query. This returns a sql.Rows result set containing
	// the result.
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	// Importantly, defer a call to rows.Close() to ensure that the result set is closed
	// before GetAll returns.
	defer func() {
		if err := rows.Close(); err != nil {
			m.ErrorLog.Println(err)
		}
	}()

	// Declare a totalRecords variable
	totalRecords := 0

	// Initialize an empty slice to hold the movie data.
	movies := []*Movie{}

	// Use rows.Next to iterate through the rows in the result set.
	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var movie Movie

		// Scan the values from the row into the Movie struct. Again, note that we're using
		// the pq.Array adapter on the genres field.
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		// Add the Movie struct to the slice
		movies = append(movies, &movie)
	}

	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Generate a Metadata struct, passing in the total record count and pagination parameters
	// from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	// If everything went OK, then return the slice of the movies and metadata.
	return movies, metadata, nil
}
