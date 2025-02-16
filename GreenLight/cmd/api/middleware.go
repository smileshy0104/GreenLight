package main

import (
	"fmt"
	"net/http"
)

// 创建中间件recoverPanic，用于处理程序恐慌
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 推迟函数调用，以便在函数返回时执行。
		defer func() {
			// 使用recover()捕获程序恐慌
			if err := recover(); err != nil {
				// 如果有对应的错误被捕获，则设置HTTP响应头Connection: close
				w.Header().Set("Connection", "close")
				// 调用serverErrorResponse函数，将错误信息作为参数传递给它
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		// 调用next.ServeHTTP()
		next.ServeHTTP(w, r)
	})
}
