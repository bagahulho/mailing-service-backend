package handler

import (
	"RIP/internal/app/ds"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

// GetChats @Summary Получить чаты пользователя
// @Description Возвращает список чатов для конкретного пользователя с указанием черновиков
// @Tags Chats
// @Param name query string false "Фильтр по имени чата"
// @Success 200 {object} ds.GetChatsResponse
// @Failure 500 {object} ds.ErrorResp
// @Router /chats [get]
func (h *Handler) GetChats(ctx *gin.Context) {
	//userID := ctx.MustGet("userID").(uint)
	userID, exists := ctx.Get("userID")
	if !exists {
		userID = 0
	}

	name := ctx.Query("name")

	// Вызов репозитория для получения чатов
	chats, draftID, draftCount, err := h.repository.GetChats(uint(userID.(int)), name)
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

// GetChatByID @Summary Получить чат по ID
// @Description Возвращает информацию о чате по его ID
// @Tags Chats
// @Param id path int true "ID чата"
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

// CreateChat @Summary Создать новый чат
// @Description Создает новый чат с указанными данными
// @Tags Chats
// @Param chat body ds.ChatRequest true "Информация о чате"
// @Accept json
// @Produce json
// @Security BearerAuth
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

// UpdateChat @Summary Обновить чат
// @Description Обновляет данные существующего чата
// @Tags Chats
// @Accept json
// @Produce json
// @Param id path int true "ID чата"
// @Param chat body ds.ChatRequest true "Обновленные данные чата"
// @Security BearerAuth
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

// DeleteChat @Summary Удалить чат
// @Description Удаляет чат по его ID и удаляет изображение из Minio
// @Tags Chats
// @Param id path int true "ID чата"
// @Security BearerAuth
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

// AddChatToMessage @Summary Добавить чат в сообщение
// @Description Добавляет чат к конкретному сообщению
// @Tags Chats
// @Param id path int true "ID чата"
// @Security BearerAuth
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

// ReplaceChatImage @Summary Заменить изображение чата
// @Description Загружает и заменяет изображение чата
// @Tags Chats
// @Accept multipart/form-data
// @Param id path int true "ID чата"
// @Param image formData file true "Новое изображение для чата"
// @Security BearerAuth
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
