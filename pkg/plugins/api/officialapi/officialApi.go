package officialapi

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/plugins"
	"WarpGPT/pkg/tools"
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/gin-gonic/gin"
	"io"
	fhttp "net/http"
	shttp "net/http"
	"strings"
)

var context *plugins.Component
var OfficialApiProcessInstance OfficialApiProcess

type Context struct {
	GinContext     *gin.Context
	RequestUrl     string
	RequestClient  tls_client.HttpClient
	RequestBody    io.ReadCloser
	RequestParam   string
	RequestMethod  string
	RequestHeaders http.Header
}
type OfficialApiProcess struct {
	Context Context
}

func (p *OfficialApiProcess) SetContext(conversation Context) {
	p.Context = conversation
}
func (p *OfficialApiProcess) GetContext() Context {
	return p.Context
}
func (p *OfficialApiProcess) ProcessMethod() {
	context.Logger.Debug("officialApi")
	var requestBody map[string]interface{}
	err := p.decodeRequestBody(&requestBody) //解析请求体
	if err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Incorrect json format"})
		return
	}

	request, err := p.createRequest(requestBody) //创建请求
	if err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Server error"})
		return
	}

	response, err := p.GetContext().RequestClient.Do(request) //发送请求
	if err != nil {
		context.Logger.Error(err)
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Server Error"})
		return
	}

	common.CopyResponseHeaders(response, p.GetContext().GinContext) //设置响应头

	if strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		err = p.streamResponse(response)
		if err != nil {
			return
		}
	}
	if strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		err = p.jsonResponse(response)
		if err != nil {
			context.Logger.Warning(err)
		}
	}
}

func (p *OfficialApiProcess) createRequest(requestBody map[string]interface{}) (*http.Request, error) {
	context.Logger.Debug("officialApi createRequest")
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	var request *http.Request
	if p.Context.RequestBody == shttp.NoBody {
		request, err = http.NewRequest(p.Context.RequestMethod, p.Context.RequestUrl, nil)
	} else {
		request, err = http.NewRequest(p.Context.RequestMethod, p.Context.RequestUrl, bytes.NewBuffer(bodyBytes))
	}
	if err != nil {
		return nil, err
	}
	p.WithHeaders(request)
	return request, nil
}

func (p *OfficialApiProcess) WithHeaders(rsq *http.Request) {
	rsq.Header.Set("Authorization", p.Context.RequestHeaders.Get("Authorization"))
	rsq.Header.Set("Content-Type", p.Context.RequestHeaders.Get("Content-Type"))
}

func (p *OfficialApiProcess) jsonResponse(response *http.Response) error {
	context.Logger.Debug("officialApi jsonResponse")
	var jsonData interface{}
	err := json.NewDecoder(response.Body).Decode(&jsonData)
	if err != nil {
		return err
	}
	p.GetContext().GinContext.JSON(response.StatusCode, jsonData)
	return nil
}

func (p *OfficialApiProcess) streamResponse(response *http.Response) error {
	context.Logger.Debug("officialApi streamResponse")
	context.Logger.Infoln("officialApiProcess stream Request")
	client := tools.NewSSEClient(response.Body)
	events := client.Read()
	for event := range events {
		if _, err := p.GetContext().GinContext.Writer.Write([]byte("data: " + event.Data + "\n\n")); err != nil {
			return err
		}
		p.GetContext().GinContext.Writer.Flush()
	}
	defer client.Close()
	return nil
}

type OfficialApiRequestUrl struct {
}

func (u OfficialApiRequestUrl) Generate(path string, rawquery string) string {
	if rawquery == "" {
		return "https://" + context.Env.OpenaiApiHost + "/v1" + path
	}
	return "https://" + context.Env.OpenaiApiHost + "/v1" + path + "?" + rawquery
}
func (p *OfficialApiProcess) decodeRequestBody(requestBody *map[string]interface{}) error {
	context.Logger.Debug("officialApi decodeRequestBody")
	conversation := p.GetContext()
	if conversation.RequestBody != fhttp.NoBody {
		if err := json.NewDecoder(conversation.RequestBody).Decode(requestBody); err != nil {
			conversation.GinContext.JSON(400, gin.H{"error": "JSON invalid"})
			return err
		}
	}
	return nil
}

func (p *OfficialApiProcess) Run(com *plugins.Component) {
	context = com
	context.Engine.Any("/v1/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c, OfficialApiRequestUrl{})
		common.Do[Context](new(OfficialApiProcess), Context(conversation))
	})
}
