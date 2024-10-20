package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func (h *Handler) DeleteChatFromMessage(ctx *gin.Context) {
	messageID, err := strconv.Atoi(ctx.Param("message_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Некорректный ID сообщения")
		return
	}

	chatID, err := strconv.Atoi(ctx.Param("chat_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Некорректный ID чата")
		return
	}

	if err := h.Repository.DeleteChatFromMessage(uint(messageID), uint(chatID)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, "Ошибка при удалении чата из сообщения")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Чат успешно удален из сообщения"})
}

func (h *Handler) ToggleSoundField(ctx *gin.Context) {
	messageID, err := strconv.Atoi(ctx.Param("message_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Некорректный ID сообщения")
		return
	}

	chatID, err := strconv.Atoi(ctx.Param("chat_id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Некорректный ID чата")
		return
	}

	sound, err := h.Repository.ToggleSoundField(uint(messageID), uint(chatID))

	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, "Ошибка изменения поля 'со звуком'")
		return
	}

	// Возвращаем успешный ответ.
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Значение успешно изменено на '%v'", sound)})
}
