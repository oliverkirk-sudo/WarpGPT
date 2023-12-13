package process

import (
	"WarpGPT/pkg/common"
	"encoding/json"
	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Process struct {
	Context common.Context
}

type ContextProcessor interface {
	SetContext(conversation common.Context)
	GetContext() common.Context
	ProcessMethod()
}

func Do(p ContextProcessor, conversation common.Context) {
	p.SetContext(conversation)
	p.ProcessMethod()
}

func DecodeRequestBody(p ContextProcessor, requestBody *map[string]interface{}) error {
	conversation := p.GetContext()
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
