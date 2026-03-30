# Cgo — System Resource Monitor

## Overview

This module demonstrates **cgo** — Go's mechanism for calling C code directly. It implements a system resource monitor that calls the Linux `sysinfo()` syscall to retrieve memory usage, system uptime, CPU load averages, and process count. These metrics are unavailable through Go's standard library, making cgo the right tool.

## Use Case

Cgo is used when:
1. A required API is only available in C (OS syscalls, hardware drivers, legacy libraries)
2. A critical hot path needs a hand-optimized C routine
3. Integrating with C libraries (OpenSSL, SQLite, libvips)

In this case, Linux's `sysinfo()` provides system-level memory and load statistics in a single syscall — more efficient than parsing `/proc/meminfo`. Go's standard library (`runtime.MemStats`) only reports Go heap stats, not total system memory.

## Code Walkthrough

### The Preamble — Inline C Code

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

The C code lives in a **comment block** immediately before `import "C"` — no blank lines between the comment and the import. This is the **preamble**. The cgo tool extracts it and compiles it as a C translation unit.

A thin C wrapper (`get_sys_stats`) is written instead of calling `sysinfo` directly from Go. This simplifies the Go side: one function call, one error code, and a pre-populated struct.

**The `loads` fixed-point conversion:**
```c
stats->load_1 = si.loads[0] / 65536.0;
```
Linux stores load averages as fixed-point integers scaled by 2^16 (65536). Dividing by 65536.0 converts them to the familiar decimal values (e.g., 0.07).

### Calling C from Go

```go
func getSystemStats() (SystemStats, error) {
    var cStats C.sys_stats       // C struct type, allocated on Go stack

    ret := C.get_sys_stats(&cStats)   // call C function, pass pointer
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

- `C.sys_stats` — the C struct type, accessible via the `C` pseudo-package
- `C.get_sys_stats(&cStats)` — calls the C function; `&cStats` passes a pointer
- All C types are accessed through the `C.` prefix: `C.int`, `C.long`, `C.sys_stats`
- **Explicit type conversion** is required for every C→Go transition: `uint64(cStats.total_ram)`, `int(cpuCount)`

### C ↔ Go Type Mapping

| C type | Go equivalent |
|--------|--------------|
| `int` | `C.int` → `int32` |
| `long` | `C.long` → `int32` or `int64` (platform-dependent) |
| `unsigned long` | `C.ulong` → `uint32` or `uint64` |
| `unsigned short` | `C.ushort` → `uint16` |
| `float` | `C.float` → `float32` |
| `char *` | `C.CString()` / `C.GoString()` for conversion |

Always cast explicitly — Go won't implicitly convert between C and Go numeric types.

## Key Concepts

- **Preamble** — C code in a `/* */` comment immediately before `import "C"`; compiled as C
- **`import "C"`** — the pseudo-package that provides access to C types and functions
- **`C.TypeName`** — access C types; `C.FuncName(...)` — call C functions
- **Explicit conversions** — every C↔Go type boundary requires an explicit cast
- **Thin C wrapper** — write a simple C function to translate C idioms into Go-friendly APIs
- **CGO_ENABLED** — set to 0 to disable cgo for cross-compilation; cgo binaries require the C toolchain

## When to Use (and When Not To)

| Situation | Recommendation |
|-----------|---------------|
| Need Linux/macOS system calls not in stdlib | Cgo is appropriate |
| Integrating an existing C library | Cgo + `// #cgo LDFLAGS: -lname` |
| Cross-compiling to another platform | Avoid cgo; use pure Go or syscall package |
| Performance-critical inner loop | Profile first; cgo has call overhead (~100ns per call) |
| Simple string/byte manipulation | Use pure Go; cgo overhead isn't worth it |
