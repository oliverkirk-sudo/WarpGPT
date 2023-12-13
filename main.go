package main

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/env"
	"WarpGPT/pkg/plugins/arkosetoken"
	"WarpGPT/pkg/plugins/backendapi"
	"WarpGPT/pkg/plugins/officialapi"
	"WarpGPT/pkg/plugins/publicapi"
	session2 "WarpGPT/pkg/plugins/session"
	"WarpGPT/pkg/plugins/unofficialapi"
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
		if apiKey != env.Env.AuthKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
func main() {
	var router = gin.Default()
	if env.Env.Verify {
		router.Use(AuthMiddleware())
	}
	router.Use(CORSMiddleware())
	router.Any("/v1/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(officialapi.OfficialApiProcess)
		common.Do(p, conversation)

	})
	router.Any("/backend-api/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(backendapi.BackendProcess)
		common.Do(p, conversation)
	})
	router.Any("/api/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(publicapi.PublicApiProcess)
		common.Do(p, conversation)
	})
	router.Any("/public-api/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(publicapi.PublicApiProcess)
		common.Do(p, conversation)
	})
	router.GET("/token", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(arkosetoken.ArkoseToken)
		common.Do(p, conversation)
	})
	router.POST("/getsession", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(session2.SessionToken)
		common.Do(p, conversation)
	})
	router.Any("/r/v1/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c)
		p := new(unofficialapi.UnofficialApiProcess)
		common.Do(p, conversation)
	})
	go proxypool.ProxyThread()
	router.Run(env.Env.Host + ":" + strconv.Itoa(env.Env.Port))
}
