package requestbody

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/tools"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/gin-gonic/gin"
	"io"
	"os"
)

type Conversation struct {
	GinContext     *gin.Context
	RequestUrl     string
	RequestClient  tls_client.HttpClient
	RequestBody    io.ReadCloser
	RequestParam   string
	RequestMethod  string
	RequestHeaders http.Header
}

func init() {
	if os.Getenv("OPENAI_HOST") == "" {
		OpenAI_HOST = "chat.openai.com"
	} else {
		OpenAI_HOST = os.Getenv("OPENAI_HOST")
	}
}

var (
	jar     = tls_client.NewCookieJar()
	options = []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(360),
		tls_client.WithClientProfile(profiles.Safari_15_6_1),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar), // create cookieJar instance and pass it as argument
	}
	client, _   = tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	user_agent  = common.Env.UserAgent
	http_proxy  = common.Env.Proxy
	OpenAI_HOST = common.Env.OpenAI_HOST
)
var converstion Conversation

func SetConversation(ctx *gin.Context) {
	converstion.GinContext = ctx
	converstion.RequestUrl = common.CheckRequest(ctx).Generate(ctx.Param("path"), ctx.Request.URL.RawQuery)
	converstion.RequestMethod = ctx.Request.Method
	converstion.RequestBody = ctx.Request.Body
	converstion.RequestParam = ctx.Param("path")
	converstion.RequestClient = tools.GetHttpClient()
	converstion.RequestHeaders = http.Header(ctx.Request.Header)
}

func GetConversation() Conversation {
	return converstion
}
