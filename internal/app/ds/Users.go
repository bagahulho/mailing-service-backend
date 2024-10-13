package ds

type Users struct {
	ID          uint   `gorm:"primary_key" json:"id"`
	Login       string `gorm:"type:varchar(25);unique;not null" json:"login"`
	Password    string `gorm:"type:varchar(100);not null" json:"-"`
	IsModerator bool   `gorm:"type:boolean;default:false" json:"is_moderator"`
}

//type Users struct {
//	ID       uint   `gorm:"primaryKey"`
//	Login    string `gorm:"type:char(15);not null"`
//	Password string `gorm:"type:char(100);not null"`
//	Admin    bool   `gorm:"default:false"`
//	// Связи
//	Chats     []Chat    `gorm:"many2many:message_chats;"`
//	Messages  []Message `gorm:"foreignKey:UserID"`
//	Moderated []Message `gorm:"foreignKey:ModeratorID"`
//}
