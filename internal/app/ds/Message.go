package ds

import (
	"database/sql"
	"time"
)

type Message struct {
	ID          uint         `gorm:"primaryKey"`
	Status      string       `gorm:"type:varchar(15);not null"`
	Text        string       `gorm:"type:text;default:null"`
	DateCreate  time.Time    `gorm:"not null"`
	DateUpdate  sql.NullTime `gorm:"default:null"`
	DateFinish  sql.NullTime `gorm:"default:null"`
	CreatorID   uint         `gorm:"not null"`
	ModeratorID *uint        `gorm:"default:null"`

	Creator   User `gorm:"foreignKey:CreatorID"`
	Moderator User `gorm:"foreignKey:ModeratorID"`
}

type MessageWithUsers struct {
	ID         uint      `json:"id"`
	Status     string    `json:"status"`
	Text       string    `json:"text,omitempty"`
	DateCreate time.Time `json:"date_create"`
	DateUpdate time.Time `json:"date_update,omitempty"`
	DateFinish time.Time `json:"date_finish,omitempty"`
	Creator    string    `json:"creator"`
	Moderator  string    `json:"moderator,omitempty"`
}

type MessageDetail struct {
	ID         uint                    `json:"id"`
	Status     string                  `json:"status"`
	Text       string                  `json:"text,omitempty"`
	DateCreate time.Time               `json:"date_create"`
	DateUpdate time.Time               `json:"date_update,omitempty"`
	DateFinish time.Time               `json:"date_finish,omitempty"`
	Creator    string                  `json:"creator"`
	Moderator  string                  `json:"moderator,omitempty"`
	Chats      []ChatResponseWithFlags `json:"chats"`
}

type UpdateMessageTextInput struct {
	Text string `json:"text" binding:"required"` // Обязательное поле текста
}
