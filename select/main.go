// 场景：服务健康检查
// 同时探测多个服务，在超时范围内收集尽可能多的结果，超时的标记为不可用
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

// fanIn 把多个 channel 汇聚到一个
func fanIn(channels ...<-chan HealthResult) <-chan HealthResult {
	merged := make(chan HealthResult, len(channels))
	for _, ch := range channels {
		go func(c <-chan HealthResult) {
			merged <- <-c
		}(ch)
	}
	return merged
}

func main() {
	services := map[string]time.Duration{
		"auth-service":    80 * time.Millisecond,
		"user-service":    120 * time.Millisecond,
		"payment-service": 200 * time.Millisecond,
		"slow-service":    600 * time.Millisecond, // 会超时
	}

	// 启动所有探测，汇聚到一个 channel
	var probes []<-chan HealthResult
	for name, latency := range services {
		probes = append(probes, probe(name, latency))
	}
	results := fanIn(probes...)

	// select 等待结果或超时
	timeout := time.After(500 * time.Millisecond)
	received := 0

	fmt.Println("Checking services (timeout=500ms)...")
	for received < len(services) {
		select {
		case r := <-results:
			status := "UP"
			if !r.Healthy {
				status = "DOWN"
			}
			fmt.Printf("[%-4s] %-20s latency=%v\n", status, r.Service, r.Latency.Round(time.Millisecond))
			received++
		case <-timeout:
			fmt.Printf("[TIMEOUT] %d service(s) did not respond within 500ms\n", len(services)-received)
			return
		}
	}
}
