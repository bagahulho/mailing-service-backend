package repository

import (
	"RIP/internal/app/ds"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
	"time"
)

func (r *Repository) GetAllChats() ([]ds.Chat, error) {
	var chats []ds.Chat
	//err := r.db.Find(&chats).Error
	err := r.db.Where("is_delete = false").Find(&chats).Error
	if err != nil {
		return nil, err
	}
	return chats, nil
}

func (r *Repository) GetChatByID(id int) (*ds.Chat, error) {
	chat := &ds.Chat{}
	err := r.db.Where("id = ?", id).First(chat).Error
	if err != nil {
		return nil, err
	}
	return chat, nil
}

func (r *Repository) SearchChatsByName(name string) ([]ds.Chat, error) {
	var chats []ds.Chat
	err := r.db.Where("name ILIKE ?", "%"+name+"%").Find(&chats).Error
	if err != nil {
		return nil, err
	}
	return chats, nil
}

func (r *Repository) GetDraftID() uint {
	var messageID uint
	err := r.db.Model(&ds.Message{}).Where("status = ?", "черновик").Select("id").First(&messageID).Error
	if err != nil {
		return 0
	}

	return messageID
}

func (r *Repository) GetMessageByID(messageID uint) (ds.Message, []ds.Chat, error) {
	var message ds.Message
	var chatIDs []uint
	var chats []ds.Chat

	err := r.db.First(&message, messageID).Error
	if err != nil {
		return message, nil, fmt.Errorf("error finding message with id %d: %w", messageID, err)
	}

	if strings.TrimSpace(strings.ToLower(message.Status)) == "удалён" {
		return message, nil, fmt.Errorf("данное сообщение удалено")
	}

	err = r.db.Model(&ds.MessageChat{}).Where("message_id = ?", messageID).Pluck("chat_id", &chatIDs).Error
	if err != nil {
		return message, nil, fmt.Errorf("error finding chat_ids for list_id %d: %w", messageID, err)
	}

	err = r.db.Model(&ds.Chat{}).Where("id IN ?", chatIDs).Find(&chats).Error
	if err != nil {
		return message, nil, fmt.Errorf("error finding chats for chat_ids %v: %w", chatIDs, err)
	}

	return message, chats, nil
}

func (r *Repository) AddChatToList(chatID uint) error {
	var message ds.Message

	err := r.db.Where("status = ?", "черновик").First(&message).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newList := ds.Message{
			Status:      "черновик",
			DateCreate:  time.Now(),
			DateUpdate:  time.Now(),
			CreatorID:   1,
			ModeratorID: 2,
		}
		if err := r.db.Create(&newList).Error; err != nil {
			return err
		}
		message = newList
	} else if err != nil {
		return err
	}

	messageChat := ds.MessageChat{
		MessageID: message.ID,
		ChatID:    chatID,
		Sound:     true,
	}

	if err := r.db.Create(&messageChat).Error; err != nil {
		return err
	}

	// Update the date of the message
	message.DateUpdate = time.Now()
	if err := r.db.Save(&message).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteList() error {
	query := "UPDATE messages SET status = 'удалён', date_update = NOW() WHERE status = 'черновик'"

	err := r.db.Exec(query)
	if err != nil {
		return err.Error
	}

	return nil
}

func (r *Repository) GetCartCount() int64 {
	var messageID uint
	var count int64

	err := r.db.Model(&ds.Message{}).Where("status = ?", "черновик").Select("id").First(&messageID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.MessageChat{}).Where("message_id = ?", messageID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in lists_chats:", err)
	}

	return count
}
