# MQTT — Publish / Subscribe Messaging

## Overview

This module demonstrates the MQTT protocol using the [Eclipse Paho](https://github.com/eclipse/paho.mqtt.golang) Go client. It connects to a public broker, subscribes to a topic, publishes 5 messages, and uses `sync.WaitGroup` to ensure all messages are received before disconnecting.

## Use Case

MQTT is a lightweight pub/sub protocol designed for constrained devices and unreliable networks (IoT sensors, telemetry, mobile apps). Compared to HTTP:
- **Persistent connection** — no per-message TCP handshake overhead
- **QoS levels** — configurable delivery guarantees (at-most-once, at-least-once, exactly-once)
- **Topic routing** — broker fans out messages to all subscribers of a topic

Real-world uses: sensor data collection, device command dispatch, real-time dashboards.

## Code Walkthrough

### Connection Setup

```go
clientID := fmt.Sprintf("go-mqtt-demo-%d", time.Now().UnixNano())

opts := mqtt.NewClientOptions()
opts.AddBroker(broker)          // "tcp://broker.emqx.io:1883"
opts.SetClientID(clientID)
opts.OnConnect = connectHandler
opts.OnConnectionLost = connectLostHandler
```

The **clientID must be unique** per active connection on a broker — two clients with the same ID will kick each other off. Using `time.Now().UnixNano()` as a suffix guarantees uniqueness across program runs.

`OnConnect` and `OnConnectionLost` are lifecycle callbacks that fire automatically when the connection state changes.

### Connecting

```go
client := mqtt.NewClient(opts)
if token := client.Connect(); token.Wait() && token.Error() != nil {
    log.Fatal(token.Error())
}
```

All Paho operations return a `Token`. Calling `token.Wait()` blocks until the operation completes, and `token.Error()` returns any failure. This pattern is used for Connect, Subscribe, and Publish.

### Subscribing with WaitGroup

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

The inline callback fires **asynchronously** in a Paho-managed goroutine each time a message arrives. `wg.Done()` counts down the WaitGroup. Using `wg.Add(msgCount)` before subscribing guarantees the counter is set before any messages can arrive.

**QoS 1** (`at-least-once`) means the broker will retry delivery until the subscriber acknowledges receipt.

### Publishing

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

`Publish(topic, qos, retained, payload)`:
- `retained=false` — the broker does not store the last message for new subscribers
- `time.Sleep` adds spacing between publishes, making the output easier to follow

### Waiting for All Messages

```go
wg.Wait() // blocks until all msgCount messages are received
client.Unsubscribe(topic)
client.Disconnect(250) // 250ms grace period for in-flight messages
```

`wg.Wait()` is far more reliable than a fixed `time.Sleep` — it exits exactly when all expected messages arrive, no sooner, no later.

## Key Concepts

- **Token pattern** — all async Paho operations return a Token; always call `token.Wait()` and check `token.Error()`
- **Dynamic clientID** — prevents broker conflicts when the same program runs multiple times
- **QoS levels** — 0 (fire-and-forget), 1 (at-least-once), 2 (exactly-once)
- **WaitGroup for async callbacks** — the correct way to synchronize with message-driven code
- **`Disconnect(quiesce)`** — the grace period allows in-flight messages to complete before the TCP connection closes

## When to Use

| Situation | Recommendation |
|-----------|---------------|
| IoT device → server | MQTT over TCP or WebSocket with QoS 1 |
| Many devices publishing to one topic | Broker handles fan-out; Go subscribes once |
| Need exactly-once delivery | QoS 2 (higher overhead) |
| Simple request-response | HTTP is simpler; MQTT is overkill |
