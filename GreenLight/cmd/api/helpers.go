package main

import (
	"DesignMode/GreenLight/internal/validator"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// TODO 用于提供公共函数的Helpers包

// envelope 是一个键值对映射，用于封装响应数据。（统一响应数据结构）
type envelope map[string]interface{}

// readIDParam 读取路由中Id参数并返回
func (app *application) readIDParam(r *http.Request) (int64, error) {
	// 获取路由参数
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	// 如果出现错误或者id小于1，则返回错误
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// writeJSONOld 使用json.Marshal 函数将数据编码为JSON格式，并返回一个包含JSON数据的字节切片。
func (app *application) writeJSONOld(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {
	// 使用json.Marshal函数将数据marshal
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// 添加换行符，使其更易阅读
	js = append(js, '\n')
	// 设置任意头信息
	for key, value := range headers {
		w.Header()[key] = value
	}
	// 设置响应头，使其兼容Json模式
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

// writeJSON 使用json.MarshalIndent 函数将数据编码为JSON格式，将每个元素放在单独的行上，并使用可选的前缀和缩进字符。
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// 使用json.MarshalIndent函数将数据marshal，并使用可选的前缀和缩进字符。
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}
	// 添加换行符，使其更易阅读
	js = append(js, '\n')
	for key, value := range headers {
		w.Header()[key] = value
	}
	// 设置响应头，使其兼容Json模式
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
	//err := json.NewDecoder(r.Body).Decode(dst)

	// 检查请求体大小是否超过1MB，如果是则返回错误
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// 创建一个JSON解码器
	dec := json.NewDecoder(r.Body)
	// 禁用未知字段，如果存在未知字段，则返回错误
	dec.DisallowUnknownFields()
	// 使用Decode方法将请求体解码到目标指针中
	err := dec.Decode(dst)

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

		// 检查是否为未知字段错误，并返回带有字段名称的详细信息
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// 检查是否为请求体过大错误，并返回错误信息
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		// 检查是否为无效解码错误，并触发 panic
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		// 其他未知错误直接返回
		default:
			return err
		}
	}
	// 检查是否还有未读取的JSON数据，如果有则返回错误
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// readString 读取查询字符串中的字符串值，并返回该值。
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	// 获取url中的key值
	s := qs.Get(key)
	// 如果没有获取到值，则返回默认值
	if s == "" {
		return defaultValue
	}
	return s
}

// readCSV 读取查询字符串中的逗号分隔值（CSV）并返回一个字符串切片。
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	// 获取url中的key值
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}
	// 使用Split函数将逗号分隔的字符串分割为字符串切片
	return strings.Split(csv, ",")
}

// readInt 读取查询字符串中的整数值，并返回该值。
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	// 获取url中的key值
	s := qs.Get(key)
	// 如果没有获取到值，则返回默认值
	if s == "" {
		return defaultValue
	}
	// 使用strconv.Atoi函数将字符串转换为整数
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	return i
}

// background 函数用于在后台运行一个函数，并使用defer语句来处理任何可能的panic。
func (app *application) background(fn func()) {
	// 使用WaitGroup来等待后台goroutine完成
	app.wg.Add(1)
	go func() {
		// defer语句会在函数返回之前执行
		defer app.wg.Done()
		// recover处理任何可能的panic
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		// 调用函数
		fn()
	}()
}
