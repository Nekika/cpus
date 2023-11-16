package main

import (
	"github.com/labstack/echo/v4"
	"github.com/shirou/gopsutil/v3/cpu"
	"golang.org/x/net/websocket"
	"log"
	"net/http"
	"time"
)

func main() {
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
		websocket.Handler(func(conn *websocket.Conn) {
			defer conn.Close()

			var fails int
			for {
				if fails >= 3 {
					log.Println("Closing connection due to too many internal errors")
					return
				}

				usages, err := cpu.Percent(0, true)
				if err != nil {
					log.Println("Error getting usages:", err.Error())
					fails += 1

					continue
				}

				if err := websocket.JSON.Send(conn, usages); err != nil {
					log.Println("Error sending message:", err.Error())
					fails += 1
				}

				time.Sleep(time.Second * 1)
			}
		}).ServeHTTP(c.Response(), c.Request())

		return nil
	})

	log.Fatal(app.Start(":4356"))
}
