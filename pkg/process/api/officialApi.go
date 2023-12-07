package api

import (
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
	"WarpGPT/pkg/tools"
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	shttp "net/http"
	"strings"
)

type OfficialApiProcess struct {
	process.Process
}

func (p *OfficialApiProcess) SetConversation(conversation requestbody.Conversation) {
	p.Conversation = conversation
}
func (p *OfficialApiProcess) GetConversation() requestbody.Conversation {
	return p.Conversation
}
func (p *OfficialApiProcess) ProcessMethod() {
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
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Server Error"})
		return
	}

	process.CopyResponseHeaders(response, p.GetConversation().GinContext) //设置响应头

	if strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		err = p.streamResponse(response)
		if err != nil {
			return
		}
	}
	if strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		err = p.jsonResponse(response)
		if err != nil {
			logger.Log.Fatal(err)
		}
	}
}

func (p *OfficialApiProcess) createRequest(requestBody map[string]interface{}) (*http.Request, error) {
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	var request *http.Request
	if p.Conversation.RequestBody == shttp.NoBody {
		request, err = http.NewRequest(p.Conversation.RequestMethod, p.Conversation.RequestUrl, nil)
	} else {
		request, err = http.NewRequest(p.Conversation.RequestMethod, p.Conversation.RequestUrl, bytes.NewBuffer(bodyBytes))
	}
	if err != nil {
		return nil, err
	}
	p.WithHeaders(request)
	return request, nil
}

func (p *OfficialApiProcess) WithHeaders(rsq *http.Request) {
	rsq.Header.Set("Authorization", p.Conversation.RequestHeaders.Get("Authorization"))
	rsq.Header.Set("Content-Type", p.Conversation.RequestHeaders.Get("Content-Type"))
}

func (p *OfficialApiProcess) jsonResponse(response *http.Response) error {
	var jsonData interface{}
	err := json.NewDecoder(response.Body).Decode(&jsonData)
	if err != nil {
		return err
	}
	p.GetConversation().GinContext.JSON(response.StatusCode, jsonData)
	return nil
}

func (p *OfficialApiProcess) streamResponse(response *http.Response) error {
	logger.Log.Infoln("officialApiProcess stream Request")
	client := tools.NewSSEClient(response.Body)
	events := client.Read()
	for event := range events {
		if _, err := p.GetConversation().GinContext.Writer.Write([]byte("data: " + event.Data + "\n\n")); err != nil {
			return err
		}
		p.GetConversation().GinContext.Writer.Flush()
	}
	defer client.Close()
	return nil
}
