package api

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/funcaptcha"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
	"WarpGPT/pkg/tools"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	shttp "net/http"
	"strconv"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"

	"github.com/gin-gonic/gin"
)

type UnofficialApiProcess struct {
	process.Process
	ID               string
	Model            string
	OldString        string
	Mode             string
	ImagePointerList []ImagePointer
}
type ImagePointer struct {
	Pointer string
	Prompt  string
}
type Result struct {
	ApiRespStrStream          common.ApiRespStrStream
	ApiRespStrStreamEnd       common.ApiRespStrStreamEnd
	ApiImageGenerationRespStr common.ApiImageGenerationRespStr
	Pass                      bool
}

func (p *UnofficialApiProcess) SetConversation(conversation requestbody.Conversation) {
	p.Conversation = conversation
}
func (p *UnofficialApiProcess) GetConversation() requestbody.Conversation {
	return p.Conversation
}

func (p *UnofficialApiProcess) ProcessMethod() {
	logger.Log.Debug("UnofficialApiProcess")
	var requestBody map[string]interface{}
	err := process.DecodeRequestBody(p, &requestBody)
	if err != nil {
		p.GetConversation().GinContext.JSON(400, gin.H{"error": "Incorrect json format"})
		return
	}
	p.ID = common.IdGenerator()
	_, exists := requestBody["model"]
	if exists {
		p.Model, _ = requestBody["model"].(string)
	} else {
		p.GetConversation().GinContext.JSON(400, gin.H{"error": "Model not provided"})
		return
	}
	if strings.HasSuffix(p.GetConversation().RequestParam, "chat/completions") {
		p.Mode = "chat"
		if err = p.chatApiProcess(requestBody); err != nil {
			println(err.Error())
			return
		}
	}
	if strings.HasSuffix(p.GetConversation().RequestParam, "images/generations") {
		p.Mode = "image"
		if err = p.imageApiProcess(requestBody); err != nil {
			println(err.Error())
			return
		}
	}
}

func (p *UnofficialApiProcess) imageApiProcess(requestBody map[string]interface{}) error {
	logger.Log.Debug("UnofficialApiProcess imageApiProcess")
	response, err := p.MakeRequest(requestBody)
	if err != nil {
		return err
	}
	result := new(Result)
	result.ApiImageGenerationRespStr = common.ApiImageGenerationRespStr{}
	err = p.response(response, func(p *UnofficialApiProcess, a string) bool {
		p.jsonImageProcess(a)
		return false
	})
	if err = p.getImageUrlByPointer(&p.ImagePointerList, result); err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "get image url failed"})
		logger.Log.Fatal(err)
	}
	if result.ApiImageGenerationRespStr.Created != 0 {
		p.GetConversation().GinContext.Header("Content-Type", "application/json")
		p.GetConversation().GinContext.JSON(response.StatusCode, result.ApiImageGenerationRespStr)
	}
	if err != nil {
		return err
	}
	return nil
}

