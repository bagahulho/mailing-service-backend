package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"RIP/internal/app/ds"
	"RIP/internal/utils"
	"github.com/gin-gonic/gin"
)

// RegisterUser Регистрация нового пользователя
// @Summary Регистрация пользователя
// @Description Создает нового пользователя с указанными логином и паролем
// @Tags Auth
// @Param input body ds.UserRespReq true "Данные для регистрации пользователя"
// @Produce json
// @Router /user/reg [post]
func (h *Handler) RegisterUser(ctx *gin.Context) {
	//var input ds.User
	var input ds.UserRegisterReq

	// Чтение JSON-запроса в структуру `input`.
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Неверный формат данных")
		return
	}

	// Примитивная валидация.
	if strings.TrimSpace(input.Login) == "" || strings.TrimSpace(input.Password) == "" {
		h.errorHandler(ctx, http.StatusBadRequest, "Логин и пароль обязательны")
		return
	}
	if input.Password != input.RepeatPassword {
		h.errorHandler(ctx, http.StatusBadRequest, "Пароли не совпадают")
		return
	}

	// Добавление нового пользователя в БД.
	if err := h.repository.RegisterUser(input.Login, input.Password); err != nil {
		if err.Error() == "пользователь с таким логином уже существует" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("пользователь %s создан", input.Login),
	})
}

// Authenticate Аутентификация пользователя
// @Summary Вход пользователя
// @Description Аутентификация пользователя и создание JWT токена
// @Tags Auth
// @Param input body ds.UserRespReq true "Данные для входа"
// @Produce json
// @Router /user/login [post]
func (h *Handler) Authenticate(ctx *gin.Context) {
	var req ds.UserRespReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	user, err := h.repository.AuthenticateUser(req.Login, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные учетные данные"})
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Login, user.IsModerator)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать токен"})
		return
	}

	// Сохранение сессии в Redis с использованием контекста запроса
	if err := h.repository.SaveSession(ctx.Request.Context(), user.ID, token, 1*time.Hour); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить сессию"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"token": token})
}

// Logout Выход пользователя
// @Summary Выход из системы
// @Description Удаляет текущую сессию пользователя и завершает сеанс
// @Tags Auth
// @Security ApiKeyAuth
// @Produce json
// @Router /user/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	// Удаление сессии из Redis
	err := h.repository.DeleteSession(c.Request.Context(), userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось завершить сессию"})
		return
	}

	// Успешный ответ
	c.JSON(http.StatusOK, gin.H{"message": "Вы успешно вышли из системы"})
}

func (h *Handler) UpdateUser(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	var input ds.UserUpdateReq

	// Читаем данные из тела запроса.
	if err := ctx.BindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Обновляем пользователя в базе данных.
	newUser, err := h.repository.UpdateUser(input, userID.(uint))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Возвращаем успешный ответ.
	ctx.JSON(http.StatusOK, gin.H{
		"id":           newUser.ID,
		"login":        newUser.Login,
		"is_moderator": newUser.IsModerator,
	})
}