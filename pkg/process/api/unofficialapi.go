package api

import (
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
)

type UnofficialApiProcess struct {
	process.Process
}

func (p *UnofficialApiProcess) SetConversation(conversation requestbody.Conversation) {
	p.Conversation = conversation
}
func (p *UnofficialApiProcess) GetConversation() requestbody.Conversation {
	return p.Conversation
}
func (p *UnofficialApiProcess) ProcessMethod() {

}
