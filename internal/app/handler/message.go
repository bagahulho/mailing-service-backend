package handler

import (
	"net/http"
	"strconv"
	"time"

	"RIP/internal/app/ds"
	"github.com/gin-gonic/gin"
)

// GetMessagesFiltered
// @Summary Получить отфильтрованные сообщения
// @Description Возвращает список сообщений, отфильтрованных по статусу и диапазону дат.
// @Tags Messages
// @Accept json
// @Produce json
// @Param status query string false "Статус сообщения"
// @Param start_date query string false "Начальная дата (YYYY-MM-DD)"
// @Param end_date query string false "Конечная дата (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {array} ds.MessageWithUsers "Список сообщений"
// @Failure 400 {object} ds.ErrorResp "Некорректный формат дат"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /messages [get]
func (h *Handler) GetMessagesFiltered(ctx *gin.Context) {
	// Получаем параметры фильтрации из запроса
	status := ctx.Query("status")
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	userID := ctx.MustGet("userID").(uint)
	isModerator := ctx.MustGet("isModerator").(bool)

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil && startDateStr != "" {
		h.errorHandler(ctx, http.StatusBadRequest, "Invalid start date format")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil && endDateStr != "" {
		h.errorHandler(ctx, http.StatusBadRequest, "Invalid end date format")
		return
	}

	// Получаем сообщения из репозитория
	messages, err := h.repository.GetMessagesFiltered(status, startDateStr != "", endDateStr != "", startDate, endDate, userID, isModerator)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Возвращаем результат клиенту
	ctx.JSON(http.StatusOK, messages)
}

// GetMessage
// @Summary Получить сообщение по ID
// @Description Возвращает полные данные о сообщении, включая чаты.
// @Tags Messages
// @Accept json
// @Produce json
// @Param id path int true "ID сообщения"
// @Security BearerAuth
// @Success 200 {object} ds.MessageDetail "Детальная информация о сообщении"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /messages/{id} [get]
func (h *Handler) GetMessage(ctx *gin.Context) {
	messageID := ctx.Param("id") // Получаем ID сообщения из URL
	userID := ctx.MustGet("userID").(uint)
	isModerator := ctx.MustGet("isModerator").(bool)
	// Получаем сообщение из репозитория
	message, chats, err := h.repository.GetMessage(messageID, userID, isModerator)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Формируем результат
	messageDetail := ds.MessageDetail{
		ID:         message.ID,
		Status:     message.Status,
		Text:       message.Text,
		DateCreate: message.DateCreate,
		Creator:    message.Creator.Login,
		Moderator:  message.Moderator.Login,
		Chats:      chats,
	}

	// Возвращаем результат клиенту
	ctx.JSON(http.StatusOK, messageDetail)
}

// UpdateMessageText
// @Summary Обновить текст сообщения
// @Description Обновляет текст сообщения по ID.
// @Tags Messages
// @Accept json
// @Produce json
// @Param id path int true "ID сообщения"
// @Param message body ds.UpdateMessageTextInput true "Новый текст сообщения"
// @Security BearerAuth
// @Success 200 {object} ds.UpdateMessageTextResp "Обновлённый текст"
// @Failure 400 {object} ds.ErrorResp "Некорректный ID или неверный формат данных"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /messages/{id}/text [put]
func (h *Handler) UpdateMessageText(ctx *gin.Context) {
	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Invalid message ID")
		return
	}

	var input struct {
		Text string `json:"text" binding:"required"` // Обязательное поле текста
	}

	// Привязываем входящие данные к структуре
	if err := ctx.ShouldBindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userID := ctx.MustGet("userID").(uint)
	// Обновляем текст сообщения через репозиторий
	if err := h.repository.UpdateMessageText(uint(messageID), input.Text, userID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Возвращаем успешный ответ
	ctx.JSON(http.StatusOK, gin.H{"text": input.Text})
}

// MessageForm
// @Summary Сформировать сообщение
// @Description Устанавливает статус сообщения на "сформирован".
// @Tags Messages
// @Accept json
// @Produce json
// @Param id path int true "ID сообщения"
// @Security BearerAuth
// @Success 200 {object} ds.OkResp "Сообщение об успешном обновлении статуса"
// @Failure 400 {object} ds.ErrorResp "Неверный ID сообщения или пустой текст"
// @Failure 403 {object} ds.ErrorResp "Действие запрещено"
// @Failure 409 {object} ds.ErrorResp "Конфликт статусов"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /messages/{id}/form [put]
func (h *Handler) MessageForm(ctx *gin.Context) {
	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Invalid message ID")
		return
	}

	creatorID := ctx.MustGet("userID").(uint)

	if err := h.repository.MessageForm(uint(messageID), creatorID); err != nil {
		if err.Error() == "только создатель заявки может ее изменить" {
			h.errorHandler(ctx, http.StatusForbidden, err.Error())
		} else if err.Error() == "статус уже установлен на 'сформирован'" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "это сообщение завершено" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "это сообщение отклонено" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "это сообщение уже удалено" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "введите текст сообщения: оно не может быть пустым" {
			h.errorHandler(ctx, http.StatusBadRequest, err.Error())
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Статус обновлен на 'сформирован' успешно"})
}

// MessageFinish
// @Summary Завершить сообщение
// @Description Устанавливает статус сообщения на "завершён".
// @Tags Messages
// @Accept json
// @Produce json
// @Param id path int true "ID сообщения"
// @Security BearerAuth
// @Success 200 {object} ds.OkResp "Сообщение об успешном обновлении статуса"
// @Failure 409 {object} ds.ErrorResp "Конфликт статусов"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /messages/{id}/finish [put]
func (h *Handler) MessageFinish(ctx *gin.Context) {
	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Неверный message ID")
		return
	}

	moderatorID := ctx.MustGet("userID").(uint)

	if err := h.repository.MessageFinish(uint(messageID), moderatorID); err != nil {
		if err.Error() == "статус уже установлен на 'завершён'" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "сообщение не сформировано, вы не можете его завершить" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "сообщение удалено" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "это сообщение уже отклонено" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Статус обновлен на 'завершён' успешно"})
}

// MessageReject
// @Summary Отклонить сообщение
// @Description Устанавливает статус сообщения на "отклонён".
// @Tags Messages
// @Accept json
// @Produce json
// @Param id path int true "ID сообщения"
// @Security BearerAuth
// @Success 200 {object} ds.OkResp "Сообщение об успешном обновлении статуса"
// @Failure 409 {object} ds.ErrorResp "Конфликт статусов"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /messages/{id}/reject [put]
func (h *Handler) MessageReject(ctx *gin.Context) {
	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Неверный message ID")
		return
	}

	moderatorID := ctx.MustGet("userID").(uint) // Получаем ID текущего пользователя

	if err := h.repository.MessageReject(uint(messageID), moderatorID); err != nil {
		if err.Error() == "статус уже установлен на 'отклонён'" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "сообщение не сформировано, вы не можете его отклонить" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "сообщение удалено" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "это сообщение уже завершено" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Статус обновлен на 'отклонён' успешно"})
}

// MessageDelete
// @Summary Удалить сообщение
// @Description Устанавливает статус сообщения на "удалён".
// @Tags Messages
// @Accept json
// @Produce json
// @Param id path int true "ID сообщения"
// @Security BearerAuth
// @Success 200 {object} ds.OkResp "Сообщение об успешном обновлении статуса"
// @Failure 403 {object} ds.ErrorResp "Действие запрещено"
// @Failure 409 {object} ds.ErrorResp "Конфликт статусов"
// @Failure 500 {object} ds.ErrorResp "Внутренняя ошибка сервера"
// @Router /messages/{id}/delete [delete]
func (h *Handler) MessageDelete(ctx *gin.Context) {
	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Invalid message ID")
		return
	}

	creatorID := ctx.MustGet("userID").(uint) // Получаем ID текущего пользователя

	if err := h.repository.MessageDelete(uint(messageID), creatorID); err != nil {
		if err.Error() == "только создатель заявки может ее изменить" {
			h.errorHandler(ctx, http.StatusForbidden, err.Error())
		} else if err.Error() == "статус уже установлен на 'удалён'" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "это сообщение завершено" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "это сообщение отклонено" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else if err.Error() == "это сообщение уже сформировано" {
			h.errorHandler(ctx, http.StatusConflict, err.Error())
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Статус обновлен на 'удалён' успешно"})
}
