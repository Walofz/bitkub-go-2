package core

import (
	"fmt"
	"os"
	"strconv"
	"sync"
)

var (
	APIKey            string
	APISecret         string
	IsDryRun          bool
	InitialInvestment float64
	Threshold         float64
	CoinAsset         string  = "ETH"
	DiscordWebhookURL string

	TargetAssets map[string]float64
)

var ConfigMutex sync.RWMutex

func init() {
	APIKey = os.Getenv("BITKUB_API_KEY")
	APISecret = os.Getenv("BITKUB_API_SECRET")
	DiscordWebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")

	if APIKey == "" || APISecret == "" {
		fmt.Println("❌ CRITICAL: BITKUB_API_KEY or SECRET is missing. Trading will fail.")
	}
	IsDryRun, _ = strconv.ParseBool(os.Getenv("IS_DRY_RUN"))

	if val, err := strconv.ParseFloat(os.Getenv("INITIAL_INVESTMENT"), 64); err == nil {
		InitialInvestment = val
	}

	if val, err := strconv.ParseFloat(os.Getenv("THRESHOLD_PERCENTAGE"), 64); err == nil {
		Threshold = val
	}

	TargetAssets = make(map[string]float64)
	TargetAssets["THB"] = 50.0
	TargetAssets[CoinAsset] = 50.0

	fmt.Printf("✅ Config loaded. Mode: %s, Initial Inv: %.2f THB\n",
		func() string {
			if IsDryRun {
				return "DRY_RUN"
			} else {
				return "PRODUCTION"
			}
		}(), InitialInvestment)
}
