package main

import (
	"context"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/nekika/cpus/internal"
	"github.com/nekika/cpus/lib"
	"github.com/shirou/gopsutil/v3/cpu"
	"log"
	"net/http"
	"nhooyr.io/websocket"
)

func main() {
	var (
		b  = lib.NewBroadCaster[[]float64]()
		ch = make(chan []float64)
	)

	go internal.CollectUsages(ch)
	go b.Broadcast(ch)

	app := echo.New()

	app.Static("/", "webapp")

	app.GET("/api/cpus", func(c echo.Context) error {
		times, err := cpu.Percent(0, true)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.JSON(http.StatusOK, times)
	})

	app.GET("/ws/cpus", func(c echo.Context) error {
		conn, err := websocket.Accept(c.Response(), c.Request(), nil)
		if err != nil {
			log.Println("Failed to accept WS connection:", err.Error())
		}

		ch := make(chan []float64)
		if _, err := b.Register(ch); err != nil {
			log.Println("Failed to register new subscriber:", err.Error())
			return c.NoContent(http.StatusInternalServerError)
		}

		for {
			msg, err := json.Marshal(<-ch)
			if err != nil {
				log.Println("Failed to marshal value to JSON:", err.Error())
				continue
			}

			if err := conn.Write(context.Background(), websocket.MessageText, msg); err != nil {
				log.Println("Failed to write message to connection:", err.Error())
				return nil
			}
		}
	})

	log.Fatal(app.Start(":4356"))
}
