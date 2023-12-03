package reverse

import (
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
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
	var requestBody map[string]interface{}
	logger.Log.Debug("ProcessBackendProcess")

	if err := process.DecodeRequestBody(p, &requestBody); err != nil {
		return
	}

	if strings.HasSuffix(p.Conversation.RequestParam, "conversation") {
		if err := process.ProcessConversationRequest(p, &requestBody, func(a string) string {
			return a
		}); err != nil {
			return
		}
	} else {
		if err := process.ProcessRegularRequest(p, &requestBody); err != nil {
			return
		}
	}
}
