package funcaptcha

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

type Session struct {
	Sid              string                        `json:"sid"`
	SessionToken     string                        `json:"session_token"`
	Hex              string                        `json:"hex"`
	ChallengeLogger  challengeLogger               `json:"challenge_logger"`
	Challenge        Challenge                     `json:"challenge"`
	ConciseChallenge ConciseChallenge              `json:"concise_challenge"`
	Headers          http.Header                   `json:"headers"`
	Client           *tls_client.HttpClient        `json:"-"`
	options          []tls_client.HttpClientOption `json:"-"`
}

type ConciseChallenge struct {
	GameType     string   `json:"game_type"`
	URLs         []string `json:"urls"`
	Instructions string   `json:"instructions"`
}
type Input struct {
	Index int
}
type ValueFunc func(Input) Input
type KeyFunc func(Input) interface{}

var Yz = map[int]struct {
	Value map[string]ValueFunc
	Key   map[string]KeyFunc
}{
	4: {
		Value: map[string]ValueFunc{
			"alpha": func(c Input) Input {
				yValueStr := strconv.Itoa(c.Index)          // 转换为字符串
				combinedStr := yValueStr + strconv.Itoa(1)  // 加1
				combinedInt, _ := strconv.Atoi(combinedStr) // 将合并后的字符串转回为整数
				return Input{Index: combinedInt - 2}
			},
			"beta":    func(c Input) Input { return Input{Index: -c.Index} },
			"gamma":   func(c Input) Input { return Input{Index: 3 * (3 - c.Index)} },
			"delta":   func(c Input) Input { return Input{Index: 7 * c.Index} },
			"epsilon": func(c Input) Input { return Input{Index: 2 * c.Index} },
			"zeta": func(c Input) Input {
				if c.Index != 0 {
					return Input{Index: 100 / c.Index}
				}
				return Input{Index: c.Index}
			},
		},
		Key: map[string]KeyFunc{
			"alpha": func(c Input) interface{} {
				return []int{rand.Intn(100), c.Index, rand.Intn(100)}
			},
			"beta": func(c Input) interface{} {
				return map[string]int{
					"size":          50 - c.Index,
					"id":            c.Index,
					"limit":         10 * c.Index,
					"req_timestamp": int(time.Now().UnixNano() / int64(time.Millisecond)),
				}
			},
			"gamma": func(c Input) interface{} {
				return c.Index
			},
			"delta": func(c Input) interface{} {
				return map[string]int{"index": c.Index}
			},
			"epsilon": func(c Input) interface{} {
				arr := make([]int, rand.Intn(5)+1)
				randIndex := rand.Intn(len(arr))
				for i := range arr {
					if i == randIndex {
						arr[i] = c.Index
					} else {
						arr[i] = rand.Intn(10)
					}
				}
				return append(arr, randIndex)
			},
			"zeta": func(c Input) interface{} {
				return append(make([]int, rand.Intn(5)+1), c.Index)
			},
		},
	},
}

func YB(gameType int, apiBreaker *ApiBreaker) func(Input) interface{} {
	return func(input Input) interface{} {
		for _, valueFuncName := range apiBreaker.Value {
			input = Yz[gameType].Value[valueFuncName](input)
		}
		return Yz[gameType].Key[apiBreaker.Key](input)
	}
}

type Challenge struct {
	SessionToken         string      `json:"session_token"`
	ChallengeID          string      `json:"challengeID"`
	ChallengeURL         string      `json:"challengeURL"`
	AudioChallengeURLs   []string    `json:"audio_challenge_urls"`
	AudioGameRateLimited interface{} `json:"audio_game_rate_limited"`
	Sec                  int         `json:"sec"`
	EndURL               interface{} `json:"end_url"`
	GameData             struct {
		GameType          int    `json:"gameType"`
		GameVariant       string `json:"game_variant"`
		InstructionString string `json:"instruction_string"`
		CustomGUI         struct {
			ChallengeIMGs       []string    `json:"_challenge_imgs"`
			ApiBreaker          *ApiBreaker `json:"api_breaker"`
			ApiBreakerV2Enabled int         `json:"api_breaker_v2_enabled"`
		} `json:"customGUI"`
	} `json:"game_data"`
	GameSID             string            `json:"game_sid"`
	SID                 string            `json:"sid"`
	Lang                string            `json:"lang"`
	StringTablePrefixes []interface{}     `json:"string_table_prefixes"`
	StringTable         map[string]string `json:"string_table"`
	EarlyVictoryMessage interface{}       `json:"earlyVictoryMessage"`
	FontSizeAdjustments interface{}       `json:"font_size_adjustments"`
	StyleTheme          string            `json:"style_theme"`
}

