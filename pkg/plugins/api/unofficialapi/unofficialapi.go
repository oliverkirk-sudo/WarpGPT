package unofficialapi

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/funcaptcha"
	"WarpGPT/pkg/plugins"
	"WarpGPT/pkg/tools"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/pkoukk/tiktoken-go"
	"io"
	shttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var context *plugins.Component
var UnofficialApiProcessInstance UnofficialApiProcess
var tke, _ = tiktoken.GetEncoding("cl100k_base")

type Context struct {
	GinContext     *gin.Context
	RequestUrl     string
	RequestClient  tls_client.HttpClient
	RequestBody    io.ReadCloser
	RequestParam   string
	RequestMethod  string
	RequestHeaders http.Header
}
type UnofficialApiProcess struct {
	Context          Context
	ID               string
	Model            string
	PromptTokens     int
	CompletionTokens int
	OldString        string
	Mode             string
	ImagePointerList []ImagePointer
}
type ImagePointer struct {
	Pointer string
	Prompt  string
}
type Result struct {
	ApiRespStrStream          ApiRespStrStream
	ApiRespStrStreamEnd       ApiRespStrStreamEnd
	ApiImageGenerationRespStr ApiImageGenerationRespStr
	Pass                      bool
}

func (p *UnofficialApiProcess) SetContext(conversation Context) {
	p.Context = conversation
}
func (p *UnofficialApiProcess) GetContext() Context {
	return p.Context
}

func (p *UnofficialApiProcess) ProcessMethod() {
	context.Logger.Debug("UnofficialApiProcess")
	var requestBody map[string]interface{}
	err := p.decodeRequestBody(&requestBody)
	if err != nil {
		return
	}
	p.ID = IdGenerator()
	_, exists := requestBody["model"]
	if exists {
		p.Model, _ = requestBody["model"].(string)
	} else {
		p.GetContext().GinContext.JSON(400, gin.H{"error": "Model not provided"})
		return
	}
	if strings.HasSuffix(p.GetContext().RequestParam, "chat/completions") {
		p.Mode = "chat"
		if err = p.chatApiProcess(requestBody); err != nil {
			println(err.Error())
			return
		}
	}
	if strings.HasSuffix(p.GetContext().RequestParam, "images/generations") {
		p.Mode = "image"
		if err = p.imageApiProcess(requestBody); err != nil {
			println(err.Error())
			return
		}
	}
}

func (p *UnofficialApiProcess) imageApiProcess(requestBody map[string]interface{}) error {
	context.Logger.Debug("UnofficialApiProcess imageApiProcess")
	response, err := p.MakeRequest(requestBody)
	if err != nil {
		return err
	}
	result := new(Result)
	result.ApiImageGenerationRespStr = ApiImageGenerationRespStr{}
	err = p.response(response, func(p *UnofficialApiProcess, a string) bool {
		p.jsonImageProcess(a)
		return false
	})
	if err = p.getImageUrlByPointer(&p.ImagePointerList, result); err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "get image url failed"})
		context.Logger.Warning(err)
	}
	if result.ApiImageGenerationRespStr.Created != 0 {
		p.GetContext().GinContext.Header("Content-Type", "application/json")
		p.GetContext().GinContext.JSON(response.StatusCode, result.ApiImageGenerationRespStr)
	}
	if err != nil {
		return err
	}
	return nil
}

