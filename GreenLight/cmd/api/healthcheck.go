package main

import (
	"fmt"
	"net/http"
)

// healthcheckHandler 是健康检查端点的处理函数。
// 它将服务的可用性、环境和版本信息写入响应中。
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "状态: 可用")
	fmt.Fprintf(w, "环境: %s\n", app.config.env)
	fmt.Fprintf(w, "版本: %s\n", version)
}
