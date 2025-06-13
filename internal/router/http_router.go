package router

import (
	"github.com/ItsDee25/exchange-rate-service/internal/controller/currency"
	"github.com/ItsDee25/exchange-rate-service/internal/usecase/currency"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	registerCurrencyRoutes(r)
}

func registerCurrencyRoutes(r *gin.Engine) {
	group := r.Group("/currency")
	controller := controller.NewCurrencyController(usecase.NewCurrencyUsecase(nil))
	group.GET("/convert", controller.ConvertCurrencyHandler)
	group.GET("/exchangeRate", controller.GetExchangeRateHandler)
}
