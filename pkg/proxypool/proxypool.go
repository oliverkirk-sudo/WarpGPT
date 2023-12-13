package proxypool

import (
	"WarpGPT/pkg/db"
	"WarpGPT/pkg/env"
	"WarpGPT/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	http "github.com/bogdanfinn/fhttp"
	"io"
	"strconv"
	"strings"
	"time"
)

type proxyUrl struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Data    []struct {
		Ip   string `json:"ip"`
		Port int    `json:"port"`
	} `json:"data"`
}

var ctx = context.Background()
var redisdb = db.RedisDB{}

// 检查代理池中的代理数量,如果数量不足,则从代理池中获取代理
func checkProxy() error {
	logger.Log.Debug("检查redis代理ip")
	client, err := redisdb.GetClient()
	if err != nil {
		return err
	}
	keys, err := client.Keys(ctx, "ip:*").Result()
	if err != nil {
		return err
	}
	if len(keys) < 20 {
		err = putIpsInRedis()
		if err != nil {
			return err
		}
	}
	return nil
}

func getProxyUrlList() (*proxyUrl, error) {
	logger.Log.Debug("请求代理ip池")
	poolUrl := env.Env.ProxyPoolUrl
	var proxy proxyUrl
	get, err := http.Get(poolUrl)
	if err != nil {
		return nil, err
	}
	all, err := io.ReadAll(get.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(all, &proxy)
	if err != nil {
		return nil, err
	}
	if proxy.Success {
		return &proxy, nil
	} else {
		return nil, errors.New("代理获取失败")
	}
}

// 从代理url中获取url,放入redis中
func putIpsInRedis() error {
	logger.Log.Debug("获取ip池并放入redis")
	proxyList, err := getProxyUrlList()
	client, err := redisdb.GetClient()
	if err != nil {
		return err
	}
	if err != nil {
		logger.Log.Fatal(err)
		return err
	}
	for _, ip := range proxyList.Data {
		ipstr := "http://" + ip.Ip + ":" + strconv.Itoa(ip.Port)
		_, err = client.Set(ctx, "ip:"+ipstr, "", time.Minute*3).Result()
		if err != nil {
			logger.Log.Error(err)
			return err
		}
	}
	return nil
}

func GetIpInRedis() (string, error) {
	logger.Log.Debug("请求代理ip")
	client, err := redisdb.GetClient()
	if err != nil {
		return "", err
	}
	statusCmd := client.RandomKey(ctx)
	result, err := statusCmd.Result()
	if err != nil {
		return "", err
	}
	size, err := client.DBSize(ctx).Result()
	if err != nil {
		return "", err
	}
	if size == 0 {
		logger.Log.Fatal("数据库为空,无法获取代理ip,尝试获取")
		err = putIpsInRedis()
		if err != nil {
			return "", err
		}
	}
	if strings.HasPrefix(result, "ip:") {
		client.Del(ctx, result)
		ip := strings.ReplaceAll(result, "ip:", "")
		logger.Log.Debug("获取的代理ip是: " + ip)
		return ip, nil
	} else {
		logger.Log.Warning("非代理ip键,跳过")
		ip, _ := GetIpInRedis()
		return ip, nil
	}
}

func ProxyThread() {
	if env.Env.ProxyPoolUrl == "" {
		logger.Log.Debug("未启动redis")
		return
	}
	logger.Log.Debug("启动redis监视线程")
	if err := checkProxy(); err != nil {
		return
	}
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := checkProxy()
			if err != nil {
				logger.Log.Fatal(err.Error())
				return
			}
		}
	}
}
