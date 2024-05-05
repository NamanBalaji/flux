package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/NamanBalaji/flux/internal/broker/service"
	"github.com/NamanBalaji/flux/pkg/config"
	"github.com/NamanBalaji/flux/pkg/message"
)

func PublishMessageHandler(cfg config.Config, broker *service.Broker) gin.HandlerFunc {
	return func(c *gin.Context) {
		jsonData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("invalid request body [ERROR]: %s", err)
			c.JSON(http.StatusBadRequest, err)

			return
		}

		var body PublishMessageRequest
		err = json.Unmarshal(jsonData, &body)
		if err != nil {
			log.Printf("invalid body format [ERROR]: %s", err)
			c.JSON(http.StatusBadRequest, err)

			return
		}

		msg := message.NewMessage(body.Id, body.Message)

		processed := broker.PublishMessage(cfg, body.Topic, msg)

		if !processed {
			log.Print("error occurred while writing the message to the topic [ERROR]: server too busy")
			c.JSON(http.StatusInternalServerError, fmt.Errorf("error occurred while writing the message to the topic [ERROR]: server too busy"))

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("message published to topic %s", body.Topic),
		})
	}
}
