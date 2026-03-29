// 场景：HTTP 中间件链
// 每个中间件（认证、日志、限流）是一个函数，用闭包捕获配置，链式组合
package main

import (
	"fmt"
	"strings"
	"time"
)

// HandlerFunc 模拟 HTTP 处理函数签名
type HandlerFunc func(path, token string) (int, string)

// Middleware 是包装 HandlerFunc 的函数
type Middleware func(HandlerFunc) HandlerFunc

// withLogging 记录请求耗时和状态码
func withLogging(next HandlerFunc) HandlerFunc {
	return func(path, token string) (int, string) {
		start := time.Now()
		status, body := next(path, token)
		fmt.Printf("[LOG] %s → %d (%v)\n", path, status, time.Since(start).Round(time.Microsecond))
		return status, body
	}
}

// withAuth 验证 Bearer token，闭包捕获 validToken
func withAuth(validToken string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(path, token string) (int, string) {
			if token != validToken {
				return 401, "Unauthorized"
			}
			return next(path, token)
		}
	}
}

// withRateLimit 限制访问次数，闭包捕获计数器（每个中间件实例独立）
func withRateLimit(maxReqs int) Middleware {
	count := 0 // 被闭包捕获，每个实例独立持有
	return func(next HandlerFunc) HandlerFunc {
		return func(path, token string) (int, string) {
			count++
			if count > maxReqs {
				return 429, fmt.Sprintf("Rate limit exceeded (%d/%d)", count, maxReqs)
			}
			return next(path, token)
		}
	}
}

// chain 将多个中间件按顺序叠加（最后一个最先执行）
func chain(h HandlerFunc, middlewares ...Middleware) HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func main() {
	// 业务 handler
	apiHandler := func(path, token string) (int, string) {
		return 200, fmt.Sprintf("Hello from %s", path)
	}

	// 组合中间件链：logging → auth → rateLimit → handler
	handler := chain(apiHandler,
		withLogging,
		withAuth("secret-token"),
		withRateLimit(3),
	)

	requests := []struct{ path, token string }{
		{"/api/users", "secret-token"},
		{"/api/orders", "secret-token"},
		{"/api/products", "wrong-token"}, // 401
		{"/api/users", "secret-token"},
		{"/api/users", "secret-token"}, // 429 rate limit
	}

	for _, req := range requests {
		status, body := handler(req.path, req.token)
		fmt.Printf("       %d %s\n", status, strings.TrimSpace(body))
	}
}
