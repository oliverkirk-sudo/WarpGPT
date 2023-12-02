package main

import (
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/process/api"
	"WarpGPT/pkg/requestbody"
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	var router = gin.Default()
	router.Any("/v1/*path", func(c *gin.Context) {
		conversation := requestbody.GetConversation(c)
		fmt.Printf("%+v\n", conversation)
		var p api.OfficialApiProcess
		p.SetConversation(conversation)
		process.Do(&p)

	})
	router.Run()
}
