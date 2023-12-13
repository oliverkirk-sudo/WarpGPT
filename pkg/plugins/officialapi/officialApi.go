package officialapi

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/tools"
	"bytes"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	"github.com/gin-gonic/gin"
	shttp "net/http"
	"strings"
)

type OfficialApiProcess struct {
	common.Process
}

func (p *OfficialApiProcess) SetContext(conversation common.Context) {
	p.Context = conversation
}
func (p *OfficialApiProcess) GetContext() common.Context {
	return p.Context
}
func (p *OfficialApiProcess) ProcessMethod() {
	var requestBody map[string]interface{}
	err := common.DecodeRequestBody(p, &requestBody) //解析请求体
	if err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Incorrect json format"})
		return
	}

	request, err := p.createRequest(requestBody) //创建请求
	if err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Server error"})
		return
	}

	response, err := p.GetContext().RequestClient.Do(request) //发送请求
	if err != nil {
		p.GetContext().GinContext.JSON(500, gin.H{"error": "Server Error"})
		return
	}

	common.CopyResponseHeaders(response, p.GetContext().GinContext) //设置响应头

	if strings.Contains(response.Header.Get("Content-Type"), "text/event-stream") {
		err = p.streamResponse(response)
		if err != nil {
			return
		}
	}
	if strings.Contains(response.Header.Get("Content-Type"), "application/json") {
		err = p.jsonResponse(response)
		if err != nil {
			logger.Log.Fatal(err)
		}
	}
}

func (p *OfficialApiProcess) createRequest(requestBody map[string]interface{}) (*http.Request, error) {
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
	p.WithHeaders(request)
	return request, nil
}

func (p *OfficialApiProcess) WithHeaders(rsq *http.Request) {
	rsq.Header.Set("Authorization", p.Context.RequestHeaders.Get("Authorization"))
	rsq.Header.Set("Content-Type", p.Context.RequestHeaders.Get("Content-Type"))
}

func (p *OfficialApiProcess) jsonResponse(response *http.Response) error {
	var jsonData interface{}
	err := json.NewDecoder(response.Body).Decode(&jsonData)
	if err != nil {
		return err
	}
	p.GetContext().GinContext.JSON(response.StatusCode, jsonData)
	return nil
}

func (p *OfficialApiProcess) streamResponse(response *http.Response) error {
	logger.Log.Infoln("officialApiProcess stream Request")
	client := tools.NewSSEClient(response.Body)
	events := client.Read()
	for event := range events {
		if _, err := p.GetContext().GinContext.Writer.Write([]byte("data: " + event.Data + "\n\n")); err != nil {
			return err
		}
		p.GetContext().GinContext.Writer.Flush()
	}
	defer client.Close()
	return nil
}