func (p *UnofficialApiProcess) chatApiProcess(requestBody map[string]interface{}) error {
	context.Logger.Debug("UnofficialApiProcess chatApiProcess")
	response, err := p.MakeRequest(requestBody)
	if err != nil {
		return err
	}
	value, exists := requestBody["stream"]

	if exists && value.(bool) {
		err = p.response(response, func(p *UnofficialApiProcess, a string) bool {
			data := p.streamChatProcess(a)
			if _, err = p.GetContext().GinContext.Writer.Write([]byte(data)); err != nil {
				context.Logger.Warning(err)
				return true
			}
			p.GetContext().GinContext.Writer.Flush()
			return false
		})
		if err != nil {
			return err
		}
	} else {
		err = p.response(response, func(p *UnofficialApiProcess, a string) bool {
			data := p.jsonChatProcess(a)
			if data != nil {
				context.Logger.Debug("Counting the number of tokens")
				p.CompletionTokens = len(tke.Encode(data.Choices[0].Message.Content, nil, nil))
				data.Usage.PromptTokens = p.PromptTokens
				data.Usage.CompletionTokens = p.CompletionTokens
				data.Usage.TotalTokens = p.PromptTokens + p.CompletionTokens
				p.GetContext().GinContext.Header("Content-Type", "application/json")
				p.GetContext().GinContext.JSON(response.StatusCode, data)
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
		p.GetContext().GinContext.JSON(400, gin.H{"error": err.Error()})
		return nil, err
	}
	req := GetChatReqStr(reqModel)
	if err = p.generateBody(req, requestBody); err != nil {
		return nil, err
	}
	jsonData, _ := json.Marshal(req)
	var requestData map[string]interface{}
	err = json.Unmarshal(jsonData, &requestData)
	if err != nil {
		p.GetContext().GinContext.JSON(400, gin.H{"error": err.Error()})
		return nil, err
	}
	request, err := p.createRequest(requestData) //创建请求
	if err != nil {
		return nil, err
	}
	response, err := p.GetContext().RequestClient.Do(request)       //发送请求
	common.CopyResponseHeaders(response, p.GetContext().GinContext) //设置响应头
	if err != nil {
		var responseBody interface{}
		err = json.NewDecoder(response.Body).Decode(&responseBody)
		if err != nil {
			p.GetContext().GinContext.JSON(500, gin.H{"error": "Request json decode error"})
			return nil, err
		}
		p.GetContext().GinContext.JSON(response.StatusCode, responseBody)
		return nil, err
	}
	return response, nil
}

func (p *UnofficialApiProcess) createRequest(requestBody map[string]interface{}) (*http.Request, error) {
	context.Logger.Debug("UnofficialApiProcess createRequest")
	token, err := p.addArkoseTokenIfNeeded(&requestBody)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	var request *http.Request
	if p.Context.RequestBody == shttp.NoBody {
		request, err = http.NewRequest(p.Context.RequestMethod, p.Context.RequestUrl, nil)
	} else {
		request, err = http.NewRequest(p.Context.RequestMethod, p.Context.RequestUrl, bytes.NewBuffer(bodyBytes))
	}
	if err != nil {
		return nil, err
	}
	p.addArkoseTokenInHeaderIfNeeded(request, token)
	p.buildHeaders(request)
	p.setCookies(request)
	return request, nil
}
func (p *UnofficialApiProcess) setCookies(request *http.Request) {
	context.Logger.Debug("UnofficialApiProcess setCookies")
	for _, cookie := range p.GetContext().GinContext.Request.Cookies() {
		request.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}
}
func (p *UnofficialApiProcess) buildHeaders(request *http.Request) {
	context.Logger.Debug("UnofficialApiProcess buildHeaders")
	headers := map[string]string{
		"Host":          context.Env.OpenaiHost,
		"Origin":        "https://" + context.Env.OpenaiHost + "/chat",
		"Authorization": p.GetContext().GinContext.Request.Header.Get("Authorization"),
		"Connection":    "keep-alive",
		"User-Agent":    context.Env.UserAgent,
		"Content-Type":  p.GetContext().GinContext.Request.Header.Get("Content-Type"),
	}

	for key, value := range headers {
		request.Header.Set(key, value)
	}

	if puid := p.GetContext().GinContext.Request.Header.Get("PUID"); puid != "" {
		request.Header.Set("cookie", "_puid="+puid+";")
	}
}
func (p *UnofficialApiProcess) addArkoseTokenInHeaderIfNeeded(request *http.Request, token string) {
	context.Logger.Debug("UnofficialApiProcess addArkoseTokenInHeaderIfNeeded")
	request.Header.Set("Openai-Sentinel-Arkose-Token", token)
}
func (p *UnofficialApiProcess) addArkoseTokenIfNeeded(requestBody *map[string]interface{}) (string, error) {
	context.Logger.Debug("UnofficialApiProcess addArkoseTokenIfNeeded")
	model, exists := (*requestBody)["model"]
	if !exists {
		return "", nil
	}
	if strings.HasPrefix(model.(string), "gpt-4") || context.Env.ArkoseMust {
		token, err := funcaptcha.GetOpenAIArkoseToken(4, p.GetContext().RequestHeaders.Get("puid"))
		if err != nil {
			p.GetContext().GinContext.JSON(500, gin.H{"error": "Get ArkoseToken Failed"})
			return "", err
		}
		(*requestBody)["arkose_token"] = token
		return token, nil
	}
	return "", nil
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
			context.Logger.Warning(err)
		}
		return "data: " + string(data) + "\n\n"
	} else if result.ApiRespStrStream.Id != "" {
		data, err := json.Marshal(result.ApiRespStrStream)
		if err != nil {
			context.Logger.Warning(err)
		}
		return "data: " + string(data) + "\n\n"
	}
	return ""
}

func (p *UnofficialApiProcess) response(response *http.Response, mid func(p *UnofficialApiProcess, a string) bool) error {
	context.Logger.Debug("UnofficialApiProcess streamResponse")
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

func (p *UnofficialApiProcess) jsonChatProcess(raw string) *ApiRespStr {
	p.getStreamResp(raw)
	if strings.Contains(raw, "[DONE]") {
		resp := GetApiRespStr(p.ID)
		choice := GetStrChoices()
		choice.Message.Content = p.OldString
		resp.Choices = append(resp.Choices, *choice)
		resp.Model = p.Model
		return resp
	}
	return nil
}

func (p *UnofficialApiProcess) jsonImageProcess(stream string) {
	context.Logger.Debug("getImageResp")
	var dalleRespStr DALLERespStr
	json.Unmarshal([]byte(stream), &dalleRespStr)
	if dalleRespStr.Message.Author.Name == "dalle.text2im" && dalleRespStr.Message.Content.ContentType == "multimodal_text" {
		context.Logger.Debug("found image")
		for _, v := range dalleRespStr.Message.Content.Parts {
			item := new(ImagePointer)
			item.Pointer = strings.ReplaceAll(v.AssetPointer, "file-service://", "")
			item.Prompt = v.Metadata.Dalle.Prompt
			p.ImagePointerList = append(p.ImagePointerList, *item)
		}
	}
}
func (p *UnofficialApiProcess) getImageUrlByPointer(imagePointerList *[]ImagePointer, result *Result) error {
	context.Logger.Debug("getImageUrlByPointer")
	for _, v := range *imagePointerList {
		imageDownloadUrl := new(ImageDownloadUrl)
		getUrl := "http://" + context.Env.Host + ":" + strconv.Itoa(context.Env.Port) + "/backend-api/files/" + v.Pointer + "/download"
		context.Logger.Debug("image url is " + getUrl)
		request, err := http.NewRequest("GET", getUrl, nil)
		if err != nil {
			return err
		}
		request.Header.Set("Authorization", p.GetContext().RequestHeaders.Get("Authorization"))
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
			context.Logger.Debug("getDownloadUrl")
			imageItem := new(ApiImageItem)
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
	context.Logger.Debug("getStreamResp")
	var chatRespStr ChatRespStr
	var chatEndRespStr ChatEndRespStr
	result := new(Result)
	result.ApiRespStrStreamEnd = ApiRespStrStreamEnd{}
	result.ApiRespStrStream = ApiRespStrStream{}
	result.Pass = false
	json.Unmarshal([]byte(stream), &chatRespStr)
	if chatRespStr.Message.Id != "" {
		if chatRespStr.Message.Metadata.ParentId == "" {
			result.Pass = true
			return result
		}
		context.Logger.Debug("chatRespStr")
		resp := GetApiRespStrStream(p.ID)
		choice := GetStreamChoice()
		resp.Model = p.Model
		choice.Delta.Content = strings.ReplaceAll(chatRespStr.Message.Content.Parts[0], p.OldString, "")
		p.OldString = chatRespStr.Message.Content.Parts[0]
		resp.Choices = resp.Choices[:0]
		resp.Choices = append(resp.Choices, *choice)
		result.ApiRespStrStream = *resp
	}
	json.Unmarshal([]byte(stream), &chatEndRespStr)
	if chatEndRespStr.IsCompletion {
		context.Logger.Debug("chatEndRespStr")
		resp := GetApiRespStrStreamEnd(p.ID)
		resp.Model = p.Model
		result.ApiRespStrStreamEnd = *resp
	}
	if result.ApiRespStrStream.Id == "" && result.ApiRespStrStreamEnd.Id == "" {
		result.Pass = true
	}
	return result
}
func (p *UnofficialApiProcess) checkModel(model string) (string, error) {
	context.Logger.Debug("UnofficialApiProcess checkModel")
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
func (p *UnofficialApiProcess) generateBody(req *ChatReqStr, requestBody map[string]interface{}) error {
	context.Logger.Debug("UnofficialApiProcess generateBody")
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
				p.PromptTokens += len(tke.Encode(content, nil, nil)) + 7
				reqMessage := GetChatReqTemplate()
				reqMessage.Content.Parts = reqMessage.Content.Parts[:0]
				reqMessage.Author.Role = role
				reqMessage.Content.Parts = append(reqMessage.Content.Parts, content)
				req.Messages = append(req.Messages, *reqMessage)
			}
			if _, ok := messageItem["content"].([]map[string]interface{}); ok {
				reqFileMessage := GetChatFileReqTemplate()
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
		reqMessage := GetChatReqTemplate()
		reqMessage.Content.Parts = reqMessage.Content.Parts[:0]
		reqMessage.Author.Role = "user"
		reqMessage.Content.Parts = append(reqMessage.Content.Parts, fmt.Sprintf("Requirements for image generation:\n- ImageCount: %s\n- Size: %s\n- Prompt:  [%s]\n- Requirements: Using the DALLE tool, each image is generated according to the number of ImageCount. It is not allowed to contain multiple elements in one image. You must call the tool multiple times to generate the number of ImageCount images, and the details of each image are different\n", count, size, prompt))
		req.Messages = append(req.Messages, *reqMessage)
	}

	return nil
}
func (p *UnofficialApiProcess) fileReqProcess(content *[]map[string]interface{}, part *[]interface{}) {

}

func (p *UnofficialApiProcess) decodeRequestBody(requestBody *map[string]interface{}) error {
	conversation := p.GetContext()
	if conversation.RequestBody != shttp.NoBody {
		if err := json.NewDecoder(conversation.RequestBody).Decode(requestBody); err != nil {
			conversation.GinContext.JSON(400, gin.H{"error": "JSON invalid"})
			return err
		}
	}
	return nil
}

type UnOfficialApiRequestUrl struct {
}

func (u UnOfficialApiRequestUrl) Generate(path string, rawquery string) string {
	if rawquery == "" {
		return "https://" + context.Env.OpenaiHost + "/backend-api" + "/conversation"
	}
	return "https://" + context.Env.OpenaiHost + "/backend-api" + "/conversation" + "?" + rawquery
}
func (p *UnofficialApiProcess) Run(com *plugins.Component) {
	context = com
	context.Engine.Any("/r/v1/*path", func(c *gin.Context) {
		conversation := common.GetContextPack(c, UnOfficialApiRequestUrl{})
		common.Do[Context](new(UnofficialApiProcess), Context(conversation))
	})
}
