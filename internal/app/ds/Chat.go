package ds

type Chat struct {
	ID          int    `gorm:"primaryKey;autoIncrement"`
	IsDelete    bool   `gorm:"type:boolean not null;default:false"`
	Img         string `gorm:"type:varchar(100)"`
	Name        string `gorm:"type:varchar(25);not null;unique"`
	Info        string `gorm:"type:varchar(100)"`
	Nickname    string `gorm:"type:varchar(15);not null;unique"`
	Friends     int
	Subscribers int
}

type ChatRequest struct {
	Name        string `json:"name"`
	Info        string `json:"info"`
	Nickname    string `json:"nickname"`
	Friends     int    `json:"friends"`
	Subscribers int    `json:"subscribers"`
}

type GetChatsResponse struct {
	Chats      []ChatResponse `json:"chats"`
	DraftID    int            `json:"draft_ID"`
	DraftCount int            `json:"draft_count"`
}

type ChatResponse struct {
	ID          int    `json:"id"`
	Img         string `json:"img"`
	Name        string `json:"name"`
	Info        string `json:"info"`
	Nickname    string `json:"nickname"`
	Friends     int    `json:"friends"`
	Subscribers int    `json:"subscribers"`
}

type ChatResponseWithFlags struct {
	ID          int    `json:"id"`
	Img         string `json:"img"`
	Name        string `json:"name"`
	Info        string `json:"info"`
	Nickname    string `json:"nickname"`
	Friends     int    `json:"friends"`
	Subscribers int    `json:"subscribers"`
	Sound       bool   `json:"sound"`
	IsRead      bool   `json:"is_read"`
}
