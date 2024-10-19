package handler

import (
	"RIP/internal/app/ds"
	"RIP/internal/app/repository"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func getCurrentUserID(c *gin.Context) uint {
	return 1 // временно возвращаем фиксированный ID
}
func getCurrentModeratorID(c *gin.Context) uint {
	return 2 // временно возвращаем фиксированный ID
}

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
