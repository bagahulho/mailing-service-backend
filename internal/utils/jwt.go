package utils

import (
	"github.com/golang-jwt/jwt"
	"os"
	"time"
)

func GenerateJWT(userID uint, isModerator bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":      userID,
		"isModerator": isModerator,
		"exp":         time.Now().Add(1 * time.Hour).Unix(), // Токен действует 1 час
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
