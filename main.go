package main

import (
	"fmt"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gobwas/ws"
	"github.com/labstack/echo/v4"
)

func main() {

	// Websocket ---------------------------------------------------------------------------
	deviceHub := NewHub()
	go deviceHub.Run()

	e := echo.New()
	e.Static("/", "./public")

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			// Bypass root path
			if c.Request().URL.Path == "/" {
				return next(c)
			}

			// Get token
			token := c.QueryParam("token")

			// Validate token
			jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
			if token == jwt {
				return next(c)
			}

			return echo.ErrUnauthorized
		}
	})

	e.Any("/device/:id", func(c echo.Context) error {

		id := c.Param("id")
		fmt.Println("id:", id)

		// Upgrade websocket
		conn, _, _, err := ws.UpgradeHTTP(c.Request(), c.Response())
		if err != nil {
			fmt.Println(err)
		}

		// Register client
		client := &Client{hub: deviceHub, conn: conn, send: make(chan []byte, 256)}
		client.hub.register <- client

		// Write message
		go client.WriteMessage()

		return nil
	})

	e.GET("/device/publish", func(c echo.Context) error {

		// Broadcast
		deviceHub.Broadcast([]byte(`{"id": "1e4832e7-1ffa-4cf4-b9d9-0b8eff286c52", "name": "Temp"}`))

		return c.JSON(http.StatusOK, echo.Map{"message": "published"})
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

	e.Logger.Fatal(e.Start(":8080"))
}
