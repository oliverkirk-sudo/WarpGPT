package arkosetoken

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/funcaptcha"
	"WarpGPT/pkg/logger"
	"github.com/gin-gonic/gin"
)

type ArkoseToken struct {
	common.Process
}

func (p *ArkoseToken) GetContext() common.Context {
	return p.Context
}
func (p *ArkoseToken) SetContext(conversation common.Context) {
	p.Context = conversation
}

func (p *ArkoseToken) ProcessMethod() {
	logger.Log.Debug("ArkoseToken")
	token, err := funcaptcha.GetOpenAIArkoseToken(funcaptcha.ArkVerChat4, p.GetContext().RequestHeaders.Get("puid"))
	if err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Unable to generate ArkoseToken"})
	}
	p.GetContext().GinContext.Header("Content-Type", "application/json")
	p.GetContext().GinContext.JSON(200, gin.H{"token": token})
}
