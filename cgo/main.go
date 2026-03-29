// 场景：系统资源监控
// 运维工具需要采集系统内存、CPU 负载、运行时间等信息
// Go 标准库不直接暴露 sysinfo()，通过 cgo 调用 Linux C API 获取底层数据
// 同时演示：inline C 代码、C 结构体映射、C/Go 类型转换

package main

/*
#include <sys/sysinfo.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>

// 封装一层 C 函数，Go 调用更清晰
typedef struct {
    long uptime_seconds;
    unsigned long total_ram;
    unsigned long free_ram;
    unsigned long shared_ram;
    unsigned long buffer_ram;
    unsigned short procs;
    float load_1;
    float load_5;
    float load_15;
} sys_stats;

int get_sys_stats(sys_stats *stats) {
    struct sysinfo si;
    if (sysinfo(&si) != 0) {
        return -1;
    }

    stats->uptime_seconds = si.uptime;
    stats->total_ram      = si.totalram * si.mem_unit;
    stats->free_ram       = si.freeram  * si.mem_unit;
    stats->shared_ram     = si.sharedram * si.mem_unit;
    stats->buffer_ram     = si.bufferram * si.mem_unit;
    stats->procs          = si.procs;

    // si.loads 是定点数，需要除以 65536.0 转为浮点
    stats->load_1  = si.loads[0] / 65536.0;
    stats->load_5  = si.loads[1] / 65536.0;
    stats->load_15 = si.loads[2] / 65536.0;

    return 0;
}

// 获取 CPU 核心数
int get_cpu_count() {
    return sysconf(_SC_NPROCESSORS_ONLN);
}
*/
import "C"

import (
	"fmt"
	"time"
)

type SystemStats struct {
	Uptime    time.Duration
	TotalRAM  uint64
	FreeRAM   uint64
	SharedRAM uint64
	BufferRAM uint64
	Procs     uint16
	Load1     float32
	Load5     float32
	Load15    float32
	CPUCores  int
}

func (s SystemStats) UsedRAM() uint64 {
	return s.TotalRAM - s.FreeRAM
}

func (s SystemStats) UsedPercent() float64 {
	if s.TotalRAM == 0 {
		return 0
	}
	return float64(s.UsedRAM()) / float64(s.TotalRAM) * 100
}

func getSystemStats() (SystemStats, error) {
	var cStats C.sys_stats

	// 调用 C 函数，传入结构体指针
	ret := C.get_sys_stats(&cStats)
	if ret != 0 {
		return SystemStats{}, fmt.Errorf("sysinfo() failed with code %d", ret)
	}

	cpuCount := C.get_cpu_count()

	// C 类型 → Go 类型转换
	return SystemStats{
		Uptime:    time.Duration(cStats.uptime_seconds) * time.Second,
		TotalRAM:  uint64(cStats.total_ram),
		FreeRAM:   uint64(cStats.free_ram),
		SharedRAM: uint64(cStats.shared_ram),
		BufferRAM: uint64(cStats.buffer_ram),
		Procs:     uint16(cStats.procs),
		Load1:     float32(cStats.load_1),
		Load5:     float32(cStats.load_5),
		Load15:    float32(cStats.load_15),
		CPUCores:  int(cpuCount),
	}, nil
}

func formatBytes(b uint64) string {
	const (
		MB = 1024 * 1024
		GB = 1024 * MB
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(MB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func main() {
	stats, err := getSystemStats()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("=== System Monitor (via cgo → sysinfo) ===")
	fmt.Println()

	fmt.Printf("  Uptime:      %s\n", stats.Uptime)
	fmt.Printf("  CPU Cores:   %d\n", stats.CPUCores)
	fmt.Printf("  Processes:   %d\n", stats.Procs)
	fmt.Println()

	fmt.Println("  Memory:")
	fmt.Printf("    Total:     %s\n", formatBytes(stats.TotalRAM))
	fmt.Printf("    Used:      %s (%.1f%%)\n", formatBytes(stats.UsedRAM()), stats.UsedPercent())
	fmt.Printf("    Free:      %s\n", formatBytes(stats.FreeRAM))
	fmt.Printf("    Shared:    %s\n", formatBytes(stats.SharedRAM))
	fmt.Printf("    Buffers:   %s\n", formatBytes(stats.BufferRAM))
	fmt.Println()

	fmt.Println("  Load Average:")
	fmt.Printf("    1 min:     %.2f\n", stats.Load1)
	fmt.Printf("    5 min:     %.2f\n", stats.Load5)
	fmt.Printf("    15 min:    %.2f\n", stats.Load15)
}
