package ds

type Users struct {
	ID          uint   `json:"id" gorm:"primary_key"`
	Login       string `gorm:"type:varchar(255);unique" json:"login"`
	Password    string `gorm:"type:varchar(255)" json:"-"`
	IsModerator bool   `gorm:"type:boolean" json:"is_moderator"`
}
