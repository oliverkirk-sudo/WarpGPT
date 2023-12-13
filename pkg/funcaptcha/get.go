package funcaptcha

import (
	"WarpGPT/pkg/env"
	"WarpGPT/pkg/logger"
)

func GetOpenAIArkoseToken(arkType arkVer, puid string) (string, error) {
	logger.Log.Debug("GetArkoseToken")
	proxyArg := WithProxy(env.Env.Proxy)
	solver := NewSolver(proxyArg)
	WithHarpool(solver)
	token, err := solver.GetOpenAIToken(arkType, puid)
	if err != nil {
		logger.Log.Fatal(err)
		return "", err
	}
	return token, nil
}
