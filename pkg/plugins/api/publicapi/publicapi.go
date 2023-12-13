package publicapi

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/plugins"
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/gin-gonic/gin"
	"io"
	shttp "net/http"
	"strings"
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
type PublicApiProcess struct {
	Context Context
}

func (p *PublicApiProcess) SetContext(conversation Context) {
	p.Context = conversation
}
func (p *PublicApiProcess) GetContext() Context {
	return p.Context
}
func (p *PublicApiProcess) ProcessMethod() {
	context.Logger.Debug("PublicApiProcess")
	var requestBody map[string]interface{}
	err := p.decodeRequestBody(&requestBody) //解析请求体
	if err != nil {
		return
	}
	request, err := p.createRequest(requestBody) //创建请求
	if err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Server error"})
		return
	}
	response, err := p.GetContext().RequestClient.Do(request) //发送请求
	if err != nil {
		var jsonData interface{}
		err := json.NewDecoder(response.Body).Decode(&jsonData)
		if err != nil {
			p.GetContext().GinContext.JSON(500, gin.H{"error": "Request json decode error"})
			return
		}
		p.GetContext().GinContext.JSON(response.StatusCode, jsonData)
		return
	}
	if strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		err := p.jsonResponse(response)
		if err != nil {
			context.Logger.Fatal(err)
		}
	}
	common.CopyResponseHeaders(response, p.GetContext().GinContext) //设置响应头
}
func (p *PublicApiProcess) createRequest(requestBody map[string]interface{}) (*http.Request, error) {
	context.Logger.Debug("PublicApiProcess createRequest")
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(bodyBytes)
	var request *http.Request
	if p.Context.RequestBody == shttp.NoBody {
		request, err = http.NewRequest(p.Context.RequestMethod, p.Context.RequestUrl, nil)
	} else {
		request, err = http.NewRequest(p.Context.RequestMethod, p.Context.RequestUrl, bodyReader)
	}
	if err != nil {
		return nil, err
	}
	p.buildHeaders(request)
	p.setCookies(request)
	return request, nil
}
func (p *PublicApiProcess) setCookies(request *http.Request) {
	context.Logger.Debug("PublicApiProcess setCookies")
	for _, cookie := range p.GetContext().GinContext.Request.Cookies() {
		request.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}
}
func (p *PublicApiProcess) buildHeaders(request *http.Request) {
	context.Logger.Debug("PublicApiProcess buildHeaders")
	headers := map[string]string{
		"Host":          context.Env.OpenaiHost,
		"Origin":        "https://" + context.Env.OpenaiHost + "/chat",
		"Authorization": p.GetContext().GinContext.Request.Header.Get("Authorization"),
		"Connection":    "keep-alive",
		"User-Agent":    context.Env.UserAgent,
		"Content-Type":  p.GetContext().GinContext.Request.Header.Get("Content-Type"),
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	if puid := p.GetContext().GinContext.Request.Header.Get("PUID"); puid != "" {
		request.Header.Set("cookie", "_puid="+puid+";")
	}
}
func (p *PublicApiProcess) jsonResponse(response *http.Response) error {
	context.Logger.Debug("PublicApiProcess jsonResponse")
	var jsonData interface{}
	err := json.NewDecoder(response.Body).Decode(&jsonData)
	if err != nil {
		return err
	}
	p.GetContext().GinContext.JSON(response.StatusCode, jsonData)
	return nil
}
func (p *PublicApiProcess) decodeRequestBody(requestBody *map[string]interface{}) error {
	conversation := p.GetContext()
	if conversation.RequestBody != shttp.NoBody {
		if err := json.NewDecoder(conversation.RequestBody).Decode(requestBody); err != nil {
			conversation.GinContext.JSON(400, gin.H{"error": "JSON invalid"})
			return err
		}
	}
	return nil
}

type ReversePublicApiRequestUrl struct {
}

func (u ReversePublicApiRequestUrl) Generate(path string, rawquery string) string {
	if rawquery == "" {
		return "https://" + context.Env.OpenaiHost + "/public-api" + path
	}
	return "https://" + context.Env.OpenaiHost + "/public-api" + path + "?" + rawquery
}

func Run(com *plugins.Component) {
	context = com
	context.Engine.Any("/public-api/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c, ReversePublicApiRequestUrl{})
		p := new(PublicApiProcess)
		common.Do[Context](p, Context(conversation))
	})
}
