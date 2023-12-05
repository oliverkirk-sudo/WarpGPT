package reverse

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
	"WarpGPT/pkg/tools"
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	"io"
	shttp "net/http"
	"strings"
)

type BackendProcess struct {
	process.Process
}

func (p *BackendProcess) GetConversation() requestbody.Conversation {
	return p.Conversation
}
func (p *BackendProcess) SetConversation(conversation requestbody.Conversation) {
	p.Conversation = conversation
}

func (p *BackendProcess) ProcessMethod() {
	logger.Log.Debug("ProcessBackendProcess")
	var requestBody map[string]interface{}
	err := process.DecodeRequestBody(p, &requestBody) //解析请求体
	if err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Incorrect json format"})
		return
	}
	request, err := p.createRequest(requestBody) //创建请求
	if err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Server error"})
		return
	}

	response, err := p.GetConversation().RequestClient.Do(request) //发送请求
	if err != nil {
		var jsonData interface{}
		err = json.NewDecoder(response.Body).Decode(&jsonData)
		if err != nil {
			p.GetConversation().GinContext.JSON(500, gin.H{"error": "Request json decode error"})
			return
		}
		p.GetConversation().GinContext.JSON(response.StatusCode, jsonData)
		return
	}

	process.CopyResponseHeaders(response, p.GetConversation().GinContext) //设置响应头

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
	if p.Conversation.RequestBody == shttp.NoBody {
		request, _ = http.NewRequest(p.Conversation.RequestMethod, p.Conversation.RequestUrl, nil)
	} else {
		err := p.addArkoseTokenIfNeeded(&requestBody)
		if err != nil {
			return nil, err
		}
		bodyBytes, err := json.Marshal(requestBody)
		request, err = http.NewRequest(p.Conversation.RequestMethod, p.Conversation.RequestUrl, bytes.NewBuffer(bodyBytes))
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
		"Host":          common.Env.OpenAI_HOST,
		"Origin":        "https://" + common.Env.OpenAI_HOST + "/chat",
		"Authorization": p.GetConversation().GinContext.Request.Header.Get("Authorization"),
		"Connection":    "keep-alive",
		"User-Agent":    common.Env.UserAgent,
		"Content-Type":  p.GetConversation().GinContext.Request.Header.Get("Content-Type"),
	}

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	if puid := p.GetConversation().GinContext.Request.Header.Get("PUID"); puid != "" {
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
	p.GetConversation().GinContext.JSON(response.StatusCode, jsonData)
	return nil
}

func (p *BackendProcess) streamResponse(response *http.Response) error {
	logger.Log.Debug("BackendProcess streamResponse")
	defer response.Body.Close()
	var accumulatedData strings.Builder

	buf := make([]byte, 1024)
	for {
		n, err := response.Body.Read(buf)
		if n > 0 {
			accumulatedData.Write(buf[:n])
			if strings.HasSuffix(accumulatedData.String(), "\n\n") || strings.Contains(accumulatedData.String(), "[DONE]") {
				if _, err := p.GetConversation().GinContext.Writer.Write([]byte(accumulatedData.String())); err != nil {
					return err
				}
				p.GetConversation().GinContext.Writer.Flush()
				accumulatedData.Reset()
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		select {
		case <-p.GetConversation().GinContext.Writer.CloseNotify():
			return nil
		default:
		}
	}
	return nil
}
func (p *BackendProcess) addArkoseTokenIfNeeded(requestBody *map[string]interface{}) error {
	logger.Log.Debug("BackendProcess addArkoseTokenIfNeeded")
	model, exists := (*requestBody)["model"]
	if !exists {
		return nil
	}
	if strings.HasPrefix(model.(string), "gpt-4") || common.Env.ArkoseMust {
		token, err := tools.NewAuthenticator("", "", "").GetLoginArkoseToken()
		if err != nil {
			p.GetConversation().GinContext.JSON(500, gin.H{"error": "Get ArkoseToken Failed"})
			return err.Error
		}
		(*requestBody)["arkose_token"] = token.Token
	}
	return nil
}
func (p *BackendProcess) setCookies(request *http.Request) {
	logger.Log.Debug("BackendProcess setCookies")
	for _, cookie := range p.GetConversation().GinContext.Request.Cookies() {
		request.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}
}
