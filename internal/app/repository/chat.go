package repository

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"RIP/internal/app/ds"
	"github.com/minio/minio-go/v7"
	_ "github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func (r *Repository) GetChats(userID uint, name string) ([]ds.ChatResponse, uint, int64, error) {
	var chats []ds.ChatResponse

	query := r.db.Model(&ds.ChatResponse{}).Where("is_delete = ?", false)

	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	err := query.Model(&ds.Chat{}).Find(&chats).Error
	if err != nil {
		return nil, 0, 0, fmt.Errorf("ошибка при получении списка чатов: %w", err)
	}

	if userID == 0 {
		return chats, 0, 0, nil
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

func (r *Repository) GetChatByID(chatID uint) (ds.ChatResponse, error) {
	var chat ds.ChatResponse

	// Поиск чата в базе данных
	err := r.db.Model(&ds.Chat{}).Where("is_delete = ?", false).First(&chat, chatID).Error
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
	err := r.db.Model(&ds.Chat{}).Where("id = ?", chat.ID).Updates(map[string]interface{}{
		"name":     chat.Name,
		"info":     chat.Info,
		"nickname": chat.Nickname,
	}).Error
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

func (r *Repository) AddChatToMessage(chatID, userID uint) error {
	var message ds.Message

	err := r.db.Where("status = ? AND creator_id = ?", "черновик", userID).First(&message).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newList := ds.Message{
			Status:     "черновик",
			DateCreate: time.Now(),
			CreatorID:  userID,
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
	var chat ds.Chat

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
	if errMinio != nil {
		return errMinio
	}

	chat.Img = fmt.Sprintf("http://127.0.0.1:9000/test/%d.png", chatID)
	errDB := r.db.Save(&chat).Error
	if errDB != nil {
		return fmt.Errorf("ошибка загрузки нового изображения для чата %d: %v", chatID, errDB)
	}

	return nil
}
