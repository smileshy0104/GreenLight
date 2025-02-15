package main

import (
	"DesignMode/GreenLight/internal/data"
	"DesignMode/GreenLight/internal/validator"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TODO 该文件存储对应的movies相关的业务函数

// createMovieHandlerOld 创建Movie
func (app *application) createMovieHandlerOld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "create a new movie")
}

// createMovieHandler 创建Movie（使用json.Decoder进行解析）
func (app *application) createMovieHandlerOld1(w http.ResponseWriter, r *http.Request) {
	// 声明一个匿名结构体，用于存储从 HTTP 请求体中预期获取的信息。
	// 该结构体的字段名和类型是之前创建的 Movie 结构体的一个子集。
	// 这个结构体将作为解码Decode的目标目标。
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	// 初始化一个新的 json.Decoder 实例，该实例从请求体中读取数据，
	// 然后使用 Decode() 方法将请求体内容解码到 input 结构体中。
	// 注意，在调用 Decode() 时，我们传递的是 input 结构体的指针作为解码的目标。
	// 如果解码过程中出现错误，我们将使用通用的 errorResponse() 辅助函数
	// 向客户端发送一个 400 Bad Request 响应，并包含错误信息。
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// 将 input 结构体的内容以 HTTP 响应的形式输出。
	fmt.Fprintf(w, "%+v\n", input)
}

// createMovieHandler 创建Movie（使用json.Unmarshal进行解析）
func (app *application) createMovieHandlerOld2(w http.ResponseWriter, r *http.Request) {
	// 声明一个匿名结构体，用于存储从 HTTP 请求体中预期获取的信息。
	// 该结构体的字段名和类型是之前创建的 Movie 结构体的一个子集。
	// 这个结构体将作为解码Decode的目标目标。
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}

	// Use io.ReadAll() to read the entire request body into a []byte slice.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// 初始化一个新的 json.Unmarshal 实例，该实例从请求体中读取数据，
	// 如果解码过程中出现错误，我们将使用通用的 errorResponse() 辅助函数
	// 向客户端发送一个 400 Bad Request 响应，并包含错误信息。
	err = json.Unmarshal(body, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// 将 input 结构体的内容以 HTTP 响应的形式输出。
	fmt.Fprintf(w, "%+v\n", input)
}

// createMovieHandler 创建Movie（使用封装函数app.readJSON(w, r, &input)）
func (app *application) createMovieHandlerOld3(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}
	// Use the new readJSON() helper to decode the request body into the input struct.
	// If this returns an error we send the client the error message along with a 400
	// Bad Request status code, just like before.
	err := app.readJSON(w, r, &input)
	if err != nil {
		//app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		app.badRequestResponse(w, r, err)
		return
	}
	fmt.Fprintf(w, "%+v\n", input)
}

// createMovieHandler 创建Movie（使用封装函数app.readJSON(w, r, &input)和解析自定义字段Runtime）
// 将输入的String类型反解析成Int类型
func (app *application) createMovieHandlerOld4(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"` // Make this field a data.Runtime type.
		Genres  []string     `json:"genres"`
	}
	// Use the new readJSON() helper to decode the request body into the input struct.
	// If this returns an error we send the client the error message along with a 400
	// Bad Request status code, just like before.
	err := app.readJSON(w, r, &input)
	if err != nil {
		//app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		app.badRequestResponse(w, r, err)
		return
	}
	fmt.Fprintf(w, "%+v\n", input)
}

// createMovieHandler 创建Movie（使用校验器validator）
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// 声明一个匿名结构体，用于存储从 HTTP 请求体中预期获取的信息。
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"` // Make this field a data.Runtime type.
		Genres  []string     `json:"genres"`
	}
	// 读取请求体
	err := app.readJSON(w, r, &input)
	if err != nil {
		//app.errorResponse(w, r, http.StatusBadRequest, err.Error())
		app.badRequestResponse(w, r, err)
		return
	}

	// 复制输入结构体movie，使用这个新的Movie结构体去进行校验
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  strings.Join(input.Genres, ","),
	}
	// 创建校验器实例
	v := validator.New()

	// 校验Movie
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// 将movie插入数据库
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// 创建Location响应头
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))
	// 将movie返回给客户端
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	//fmt.Fprintf(w, "%+v\n", input)
}

