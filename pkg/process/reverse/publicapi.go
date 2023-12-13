package reverse

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
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

func (p *PublicApiProcess) SetContext(conversation common.Context) {
	p.Context = conversation
}
func (p *PublicApiProcess) GetContext() common.Context {
	return p.Context
}
func (p *PublicApiProcess) ProcessMethod() {
	logger.Log.Debug("PublicApiProcess")
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
			logger.Log.Fatal(err)
		}
	}
	process.CopyResponseHeaders(response, p.GetContext().GinContext) //设置响应头
}
func (p *PublicApiProcess) createRequest(requestBody map[string]interface{}) (*http.Request, error) {
	logger.Log.Debug("PublicApiProcess createRequest")
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
	logger.Log.Debug("PublicApiProcess setCookies")
	for _, cookie := range p.GetContext().GinContext.Request.Cookies() {
		request.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}
}
func (p *PublicApiProcess) buildHeaders(request *http.Request) {
	logger.Log.Debug("PublicApiProcess buildHeaders")
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
func (p *PublicApiProcess) jsonResponse(response *http.Response) error {
	logger.Log.Debug("PublicApiProcess jsonResponse")
	var jsonData interface{}
	err := json.NewDecoder(response.Body).Decode(&jsonData)
	if err != nil {
		return err
	}
	p.GetContext().GinContext.JSON(response.StatusCode, jsonData)
	return nil
}
