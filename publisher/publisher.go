package publisher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/NamanBalaji/flux/pkg/request"
)

type Publisher struct {
	brokerAddress string
}

func NewPublisher(brokerAddress string) *Publisher {
	return &Publisher{brokerAddress: brokerAddress}
}

func (p *Publisher) Publish(topic string, message string) (*request.PublishMessageRequest, error) {
	id := uuid.New().String()
	requestBody := request.PublishMessageRequest{
		Id:      id,
		Message: message,
		Topic:   topic,
	}

	requestBodyJson, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	_, status, err := request.SendHTTPRequest(http.MethodPost, fmt.Sprintf("%s/publish", p.brokerAddress), bytes.NewBuffer(requestBodyJson))
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("publish request failed with status code %d", status)
	}

	return &requestBody, nil
}
