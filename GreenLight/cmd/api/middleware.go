package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
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

// 创建个中间件rateLimit，用于限制请求速率
func (app *application) rateLimitOld0(next http.Handler) http.Handler {
	// 初始化一个rate.Limiter对象，并设置其速率为2个请求/秒，并且最多允许4个请求。
	limiter := rate.NewLimiter(2, 4)
	// 使用defer关键字，在函数返回时执行limiter.Allow()。
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果请求速率超出了限制，则调用rateLimitExceededResponse函数，并返回。
		// Allow（）方法是由互斥锁保护，并且对于并发使用是安全的！
		if !limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// 创建个中间件rateLimitNew，用于限制请求速率（其中包含对应IP地址限制）
func (app *application) rateLimitOld1(next http.Handler) http.Handler {
	// 初始化一个map，用于存储IP地址和对应的rate.Limiter对象。
	var (
		mu      sync.Mutex
		clients = make(map[string]*rate.Limiter)
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 取出请求的IP地址
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		// 使用mu.Lock()和mu.Unlock()来锁定和释放互斥锁，以保护对clients map的并发访问。
		mu.Lock()
		// 如果该IP地址不在map中，则创建一个新的rate.Limiter对象，并将其添加到map中。
		if _, found := clients[ip]; !found {
			// 设置其速率为2个请求/秒，并且最多允许4个请求。
			clients[ip] = rate.NewLimiter(2, 4)
		}
		// 如果对应IP的请求速率超出了限制，则调用rateLimitExceededResponse函数，并返回。
		if !clients[ip].Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}
		// 释放互斥锁
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}

// 创建个中间件rateLimit，用于限制请求速率（其中包含对应IP地址限制和IP存活时间限制）
// （这种reteLimit只能作用于单台机器的情况，若为分布式情况，则无法适应）
// 当为分布式系统时，我们应该使用HAProxy或Nginx作为负载均衡器或反向代理，它们都有内置的速率限制功能！
func (app *application) rateLimit(next http.Handler) http.Handler {
	// 初始化一个结构体，用于存储IP地址和对应的rate.Limiter对象和最后使用时间。
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	// 声明对应的互斥锁和client
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// 创建一个goroutine，用于清理过期的IP地址。
	go func() {
		for {
			time.Sleep(time.Minute)

			// 上锁
			mu.Lock()

			// 遍历clients，并删除超过3分钟没有使用的IP地址
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// 释放互斥锁
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 取出请求的IP地址
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		mu.Lock()
		// 如果该IP地址不在map中，则创建一个新的rate.Limiter对象，并将其添加到map中。
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
		}
		// 更新最后使用时间
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
