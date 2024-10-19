package handler

//
//import (
//	"RIP/internal/app/ds"
//	"github.com/gin-gonic/gin"
//	"net/http"
//	"strconv"
//	"strings"
//)
//
//func (h *Handler) GetAllChats(ctx *gin.Context) {
//	var chats []ds.Chat
//	var err error
//
//	search := ctx.Query("search")
//	if search == "" {
//		chats, err = h.Repository.GetAllChats()
//	} else {
//		chats, err = h.Repository.SearchChatsByName(search)
//	}
//
//	if err != nil {
//		ctx.JSON(http.StatusInternalServerError, gin.H{
//			"error": err.Error(),
//		})
//		return
//	}
//
//	ctx.HTML(http.StatusOK, "chats.page.tmpl", gin.H{
//		"data":       chats,
//		"cart_count": h.Repository.GetCartCount(),
//		"search":     search,
//		"draft_id":   h.Repository.GetDraftID(),
//	})
//}
//
//func (h *Handler) GetChatById(ctx *gin.Context) {
//	strId := ctx.Param("id")
//	id, err := strconv.Atoi(strId)
//	if err != nil {
//		ctx.JSON(http.StatusInternalServerError, gin.H{
//			"error": err.Error(),
//		})
//	}
//
//	chat, err := h.Repository.GetChatByID(id)
//	if err != nil {
//		ctx.JSON(http.StatusInternalServerError, gin.H{
//			"error": err.Error(),
//		})
//	}
//
//	ctx.HTML(http.StatusOK, "chat.page.tmpl", chat)
//}
//
//func (h *Handler) GetMessage(ctx *gin.Context) {
//	strId := ctx.Param("id")
//	id, err := strconv.Atoi(strId)
//	if err != nil {
//		ctx.JSON(http.StatusInternalServerError, gin.H{
//			"error": err.Error(),
//		})
//		return
//	}
//
//	Message, Chats, err := h.Repository.GetMessage(uint(id))
//	if err != nil {
//		ctx.Redirect(http.StatusFound, "/chats")
//		return
//	}
//
//	// Рендерим страницу, если нет ошибок
//	ctx.HTML(http.StatusOK, "message.page.tmpl", gin.H{
//		"message_text": Message.Text.String,
//		"chats":        Chats,
//	})
//}
//
//func (h *Handler) AddChatToList(ctx *gin.Context) {
//	strId := ctx.PostForm("chat_id")
//	id, err := strconv.Atoi(strId)
//	if err != nil {
//		ctx.JSON(http.StatusInternalServerError, gin.H{
//			"error": err.Error(),
//		})
//	}
//	// Вызов функции добавления чата в заявку
//	err = h.Repository.AddChatToList(uint(id))
//	if err != nil && !strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
//		return
//	}
//
//	ctx.Redirect(http.StatusFound, "/chats")
//}
//
//func (h *Handler) DeleteList(ctx *gin.Context) {
//	err := h.Repository.DeleteList()
//	if err != nil {
//		ctx.JSON(http.StatusInternalServerError, gin.H{
//			"error": err.Error(),
//		})
//	}
//
//	ctx.Redirect(http.StatusSeeOther, "/chats")
//}
