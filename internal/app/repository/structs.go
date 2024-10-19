package repository

import "time"

type Chat struct {
	ID          int
	Img         string
	Name        string
	Info        string
	Nickname    string
	Friends     int
	Subscribers int
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
	ID         uint      `json:"id"`
	Status     string    `json:"status"`
	Text       string    `json:"text,omitempty"`
	DateCreate time.Time `json:"date_create"`
	DateUpdate time.Time `json:"date_update,omitempty"`
	DateFinish time.Time `json:"date_finish,omitempty"`
	Creator    string    `json:"creator"`
	Moderator  string    `json:"moderator,omitempty"`
	Chats      []Chat    `json:"chats"`
}
