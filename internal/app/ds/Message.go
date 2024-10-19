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
