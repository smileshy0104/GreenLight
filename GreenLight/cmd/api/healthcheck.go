package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// healthcheckHandler 是健康检查端点的处理函数。
// 它将服务的可用性、环境和版本信息写入响应中。
// healthcheckHandler是作为application struct的方法，所以它需要一个指针。
func (app *application) healthcheckHandlerOld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "environment: %s\n", app.config.env)
	fmt.Fprintf(w, "version: %s\n", version)
}

// healthcheckHandlerOld1 使用设置响应头，使其兼容Json模式
func (app *application) healthcheckHandlerOld1(w http.ResponseWriter, r *http.Request) {
	// 使用字符串格式化来构建JSON。
	js := `{"status": "available", "environment": %q, "version": %q}`
	js = fmt.Sprintf(js, app.config.env, version)
	// Set the "Content-Type: application/json" header on the response. If you forget to
	// this, Go will default to sending a "Content-Type: text/plain; charset=utf-8"
	// header instead.
	w.Header().Set("Content-Type", "application/json") // 设置响应头，使其兼容Json模式
	// Write the JSON as the HTTP response body.
	w.Write([]byte(js))
}

// healthcheckHandlerOld2 使用map来构建JSON，然后使用json.Marshal()函数来编码它。
func (app *application) healthcheckHandlerOld2(w http.ResponseWriter, r *http.Request) {
	// Create a map which holds the information that we want to send in the response.
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}
	// Pass the map to the json.Marshal() function. This returns a []byte slice
	// containing the encoded JSON. If there was an error, we log it and send the client
	// a generic error message.
	js, err := json.Marshal(data)
	if err != nil {
		app.logger.PrintError(err, nil) // log日志包
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
		return
	}
	// append “\n” 为了更好的阅读
	js = append(js, '\n')
	// At this point we know that encoding the data worked without any problems, so we
	// can safely set any necessary HTTP headers for a successful response.
	w.Header().Set("Content-Type", "application/json") // 设置响应头，使其兼容Json模式
	w.Write(js)
}

// healthcheckHandlerOld3 使用封装的writeJSON()函数来构建JSON，然后使用json.Marshal()函数来编码它。
func (app *application) healthcheckHandlerOld3(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}
	err := app.writeJSONOld(w, http.StatusOK, data, nil) // 调用封装的writeJSON()函数
	if err != nil {
		app.logger.PrintError(err, nil)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

// healthcheckHandler 使用封装的writeJSON()函数来构建JSON，然后使用json.Marshal()函数来编码它。（使用统一的envelope作为统一响应数据结构）
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an envelope map containing the data for the response. Notice that the way
	// we've constructed this means the environment and version data will now be nested
	// under a system_info key in the JSON response.
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}
	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		// 使用未封装的logger和http.Error()函数
		//app.logger.Println(err)
		//http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
		// Use the new serverErrorResponse() helper.
		app.serverErrorResponse(w, r, err)
	}
}
