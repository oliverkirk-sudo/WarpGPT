package tools

import (
	"WarpGPT/pkg/common"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

func GetHttpClient() tls_client.HttpClient {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(120),
		tls_client.WithClientProfile(profiles.Safari_15_6_1),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
		tls_client.WithProxyUrl(common.Env.Proxy),
	}
	client, _ := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	return client
}
