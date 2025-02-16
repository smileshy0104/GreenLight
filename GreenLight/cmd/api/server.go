package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// 创建一个http.Server结构体
	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", app.config.port), // 设置监听地址和端口
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	// 创建一个通道，用于接收优雅关闭的错误信号
	shutdownError := make(chan error, 1)

	// 创建一个goroutine，用于处理SIGINT和SIGTERM信号，并记录任何错误。
	go func() {
		// 创建一个通道，用于接收信号
		quit := make(chan os.Signal, 1)
		// 使用signal.Notify()函数注册quit通道以接收SIGINT和SIGTERM信号。
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		// 创建一个通道，用于接收信号
		s := <-quit

		// 使用app.logger.PrintInfo()记录接收到的信号。
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})

		// 创建一个带有超时的上下文，并调用srv.Shutdown()方法来关闭服务器。
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// 使用srv.Shutdown()方法关闭服务器，并记录任何错误。
		// 利用优雅的关闭，它不会中断任何活动连接的服务器。首先关闭所有打开的侦听器，然后关闭所有空闲的连接，然后无限期地等待连接恢复到空闲状态，然后关闭。
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		//// 使用app.logger.PrintInfo()记录接收到的信号。
		//app.logger.PrintInfo("caught signal", map[string]string{
		//	"signal": s.String(),
		//})
		//// 退出程序
		//os.Exit(0)
	}()

	// 使用srv.ListenAndServe()方法启动服务器，并记录任何错误。
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// 创建一个goroutine，用于等待服务器关闭，并记录任何错误。
	err = <-shutdownError
	if err != nil {
		return err
	}

	// 使用app.logger.PrintInfo()记录服务器已关闭。
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
