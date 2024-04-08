package handler

import (
	"encoding/json"
	"fmt"
	"github.com/NamanBalaji/flux/internal/broker/service"
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

		// add address validation ?

		// if topics absent return error
		err = broker.ValidateTopics(body.Topics)
		if err != nil {
			log.Printf("topics not valid: %s", err)

			c.JSON(http.StatusBadRequest, fmt.Errorf("topics not valid: %s", err))
		}

		// create new subscriber for each topic
		for _, topic := range body.Topics {
			broker.Subscribe(topic, body.Address)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "subscriber registered successfully",
			"topics":  body.Topics,
		})
	}
}
