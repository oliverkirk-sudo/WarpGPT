package reverse

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	shttp "net/http"
	"strings"
)

type PublicApiProcess struct {
	process.Process
}

func (p *PublicApiProcess) SetConversation(conversation requestbody.Conversation) {
	p.Conversation = conversation
}
func (p *PublicApiProcess) GetConversation() requestbody.Conversation {
	return p.Conversation
}
func (p *PublicApiProcess) ProcessMethod() {
	logger.Log.Debug("PublicApiProcess")
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
		err := json.NewDecoder(response.Body).Decode(&jsonData)
		if err != nil {
			p.GetConversation().GinContext.JSON(500, gin.H{"error": "Request json decode error"})
			return
		}
		p.GetConversation().GinContext.JSON(response.StatusCode, jsonData)
		return
	}
	if strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		err := p.jsonResponse(response)
		if err != nil {
			logger.Log.Fatal(err)
		}
	}
	process.CopyResponseHeaders(response, p.GetConversation().GinContext) //设置响应头
}
func (p *PublicApiProcess) createRequest(requestBody map[string]interface{}) (*http.Request, error) {
	logger.Log.Debug("PublicApiProcess createRequest")
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	bodyReader := bytes.NewReader(bodyBytes)
	var request *http.Request
	if p.Conversation.RequestBody == shttp.NoBody {
		request, err = http.NewRequest(p.Conversation.RequestMethod, p.Conversation.RequestUrl, nil)
	} else {
		request, err = http.NewRequest(p.Conversation.RequestMethod, p.Conversation.RequestUrl, bodyReader)
	}
	if err != nil {
		return nil, err
	}
	p.buildHeaders(request)
	p.setCookies(request)
	return request, nil
}
func (p *PublicApiProcess) setCookies(request *http.Request) {
	logger.Log.Debug("PublicApiProcess setCookies")
	for _, cookie := range p.GetConversation().GinContext.Request.Cookies() {
		request.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}
}
func (p *PublicApiProcess) buildHeaders(request *http.Request) {
	logger.Log.Debug("PublicApiProcess buildHeaders")
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
func (p *PublicApiProcess) jsonResponse(response *http.Response) error {
	logger.Log.Debug("PublicApiProcess jsonResponse")
	var jsonData interface{}
	err := json.NewDecoder(response.Body).Decode(&jsonData)
	if err != nil {
		return err
	}
	p.GetConversation().GinContext.JSON(response.StatusCode, jsonData)
	return nil
}
