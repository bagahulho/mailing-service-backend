package repository

import (
	"RIP/internal/app/ds"
	"database/sql"
	"errors"
	"time"
)

func (r *Repository) GetMessagesFiltered(status string, hasStartDate, hasEndDate bool, startDate, endDate time.Time) ([]MessageWithUsers, error) {
	var messages []MessageWithUsers

	query := r.db.Table("messages").
		Select("messages.id, messages.status, messages.text, messages.date_create, messages.date_update, messages.date_finish, u1.login as creator").
		Joins("JOIN users u1 ON messages.creator_id = u1.id").
		Where("messages.status != ? AND messages.status != ?", "удалён", "черновик")

	// Добавляем фильтрацию по статусу, если он указан
	if status != "" {
		query = query.Where("messages.status = ?", status)
	}

	// Добавляем фильтрацию по диапазону дат, если даты указаны
	if hasStartDate {
		query = query.Where("messages.date_update >= ?", startDate)
	}
	if hasEndDate {
		query = query.Where("messages.date_update <= ?", endDate)
	}

	if err := query.Find(&messages).Error; err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *Repository) GetMessage(messageID string) (ds.Message, []Chat, error) {
	var message ds.Message
	var chats []Chat

	// Получаем сообщение
	if err := r.db.Preload("Creator").Preload("Moderator").First(&message, messageID).Error; err != nil {
		return ds.Message{}, nil, err
	}

	// Получаем чаты, связанные с сообщением
	if err := r.db.Model(&ds.MessageChat{}).
		Select("chats.*").
		Joins("JOIN chats ON message_chats.chat_id = chats.id").
		Where("message_chats.message_id = ?", message.ID).
		Find(&chats).Error; err != nil {
		return ds.Message{}, nil, err
	}

	return message, chats, nil
}

func (r *Repository) UpdateMessageText(messageID uint, newText string) error {
	var message ds.Message

	// Находим сообщение по ID
	if err := r.db.First(&message, messageID).Error; err != nil {
		return err // Возвращаем ошибку, если сообщение не найдено
	}

	// Обновляем текст сообщения
	//message.Text = sql.NullString{String: newText, Valid: true}
	message.Text = newText
	message.DateUpdate = sql.NullTime{Time: time.Now(), Valid: true} // Обновляем дату изменения

	// Сохраняем изменения
	return r.db.Save(&message).Error
}

func (r *Repository) MessageForm(messageID uint, creatorID uint) error {
	var message ds.Message

	// Находим сообщение по ID
	if err := r.db.First(&message, messageID).Error; err != nil {
		return err // Возвращаем ошибку, если сообщение не найдено
	}

	// Проверяем, является ли текущий пользователь создателем сообщения
	if message.CreatorID != creatorID {
		return errors.New("только создатель заявки может ее изменить")
	}

	if message.Text == "" {
		return errors.New("введите текст сообщения: оно не может быть пустым")
	}

	// Проверяем, что статус не изменяется на "сформирован" повторно
	if message.Status == "сформирован" {
		return errors.New("статус уже установлен на 'сформирован'")
	} else if message.Status == "завершён" {
		return errors.New("это сообщение завершено")
	} else if message.Status == "отклонён" {
		return errors.New("это сообщение отклонено")
	} else if message.Status == "удалён" {
		return errors.New("это сообщение уже удалено")
	}

	// Обновляем статус и дату формирования
	message.Status = "сформирован"
	message.DateUpdate = sql.NullTime{Time: time.Now(), Valid: true} // Устанавливаем дату завершения

	return r.db.Save(&message).Error
}

func (r *Repository) MessageFinish(messageID uint, moderatorID uint) error {
	var message ds.Message

	// Находим сообщение по ID
	if err := r.db.First(&message, messageID).Error; err != nil {
		return err // Возвращаем ошибку, если сообщение не найдено
	}

	message.ModeratorID = &moderatorID

	// Проверяем, что статус не изменяется на "сформирован" повторно
	if message.Status == "завершён" {
		return errors.New("статус уже установлен на 'завершён'")
	} else if message.Status == "черновик" {
		return errors.New("сообщение не сформировано, вы не можете его завершить")
	} else if message.Status == "удалён" {
		return errors.New("сообщение удалено")
	} else if message.Status == "отклонён" {
		return errors.New("это сообщение уже отклонено")
	}

	message.Status = "завершён"
	message.DateFinish = sql.NullTime{Time: time.Now(), Valid: true}

	return r.db.Save(&message).Error
}

func (r *Repository) MessageReject(messageID uint, moderatorID uint) error {
	var message ds.Message

	// Находим сообщение по ID
	if err := r.db.First(&message, messageID).Error; err != nil {
		return err // Возвращаем ошибку, если сообщение не найдено
	}

	message.ModeratorID = &moderatorID

	// Проверяем, что статус не изменяется на "сформирован" повторно
	if message.Status == "отклонён" {
		return errors.New("статус уже установлен на 'отклонён'")
	} else if message.Status == "черновик" {
		return errors.New("сообщение не сформировано, вы не можете его отклонить")
	} else if message.Status == "удалён" {
		return errors.New("сообщение удалено")
	} else if message.Status == "завершён" {
		return errors.New("это сообщение уже завершено")
	}

	message.Status = "отклонён"
	message.DateFinish = sql.NullTime{Time: time.Now(), Valid: true}

	return r.db.Save(&message).Error
}

func (r *Repository) MessageDelete(messageID uint, creatorID uint) error {
	var message ds.Message

	// Находим сообщение по ID
	if err := r.db.First(&message, messageID).Error; err != nil {
		return err // Возвращаем ошибку, если сообщение не найдено
	}

	// Проверяем, является ли текущий пользователь создателем сообщения
	if message.CreatorID != creatorID {
		return errors.New("только создатель заявки может ее изменить")
	}

	// Проверяем, что статус не изменяется на "сформирован" повторно
	if message.Status == "удалён" {
		return errors.New("статус уже установлен на 'удалён'")
	} else if message.Status == "завершён" {
		return errors.New("это сообщение завершено")
	} else if message.Status == "отклонён" {
		return errors.New("это сообщение отклонено")
	} else if message.Status == "сформирован" {
		return errors.New("это сообщение уже сформировано")
	}

	// Обновляем статус и дату формирования
	message.Status = "удалён"
	message.DateUpdate = sql.NullTime{Time: time.Now(), Valid: true} // Устанавливаем дату завершения

	return r.db.Save(&message).Error
}
