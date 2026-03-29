// 场景：服务健康检查
// 同时探测多个服务，哪个先响应就用哪个，超过阈值则报告不可用
package main

import (
	"fmt"
	"math/rand"
	"time"
)

type HealthResult struct {
	Service string
	Latency time.Duration
	Healthy bool
}

// probe 模拟对某个服务发起健康检查
func probe(service string, latency time.Duration) <-chan HealthResult {
	ch := make(chan HealthResult, 1)
	go func() {
		time.Sleep(latency)
		ch <- HealthResult{
			Service: service,
			Latency: latency,
			Healthy: rand.Intn(5) != 0, // 80% 概率健康
		}
	}()
	return ch
}

func checkAll(services map[string]time.Duration, timeout time.Duration) {
	type result struct {
		name string
		ch   <-chan HealthResult
	}

	probes := make([]result, 0, len(services))
	for name, latency := range services {
		probes = append(probes, result{name, probe(name, latency)})
	}

	deadline := time.After(timeout)
	received := 0

	for received < len(probes) {
		// select 动态等待所有 channel，超时则放弃
		select {
		case r := <-probes[0].ch:
			printResult(r)
			probes = probes[1:]
			received++
		case r := <-probes[min(1, len(probes)-1)].ch:
			printResult(r)
			if len(probes) > 1 {
				probes = append(probes[:1], probes[2:]...)
			}
			received++
		case <-deadline:
			fmt.Printf("[TIMEOUT] %d service(s) did not respond within %v\n",
				len(probes)-received, timeout)
			return
		}
	}
}

func printResult(r HealthResult) {
	status := "UP"
	if !r.Healthy {
		status = "DOWN"
	}
	fmt.Printf("[%-4s] %-20s latency=%v\n", status, r.Service, r.Latency.Round(time.Millisecond))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	services := map[string]time.Duration{
		"auth-service":    80 * time.Millisecond,
		"user-service":    120 * time.Millisecond,
		"payment-service": 200 * time.Millisecond,
		"slow-service":    600 * time.Millisecond, // 会超时
	}

	fmt.Println("Checking services (timeout=500ms)...")
	checkAll(services, 500*time.Millisecond)
}
