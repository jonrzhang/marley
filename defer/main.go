// 场景：数据库事务管理
// 执行一组操作，任何一步失败则自动回滚，成功则提交；文件操作也确保关闭
package main

import (
	"errors"
	"fmt"
)

// --- 模拟数据库事务 ---

type TX struct {
	ops    []string
	closed bool
}

func beginTX() *TX {
	fmt.Println("[DB] Transaction started")
	return &TX{}
}

func (tx *TX) Exec(sql string) error {
	if sql == "INSERT INTO orders FAIL" {
		return errors.New("duplicate key violation")
	}
	tx.ops = append(tx.ops, sql)
	fmt.Printf("[DB] Exec: %s\n", sql)
	return nil
}

func (tx *TX) Commit() error {
	fmt.Printf("[DB] Committed (%d ops)\n", len(tx.ops))
	tx.closed = true
	return nil
}

func (tx *TX) Rollback() {
	if !tx.closed {
		fmt.Printf("[DB] Rolled back (%d ops undone)\n", len(tx.ops))
		tx.closed = true
	}
}

// createOrder 演示 defer 确保事务一定被关闭
func createOrder(userID int, fail bool) error {
	tx := beginTX()
	// defer 在函数返回前执行，无论正常还是 error 返回
	// 如果已 commit，Rollback 是空操作
	defer tx.Rollback()

	if err := tx.Exec(fmt.Sprintf("INSERT INTO orders (user_id) VALUES (%d)", userID)); err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	orderSQL := "INSERT INTO order_items (product_id) VALUES (42)"
	if fail {
		orderSQL = "INSERT INTO orders FAIL"
	}
	if err := tx.Exec(orderSQL); err != nil {
		return fmt.Errorf("insert items: %w", err)
	}

	if err := tx.Exec("UPDATE inventory SET stock = stock - 1 WHERE product_id = 42"); err != nil {
		return fmt.Errorf("update inventory: %w", err)
	}

	return tx.Commit() // commit 成功后 Rollback 变为空操作
}

// --- 模拟文件操作 ---

type File struct{ name string }

func openFile(name string) *File {
	fmt.Printf("\n[FILE] Opened %s\n", name)
	return &File{name: name}
}

func (f *File) Close() {
	fmt.Printf("[FILE] Closed %s\n", f.name)
}

func processReport(filename string) {
	f := openFile(filename)
	defer f.Close() // 无论函数如何返回，文件必定关闭

	fmt.Printf("[FILE] Processing %s...\n", f.name)
}

func main() {
	fmt.Println("=== Order 1: success ===")
	if err := createOrder(101, false); err != nil {
		fmt.Println("Error:", err)
	}

	fmt.Println("\n=== Order 2: failure ===")
	if err := createOrder(102, true); err != nil {
		fmt.Println("Error:", err)
	}

	processReport("monthly_report.csv")
}
