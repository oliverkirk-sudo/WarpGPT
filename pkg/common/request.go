package common

import (
	"WarpGPT/pkg/env"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/plugins/service/proxypool"
	"encoding/json"
	browser "github.com/EDDYCJY/fake-useragent"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/gin-gonic/gin"
	"io"
	"math/rand"
	"sync"
)

type Context struct {
	GinContext     *gin.Context
	RequestUrl     string
	RequestClient  tls_client.HttpClient
	RequestBody    io.ReadCloser
	RequestParam   string
	RequestMethod  string
	RequestHeaders http.Header
}

var tu sync.Mutex

type RequestUrl interface {
	Generate(path string, rawquery string) string
}

func GetContextPack[T RequestUrl](ctx *gin.Context, reqUrl T) Context {
	conversation := Context{}
	conversation.GinContext = ctx
	conversation.RequestUrl = reqUrl.Generate(ctx.Param("path"), ctx.Request.URL.RawQuery)
	conversation.RequestMethod = ctx.Request.Method
	conversation.RequestBody = ctx.Request.Body
	conversation.RequestParam = ctx.Param("path")
	conversation.RequestClient = GetHttpClient()
	conversation.RequestHeaders = http.Header(ctx.Request.Header)
	return conversation
}
func setUserAgent() {
	tu.Lock()
	env.E.UserAgent = browser.Safari()
	tu.Unlock()
}

func GetHttpClient() tls_client.HttpClient {
	jar := tls_client.NewCookieJar()
	userAgent := map[int]profiles.ClientProfile{
		1: profiles.Safari_15_6_1,
		2: profiles.Safari_16_0,
		3: profiles.Safari_IOS_15_5,
		4: profiles.Safari_IOS_15_6,
		5: profiles.Safari_IOS_16_0,
	}
	go setUserAgent()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(120),
		tls_client.WithClientProfile(userAgent[rand.Intn(6)]),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
		tls_client.WithRandomTLSExtensionOrder(),
	}
	if env.E.ProxyPoolUrl != "" {
		ip, err := proxypool.ProxyPoolInstance.GetIpInRedis()
		if err != nil {
			logger.Log.Warning(err.Error())
			return nil
		}
		options = append(options, tls_client.WithProxyUrl(ip))
	} else {
		options = append(options, tls_client.WithProxyUrl(env.E.Proxy))
	}
	client, _ := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	return client
}

func RequestOpenAI[T any](path string, body io.Reader, accessToken string, requestMethod string) (*T, error) {
	url := "https://" + env.E.OpenaiHost + path
	req, err := http.NewRequest(requestMethod, url, body)
	if err != nil {
		logger.Log.Error("Error creating request:", err)
		return nil, err
	}
	headers := map[string]string{
		"Host":          env.E.OpenaiHost,
		"Origin":        "https://" + env.E.OpenaiHost,
		"Authorization": accessToken,
		"Connection":    "keep-alive",
		"User-Agent":    env.E.UserAgent,
		"Referer":       "https://" + env.E.OpenaiHost,
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := GetHttpClient().Do(req)
	if err != nil {
		logger.Log.Error("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()
	var data T
	readAll, err := io.ReadAll(resp.Body)
	if readAll == nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(readAll, &data)
	return &data, nil
}
