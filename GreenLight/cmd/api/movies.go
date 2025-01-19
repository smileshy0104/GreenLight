package main

import (
	"DesignMode/GreenLight/internal/data"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"strconv"
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
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
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
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
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
	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
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
