package handler

import (
	"encoding/json"
	"fmt"
	"github.com/NamanBalaji/flux/internal/broker/service"
	"github.com/NamanBalaji/flux/pkg/subscriber"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
)

func RegisterSubscriberHandler(broker *service.Broker) gin.HandlerFunc {
	return func(c *gin.Context) {
		jsonData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("invalid request body [ERROR]: %s", err)
			c.JSON(http.StatusBadRequest, err)

			return
		}

		var body RegisterSubscriberRequest
		err = json.Unmarshal(jsonData, &body)
		if err != nil {
			log.Printf("invalid body format [ERROR]: %s", err)
			c.JSON(http.StatusBadRequest, err)

			return
		}

		for _, topic := range body.Topics {
			t := broker.FindTopic(topic)
			if t == nil {
				log.Printf("invalid topic [%s]", topic)
				c.JSON(http.StatusBadRequest, fmt.Errorf("invalid topic [%s]", topic))
			}
		}

		latestOrder := broker.Messages.GetLatestOrder()
		// create new subscriber for each topic
		for _, topic := range body.Topics {
			sub := subscriber.NewSubscriber(body.Address, latestOrder)
			_, err := broker.AddSubscriber(topic, *sub)
			if err != nil {
				log.Printf("error occured while regestering the subscriber to the topic [ERROR]: %s", err)
				c.JSON(http.StatusInternalServerError, err)

				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "subscriber registered successfully",
			"topics":  body.Topics,
		})
	}
}
