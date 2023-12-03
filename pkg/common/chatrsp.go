package common

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type ChatReqStr struct {
	Action   string `json:"action"`
	Messages []struct {
		Id     string `json:"id"`
		Author struct {
			Role string `json:"role"`
		} `json:"author"`
		Content struct {
			ContentType string   `json:"content_type"`
			Parts       []string `json:"parts"`
		} `json:"content"`
		Metadata struct {
		} `json:"metadata"`
	} `json:"messages"`
	ParentMessageId            string        `json:"parent_message_id"`
	Model                      string        `json:"model"`
	TimezoneOffsetMin          int           `json:"timezone_offset_min"`
	Suggestions                []interface{} `json:"suggestions"`
	HistoryAndTrainingDisabled bool          `json:"history_and_training_disabled"`
	ArkoseToken                string        `json:"arkose_token"`
	ConversationMode           struct {
		Kind string `json:"kind"`
	} `json:"conversation_mode"`
	ForceParagen   bool `json:"force_paragen"`
	ForceRateLimit bool `json:"force_rate_limit"`
}
type ChatReqTemplate struct {
	Id     string `json:"id"`
	Author struct {
		Role string `json:"role"`
	} `json:"author"`
	Content struct {
		ContentType string   `json:"content_type"`
		Parts       []string `json:"parts"`
	} `json:"content"`
	Metadata struct {
	} `json:"metadata"`
}
type ChatFileReqTemplate struct {
	Id     string `json:"id"`
	Author struct {
		Role string `json:"role"`
	} `json:"author"`
	Content struct {
		ContentType string        `json:"content_type"`
		Parts       []interface{} `json:"parts"`
	} `json:"content"`
	Metadata struct {
		Attachments []interface{} `json:"attachments"`
	} `json:"metadata"`
}
type ChatRespStr struct {
	Message struct {
		Id     string `json:"id"`
		Author struct {
			Role     string      `json:"role"`
			Name     interface{} `json:"name"`
			Metadata struct {
			} `json:"metadata"`
		} `json:"author"`
		CreateTime float64     `json:"create_time"`
		UpdateTime interface{} `json:"update_time"`
		Content    struct {
			ContentType string   `json:"content_type"`
			Parts       []string `json:"parts"`
		} `json:"content"`
		Status   string  `json:"status"`
		EndTurn  bool    `json:"end_turn"`
		Weight   float64 `json:"weight"`
		Metadata struct {
			FinishDetails struct {
				Type       string `json:"type"`
				StopTokens []int  `json:"stop_tokens"`
			} `json:"finish_details"`
			IsComplete             bool   `json:"is_complete"`
			MessageType            string `json:"message_type"`
			ModelSlug              string `json:"model_slug"`
			ParentId               string `json:"parent_id"`
			Timestamp              string `json:"timestamp_"`
			IsUserSystemMessage    bool   `json:"is_user_system_message"`
			UserContextMessageData struct {
				AboutModelMessage string `json:"about_model_message"`
			} `json:"user_context_message_data"`
		} `json:"metadata"`
		Recipient string `json:"recipient"`
	} `json:"message"`
	ConversationId string      `json:"conversation_id"`
	Error          interface{} `json:"error"`
}
type ChatEndRespStr struct {
	ConversationId     string `json:"conversation_id"`
	MessageId          string `json:"message_id"`
	IsCompletion       bool   `json:"is_completion"`
	ModerationResponse struct {
		Flagged      bool   `json:"flagged"`
		Blocked      bool   `json:"blocked"`
		ModerationId string `json:"moderation_id"`
	} `json:"moderation_response"`
}
type ChatUserSystemMsgReqStr struct {
	AboutUserMessage  string `json:"about_user_message"`
	AboutModelMessage string `json:"about_model_message"`
	Enabled           bool   `json:"enabled"`
}
type ChatUserSystemMsgRespStr struct {
	Object            string `json:"object"`
	Enabled           bool   `json:"enabled"`
	AboutUserMessage  string `json:"about_user_message"`
	AboutModelMessage string `json:"about_model_message"`
}
type ChatDetectedErrorRespStr struct {
	Message        interface{} `json:"message"`
	ConversationId string      `json:"conversation_id"`
	Error          string      `json:"error"`
}
type DALLERespStr struct {
	Message struct {
		Id     string `json:"id"`
		Author struct {
			Role     string `json:"role"`
			Name     string `json:"name"`
			Metadata struct {
			} `json:"metadata"`
		} `json:"author"`
		CreateTime interface{} `json:"create_time"`
		UpdateTime interface{} `json:"update_time"`
		Content    struct {
			ContentType string `json:"content_type"`
			Parts       []struct {
				ContentType  string `json:"content_type"`
				AssetPointer string `json:"asset_pointer"`
				SizeBytes    int    `json:"size_bytes"`
				Width        int    `json:"width"`
				Height       int    `json:"height"`
				Fovea        int    `json:"fovea"`
				Metadata     struct {
					Dalle struct {
						GenId              string `json:"gen_id"`
						Prompt             string `json:"prompt"`
						Seed               int64  `json:"seed"`
						SerializationTitle string `json:"serialization_title"`
					} `json:"dalle"`
				} `json:"metadata"`
			} `json:"parts"`
		} `json:"content"`
		Status   string      `json:"status"`
		EndTurn  interface{} `json:"end_turn"`
		Weight   float64     `json:"weight"`
		Metadata struct {
			MessageType string `json:"message_type"`
			ModelSlug   string `json:"model_slug"`
			ParentId    string `json:"parent_id"`
		} `json:"metadata"`
		Recipient string `json:"recipient"`
	} `json:"message"`
	ConversationId string      `json:"conversation_id"`
	Error          interface{} `json:"error"`
}
type ApiRespStr struct {
	Id                string `json:"id"`
	Object            string `json:"object"`
	Created           int64  `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}
type ApiRespStrStream struct {
	Id                string `json:"id"`
	Object            string `json:"object"`
	Created           int64  `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		Index        int         `json:"index"`
		FinishReason interface{} `json:"finish_reason"`
	} `json:"choices"`
}
type ApiRespStrStreamEnd struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
		} `json:"delta"`
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}
type ApiImageGenerationRespStr struct {
	Created int64 `json:"created"`
	Data    []struct {
		RevisedPrompt string `json:"revised_prompt"`
		Url           string `json:"url"`
	} `json:"data"`
}
type ApiImageGenerationErrorRespStr struct {
	Error struct {
		Code    interface{} `json:"code"`
		Message string      `json:"message"`
		Param   interface{} `json:"param"`
		Type    string      `json:"type"`
	} `json:"error"`
}

