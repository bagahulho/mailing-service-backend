package handler

import (
	"net/http"
	"strings"

	"RIP/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func (h *Handler) AuthMiddleware(ctx *gin.Context) {
	// Извлечение токена из заголовка Authorization
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status":      "fail",
			"description": "Отсутствует токен",
		})
		ctx.Abort()
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := utils.ParseJWT(tokenString)

	if err != nil || !token.Valid {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status":      "fail",
			"description": "Недействительный токен",
		})
		ctx.Abort()
		return
	}

	// Извлечение userID из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status":      "fail",
			"description": "Ошибка обработки токена",
		})
		ctx.Abort()
		return
	}

	userID := uint(claims["userID"].(float64))
	isModerator := claims["isModerator"].(bool)

	// Проверка сессии в Redis
	redisToken, err := h.repository.GetSession(ctx.Request.Context(), userID)
	if err != nil || redisToken != tokenString {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status":      "fail",
			"description": "Сессия не найдена или истекла",
		})
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
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status":      "fail",
			"description": "Пользователь не авторизован",
		})
		ctx.Abort()
		return
	}

	isModerator := ctx.MustGet("isModerator").(bool)

	if !isModerator {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"status":      "fail",
			"description": "Доступ запрещен",
		})
		ctx.Abort()
		return
	}

	ctx.Next()
}
