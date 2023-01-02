package main

import (
	"fmt"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gobwas/ws"
)

type Device struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func main() {

	// Websocket ---------------------------------------------------------------------------
	deviceHub := NewHub()
	go deviceHub.Run()

	http.HandleFunc("/device", func(w http.ResponseWriter, r *http.Request) {

		token := r.URL.Query().Get("token")
		fmt.Println("token:", token)

		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			fmt.Println(err)
		}

		client := &Client{hub: deviceHub, conn: conn, send: make(chan []byte, 256)}
		client.hub.register <- client

		go client.WriteMessage()
	})

	// Mqtt --------------------------------------------------------------------------------
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://broker.emqx.io:1883").
		SetClientID("emqx_test_client").
		SetUsername("emqx_test").
		SetPassword("emqx_test")

	opts.SetKeepAlive(60 * time.Second)
	// Set the message callback handler
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		fmt.Println(msg.Topic(), ":", string(msg.Payload()))

		// Broadcast
		deviceHub.Broadcast(msg.Payload())
	})
	opts.SetPingTimeout(1 * time.Second)

	c := mqtt.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		fmt.Println("Mqtt connected")
	}

	// Subscribe to a topic
	if token := c.Subscribe("device", 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	fmt.Println("Listening *:8080")
	http.ListenAndServe(":8080", nil)
}
