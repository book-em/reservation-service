package util

import (
	"bookem-reservation-service/client/userclient"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Jwt struct {
	ID       uint
	Username string
	Role     userclient.UserRole
}

func GetJwtString(ctx *gin.Context) (string, error) {
	header := ctx.GetHeader("Authorization")
	if header == "" {
		return "", fmt.Errorf("Unauthenticated")
	}

	if !strings.HasPrefix(header, "Bearer ") {
		return "", errors.New("invalid authorization model (must be Bearer)")
	}

	jwt := strings.SplitN(header, "Bearer ", 2)[1]
	_, err := ParseJWT(jwt)

	return jwt, err
}

func GetJwtData(ctx *gin.Context) (jwt.MapClaims, error) {
	jwtString, err := GetJwtString(ctx)
	if err != nil {
		return nil, err
	}

	jwt, err := ParseJWT(jwtString)
	if err != nil {
		return nil, err
	}

	return jwt, err
}

// GetJwt returns the JWT data embedded in the request header. If the user is
// unauthenticated (no JWT in the request), the function returns nil and an error.
func GetJwt(ctx *gin.Context) (*Jwt, error) {
	jwtData, err := GetJwtData(ctx)
	if err != nil {
		return nil, err
	}

	jwt := Jwt{
		ID:       uint(jwtData["sub"].(float64)),
		Username: jwtData["username"].(string),
		Role:     userclient.UserRole(jwtData["role"].(string)),
	}

	return &jwt, nil
}

func GetJwtFromString(jwtString string) (*Jwt, error) {
	jwtData, err := ParseJWT(jwtString)
	if err != nil {
		return nil, err
	}

	jwt := Jwt{
		ID:       uint(jwtData["sub"].(float64)),
		Username: jwtData["username"].(string),
		Role:     userclient.UserRole(jwtData["role"].(string)),
	}

	return &jwt, nil
}
