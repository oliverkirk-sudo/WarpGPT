package common

import (
	"WarpGPT/pkg/logger"
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
	Env = ENV{
		Proxy:      os.Getenv("proxy"),
		Port:       port,
		Host:       os.Getenv("host"),
		Verify:     verify,
		AuthKey:    os.Getenv("auth_key"),
		ArkoseMust: arkoseMust,
	}
}
