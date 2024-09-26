package ds

import (
	"database/sql"
	"time"
)

type List struct {
	ID          uint `gorm:"primaryKey"`
	Status      string
	DateCreate  time.Time
	DateUpdate  time.Time
	DateFinish  sql.NullTime `gorm:"default:null"`
	CreatorID   uint
	ModeratorID uint
	Creator     Users `gorm:"foreignKey:CreatorID"`
	Moderator   Users `gorm:"foreignKey:ModeratorID"`
}
