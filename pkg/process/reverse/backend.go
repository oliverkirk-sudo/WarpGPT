package reverse

import (
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
)

type BackendProcess struct {
	process.Process
}

func (p *BackendProcess) SetConversation(conversation requestbody.Conversation) {
	p.Conversation = conversation
}

func (p *BackendProcess) ProcessMethod() {

}
