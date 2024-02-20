package funcaptcha

import (
	"WarpGPT/pkg/env"
	"encoding/json"
	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type arkVer int

const (
	ArkVerAuth  arkVer = 0
	ArkVerReg   arkVer = 1
	ArkVerChat3 arkVer = 3
	ArkVerChat4 arkVer = 4
)

type Solver struct {
	initVer string
	initHex string
	arks    map[arkVer][]arkReq
	client  *tls_client.HttpClient
}

type solverArg func(*Solver)

func NewSolver(args ...solverArg) *Solver {
	var (
		jar     = tls_client.NewCookieJar()
		options = []tls_client.HttpClientOption{
			tls_client.WithTimeoutSeconds(360),
			tls_client.WithClientProfile(profiles.Chrome_117),
			tls_client.WithRandomTLSExtensionOrder(),
			tls_client.WithNotFollowRedirects(),
			tls_client.WithCookieJar(jar),
		}
		client, _ = tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	)
	s := &Solver{
		arks:    make(map[arkVer][]arkReq),
		client:  &client,
		initVer: "1.5.4",
		initHex: "cd12da708fe6cbe6e068918c38de2ad9",
	}
	for _, arg := range args {
		arg(s)
	}
	return s
}

func WithInitVer(ver string) solverArg {
	return func(s *Solver) {
		s.initVer = ver
	}
}

func WithProxy(proxy string) solverArg {
	return func(s *Solver) {
		(*s.client).SetProxy(proxy)
	}
}

func WithInitHex(hex string) solverArg {
	return func(s *Solver) {
		s.initHex = hex
	}
}

func WithClient(client *tls_client.HttpClient) solverArg {
	return func(s *Solver) {
		s.client = client
	}
}

func WithHarData(harData HARData) solverArg {
	return func(s *Solver) {
		for _, v := range harData.Log.Entries {
			if strings.HasPrefix(v.Request.URL, arkPreURL) || strings.HasPrefix(v.Request.URL, arkAuthPreURL) {
				var tmpArk arkReq
				tmpArk.arkURL = v.Request.URL
				if v.StartedDateTime == "" {
					println("Error: no arkose request!")
					continue
				}
				t, _ := time.Parse(time.RFC3339, v.StartedDateTime)
				bw := getBw(t.Unix())
				fallbackBw := getBw(t.Unix() - 21600)
				tmpArk.arkHeader = make(http.Header)
				for _, h := range v.Request.Headers {
					if !strings.EqualFold(h.Name, "content-length") && !strings.EqualFold(h.Name, "cookie") && !strings.HasPrefix(h.Name, ":") {
						tmpArk.arkHeader.Set(h.Name, h.Value)
						if strings.EqualFold(h.Name, "user-agent") {
							tmpArk.userAgent = h.Value
						}
					}
				}
				tmpArk.arkCookies = []*http.Cookie{}
				for _, cookie := range v.Request.Cookies {
					expire, _ := time.Parse(time.RFC3339, cookie.Expires)
					if expire.After(time.Now()) {
						tmpArk.arkCookies = append(tmpArk.arkCookies, &http.Cookie{Name: cookie.Name, Value: cookie.Value, Expires: expire.UTC()})
					}
				}
				var arkType string
				tmpArk.arkBody = make(url.Values)
				for _, p := range v.Request.PostData.Params {
					if p.Name == "bda" {
						cipher, err := url.QueryUnescape(p.Value)
						if err != nil {
							panic(err)
						}
						tmpArk.arkBx = Decrypt(cipher, tmpArk.userAgent+bw, tmpArk.userAgent+fallbackBw)
					} else if p.Name != "rnd" {
						query, err := url.QueryUnescape(p.Value)
						if err != nil {
							panic(err)
						}
						tmpArk.arkBody.Set(p.Name, query)
						if p.Name == "public_key" {
							if query == "0A1D34FC-659D-4E23-B17B-694DCFCF6A6C" {
								arkType = "auth"
								s.arks[ArkVerAuth] = append(s.arks[ArkVerAuth], tmpArk)
							} else if query == "3D86FBBA-9D22-402A-B512-3420086BA6CC" {
								arkType = "chat3"
								s.arks[ArkVerChat3] = append(s.arks[ArkVerChat3], tmpArk)
							} else if query == "35536E1E-65B4-4D96-9D97-6ADB7EFF8147" {
								arkType = "chat4"
								s.arks[ArkVerChat4] = append(s.arks[ArkVerChat4], tmpArk)
							} else if query == "0655BC92-82E1-43D9-B32E-9DF9B01AF50C" {
								arkType = "reg"
								s.arks[ArkVerReg] = append(s.arks[ArkVerReg], tmpArk)
							}
						}
					}
				}
				if tmpArk.arkBx != "" {
					println("success read " + arkType + " arkose")
				} else {
					println("failed to decrypt HAR file")
				}
			}
		}

	}
}

func WithHarpool(s *Solver) {
	dirPath := path.Join(path.Dir(env.EnvFile), "harPool")
	var harPath []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := filepath.Ext(info.Name())
			if ext == ".har" {
				harPath = append(harPath, path)
			}
		}
		return nil
	})
	if err != nil {
		println("Error: please put HAR files in harPool directory!")
	}
	for _, path := range harPath {
		file, err := os.ReadFile(path)
		if err != nil {
			return
		}
		var harFile HARData
		err = json.Unmarshal(file, &harFile)
		if err != nil {
			println("Error: not a HAR file!")
			return
		}
		WithHarData(harFile)(s)
	}

}
