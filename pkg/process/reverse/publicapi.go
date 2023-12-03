package reverse

import (
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
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
	var requestBody map[string]interface{}
	logger.Log.Debug("ProcessPublicApiProcess")
	err := process.DecodeRequestBody(p, &requestBody)
	if err != nil {
		return
	}
	err = process.ProcessRegularRequest(p, &requestBody)
	if err != nil {
		return
	}
}
