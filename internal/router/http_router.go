package router

import (
	"github.com/ItsDee25/exchange-rate-service/internal/controller"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	registerCurrencyRoutes(r)
}

func registerCurrencyRoutes(r *gin.Engine) {
	group := r.Group("/currency")
	controller := controller.NewCurrencyController(nil)
	group.GET("/convert", controller.ConvertCurrencyHandler)
	group.GET("/exchangeRate", controller.GetExchangeRateHandler)
}
