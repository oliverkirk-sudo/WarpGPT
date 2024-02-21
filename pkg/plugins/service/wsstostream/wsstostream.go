package wsstostream

import (
	"WarpGPT/pkg/common"
	"WarpGPT/pkg/env"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/tools"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	http "github.com/bogdanfinn/fhttp"
	"github.com/gorilla/websocket"
	"golang.org/x/net/proxy"
	"io"
	shttp "net/http"
	"net/url"
	"time"
)

type RegisterWebsocket struct {
	ExpiresAt time.Time `json:"expires_at"`
	WssUrl    string    `json:"wss_url"`
}
type WsResponse struct {
	SequenceId int    `json:"sequenceId"`
	Type       string `json:"type"`
	From       string `json:"from"`
	DataType   string `json:"dataType"`
	Data       struct {
		Type           string `json:"type"`
		Body           string `json:"body"`
		MoreBody       bool   `json:"more_body"`
		ResponseId     string `json:"response_id"`
		ConversationId string `json:"conversation_id"`
		MessageId      string `json:"message_id"`
	} `json:"data"`
}
type Reconnect struct {
	Type              string `json:"type"`
	Event             string `json:"event"`
	UserId            string `json:"userId"`
	ConnectionId      string `json:"connectionId"`
	ReconnectionToken string `json:"reconnectionToken"`
}
type WssToStream struct {
	ConversationId string
	ResponseId     string
	AccessToken    string
	Server         *websocket.Conn
	WS             *RegisterWebsocket
	Reconnect
}

func NewWssToStream(accessToken string) *WssToStream {
	return &WssToStream{AccessToken: accessToken}
}

func GetRegisterWebsocket(accessToken string) (*RegisterWebsocket, error) {
	logger.Log.Debug("GetRegisterWebsocket")
	WS, err := common.RequestOpenAI[RegisterWebsocket]("/backend-api/register-websocket", nil, accessToken, http.MethodPost)
	if err != nil {
		return nil, err
	}
	if err != nil {
		logger.Log.Error("Error decoding response:", err)
		return nil, err
	}
	if WS.WssUrl != "" {
		logger.Log.Debug("GetRegisterWebsocket Success WssUrl:", WS.WssUrl)
		return WS, nil
	} else {
		return nil, errors.New("check your access_key")
	}
}
func (s *WssToStream) InitConnect() error {
	logger.Log.Debug("Try Connect To WS")
	var dialer websocket.Dialer

	// 当 env.Env.Proxy 不为空字符串时，才配置代理
	if env.Env.Proxy != "" {
		proxyAddr, err := url.Parse(env.Env.Proxy)
		if err != nil {
			logger.Log.Error("Error parsing proxy URL:", err)
			return err
		}

		switch proxyAddr.Scheme {
		case "http", "https":
			dialer.Proxy = shttp.ProxyURL(proxyAddr)
		case "socks5":
			socksDialer, err := proxy.FromURL(proxyAddr, proxy.Direct)
			if err != nil {
				logger.Log.Error("Error creating SOCKS proxy dialer:", err)
				return err
			}
			dialer.NetDial = socksDialer.Dial
		default:
			logger.Log.Error("Unsupported proxy scheme:", proxyAddr.Scheme)
			return errors.New("unsupported proxy scheme")
		}
	}

	headers := http.Header{}
	headers.Set("Origin", "https://"+env.Env.OpenaiHost)
	headers.Set("Sec-WebSocket-Protocol", "json.reliable.webpubsub.azure.v1")
	headers.Set("User-Agent", env.Env.UserAgent)

	item, exist := tools.AllCache.CacheGet(s.AccessToken)
	if !exist {
		registerWebsocket, err := GetRegisterWebsocket(s.AccessToken)
		if err != nil {
			return err
		}
		tools.AllCache.CacheSet(s.AccessToken, registerWebsocket)
		s.WS = registerWebsocket
	} else {
		if s.WS.ExpiresAt.Before(time.Now()) {
			registerWebsocket, err := GetRegisterWebsocket(s.AccessToken)
			if err != nil {
				return err
			}
			tools.AllCache.CacheSet(s.AccessToken, registerWebsocket)
			s.WS = registerWebsocket
		} else {
			s.WS = item.(*RegisterWebsocket)
		}
	}

	c, _, err := dialer.Dial(s.WS.WssUrl, shttp.Header(headers))
	if err != nil {
		logger.Log.Error("Dial error:", err)
		return err
	}
	logger.Log.Debug("WS Connect Success")
	s.Server = c
	_, msg, err := s.Server.ReadMessage()
	if err != nil {
		return err
	}
	logger.Log.Debug("Init Read Message:", string(msg))
	return nil
}

type NopCloser struct {
	*bytes.Reader
}

func (NopCloser) Close() error {
	return nil
}
func NewNopCloser(data []byte) io.ReadCloser {
	return NopCloser{Reader: bytes.NewReader(data)}
}

func (s *WssToStream) ReadMessage() (io.ReadCloser, error) {
	logger.Log.Debug("Read Messages")
	_, msg, err := s.Server.ReadMessage()
	if err != nil {
		logger.Log.Error("read message error:", err)
	}
	var response WsResponse
	if err = json.Unmarshal(msg, &response); err != nil {
		logger.Log.Error("unmarshal message error:", err)
	}
	if response.Data.ResponseId == s.ResponseId && response.Data.ConversationId == s.ConversationId {
		if response.Data.Body == "ZGF0YTogW0RPTkVdCgo=" {
			s.Server.Close()
		}
		data, err := base64.StdEncoding.DecodeString(response.Data.Body)
		if err != nil {
			return nil, err
		}
		return NewNopCloser(data), nil
	} else {
		return nil, nil
	}
}
func (s *WssToStream) Read(p []byte) (n int, err error) {
	logger.Log.Debug("Read")
	_, message, err := s.Server.ReadMessage()
	if err != nil {
		return 0, err
	}
	var response WsResponse
	if err = json.Unmarshal(message, &response); err != nil {
		logger.Log.Error("unmarshal message error:", err)
	}
	if response.Data.ResponseId == s.ResponseId && response.Data.ConversationId == s.ConversationId {
		if response.Data.Body == "ZGF0YTogW0RPTkVdCgo=" {
			s.Server.Close()
		}
		data, err := base64.StdEncoding.DecodeString(response.Data.Body)
		if err != nil {
			return 0, err
		}
		copyLen := copy(p, data)
		if copyLen < len(data) {
			return copyLen, errors.New("buffer too small to hold message")
		}
		return copyLen, nil
	} else {
		return 0, nil
	}
}

func (s *WssToStream) Close() error {
	return s.Server.Close()
}
