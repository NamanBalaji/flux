package handler

import (
	"encoding/json"
	"github.com/NamanBalaji/flux/internal/broker/service"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
)

func SubscribeMessageHandler(broker *service.Broker) gin.HandlerFunc {
	return func(c *gin.Context) {
		jsonData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("invalid request body [ERROR]: %s", err)
			c.JSON(http.StatusBadRequest, err)

			return
		}

		var body ReadMessageRequest
		err = json.Unmarshal(jsonData, &body)
		if err != nil {
			log.Printf("invalid body format [ERROR]: %s", err)
			c.JSON(http.StatusBadRequest, err)

			return
		}

		messages, err := broker.ReadMessage(body.Topic, body.Partition, body.Offset)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})

			return
		}

		c.JSON(http.StatusOK, messages)
	}
}
