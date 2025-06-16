package controller

import (
	"time"

	constants "github.com/ItsDee25/exchange-rate-service/internal/constants/currency"
)

func isValidCurrency(code string) bool {
	_, exists := constants.SupportedCurrencies[code]
	return exists
}

func isWithin90Days(dateStr string) bool {
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}

	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	return parsedDate.After(ninetyDaysAgo) && parsedDate.Before(time.Now())
}
