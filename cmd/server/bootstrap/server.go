package bootstrap

import (
	"net/http"

	"github.com/ItsDee25/exchange-rate-service/internal/router"
	"github.com/gin-gonic/gin"
)

func InitServer() {

	r := gin.Default()

	// health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.RegisterRoutes(r)

	if err := r.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
	

}
