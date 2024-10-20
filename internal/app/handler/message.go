package handler

import (
	"RIP/internal/app/repository"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func (h *Handler) GetMessagesFiltered(ctx *gin.Context) {
	// Получаем параметры фильтрации из запроса
	status := ctx.Query("status")
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	// Парсим даты
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
	messages, err := h.Repository.GetMessagesFiltered(status, startDateStr != "", endDateStr != "", startDate, endDate)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Возвращаем результат клиенту
	ctx.JSON(http.StatusOK, messages)
}

func (h *Handler) GetMessage(ctx *gin.Context) {
	messageID := ctx.Param("id") // Получаем ID сообщения из URL

	// Получаем сообщение из репозитория
	message, chats, err := h.Repository.GetMessage(messageID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Формируем результат
	messageDetail := repository.MessageDetail{
		ID:     message.ID,
		Status: message.Status,
		//Text:       message.Text.String,
		Text:       message.Text,
		DateCreate: message.DateCreate,
		DateUpdate: message.DateUpdate.Time,
		DateFinish: message.DateFinish.Time,
		Creator:    message.Creator.Login,
		Moderator:  message.Moderator.Login,
		Chats:      chats,
	}

	// Возвращаем результат клиенту
	ctx.JSON(http.StatusOK, messageDetail)
}

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

	// Обновляем текст сообщения через репозиторий
	if err := h.Repository.UpdateMessageText(uint(messageID), input.Text); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	// Возвращаем успешный ответ
	ctx.JSON(http.StatusOK, gin.H{"message": "Text updated successfully"})
}

func (h *Handler) MessageForm(ctx *gin.Context) {
	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Invalid message ID")
		return
	}

	creatorID := getCurrentUserID(ctx) // Получаем ID текущего пользователя

	if err := h.Repository.MessageForm(uint(messageID), creatorID); err != nil {
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

func (h *Handler) MessageFinish(ctx *gin.Context) {
	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Неверный message ID")
		return
	}

	moderatorID := getCurrentModeratorID(ctx) // Получаем ID текущего пользователя

	if err := h.Repository.MessageFinish(uint(messageID), moderatorID); err != nil {
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

func (h *Handler) MessageReject(ctx *gin.Context) {
	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Неверный message ID")
		return
	}

	moderatorID := getCurrentModeratorID(ctx) // Получаем ID текущего пользователя

	if err := h.Repository.MessageReject(uint(messageID), moderatorID); err != nil {
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

func (h *Handler) MessageDelete(ctx *gin.Context) {
	messageID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "Invalid message ID")
		return
	}

	creatorID := getCurrentUserID(ctx) // Получаем ID текущего пользователя

	if err := h.Repository.MessageDelete(uint(messageID), creatorID); err != nil {
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
