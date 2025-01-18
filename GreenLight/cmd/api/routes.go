package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// 路由配置文件（将所有的路由相关内容写到这个文件）
func (app *application) routes() *httprouter.Router {
	// 创建一个路由实例（使用httprouter中的对应函数）
	router := httprouter.New()
	// Register the relevant methods, URL patterns and handler functions for our endpoints using the HandlerFunc() method.
	// Note that http.MethodGet and http.MethodPost are constants which equate to the strings "GET" and "POST"
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler) // 健康检查端点的处理函数。
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	// 返回路由实例
	return router
}
