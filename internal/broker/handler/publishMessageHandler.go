package handler

import (
	"encoding/json"
	"github.com/NamanBalaji/flux/internal/broker/service"
	"github.com/NamanBalaji/flux/pkg/model"
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
		msg, err := broker.AddMessage(topic.Name, model.Message{
			Id:      body.Id,
			Topic:   body.Topic,
			Payload: body.Message,
		})

		if err != nil {
			log.Printf("error occured while writing the message to the topic [ERROR]: %s", err)
			c.JSON(http.StatusBadRequest, err)

			return
		}

		c.JSON(http.StatusOK, gin.H{
			"topic":   topic.Name,
			"message": *msg,
		})
	}
}
