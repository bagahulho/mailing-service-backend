package handler

import (
	"RIP/internal/app/ds"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) GetChats(ctx *gin.Context) {
	userID := getCurrentUserID(ctx)

	name := ctx.Query("name")

	// Вызов репозитория для получения чатов
	chats, draftID, draftCount, err := h.Repository.GetChats(userID, name)
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

func (h *Handler) GetChatByID(ctx *gin.Context) {
	chatID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "некорректный ID чата")
		return
	}

	chat, err := h.Repository.GetChatByID(uint(chatID))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, chat)
}

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

	chatID, err := h.Repository.CreateChat(chat)

	chat.ID = chatID
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Возвращение успешного ответа с созданным чатом
	ctx.JSON(http.StatusCreated, chat)
}

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
	err = h.Repository.UpdateChat(chat)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Возвращение успешного ответа с обновленным чатом
	ctx.JSON(http.StatusOK, chat)
}

func (h *Handler) DeleteChat(ctx *gin.Context) {
	chatID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "неверный id чата")
		return
	}

	// Форматирование имени изображения для удаления из Minio
	imageName := fmt.Sprintf("%d.png", chatID)

	// Вызов репозитория для удаления чата
	err = h.Repository.DeleteChat(uint(chatID), imageName)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("чат (id-%d) успешно удален", chatID)})
}

func (h *Handler) AddChatToMessage(ctx *gin.Context) {
	chatID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "некорректный ID чата")
		return
	}

	err = h.Repository.AddChatToMessage(uint(chatID))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": fmt.Sprintf("чат (id-%d) успешно добавлен в сообщение", chatID)})
}

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
	err = h.Repository.ReplaceChatImage(uint(chatID), imageName, file, fileHeader.Size)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Image successfully uploaded"})
}
