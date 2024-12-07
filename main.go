package main

import (
	"akile_monitor/client/model"
	"encoding/json"
	"os"
	"os/signal"

	"fmt"
	"log"
	"time"
)

var offline = make(map[string]bool)

func main() {
	LoadConfig()
	initDb()
	initFileDb()
	if cfg.EnableTG {
		go startbot()
	}

	if cfg.TgChatID != 0 {
		go func() {
			for {
				var mm []model.Data
				data := fetchData()
				json.Unmarshal(data, &mm)
				for _, v := range mm {
					// 30秒内离线
					if v.Timestamp < time.Now().Unix()-60 {
						if !offline[v.Host.Name] {
							offline[v.Host.Name] = true
							msg := fmt.Sprintf("❌ %s 离线了", v.Host.Name)
							SendTGMessage(msg)
						}
					} else {
						if offline[v.Host.Name] {
							offline[v.Host.Name] = false
							msg := fmt.Sprintf("✅ %s 上线了", v.Host.Name)
							SendTGMessage(msg)
						}
					}
				}
				time.Sleep(time.Second * 20)

			}
		}()
	}

	server := newServer()
	if err := server.Start(); err != nil {
		log.Printf("http server stopped: %s", err.Error())
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
}