type challengeLogger struct {
	Sid           string `json:"sid"`
	SessionToken  string `json:"session_token"`
	AnalyticsTier int    `json:"analytics_tier"`
	RenderType    string `json:"render_type"`
	Category      string `json:"category"`
	Action        string `json:"action"`
	// Omit if empty
	GameToken string `json:"game_token,omitempty"`
	GameType  string `json:"game_type,omitempty"`
}

type requestChallenge struct {
	Sid               string `json:"sid"`
	Token             string `json:"token"`
	AnalyticsTier     int    `json:"analytics_tier"`
	RenderType        string `json:"render_type"`
	Lang              string `json:"lang"`
	IsAudioGame       bool   `json:"isAudioGame"`
	APIBreakerVersion string `json:"apiBreakerVersion"`
}

type submitChallenge struct {
	SessionToken  string `json:"session_token"`
	Sid           string `json:"sid"`
	GameToken     string `json:"game_token"`
	Guess         string `json:"guess"`
	RenderType    string `json:"render_type"`
	AnalyticsTier int    `json:"analytics_tier"`
	Bio           string `json:"bio"`
}

type ApiBreaker struct {
	Key   string   `json:"key"`
	Value []string `json:"value"`
}

func StartChallenge(full_session, hex string) (*Session, error) {
	fields := strings.Split(full_session, "|")
	session_token := fields[0]
	sid := strings.Split(fields[1], "=")[1]

	session := Session{
		Sid:          sid,
		SessionToken: session_token,
		Hex:          hex,
	}
	session.Headers = headers
	session.Headers.Set("Referer", fmt.Sprintf("https://client-api.arkoselabs.com/fc/assets/ec-game-core/game-core/1.15.0/standard/index.html?session=%s", strings.Replace(full_session, "|", "&", -1)))
	session.ChallengeLogger = challengeLogger{
		Sid:           sid,
		SessionToken:  session_token,
		AnalyticsTier: 40,
		RenderType:    "canvas",
	}
	err := session.log("", 0, "Site URL", fmt.Sprintf("https://client-api.arkoselabs.com/v2/1.5.5/enforcement.%s.html", hex))
	jar := tls_client.NewCookieJar()
	session.options = []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(360),
		tls_client.WithClientProfile(profiles.Chrome_117),
		tls_client.WithRandomTLSExtensionOrder(),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
	}
	client, _ := tls_client.NewHttpClient(tls_client.NewNoopLogger(), session.options...)
	session.Client = &client
	return &session, err
}

