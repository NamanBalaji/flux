package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/NamanBalaji/flux/internal/subscriber/handler"
	"github.com/NamanBalaji/flux/pkg/request"
)

func SetupRouter(messageChan chan request.PollMessage) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "alive",
		})
	})

	r.POST("/poll", handler.PollMessage(messageChan))

	return r
}
