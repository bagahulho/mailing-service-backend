package handler

import (
	"RIP/internal/app/repository"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	//router.GET("/chats", h.GetAllChats)
	//router.GET("/chats/:id", h.GetChatById)
	//router.GET("/message/:id", h.GetMessage)
	//router.POST("/add-chat", h.AddChatToList)
	//router.POST("/list/delete", h.DeleteList)

	router.GET("/chats", h.GetChats)
	router.GET("/chats/:id", h.GetChatByID)
	router.POST("/chats/create", h.CreateChat)
	router.PUT("/chats/:id", h.UpdateChat)
	router.DELETE("/chats/:id", h.DeleteChat)
	router.POST("/chat-to-message/:id", h.AddChatToMessage)
	router.POST("/chats/:id/new-image", h.ReplaceChatImage)

	router.GET("/messages", h.GetMessagesFiltered)
	router.GET("/messages/:id", h.GetMessage)
	router.PUT("/messages/:id/text", h.UpdateMessageText)
	router.PUT("/messages/:id/form", h.MessageForm)
	router.PUT("/messages/:id/finish", h.MessageFinish)
	router.PUT("/messages/:id/reject", h.MessageReject)
	router.DELETE("/messages/:id/delete", h.MessageDelete)

	router.DELETE("/message-chats/delete/:message_id/:chat_id", h.DeleteChatFromMessage)
	router.PUT("/message-chats/switch/:message_id/:chat_id", h.ToggleSoundField)

	router.POST("/user/reg", h.CreateUser)
	router.PUT("/user/update/:id", h.UpdateUser)
	router.POST("/user/auth", h.AuthUser)
	router.POST("/user/de-auth", h.DeAuthUser)
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*")
	router.Static("/styles", "./styles")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, description string) {
	logrus.Error(description)
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "fail",
		"description": description,
	})
}
