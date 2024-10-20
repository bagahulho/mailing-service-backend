package repository

import "RIP/internal/app/ds"

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
