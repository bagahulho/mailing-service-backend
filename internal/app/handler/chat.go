package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"RIP/internal/app/ds"
	"RIP/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// GetChats @Summary Получить чаты пользователя
// @Description Возвращает список чатов для конкретного пользователя с указанием черновиков.
// @Tags Chats
// @Param name query string false "Фильтр по имени чата"
// @Produce json
// @Success 200 {object} ds.GetChatsResponse "Список чатов с черновиками"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /chats [get]
func (h *Handler) GetChats(ctx *gin.Context) {
	userID := uint(0)
	name := ctx.Query("name")
	authHeader := ctx.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenStr := parts[1]
			token, _ := utils.ParseJWT(tokenStr)
			if token != nil {
				claims, ok := token.Claims.(jwt.MapClaims)
				if ok {
					userID = uint(claims["userID"].(float64))
				}
			}
		}
	}

	chats, draftID, draftCount, err := h.repository.GetChats(userID, name)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"chats":       chats,
		"draft_count": draftCount,
		"draft_ID":    draftID,
	})
}

// GetChatByID
// @Summary Получить чат по ID
// @Description Возвращает информацию о чате по его ID.
// @Tags Chats
// @Produce json
// @Param id path int true "ID чата"
// @Success 200 {object} ds.ChatResponse "Информация о чате"
// @Failure 400 {object} ds.ErrorResp "Некорректный ID чата"
// @Failure 404 {object} ds.ErrorResp "Чат не найден"
// @Router /chats/{id} [get]
func (h *Handler) GetChatByID(ctx *gin.Context) {
	chatID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "некорректный ID чата")
		return
	}

	chat, err := h.repository.GetChatByID(uint(chatID))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, chat)
}

// CreateChat
// @Summary Создать новый чат
// @Description Создает новый чат с указанными данными.
// @Tags Chats
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param chat body ds.ChatRequest true "Информация о чате"
// @Success 201 {object} ds.Chat "Созданный чат"
// @Failure 400 {object} ds.ErrorResp "Неверные данные или отсутствует имя/никнейм"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /chats/create [post]
func (h *Handler) CreateChat(ctx *gin.Context) {
	var chat ds.Chat

	// Парсинг данных из тела запроса
	if err := ctx.BindJSON(&chat); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "неверные данные")
		return
	}

	// Проверка обязательных полей
	if strings.TrimSpace(chat.Name) == "" || strings.TrimSpace(chat.Nickname) == "" {
		h.errorHandler(ctx, http.StatusBadRequest, "необходимо указать имя и никнейм")
		return
	}

	chat.IsDelete = false

	chatID, err := h.repository.CreateChat(chat)

	chat.ID = chatID
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Возвращение успешного ответа с созданным чатом
	ctx.JSON(http.StatusCreated, chat)
}

// UpdateChat
// @Summary Обновить чат
// @Description Обновляет данные существующего чата.
// @Tags Chats
// @Accept json
// @Produce json
// @Param id path int true "ID чата"
// @Param chat body ds.ChatRequest true "Обновленные данные чата"
// @Security BearerAuth
// @Success 200 {object} ds.Chat "Обновлённый чат"
// @Failure 400 {object} ds.ErrorResp "Некорректный ID чата или неверные данные"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /chats/{id} [put]
func (h *Handler) UpdateChat(ctx *gin.Context) {
	var chat ds.Chat
	chatID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "неверный id чата")
		return
	}

	if err := ctx.BindJSON(&chat); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "неверные данные")
		return
	}

	chat.ID = chatID

	// Проверка обязательных полей
	if strings.TrimSpace(chat.Name) == "" || strings.TrimSpace(chat.Nickname) == "" {
		h.errorHandler(ctx, http.StatusBadRequest, "необходимо указать имя и никнейм")
		return
	}

	// Вызов репозитория для обновления чата
	err = h.repository.UpdateChat(chat)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Возвращение успешного ответа с обновленным чатом
	ctx.JSON(http.StatusOK, chat)
}

// DeleteChat
// @Summary Удалить чат
// @Description Удаляет чат по его ID и удаляет изображение из Minio.
// @Tags Chats
// @Param id path int true "ID чата"
// @Security BearerAuth
// @Produce json
// @Success 201 {object} ds.OkResp "Сообщение об успешном удалении"
// @Failure 400 {object} ds.ErrorResp "Некорректный ID чата"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /chats/{id} [delete]
func (h *Handler) DeleteChat(ctx *gin.Context) {
	chatID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "неверный id чата")
		return
	}

	// Форматирование имени изображения для удаления из Minio
	imageName := fmt.Sprintf("%d.png", chatID)

	// Вызов репозитория для удаления чата
	err = h.repository.DeleteChat(uint(chatID), imageName)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("чат (id-%d) успешно удален", chatID)})
}

// AddChatToMessage
// @Summary Добавить чат в сообщение
// @Description Добавляет чат к конкретному сообщению.
// @Tags Chats
// @Param id path int true "ID чата"
// @Security BearerAuth
// @Produce json
// @Success 201 {object} ds.OkResp "Сообщение о добавлении чата"
// @Failure 400 {object} ds.ErrorResp "Некорректный ID чата"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /chat/{id}/in-message [post]
func (h *Handler) AddChatToMessage(ctx *gin.Context) {
	chatID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "некорректный ID чата")
		return
	}
	userID := ctx.MustGet("userID").(uint)

	err = h.repository.AddChatToMessage(uint(chatID), userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("чат (id-%d) успешно добавлен в сообщение", chatID)})
}

// ReplaceChatImage
// @Summary Заменить изображение чата
// @Description Загружает и заменяет изображение чата.
// @Tags Chats
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID чата"
// @Param image formData file true "Новое изображение для чата"
// @Security BearerAuth
// @Success 200 {object} ds.OkResp "Сообщение об успешной загрузке"
// @Failure 400 {object} ds.ErrorResp "Невалидные данные или неправильный формат файла"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /chats/{id}/new-image [post]
func (h *Handler) ReplaceChatImage(ctx *gin.Context) {
	chatID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Invalid chat ID")
		return
	}

	// Получаем файл изображения из запроса
	file, fileHeader, err := ctx.Request.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Failed to upload image")
		return
	}
	defer file.Close()

	imageName := fmt.Sprintf("%d.png", chatID)

	// Передаем файл в репозиторий для обработки
	err = h.repository.ReplaceChatImage(uint(chatID), imageName, file, fileHeader.Size)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Image successfully uploaded"})
}
