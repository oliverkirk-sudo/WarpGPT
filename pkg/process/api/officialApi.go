package api

import (
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	"strings"
)

type OfficialApiProcess struct {
	process.Process
}

func (p *OfficialApiProcess) SetConversation(conversation requestbody.Conversation) {
	p.Conversation = conversation
}
func (p *OfficialApiProcess) ProcessMethod() {
	var requestBody map[string]interface{}
	logger.Log.Debug("ProcessOfficialApiConversation")
	if p.Conversation.RequestBody != nil {
		err := json.NewDecoder(p.Conversation.RequestBody).Decode(&requestBody)
		if err != nil {
			p.Conversation.GinContext.JSON(400, gin.H{"error": "JSON invalid"})
			return
		}
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		p.Conversation.GinContext.JSON(500, gin.H{"error": "Error encoding JSON"})
		return
	}
	request, err := http.NewRequest(p.Conversation.RequestMethod, p.Conversation.RequestUrl, bytes.NewBuffer(bodyBytes))
	if err != nil {
		p.Conversation.GinContext.JSON(500, gin.H{"error": "Wait a minute. There's a server error."})
		return
	}
	request.Header.Set("Authorization", p.Conversation.GinContext.Request.Header.Get("Authorization"))
	request.Header.Set("Content-Type", "application/json")
	response, err := p.Conversation.RequestClient.Do(request)
	if err != nil {
		p.Conversation.GinContext.JSON(500, gin.H{"error": "Return Error"})
		return
	}
	defer response.Body.Close()
	_, exists := requestBody["stream"].(bool)
	for name, values := range response.Header {
		for _, value := range values {
			p.Conversation.GinContext.Header(name, value)
		}
	}
	if exists {
		p.Conversation.GinContext.Header("Content-Type", response.Header.Get("Content-Type"))
		logger.Log.Infoln("Stream Request")
		buf := make([]byte, 1024)
		defer response.Body.Close()
		for {
			closeNotifier := p.Conversation.GinContext.Writer.CloseNotify()
			select {
			case <-closeNotifier:
				break
			default:
				n, err := response.Body.Read(buf)
				p.Conversation.GinContext.Writer.Write(buf[:n])
				p.Conversation.GinContext.Writer.Flush()
				if strings.Contains(string(buf[:n]), "[DONE]") {
					break
				}
				if err != nil {
					p.Conversation.GinContext.JSON(response.StatusCode, response.Body)
					return
				}
			}
		}
	} else {
		var responseBody map[string]interface{}
		err := json.NewDecoder(response.Body).Decode(&responseBody)
		p.Conversation.GinContext.Header("Content-Encoding", "")
		if err != nil {
			p.Conversation.GinContext.JSON(response.StatusCode, response.Body)
			return
		}
		p.Conversation.GinContext.JSON(response.StatusCode, responseBody)
	}

}
