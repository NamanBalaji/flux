package handler

import (
	"encoding/json"
	"github.com/NamanBalaji/flux/internal/broker/service"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
)

func PublishMessageHandler(broker *service.Broker) gin.HandlerFunc {
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

		topic := broker.GetOrCreateTopic(body.Topic)
		partition, err := topic.AssignPartition(body.Key)
		if err != nil {
			log.Printf("error occured while assigning partition [ERROR]: %s", err)
			c.JSON(http.StatusBadRequest, err)

			return
		}

		offset := partition.AppendMessageToPartition(body.Message, body.Key)

		c.JSON(http.StatusOK, gin.H{
			"key":     body.Key,
			"message": body.Message,
			"topic":   body.Topic,
			"offset":  offset,
		})
	}
}
