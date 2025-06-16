package router

import (
	"net/http"

	"github.com/ItsDee25/exchange-rate-service/cmd/server/bootstrap/builders"
	controller "github.com/ItsDee25/exchange-rate-service/internal/controller/currency"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, usecases *builders.Usecases) {

	// health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	registerCurrencyRoutes(r, usecases)
}

func registerCurrencyRoutes(r *gin.Engine, usecases *builders.Usecases) {
	group := r.Group("/currency")
	controller := controller.NewCurrencyController(usecases.CurrencyUsecase)
	group.GET("/convert", controller.ConvertCurrencyHandler)
	group.GET("/exchangeRate", controller.GetExchangeRateHandler)
}
