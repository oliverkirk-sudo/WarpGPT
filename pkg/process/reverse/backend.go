package reverse

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/funcaptcha"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/tools"
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	shttp "net/http"
	"strings"
)

type BackendProcess struct {
	process.Process
}

func (p *BackendProcess) GetContext() common.Context {
	return p.Context
}
func (p *BackendProcess) SetContext(conversation common.Context) {
	p.Context = conversation
}

func (p *BackendProcess) ProcessMethod() {
	logger.Log.Debug("ProcessBackendProcess")
	var requestBody map[string]interface{}
	err := process.DecodeRequestBody(p, &requestBody) //解析请求体
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
		var jsonData interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonData)
		if err != nil {
			p.GetContext().GinContext.JSON(500, gin.H{"error": "Request json decode error"})
			return
		}
		p.GetContext().GinContext.JSON(response.StatusCode, jsonData)
		return
	}

	process.CopyResponseHeaders(response, p.GetContext().GinContext) //设置响应头

	if strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		err := p.streamResponse(response)
		if err != nil {
			return
		}
	}
	if strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		err := p.jsonResponse(response)
		if err != nil {
			logger.Log.Fatal(err)
		}
	}
}
func (p *BackendProcess) createRequest(requestBody map[string]interface{}) (*http.Request, error) {
	logger.Log.Debug("BackendProcess createRequest")
	var request *http.Request
	if p.Context.RequestBody == shttp.NoBody {
		request, _ = http.NewRequest(p.Context.RequestMethod, p.Context.RequestUrl, nil)
	} else {
		err := p.addArkoseTokenIfNeeded(&requestBody)
		if err != nil {
			return nil, err
		}
		bodyBytes, err := json.Marshal(requestBody)
		request, err = http.NewRequest(p.Context.RequestMethod, p.Context.RequestUrl, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, err
		}
	}
	p.buildHeaders(request)
	p.setCookies(request)
	return request, nil
}
func (p *BackendProcess) buildHeaders(request *http.Request) {
	logger.Log.Debug("BackendProcess buildHeaders")
	headers := map[string]string{
		"Host":          common.Env.OpenaiHost,
		"Origin":        "https://" + common.Env.OpenaiHost + "/chat",
		"Authorization": p.GetContext().GinContext.Request.Header.Get("Authorization"),
		"Connection":    "keep-alive",
		"User-Agent":    common.Env.UserAgent,
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
	logger.Log.Debug("BackendProcess jsonResponse")
	var jsonData interface{}
	err := json.NewDecoder(response.Body).Decode(&jsonData)
	if err != nil {
		return err
	}
	p.GetContext().GinContext.JSON(response.StatusCode, jsonData)
	return nil
}

func (p *BackendProcess) streamResponse(response *http.Response) error {
	logger.Log.Debug("BackendProcess streamResponse")
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
func (p *BackendProcess) addArkoseTokenIfNeeded(requestBody *map[string]interface{}) error {
	logger.Log.Debug("BackendProcess addArkoseTokenIfNeeded")
	model, exists := (*requestBody)["model"]
	if !exists {
		return nil
	}
	if strings.HasPrefix(model.(string), "gpt-4") || common.Env.ArkoseMust {
		token, err := funcaptcha.GetOpenAIArkoseToken(funcaptcha.ArkVerChat4, p.GetContext().RequestHeaders.Get("puid"))
		if err != nil {
			p.GetContext().GinContext.JSON(500, gin.H{"error": "Get ArkoseToken Failed"})
			return err
		}
		(*requestBody)["arkose_token"] = token
	}
	return nil
}
func (p *BackendProcess) setCookies(request *http.Request) {
	logger.Log.Debug("BackendProcess setCookies")
	for _, cookie := range p.GetContext().GinContext.Request.Cookies() {
		request.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}
}