// showMovieHandlerOld1 查看Movie（使用httprouter.ParamsFromContext(r.Context())读取路由）
func (app *application) showMovieHandlerOld(w http.ResponseWriter, r *http.Request) {
	// 获取路由参数
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	// Otherwise, interpolate the movie ID in a placeholder response.
	fmt.Fprintf(w, "show the details of movie %d\n", id)
}

// showMovieHandlerOld2 查看Movie（使用封装函数app.readIDParam(r)）
func (app *application) showMovieHandlerOld1(w http.ResponseWriter, r *http.Request) {
	// 获取路由参数（使用封装函数）
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// Otherwise, interpolate the movie ID in a placeholder response.
	fmt.Fprintf(w, "show the details of movie %d\n", id)
}

// showMovieHandlerOld3 查看Movie（使用封装函数app.writeJSONOld(w, http.StatusOK, movie, nil)）
// 使用结构体movie去传递
func (app *application) showMovieHandlerOld2(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// 使用结构体movie去传递
	movie := data.Movie{
		ID:        id,
		CreatedAt: int32(time.Now().Unix()),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    strings.Join([]string{"drama", "romance", "war"}, ","),
		Version:   1,
	}
	// Encode the struct to JSON and send it as the HTTP response.
	err = app.writeJSONOld(w, http.StatusOK, movie, nil)
	if err != nil {
		app.logger.Println(err)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

// showMovieHandler 查看Movie（使用统一的envelope作为统一响应数据结构）
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		// 使用未封装的http.NotFound()函数
		//http.NotFound(w, r)
		// Use the new notFoundResponse() helper.
		app.notFoundResponse(w, r)
		return
	}
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Create an envelope{"movie": movie} instance and pass it to writeJSON(), instead
	// of passing the plain movie struct.
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		// 使用未封装的logger和http.Error()函数
		//app.logger.Println(err)
		//http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
		// Use the new serverErrorResponse() helper.
		app.serverErrorResponse(w, r, err)
	}
}

// updateMovieHandler 更新Movie
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// 获取路由参数
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// 获取对应id的movie
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 如果请求包含X-Exped-Version标头，请验证电影数据库中的版本与标头中指定的预期版本匹配
	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(movie.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	// 声明结构体 input，并且使用指针判断其输入是否为空！
	var input struct {
		//Title   string       `json:"title"`
		//Year    int32        `json:"year"`
		//Runtime data.Runtime `json:"runtime"`
		ID      int64         `json:"id"`
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}
	// Read the JSON request body data into the input struct.
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// TODO 整个更新使用PUT
	//movie.Title = input.Title
	//movie.Year = input.Year
	//movie.Runtime = input.Runtime
	//movie.Genres = input.Genres

	// TODO 部分更新使用PATCH
	if input.Title != nil {
		movie.Title = *input.Title
	}
	// We also do the same for the other fields in the input struct.
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = strings.Join(input.Genres, ",") // Note that we don't need to dereference a slice.
	}

	// 校验器
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// 更新Movie
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r) // 使用未封装的editConflictResponse()函数
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// 使用writeJSON()函数
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteMovieHandler 删除Movie
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL.
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Delete the movie from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return a 200 OK status code along with a success message.
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// listMoviesHandler 电影列表
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	// 声明结构体 input
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}
	v := validator.New()
	// 获取查询字符串参数
	qs := r.URL.Query()
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// 使用一个硬编码的slice来验证用户输入的排序参数
	input.Filters.SortSafeList = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}
	// 验证过滤器
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// 调用MovieModel的GetAll()方法获取电影列表
	movies, _, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// 将电影列表写入JSON响应
	err = app.writeJSON(w, http.StatusOK, envelope{"movies": movies}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
