package utils

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt"
)

func GenerateJWT(userID uint, username string, isModerator bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":      userID,
		"username":    username,
		"isModerator": isModerator,
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неверный метод подписи")
		}
		return []byte(os.Getenv("JWT_KEY")), nil // Используем тот же секретный ключ
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}
