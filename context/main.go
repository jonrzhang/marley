// 场景：HTTP API 客户端
// 调用外部 API 时需要：超时控制、支持取消、请求 ID 贯穿整个调用链
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type APIResponse struct {
	StatusCode int
	Body       string
}

// requestIDKey 用自定义类型避免 context key 冲突
type contextKey string

const requestIDKey contextKey = "requestID"

// httpGet 模拟 HTTP 请求，从 context 读取 requestID 用于日志追踪
func httpGet(ctx context.Context, url string) (*APIResponse, error) {
	reqID, _ := ctx.Value(requestIDKey).(string)
	fmt.Printf("[%s] GET %s\n", reqID, url)

	// 模拟网络延迟
	delay := time.Duration(rand.Intn(300)+100) * time.Millisecond

	select {
	case <-time.After(delay):
		return &APIResponse{StatusCode: 200, Body: `{"status":"ok"}`}, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("[%s] request to %s: %w", reqID, url, ctx.Err())
	}
}

// fetchUserOrders 组合多个 API 调用，共享同一个 context
func fetchUserOrders(ctx context.Context, userID int) error {
	// 串行调用：每步都检查 context 是否已取消
	user, err := httpGet(ctx, fmt.Sprintf("/api/users/%d", userID))
	if err != nil {
		return fmt.Errorf("fetch user: %w", err)
	}
	fmt.Printf("  User response: %s\n", user.Body)

	orders, err := httpGet(ctx, fmt.Sprintf("/api/users/%d/orders", userID))
	if err != nil {
		return fmt.Errorf("fetch orders: %w", err)
	}
	fmt.Printf("  Orders response: %s\n", orders.Body)

	return nil
}

func main() {
	// 1. 正常请求：带超时和 requestID
	ctx := context.WithValue(context.Background(), requestIDKey, "req-abc-001")
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	fmt.Println("=== Request 1: normal ===")
	if err := fetchUserOrders(ctx, 42); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Success")
	}

	// 2. 超时请求：超时时间设得很短
	ctx2 := context.WithValue(context.Background(), requestIDKey, "req-xyz-002")
	ctx2, cancel2 := context.WithTimeout(ctx2, 50*time.Millisecond)
	defer cancel2()

	fmt.Println("\n=== Request 2: timeout ===")
	if err := fetchUserOrders(ctx2, 99); err != nil {
		fmt.Println("Error:", err)
	}

	// 3. 手动取消：模拟用户中断请求
	ctx3 := context.WithValue(context.Background(), requestIDKey, "req-zzz-003")
	ctx3, cancel3 := context.WithCancel(ctx3)

	fmt.Println("\n=== Request 3: manual cancel ===")
	go func() {
		time.Sleep(80 * time.Millisecond)
		fmt.Println("  [client] user cancelled the request")
		cancel3()
	}()
	if err := fetchUserOrders(ctx3, 77); err != nil {
		fmt.Println("Error:", err)
	}
}
