# Cgo — 系统资源监控

## 概述

本模块演示 **cgo**——Go 直接调用 C 代码的机制。它实现了一个系统资源监控器，调用 Linux 的 `sysinfo()` 系统调用来获取内存使用情况、系统运行时间、CPU 负载均衡和进程数量。这些指标无法通过 Go 标准库获取，因此 cgo 是正确的工具。

## 使用场景

以下情况适合使用 cgo：
1. 所需 API 只在 C 中可用（OS 系统调用、硬件驱动、遗留库）
2. 关键热路径需要手工优化的 C 例程
3. 与 C 库集成（OpenSSL、SQLite、libvips）

本例中，Linux 的 `sysinfo()` 通过单次系统调用提供系统级内存和负载统计——比解析 `/proc/meminfo` 更高效。Go 标准库（`runtime.MemStats`）只报告 Go 堆的统计信息，不报告系统总内存。

## 代码解析

### 前言块（Preamble）— 内联 C 代码

```go
package main

/*
#include <sys/sysinfo.h>
#include <unistd.h>

typedef struct {
    long uptime_seconds;
    unsigned long total_ram;
    ...
} sys_stats;

int get_sys_stats(sys_stats *stats) {
    struct sysinfo si;
    if (sysinfo(&si) != 0) {
        return -1;
    }
    stats->uptime_seconds = si.uptime;
    stats->total_ram      = si.totalram * si.mem_unit;
    ...
    stats->load_1  = si.loads[0] / 65536.0;
    ...
    return 0;
}

int get_cpu_count() {
    return sysconf(_SC_NPROCESSORS_ONLN);
}
*/
import "C"
```

C 代码放在紧跟 `import "C"` **之前**的 `/* */` 注释块中——注释和 import 之间**不能有空行**。这就是**前言块（preamble）**。cgo 工具提取它并作为 C 翻译单元编译。

编写了一个薄薄的 C 包装器（`get_sys_stats`），而不是直接从 Go 调用 `sysinfo`。这简化了 Go 端：一次函数调用、一个错误码、一个预填充的结构体。

**`loads` 定点数转换：**
```c
stats->load_1 = si.loads[0] / 65536.0;
```
Linux 将负载均衡存储为按 2^16（65536）缩放的定点整数。除以 65536.0 将其转换为我们熟悉的小数值（如 0.07）。

### 从 Go 调用 C

```go
func getSystemStats() (SystemStats, error) {
    var cStats C.sys_stats       // C 结构体类型，分配在 Go 栈上

    ret := C.get_sys_stats(&cStats)   // 调用 C 函数，传入指针
    if ret != 0 {
        return SystemStats{}, fmt.Errorf("sysinfo() failed with code %d", ret)
    }

    cpuCount := C.get_cpu_count()

    return SystemStats{
        Uptime:   time.Duration(cStats.uptime_seconds) * time.Second,
        TotalRAM: uint64(cStats.total_ram),
        ...
    }, nil
}
```

- `C.sys_stats` — C 结构体类型，通过 `C` 伪包访问
- `C.get_sys_stats(&cStats)` — 调用 C 函数；`&cStats` 传入指针
- 所有 C 类型通过 `C.` 前缀访问：`C.int`、`C.long`、`C.sys_stats`
- 每次 C→Go 转换都需要**显式类型转换**：`uint64(cStats.total_ram)`、`int(cpuCount)`

### C ↔ Go 类型映射

| C 类型 | Go 等价类型 |
|--------|------------|
| `int` | `C.int` → `int32` |
| `long` | `C.long` → `int32` 或 `int64`（平台相关）|
| `unsigned long` | `C.ulong` → `uint32` 或 `uint64` |
| `unsigned short` | `C.ushort` → `uint16` |
| `float` | `C.float` → `float32` |
| `char *` | `C.CString()` / `C.GoString()` 转换 |

始终进行显式转换——Go 不会隐式转换 C 和 Go 数值类型之间的值。

## 核心知识点

- **前言块** — `import "C"` 紧前方的 `/* */` 注释中的 C 代码；被编译为 C
- **`import "C"`** — 提供访问 C 类型和函数的伪包
- **`C.TypeName`** — 访问 C 类型；`C.FuncName(...)` — 调用 C 函数
- **显式转换** — 每个 C↔Go 类型边界都需要显式转换
- **薄 C 包装器** — 编写简单的 C 函数将 C 惯用法转换为 Go 友好的 API
- **CGO_ENABLED** — 设为 0 可禁用 cgo 以支持交叉编译；cgo 二进制文件需要 C 工具链

## 适用场景

| 场景 | 建议 |
|------|------|
| 需要标准库中没有的 Linux/macOS 系统调用 | 适合使用 cgo |
| 集成现有 C 库 | cgo + `// #cgo LDFLAGS: -lname` |
| 交叉编译到其他平台 | 避免 cgo；使用纯 Go 或 syscall 包 |
| 性能关键的内部循环 | 先做性能分析；cgo 每次调用约有 ~100ns 开销 |
| 简单的字符串/字节处理 | 使用纯 Go；cgo 开销不值得 |
