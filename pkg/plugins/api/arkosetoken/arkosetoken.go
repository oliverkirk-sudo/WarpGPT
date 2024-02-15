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
var ArkoseTokenInstance ArkoseToken

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
	id := p.GetContext().GinContext.Param("id")
	var (
		token string
		err   error
	)
	if id == "35536E1E-65B4-4D96-9D97-6ADB7EFF8147" {
		token, err = funcaptcha.GetOpenAIArkoseToken(4, p.GetContext().RequestHeaders.Get("puid"))
	} else if id == "0A1D34FC-659D-4E23-B17B-694DCFCF6A6C" {
		token, err = funcaptcha.GetOpenAIArkoseToken(0, p.GetContext().RequestHeaders.Get("puid"))
	} else if id == "3D86FBBA-9D22-402A-B512-3420086BA6CC" {
		token, err = funcaptcha.GetOpenAIArkoseToken(3, p.GetContext().RequestHeaders.Get("puid"))
	} else {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Invalid id"})
		return
	}
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
func (p *ArkoseToken) Run(com *plugins.Component) {
	context = com
	context.Engine.GET("/token/:id", func(c *gin.Context) {
		conversation := common.GetContextPack(c, NotHaveUrl{})
		common.Do[Context](new(ArkoseToken), Context(conversation))
	})
}
