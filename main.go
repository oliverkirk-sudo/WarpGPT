package main

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/process/api"
	"WarpGPT/pkg/process/reverse"
	"WarpGPT/pkg/process/session"
	"WarpGPT/pkg/proxypool"
	"github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	"strconv"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("AuthKey")
		if apiKey != common.Env.AuthKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
func main() {
	var router = gin.Default()
	if common.Env.Verify {
		router.Use(AuthMiddleware())
	}
	router.Use(CORSMiddleware())
	router.Any("/v1/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(api.OfficialApiProcess)
		process.Do(p, conversation)

	})
	router.Any("/backend-api/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(reverse.BackendProcess)
		process.Do(p, conversation)
	})
	router.Any("/api/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(reverse.PublicApiProcess)
		process.Do(p, conversation)
	})
	router.Any("/public-api/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(reverse.PublicApiProcess)
		process.Do(p, conversation)
	})
	router.GET("/token", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(session.ArkoseToken)
		process.Do(p, conversation)
	})
	router.POST("/getsession", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(session.SessionToken)
		process.Do(p, conversation)
	})
	router.Any("/r/v1/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(api.UnofficialApiProcess)
		process.Do(p, conversation)
	})
	go proxypool.ProxyThread()
	router.Run(common.Env.Host + ":" + strconv.Itoa(common.Env.Port))
}
