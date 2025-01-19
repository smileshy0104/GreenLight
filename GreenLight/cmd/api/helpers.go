package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"strconv"
)

// TODO 用于提供公共函数的Helpers包

// envelope 是一个键值对映射，用于封装响应数据。（统一响应数据结构）
type envelope map[string]interface{}

// readIDParam 读取路由中Id参数并返回
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// writeJSONOld 使用json.Marshal 函数将数据编码为JSON格式，并返回一个包含JSON数据的字节切片。
func (app *application) writeJSONOld(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	// Encode the data to JSON, returning the error if there was one.
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// Append a newline to make it easier to view in terminal applications.
	js = append(js, '\n')
	// 设置任意头信息
	for key, value := range headers {
		w.Header()[key] = value
	}
	// Add the "Content-Type: application/json" header, then write the status code and
	// JSON response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

// Change the data parameter to have the type envelope instead of interface{}.
// writeJSON 使用json.MarshalIndent 函数将数据编码为JSON格式，将每个元素放在单独的行上，并使用可选的前缀和缩进字符。
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	js = append(js, '\n')
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

// readJSON 尝试读取并解码 JSON 请求体到提供的目标指针中。
// 它处理各种错误情况，并返回适当的错误消息。
// 参数：
//   - w: http.ResponseWriter，用于写入 HTTP 响应。
//   - r: *http.Request，包含 HTTP 请求信息。
//   - dst: interface{}，指向要解码的目标的指针。
//
// 返回值：
//   - error：如果发生错误，则返回错误；否则返回 nil。
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Use io.ReadAll() to read the entire request body into a []byte slice.
	//body, err := io.ReadAll(r.Body)
	//if err != nil {
	//	app.serverErrorResponse(w, r, err)
	//	return
	//}
	// TODO 使用Unmarshal解码请求体到目标指针中
	//json.Unmarshal(body, dst)

	// TODO 使用Decode解码请求体到目标指针中
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		// 如果解码过程中出现错误，开始分类处理
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// 检查是否为语法错误，并返回带有错误位置的详细信息
		case errors.As(err, &syntaxError):
			return fmt.Errorf("请求体包含格式错误的 JSON（在字符 %d 处）", syntaxError.Offset)

		// 检查是否为意外结束符错误，通常由 JSON 语法错误引起
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("请求体包含格式错误的 JSON")

		// 检查是否为类型不匹配错误，并返回带有字段名称的详细信息
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("请求体包含字段 %q 类型错误", unmarshalTypeError.Field)
			}
			return fmt.Errorf("请求体包含类型错误（在字符 %d 处）", unmarshalTypeError.Offset)

		// 检查是否为空请求体错误
		case errors.Is(err, io.EOF):
			return errors.New("请求体不能为空")

		// 检查是否为无效解码错误，并触发 panic
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// 其他未知错误直接返回
		default:
			return err
		}
	}

	return nil
}
