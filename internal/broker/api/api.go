package api

import (
	"github.com/NamanBalaji/flux/internal/broker/handler"
	"github.com/NamanBalaji/flux/internal/broker/service"
	"github.com/NamanBalaji/flux/pkg/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg config.Config, broker *service.Broker) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "alive",
		})
	})

	r.POST("/publish", handler.PublishMessageHandler(cfg, broker))
	r.POST("/subscribe", handler.RegisterSubscriberHandler(cfg, broker))
	r.POST("/unsubscribe", handler.UnsubscribeHandler(broker))

	return r
}
