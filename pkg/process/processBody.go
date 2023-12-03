package process

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/requestbody"
	"WarpGPT/pkg/tools"
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	"log"
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
	if conversation.RequestBody != nil {
		if err := json.NewDecoder(conversation.RequestBody).Decode(requestBody); err != nil {
			conversation.GinContext.JSON(400, gin.H{"error": "JSON invalid"})
			return err
		}
	}
	return nil
}

func ProcessConversationRequest(p ProcessInterface, requestBody *map[string]interface{}) error {
	if _, modelExists := (*requestBody)["model"]; modelExists {
		if err := addArkoseTokenIfNeeded(p, requestBody); err != nil {
			return err
		}
	}

	return makeRequest(p, requestBody)
}

func addArkoseTokenIfNeeded(p ProcessInterface, requestBody *map[string]interface{}) error {
	model := (*requestBody)["model"].(string)
	if strings.HasPrefix(model, "gpt-4") {
		token, err := tools.NewAuthenticator("", "").GetLoginArkoseToken()
		if err != nil {
			p.GetConversation().GinContext.JSON(500, gin.H{"error": "Get ArkoseToken Failed"})
			return err.Error
		}
		(*requestBody)["arkose_token"] = token.Token
	}
	return nil
}

func ProcessRegularRequest(p ProcessInterface, requestBody *map[string]interface{}) error {
	return makeRequest(p, requestBody)
}

func makeRequest(p ProcessInterface, requestBody *map[string]interface{}) error {
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Error encoding JSON"})
		return err
	}

	request, err := http.NewRequest(p.GetConversation().RequestMethod, p.GetConversation().RequestUrl, bytes.NewBuffer(bodyBytes))
	if err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Server error"})
		return err
	}

	BuildHeaders(p.GetConversation().GinContext, request)
	response, err := p.GetConversation().RequestClient.Do(request)
	if err != nil {
		p.GetConversation().GinContext.JSON(500, gin.H{"error": "Return Error"})
		return err
	}
	defer response.Body.Close()

	CopyResponseHeaders(response, p.GetConversation().GinContext)

	if err := streamResponse(p, response); err != nil {
		return err
	}
	return nil
}

func streamResponse(p ProcessInterface, response *http.Response) error {
	buf := make([]byte, 1024)
	for {
		n, err := response.Body.Read(buf)
		if n > 0 {
			p.GetConversation().GinContext.Writer.Write(buf[:n])
			p.GetConversation().GinContext.Writer.Flush()
		}
		if err != nil {
			return err
		}
	}
}
func CopyResponseHeaders(response *http.Response, ctx *gin.Context) {
	for name, values := range response.Header {
		if name == "Content-Encoding" {
			continue
		}
		for _, value := range values {
			ctx.Header(name, value)
		}
	}
}

func BuildHeaders(c *gin.Context, request *http.Request) {
	log.Println("build_headers")
	request.Header.Set("Host", common.Env.OpenAI_HOST)
	request.Header.Set("Origin", "https://"+common.Env.OpenAI_HOST+"/chat")
	request.Header.Set("Authorization", c.Request.Header.Get("Authorization"))
	request.Header.Set("user-agent", common.Env.UserAgent)
	if c.Request.Header.Get("PUID") != "" {
		request.Header.Set("cookie", "_puid="+c.Request.Header.Get("PUID")+";")
	}
}
