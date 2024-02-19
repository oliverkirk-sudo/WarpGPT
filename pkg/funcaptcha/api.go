package funcaptcha

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
)

const arkPreURL = "https://tcr9i.chat.openai.com/fc/gt2/"
const arkAuthPreURL = "https://tcr9i.openai.com/fc/gt2/"

var arkURLIns, _ = url.Parse(arkPreURL)

type arkReq struct {
	arkURL     string
	arkBx      string
	arkHeader  http.Header
	arkBody    url.Values
	arkCookies []*http.Cookie
	userAgent  string
}

type kvPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type cookie struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Expires string `json:"expires"`
}
type postBody struct {
	Params []kvPair `json:"params"`
}
type request struct {
	URL      string   `json:"url"`
	Headers  []kvPair `json:"headers,omitempty"`
	PostData postBody `json:"postData,omitempty"`
	Cookies  []cookie `json:"cookies,omitempty"`
}
type entries struct {
	StartedDateTime string  `json:"startedDateTime"`
	Request         request `json:"request"`
}
type logData struct {
	Entries []entries `json:"entries"`
}
type HARData struct {
	Log logData `json:"log"`
}

func (s *Solver) GetOpenAIToken(version arkVer, puid string) (string, error) {
	token, err := s.sendRequest(version, "", puid)
	return token, err
}

func (s *Solver) GetOpenAITokenWithBx(version arkVer, bx string, puid string) (string, error) {
	token, err := s.sendRequest(version, getBdaWitBx(bx), puid)
	return token, err
}

func (s *Solver) sendRequest(arkType arkVer, bda string, puid string) (string, error) {
	if len(s.arks[arkType]) == 0 {
		return "", errors.New("a valid HAR file with arkType " + strconv.Itoa(int(arkType)) + " required")
	}
	var tmpArk *arkReq = &s.arks[arkType][0]
	s.arks[arkType] = append(s.arks[arkType][1:], s.arks[arkType][0])
	if tmpArk == nil || tmpArk.arkBx == "" || len(tmpArk.arkBody) == 0 || len(tmpArk.arkHeader) == 0 {
		return "", errors.New("a valid HAR file required")
	}
	if bda == "" {
		bda = s.getBDA(tmpArk)
	}
	tmpArk.arkBody.Set("bda", base64.StdEncoding.EncodeToString([]byte(bda)))
	tmpArk.arkBody.Set("rnd", strconv.FormatFloat(rand.Float64(), 'f', -1, 64))
	req, _ := http.NewRequest(http.MethodPost, tmpArk.arkURL, strings.NewReader(tmpArk.arkBody.Encode()))
	req.Header = tmpArk.arkHeader.Clone()
	(*s.client).GetCookieJar().SetCookies(arkURLIns, tmpArk.arkCookies)
	if puid != "" {
		req.Header.Set("cookie", "_puid="+puid+";")
	}
	resp, err := (*s.client).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New("status code " + resp.Status)
	}

	type arkoseResponse struct {
		Token string `json:"token"`
	}
	var arkose arkoseResponse
	err = json.NewDecoder(resp.Body).Decode(&arkose)
	if err != nil {
		return "", err
	}
	// Check if rid is empty
	if !strings.Contains(arkose.Token, "pk=") {
		return arkose.Token, errors.New("captcha required")
	}

	return arkose.Token, nil
}

//goland:noinspection SpellCheckingInspection
func (s *Solver) getBDA(arkReq *arkReq) string {
	var bx string = arkReq.arkBx
	if bx == "" {
		bx = fmt.Sprintf(bx_template,
			getF(),
			getN(),
			getWh(),
			webglExtensions,
			getWebglExtensionsHash(),
			webglRenderer,
			webglVendor,
			webglVersion,
			webglShadingLanguageVersion,
			webglAliasedLineWidthRange,
			webglAliasedPointSizeRange,
			webglAntialiasing,
			webglBits,
			webglMaxParams,
			webglMaxViewportDims,
			webglUnmaskedVendor,
			webglUnmaskedRenderer,
			webglVsfParams,
			webglVsiParams,
			webglFsfParams,
			webglFsiParams,
			getWebglHashWebgl(),
			s.initVer,
			s.initHex,
			getFe(),
			getIfeHash(),
		)
	} else {
		re := regexp.MustCompile(`"key"\:"n","value"\:"\S*?"`)
		bx = re.ReplaceAllString(bx, `"key":"n","value":"`+getN()+`"`)
	}
	bt := getBt()
	bw := getBw(bt)
	return Encrypt(bx, arkReq.userAgent+bw)
}

func getBt() int64 {
	return time.Now().UnixMicro() / 1000000
}

func getBw(bt int64) string {
	return strconv.FormatInt(bt-(bt%21600), 10)
}

func getBdaWitBx(bx string) string {
	bt := getBt()
	bw := getBw(bt)
	return Encrypt(bx, bv+bw)
}
