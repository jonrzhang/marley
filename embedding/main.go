// 场景：用户权限系统
// 普通用户和管理员共享基础属性和行为，管理员在此基础上扩展权限
// 用嵌入替代继承，组合更灵活
package main

import (
	"fmt"
	"strings"
	"time"
)

// BaseUser 包含所有用户共有的字段和方法
type BaseUser struct {
	ID        int
	Name      string
	Email     string
	CreatedAt time.Time
}

func (u BaseUser) String() string {
	return fmt.Sprintf("User{id=%d, name=%s, email=%s}", u.ID, u.Name, u.Email)
}

func (u BaseUser) CanAccess(resource string) bool {
	// 普通用户只能访问 public 资源
	return strings.HasPrefix(resource, "public/")
}

// AuditLog 给需要审计的类型提供日志能力
type AuditLog struct {
	logs []string
}

func (a *AuditLog) Log(action string) {
	entry := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), action)
	a.logs = append(a.logs, entry)
}

func (a *AuditLog) PrintLogs() {
	for _, l := range a.logs {
		fmt.Println(" ", l)
	}
}

// AdminUser 嵌入 BaseUser 和 AuditLog，扩展管理员能力
type AdminUser struct {
	BaseUser               // 提升 String()、Email 等
	AuditLog               // 提升 Log()、PrintLogs()
	Permissions []string
}

// CanAccess 覆盖 BaseUser 的实现
func (a *AdminUser) CanAccess(resource string) bool {
	for _, perm := range a.Permissions {
		if strings.HasPrefix(resource, perm) {
			a.Log(fmt.Sprintf("access granted: %s", resource))
			return true
		}
	}
	a.Log(fmt.Sprintf("access denied: %s", resource))
	return false
}

func checkAccess(user interface{ CanAccess(string) bool }, resource string) {
	if user.CanAccess(resource) {
		fmt.Printf("  ✓ %s\n", resource)
	} else {
		fmt.Printf("  ✗ %s\n", resource)
	}
}

func main() {
	regular := BaseUser{ID: 1, Name: "Alice", Email: "alice@example.com", CreatedAt: time.Now()}
	admin := &AdminUser{
		BaseUser:    BaseUser{ID: 2, Name: "Bob", Email: "bob@example.com", CreatedAt: time.Now()},
		Permissions: []string{"public/", "admin/", "api/"},
	}

	fmt.Println("Regular user:", regular)
	fmt.Println("Admin user:", admin.String()) // 提升自 BaseUser

	resources := []string{"public/homepage", "admin/users", "api/orders", "internal/secrets"}

	fmt.Println("\nRegular user access:")
	for _, r := range resources {
		checkAccess(regular, r)
	}

	fmt.Println("\nAdmin user access:")
	for _, r := range resources {
		checkAccess(admin, r)
	}

	fmt.Println("\nAdmin audit log:")
	admin.PrintLogs() // 提升自 AuditLog
}
