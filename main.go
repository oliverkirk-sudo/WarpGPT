package main

import (
	"WarpGPT/pkg/db"
	"WarpGPT/pkg/env"
	"WarpGPT/pkg/funcaptcha"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/plugins"
	"WarpGPT/pkg/plugins/api/arkosetoken"
	"WarpGPT/pkg/plugins/api/backendapi"
	"WarpGPT/pkg/plugins/api/officialapi"
	"WarpGPT/pkg/plugins/api/publicapi"
	"WarpGPT/pkg/plugins/api/rapi"
	"WarpGPT/pkg/plugins/api/session"
	"WarpGPT/pkg/plugins/api/unofficialapi"
	"WarpGPT/pkg/plugins/service/proxypool"
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
	component := &plugins.Component{
		Engine: router,
		Db:     db.DB{GetRedisClient: db.RedisDB{}.GetRedisClient},
		Logger: logger.Log,
		Env:    &env.Env,
		Auth:   funcaptcha.GetOpenAIArkoseToken,
	}
	arkosetoken.Run(component)
	session.Run(component)
	backendapi.Run(component)
	officialapi.Run(component)
	unofficialapi.Run(component)
	publicapi.Run(component)
	rapi.Run(component)
	proxypool.Run(component)

	router.Run(env.Env.Host + ":" + strconv.Itoa(env.Env.Port))
}
