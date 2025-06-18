package mocks

import (
	"context"
	"fmt"
)

// MockRateFetcher implements domain.RateFetcher
type MockRateFetcher struct {
	Rates map[string]float64 // key: "FROM#TO"
}

func NewMockRateFetcher() *MockRateFetcher {
	return &MockRateFetcher{
		Rates: map[string]float64{
			"USD#INR": 83.12,
			"USD#EUR": 0.93,
			"USD#JPY": 155.42,
			"INR#USD": 0.012,
			"INR#EUR": 0.011,
			"INR#JPY": 1.87,
			"EUR#USD": 1.07,
			"EUR#INR": 89.31,
			"EUR#JPY": 166.94,
			"JPY#USD": 0.0064,
			"JPY#INR": 0.53,
			"JPY#EUR": 0.0060,
			"AUD#USD": 0.66,
			"USD#AUD": 1.51,
			"GBP#USD": 1.27,
			"USD#GBP": 0.79,
			"CAD#INR": 61.23,
			"INR#CAD": 0.016,
			"CHF#USD": 1.12,
			"USD#CHF": 0.89,
		},
	}
}

func (m *MockRateFetcher) FetchRate(ctx context.Context, from, to, date string) (float64, error) {
	key := fmt.Sprintf("%s#%s", from, to)
	if rate, ok := m.Rates[key]; ok {
		return rate, nil
	}
	return 0, fmt.Errorf("mock rate not found for %s to %s", from, to)
}
