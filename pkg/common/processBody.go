package common

import (
	"WarpGPT/pkg/logger"
	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
)

type ContextProcessor[T any] interface {
	SetContext(conversation T)
	GetContext() T
	ProcessMethod()
}

func Do[T any](p ContextProcessor[T], conversation T) {
	p.SetContext(conversation)
	p.ProcessMethod()
}

func CopyResponseHeaders(response *fhttp.Response, ctx *gin.Context) {
	logger.Log.Debug("CopyResponseHeaders")
	if response == nil {
		logger.Log.Warning("response is empty")
		ctx.JSON(400, gin.H{"error": "response is empty"})
		return
	}
	skipHeaders := map[string]bool{
		"content-encoding":true,
		"content-length":true,
		"transfer-encoding":true,
		"connection":true,
	}
	for name, values := range response.Header {
		if !skipHeaders[name] {
			for _, value := range values {
				ctx.Writer.Header().Add(name, value)
			}
		}
	}
}
