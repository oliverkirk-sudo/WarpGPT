package api

import (
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
)

type ApiProcess struct {
	process.Process
}

var api ApiProcess

func (p *ApiProcess) SetConversation(conversation requestbody.Conversation) *ApiProcess {
	api.Conversation = conversation
	return p
}
func (p *ApiProcess) ProcessOfficialApiConversation() {
	var request_body map[string]interface{}
	logger.Log.Debug("ProcessOfficialApiConversation")
	if api.Conversation.RequestBody != nil {
		err := json.NewDecoder(api.Conversation.RequestBody).Decode(&request_body)
		if err != nil {
			api.Conversation.GinContext.JSON(400, gin.H{"error": "JSON invalid"})
			return
		}
	}
	request, err := http.NewRequest(api.Conversation.RequestMethod, api.Conversation.RequestUrl, api.Conversation.RequestBody)
	if err != nil {
		api.Conversation.GinContext.JSON(500, gin.H{"error": "Wait a minute. There's a server error."})
		return
	}
	request.Header.Set("Authorization", api.Conversation.GinContext.Request.Header.Get("Authorization"))
	response, err := api.Conversation.RequestClient.Do(request)
	if err != nil {
		api.Conversation.GinContext.JSON(500, gin.H{"error": "Return Error"})
		return
	}

}
