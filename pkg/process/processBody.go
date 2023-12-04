package process

import (
	"WarpGPT/pkg/requestbody"
	"encoding/json"
	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	"net/http"
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

func CopyResponseHeaders(response *fhttp.Response, ctx *gin.Context) {
	skipHeaders := map[string]bool{"Content-Encoding": true, "Content-Length": true, "transfer-encoding": true, "connection": true}
	for name, values := range response.Header {
		if !skipHeaders[name] {
			for _, value := range values {
				ctx.Header(name, value)
			}
		}
	}
}
