package common

import (
	"WarpGPT/pkg/tools"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/gin-gonic/gin"
	"io"
)

type Context struct {
	GinContext     *gin.Context
	RequestUrl     string
	RequestClient  tls_client.HttpClient
	RequestBody    io.ReadCloser
	RequestParam   string
	RequestMethod  string
	RequestHeaders http.Header
}

func GetContextPack(ctx *gin.Context) Context {
	conversation := Context{}
	conversation.GinContext = ctx
	conversation.RequestUrl = CheckRequest(ctx).Generate(ctx.Param("path"), ctx.Request.URL.RawQuery)
	conversation.RequestMethod = ctx.Request.Method
	conversation.RequestBody = ctx.Request.Body
	conversation.RequestParam = ctx.Param("path")
	conversation.RequestClient = tools.GetHttpClient()
	conversation.RequestHeaders = http.Header(ctx.Request.Header)
	return conversation
}
