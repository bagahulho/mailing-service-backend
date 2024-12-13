package handler

import (
	"RIP/internal/app/repository"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		repository: r,
	}
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//router.GET("/chats", h.GetAllChats)
	//router.GET("/chats/:id", h.GetChatById)
	//router.GET("/message/:id", h.GetMessage)
	//router.POST("/add-chat", h.AddChatToList)
	//router.POST("/list/delete", h.DeleteList)
	protected := router.Group("/")
	protected.Use(h.AuthMiddleware)
	{
		protected.POST("/chat/:id/in-message", h.AddChatToMessage)

		protected.GET("/messages", h.GetMessagesFiltered)
		protected.GET("/messages/:id", h.GetMessage)
		protected.PUT("/messages/:id/text", h.UpdateMessageText)
		protected.PUT("/messages/:id/form", h.MessageForm)
		protected.DELETE("/messages/:id/delete", h.MessageDelete)

		protected.DELETE("/message-chats/delete/:message_id/:chat_id", h.DeleteChatFromMessage)
		protected.PUT("/message-chats/switch/:message_id/:chat_id", h.ToggleSoundField)

		protected.POST("/chats/create", h.ModeratorMiddleware, h.CreateChat)
		protected.PUT("/chats/:id", h.ModeratorMiddleware, h.UpdateChat)
		protected.POST("/chats/:id/new-image", h.ModeratorMiddleware, h.ReplaceChatImage)
		protected.DELETE("/chats/:id", h.DeleteChat)

		protected.PUT("/messages/:id/finish", h.ModeratorMiddleware, h.MessageFinish)
		protected.PUT("/messages/:id/reject", h.ModeratorMiddleware, h.MessageReject)

		protected.POST("/user/logout", h.Logout)
		protected.PUT("/user/update", h.UpdateUser)
	}
	router.GET("/chats", h.GetChats)
	router.GET("/chats/:id", h.GetChatByID)

	router.POST("/user/reg", h.RegisterUser)
	router.POST("/user/login", h.Authenticate)
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
