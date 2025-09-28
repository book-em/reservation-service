package roomclient

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
	util.TEL.Eventf("find room %d", nil, id)

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
		util.TEL.Eventf("room %d not found: http %d", nil, id, resp.StatusCode)
		return nil, fmt.Errorf("room %d not found", id)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		util.TEL.Eventf("could not parse bytes from response", err)
		return nil, err
	}
	defer resp.Body.Close()

	var obj RoomDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		util.TEL.Eventf("could not unmarshall JSON", err)
		return nil, err
	}

	return &obj, nil
}

func (c *roomClient) FindCurrentAvailabilityListOfRoom(context context.Context, roomId uint) (*RoomAvailabilityListDTO, error) {
	util.TEL.Eventf("find current availability list of room %d", nil, roomId)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/available/room/%d", c.baseURL, roomId), nil)
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
		util.TEL.Eventf("current availability list of room %d not found: http %d", nil, roomId, resp.StatusCode)
		return nil, fmt.Errorf("current availability list of room %d not found", roomId)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		util.TEL.Eventf("could not parse bytes from response", err)
		return nil, err
	}
	defer resp.Body.Close()

	var obj RoomAvailabilityListDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		util.TEL.Eventf("could not unmarshall JSON", err)
		return nil, err
	}

	return &obj, nil
}

func (c *roomClient) FindCurrentPricelistOfRoom(context context.Context, roomId uint) (*RoomPriceListDTO, error) {
	util.TEL.Eventf("find current price list of room %d", nil, roomId)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/price/room/%d", c.baseURL, roomId), nil)
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
		util.TEL.Eventf("current price list of room %d not found: http %d", nil, roomId, resp.StatusCode)
		return nil, fmt.Errorf("current price list of room %d not found", roomId)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		util.TEL.Eventf("could not parse bytes from response", err)
		return nil, err
	}
	defer resp.Body.Close()

	var obj RoomPriceListDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		util.TEL.Eventf("could not unmarshall JSON", err)
		return nil, err
	}

	return &obj, nil
}

func (c *roomClient) QueryForReservation(context context.Context, jwt string, dto RoomReservationQueryDTO) (*RoomReservationQueryResponseDTO, error) {
	util.TEL.Eventf("query room %d for potential reservation", nil, dto.RoomID)

	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		util.TEL.Eventf("could not unmarshall input JSON", err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/reservation/query", c.baseURL), bytes.NewBuffer(jsonBytes))
	if err != nil {
		util.TEL.Eventf("could not create request", err)
		return nil, err
	}
	otel.GetTextMapPropagator().Inject(context, propagation.HeaderCarrier(req.Header))
	req.Header.Add("Authorization", "Bearer "+jwt)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		util.TEL.Eventf("could not send request", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		util.TEL.Eventf("could not query room %d for potential reservation: http %d", nil, dto.RoomID, resp.StatusCode)
		return nil, fmt.Errorf("failed querying room %d", dto.RoomID)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		util.TEL.Eventf("could not parse bytes from response", err)
		return nil, err
	}
	defer resp.Body.Close()

	var obj RoomReservationQueryResponseDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		util.TEL.Eventf("could not unmarshall JSON", err)
		return nil, err
	}

	return &obj, nil
}
