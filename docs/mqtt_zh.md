# MQTT — 发布/订阅消息

## 概述

本模块使用 [Eclipse Paho](https://github.com/eclipse/paho.mqtt.golang) Go 客户端演示 MQTT 协议。程序连接公共 Broker，订阅一个 Topic，发布 5 条消息，并使用 `sync.WaitGroup` 确保所有消息都被接收后再断开连接。

## 使用场景

MQTT 是专为受限设备和不稳定网络（IoT 传感器、遥测、移动应用）设计的轻量级发布/订阅协议。与 HTTP 相比：
- **持久连接** — 无需每条消息都建立 TCP 握手
- **QoS 等级** — 可配置的投递保障（最多一次、至少一次、恰好一次）
- **Topic 路由** — Broker 自动将消息推送给所有订阅者

实际应用场景：传感器数据采集、设备指令下发、实时仪表盘。

## 代码解析

### 连接配置

```go
clientID := fmt.Sprintf("go-mqtt-demo-%d", time.Now().UnixNano())

opts := mqtt.NewClientOptions()
opts.AddBroker(broker)          // "tcp://broker.emqx.io:1883"
opts.SetClientID(clientID)
opts.OnConnect = connectHandler
opts.OnConnectionLost = connectLostHandler
```

**clientID 在 Broker 上必须唯一**——两个相同 clientID 的客户端会互相踢掉对方。使用 `time.Now().UnixNano()` 作为后缀，保证每次运行都不重复。

`OnConnect` 和 `OnConnectionLost` 是生命周期回调，在连接状态变化时自动触发。

### 建立连接

```go
client := mqtt.NewClient(opts)
if token := client.Connect(); token.Wait() && token.Error() != nil {
    log.Fatal(token.Error())
}
```

Paho 所有操作都返回 `Token`。调用 `token.Wait()` 阻塞直到操作完成，`token.Error()` 返回错误（如有）。Connect、Subscribe、Publish 都使用这个模式。

### 带 WaitGroup 的订阅

```go
var wg sync.WaitGroup
wg.Add(msgCount)

token := client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
    fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
    wg.Done()
})
if token.Wait() && token.Error() != nil {
    log.Fatal(token.Error())
}
```

内联回调函数在 Paho 管理的 goroutine 中**异步触发**，每次消息到达时调用一次。`wg.Done()` 递减计数器。在订阅前调用 `wg.Add(msgCount)` 确保计数器在消息到达之前就已设置好。

**QoS 1**（至少一次）表示 Broker 会重试投递，直到订阅者确认收到。

### 发布消息

```go
for i := 1; i <= msgCount; i++ {
    text := fmt.Sprintf("message %d", i)
    token := client.Publish(topic, 1, false, text)
    if token.Wait() && token.Error() != nil {
        log.Printf("Publish failed: %v", token.Error())
    }
    time.Sleep(time.Second)
}
```

`Publish(topic, qos, retained, payload)`：
- `retained=false` — Broker 不为新订阅者保存最后一条消息
- `time.Sleep` 在发布间加入间隔，让输出更清晰

### 等待所有消息接收

```go
wg.Wait() // 阻塞直到所有 msgCount 条消息都被接收
client.Unsubscribe(topic)
client.Disconnect(250) // 250ms 宽限期，等待在途消息完成
```

`wg.Wait()` 比固定的 `time.Sleep` 可靠得多——恰好在所有期望消息到达时退出，不早也不晚。

## 核心知识点

- **Token 模式** — 所有 Paho 异步操作返回 Token；始终调用 `token.Wait()` 并检查 `token.Error()`
- **动态 clientID** — 防止同一程序多次运行时在 Broker 上发生冲突
- **QoS 等级** — 0（发送即忘）、1（至少一次）、2（恰好一次）
- **WaitGroup 配合异步回调** — 与消息驱动代码同步的正确方式
- **`Disconnect(quiesce)`** — 宽限期允许在途消息在 TCP 关闭前完成

## 适用场景

| 场景 | 建议 |
|------|------|
| IoT 设备 → 服务器 | TCP 或 WebSocket 上的 MQTT，QoS 1 |
| 多设备发布到同一 Topic | Broker 处理扇出；Go 只需订阅一次 |
| 需要恰好一次投递 | 使用 QoS 2（开销更高）|
| 简单请求-响应 | HTTP 更简单，MQTT 过于复杂 |
