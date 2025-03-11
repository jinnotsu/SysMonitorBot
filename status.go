package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

// 指定された間隔でシステムのステータスを更新
func UpdateSystemStatus(s *discordgo.Session, intervalSeconds int) {
    ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
    defer ticker.Stop()

	for {
		// CPU使用率を取得
		cpuPercent, err := cpu.Percent(0, false)
		if err != nil {
			log.Fatalf("Error: Failed to get CPU information: %v\n", err)
			continue
		}

		// メモリ使用率を取得
		memInfo, err := mem.VirtualMemory()
		if err != nil {
			log.Fatalf("Error: Failed to get memory information: %v\n", err)
			continue
		}

		// ステータスメッセージを作成
		statusMsg := fmt.Sprintf("CPU: %.1f%% | MEM: %.1f%%", cpuPercent[0], memInfo.UsedPercent)

		// Botのステータスを更新
		err = s.UpdateStatusComplex(discordgo.UpdateStatusData{
			Activities: []*discordgo.Activity{
				{
					Name: statusMsg,
					Type: discordgo.ActivityTypeGame,
				},
			},
			Status: "online",
		})
		if err != nil {
			log.Fatalf("Error: Failed to update status: %v\n", err)
		}

		<-ticker.C
	}
}
