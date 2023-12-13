package common

import (
	"WarpGPT/pkg/env"
	"github.com/gin-gonic/gin"
	"strings"
)

type RequestUrl interface {
	Generate(path string, rawquery string) string
}
type OfficialApiRequestUrl struct {
}
type UnOfficialApiRequestUrl struct {
}
type ReverseApiRequestUrl struct {
}
type ReverseBackendRequestUrl struct {
}
type ReversePublicApiRequestUrl struct {
}
type NotHaveUrl struct {
}

func (u OfficialApiRequestUrl) Generate(path string, rawquery string) string {
	if rawquery == "" {
		return "https://api.openai.com/v1" + path
	}
	return "https://api.openai.com/v1" + path + "?" + rawquery
}
func (u UnOfficialApiRequestUrl) Generate(path string, rawquery string) string {
	if rawquery == "" {
		return "https://" + env.Env.OpenaiHost + "/backend-api" + "/conversation"
	}
	return "https://" + env.Env.OpenaiHost + "/backend-api" + "/conversation" + "?" + rawquery
}
func (u ReverseApiRequestUrl) Generate(path string, rawquery string) string {
	if rawquery == "" {
		return "https://" + env.Env.OpenaiHost + "/api" + path
	}
	return "https://" + env.Env.OpenaiHost + "/api" + path + "?" + rawquery
}
func (u ReverseBackendRequestUrl) Generate(path string, rawquery string) string {
	if rawquery == "" {
		return "https://" + env.Env.OpenaiHost + "/backend-api" + path
	}
	return "https://" + env.Env.OpenaiHost + "/backend-api" + path + "?" + rawquery
}
func (u ReversePublicApiRequestUrl) Generate(path string, rawquery string) string {
	if rawquery == "" {
		return "https://" + env.Env.OpenaiHost + "/public-api" + path
	}
	return "https://" + env.Env.OpenaiHost + "/public-api" + path + "?" + rawquery
}

func (u NotHaveUrl) Generate(path string, rawquery string) string {
	return ""
}
func CheckRequest(c *gin.Context) RequestUrl {
	path := c.Request.URL.Path
	if strings.HasPrefix(path, "/backend-api") {
		return ReverseBackendRequestUrl{}
	}
	if strings.HasPrefix(path, "/api") {
		return ReverseApiRequestUrl{}
	}
	if strings.HasPrefix(path, "/public-api") {
		return ReversePublicApiRequestUrl{}
	}
	if strings.HasPrefix(path, "/v1") {
		return OfficialApiRequestUrl{}
	}
	if strings.HasPrefix(path, "/r") {
		return UnOfficialApiRequestUrl{}
	}
	return NotHaveUrl{}
}
