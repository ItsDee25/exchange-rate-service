package controller

import (
	"log"
	"net/http"
	"strconv"

	"github.com/ItsDee25/exchange-rate-service/internal/domain"
	"github.com/gin-gonic/gin"
)

type currencyController struct {
	currencyUsecase domain.ICurrencyUsecase
}

func NewCurrencyController(u domain.ICurrencyUsecase) *currencyController {
	return &currencyController{
		currencyUsecase: u,
	}
}

func (controller *currencyController) ConvertCurrencyHandler(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	amountStr := c.Query("amount")
	date := c.Query("date")

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		log.Println("Error parsing amount:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}
	if !isValidCurrency(from) || !isValidCurrency(to) || amount <= 0 {
		log.Printf("Invalid parameters: from: %s, to: %s, amount %s", from, to, amountStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters"})
		return
	}
	if date != "" && !isWithin90Days(date) {
		log.Printf("Invalid date: %s", date)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date must be within the last 90 days"})
		return
	}

	convertedAmount, err := controller.currencyUsecase.GetConvertedCurrency(from, to, date, amount)
	if err != nil {
		log.Println("Error converting currency:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert currency"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from":             from,
		"to":               to,
		"amount":           amount,
		"date":             date,
		"converted_amount": convertedAmount,
	})
}

func (controller *currencyController) GetExchangeRateHandler(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	date := c.Query("date")

	if !isValidCurrency(from) || !isValidCurrency(to) {
		log.Printf("Invalid parameters: from: %s, to: %s", from, to)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters"})
		return
	}
	if date != "" && !isWithin90Days(date) {
		log.Printf("Invalid date: %s", date)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date must be within the last 90 days"})
		return
	}

	rate, err := controller.currencyUsecase.GetExchangeRate(from, to, date)
	if err != nil {
		log.Println("Error getting exchange rate:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get exchange rate"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"from": from,
		"to":   to,
		"rate": rate,
	})
}
