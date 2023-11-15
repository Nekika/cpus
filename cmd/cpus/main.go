package main

import (
	"github.com/labstack/echo/v4"
	"github.com/shirou/gopsutil/v3/cpu"
	"log"
	"net/http"
)

func main() {
	app := echo.New()

	app.GET("/api/cpus", func(c echo.Context) error {
		times, err := cpu.Percent(0, true)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.JSON(http.StatusOK, times)
	})

	log.Fatal(app.Start(":4356"))
}
