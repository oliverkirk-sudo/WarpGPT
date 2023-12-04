package main

import (
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/process/api"
	"WarpGPT/pkg/process/reverse"
	"WarpGPT/pkg/process/session"
	"WarpGPT/pkg/requestbody"
	"github.com/gin-gonic/gin"
)

func main() {
	var router = gin.Default()
	router.Any("/v1/*path", func(c *gin.Context) {
		conversation := requestbody.GetConversation(c)
		var p api.OfficialApiProcess
		process.Do(&p, conversation)

	})
	router.Any("/backend-api/*path", func(c *gin.Context) {
		conversation := requestbody.GetConversation(c)
		var p reverse.BackendProcess
		process.Do(&p, conversation)
	})
	router.Any("/api/*path", func(c *gin.Context) {
		conversation := requestbody.GetConversation(c)
		var p reverse.PublicApiProcess
		process.Do(&p, conversation)
	})
	router.Any("/public-api/*path", func(c *gin.Context) {
		conversation := requestbody.GetConversation(c)
		var p reverse.PublicApiProcess
		process.Do(&p, conversation)
	})
	router.GET("/token", func(c *gin.Context) {
		conversation := requestbody.GetConversation(c)
		var p session.ArkoseToken
		process.Do(&p, conversation)
	})
	router.POST("/getsession", func(c *gin.Context) {
		conversation := requestbody.GetConversation(c)
		var p session.SessionToken
		process.Do(&p, conversation)
	})
	router.Any("/r/v1/*path", func(c *gin.Context) {
		conversation := requestbody.GetConversation(c)
		var p api.UnofficialApiProcess
		process.Do(&p, conversation)
	})
	router.Run()
}