func (c *Session) RequestChallenge(isAudioGame bool) (*ApiBreaker, error) {
	challenge_request := requestChallenge{
		Sid:               c.Sid,
		Token:             c.SessionToken,
		AnalyticsTier:     40,
		RenderType:        "canvas",
		Lang:              "en-us",
		IsAudioGame:       isAudioGame,
		APIBreakerVersion: "green",
	}
	payload := jsonToForm(toJSON(challenge_request))

	req, _ := http.NewRequest(http.MethodPost, "https://client-api.arkoselabs.com/fc/gfct/", strings.NewReader(payload))
	req.Header = c.Headers
	req.Header.Set("X-NewRelic-Timestamp", getTimeStamp())
	resp, err := (*c.Client).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var challenge_data Challenge
	err = json.Unmarshal(body, &challenge_data)
	if err != nil {
		return nil, err
	}
	err = c.log(challenge_data.ChallengeID, challenge_data.GameData.GameType, "loaded", "game loaded")
	c.Challenge = challenge_data
	// Build concise challenge
	var challenge_type string
	var challenge_urls []string
	var key string
	var apiBreaker *ApiBreaker
	switch challenge_data.GameData.GameType {
	case 4:
		challenge_type = "image"
		challenge_urls = challenge_data.GameData.CustomGUI.ChallengeIMGs
		instruction_string := challenge_data.GameData.InstructionString
		key = fmt.Sprintf("4.instructions-%s", instruction_string)
		if challenge_data.GameData.CustomGUI.ApiBreakerV2Enabled == 1 {
			apiBreaker = challenge_data.GameData.CustomGUI.ApiBreaker
		}
	case 101:
		challenge_type = "audio"
		challenge_urls = challenge_data.AudioChallengeURLs
		instruction_string := challenge_data.GameData.GameVariant
		key = fmt.Sprintf("audio_game.instructions-%s", instruction_string)

	default:
		challenge_type = "unknown"
		challenge_urls = []string{}
	}

	c.ConciseChallenge = ConciseChallenge{
		GameType:     challenge_type,
		URLs:         challenge_urls,
		Instructions: strings.ReplaceAll(strings.ReplaceAll(challenge_data.StringTable[key], "<strong>", ""), "</strong>", ""),
	}
	return apiBreaker, err
}

