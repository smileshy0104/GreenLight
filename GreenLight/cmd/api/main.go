package main

import (
	"DesignMode/GreenLight/internal/data"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// 定义应用程序的版本号。
const version = "1.0.0"

// 配置结构体，用于存储服务器端口和环境变量。
type config struct {
	port int    // 端口
	env  string // 环境
	// For now this only holds the DSN, which we read in from a command-line flag.
	// 数据库相关配置信息，用于数据库连接池配置
	db struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// application 结构体包含配置和日志记录器。
type application struct {
	config config      // 相关配置结构体
	logger *log.Logger // 日志记录器
	//logger *jsonlog.Logger  // json日志记录器
	models data.Models    // 数据库模型
	wg     sync.WaitGroup // 等待组
}

// TestHttpServer 是一个测试函数，用于启动 HTTP 服务器。
// 主要功能包括解析命令行参数、初始化日志记录器和应用程序结构体、设置路由并启动服务器。
func main() {
	// 定义配置结构体，用于存储服务器端口、环境变量等相关配置。
	var cfg config

	// 解析命令行参数以初始化配置。
	// TODO 端口和环境变量都可以在终端自定义
	// go run ./cmd/api -port=3030 -env=production
	flag.IntVar(&cfg.port, "port", 4000, "API 服务器端口")
	flag.StringVar(&cfg.env, "env", "development", "环境 (development|staging|production)")
	// 解析命令行参数以初始化配置。
	flag.Parse()

	// 初始化日志记录器。
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	//logger := jsonlog.NewLogger(os.Stdout, jsonlog.LevelInfo)

	// 创建应用程序结构体，并初始化相关配置和日志记录器。
	app := &application{
		config: cfg,
		logger: logger,
	}

	// 初始化请求路由器。
	// TODO 使用http.NewServeMux()功能非常受限，不方便使用。
	//mux := http.NewServeMux()
	//mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// 配置 HTTP 服务器。
	// TODO 使用 httprouter 作为请求路由器，并设置相关配置。
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.port), // 设置监听地址和端口
		//Handler:      mux,                 // 设置请求路由器
		Handler:      app.routes(),     // 设置请求路由器
		IdleTimeout:  time.Minute,      // 空闲超时
		ReadTimeout:  10 * time.Second, // 读取超时
		WriteTimeout: 30 * time.Second, // 写入超时
	}

	// 启动 HTTP 服务器。
	logger.Printf("正在启动 %s 环境下的服务器，监听地址为 %s", cfg.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}
