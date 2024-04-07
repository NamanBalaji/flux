package api

import (
	"github.com/NamanBalaji/flux/internal/broker/handler"
	"github.com/NamanBalaji/flux/internal/broker/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(broker *service.Broker) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "alive",
		})
	})

	r.POST("/publish", handler.PublishMessageHandler(broker))

	// subscriber register
	return r
}
