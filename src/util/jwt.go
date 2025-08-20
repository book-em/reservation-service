package util

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var JWT_PUBLIC_KEY_PATH = os.Getenv("JWT_PUBLIC_KEY_PATH")

var ParseJWT = parseJWT

// ParseJWT validates and extracts claims from an encoded JWT string.
func parseJWT(tokenString string) (jwt.MapClaims, error) {
	publicKeyData, err := os.ReadFile(JWT_PUBLIC_KEY_PATH)
	if err != nil {
		return nil, fmt.Errorf("could not open public key %s: %w", JWT_PUBLIC_KEY_PATH, err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		return nil, fmt.Errorf("could not parse public key: %w", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return nil, fmt.Errorf("invalid jwt token or claims")
	}

	return claims, nil
}
