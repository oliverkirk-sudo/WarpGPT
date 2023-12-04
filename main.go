package main

import (
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/process/api"
	"WarpGPT/pkg/process/reverse"
	"WarpGPT/pkg/process/session"
	"WarpGPT/pkg/requestbody"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
func main() {
	var router = gin.Default()
	router.Use(CORSMiddleware())
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
	router.OPTIONS("/r/v1/*path", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})
	router.POST("/r/v1/*path", func(c *gin.Context) {
		conversation := requestbody.GetConversation(c)
		var p api.UnofficialApiProcess
		process.Do(&p, conversation)
	})
	router.GET("/r/v1/*path", func(c *gin.Context) {
		conversation := requestbody.GetConversation(c)
		var p api.UnofficialApiProcess
		process.Do(&p, conversation)
	})
	router.Run()
}
