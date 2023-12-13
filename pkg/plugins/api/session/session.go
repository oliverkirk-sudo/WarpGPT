package session

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/plugins"
	"WarpGPT/pkg/tools"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/gin-gonic/gin"
	"io"
	shttp "net/http"
)

var context *plugins.Component
var SessionTokenInstance SessionToken

type Context struct {
	GinContext     *gin.Context
	RequestUrl     string
	RequestClient  tls_client.HttpClient
	RequestBody    io.ReadCloser
	RequestParam   string
	RequestMethod  string
	RequestHeaders http.Header
}
type SessionToken struct {
	Context Context
}

func (p *SessionToken) GetContext() Context {
	return p.Context
}
func (p *SessionToken) SetContext(conversation Context) {
	p.Context = conversation
}

func (p *SessionToken) ProcessMethod() {
	context.Logger.Debug("SessionToken")
	var requestBody map[string]interface{}
	if err := p.decodeRequestBody(&requestBody); err != nil {
		return
	}
	var auth *tools.Authenticator
	username, usernameExists := requestBody["username"]
	password, passwordExists := requestBody["password"]
	puid, puidExists := requestBody["puid"]
	refreshCookie, refreshCookieExists := requestBody["refreshCookie"]
	if !refreshCookieExists {
		if usernameExists && passwordExists {
			if puidExists {
				auth = tools.NewAuthenticator(username.(string), password.(string), puid.(string))
			} else {
				auth = tools.NewAuthenticator(username.(string), password.(string), "")
			}
			if err := auth.Begin(); err != nil {
				p.GetContext().GinContext.JSON(400, err)
				return
			}
			auth.GetModels()
			all := auth.GetAuthResult()
			var result map[string]interface{}
			accessToken := all.AccessToken
			model := all.Model
			refreshToken := all.FreshToken
			result = accessToken
			result["refreshCookie"] = refreshToken
			result["models"] = model["models"]
			p.GetContext().GinContext.JSON(200, result)
		} else {
			p.GetContext().GinContext.JSON(400, gin.H{"error": "Please provide a refreshCookie or username and password."})
			return
		}
	} else {
		auth = tools.NewAuthenticator("", "", "")
		err := auth.GetAccessTokenByRefreshToken(refreshCookie.(string))
		if err != nil {
			p.GetContext().GinContext.JSON(400, err)
			return
		}
		auth.GetModels()
		all := auth.GetAuthResult()
		var result map[string]interface{}
		accessToken := all.AccessToken
		model := all.Model
		refreshToken := all.FreshToken
		result = accessToken
		result["refreshCookie"] = refreshToken
		result["models"] = model["models"]
		p.GetContext().GinContext.JSON(200, result)
	}
}
func (p *SessionToken) decodeRequestBody(requestBody *map[string]interface{}) error {
	conversation := p.GetContext()
	if conversation.RequestBody != shttp.NoBody {
		if err := json.NewDecoder(conversation.RequestBody).Decode(requestBody); err != nil {
			conversation.GinContext.JSON(400, gin.H{"error": "JSON invalid"})
			return err
		}
	}
	return nil
}

type NotHaveUrl struct {
}

func (u NotHaveUrl) Generate(path string, rawquery string) string {
	return ""
}
func (p *SessionToken) Run(com *plugins.Component) {
	context = com
	context.Engine.POST("/getsession", func(c *gin.Context) {
		conversation := common.GetContextPack(c, NotHaveUrl{})
		common.Do[Context](new(SessionToken), Context(conversation))
	})
}
