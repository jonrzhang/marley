// 场景：并行下载图片
// 有一批图片 URL，逐个下载太慢，用 goroutine 并发下载，最后汇总结果
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type DownloadResult struct {
	URL      string
	Size     int
	Duration time.Duration
	Err      error
}

// download 模拟下载一张图片（用随机延迟模拟网络请求）
func download(url string) DownloadResult {
	start := time.Now()
	delay := time.Duration(rand.Intn(300)+100) * time.Millisecond
	time.Sleep(delay)

	// 模拟偶发失败
	if rand.Intn(10) == 0 {
		return DownloadResult{URL: url, Err: fmt.Errorf("connection timeout")}
	}

	return DownloadResult{
		URL:      url,
		Size:     rand.Intn(500) + 100, // KB
		Duration: time.Since(start),
	}
}

func main() {
	urls := []string{
		"https://example.com/img/photo1.jpg",
		"https://example.com/img/photo2.jpg",
		"https://example.com/img/photo3.jpg",
		"https://example.com/img/photo4.jpg",
		"https://example.com/img/photo5.jpg",
	}

	results := make([]DownloadResult, len(urls))
	var wg sync.WaitGroup

	start := time.Now()
	for i, url := range urls {
		wg.Add(1)
		go func(idx int, u string) {
			defer wg.Done()
			results[idx] = download(u)
		}(i, url)
	}
	wg.Wait()
	total := time.Since(start)

	// 统计结果
	var success, failed, totalSize int
	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("[FAIL] %s — %v\n", r.URL, r.Err)
			failed++
		} else {
			fmt.Printf("[OK]   %s — %dKB in %v\n", r.URL, r.Size, r.Duration.Round(time.Millisecond))
			totalSize += r.Size
			success++
		}
	}
	fmt.Printf("\n%d OK, %d failed, total size %dKB, elapsed %v\n",
		success, failed, totalSize, total.Round(time.Millisecond))
}