func GetChatReqStr() *ChatReqStr {
	jsonStr := `{
        "action": "next",
        "messages": [
        ],
        "parent_message_id": "",
        "model": "gpt-4-code-interpreter",
        "timezone_offset_min": -480,
        "suggestions": [],
        "history_and_training_disabled": true,
        "arkose_token": "",
        "conversation_mode": {
            "kind": "primary_assistant"
        },
        "force_paragen": false,
        "force_rate_limit": false
    }`

	t := new(ChatReqStr)
	err := json.Unmarshal([]byte(jsonStr), &t)
	t.ParentMessageId = uuid.New().String()
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	return t
}
func GetChatReqTemplate() *ChatReqTemplate {
	jsonStr := `{
    "id": "aaa25982-c951-4be5-9e33-02ec1ffad43f",
    "author": {
        "role": "user"
    },
    "content": {
        "content_type": "text",
        "parts": [
            "如何使用呢"
        ]
    },
    "metadata": {}
	}`
	t := new(ChatReqTemplate)
	err := json.Unmarshal([]byte(jsonStr), &t)
	t.Id = uuid.New().String()
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	return t
}
func GetChatFileReqTemplate() *ChatFileReqTemplate {
	jsonStr := `{
    "id": "aaa2d6bc-ef84-4f43-984f-c2996f1051bd",
    "author": {
        "role": "user"
    },
    "content": {
        "content_type": "multimodal_text",
        "parts": [
        ]
    },
    "metadata": {
        "attachments": [
        ]
    }
}`
	t := new(ChatFileReqTemplate)
	err := json.Unmarshal([]byte(jsonStr), &t)
	t.Id = uuid.New().String()
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	return t
}

