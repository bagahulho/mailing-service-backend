package handler

import (
	"RIP/internal/app/ds"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) CreateUser(ctx *gin.Context) {
	var input ds.User

	// Чтение JSON-запроса в структуру `input`.
	if err := ctx.BindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Примитивная валидация.
	if strings.TrimSpace(input.Login) == "" || strings.TrimSpace(input.Password) == "" {
		h.errorHandler(ctx, http.StatusBadRequest, "Login and password are required")
		return
	}

	// Добавление нового пользователя в БД.
	if err := h.Repository.CreateUser(&input); err != nil {
		if err.Error() == "пользователь с таким логином уже существует" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"id":           input.ID,
		"login":        input.Login,
		"is_moderator": input.IsModerator,
	})
}

func (h *Handler) UpdateUser(ctx *gin.Context) {
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "некорректный ID пользователя")
		return
	}

	// Находим пользователя по ID.
	//user, err := repo.GetUserByID(userID)
	//if err != nil {
	//	c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	//	return
	//}

	var input ds.User

	// Читаем данные из тела запроса.
	if err := ctx.BindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	//// Проверяем, изменился ли логин, и валидируем его.
	//if strings.TrimSpace(input.Login) != "" && input.Login != user.Login {
	//	user.Login = input.Login
	//} else {
	//	c.JSON(http.StatusBadRequest, gin.H{"error": "Login is required and must be unique"})
	//	return
	//}

	// Обновляем пользователя в базе данных.
	newUser, err := h.Repository.UpdateUser(input, uint(userID))
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

func (h *Handler) AuthUser(ctx *gin.Context) {

}

func (h *Handler) DeAuthUser(ctx *gin.Context) {

}
