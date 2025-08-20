package test

import (
	"bookem-reservation-service/client/userclient"
	"bookem-reservation-service/internal"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
)

const URL_user = "http://user-service:8080/api/"
const URL_room = "http://room-service:8080/api/"
const URL_reservation = "http://reservation-service:8080/api/"

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenName(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RegisterUser(username_or_email string, password string, role userclient.UserRole) (*http.Response, error) {
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
