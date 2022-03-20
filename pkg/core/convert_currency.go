package core

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
)

// ExchangeRateResponse represents models of API response
type ExchangeRateResponse struct {
	Result             string  `json:"result"`
	Documentation      string  `json:"documentation"`
	TermsOfUse         string  `json:"terms_of_use"`
	TimeLastUpdateUnix int     `json:"time_last_update_unix"`
	TimeLastUpdateUtc  string  `json:"time_last_update_utc"`
	TimeNextUpdateUnix int     `json:"time_next_update_unix"`
	TimeNextUpdateUtc  string  `json:"time_next_update_utc"`
	BaseCode           string  `json:"base_code"`
	TargetCode         string  `json:"target_code"`
	ConversionRate     float64 `json:"conversion_rate"`
	ConversionResult   float64 `json:"conversion_result"`
}

type currency string

const (
	usd currency = "usd"
	rub currency = "rub"
	gel currency = "gel"
)

const apiEndpoint = "https://v6.exchangerate-api.com/v6/%s/pair/%s/%s/%f"

func parseCurrency(s string) currency {
	switch s {
	case "$", "USD", "usd", "долларов", "доллара":
		return usd
	case "₽", "RUB", "rub", "рублей", "рубля":
		return rub
	case "₾", "GEL", "gel", "лари", "лар":
		return gel
	default:
		return ""
	}
}

func (c Core) convertToUSD(cur currency, floatAmount float64) (int, error) {
	if cur == "" {
		return 0, fmt.Errorf("failed to parse currency")
	}

	url := fmt.Sprintf(apiEndpoint, c.apiKey, cur, "usd", floatAmount)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("api returned bad status: %v", resp.StatusCode)
	}
	defer resp.Body.Close()

	var exchangeRateResp ExchangeRateResponse
	if err = json.NewDecoder(resp.Body).Decode(&exchangeRateResp); err != nil {
		return 0, err
	}
	return int(math.Round(exchangeRateResp.ConversionResult) * 100), nil
}
