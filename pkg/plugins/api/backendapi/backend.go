package backendapi

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/funcaptcha"
	"WarpGPT/pkg/plugins"
	"WarpGPT/pkg/tools"
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
var BackendProcessInstance BackendProcess

type Context struct {
	GinContext     *gin.Context
	RequestUrl     string
	RequestClient  tls_client.HttpClient
	RequestBody    io.ReadCloser
	RequestParam   string
	RequestMethod  string
	RequestHeaders http.Header
}
type BackendProcess struct {
	Context Context
}

func (p *BackendProcess) GetContext() Context {
	return p.Context
}
func (p *BackendProcess) SetContext(conversation Context) {
	p.Context = conversation
}

func (p *BackendProcess) ProcessMethod() {
	context.Logger.Debug("ProcessBackendProcess")
	var requestBody map[string]interface{}
	err := p.decodeRequestBody(&requestBody)
	if err != nil {
		return
	}
	request, err := p.createRequest(requestBody)
	if err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Server error"})
		return
	}

	response, err := p.GetContext().RequestClient.Do(request)
	if err != nil {
		var jsonData interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonData)
		if err != nil {
			p.GetContext().GinContext.JSON(500, gin.H{"error": "Request json decode error"})
			return
		}
		p.GetContext().GinContext.JSON(response.StatusCode, jsonData)
		return
	}

	common.CopyResponseHeaders(response, p.GetContext().GinContext)

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
func (p *BackendProcess) createRequest(requestBody map[string]interface{}) (*http.Request, error) {
	context.Logger.Debug("BackendProcess createRequest")
	var request *http.Request
	if p.Context.RequestBody == shttp.NoBody {
		request, _ = http.NewRequest(p.Context.RequestMethod, p.Context.RequestUrl, nil)
	} else {
		token, err := p.addArkoseTokenIfNeeded(&requestBody)
		if err != nil {
			return nil, err
		}
		bodyBytes, err := json.Marshal(requestBody)
		request, err = http.NewRequest(p.Context.RequestMethod, p.Context.RequestUrl, bytes.NewBuffer(bodyBytes))
		p.addArkoseTokenInHeaderIfNeeded(request, token)
		if err != nil {
			return nil, err
		}
	}
	p.buildHeaders(request)
	p.setCookies(request)
	return request, nil
}
func (p *BackendProcess) buildHeaders(request *http.Request) {
	context.Logger.Debug("BackendProcess buildHeaders")
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

func (p *BackendProcess) jsonResponse(response *http.Response) error {
	context.Logger.Debug("BackendProcess jsonResponse")
	var jsonData interface{}
	err := json.NewDecoder(response.Body).Decode(&jsonData)
	if err != nil {
		return err
	}
	p.GetContext().GinContext.JSON(response.StatusCode, jsonData)
	return nil
}

func (p *BackendProcess) streamResponse(response *http.Response) error {
	context.Logger.Debug("BackendProcess streamResponse")
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
func (p *BackendProcess) addArkoseTokenInHeaderIfNeeded(request *http.Request, token string) {
	request.Header.Set("Openai-Sentinel-Arkose-Token", token)
}
func (p *BackendProcess) addArkoseTokenIfNeeded(requestBody *map[string]interface{}) (string, error) {
	context.Logger.Debug("BackendProcess addArkoseTokenIfNeeded")
	model, exists := (*requestBody)["model"]
	if !exists {
		return "", nil
	}
	if strings.HasPrefix(model.(string), "gpt-4") || context.Env.ArkoseMust {
		token, err := funcaptcha.GetOpenAIArkoseToken(4, p.GetContext().RequestHeaders.Get("puid"))
		if err != nil {
			p.GetContext().GinContext.JSON(500, gin.H{"error": "Get ArkoseToken Failed"})
			return "", err
		}
		(*requestBody)["arkose_token"] = token
		return token, nil
	}
	return "", nil
}
func (p *BackendProcess) setCookies(request *http.Request) {
	context.Logger.Debug("BackendProcess setCookies")
	for _, cookie := range p.GetContext().GinContext.Request.Cookies() {
		request.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}
}

func (p *BackendProcess) decodeRequestBody(requestBody *map[string]interface{}) error {
	conversation := p.GetContext()
	if conversation.RequestBody != shttp.NoBody {
		if err := json.NewDecoder(conversation.RequestBody).Decode(requestBody); err != nil {
			conversation.GinContext.JSON(400, gin.H{"error": "JSON invalid"})
			return err
		}
	}
	return nil
}

type ReverseBackendRequestUrl struct {
}

func (u ReverseBackendRequestUrl) Generate(path string, rawquery string) string {
	if rawquery == "" {
		return "https://" + context.Env.OpenaiHost + "/backend-api" + path
	}
	return "https://" + context.Env.OpenaiHost + "/backend-api" + path + "?" + rawquery
}

func (p *BackendProcess) Run(com *plugins.Component) {
	context = com
	context.Engine.Any("/backend-api/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c, ReverseBackendRequestUrl{})
		common.Do[Context](new(BackendProcess), Context(conversation))
	})
}
