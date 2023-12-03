package api

import (
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
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
	err := process.DecodeRequestBody(p, &requestBody)
	if err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Json Error"})
		return
	}
	request, err := createRequest(p, requestBody)
	if err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Server error"})
		return
	}
	response, err := p.GetConversation().RequestClient.Do(request)
	if err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Server Error"})
		return
	}

	process.CopyResponseHeaders(response, p.GetConversation().GinContext)

	if _, exists := requestBody["stream"].(bool); exists {
		err := process.StreamResponse(p, response)
		if err != nil {
			logger.Log.Fatal(err)
		}
	} else {
		err := process.SendJsonResponse(p, response)
		if err != nil {
			logger.Log.Fatal(err)
		}
	}
}

func createRequest(p *OfficialApiProcess, requestBody map[string]interface{}) (*http.Request, error) {
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(p.Conversation.RequestMethod, p.Conversation.RequestUrl, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", p.Conversation.GinContext.Request.Header.Get("Authorization"))
	request.Header.Set("Content-Type", "application/json")
	return request, nil
}
