package internal

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"log"
	"time"
)

func CollectUsages(ch chan<- []float64) {
	for {
		usages, err := cpu.Percent(time.Second*1, true)
		if err != nil {
			log.Println("error retrieving usages:", err.Error())

			continue
		}

		ch <- usages
	}
}
