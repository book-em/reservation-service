package roomclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type RoomClient interface {
	FindById(context context.Context, it uint) (*RoomDTO, error)

	FindCurrentAvailabilityListOfRoom(context context.Context, roomId uint) (*RoomAvailabilityListDTO, error)
	FindCurrentPricelistOfRoom(context context.Context, roomId uint) (*RoomPriceListDTO, error)
	QueryForReservation(context context.Context, jwt string, dto RoomReservationQueryDTO) (*RoomReservationQueryResponseDTO, error)
}

type roomClient struct {
	baseURL string
}

func NewRoomClient() RoomClient {
	return &roomClient{
		baseURL: "http://room-service:8080/api", // TODO: This should not be hardcoded
	}
}

func (c *roomClient) FindById(context context.Context, id uint) (*RoomDTO, error) {
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

func (c *roomClient) FindCurrentAvailabilityListOfRoom(context context.Context, roomId uint) (*RoomAvailabilityListDTO, error) {
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

func (c *roomClient) FindCurrentPricelistOfRoom(context context.Context, roomId uint) (*RoomPriceListDTO, error) {
	log.Printf("Find current price list of room %d", roomId)

	resp, err := http.Get(fmt.Sprintf("%s/price/room/%d", c.baseURL, roomId))

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

func (c *roomClient) QueryForReservation(context context.Context, jwt string, dto RoomReservationQueryDTO) (*RoomReservationQueryResponseDTO, error) {
	log.Printf("Query room %d for potential reservation", dto.RoomID)

	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		log.Printf("JSON marshall error %v", err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/reservation/query", c.baseURL), bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Printf("Error %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Could not query room %d for potential reservation", dto.RoomID)
		return nil, fmt.Errorf("failed querying room %d", dto.RoomID)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Parsing response error: %v", err)
		return nil, err
	}

	var obj RoomReservationQueryResponseDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		log.Printf("JSON Unmarshall error: %v", err)
		return nil, err
	}

	return &obj, nil
}
