package common

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type ENV struct {
	Proxy      string
	Port       int
	Host       string
	Verify     bool
	AuthKey    string
	ArkoseMust bool
	OpenaiHost string
	UserAgent  string
	LogLevel   string
}

var Env ENV

func init() {
	err := godotenv.Load()
	if err != nil {
		return
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
	OpenaiHost := os.Getenv("openai_host")
	if err != nil {
		OpenaiHost = "chat.openai.com"
	}
	loglevel := os.Getenv("log_level")
	if loglevel == "" {
		loglevel = "info"
	}
	Env = ENV{
		Proxy:      os.Getenv("proxy"),
		Port:       port,
		Host:       os.Getenv("host"),
		Verify:     verify,
		AuthKey:    os.Getenv("auth_key"),
		ArkoseMust: arkoseMust,
		OpenaiHost: OpenaiHost,
		UserAgent:  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.4 Safari/605.1.15",
		LogLevel:   loglevel,
	}
}
