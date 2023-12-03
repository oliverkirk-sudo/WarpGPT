package api

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"strings"
)

var id string
var model string

type UnofficialApiProcess struct {
	process.Process
}
type Result struct {
	ApiRespStrStream    common.ApiRespStrStream
	ApiRespStrStreamEnd common.ApiRespStrStreamEnd
	Pass                bool
}

func (p *UnofficialApiProcess) SetConversation(conversation requestbody.Conversation) {
	p.Conversation = conversation
}
func (p *UnofficialApiProcess) GetConversation() requestbody.Conversation {
	return p.Conversation
}

func (p *UnofficialApiProcess) ProcessMethod() {
	var requestBody map[string]interface{}
	err := process.DecodeRequestBody(p, &requestBody)
	if err != nil {
		p.GetConversation().GinContext.JSON(400, gin.H{"error": "Incorrect json format"})
	}
	id = common.IdGenerator()
	_, exists := requestBody["model"]
	if exists {
		model, _ = requestBody["model"].(string)
	} else {
		p.GetConversation().GinContext.JSON(400, gin.H{"error": "Model not provided"})
	}
	if strings.HasSuffix(p.GetConversation().RequestParam, "chat/completions") {
		if err := p.chatApiProcess(); err != nil {
			return
		}
	}
	if strings.HasSuffix(p.GetConversation().RequestParam, "images/generations") {
		if err := p.imageApiProcess(); err != nil {
			return
		}
	}
}

func (p *UnofficialApiProcess) imageApiProcess() error {
	var requestBody map[string]interface{}
	logger.Log.Debug("imageApiProcess")

	if err := process.DecodeRequestBody(p, &requestBody); err != nil {
		return err
	}
	if err := process.ProcessConversationRequest(p, &requestBody, jsonImageProcess); err != nil {
		return err
	}
	return nil
}

func (p *UnofficialApiProcess) chatApiProcess() error {
	var requestBody map[string]interface{}
	logger.Log.Debug("chatApiProcess")

	if err := process.DecodeRequestBody(p, &requestBody); err != nil {
		return err
	}
	value, exists := requestBody["stream"]
	if exists && value.(string) == "true" {
		if err := process.ProcessConversationRequest(p, &requestBody, streamChatProcess); err != nil {
			return err
		}
	} else {
		if err := process.ProcessConversationRequest(p, &requestBody, jsonChatProcess); err != nil {
			return err
		}
	}

	return nil
}

func streamChatProcess(raw string) string {
	jsonData := strings.Trim(strings.SplitN(raw, ":", 1)[1], "\n")
	checkStreamClass(jsonData)
	return raw
}
func jsonChatProcess(raw string) string {
	println(raw)
	return raw
}
func jsonImageProcess(raw string) string {
	println(raw)
	return raw
}
func checkStreamClass(stream string) *Result {
	var chatRespStr common.ChatRespStr
	var chatEndRespStr common.ChatEndRespStr
	result := &Result{
		ApiRespStrStream:    common.ApiRespStrStream{},
		ApiRespStrStreamEnd: common.ApiRespStrStreamEnd{},
		Pass:                false,
	}
	err := json.Unmarshal([]byte(stream), &chatRespStr)
	if err == nil {
		resp := common.GetApiRespStrStream(id)
		resp.Model = model
		result.ApiRespStrStream = *resp
	}
	err = json.Unmarshal([]byte(stream), &chatEndRespStr)
	if err == nil {
		resp := common.GetApiRespStrStreamEnd(id)
		resp.Model = model
		result.ApiRespStrStreamEnd = *resp
	}
	if result.ApiRespStrStream.Id == "" && result.ApiRespStrStreamEnd.Id == "" {
		result.Pass = true
		return result
	}
	return result
}
