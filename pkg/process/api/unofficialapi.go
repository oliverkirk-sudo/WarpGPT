package api

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

var id string
var model string
var oldString = ""

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
		return
	}
	id = common.IdGenerator()
	_, exists := requestBody["model"]
	if exists {
		model, _ = requestBody["model"].(string)
	} else {
		p.GetConversation().GinContext.JSON(400, gin.H{"error": "Model not provided"})
		return
	}
	if strings.HasSuffix(p.GetConversation().RequestParam, "chat/completions") {
		if err := p.chatApiProcess(requestBody); err != nil {
			println(err.Error())
			return
		}
	}
	if strings.HasSuffix(p.GetConversation().RequestParam, "images/generations") {
		if err := p.imageApiProcess(requestBody); err != nil {
			println(err.Error())
			return
		}
	}
}

func (p *UnofficialApiProcess) imageApiProcess(requestBody map[string]interface{}) error {
	logger.Log.Debug("imageApiProcess")
	if err := process.ProcessConversationRequest(p, &requestBody, jsonImageProcess); err != nil {
		return err
	}
	return nil
}

func (p *UnofficialApiProcess) chatApiProcess(requestBody map[string]interface{}) error {
	logger.Log.Debug("chatApiProcess")

	value, exists := requestBody["stream"]
	reqModel, err := checkModel(model)
	if err != nil {
		p.GetConversation().GinContext.JSON(400, gin.H{"error": err.Error()})
	}
	req := common.GetChatReqStr(reqModel)
	if err := generateBody(req, requestBody); err != nil {
		return err
	}
	jsonData, _ := json.Marshal(req)
	var request map[string]interface{}
	err = json.Unmarshal(jsonData, &request)
	if err != nil {
		p.GetConversation().GinContext.JSON(400, gin.H{"error": err.Error()})
	}
	if exists && value.(bool) {
		if err := process.ProcessConversationRequest(p, &request, streamChatProcess); err != nil {
			println(err.Error())
			return err
		}
	} else {
		if err := process.ProcessConversationRequest(p, &request, jsonChatProcess); err != nil {
			return err
		}
	}

	return nil
}

func streamChatProcess(raw string) string {
	jsonData := strings.Trim(strings.SplitN(raw, "data: ", 2)[1], "\n")
	result := getStreamResp(jsonData)
	if strings.Contains(raw, "[DONE]") {
		return raw
	} else if result.ApiRespStrStreamEnd.Id != "" {
		jsonData, err := json.Marshal(result.ApiRespStrStreamEnd)
		if err != nil {
			logger.Log.Fatal(err)
		}
		return "data: " + string(jsonData) + "\n\n"
	} else if result.ApiRespStrStream.Id != "" {
		jsonData, err := json.Marshal(result.ApiRespStrStream)
		if err != nil {
			logger.Log.Fatal(err)
		}
		return "data: " + string(jsonData) + "\n\n"
	}
	return ""
}
func jsonChatProcess(raw string) string {
	jsonData := strings.Trim(strings.SplitN(raw, "data: ", 2)[1], "\n")
	getStreamResp(jsonData)
	if strings.Contains(raw, "[DONE]") {
		resp := common.GetApiRespStr(id)
		choice := common.GetStrChoices()
		choice.Message.Content = oldString
		resp.Choices = append(resp.Choices, *choice)
		resp.Model = model
		data, _ := json.Marshal(resp)
		return string(data)
	}
	return ""
}
func jsonImageProcess(raw string) string {
	println(raw)
	return raw
}
func getStreamResp(stream string) *Result {
	var chatRespStr common.ChatRespStr
	var chatEndRespStr common.ChatEndRespStr
	result := &Result{
		ApiRespStrStream:    common.ApiRespStrStream{},
		ApiRespStrStreamEnd: common.ApiRespStrStreamEnd{},
		Pass:                false,
	}
	json.Unmarshal([]byte(stream), &chatRespStr)
	if chatRespStr.Message.Id != "" {
		resp := common.GetApiRespStrStream(id)
		choice := common.GetStreamChoice()
		resp.Model = model
		choice.Delta.Content = strings.ReplaceAll(chatRespStr.Message.Content.Parts[0], oldString, "")
		oldString = chatRespStr.Message.Content.Parts[0]
		resp.Choices = append(resp.Choices, *choice)
		result.ApiRespStrStream = *resp
	}
	json.Unmarshal([]byte(stream), &chatEndRespStr)
	if chatEndRespStr.MessageId != "" {
		resp := common.GetApiRespStrStreamEnd(id)
		resp.Model = model
		result.ApiRespStrStreamEnd = *resp
	}
	if result.ApiRespStrStream.Id == "" && result.ApiRespStrStreamEnd.Id == "" {
		result.Pass = true
		return result
	}
	if chatRespStr.Message.Metadata.ParentId == "" {
		result.Pass = true
		return result
	}
	return result
}
func checkModel(model string) (string, error) {
	logger.Log.Debug("checkModel")
	if strings.HasPrefix(model, "dalle") || strings.HasPrefix(model, "gpt-4-vision") {
		return "gpt-4", nil
	} else if strings.HasPrefix(model, "gpt-3") {
		return "text-davinci-002-render-sha", nil
	} else if strings.HasPrefix(model, "gpt-4") {
		return "gpt-4-gizmo", nil
	} else {
		return "", errors.New("unsupported model")
	}
}
func generateBody(req *common.ChatReqStr, requestBody map[string]interface{}) error {
	reqMessage := common.GetChatReqTemplate()
	reqFileMessage := common.GetChatFileReqTemplate()
	messageList, exists := requestBody["messages"]
	if !exists {
		return errors.New("no message body")
	}
	messages, _ := messageList.([]interface{})

	for _, message := range messages {
		messageItem, _ := message.(map[string]interface{})
		role, _ := messageItem["role"].(string)
		if _, ok := messageItem["content"].(string); ok {
			content, _ := messageItem["content"].(string)
			reqMessage.Content.Parts = reqMessage.Content.Parts[:0]
			reqMessage.Author.Role = role
			reqMessage.Content.Parts = append(reqMessage.Content.Parts, content)
			req.Messages = append(req.Messages, *reqMessage)
		}
		if _, ok := messageItem["content"].([]map[string]interface{}); ok {
			content, _ := messageItem["content"].([]map[string]interface{})
			reqFileMessage.Content.Parts = reqFileMessage.Content.Parts[:0]
			reqFileMessage.Author.Role = role
			fileReqProcess(&content, &reqFileMessage.Content.Parts)
			//reqMessage.Content.Parts = append(reqMessage.Content.Parts, content)
			//req.Messages = append(req.Messages, *reqFileMessage)
		}
	}
	return nil
}
func fileReqProcess(content *[]map[string]interface{}, part *[]interface{}) {

}
