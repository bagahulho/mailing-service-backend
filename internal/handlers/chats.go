package handlers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Chat struct {
	ID             int
	Img            string
	Name           string
	Info           string
	Messages       []string
	LastMessageInd int
	Nickname       string
	Friends        int
	Subscribers    int
}

type Recipient struct {
	ID            int
	ApplicationID int
	ChatID        int
	Img           string
	Name          string
	Info          string
}

type Request struct {
	ID        int
	Recipient []Recipient
}

var chats []Chat = []Chat{
	{
		ID:   1,
		Img:  "http://127.0.0.1:9000/test/1.png",
		Name: "Леонардо Ди Каприо",
		Info: "актёр, продюсер и активист, известный своими выдающимися ролями и вкладом в защиту окружающей среды.",
		Messages: []string{
			"Привет, Лео! Как у тебя дела?",
			"Только что пересматривал 'Выжившего'. Впечатляюще!",
			"У тебя есть планы на новые фильмы?",
			"Как тебе удаётся поддерживать такую активную карьеру?",
			"Что для тебя самое сложное в актёрской работе?",
			"Есть ли роли, которые ты жалеешь, что не сыграл?",
		},
		LastMessageInd: 5,
		Nickname:       "@leo_di",
		Friends:        500,
		Subscribers:    9213499,
	},
	{
		ID:   2,
		Img:  "http://127.0.0.1:9000/test/2.png",
		Name: "Хью Джекман",
		Info: "Известен своей ролью Росомахи, также проявил себя в театре и кино",
		Messages: []string{
			"Привет, Хью! Как поживаешь?",
			"Смотрел 'Величайший шоумен' — просто невероятно!",
			"Какие у тебя следующие проекты?",
			"Как удаётся балансировать между кино и театром?",
			"Ты скучаешь по роли Росомахи?",
			"Что для тебя значит быть частью франшизы 'Люди Икс'?",
			"Есть ли роль, которую ты мечтаешь сыграть?",
			"Как ты готовишься к физически сложным ролям?",
			"Какой твой самый запоминающийся момент на сцене?",
			"Что ты любишь делать в свободное время?",
		},
		LastMessageInd: 9,
		Nickname:       "@hyu_jack",
		Friends:        123,
		Subscribers:    882378,
	},
	{
		ID:   3,
		Img:  "http://127.0.0.1:9000/test/3.png",
		Name: "Магнус Карлсен",
		Info: "Норвежский гроссмейстер, ставший чемпионом мира по шахматам",
		Messages: []string{
			"Привет, Магнус! Как твои шахматные тренировки?",
			"Недавно смотрел твои партии — впечатляющая игра!",
			"Что для тебя самое сложное в шахматах?",
			"Есть ли стратегия, которой ты придерживаешься чаще всего?",
			"Как ты справляешься с давлением на крупных турнирах?",
			"Кто был твоим самым сложным соперником?",
			"Какие цели ты ставишь перед собой сейчас?",
			"Есть ли какой-то совет для начинающих шахматистов?",
		},
		LastMessageInd: 7,
		Nickname:       "@magnusK",
		Friends:        1,
		Subscribers:    321421,
	},
	{
		ID:   4,
		Img:  "http://127.0.0.1:9000/test/4.png",
		Name: "Антон Канев",
		Info: "Гений, миллиардер, плейбой, филантроп и говорит (по мелочи) на китайском",
		Messages: []string{
			"Здравствуйте, поставьте пж отл",
			"Докончил правки",
			"Можно на консультацию прийти?",
		},
		LastMessageInd: 2,
		Nickname:       "@neughost",
		Friends:        265,
		Subscribers:    8,
	},
}

var mockRequest Request = Request{
	ID: 1,
	Recipient: []Recipient{
		{
			ID:            1,
			ApplicationID: 1,
			ChatID:        1,
			Img:           "http://127.0.0.1:9000/test/1.png",
			Name:          "Леонардо Ди Каприо",
			Info:          "Актёр, продюсер и активист, известный своими выдающимися ролями и вкладом в защиту окружающей среды.",
		},
		{
			ID:            2,
			ApplicationID: 1,
			ChatID:        4,
			Img:           "http://127.0.0.1:9000/test/4.png",
			Name:          "Антон Канев",
			Info:          "Гений, миллиардер, плейбой, филантроп и говорит (по мелочи) на китайском",
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
		"data":  result,
		"req":   mockRequest,
		"query": query,
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

func RequestHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "request.page.tmpl", mockRequest)
}
