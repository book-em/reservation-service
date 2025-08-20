package userclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type UserClient interface {
	FindById(it uint) (*UserDTO, error)
}

type userClient struct {
	baseURL string
}

func NewUserClient() UserClient {
	return &userClient{
		baseURL: "http://user-service:8080/api", // TODO: This should not be hardcoded
	}
}

func (c *userClient) FindById(id uint) (*UserDTO, error) {
	log.Printf("Find user %d", id)

	resp, err := http.Get(fmt.Sprintf("%s/%d", c.baseURL, id))

	if err != nil {
		log.Printf("Error %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("User %d not found: http %d", id, resp.StatusCode)
		return nil, fmt.Errorf("user %d not found", id)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Parsing response error: %v", err)
		return nil, err
	}

	var obj UserDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		log.Printf("JSON Unmarshall error: %v", err)
		return nil, err
	}

	return &obj, nil
}
