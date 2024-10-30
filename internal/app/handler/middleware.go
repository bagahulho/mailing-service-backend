package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"net/http"
	"os"
	"strings"
)

func (h *Handler) AuthMiddleware(ctx *gin.Context) {
	// Извлечение токена из заголовка Authorization
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Отсутствует токен"})
		ctx.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Парсинг и валидация токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неверный метод подписи")
		}
		return []byte(os.Getenv("JWT_KEY")), nil // Используем тот же секретный ключ
	})

	if err != nil || !token.Valid {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
		ctx.Abort()
		return
	}

	// Извлечение userID из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Ошибка обработки токена"})
		ctx.Abort()
		return
	}

	userID := uint(claims["userID"].(float64))
	isModerator := claims["isModerator"].(bool)

	// Проверка сессии в Redis
	redisToken, err := h.repository.GetSession(ctx.Request.Context(), userID)
	if err != nil || redisToken != tokenString {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Сессия не найдена или истекла"})
		ctx.Abort()
		return
	}

	// Передаем userID в контекст
	ctx.Set("userID", userID)
	ctx.Set("isModerator", isModerator)
	ctx.Next()
}

func (h *Handler) ModeratorMiddleware(ctx *gin.Context) {
	_, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		ctx.Abort()
		return
	}

	isModerator := ctx.MustGet("isModerator").(bool)

	if !isModerator {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
		ctx.Abort()
		return
	}

	ctx.Next()
}
