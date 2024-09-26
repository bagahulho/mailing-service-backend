package repository

import (
	"RIP/internal/app/ds"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
)

func (r *Repository) GetAllChats() ([]ds.Chat, error) {
	var chats []ds.Chat
	err := r.db.Find(&chats).Error
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
	err := r.db.Where("name ILIKE ?", "%"+name+"%").First(&chats).Error
	if err != nil {
		return nil, err
	}
	return chats, nil
}

func (r *Repository) GetRequest() ([]ds.Chat, error) {
	var listID uint
	var chatIDs []uint
	var chats []ds.Chat

	err := r.db.Model(&ds.List{}).Where("status = ?", "черновик").Select("id").First(&listID).Error
	if err != nil {
		return nil, fmt.Errorf("error finding list with status 'черновик': %w", err)
	}

	err = r.db.Model(&ds.ListsChat{}).Where("list_id = ?", listID).Pluck("chat_id", &chatIDs).Error
	if err != nil {
		return nil, fmt.Errorf("error finding chat_ids for list_id %d: %w", listID, err)
	}

	err = r.db.Model(&ds.Chat{}).Where("id IN ?", chatIDs).Find(&chats).Error
	if err != nil {
		return nil, fmt.Errorf("error finding chats for chat_ids %v: %w", chatIDs, err)
	}

	return chats, nil
}

func (r *Repository) AddChatToList(chatID uint) error {
	var list ds.List

	err := r.db.Where("status = ?", "черновик").First(&list).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		newList := ds.List{
			Status:      "черновик",
			DateCreate:  time.Now(),
			DateUpdate:  time.Now(),
			CreatorID:   1,
			ModeratorID: 2,
		}
		if err := r.db.Create(&newList).Error; err != nil {
			return err
		}
		list = newList
	} else if err != nil {
		return err
	}

	listChat := ds.ListsChat{
		ListID:        list.ID,
		ChatID:        chatID,
		MessagesCount: 0,
	}

	if err := r.db.Create(&listChat).Error; err != nil {
		return err
	}

	// Update the date of the list
	list.DateUpdate = time.Now()
	if err := r.db.Save(&list).Error; err != nil {
		return err
	}

	return nil
}

func (r *Repository) DeleteList() error {
	query := "UPDATE lists SET status = 'удалён', date_update = NOW() WHERE status = 'черновик'"

	err := r.db.Exec(query)
	if err != nil {
		return err.Error
	}

	return nil
}

func (r *Repository) GetCartCount() int64 {
	var listID uint
	var count int64

	err := r.db.Model(&ds.List{}).Where("status = ?", "черновик").Select("id").First(&listID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.ListsChat{}).Where("list_id = ?", listID).Count(&count).Error
	if err != nil {
		logrus.Println("Error counting records in lists_chats:", err)
	}

	return count
}
