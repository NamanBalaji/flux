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
)

func RegisterSubscriberHandler(cfg config.Config, broker *service.Broker) gin.HandlerFunc {
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

		// if topics absent return error
		err = broker.ValidateTopics(body.Topics)
		if err != nil {
			log.Printf("topics not valid: %s", err)

			c.JSON(http.StatusBadRequest, fmt.Errorf("topics not valid: %s", err))
		}

		// create new subscriber for each topic
		for _, topic := range body.Topics {
			broker.Subscribe(c, cfg, topic, body.Address, body.ReadOld)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "subscriber registered successfully",
			"topics":  body.Topics,
		})
	}
}
