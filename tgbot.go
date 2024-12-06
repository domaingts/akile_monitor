package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func startbot() {
	bot, err := tgbotapi.NewBotAPI(cfg.TgToken)
	if err != nil {
		log.Println("Error creating bot", err)
		return
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if update.Message.IsCommand() {
				if update.Message.Command() == "akall" {

					var ret []Data
					db.Model(&Data{}).Find(&ret)

					var mm []M
					for _, v := range ret {
						var m M
						json.Unmarshal([]byte(v.Data), &m)
						mm = append(mm, m)
					}

					var online int
					var cpu int
					var mem uint64
					var memused uint64
					var downspeed uint64
					var upspeed uint64
					var downflow uint64
					var upflow uint64
					var swapused uint64
					var swap uint64
					time_now := time.Now().Unix()
					for _, v := range mm {
						if v.TimeStamp > time_now-30 {
							online++
						}
						cpu += parseCPU(v.Host.CPU[0])
						mem += v.Host.MemTotal
						memused += v.State.MemUsed
						swap += v.Host.SwapTotal
						swapused += v.State.SwapUsed
						downspeed += v.State.NetInSpeed
						upspeed += v.State.NetOutSpeed
						downflow += v.State.NetInTransfer
						upflow += v.State.NetOutTransfer
					}

					//流量对等性
					var duideng string
					if downflow > upflow {
						duideng = fmt.Sprintf("%.2f%%", float64(upflow)/float64(downflow)*100)
					} else {
						duideng = fmt.Sprintf("%.2f%%", float64(downflow)/float64(upflow)*100)
					}

					msg := fmt.Sprintf(`统计信息
===========================
服务器数量： %d
在线服务器： %d
CPU核心数： %d
内存： %s [%s/%s]
交换分区： %s [%s/%s]
下行速度： ↓%s/s
上行速度： ↑%s/s
下行流量： ↓%s
上行流量： ↑%s
流量对等性： %s

更新于：%s`, len(mm), online, cpu, fmt.Sprintf("%.2f%%", float64(memused)/float64(mem)*100), formatSize(memused), formatSize(mem), fmt.Sprintf("%.2f%%", float64(swapused)/float64(swap)*100), formatSize(swapused), formatSize(swap), formatSize(downspeed), formatSize(upspeed), formatSize(downflow), formatSize(upflow), duideng, time.Now().Format("2006-01-02 15:04:05"))
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
				}
			} else if update.Message.Command() == "id" {
				msg := tgbotapi.NewMessage(update.Message.From.ID, "Your ID is "+strconv.FormatInt(update.Message.From.ID, 10))
				bot.Send(msg)
			}
		}
	}
}

// 解析CPU数量
func parseCPU(cpu string) int {
	re := regexp.MustCompile(`(\d+) (Virtual) Core`)

	// 查找匹配项
	matches := re.FindStringSubmatch(cpu)
	if len(matches) >= 2 {
		virtualCores := matches[1]

		vint, err := strconv.Atoi(virtualCores)
		if err != nil {
			return 0
		}
		return vint
	}
	return 0
}

// 格式化字节大小
func formatSize(size uint64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MB", float64(size)/1024/1024)
	} else if size < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2f GB", float64(size)/1024/1024/1024)
	} else if size < 1024*1024*1024*1024*1024 {
		return fmt.Sprintf("%.2f TB", float64(size)/1024/1024/1024/1024)
	} else if size < 1024*1024*1024*1024*1024*1024 {
		return fmt.Sprintf("%.2f PB", float64(size)/1024/1024/1024/1024/1024)
	} else {
		return fmt.Sprintf("%.2f EB", float64(size)/1024/1024/1024/1024/1024/1024)
	}
}

func SendTGMessage(msg string) {
	bot, err := tgbotapi.NewBotAPI(cfg.TgToken)
	if err != nil {
		log.Println("Error creating bot", err)
		return
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	msgs := tgbotapi.NewMessage(cfg.TgChatID, msg)
	bot.Send(msgs)
}
