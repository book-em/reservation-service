package roomclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type RoomClient interface {
	FindById(it uint) (*RoomDTO, error)

	FindCurrentAvailabilityListOfRoom(roomId uint) (*RoomAvailabilityListDTO, error)
	FindCurrentPricelistOfRoom(roomId uint) (*RoomPriceListDTO, error)
}

type roomClient struct {
	baseURL string
}

func NewRoomClient() RoomClient {
	return &roomClient{
		baseURL: "http://room-service:8080/api", // TODO: This should not be hardcoded
	}
}

func (c *roomClient) FindById(id uint) (*RoomDTO, error) {
	log.Printf("Find room %d", id)

	resp, err := http.Get(fmt.Sprintf("%s/%d", c.baseURL, id))

	if err != nil {
		log.Printf("Error %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Room %d not found: http %d", id, resp.StatusCode)
		return nil, fmt.Errorf("user %d not found", id)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Parsing response error: %v", err)
		return nil, err
	}

	var obj RoomDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		log.Printf("JSON Unmarshall error: %v", err)
		return nil, err
	}

	return &obj, nil
}

func (c *roomClient) FindCurrentAvailabilityListOfRoom(roomId uint) (*RoomAvailabilityListDTO, error) {
	log.Printf("Find current availability list of room %d", roomId)

	resp, err := http.Get(fmt.Sprintf("%s/available/room/%d", c.baseURL, roomId))

	if err != nil {
		log.Printf("Error %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Room %d availability list not found: http %d", roomId, resp.StatusCode)
		return nil, fmt.Errorf("user %d not found", roomId)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Parsing response error: %v", err)
		return nil, err
	}

	var obj RoomAvailabilityListDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		log.Printf("JSON Unmarshall error: %v", err)
		return nil, err
	}

	return &obj, nil
}

func (c *roomClient) FindCurrentPricelistOfRoom(roomId uint) (*RoomPriceListDTO, error) {
	log.Printf("Find current price list of room %d", roomId)

	resp, err := http.Get(fmt.Sprintf("%s/available/room/%d", c.baseURL, roomId))

	if err != nil {
		log.Printf("Error %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Room %d price list not found: http %d", roomId, resp.StatusCode)
		return nil, fmt.Errorf("user %d not found", roomId)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Parsing response error: %v", err)
		return nil, err
	}

	var obj RoomPriceListDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		log.Printf("JSON Unmarshall error: %v", err)
		return nil, err
	}

	return &obj, nil
}
