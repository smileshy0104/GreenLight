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
	port int
	env  string
	// db struct field holds the configuration settings for our database connection pool.
	// For now this only holds the DSN, which we read in from a command-line flag.
	db struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	// Add a new limiter struct containing fields for the request-per-second and burst
	// values, and a boolean field which we can use to enable/disable rate limiting.
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

// application 结构体包含配置和日志记录器。
type application struct {
	config config
	logger *log.Logger
	//logger *jsonlog.Logger
	models data.Models
	wg     sync.WaitGroup
}

// TestHttpServer 是一个测试函数，用于启动 HTTP 服务器。
// 主要功能包括解析命令行参数、初始化日志记录器和应用程序结构体、设置路由并启动服务器。
func main() {
	var cfg config

	// 解析命令行参数以初始化配置。
	// TODO 端口和环境变量都可以自定义
	// go run ./cmd/api -port=3030 -env=production
	flag.IntVar(&cfg.port, "port", 4000, "API 服务器端口")
	flag.StringVar(&cfg.env, "env", "development", "环境 (development|staging|production)")
	flag.Parse()

	// 初始化日志记录器。
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	//logger := jsonlog.NewLogger(os.Stdout, jsonlog.LevelInfo)

	app := &application{
		config: cfg,
		logger: logger,
	}

	// 初始化请求路由器。
	//mux := http.NewServeMux()
	//mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// 配置 HTTP 服务器。
	// 使用 httprouter 作为请求路由器，并设置相关配置。
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 启动 HTTP 服务器。
	logger.Printf("正在启动 %s 环境下的服务器，监听地址为 %s", cfg.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}