func (c *Session) SubmitAnswer(indices []int, isAudio bool, apiBreaker *ApiBreaker) error {
	submission := submitChallenge{
		SessionToken:  c.SessionToken,
		Sid:           c.Sid,
		GameToken:     c.Challenge.ChallengeID,
		RenderType:    "canvas",
		AnalyticsTier: 40,
		Bio:           "eyJtYmlvIjoiMTUwLDAsMTE3LDIzOTszMDAsMCwxMjEsMjIxOzMxNywwLDEyNCwyMTY7NTUwLDAsMTI5LDIxMDs1NjcsMCwxMzQsMjA3OzYxNywwLDE0NCwyMDU7NjUwLDAsMTU1LDIwNTs2NjcsMCwxNjUsMjA1OzY4NCwwLDE3MywyMDc7NzAwLDAsMTc4LDIxMjs4MzQsMCwyMjEsMjI4OzI2MDY3LDAsMTkzLDM1MTsyNjEwMSwwLDE4NSwzNTM7MjYxMDEsMCwxODAsMzU3OzI2MTM0LDAsMTcyLDM2MTsyNjE4NCwwLDE2NywzNjM7MjYyMTcsMCwxNjEsMzY1OzI2MzM0LDAsMTU2LDM2NDsyNjM1MSwwLDE1MiwzNTQ7MjYzNjcsMCwxNTIsMzQzOzI2Mzg0LDAsMTUyLDMzMTsyNjQ2NywwLDE1MSwzMjU7MjY0NjcsMCwxNTEsMzE3OzI2NTAxLDAsMTQ5LDMxMTsyNjY4NCwxLDE0NywzMDc7MjY3NTEsMiwxNDcsMzA3OzMwNDUxLDAsMzcsNDM3OzMwNDY4LDAsNTcsNDI0OzMwNDg0LDAsNjYsNDE0OzMwNTAxLDAsODgsMzkwOzMwNTAxLDAsMTA0LDM2OTszMDUxOCwwLDEyMSwzNDk7MzA1MzQsMCwxNDEsMzI0OzMwNTUxLDAsMTQ5LDMxNDszMDU4NCwwLDE1MywzMDQ7MzA2MTgsMCwxNTUsMjk2OzMwNzUxLDAsMTU5LDI4OTszMDc2OCwwLDE2NywyODA7MzA3ODQsMCwxNzcsMjc0OzMwODE4LDAsMTgzLDI3MDszMDg1MSwwLDE5MSwyNzA7MzA4ODQsMCwyMDEsMjY4OzMwOTE4LDAsMjA4LDI2ODszMTIzNCwwLDIwNCwyNjM7MzEyNTEsMCwyMDAsMjU3OzMxMzg0LDAsMTk1LDI1MTszMTQxOCwwLDE4OSwyNDk7MzE1NTEsMSwxODksMjQ5OzMxNjM0LDIsMTg5LDI0OTszMTcxOCwxLDE4OSwyNDk7MzE3ODQsMiwxODksMjQ5OzMxODg0LDEsMTg5LDI0OTszMTk2OCwyLDE4OSwyNDk7MzIyODQsMCwyMDIsMjQ5OzMyMzE4LDAsMjE2LDI0NzszMjMxOCwwLDIzNCwyNDU7MzIzMzQsMCwyNjksMjQ1OzMyMzUxLDAsMzAwLDI0NTszMjM2OCwwLDMzOSwyNDE7MzIzODQsMCwzODgsMjM5OzMyNjE4LDAsMzkwLDI0NzszMjYzNCwwLDM3NCwyNTM7MzI2NTEsMCwzNjUsMjU1OzMyNjY4LDAsMzUzLDI1NzszMjk1MSwxLDM0OCwyNTc7MzMwMDEsMiwzNDgsMjU3OzMzNTY4LDAsMzI4LDI3MjszMzU4NCwwLDMxOSwyNzg7MzM2MDEsMCwzMDcsMjg2OzMzNjUxLDAsMjk1LDI5NjszMzY1MSwwLDI5MSwzMDA7MzM2ODQsMCwyODEsMzA5OzMzNjg0LDAsMjcyLDMxNTszMzcxOCwwLDI2NiwzMTc7MzM3MzQsMCwyNTgsMzIzOzMzNzUxLDAsMjUyLDMyNzszMzc1MSwwLDI0NiwzMzM7MzM3NjgsMCwyNDAsMzM3OzMzNzg0LDAsMjM2LDM0MTszMzgxOCwwLDIyNywzNDc7MzM4MzQsMCwyMjEsMzUzOzM0MDUxLDAsMjE2LDM1NDszNDA2OCwwLDIxMCwzNDg7MzQwODQsMCwyMDQsMzQ0OzM0MTAxLDAsMTk4LDM0MDszNDEzNCwwLDE5NCwzMzY7MzQ1ODQsMSwxOTIsMzM0OzM0NjUxLDIsMTkyLDMzNDsiLCJ0YmlvIjoiIiwia2JpbyI6IiJ9",
	}
	var answerIndex []string
	if isAudio {
		for _, answer := range indices {
			answerIndex = append(answerIndex, strconv.Itoa(answer))
		}
	} else {
		for _, answer := range indices {
			input := Input{Index: answer}
			encoder := YB(4, apiBreaker)
			result := encoder(input)
			marshal, _ := json.Marshal(result)
			answerIndex = append(answerIndex, string(marshal))
		}
	}
	answer := "[" + strings.Join(answerIndex, ",") + "]"
	submission.Guess = Encrypt(answer, c.SessionToken)
	payload := jsonToForm(toJSON(submission))
	req, _ := http.NewRequest(http.MethodPost, "https://client-api.arkoselabs.com/fc/ca/", strings.NewReader(payload))
	req.Header = c.Headers
	req.Header.Set("X-Requested-ID", getRequestId(c.SessionToken))
	req.Header.Set("X-NewRelic-Timestamp", getTimeStamp())

	resp, err := (*c.Client).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var response struct {
		Error          string `json:"error"`
		Response       string `json:"response"`
		Solved         bool   `json:"solved"`
		IncorrectGuess string `json:"incorrect_guess"`
		Score          int    `json:"score"`
	}
	log.Println(string(body))
	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}
	if response.Error != "" {
		return errors.New(response.Error)
	}
	if !response.Solved {
		return fmt.Errorf("incorrect guess: %s", response.IncorrectGuess)
	}
	// Set new client
	cli, _ := tls_client.NewHttpClient(tls_client.NewNoopLogger(), c.options...)
	c.Client = &cli
	return nil
}

func (c *Session) log(game_token string, game_type int, category, action string) error {
	v := c.ChallengeLogger
	v.GameToken = game_token
	if game_type != 0 {
		v.GameType = fmt.Sprintf("%d", game_type)
	}
	v.Category = category
	v.Action = action

	request, _ := http.NewRequest(http.MethodPost, "https://client-api.arkoselabs.com/fc/a/", strings.NewReader(jsonToForm(toJSON(v))))
	request.Header = headers
	resp, err := (*c.Client).Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	return nil
}
