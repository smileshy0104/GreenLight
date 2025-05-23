package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

// TestHttpServer 是一个测试函数，用于启动 HTTP 服务器。
// 主要功能包括解析命令行参数、初始化日志记录器和应用程序结构体、设置路由并启动服务器。
func TestHttpServer(t *testing.T) {
	var cfg config

	// 解析命令行参数以初始化配置。
	// TODO 端口和环境变量都可以自定义
	// go run ./cmd/api -port=3030 -env=production
	flag.IntVar(&cfg.port, "port", 4000, "API 服务器端口")
	flag.StringVar(&cfg.env, "env", "development", "环境 (development|staging|production)")
	flag.Parse()

	// 初始化日志记录器。
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	app := &application{
		config: cfg,
		logger: logger,
	}

	// 初始化请求路由器。
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// 配置 HTTP 服务器。
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 启动 HTTP 服务器。
	logger.Printf("正在启动 %s 环境下的服务器，监听地址为 %s", cfg.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}
