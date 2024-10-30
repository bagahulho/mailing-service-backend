package ds

type MessageChat struct {
	ID        uint `gorm:"primaryKey"`
	MessageID uint `gorm:"not null;uniqueIndex:idx_message_chat"`
	ChatID    uint `gorm:"not null;uniqueIndex:idx_message_chat"`
	Sound     bool `gorm:"default:true"`
	IsRead    bool `gorm:"default:false"`

	Message Message `gorm:"foreignKey:MessageID"`
	Chat    Chat    `gorm:"foreignKey:ChatID"`
}
