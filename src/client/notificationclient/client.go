package notificationclient

import (
	"bookem-reservation-service/util"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type NotificationClient interface {
	CreateNotification(ctx context.Context, jwt string, dto CreateNotificationDTO) (*NotificationDTO, error)
}

type notificationClient struct {
	baseURL string
}

func NewNotificationClient() NotificationClient {
	return &notificationClient{
		baseURL: "http://notification-service:8080/api",
	}
}

func (c *notificationClient) CreateNotification(ctx context.Context, jwt string, dto CreateNotificationDTO) (*NotificationDTO, error) {
	util.TEL.Info("creating notification", "receiver_id", dto.ReceiverID, "type", dto.Type)

	body, err := json.Marshal(dto)
	if err != nil {
		util.TEL.Error("failed to marshal notification DTO", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/notification", c.baseURL), bytes.NewBuffer(body))
	if err != nil {
		util.TEL.Error("could not create HTTP request", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwt)
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		util.TEL.Error("failed sending request to notification-service", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		util.TEL.Error("unexpected response from notification-service", nil, "status", resp.StatusCode, "body", string(bodyBytes))
		return nil, fmt.Errorf("failed to create notification, status: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		util.TEL.Error("failed reading response body", err)
		return nil, err
	}

	var notification NotificationDTO
	if err := json.Unmarshal(bodyBytes, &notification); err != nil {
		util.TEL.Error("failed unmarshalling response", err)
		return nil, err
	}

	return &notification, nil
}
