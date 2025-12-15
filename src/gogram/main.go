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
    msgID, _ := strconv.Atoi(parts[4])

    message, _ := client.GetMessageByID(chat, int32(msgID))
    fileSize := message.File.Size

    // Download progress
    var peakDown int64
    start := time.Now().Unix()
    var pmDown *tg.ProgressManager

    downloaded, _ := message.Download(&tg.DownloadOptions{
        Progress: func(current, total int) {
            if pmDown == nil {
                pmDown = tg.NewProgressManager(total, 3)
            }
            if pmDown.ShouldEdit(current) {
                fmt.Println(pmDown.GetStats(current))
            }
            elapsed := time.Now().Unix() - start
            if elapsed > 0 {
                speed := int64(current) / elapsed
                if speed > peakDown {
                    peakDown = speed
                }
            }
        },
    })
    defer os.Remove(downloaded)

    avgDown := float64(fileSize) / float64(time.Now().Unix()-start)
    benchmark.Download = Entry{
        PeakSpeed: HumanizeBytes(peakDown) + "/s",
        AvgSpeed:  HumanizeBytes(int64(avgDown)) + "/s",
        TimeTaken: time.Now().Unix() - start,
        StartTime: start,
        EndTime:   time.Now().Unix(),
    }

    // Upload progress
    start = time.Now().Unix()
    var peakUp int64
    var pmUp *tg.ProgressManager

    client.SendMedia(message.Chat, downloaded, &tg.MediaOptions{
        Progress: func(current, total int) {
            if pmUp == nil {
                pmUp = tg.NewProgressManager(total, 3)
            }
            if pmUp.ShouldEdit(current) {
                fmt.Println(pmUp.GetStats(current))
            }
            elapsed := time.Now().Unix() - start
            if elapsed > 0 {
                speed := int64(current) / elapsed
                if speed > peakUp {
                    peakUp = speed
                }
            }
        },
        ForceDocument: true,
        ReplyID:       message.ID,
        Caption:       "gogram",
        Attributes:    message.Document().Attributes,
    })

    avgUp := float64(fileSize) / float64(time.Now().Unix()-start)
    benchmark.Upload = Entry{
        PeakSpeed: HumanizeBytes(peakUp) + "/s",
        AvgSpeed:  HumanizeBytes(int64(avgUp)) + "/s",
        TimeTaken: time.Now().Unix() - start,
        StartTime: start,
        EndTime:   time.Now().Unix(),
    }

    benchmark.FileSize = fileSize

    out, _ := json.MarshalIndent(benchmark, "", "  ")
    os.WriteFile("../../out/gogram.json", out, 0644)
}

func HumanizeBytes(size int64) string {
    units := []string{"B", "KB", "MB", "GB", "TB"}
    i := 0
    for size > 1024 {
        size /= 1024
        i++
    }
    return fmt.Sprintf("%.2f %s", float64(size), units[i])
}
