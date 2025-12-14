package core

import (
	"fmt"
	"math"
	"sort"
	"time"
)

var LastCoinPrice float64

func RoundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func fetchCurrentPrice(sym string) float64 {
	if sym == "THB" {
		return 1.0
	}

	price, err := FetchTickerPrice("THB_" + sym)
	if err != nil {
		fmt.Printf("Error fetching price for %s: %v\n", sym, err)
		return 0.0
	}
	return price
}

func fetchCurrentBalance() map[string]float64 {
	balances, err := FetchWalletBalance()
	if err != nil {
		fmt.Printf("Error fetching wallet balance: %v\n", err)
		return map[string]float64{"THB": 0.0, CoinAsset: 0.0}
	}
	return balances
}

func CalculatePortfolio() PortfolioSummary {
	balance := fetchCurrentBalance()
	coinPrice := fetchCurrentPrice(CoinAsset)
	LastCoinPrice = coinPrice

	coinValue := balance[CoinAsset] * coinPrice
	totalValue := balance["THB"] + coinValue

	if totalValue < 0.00000001 {
		totalValue = 1.0
	}

	roi := 0.0
	if InitialInvestment > 0 {
		roi = ((totalValue - float64(InitialInvestment)) / float64(InitialInvestment)) * 100.0
	}

	ConfigMutex.RLock()
	targetAssets := TargetAssets
	ConfigMutex.RUnlock()

	portfolio := []AssetData{}

	for asset, targetPercent := range targetAssets {
		var assetValue float64
		var price float64
		var rawBalance float64

		switch asset {
		case "THB":
			assetValue = balance["THB"]
			price = 1.0
			rawBalance = balance["THB"]
		case CoinAsset:
			assetValue = coinValue
			price = coinPrice
			rawBalance = balance[CoinAsset]
		default:
			assetValue = 0.0
			price = 0.0
			rawBalance = 0.0
		}

		actualPercent := (assetValue / totalValue) * 100

		portfolio = append(portfolio, AssetData{
			Asset:        asset,
			CurrentPrice: price,
			CoinBalance:  rawBalance,
			BalanceTHB:   assetValue,
			ActualPct:    actualPercent,
			TargetPct:    targetPercent,
		})
	}

	sort.Sort(ByTargetAndAsset(portfolio))
	return PortfolioSummary{
		TotalValue: totalValue,
		ROI:        roi,
		Portfolio:  portfolio,
	}
}

func RunRebalance() {
	summary := CalculatePortfolio()
	portfolio := summary.Portfolio
	totalValue := summary.TotalValue
	ConfigMutex.RLock()
	dryRun := IsDryRun
	threshold := Threshold
	ConfigMutex.RUnlock()

	fmt.Printf("\n--- Rebalance Check (%s) | Total Value: %.2f THB | ROI: %.2f%% ---\n",
		time.Now().Format("15:04:05"), totalValue, summary.ROI)

	for _, assetData := range portfolio {
		if assetData.Asset == "THB" {
			continue
		}
		deviation := math.Abs(assetData.ActualPct - assetData.TargetPct)
		if deviation > threshold {
			fmt.Printf("⚠️ %s: สัดส่วนจริง %.2f%% (เป้าหมาย %.2f%%) | เบี่ยงเบน %.2f%% > THRESHOLD\n",
				assetData.Asset, assetData.ActualPct, assetData.TargetPct, deviation)

			diff := assetData.ActualPct - assetData.TargetPct
			operation := ""
			amountToTrade := 0.0

			if diff > 0 {
				operation = "sell"
			} else {
				operation = "buy"
			}
			amountToTrade = (math.Abs(diff) / 100.0) * totalValue
			amountToTrade = RoundFloat(amountToTrade, 2)

			if assetData.CurrentPrice <= 0 {
				fmt.Printf("❌ ERROR: ราคา %s เป็นศูนย์. ไม่สามารถคำนวณปริมาณได้.\n", assetData.Asset)
				continue
			}

			coinAmount := amountToTrade / assetData.CurrentPrice
			coinAmount = RoundFloat(coinAmount, 8)
			var finalAmount float64

			if operation == "buy" {
				if amountToTrade < 10.0 {
					fmt.Printf("⏸️ SKIP: BUY มูลค่า %.2f THB น้อยกว่าขั้นต่ำ 10.00 THB\n", amountToTrade)
					continue
				}
				finalAmount = amountToTrade
			} else {
				finalAmount = coinAmount
			}
			tradeSym := assetData.Asset + "_THB"

			if dryRun {
				mode := "DRY_RUN"
				logMessage := fmt.Sprintf(
					"จำลองคำสั่ง %s %.8f %s มูลค่า %.2f THB บนคู่ %s",
					operation, coinAmount, assetData.Asset, amountToTrade, tradeSym)
				fmt.Println("🔥 " + mode + ": " + logMessage)

				SendDiscordTrade(assetData.Asset, operation, amountToTrade, coinAmount, assetData.CurrentPrice, "DRY_RUN")
				LogTrade(assetData.Asset, operation, amountToTrade, coinAmount, assetData.CurrentPrice, mode, deviation, logMessage)
			} else {
				mode := "PRODUCTION"
				fmt.Printf("✅ PRODUCTION: ส่งคำสั่ง %s %.8f %s (มูลค่า %.2f THB)\n", operation, coinAmount, assetData.Asset, amountToTrade)
				err := SendOrder(tradeSym, finalAmount, operation)
				logMessage := ""
				if err != nil {
					logMessage = fmt.Sprintf("คำสั่งล้มเหลว: %v", err)
					fmt.Printf("❌ ERROR: %s\n", logMessage)
				} else {
					logMessage = "คำสั่งสำเร็จ: Order sent to Bitkub"
					SendDiscordTrade(assetData.Asset, operation, amountToTrade, coinAmount, assetData.CurrentPrice, "PRODUCTION")
				}

				LogTrade(assetData.Asset, operation, amountToTrade, coinAmount, assetData.CurrentPrice, mode, deviation, logMessage)
			}
		} else {
			fmt.Printf("✅ %s: สัดส่วนปกติ (%.2f%%) | ไม่ต้อง Rebalance\n", assetData.Asset, assetData.ActualPct)
		}
	}
}

type ByTargetAndAsset []AssetData

func (p ByTargetAndAsset) Len() int      { return len(p) }
func (p ByTargetAndAsset) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p ByTargetAndAsset) Less(i, j int) bool {
	if p[i].TargetPct != p[j].TargetPct {
		return p[i].TargetPct > p[j].TargetPct
	}
	return p[i].Asset < p[j].Asset
}

func StartBotLoop() {
	for {
		RunRebalance()
		time.Sleep(1 * time.Minute)
	}
}
