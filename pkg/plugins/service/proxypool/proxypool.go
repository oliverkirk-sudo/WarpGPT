package proxypool

import (
	"WarpGPT/pkg/db"
	"WarpGPT/pkg/plugins"
	ctx "context"
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

var context *plugins.Component
var redisdb db.DB
var ProxyPoolInstance ProxyPool

type ProxyPool struct {
}

// 检查代理池中的代理数量,如果数量不足,则从代理池中获取代理
func (p *ProxyPool) checkProxy() error {
	context.Logger.Debug("检查redis代理ip")
	client, err := redisdb.GetRedisClient()
	if err != nil {
		return err
	}
	keys, err := client.Keys(ctx.Background(), "ip:*").Result()
	if err != nil {
		return err
	}
	if len(keys) < 20 {
		err = p.putIpsInRedis()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *ProxyPool) getProxyUrlList() (*proxyUrl, error) {
	context.Logger.Debug("请求代理ip池")
	poolUrl := context.Env.ProxyPoolUrl
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
func (p *ProxyPool) putIpsInRedis() error {
	context.Logger.Debug("获取ip池并放入redis")
	proxyList, err := p.getProxyUrlList()
	client, err := redisdb.GetRedisClient()
	if err != nil {
		return err
	}
	if err != nil {
		context.Logger.Fatal(err)
		return err
	}
	for _, ip := range proxyList.Data {
		ipstr := "http://" + ip.Ip + ":" + strconv.Itoa(ip.Port)
		_, err = client.Set(ctx.Background(), "ip:"+ipstr, "", time.Minute*3).Result()
		if err != nil {
			context.Logger.Error(err)
			return err
		}
	}
	return nil
}

func (p *ProxyPool) GetIpInRedis() (string, error) {
	context.Logger.Debug("请求代理ip")
	client, err := redisdb.GetRedisClient()
	if err != nil {
		return "", err
	}
	statusCmd := client.RandomKey(ctx.Background())
	result, err := statusCmd.Result()
	if err != nil {
		return "", err
	}
	size, err := client.DBSize(ctx.Background()).Result()
	if err != nil {
		return "", err
	}
	if size == 0 {
		context.Logger.Fatal("数据库为空,无法获取代理ip,尝试获取")
		err = p.putIpsInRedis()
		if err != nil {
			return "", err
		}
	}
	if strings.HasPrefix(result, "ip:") {
		client.Del(ctx.Background(), result)
		ip := strings.ReplaceAll(result, "ip:", "")
		context.Logger.Debug("获取的代理ip是: " + ip)
		return ip, nil
	} else {
		context.Logger.Warning("非代理ip键,跳过")
		ip, _ := p.GetIpInRedis()
		return ip, nil
	}
}

func (p *ProxyPool) ProxyThread() {
	if context.Env.ProxyPoolUrl == "" {
		context.Logger.Debug("未启动redis")
		return
	}
	context.Logger.Debug("启动redis监视线程")
	if err := p.checkProxy(); err != nil {
		return
	}
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err := p.checkProxy()
			if err != nil {
				context.Logger.Fatal(err.Error())
				return
			}
		}
	}
}

func (p *ProxyPool) Run(com *plugins.Component) {
	context = com
	redisdb = context.Db
	go p.ProxyThread()
}
