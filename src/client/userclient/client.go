package userclient

import (
	"bookem-reservation-service/util"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type UserClient interface {
	FindById(context context.Context, it uint) (*UserDTO, error)
}

type userClient struct {
	baseURL string
}

func NewUserClient() UserClient {
	return &userClient{
		baseURL: "http://user-service:8080/api", // TODO: This should not be hardcoded
	}
}

func (c *userClient) FindById(context context.Context, id uint) (*UserDTO, error) {
	util.TEL.Eventf("find user %d", nil, id)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%d", c.baseURL, id), nil)
	if err != nil {
		util.TEL.Eventf("could not create request", err)
		return nil, err
	}
	otel.GetTextMapPropagator().Inject(context, propagation.HeaderCarrier(req.Header))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		util.TEL.Eventf("could not send request", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		util.TEL.Eventf("user %d not found: http %d", nil, id, resp.StatusCode)
		return nil, fmt.Errorf("user %d not found", id)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		util.TEL.Eventf("could not parse bytes from response", err)
		return nil, err
	}
	defer resp.Body.Close()

	var obj UserDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		util.TEL.Eventf("could not unmarshall JSON", err)
		return nil, err
	}

	return &obj, nil
}
