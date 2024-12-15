package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DeleteChatFromMessage
// @Summary Удаление чата из сообщения
// @Description Удаляет чат с указанным ID из сообщения с указанным ID.
// @Tags Message-Chats
// @Produce json
// @Param message_id path int true "ID сообщения"
// @Param chat_id path int true "ID чата"
// @Security BearerAuth
// @Success 200 {object} ds.OkResp "Сообщение об успешном удалении чата"
// @Failure 400 {object} ds.ErrorResp "Некорректный ID"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /message-chats/delete/{message_id}/{chat_id} [delete]
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

	if err := h.repository.DeleteChatFromMessage(uint(messageID), uint(chatID)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, "Ошибка при удалении чата из сообщения")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Чат успешно удален из сообщения"})
}

// ToggleSoundField
// @Summary Переключение поля "со звуком"
// @Description Переключает значение поля "со звуком" у чата в сообщении.
// @Tags Message-Chats
// @Produce json
// @Param message_id path int true "ID сообщения"
// @Param chat_id path int true "ID чата"
// @Security BearerAuth
// @Success 200 {object} ds.OkResp "Сообщение об успешном изменении значения"
// @Failure 400 {object} ds.ErrorResp "Некорректный ID"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /message-chats/switch/{message_id}/{chat_id} [put]
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

	sound, err := h.repository.ToggleSoundField(uint(messageID), uint(chatID))

	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, "Ошибка изменения поля 'со звуком'")
		return
	}

	// Возвращаем успешный ответ.
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Значение успешно изменено на '%v'", sound)})
}
