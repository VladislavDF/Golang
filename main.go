package main

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Счетчик ошибок
	errorCount := 0

	// Бесконечный цикл для периодического опроса
	for {
		// Делаем запрос к серверу
		resp, err := http.Get("http://srv.msk01.gigacorp.local/_stats")

		// Проверяем ошибку при запросе
		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
			time.Sleep(5 * time.Second)
			continue
		}

		// Проверяем HTTP статус ответа
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
			time.Sleep(5 * time.Second)
			continue
		}

		// Читаем тело ответа
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
			time.Sleep(5 * time.Second)
			continue
		}

		// Разбиваем строку на части по запятой
		parts := strings.Split(string(body), ",")

		// Проверяем что получили 7 чисел
		if len(parts) != 7 {
			errorCount++
			if errorCount >= 3 {
				fmt.Println("Unable to fetch server statistic")
				errorCount = 0
			}
			time.Sleep(5 * time.Second)
			continue
		}

		// Преобразуем строки в числа
		loadAvg, _ := strconv.ParseFloat(parts[0], 64)
		totalMemory, _ := strconv.ParseInt(parts[1], 10, 64)
		usedMemory, _ := strconv.ParseInt(parts[2], 10, 64)
		totalDisk, _ := strconv.ParseInt(parts[3], 10, 64)
		usedDisk, _ := strconv.ParseInt(parts[4], 10, 64)
		totalBandwidth, _ := strconv.ParseInt(parts[5], 10, 64)
		usedBandwidth, _ := strconv.ParseInt(parts[6], 10, 64)

		// Сброс счетчика ошибок при успешном получении данных
		errorCount = 0

		// 1. Проверка Load Average - ИСПРАВЛЕНО: используем >= 30
		if loadAvg >= 30 {
			fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
		}

		// 2. Проверка памяти (80%)
		memoryPercent := float64(usedMemory) / float64(totalMemory) * 100
		if memoryPercent > 80 {
			// Округляем вниз
			roundedPercent := math.Floor(memoryPercent)
			fmt.Printf("Memory usage too high: %.0f%%\n", roundedPercent)
		}

		// 3. Проверка диска (90%)
		freeDisk := totalDisk - usedDisk
		diskPercent := float64(usedDisk) / float64(totalDisk) * 100
		if diskPercent > 90 {
			freeDiskMB := freeDisk / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeDiskMB)
		}

		// 4. Проверка сети (90%) - ИСПРАВЛЕНО: используем >= 90
		bandwidthPercent := float64(usedBandwidth) / float64(totalBandwidth) * 100
		if bandwidthPercent >= 90 {
			freeBandwidth := totalBandwidth - usedBandwidth
			freeBandwidthMbits := float64(freeBandwidth) / 1_000_000
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeBandwidthMbits)
		}

		// Ждем перед следующим запросом
		time.Sleep(5 * time.Second)
	}
}
