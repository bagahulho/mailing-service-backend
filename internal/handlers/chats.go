package handlers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Chat struct {
	ID          int
	Img         string
	Name        string
	Info        string
	Nickname    string
	Friends     int
	Subscribers int
}

type Recipient struct {
	ChatID int
	Sound  int
}

type Message struct {
	ID        int
	Text      string
	Date      string
	Recipient []Recipient
}

var chats []Chat = []Chat{
	{
		ID:          1,
		Img:         "http://127.0.0.1:9000/test/1.png",
		Name:        "Леонардо Ди Каприо",
		Info:        "актёр, продюсер и активист, известный своими выдающимися ролями и вкладом в защиту окружающей среды.",
		Nickname:    "@leo_di",
		Friends:     500,
		Subscribers: 9213499,
	},
	{
		ID:          2,
		Img:         "http://127.0.0.1:9000/test/2.png",
		Name:        "Хью Джекман",
		Info:        "Известен своей ролью Росомахи, также проявил себя в театре и кино",
		Nickname:    "@hyu_jack",
		Friends:     123,
		Subscribers: 882378,
	},
	{
		ID:          3,
		Img:         "http://127.0.0.1:9000/test/3.png",
		Name:        "Магнус Карлсен",
		Info:        "Норвежский гроссмейстер, ставший чемпионом мира по шахматам",
		Nickname:    "@magnusK",
		Friends:     1,
		Subscribers: 321421,
	},
	{
		ID:          4,
		Img:         "http://127.0.0.1:9000/test/4.png",
		Name:        "Антон Канев",
		Info:        "Гений, миллиардер, плейбой, филантроп и говорит (по мелочи) на китайском",
		Nickname:    "@neughost",
		Friends:     265,
		Subscribers: 8,
	},
}

var mockMessage Message = Message{
	ID:   1,
	Text: "It`s first message",
	Date: "19.06.2024",
	Recipient: []Recipient{
		{
			ChatID: 1,
			Sound:  1,
		},
		{
			ChatID: 4,
			Sound:  0,
		},
	},
}

func ChatsHandle(c *gin.Context) {
	// Получаем параметр query из GET-запроса
	query := c.Query("query")
	// это реализация поиска!
	result := make([]Chat, 0)
	if query != "" {
		for _, m := range chats {
			if strings.Contains(strings.ToLower(m.Name), strings.ToLower(query)) {
				result = append(result, m)
			}
		}
	} else {
		result = chats
	}

	c.HTML(http.StatusOK, "chats.page.tmpl", gin.H{
		"data":        result,
		"message-len": len(mockMessage.Recipient),
		"query":       query,
	})
}

func ChatHandle(c *gin.Context) {
	idStr := c.Param("id") // Получаем параметр id из маршрута
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Fatal(err)
	}

	ind := 0
	for i, chat := range chats {
		if chat.ID == id {
			ind = i
		}
	}

	c.HTML(http.StatusOK, "chat.page.tmpl", chats[ind])
}

func SendingHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "sending.page.tmpl", mockMessage)
}
