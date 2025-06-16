package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Ensure it implements usecase.RateFetcher
type ExchangeRateAPI struct {
	baseURL    string
	httpClient *http.Client
}

func NewExchangeRateAPI() *ExchangeRateAPI {
	return &ExchangeRateAPI{
		baseURL:    "https://api.exchangerate.host",
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

type apiResponse struct {
	Success bool    `json:"success"`
	Rate    float64 `json:"rate"`
}

// FetchRate implements RateFetcher
func (e *ExchangeRateAPI) FetchRate(ctx context.Context, from, to, date string) (float64, error) {
	url := fmt.Sprintf("%s/%s?from=%s&to=%s", e.baseURL, date, from, to)

	resp, err := e.httpClient.Get(url)
	if err != nil {
		return 0, fmt.Errorf("error calling rate API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("rate API returned non-200: %d", resp.StatusCode)
	}

	var parsed apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, fmt.Errorf("decode error: %w", err)
	}

	return parsed.Rate, nil
}
