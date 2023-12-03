package reverse

import (
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
)

type ApiProcess struct {
	process.Process
}

func (p *ApiProcess) SetConversation(conversation requestbody.Conversation) {
	p.Conversation = conversation
}
func (p *ApiProcess) GetConversation() requestbody.Conversation {
	return p.Conversation
}
func (p *ApiProcess) ProcessMethod() {
	var requestBody map[string]interface{}
	logger.Log.Debug("ProcessApiProcess")
	if err := process.DecodeRequestBody(p, &requestBody); err != nil {
		return
	}
	if err := process.ProcessRegularRequest(p, &requestBody); err != nil {
		return
	}
}
