// 场景：配置文件解析器
// 解析应用配置时，错误需要精确定位（哪个文件、哪个字段、什么问题）
// 用 %w 包装错误链，调用方用 errors.Is/As 精确处理
package main

import (
	"errors"
	"fmt"
	"strconv"
)

// --- 自定义错误类型 ---

type ParseError struct {
	File  string
	Field string
	Value string
	Cause error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse %q in %s: invalid value %q", e.Field, e.File, e.Value)
}

func (e *ParseError) Unwrap() error { return e.Cause }

var ErrMissingField = errors.New("missing required field")
var ErrOutOfRange = errors.New("value out of range")

// --- 解析函数 ---

func parsePort(raw string, file string) (int, error) {
	if raw == "" {
		return 0, &ParseError{
			File:  file,
			Field: "port",
			Value: raw,
			Cause: ErrMissingField,
		}
	}

	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, &ParseError{File: file, Field: "port", Value: raw, Cause: err}
	}

	if port < 1 || port > 65535 {
		return 0, &ParseError{
			File:  file,
			Field: "port",
			Value: raw,
			Cause: ErrOutOfRange,
		}
	}
	return port, nil
}

func loadConfig(file string, raw map[string]string) (map[string]interface{}, error) {
	config := make(map[string]interface{})

	port, err := parsePort(raw["port"], file)
	if err != nil {
		// %w 包装：调用方能用 errors.Is/As 穿透到原始错误
		return nil, fmt.Errorf("loadConfig(%s): %w", file, err)
	}
	config["port"] = port

	host := raw["host"]
	if host == "" {
		return nil, fmt.Errorf("loadConfig(%s): %w", file, &ParseError{
			File:  file,
			Field: "host",
			Value: "",
			Cause: ErrMissingField,
		})
	}
	config["host"] = host
	return config, nil
}

func main() {
	cases := []struct {
		name string
		raw  map[string]string
	}{
		{"valid", map[string]string{"port": "8080", "host": "localhost"}},
		{"missing port", map[string]string{"port": "", "host": "localhost"}},
		{"bad port", map[string]string{"port": "abc", "host": "localhost"}},
		{"port out of range", map[string]string{"port": "99999", "host": "localhost"}},
		{"missing host", map[string]string{"port": "8080", "host": ""}},
	}

	for _, c := range cases {
		cfg, err := loadConfig("app.yaml", c.raw)
		if err == nil {
			fmt.Printf("[OK]   %-20s → %v\n", c.name, cfg)
			continue
		}

		// errors.Is: 检查错误链中是否包含特定哨兵错误
		switch {
		case errors.Is(err, ErrMissingField):
			fmt.Printf("[MISS] %-20s → %v\n", c.name, err)
		case errors.Is(err, ErrOutOfRange):
			fmt.Printf("[RANGE]%-20s → %v\n", c.name, err)
		default:
			// errors.As: 提取具体错误类型获取详细信息
			var pe *ParseError
			if errors.As(err, &pe) {
				fmt.Printf("[PARSE]%-20s → field=%s file=%s\n", c.name, pe.Field, pe.File)
			}
		}
	}
}
