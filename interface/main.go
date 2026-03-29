// 场景：通知系统
// 应用需要通过多种渠道（Email、SMS、Push）发送通知
// 新增渠道无需修改发送逻辑，只需实现 Notifier 接口
package main

import (
	"fmt"
	"strings"
)

type Message struct {
	To      string
	Subject string
	Body    string
}

// Notifier 是所有通知渠道共同遵守的契约
type Notifier interface {
	Send(msg Message) error
	Channel() string
}

// --- Email ---
type EmailNotifier struct {
	SMTPHost string
}

func (e EmailNotifier) Send(msg Message) error {
	fmt.Printf("[Email] To: %s | Subject: %s\n", msg.To, msg.Subject)
	return nil
}
func (e EmailNotifier) Channel() string { return "email" }

// --- SMS ---
type SMSNotifier struct {
	APIKey string
}

func (s SMSNotifier) Send(msg Message) error {
	body := msg.Body
	if len(body) > 70 {
		body = body[:67] + "..."
	}
	fmt.Printf("[SMS]   To: %s | %s\n", msg.To, body)
	return nil
}
func (s SMSNotifier) Channel() string { return "sms" }

// --- Push ---
type PushNotifier struct {
	AppID string
}

func (p PushNotifier) Send(msg Message) error {
	fmt.Printf("[Push]  To: %s | %s\n", msg.To, msg.Subject)
	return nil
}
func (p PushNotifier) Channel() string { return "push" }

// NotificationService 面向接口编程，不感知具体渠道
type NotificationService struct {
	notifiers []Notifier
}

func (ns *NotificationService) Register(n Notifier) {
	ns.notifiers = append(ns.notifiers, n)
}

func (ns *NotificationService) Broadcast(msg Message) {
	var failed []string
	for _, n := range ns.notifiers {
		if err := n.Send(msg); err != nil {
			failed = append(failed, n.Channel())
		}
	}
	if len(failed) > 0 {
		fmt.Printf("Failed channels: %s\n", strings.Join(failed, ", "))
	}
}

func main() {
	svc := &NotificationService{}
	svc.Register(EmailNotifier{SMTPHost: "smtp.example.com"})
	svc.Register(SMSNotifier{APIKey: "sk-xxx"})
	svc.Register(PushNotifier{AppID: "com.example.app"})

	msg := Message{
		To:      "user@example.com",
		Subject: "Your order has shipped",
		Body:    "Your order #12345 has been shipped and will arrive in 2-3 days.",
	}

	fmt.Println("Broadcasting notification:")
	svc.Broadcast(msg)
}
