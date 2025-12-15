package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
	dotenv "github.com/joho/godotenv"
)

type Entry struct {
	PeakSpeed string `json:"peak_speed"`
	AvgSpeed  string `json:"avg_speed"`
	TimeTaken int64  `json:"time_taken"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}

type Benchmark struct {
	Version  string `json:"version"`
	Layer    int    `json:"layer"`
	Download Entry  `json:"download"`
	Upload   Entry  `json:"upload"`
	FileSize int64  `json:"file_size"`
}

func main() {
	var benchmark = Benchmark{
		Version: tg.Version,
		Layer:   tg.ApiVersion,
	}

	dotenv.Load()
	var (
		APP_ID       = os.Getenv("APP_ID")
		APP_HASH     = os.Getenv("API_HASH")
		BOT_TOKEN    = os.Getenv("BOT_TOKEN")
		MESSAGE_LINK = os.Getenv("MESSAGE_LINK")
		TG_SESSION   = os.Getenv("TG_SESSION")
	)

	appIdInt, _ := strconv.Atoi(APP_ID)

	cfg := tg.ClientConfig{
		AppID:         int32(appIdInt),
		AppHash:       APP_HASH,
		LogLevel:      tg.LogInfo,
		MemorySession: true,
		DisableCache:  true,
	}

	if TG_SESSION != "" {
		cfg.StringSession = TG_SESSION
	}

	client, _ := tg.NewClient(cfg)
	client.LoginBot(BOT_TOKEN)

	parts := strings.Split(MESSAGE_LINK, "/")
	chat := parts[3]
	messageId, _ := strconv.Atoi(parts[4])

	message, _ := client.GetMessageByID(chat, int32(messageId))

	fileSize := message.File.Size

	// ----- Download logic -----
	var peakSpeed int64
	startTime := time.Now().Unix()

	var pmDownload *tg.ProgressManager

	downloaded, _ := message.Download(&tg.DownloadOptions{
		Progress: func(current, total int64) {
			if pmDownload == nil {
				pmDownload = tg.NewProgressManager(int(total), 3)
			}
			if pmDownload.ShouldEdit(int(current)) {
				stats := pmDownload.GetStats(int(current))
				fmt.Println(stats) // or send as client.EditMessage
			}

			if time.Now().Unix()-startTime > 0 {
				speed := current / (time.Now().Unix() - startTime)
				if speed > peakSpeed {
					peakSpeed = speed
				}
			}
		},
	})
	defer os.Remove(downloaded)

	avgSpeed := float64(fileSize) / float64(time.Now().Unix()-startTime)

	benchmark.Download = Entry{
		PeakSpeed: HumanizeBytes(peakSpeed) + "/s",
		AvgSpeed:  HumanizeBytes(int64(avgSpeed)) + "/s",
		TimeTaken: time.Now().Unix() - startTime,
		StartTime: startTime,
		EndTime:   time.Now().Unix(),
	}

	// ----- Upload logic -----
	startTime = time.Now().Unix()
	peakSpeed = 0

	var pmUpload *tg.ProgressManager

	client.SendMedia(message.Chat, downloaded, &tg.MediaOptions{
		Progress: func(current, total int64) {
			if pmUpload == nil {
				pmUpload = tg.NewProgressManager(int(total), 3)
			}
			if pmUpload.ShouldEdit(int(current)) {
				stats := pmUpload.GetStats(int(current))
				fmt.Println(stats) // or client.EditMessage
			}
			if time.Now().Unix()-startTime > 0 {
				speed := current / (time.Now().Unix() - startTime)
				if speed > peakSpeed {
					peakSpeed = speed
				}
			}
		},
		ForceDocument: true,
		ReplyID:       message.ID,
		Caption:       "gogram",
		Attributes:    message.Document().Attributes,
	})

	avgSpeed = float64(fileSize) / float64(time.Now().Unix()-startTime)

	benchmark.Upload = Entry{
		PeakSpeed: HumanizeBytes(peakSpeed) + "/s",
		AvgSpeed:  HumanizeBytes(int64(avgSpeed)) + "/s",
		TimeTaken: time.Now().Unix() - startTime,
		StartTime: startTime,
		EndTime:   time.Now().Unix(),
	}

	benchmark.FileSize = fileSize

	jsonBenchmark, _ := json.MarshalIndent(benchmark, "", "  ")
	os.WriteFile("../../out/gogram.json", jsonBenchmark, 0644)
}

func HumanizeBytes(size int64) string {
	var units = []string{"B", "KB", "MB", "GB", "TB"}
	var i = 0
	for size > 1024 {
		size = size / 1024
		i++
	}
	return fmt.Sprintf("%.2f %s", float64(size), units[i])
}