func (p *UnofficialApiProcess) chatApiProcess(requestBody map[string]interface{}) error {
	logger.Log.Debug("UnofficialApiProcess chatApiProcess")
	response, err := p.MakeRequest(requestBody)
	if err != nil {
		return err
	}
	value, exists := requestBody["stream"]

	if exists && value.(bool) {
		err = p.response(response, func(p *UnofficialApiProcess, a string) bool {
			data := p.streamChatProcess(a)
			if _, err = p.GetConversation().GinContext.Writer.Write([]byte(data)); err != nil {
				logger.Log.Fatal(err)
				return true
			}
			p.GetConversation().GinContext.Writer.Flush()
			return false
		})
		if err != nil {
			return err
		}
	} else {
		err = p.response(response, func(p *UnofficialApiProcess, a string) bool {
			data := p.jsonChatProcess(a)
			if data != nil {
				p.GetConversation().GinContext.Header("Content-Type", "application/json")
				p.GetConversation().GinContext.JSON(response.StatusCode, data)
				return true
			}
			return false
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *UnofficialApiProcess) MakeRequest(requestBody map[string]interface{}) (*http.Response, error) {
	reqModel, err := p.checkModel(p.Model)
	if err != nil {
		p.GetConversation().GinContext.JSON(400, gin.H{"error": err.Error()})
		return nil, err
	}
	req := common.GetChatReqStr(reqModel)
	if err = p.generateBody(req, requestBody); err != nil {
		return nil, err
	}
	jsonData, _ := json.Marshal(req)
	var requestData map[string]interface{}
	err = json.Unmarshal(jsonData, &requestData)
	if err != nil {
		p.GetConversation().GinContext.JSON(400, gin.H{"error": err.Error()})
		return nil, err
	}
	request, err := p.createRequest(requestData) //创建请求
	if err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Server error"})
		return nil, err
	}
	response, err := p.GetConversation().RequestClient.Do(request)        //发送请求
	process.CopyResponseHeaders(response, p.GetConversation().GinContext) //设置响应头
	if err != nil {
		var responseBody interface{}
		err = json.NewDecoder(response.Body).Decode(&responseBody)
		if err != nil {
			p.GetConversation().GinContext.JSON(500, gin.H{"error": "Request json decode error"})
			return nil, err
		}
		p.GetConversation().GinContext.JSON(response.StatusCode, responseBody)
		return nil, err
	}
	return response, nil
}

func (p *UnofficialApiProcess) createRequest(requestBody map[string]interface{}) (*http.Request, error) {
	logger.Log.Debug("UnofficialApiProcess createRequest")
	err := p.addArkoseTokenIfNeeded(&requestBody)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	var request *http.Request
	if p.Conversation.RequestBody == shttp.NoBody {
		request, err = http.NewRequest(p.Conversation.RequestMethod, p.Conversation.RequestUrl, nil)
	} else {
		request, err = http.NewRequest(p.Conversation.RequestMethod, p.Conversation.RequestUrl, bytes.NewBuffer(bodyBytes))
	}
	if err != nil {
		return nil, err
	}
	p.buildHeaders(request)
	p.setCookies(request)
	return request, nil
}
func (p *UnofficialApiProcess) setCookies(request *http.Request) {
	logger.Log.Debug("UnofficialApiProcess setCookies")
	for _, cookie := range p.GetConversation().GinContext.Request.Cookies() {
		request.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}
}
func (p *UnofficialApiProcess) buildHeaders(request *http.Request) {
	logger.Log.Debug("UnofficialApiProcess buildHeaders")
	headers := map[string]string{
		"Host":          common.Env.OpenaiHost,
		"Origin":        "https://" + common.Env.OpenaiHost + "/chat",
		"Authorization": p.GetConversation().GinContext.Request.Header.Get("Authorization"),
		"Connection":    "keep-alive",
		"User-Agent":    common.Env.UserAgent,
		"Content-Type":  p.GetConversation().GinContext.Request.Header.Get("Content-Type"),
	}

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	if puid := p.GetConversation().GinContext.Request.Header.Get("PUID"); puid != "" {
		request.Header.Set("cookie", "_puid="+puid+";")
	}
}
func (p *UnofficialApiProcess) addArkoseTokenIfNeeded(requestBody *map[string]interface{}) error {
	logger.Log.Debug("UnofficialApiProcess addArkoseTokenIfNeeded")
	model, exists := (*requestBody)["model"]
	if !exists {
		return nil
	}
	if strings.HasPrefix(model.(string), "gpt-4") || common.Env.ArkoseMust {
		token, err := funcaptcha.GetOpenAIArkoseToken(funcaptcha.ArkVerChat4, p.GetConversation().RequestHeaders.Get("puid"))
		if err != nil {
			p.GetConversation().GinContext.JSON(500, gin.H{"error": "Get ArkoseToken Failed"})
			return err
		}
		(*requestBody)["arkose_token"] = token
	}
	return nil
}
func (p *UnofficialApiProcess) streamChatProcess(raw string) string {
	result := p.getStreamResp(raw)
	if strings.Contains(raw, "[DONE]") {
		return "data: " + raw + "\n\n"
	} else if result.Pass {
		return ""
	} else if result.ApiRespStrStreamEnd.Id != "" {
		data, err := json.Marshal(result.ApiRespStrStreamEnd)
		if err != nil {
			logger.Log.Fatal(err)
		}
		return "data: " + string(data) + "\n\n"
	} else if result.ApiRespStrStream.Id != "" {
		data, err := json.Marshal(result.ApiRespStrStream)
		if err != nil {
			logger.Log.Fatal(err)
		}
		return "data: " + string(data) + "\n\n"
	}
	return ""
}

func (p *UnofficialApiProcess) response(response *http.Response, mid func(p *UnofficialApiProcess, a string) bool) error {
	logger.Log.Debug("UnofficialApiProcess streamResponse")
	client := tools.NewSSEClient(response.Body)
	events := client.Read()
	for event := range events {
		if event.Event == "message" {
			if mid(p, event.Data) {
				return nil
			}
		}
	}
	defer client.Close()
	return nil
}

func (p *UnofficialApiProcess) jsonChatProcess(raw string) *common.ApiRespStr {
	p.getStreamResp(raw)
	if strings.Contains(raw, "[DONE]") {
		resp := common.GetApiRespStr(p.ID)
		choice := common.GetStrChoices()
		choice.Message.Content = p.OldString
		resp.Choices = append(resp.Choices, *choice)
		resp.Model = p.Model
		return resp
	}
	return nil
}

func (p *UnofficialApiProcess) jsonImageProcess(stream string) {
	logger.Log.Debug("getImageResp")
	var dalleRespStr common.DALLERespStr
	json.Unmarshal([]byte(stream), &dalleRespStr)
	if dalleRespStr.Message.Author.Name == "dalle.text2im" && dalleRespStr.Message.Content.ContentType == "multimodal_text" {
		logger.Log.Debug("found image")
		for _, v := range dalleRespStr.Message.Content.Parts {
			item := new(ImagePointer)
			item.Pointer = strings.ReplaceAll(v.AssetPointer, "file-service://", "")
			item.Prompt = v.Metadata.Dalle.Prompt
			p.ImagePointerList = append(p.ImagePointerList, *item)
		}
	}
}
func (p *UnofficialApiProcess) getImageUrlByPointer(imagePointerList *[]ImagePointer, result *Result) error {
	logger.Log.Debug("getImageUrlByPointer")
	for _, v := range *imagePointerList {
		imageDownloadUrl := new(common.ImageDownloadUrl)
		getUrl := "http://" + common.Env.Host + ":" + strconv.Itoa(common.Env.Port) + "/backend-api/files/" + v.Pointer + "/download"
		logger.Log.Debug("image url is " + getUrl)
		request, err := http.NewRequest("GET", getUrl, nil)
		if err != nil {
			return err
		}
		request.Header.Set("Authorization", p.GetConversation().RequestHeaders.Get("Authorization"))
		response, err := (&http.Client{}).Do(request)
		if err != nil {
			return err
		}
		if response.Body != shttp.NoBody {
			err = json.NewDecoder(response.Body).Decode(&imageDownloadUrl)
			if err != nil {
				return err
			}
		}
		if imageDownloadUrl.DownloadUrl != "" {
			logger.Log.Debug("getDownloadUrl")
			imageItem := new(common.ApiImageItem)
			result.ApiImageGenerationRespStr.Created = time.Now().Unix()
			imageItem.Url = imageDownloadUrl.DownloadUrl
			imageItem.RevisedPrompt = v.Prompt
			result.ApiImageGenerationRespStr.Data = append(result.ApiImageGenerationRespStr.Data, *imageItem)
		}
		response.Body.Close()
	}
	return nil
}

func (p *UnofficialApiProcess) getStreamResp(stream string) *Result {
	logger.Log.Debug("getStreamResp")
	var chatRespStr common.ChatRespStr
	var chatEndRespStr common.ChatEndRespStr
	result := new(Result)
	result.ApiRespStrStreamEnd = common.ApiRespStrStreamEnd{}
	result.ApiRespStrStream = common.ApiRespStrStream{}
	result.Pass = false
	json.Unmarshal([]byte(stream), &chatRespStr)
	if chatRespStr.Message.Id != "" {
		if chatRespStr.Message.Metadata.ParentId == "" {
			result.Pass = true
			return result
		}
		logger.Log.Debug("chatRespStr")
		resp := common.GetApiRespStrStream(p.ID)
		choice := common.GetStreamChoice()
		resp.Model = p.Model
		choice.Delta.Content = strings.ReplaceAll(chatRespStr.Message.Content.Parts[0], p.OldString, "")
		p.OldString = chatRespStr.Message.Content.Parts[0]
		resp.Choices = resp.Choices[:0]
		resp.Choices = append(resp.Choices, *choice)
		result.ApiRespStrStream = *resp
	}
	json.Unmarshal([]byte(stream), &chatEndRespStr)
	if chatEndRespStr.IsCompletion {
		logger.Log.Debug("chatEndRespStr")
		resp := common.GetApiRespStrStreamEnd(p.ID)
		resp.Model = p.Model
		result.ApiRespStrStreamEnd = *resp
	}
	if result.ApiRespStrStream.Id == "" && result.ApiRespStrStreamEnd.Id == "" {
		result.Pass = true
	}
	return result
}
func (p *UnofficialApiProcess) checkModel(model string) (string, error) {
	logger.Log.Debug("UnofficialApiProcess checkModel")
	if strings.HasPrefix(model, "dall-e") || strings.HasPrefix(model, "gpt-4-vision") {
		return "gpt-4", nil
	} else if strings.HasPrefix(model, "gpt-3") {
		return "text-davinci-002-render-sha", nil
	} else if strings.HasPrefix(model, "gpt-4") {
		return "gpt-4-gizmo", nil
	} else {
		return "", errors.New("unsupported model")
	}
}
func (p *UnofficialApiProcess) generateBody(req *common.ChatReqStr, requestBody map[string]interface{}) error {
	logger.Log.Debug("UnofficialApiProcess generateBody")
	if p.Mode == "chat" {
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
				reqMessage := common.GetChatReqTemplate()
				reqMessage.Content.Parts = reqMessage.Content.Parts[:0]
				reqMessage.Author.Role = role
				reqMessage.Content.Parts = append(reqMessage.Content.Parts, content)
				req.Messages = append(req.Messages, *reqMessage)
			}
			if _, ok := messageItem["content"].([]map[string]interface{}); ok {
				reqFileMessage := common.GetChatFileReqTemplate()
				content, _ := messageItem["content"].([]map[string]interface{})
				reqFileMessage.Content.Parts = reqFileMessage.Content.Parts[:0]
				reqFileMessage.Author.Role = role
				p.fileReqProcess(&content, &reqFileMessage.Content.Parts)
				//reqMessage.Content.Parts = append(reqMessage.Content.Parts, content)
				//req.Messages = append(req.Messages, *reqFileMessage)
			}
		}
	}
	if p.Mode == "image" {
		prompt, exists := requestBody["prompt"]
		if !exists {
			return errors.New("please provide prompt")
		}
		count, exists := requestBody["n"]
		if !exists {
			count = 1
		}
		size, exists := requestBody["size"]
		if !exists {
			size = "1024x1024"
		}
		reqMessage := common.GetChatReqTemplate()
		reqMessage.Content.Parts = reqMessage.Content.Parts[:0]
		reqMessage.Author.Role = "user"
		reqMessage.Content.Parts = append(reqMessage.Content.Parts, fmt.Sprintf("Requirements for image generation:\n- ImageCount: %s\n- Size: %s\n- Prompt:  [%s]\n- Requirements: Using the DALLE tool, each image is generated according to the number of ImageCount. It is not allowed to contain multiple elements in one image. You must call the tool multiple times to generate the number of ImageCount images, and the details of each image are different\n", count, size, prompt))
		req.Messages = append(req.Messages, *reqMessage)
	}

	return nil
}
func (p *UnofficialApiProcess) fileReqProcess(content *[]map[string]interface{}, part *[]interface{}) {

}
