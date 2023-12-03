package api

import (
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	"io"
)

type OfficialApiProcess struct {
	process.Process
}

func (p *OfficialApiProcess) SetConversation(conversation requestbody.Conversation) {
	p.Conversation = conversation
}

func (p *OfficialApiProcess) ProcessMethod() {
	requestBody := decodeRequestBody(p)
	if requestBody == nil {
		return
	}

	request, err := createRequest(p, requestBody)
	if err != nil {
		p.Conversation.GinContext.JSON(500, gin.H{"error": "Server error"})
		return
	}

	response, err := p.Conversation.RequestClient.Do(request)
	if err != nil {
		p.Conversation.GinContext.JSON(500, gin.H{"error": "Return Error"})
		return
	}
	defer response.Body.Close()

	process.CopyResponseHeaders(response, p.Conversation.GinContext)

	if _, exists := requestBody["stream"].(bool); exists {
		streamResponse(p, response)
	} else {
		sendJsonResponse(p, response)
	}
}

func decodeRequestBody(p *OfficialApiProcess) map[string]interface{} {
	var requestBody map[string]interface{}
	if p.Conversation.RequestBody != nil {
		err := json.NewDecoder(p.Conversation.RequestBody).Decode(&requestBody)
		if err != nil {
			p.Conversation.GinContext.JSON(400, gin.H{"error": "JSON invalid"})
			return nil
		}
	}
	return requestBody
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

func streamResponse(p *OfficialApiProcess, response *http.Response) {
	logger.Log.Infoln("Stream Request")
	buf := make([]byte, 1024)
	for {
		n, err := response.Body.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			p.Conversation.GinContext.JSON(response.StatusCode, response.Body)
			return
		}
		p.Conversation.GinContext.Writer.Write(buf[:n])
		p.Conversation.GinContext.Writer.Flush()
	}
}

func sendJsonResponse(p *OfficialApiProcess, response *http.Response) {
	var responseBody map[string]interface{}
	err := json.NewDecoder(response.Body).Decode(&responseBody)
	if err != nil {
		p.Conversation.GinContext.JSON(response.StatusCode, response.Body)
		return
	}
	p.Conversation.GinContext.JSON(response.StatusCode, responseBody)
}
