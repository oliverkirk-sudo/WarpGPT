package backendapi

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/funcaptcha"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/plugins"
	"WarpGPT/pkg/plugins/service/wsstostream"
	"WarpGPT/pkg/tools"
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/gin-gonic/gin"
	"io"
	shttp "net/http"
	"strings"
	"time"
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

type WsResponse struct {
	ConversationId string    `json:"conversation_id"`
	ExpiresAt      time.Time `json:"expires_at"`
	ResponseId     string    `json:"response_id"`
	WssUrl         string    `json:"wss_url"`
}

type BackendProcess struct {
	ConversationId string
	Context        Context
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
	var ws *wsstostream.WssToStream
	err := p.decodeRequestBody(&requestBody)
	if err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Request json decode error"})
		context.Logger.Error(err)
		return
	}
	request, err := p.createRequest(requestBody)
	if err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Server error"})
		context.Logger.Error(err)
		return
	}
	if strings.Contains(p.Context.RequestParam, "/conversation/ws") {
		ws = wsstostream.NewWssToStream(p.GetContext().RequestHeaders.Get("Authorization"))
		err = ws.InitConnect()
	}
	if err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": err.Error()})
		context.Logger.Error(err)
		return
	}
	context.Logger.Debug("Requesting to ", p.GetContext().RequestUrl)
	response, err := p.GetContext().RequestClient.Do(request)
	if err != nil {
		var jsonData interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonData)
		if err != nil {
			p.GetContext().GinContext.JSON(500, gin.H{"error": "Request json decode error"})
			context.Logger.Error(err)
			return
		}
		p.GetContext().GinContext.JSON(response.StatusCode, jsonData)
		context.Logger.Error(err)
		return
	}

	common.CopyResponseHeaders(response, p.GetContext().GinContext)

	if strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		err = p.streamResponse(response)
		if err != nil {
			p.GetContext().GinContext.JSON(500, gin.H{"error": err.Error()})
			context.Logger.Error(err)
			return
		}
	}
	if strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		if strings.Contains(p.Context.RequestParam, "/conversation/ws") {
			context.Logger.Debug("WsToStreamResponse")
			p.WsToStreamResponse(ws, response)
		} else {
			err = p.jsonResponse(response)
			if err != nil {
				p.GetContext().GinContext.JSON(500, gin.H{"error": err.Error()})
				context.Logger.Error(err)
				return
			}
		}
	}
}
func (p *BackendProcess) WsToStreamResponse(ws *wsstostream.WssToStream, response *http.Response) {
	var jsonData WsResponse
	err := json.NewDecoder(response.Body).Decode(&jsonData)
	if err != nil {
		context.Logger.Error(err)
	}
	ws.ResponseId = jsonData.ResponseId
	ws.ConversationId = jsonData.ConversationId
	p.GetContext().GinContext.Writer.Header().Set("Content-Type", "text/event-stream")
	p.GetContext().GinContext.Writer.Header().Set("Cache-Control", "no-cache")
	p.GetContext().GinContext.Writer.Header().Set("Connection", "keep-alive")
	ctx := p.GetContext().GinContext.Request.Context()
	for {
		select {
		case <-p.GetContext().GinContext.Writer.CloseNotify():
			logger.Log.Debug("WsToStreamResponse Writer.CloseNotify")
			return
		case <-ctx.Done():
			logger.Log.Debug("WsToStreamResponse ctx.Done")
			return
		default:
			message, err := ws.ReadMessage()
			if err != nil {
				context.Logger.Error(err)
				break
			}
			if message != nil {
				data, err := io.ReadAll(message)
				if err != nil {
					context.Logger.Error(err)
					return
				}
				_, writeErr := p.GetContext().GinContext.Writer.Write(data)
				if writeErr != nil {
					return
				}
				p.GetContext().GinContext.Writer.Flush()
				if strings.Contains(string(data), "data: [DONE]") {
					return
				}
			}
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
		if token != "" {
			p.addArkoseTokenInHeaderIfNeeded(request, token)
		}
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
	context.Logger.Debug("BackendProcess addArkoseTokenInHeaderIfNeeded")
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
	if strings.Contains(path, "/ws") {
		path = strings.ReplaceAll(path, "/ws", "")
	}
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
