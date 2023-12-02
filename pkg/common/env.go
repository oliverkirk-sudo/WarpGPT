package common

import (
	"WarpGPT/pkg/logger"
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type ENV struct {
	Proxy       string
	Port        int
	Host        string
	Verify      bool
	AuthKey     string
	ArkoseMust  bool
	OpenAI_HOST string
	UserAgent   string
}

var Env ENV

func init() {
	err := godotenv.Load()
	if err != nil {
		logger.Log.Fatalf("Error loading .env file: %v", err)
	}
	port, err := strconv.Atoi(os.Getenv("port"))
	if err != nil {
		port = 5000
	}
	verify, err := strconv.ParseBool(os.Getenv("verify"))
	if err != nil {
		verify = false
	}
	arkoseMust, err := strconv.ParseBool(os.Getenv("verify"))
	if err != nil {
		arkoseMust = false
	}
	OpenAI_HOST := os.Getenv("OpenAI_HOST")
	if err != nil {
		OpenAI_HOST = "chat.openai.com"
	}
	Env = ENV{
		Proxy:       os.Getenv("proxy"),
		Port:        port,
		Host:        os.Getenv("host"),
		Verify:      verify,
		AuthKey:     os.Getenv("auth_key"),
		ArkoseMust:  arkoseMust,
		OpenAI_HOST: OpenAI_HOST,
		UserAgent:   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Safari/605.1.15",
	}
}
