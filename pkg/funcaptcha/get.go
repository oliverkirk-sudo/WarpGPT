package funcaptcha

import (
	"WarpGPT/pkg/env"
	"WarpGPT/pkg/logger"
	"WarpGPT/pkg/plugins/service/proxypool"
)

func GetOpenAIArkoseToken(arkType int, puid string) (string, error) {
	logger.Log.Debug("GetArkoseToken")
	var proxyArg solverArg
	if env.Env.ProxyPoolUrl != "" {
		ip, err := proxypool.ProxyPoolInstance.GetIpInRedis()
		if err != nil {
			logger.Log.Warning(err.Error())
			return "", nil
		}
		proxyArg = WithProxy(ip)
	} else {
		proxyArg = WithProxy(env.Env.Proxy)
	}

	solver := NewSolver(proxyArg)
	WithHarpool(solver)
	token, err := solver.GetOpenAIToken(arkVer(arkType), puid)
	if err != nil {
		logger.Log.Warning(err)
		return "", err
	}
	return token, nil
}
