package process

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/requestbody"
	"WarpGPT/pkg/tools"
	"bytes"
	"encoding/json"
	"fmt"
	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strings"
)

type Process struct {
	Conversation requestbody.Conversation
}

type ConversationProcessor interface {
	SetConversation(conversation requestbody.Conversation)
	ProcessMethod()
}

func Do(p ConversationProcessor, conversation requestbody.Conversation) {
	p.SetConversation(conversation)
	p.ProcessMethod()
}

type ProcessInterface interface {
	GetConversation() requestbody.Conversation
}

func DecodeRequestBody(p ProcessInterface, requestBody *map[string]interface{}) error {
	conversation := p.GetConversation()
	if conversation.RequestBody != http.NoBody {
		if err := json.NewDecoder(conversation.RequestBody).Decode(requestBody); err != nil {
			conversation.GinContext.JSON(400, gin.H{"error": "JSON invalid"})
			return err
		}
	}
	return nil
}

func ProcessConversationRequest(p ProcessInterface, requestBody *map[string]interface{}, mid func(a string) string) error {
	if _, modelExists := (*requestBody)["model"]; modelExists {
		if err := addArkoseTokenIfNeeded(p, requestBody); err != nil {
			return err
		}
	}
	response, err := makeRequest(p, requestBody)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	err = StreamResponse(p, response, mid)
	if err != nil {
		return err
	}
	return nil
}

func addArkoseTokenIfNeeded(p ProcessInterface, requestBody *map[string]interface{}) error {
	model := (*requestBody)["model"].(string)
	if strings.HasPrefix(model, "gpt-4") {
		token, err := tools.NewAuthenticator("", "", "").GetLoginArkoseToken()
		if err != nil {
			p.GetConversation().GinContext.JSON(500, gin.H{"error": "Get ArkoseToken Failed"})
			return err.Error
		}
		(*requestBody)["arkose_token"] = token.Token
	}
	return nil
}

func ProcessRegularRequest(p ProcessInterface, requestBody *map[string]interface{}) error {
	logger.Log.Debug("Json Request")
	resp, err := makeRequest(p, requestBody)
	if err != nil {
		return err
	}
	err = SendJsonResponse(p, resp)
	if err != nil {
		return err
	}
	return nil
}

func makeRequest(p ProcessInterface, requestBody *map[string]interface{}) (*fhttp.Response, error) {
	logger.Log.Debug("makeRequest")
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Error encoding JSON"})
		return nil, err
	}
	var request *fhttp.Request
	if len(*requestBody) == 0 {
		logger.Log.Debug("Empty requestBody")
		request, _ = fhttp.NewRequest(p.GetConversation().RequestMethod, p.GetConversation().RequestUrl, nil)
	} else {
		logger.Log.Debug("RequestBody")
		request, _ = fhttp.NewRequest(p.GetConversation().RequestMethod, p.GetConversation().RequestUrl, bytes.NewBuffer(bodyBytes))
	}
	SetCookies(p.GetConversation().GinContext, request)
	BuildHeaders(p.GetConversation().GinContext, request)
	response, err := p.GetConversation().RequestClient.Do(request)
	if err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Return Error"})
		return nil, err
	}

	CopyResponseHeaders(response, p.GetConversation().GinContext)
	return response, nil
}

func StreamResponse(p ProcessInterface, response *fhttp.Response, mid func(a string) string) error {
	logger.Log.Infoln("Stream Request")
	defer response.Body.Close()

	var accumulatedData strings.Builder

Loop:
	for {
		buf := make([]byte, 1024)
		n, err := response.Body.Read(buf)
		if n > 0 {
			accumulatedData.Write(buf[:n])
			if strings.HasSuffix(accumulatedData.String(), "\n\n") {
				_, err := p.GetConversation().GinContext.Writer.Write([]byte(mid(accumulatedData.String())))
				if err != nil {
					return err
				}
				p.GetConversation().GinContext.Writer.Flush()
				accumulatedData.Reset()
			}
		}
		if err != nil {
			if err == io.EOF {
				break Loop
			}
			return err
		}

		if strings.Contains(accumulatedData.String(), "[DONE]") {
			break
		}

		select {
		case <-p.GetConversation().GinContext.Writer.CloseNotify():
			break Loop
		default:
			continue
		}
	}
	return nil
}

func CopyResponseHeaders(response *fhttp.Response, ctx *gin.Context) {
	for name, values := range response.Header {
		if name == "Content-Encoding" {
			continue
		}
		for _, value := range values {
			ctx.Header(name, value)
		}
	}
}

func BuildHeaders(c *gin.Context, request *fhttp.Request) {
	log.Println("build_headers")
	request.Header.Set("Host", common.Env.OpenAI_HOST)
	request.Header.Set("Origin", "https://"+common.Env.OpenAI_HOST+"/chat")
	request.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("User-Agent", common.Env.UserAgent)
	request.Header.Set("Content-Type", c.Request.Header.Get("Content-Type"))
	if c.Request.Header.Get("PUID") != "" {
		request.Header.Set("cookie", "_puid="+c.Request.Header.Get("PUID")+";")
	}
}

func SendJsonResponse(p ProcessInterface, response *fhttp.Response) error {
	var responseBody map[string]interface{}
	defer response.Body.Close()
	err := json.NewDecoder(response.Body).Decode(&responseBody)
	if err != nil {
		p.GetConversation().GinContext.JSON(response.StatusCode, response.Body)
		return err
	}
	p.GetConversation().GinContext.JSON(response.StatusCode, responseBody)
	return nil
}

func SetCookies(c *gin.Context, request *fhttp.Request) {
	fhttpCookies := make([]*fhttp.Cookie, len(c.Request.Cookies()))
	for i, cookie := range c.Request.Cookies() {
		fhttpCookies[i] = &fhttp.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		}
		fmt.Printf("%+v", cookie)
	}
	for _, cookie := range fhttpCookies {
		request.AddCookie(cookie)
	}
}
