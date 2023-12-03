package session

import (
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/process"
	"WarpGPT/pkg/requestbody"
	"WarpGPT/pkg/tools"
	"github.com/gin-gonic/gin"
)

type SessionToken struct {
	process.Process
}

func (p *SessionToken) GetConversation() requestbody.Conversation {
	return p.Conversation
}
func (p *SessionToken) SetConversation(conversation requestbody.Conversation) {
	p.Conversation = conversation
}

func (p *SessionToken) ProcessMethod() {
	logger.Log.Debug("SessionToken")
	var requestBody map[string]interface{}
	if err := process.DecodeRequestBody(p, &requestBody); err != nil {
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
				p.GetConversation().GinContext.JSON(400, err)
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
			p.GetConversation().GinContext.JSON(200, result)
		} else {
			p.GetConversation().GinContext.JSON(400, gin.H{"error": "Please provide a refreshCookie or username and password."})
			return
		}
	} else {
		auth = tools.NewAuthenticator("", "", "")
		err := auth.GetAccessTokenByRefreshToken(refreshCookie.(string))
		if err != nil {
			p.GetConversation().GinContext.JSON(400, err)
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
		p.GetConversation().GinContext.JSON(200, result)
	}
}
