package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	broker    = "tcp://broker.emqx.io:1883"
	topic     = "go/mqtt/demo"
	msgCount  = 5
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected to broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v\n", err)
}

func main() {
	clientID := fmt.Sprintf("go-mqtt-demo-%d", time.Now().UnixNano())

	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	// Subscribe
	var wg sync.WaitGroup
	wg.Add(msgCount)

	token := client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
		wg.Done()
	})
	if token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	fmt.Printf("Subscribed to topic: %s\n", topic)

	// Publish messages
	for i := 1; i <= msgCount; i++ {
		text := fmt.Sprintf("message %d", i)
		token := client.Publish(topic, 1, false, text)
		if token.Wait() && token.Error() != nil {
			log.Printf("Publish failed: %v", token.Error())
		} else {
			fmt.Printf("Published: %s\n", text)
		}
		time.Sleep(time.Second)
	}

	// Wait for all messages to be received
	wg.Wait()

	client.Unsubscribe(topic)
	client.Disconnect(250)
	fmt.Println("Disconnected")
}