func GetChatRespStr() *ChatRespStr {
	jsonStr := `{
    "message":
        {
            "id": "7e30dbfe-e648-48f2-80ac-166363233f22",
            "author":
                {
                    "role": "assistant",
                    "name": null,
                    "metadata": {}
                },
            "create_time": 1699032699.636848,
            "update_time": null,
            "content": {
                "content_type": "text",
                "parts": [""]
            },
            "status": "finished_successfully",
            "end_turn": true,
            "weight": 1.0,
            "metadata": {
                "finish_details":
                    {
                        "type": "stop",
                        "stop_tokens": [100260]
                    },
                "is_complete": true,
                "message_type": "next",
                "model_slug": "gpt-4",
                "parent_id": "aaa27d58-b93c-4bc0-9fd4-0632b6b0b7b7",
				"timestamp_": "absolute",
				"message_type": null,
				"is_user_system_message": true,
				"user_context_message_data": {
					"about_model_message": "Strict adherence to Instructions"
				}
            }, "recipient": "all"
        },
    "conversation_id": "611228f2-94fd-44ed-b5d9-4f229ef3c400",
    "error": null
}`
	t := new(ChatRespStr)
	err := json.Unmarshal([]byte(jsonStr), &t)
	nowTime := fmt.Sprintf("%.6f\n", float64(time.Now().UnixNano())/1e9)
	floatTime, _ := strconv.ParseFloat(nowTime, 64)
	t.Message.CreateTime = floatTime
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	return t
}
func GetChatEndRespStr() *ChatEndRespStr {
	jsonStr := `{
    "conversation_id": "59dcb4e2-25e4-44e8-b231-ff0df525d101",
    "message_id": "2f27b73f-faec-4ad8-a1d2-25cc30ba40b4",
    "is_completion": true,
    "moderation_response": {
        "flagged": false,
        "blocked": false,
        "moderation_id": "modr-8H0ss3XXpdakuGslnbTgSTcUsUceM"
    }
}`
	t := new(ChatEndRespStr)
	err := json.Unmarshal([]byte(jsonStr), &t)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	return t
}
func GetChatUserSystemMsgReqStr() *ChatUserSystemMsgReqStr {
	jsonStr := `{
    "about_user_message": "",
    "about_model_message": "",
    "enabled": true
}`
	t := new(ChatUserSystemMsgReqStr)
	err := json.Unmarshal([]byte(jsonStr), &t)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	return t
}
func GetApiRespStr() *ApiRespStr {
	jsonStr := `{
    "id": "chatcmpl-8H3JOH8ErCUKSM0KVZ6tHE06T0XLi",
    "object": "chat.completion",
    "created": 1699074998,
    "model": "gpt-3.5-turbo-0613",
	"system_fingerprint": null,
    "choices": [
        {
            "index": 0,
            "message": {
                "role": "assistant",
                "content": ""
            },
            "finish_reason": "stop"
        }
    ],
    "usage": {
        "prompt_tokens": 0,
        "completion_tokens": 0,
        "total_tokens": 0
    }
}`
	t := new(ApiRespStr)
	err := json.Unmarshal([]byte(jsonStr), &t)
	t.Id = IdGenerator()
	t.Created = time.Now().Unix()
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	return t
}
func IdGenerator() string {
	const prefix = "chatcmpl-"
	const characters = "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var uniqueString strings.Builder

	rand.Seed(time.Now().UnixNano()) // 初始化随机数生成器
	for i := 0; i < 32; i++ {
		uniqueString.WriteByte(characters[rand.Intn(len(characters))])
	}

	log.Println("id_generator")
	return prefix + uniqueString.String()
}
func GetApiRespStrStream(id string) *ApiRespStrStream {
	jsonStr := `{
    "id": "chatcmpl-123",
    "object": "chat.completion.chunk",
    "created": 123,
    "model": "gpt-3.5-turbo",
	"system_fingerprint":null
    "choices": [
        {
            "delta": {
                "content": ""
            },
            "index": 0,
            "finish_reason": null
        }
    ]
}`
	t := new(ApiRespStrStream)
	err := json.Unmarshal([]byte(jsonStr), &t)
	t.Id = id
	t.Created = time.Now().Unix()
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	return t
}
func GetApiRespStrStreamEnd(id string) *ApiRespStrStreamEnd {
	jsonStr := `{
    "id": "chatcmpl-123",
    "object": "chat.completion.chunk",
    "created": 123,
    "model": "gpt-3.5-turbo",
    "choices": [
        {
            "delta": {},
            "index": 0,
            "finish_reason": "stop"
        }
    ]
}`
	t := new(ApiRespStrStreamEnd)
	err := json.Unmarshal([]byte(jsonStr), &t)
	t.Id = id
	t.Created = time.Now().Unix()
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	return t
}
func GetApiImageGenerationRespStr() *ApiImageGenerationRespStr {
	jsonStr := `{
    "created": 1700809991,
    "data": [
        {
            "revised_prompt": "",
            "url": ""
        }
    ]
}`
	t := new(ApiImageGenerationRespStr)
	err := json.Unmarshal([]byte(jsonStr), &t)
	t.Created = time.Now().Unix()
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	return t
}
