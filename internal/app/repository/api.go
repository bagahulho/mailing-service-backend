package repository

import (
	"RIP/internal/app/ds"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"io"
	"time"
)

func (r *Repository) GetChats(userID uint, name string) ([]Chat, uint, int64, error) {
	var chats []Chat

	query := r.db.Model(&Chat{}).Where("is_delete = ?", false)

	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	err := query.Find(&chats).Error
	if err != nil {
		return nil, 0, 0, fmt.Errorf("ошибка при получении списка чатов: %w", err)
	}

	var draft ds.Message
	err = r.db.Where("creator_id = ? AND status = ?", userID, "черновик").First(&draft).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, 0, fmt.Errorf("ошибка получения черновика: %w", err)
	}

	var count int64
	if draft.ID != 0 {
		err = r.db.Model(&ds.MessageChat{}).Where("message_id = ?", draft.ID).Count(&count).Error
		if err != nil {
			return nil, 0, 0, fmt.Errorf("ошибка подсчета чатов в черновике: %w", err)
		}
	}

	return chats, draft.ID, count, nil
}

func (r *Repository) GetChatByID(chatID uint) (Chat, error) {
	var chat Chat

	// Поиск чата в базе данных
	err := r.db.Where("is_delete = ?", false).First(&chat, chatID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return chat, fmt.Errorf("чат с id %d не найден", chatID)
		}
		return chat, fmt.Errorf("ошибка при получении чата: %w", err)
	}

	return chat, nil
}

func (r *Repository) CreateChat(chat ds.Chat) (int, error) {
	// Сохранение нового чата в базе данных
	err := r.db.Create(&chat).Error
	if err != nil {
		return 0, fmt.Errorf("ошибка при создании чата: %w", err)
	}

	return chat.ID, nil
}

func (r *Repository) UpdateChat(chat ds.Chat) error {
	err := r.db.Save(&chat).Error
	if err != nil {
		return fmt.Errorf("ошибка при обновлении чата с id %d: %w", chat.ID, err)
	}
	return nil
}

func (r *Repository) DeleteChat(chatID uint, imageName string) error {
	// Удаление изображения из Minio
	err := r.MinioClient.RemoveObject(context.Background(), "test", imageName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("ошибка при удалении изображения")
	}

	err = r.db.Delete(&ds.Chat{}, chatID).Error
	if err != nil {
		return fmt.Errorf("ошибка при удалении чата с id %d: %w", chatID, err)
	}

	return nil
}

func (r *Repository) AddChatToMessage(chatID uint) error {
	var message ds.Message

	err := r.db.Where("status = ?", "черновик").First(&message).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newList := ds.Message{
			Status:     "черновик",
			DateCreate: time.Now(),
			CreatorID:  1,
		}
		if err := r.db.Create(&newList).Error; err != nil {
			return fmt.Errorf("ошибка создания черновика")
		}
		message = newList
	} else if err != nil {
		return fmt.Errorf("ошибка при поиске черновика")
	}

	// Сохраняем изменения
	err = r.db.Save(&message).Error
	if err != nil {
		return fmt.Errorf("ошибка обновления даты изменения")
	}

	messageChat := ds.MessageChat{
		MessageID: message.ID,
		ChatID:    chatID,
		Sound:     true,
	}

	if err := r.db.Create(&messageChat).Error; err != nil {
		return fmt.Errorf("ошибка при добавлении чата в сообщение")
	}

	return nil
}

func (r *Repository) ReplaceChatImage(chatID uint, imageName string, imageFile io.Reader, imageSize int64) error {
	var chat Chat

	// Найти чат по ID
	if err := r.db.First(&chat, chatID).Error; err != nil {
		return fmt.Errorf("чат с id %d не найден: %w", chatID, err)
	}

	// Если старое изображение существует, удалить его из Minio
	if chat.Img != "" {
		err := r.MinioClient.RemoveObject(context.Background(), "test", imageName, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("ошибка удаления старого изображения %s: %v", chat.Img, err)
		}
	}

	// Загрузить новое изображение в Minio
	_, errMinio := r.MinioClient.PutObject(context.Background(), "test", imageName, imageFile, imageSize, minio.PutObjectOptions{
		ContentType: "image/png",
	})

	chat.Img = fmt.Sprintf("http://127.0.0.1:9000/test/%d.png", chatID)
	errDB := r.db.Save(&chat).Error

	if errMinio != nil || errDB != nil {
		return fmt.Errorf("ошибка загрузки нового изображения для чата %d", chatID)
	}

	return nil
}

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

func (r *Repository) DeleteChatFromMessage(messageID uint, chatID uint) error {
	if err := r.db.Where("message_id = ? AND chat_id = ?", messageID, chatID).Delete(&ds.MessageChat{}).Error; err != nil {
		return err
	}
	return nil
}

func (r *Repository) ToggleSoundField(messageID uint, chatID uint) (bool, error) {
	var messageChat ds.MessageChat

	if err := r.db.Where("message_id = ? AND chat_id = ?", messageID, chatID).First(&messageChat).Error; err != nil {
		return false, err
	}

	messageChat.Sound = !messageChat.Sound

	// Сохраняем изменения.
	if err := r.db.Save(&messageChat).Error; err != nil {
		return false, err
	}

	return messageChat.Sound, nil
}

func (r *Repository) CreateUser(user *ds.User) error {
	// Проверяем, существует ли уже пользователь с таким логином
	var existingUser ds.User
	if err := r.db.Where("login = ?", user.Login).First(&existingUser).Error; err == nil {
		return errors.New("пользователь с таким логином уже существует")
	}

	return r.db.Create(user).Error
}

func (r *Repository) UpdateUser(newUser ds.User, id uint) (ds.User, error) {
	var user ds.User
	if err := r.db.First(&user, id).Error; err != nil {
		return user, fmt.Errorf("пользователь с id %d не найден", id)
	}

	if newUser.Login != "" {
		user.Login = newUser.Login
	}
	if newUser.Password != "" {
		user.Password = newUser.Password
	}

	if err := r.db.Save(user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func (r *Repository) AuthUser() {

}

func (r *Repository) DeAuthUser() {

}
