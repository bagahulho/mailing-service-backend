package api

import (
	"RIP/internal/handlers"
	"github.com/gin-gonic/gin"
	"log"
)

func StartServer() {
	log.Println("Starting server")

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.GET("/chats", handlers.ChatsHandle)
	r.GET("/sending", handlers.SendingHandle)
	r.GET("/chats/:id", handlers.ChatHandle)

	r.Static("/styles", "./styles")

	err := r.Run()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Server down")
}
