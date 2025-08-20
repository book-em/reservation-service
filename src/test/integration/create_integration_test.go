package test

import (
	"bookem-reservation-service/client/userclient"
	test "bookem-reservation-service/test/unit"
	"bookem-reservation-service/util"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegration_Create_Stub(t *testing.T) {
	RegisterUser("user1", "1234", userclient.Guest)
	jwt := LoginUser2("user1", "1234")
	jwtObj, err := util.GetJwtFromString(jwt)
	if err != nil {
		panic(err)
	}

	dto := test.DefaultReservationDTO
	dto.GuestID = jwtObj.ID
	resp, err := CreateReservation(jwt, dto)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
