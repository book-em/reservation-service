package test

import (
	"bookem-reservation-service/client/roomclient"
	"bookem-reservation-service/client/userclient"
	"bookem-reservation-service/internal"
	"bookem-reservation-service/util"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const URL_user = "http://user-service:8080/api/"
const URL_room = "http://room-service:8080/api/"
const URL_reservation = "http://reservation-service:8080/api/"

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type AvailabilityResponse struct {
	Available bool `json:"available"`
}

func GenName(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RegisterUser(username_or_email string, password string, role util.UserRole) (*http.Response, error) {
	username := username_or_email
	email := username + "@gmail.com"

	if strings.HasSuffix(username_or_email, "@gmail.com") {
		username = strings.Split(username_or_email, "@")[0]
		email = username_or_email
	}

	dto := userclient.UserCreateDTO{
		Username: username,
		Password: password,
		Email:    email,
		Role:     string(role),
		Name:     GenName(6),
		Surname:  GenName(6),
		Address:  GenName(10),
	}

	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(URL_user+"register", "application/json", bytes.NewBuffer(jsonBytes))
	return resp, err
}

func LoginUser(username_or_email string, password string) (*http.Response, error) {
	dto := userclient.LoginDTO{
		UsernameOrEmail: username_or_email,
		Password:        password,
	}

	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(URL_user+"login", "application/json", bytes.NewBuffer(jsonBytes))
	return resp, err
}

func LoginUser2(username_or_email string, password string) string {
	resp, _ := LoginUser(username_or_email, password)

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read response body: %v", err))
	}

	var token userclient.JWTDTO
	if err := json.Unmarshal(bodyBytes, &token); err != nil {
		panic(fmt.Sprintf("failed to unmarshal jwt: %v", err))
	}

	return token.Jwt
}

func CreateRoom(jwt string, dto roomclient.CreateRoomDTO) (*http.Response, error) {
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, URL_room+"new", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func ResponseToRoom(resp *http.Response) roomclient.RoomDTO {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read response body: %v", err))
	}

	var obj roomclient.RoomDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		fmt.Print(string(bodyBytes))
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	return obj
}

func CreateRoomAvailability(jwt string, dto roomclient.CreateRoomAvailabilityListDTO) (*http.Response, error) {
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, URL_room+"available", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func CreateRoomPrice(jwt string, dto roomclient.CreateRoomPriceListDTO) (*http.Response, error) {
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, URL_room+"price", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func CreateReservation(jwt string, dto internal.ReservationDTO) (*http.Response, error) {
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, URL_reservation+"new", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func ResponseToReservation(resp *http.Response) internal.ReservationDTO {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read response body: %v", err))
	}

	var obj internal.ReservationDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		fmt.Print(string(bodyBytes))
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	return obj
}

func ResponseToReservations(resp *http.Response) []internal.ReservationDTO {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read response body: %v", err))
	}

	var obj []internal.ReservationDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	return obj
}

func CreateReservationRequest(jwt string, dto internal.CreateReservationRequestDTO) (*http.Response, error) {
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, URL_reservation+"req", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func GetPendingRequestsByGuest(jwt string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, URL_reservation+"req/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func GetPendingRequestsByRoom(jwt string, roomID uint) (*http.Response, error) {
	url := fmt.Sprintf(URL_reservation+"req/room/%d", roomID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func DeleteReservationRequest(jwt string, requestID uint) (*http.Response, error) {
	url := fmt.Sprintf(URL_reservation+"req/%d", requestID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func CheckRoomAvailability(jwt string, roomID uint, from, to string) (*http.Response, error) {
	url := fmt.Sprintf(URL_reservation+"room/%d/availability?from=%s&to=%s", roomID, from, to)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func ResponseToReservationRequest(resp *http.Response) internal.ReservationRequestDTO {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read response body: %v", err))
	}

	var obj internal.ReservationRequestDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	return obj
}

func ResponseToReservationRequests(resp *http.Response) []internal.ReservationRequestDTO {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read response body: %v", err))
	}

	var obj []internal.ReservationRequestDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	return obj
}

func GetActiveGuestReservations(jwt string) (*http.Response, error) {
	url := URL_reservation + "reservations/guest/active"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

const (
	SMALL_IMG = "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAAMCAgMCAgMDAwMEAwMEBQgFBQQEBQoHBwYIDAoMDAsKCwsNDhIQDQ4RDgsLEBYQERMUFRUVDA8XGBYUGBIUFRT/wAALCAABAAEBAREA/8QAFAABAAAAAAAAAAAAAAAAAAAACf/EABQQAQAAAAAAAAAAAAAAAAAAAAD/2gAIAQEAAD8AKp//2Q=="
)

func SetupHostRoomAvailabilityPrice(hostUsername string, t *testing.T) (string, string, string, roomclient.RoomDTO) {
	// Step 1: Register unique host
	username := hostUsername
	password := "pass"
	RegisterUser(username, password, util.Host)
	jwt := LoginUser2(username, password)
	jwtObj, err := util.GetJwtFromString(jwt)
	require.NoError(t, err)

	// Step 2: Create room
	roomDTO := roomclient.CreateRoomDTO{
		HostID:        jwtObj.ID,
		Name:          "Room_" + GenName(6),
		Description:   "Test room",
		Address:       "Test address",
		MinGuests:     1,
		MaxGuests:     4,
		PhotosPayload: []string{SMALL_IMG},
		Commodities:   []string{"WiFi", "AC"},
		AutoApprove:   false,
	}
	roomResp, err := CreateRoom(jwt, roomDTO)
	require.NoError(t, err)
	defer roomResp.Body.Close()
	room := ResponseToRoom(roomResp)

	// Step 3: Create availability list
	availabilityDTO := roomclient.CreateRoomAvailabilityListDTO{
		RoomID: room.ID,
		Items: []roomclient.CreateRoomAvailabilityItemDTO{
			{
				ExistingID: 0,
				DateFrom:   time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
				DateTo:     time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
				Available:  true,
			},
			{
				ExistingID: 0,
				DateFrom:   time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
				DateTo:     time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC),
				Available:  true,
			},
			{
				ExistingID: 0,
				DateFrom:   time.Date(2025, 9, 22, 0, 0, 0, 0, time.UTC),
				DateTo:     time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC),
				Available:  true,
			},
			{
				ExistingID: 0,
				DateFrom:   time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
				DateTo:     time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC),
				Available:  true,
			},
		},
	}
	availResp, err := CreateRoomAvailability(jwt, availabilityDTO)
	require.NoError(t, err)
	defer availResp.Body.Close()

	// Step 4: Create price list
	priceDTO := roomclient.CreateRoomPriceListDTO{
		RoomID:    room.ID,
		BasePrice: 80,
		PerGuest:  false,
		Items: []roomclient.CreateRoomPriceItemDTO{
			{
				ExistingID: 0,
				DateFrom:   time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
				DateTo:     time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
				Price:      100,
			},
			{
				ExistingID: 0,
				DateFrom:   time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
				DateTo:     time.Date(2025, 9, 20, 0, 0, 0, 0, time.UTC),
				Price:      120,
			},
			{
				ExistingID: 0,
				DateFrom:   time.Date(2025, 9, 22, 0, 0, 0, 0, time.UTC),
				DateTo:     time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC),
				Price:      200,
			},
			{
				ExistingID: 0,
				DateFrom:   time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
				DateTo:     time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC),
				Price:      200,
			},
		},
	}
	priceResp, err := CreateRoomPrice(jwt, priceDTO)
	require.NoError(t, err)
	defer priceResp.Body.Close()

	return username, password, jwt, room
}

func QueryRoomAvailability(jwt string, dto roomclient.RoomReservationQueryDTO) (roomclient.RoomReservationQueryResponseDTO, error) {
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return roomclient.RoomReservationQueryResponseDTO{}, fmt.Errorf("failed to marshal query DTO: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, URL_room+"reservation/query", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return roomclient.RoomReservationQueryResponseDTO{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+jwt)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return roomclient.RoomReservationQueryResponseDTO{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return roomclient.RoomReservationQueryResponseDTO{}, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Println("Raw response body:", string(bodyBytes))
	var result roomclient.RoomReservationQueryResponseDTO
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return roomclient.RoomReservationQueryResponseDTO{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

func CheckReservationAvailability(roomID uint, from, to string) (*http.Response, error) {
	url := fmt.Sprintf("/room/%d/availability?from=%s&to=%s", roomID, from, to)
	return http.Get(url)
}

func ResponseToReservationAvailability(resp *http.Response) AvailabilityResponse {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read response body: %v", err))
	}

	var obj AvailabilityResponse
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		panic(fmt.Sprintf("failed to unmarshal: %v", err))
	}

	return obj
}
