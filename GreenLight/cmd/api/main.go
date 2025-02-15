package main

import (
	configFile "DesignMode/GreenLight/internal/config"
	"DesignMode/GreenLight/internal/data"
	"context" // New import
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	// Import the pq driver so that it can register itself with the database/sql
	// package. Note that we alias this import to the blank identifier, to stop the Go
	// compiler complaining that the package isn't being used.
	_ "github.com/lib/pq"
)

// 定义应用程序的版本号。
const version = "1.0.0"

// 配置结构体，用于存储服务器端口和环境变量。
type config struct {
	port int    // 端口
	env  string // 环境
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

//go:embed config/*
var f embed.FS

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

	// TODO  postgres://username:password@localhost/db_name?sslmode=disable
	// 初始化配置文件
	configFile.InitConfig(f)
	postgreSqlDsn := configFile.AppConf.GetString("database.dsn")

	// 设置命令行参数，用于配置数据库连接信息。q
	flag.StringVar(&cfg.db.dsn, "db-dsn", postgreSqlDsn, "PostgreSQL DSN")
	// 设置数据库连接池配置
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	// 解析命令行参数以初始化配置。
	flag.Parse()

	// 初始化日志记录器。
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	//logger := jsonlog.NewLogger(os.Stdout, jsonlog.LevelInfo)

	// 尝试打开数据库连接池。
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	// 创建一个日志记录器，并记录一条消息，表示数据库连接池已建立。
	logger.Printf("database connection pool established")

	// 创建应用程序结构体，并初始化相关配置和日志记录器。
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	// 初始化请求路由器。
	// TODO 使用http.NewServeMux()功能非常受限，不方便使用。
	//mux := http.NewServeMux()
	//mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// 配置 HTTP 服务器。
	// TODO 使用 httprouter 作为请求路由器，并设置相关配置。
	srv := &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%d", cfg.port), // 设置监听地址和端口
		//Handler:      mux,                 // 设置请求路由器
		Handler:      app.routes(),     // 设置请求路由器
		IdleTimeout:  time.Minute,      // 空闲超时
		ReadTimeout:  10 * time.Second, // 读取超时
		WriteTimeout: 30 * time.Second, // 写入超时
	}

	// 启动 HTTP 服务器。
	logger.Printf("正在启动 %s 环境下的服务器，监听地址为 %s", cfg.env, srv.Addr)
	// 使用srv.ListenAndServe()方法启动服务器，并记录任何错误。
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

// openDB()函数用于创建并返回一个数据库连接池。
func openDB(cfg config) (*sql.DB, error) {
	// 使用数据库连接池配置，创建并返回一个数据库连接池。
	conn, err := gorm.Open(mysql.Open(cfg.db.dsn))
	if err != nil {
		return nil, err
	}
	//设置数据库连接池参数
	sqlDB, _ := conn.DB()
	// 设置连接池的最大打开连接数。
	sqlDB.SetMaxOpenConns(cfg.db.maxOpenConns)
	// 设置连接池的最大空闲连接数。
	sqlDB.SetMaxIdleConns(cfg.db.maxIdleConns)
	// 设置连接池中每个连接的最大空闲时间。
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	sqlDB.SetConnMaxIdleTime(duration)

	// 创建一个上下文，并设置5秒超时。
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// 使用PingContext()方法执行ping操作，以确认连接池是否工作正常。
	err = sqlDB.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return sqlDB, nil
}
