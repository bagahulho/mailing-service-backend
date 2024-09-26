package ds

type ListsChat struct {
	ListID        uint `gorm:"primary_key"`
	ChatID        uint `gorm:"primary_key"`
	List          List `gorm:"foreignKey:ListID"`
	Chat          Chat `gorm:"foreignKey:ChatID"`
	MessagesCount uint `gorm:"default:0"`
}
