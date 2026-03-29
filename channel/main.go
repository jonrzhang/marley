// 场景：日志处理 Pipeline
// 日志系统需要：读取原始日志行 → 解析为结构化数据 → 过滤并写入存储
// 三个阶段用 channel 连接，各自独立并发运行
package main

import (
	"fmt"
	"strings"
	"time"
)

type LogEntry struct {
	Level   string
	Message string
	Time    time.Time
}

// stage1: 模拟读取原始日志行
func readLines(lines []string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, line := range lines {
			out <- line
		}
	}()
	return out
}

// stage2: 解析原始行为结构化 LogEntry
func parseEntries(in <-chan string) <-chan LogEntry {
	out := make(chan LogEntry)
	go func() {
		defer close(out)
		for line := range in {
			parts := strings.SplitN(line, " ", 3)
			if len(parts) < 3 {
				continue
			}
			out <- LogEntry{
				Level:   parts[0],
				Time:    time.Now(),
				Message: parts[2],
			}
		}
	}()
	return out
}

// stage3: 只保留 ERROR 和 WARN 级别，写入"存储"
func filterAndWrite(in <-chan LogEntry) <-chan int {
	count := make(chan int, 1)
	go func() {
		n := 0
		for entry := range in {
			if entry.Level == "ERROR" || entry.Level == "WARN" {
				fmt.Printf("[STORE] %s | %s\n", entry.Level, entry.Message)
				n++
			}
		}
		count <- n
	}()
	return count
}

func main() {
	rawLogs := []string{
		"INFO  2024-01-01 Server started on :8080",
		"DEBUG 2024-01-01 Received request GET /health",
		"WARN  2024-01-01 Response time exceeded 500ms",
		"ERROR 2024-01-01 Database connection failed: timeout",
		"INFO  2024-01-01 Retry attempt 1/3",
		"ERROR 2024-01-01 Max retries exceeded, giving up",
	}

	// 构建 pipeline：readLines → parseEntries → filterAndWrite
	lines := readLines(rawLogs)
	entries := parseEntries(lines)
	count := filterAndWrite(entries)

	fmt.Printf("\nStored %d entries\n", <-count)
}
