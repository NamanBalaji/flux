package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/NamanBalaji/flux/pkg/request"
)

func PollMessage(messageChan chan request.PollMessage) gin.HandlerFunc {
	return func(c *gin.Context) {
		jsonData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("invalid request body [ERROR]: %s", err)
			c.JSON(http.StatusBadRequest, err)

			return
		}

		var body request.PollMessage
		err = json.Unmarshal(jsonData, &body)
		if err != nil {
			log.Printf("invalid body format [ERROR]: %s", err)
			c.JSON(http.StatusBadRequest, err)

			return
		}

		messageChan <- body

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("message with id %s consumed successfully", body.Topic),
		})
	}
}
