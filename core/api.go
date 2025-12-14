package core

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func API_URL() string {
	return os.Getenv("BITKUB_API_BASE_URL")
}

func signPayload(apiSecret string, timestamp string, method string, endpoint string, body []byte) string {
	sigBody := ""
	if len(body) > 0 {
		sigBody = string(body)
	}

	payload := fmt.Sprintf("%s%s%s%s", timestamp, strings.ToUpper(method), endpoint, sigBody)

	h := hmac.New(sha256.New, []byte(apiSecret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func sendPrivateRequest(endpoint string, method string, payload map[string]interface{}) ([]byte, error) {
	if APIKey == os.Getenv("BITKUB_API_KEY") || APISecret == os.Getenv("BITKUB_API_SECRET") {
		return nil, fmt.Errorf("API Keys not configured. Please check config.go")
	}

	payloadBytes := []byte{}
	if len(payload) > 0 {
		payloadBytes, _ = json.Marshal(payload)
	}
	timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	signature := signPayload(APISecret, timestamp, method, "/api/"+endpoint, payloadBytes)
	req, _ := http.NewRequest(method, API_URL()+"/"+endpoint, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BTK-TIMESTAMP", timestamp)
	req.Header.Set("X-BTK-SIGN", signature)
	req.Header.Set("X-BTK-APIKEY", APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var errorCheck map[string]interface{}
	if err := json.Unmarshal(body, &errorCheck); err == nil {
		if code, ok := errorCheck["error"].(float64); ok && code != 0 {
			return nil, fmt.Errorf("bitkub API error code %d: %s", int(code), string(body))
		}
	}

	return body, nil
}

func FetchTickerPrice(sym string) (float64, error) {
	resp, err := http.Get(API_URL() + "/market/ticker?sym=" + sym)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("failed to decode ticker JSON: %v", err)
	}

	if data, ok := result[sym]; ok {
		if lastPriceStr, ok := data["last"].(string); ok {
			if lastPrice, err := strconv.ParseFloat(lastPriceStr, 64); err == nil {
				return lastPrice, nil
			}
		}
		if lastPriceFloat, ok := data["last"].(float64); ok {
			return lastPriceFloat, nil
		}
	}

	return 0, fmt.Errorf("price not found or invalid format for %s", sym)
}

func FetchWalletBalance() (map[string]float64, error) {	
	respBody, err := sendPrivateRequest("v3/market/wallet", "POST", map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(respBody, &walletResp); err != nil {
		return nil, fmt.Errorf("failed to decode wallet JSON: %v", err)
	}

	balances := make(map[string]float64)
	balances["THB"] = 0.0
	balances[CoinAsset] = 0.0

	for asset, balanceValue := range walletResp.Result {
		if asset == "THB" || asset == CoinAsset {
			balances[asset] = balanceValue
		}
	}

	return balances, nil
}

func SendOrder(sym string, amount float64, op string) error {
	if amount <= 0 {
		return fmt.Errorf("cannot send order with non-positive amount: %.8f", amount)
	}

	if !strings.HasSuffix(sym, "_THB") {
		return fmt.Errorf("invalid trading symbol format: must end with _THB")
	}

	switch op {
	case "buy":
		return sendOrderRequest("v3/market/place-bid", sym, amount, op)
	case "sell":
		return sendOrderRequest("v3/market/place-ask", sym, amount, op)
	}

	return fmt.Errorf("invalid operation: must be 'buy' or 'sell'")
}

func sendOrderRequest(endpoint string, sym string, amount float64, op string) error {
	rate := 0.0
	precision := 8
	if op == "buy" {
		precision = 2
	}
	amountStr := fmt.Sprintf(fmt.Sprintf("%%.%df", precision), amount)
	amountStr = strings.TrimRight(amountStr, "0")
	amountStr = strings.TrimRight(amountStr, ".")
	finalAmount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return fmt.Errorf("failed to parse final amount string to float (%s): %w", amountStr, err)
	}

	payload := map[string]interface{}{
		"sym": sym,
		"amt": finalAmount,
		"rat": rate,
		"typ": "market",
	}

	respBody, err := sendPrivateRequest(endpoint, "POST", payload)
	if err != nil {
		return err
	}

	var orderResp map[string]interface{}
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return fmt.Errorf("order sent to %s, but failed to decode response: %s", endpoint, string(respBody))
	}

	if code, ok := orderResp["error"].(float64); ok && code == 0 {
		return nil
	}

	return fmt.Errorf("order to %s failed. Response: %s", endpoint, string(respBody))
}
