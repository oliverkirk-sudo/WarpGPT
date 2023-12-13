package arkosetoken

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/funcaptcha"
	"WarpGPT/pkg/plugins"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/gin-gonic/gin"
	"io"
)

var context *plugins.Component

type Context struct {
	GinContext     *gin.Context
	RequestUrl     string
	RequestClient  tls_client.HttpClient
	RequestBody    io.ReadCloser
	RequestParam   string
	RequestMethod  string
	RequestHeaders http.Header
}
type ArkoseToken struct {
	Context Context
}

func (p *ArkoseToken) GetContext() Context {
	return p.Context
}
func (p *ArkoseToken) SetContext(conversation Context) {
	p.Context = conversation
}

func (p *ArkoseToken) ProcessMethod() {
	context.Logger.Debug("ArkoseToken")
	token, err := funcaptcha.GetOpenAIArkoseToken(4, p.GetContext().RequestHeaders.Get("puid"))
	if err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Unable to generate ArkoseToken"})
	}
	p.GetContext().GinContext.Header("Content-Type", "application/json")
	p.GetContext().GinContext.JSON(200, gin.H{"token": token})
}

type NotHaveUrl struct {
}

func (u NotHaveUrl) Generate(path string, rawquery string) string {
	return ""
}
func Run(com *plugins.Component) {
	context = com
	context.Engine.Any("/token", func(c *gin.Context) {
		conversation := common.GetContextPack(c, NotHaveUrl{})
		p := new(ArkoseToken)
		common.Do[Context](p, Context(conversation))
	})
}
