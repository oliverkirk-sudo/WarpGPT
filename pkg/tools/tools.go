package tools

import (
	"WarpGPT/pkg/common"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/gin-gonic/gin"
	"log"
)

var OpenAI_HOST string

const user_agent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Safari/605.1.15"

func BuildHeaders(c *gin.Context, request *http.Request) {
	log.Println("build_headers")
	request.Header.Set("Host", ""+OpenAI_HOST+"")
	request.Header.Set("Origin", "https://"+OpenAI_HOST+"/chat")
	request.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
	request.Header.Set("user-agent", user_agent)
	if c.Request.Header.Get("PUID") != "" {
		request.Header.Set("cookie", "_puid="+c.Request.Header.Get("PUID")+";")
	}
}
func GetHttpClient() tls_client.HttpClient {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(120),
		tls_client.WithClientProfile(profiles.Safari_15_6_1),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
		tls_client.WithProxyUrl(common.Env.Proxy),
	}
	client, _ := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	return client
}
