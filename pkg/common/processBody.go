package common

import (
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
	skipHeaders := map[string]bool{"Content-Encoding": true, "Content-Length": true, "transfer-encoding": true, "connection": true}
	for name, values := range response.Header {
		if !skipHeaders[name] {
			for _, value := range values {
				ctx.Header(name, value)
			}
		}
	}
}
