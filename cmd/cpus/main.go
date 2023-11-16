package main

import (
	"context"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/shirou/gopsutil/v3/cpu"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"sync"
	"time"
)

func CollectCpuUsages(ch chan<- []float64) {
	for {
		usages, err := cpu.Percent(time.Second*1, true)
		if err != nil {
			log.Println("Error getting usages:", err.Error())

			continue
		}

		ch <- usages
	}
}

func BroadCastCpuUsages(ch <-chan []float64, clients *map[int]*Client) {
	for {
		usages, ok := <-ch
		if !ok {
			return
		}

		if len(*clients) == 0 {
			continue
		}

		for _, client := range *clients {
			client.Chan <- usages
		}
		log.Printf("sent cpu usages to %v clients\n", len(*clients))
	}
}

type Client struct {
	Id int
	*websocket.Conn
	Chan chan []float64
}

func NewClient(id int, conn *websocket.Conn) *Client {
	return &Client{
		Id:   id,
		Conn: conn,
		Chan: make(chan []float64),
	}
}

type SafeId struct {
	mx  sync.Mutex
	val int
}

func (si *SafeId) Increment() {
	si.mx.Lock()
	defer si.mx.Unlock()

	si.val += 1
}

func (si *SafeId) Value() int {
	return si.val
}

func main() {
	var (
		id       = new(SafeId)
		clients  = make(map[int]*Client)
		usagesch = make(chan []float64)
	)

	go CollectCpuUsages(usagesch)
	go BroadCastCpuUsages(usagesch, &clients)

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

		id.Increment()

		client := NewClient(id.Value(), conn)
		clients[client.Id] = client

		defer func() {
			conn.CloseNow()
			delete(clients, client.Id)
		}()

		for {
			ctx := conn.CloseRead(context.Background())

			if err := client.Conn.Ping(ctx); err != nil {
				log.Printf("Client %v closed the connection\n", client.Id)
				break
			}

			usages, ok := <-client.Chan
			if !ok {
				break
			}

			msg, err := json.Marshal(usages)
			if err != nil {
				log.Println("Failed to marshal usages to JSON")

				continue
			}

			if err := client.Conn.Write(context.Background(), websocket.MessageText, msg); err != nil {
				log.Println("Failed to write to connection:", err.Error())
			}
		}

		return nil
	})

	log.Fatal(app.Start(":4356"))
}
