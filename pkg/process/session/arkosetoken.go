package session

import (
	"WarpGPT/pkg/funcaptcha"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
	"github.com/gin-gonic/gin"
)

type ArkoseToken struct {
	process.Process
}

func (p *ArkoseToken) GetConversation() requestbody.Conversation {
	return p.Conversation
}
func (p *ArkoseToken) SetConversation(conversation requestbody.Conversation) {
	p.Conversation = conversation
}

func (p *ArkoseToken) ProcessMethod() {
	logger.Log.Debug("ArkoseToken")
	token, err := funcaptcha.GetArkoseToken(funcaptcha.ArkVerChat4, p.GetConversation().RequestHeaders.Get("puid"))
	if err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Unable to generate ArkoseToken"})
	}
	p.GetConversation().GinContext.Header("Content-Type", "application/json")
	p.GetConversation().GinContext.JSON(200, gin.H{"token": token})
}
