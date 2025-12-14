package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func SendDiscordStartup() {
	if DiscordWebhookURL == "" {
		return
	}

	title := "🚀 Bot Started / Restarted"
	description := "บอทเริ่มทำงานแล้วในโหมด **PRODUCTION** (เงินจริง)"
	color := 0x00ff00

	if IsDryRun {
		title = "🧪 Bot Started (DRY RUN)"
		description = "บอทเริ่มทำงานในโหมด **DRY RUN** (จำลองการเทรด)"
		color = 0xffa500
	}

	payload := map[string]interface{}{
		"username": "Bitkub Bot",
		"embeds": []map[string]interface{}{
			{
				"title":       title,
				"description": description,
				"color":       color,
				"fields": []map[string]interface{}{
					{"name": "Initial Investment", "value": fmt.Sprintf("%.2f THB", InitialInvestment), "inline": true},
					{"name": "Rebalance Threshold", "value": fmt.Sprintf("%.2f%%", Threshold), "inline": true},
					{"name": "Time", "value": time.Now().Format("15:04:05 02/01/2006"), "inline": false},
				},
				"footer": map[string]interface{}{
					"text": "Bitkub Rebalance Bot (GoLang)",
				},
			},
		},
	}

	sendToDiscord(payload)
	fmt.Println("🔔 Startup notification sent to Discord.")
}

func SendDiscordTrade(asset, operation string, amountTHB, coinAmount, price float64, mode string) {
	if DiscordWebhookURL == "" {
		return
	}

	color := 0x00ff00
	if operation == "sell" {
		color = 0xff0000
	}
	
	title := "✅ Trade Executed"
	if mode == "DRY_RUN" {
		title = "🔥 Dry Run Trade"
		color = 0xffcc00
	}

	payload := map[string]interface{}{
		"username": "Bitkub Bot",
		"embeds": []map[string]interface{}{
			{
				"title":       title,
				"description": fmt.Sprintf("Action: **%s** on **%s_THB**", operation, asset),
				"color":       color,
				"fields": []map[string]interface{}{
					{"name": "Price", "value": fmt.Sprintf("%.2f", price), "inline": true},
					{"name": "Amount (THB)", "value": fmt.Sprintf("%.2f", amountTHB), "inline": true},
					{"name": "Amount (Coin)", "value": fmt.Sprintf("%.8f", coinAmount), "inline": true},
				},
				"timestamp": time.Now().Format(time.RFC3339),
			},
		},
	}

	sendToDiscord(payload)
}

func sendToDiscord(payload map[string]interface{}) {
	jsonPayload, _ := json.Marshal(payload)
	
	go func() {
		resp, err := http.Post(DiscordWebhookURL, "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			fmt.Println("❌ Failed to send Discord webhook:", err)
			return
		}
		defer resp.Body.Close()
	}()
}

func SendDiscordModeChange(isDryRun bool) {
	if DiscordWebhookURL == "" {
		return
	}

	title := "🔄 Bot Mode Changed"
	description := "เปลี่ยนโหมดเป็น **PRODUCTION** (เริ่มใช้งานเงินจริง) 💸"
	color := 0x00ff00

	if isDryRun {
		description = "เปลี่ยนโหมดเป็น **DRY RUN** (จำลองการเทรด) 🧪"
		color = 0xffa500
	}

	payload := map[string]interface{}{
		"username": "Bitkub Bot",
		"embeds": []map[string]interface{}{
			{
				"title":       title,
				"description": description,
				"color":       color,
				"fields": []map[string]interface{}{
					{"name": "Time", "value": time.Now().Format("15:04:05 02/01/2006"), "inline": false},
				},
				"footer": map[string]interface{}{
					"text": "Bitkub Rebalance Bot",
				},
			},
		},
	}

	sendToDiscord(payload)
}