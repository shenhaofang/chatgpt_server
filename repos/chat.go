package repos

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"meipian.cn/meigo/v2/config"
	"meipian.cn/meigo/v2/log"

	"chatgpt_server/models"
	"chatgpt_server/utils"
)

type GPTConfig struct {
	APIKey string
	Client *http.Client
}

type GPTClients []*GPTConfig

func (g GPTClients) Get(userID int64) *GPTConfig {
	return g[userID%int64(len(g))]
}

var gptClients = make(GPTClients, 0, 5)

type Chat interface {
	SendMsg(ctx context.Context, req models.ReqChat) (*models.RespGPT3, error)
}

type chat struct {
}

func getAPIKeys() []string {
	apiKeyStr := config.GetStr("default_api_keys")
	return strings.Split(apiKeyStr, ",")
}

const (
	DefaultRequestTimeout      = 30 * time.Second
	DefaultMaxIdleConns        = 1000
	DefaultMaxIdleConnsPerHost = 50
	DefaultMaxConnsPerHost     = 500
	DefaultIdleConnTimeout     = 20 * time.Minute
)

func InitChatGPTs() {
	proxyUrl := "http://127.0.0.1:7890"
	u, _ := url.Parse(proxyUrl)
	apiKeys := getAPIKeys()
	if len(apiKeys) == 0 {
		panic("no avalible api keys")
	}
	for _, apiKey := range apiKeys {
		gpt := &GPTConfig{
			APIKey: apiKey,
			Client: &http.Client{
				Transport: &http.Transport{
					MaxIdleConns:        DefaultMaxIdleConns,
					MaxIdleConnsPerHost: DefaultMaxIdleConnsPerHost,
					MaxConnsPerHost:     DefaultMaxConnsPerHost,
					IdleConnTimeout:     DefaultIdleConnTimeout,
					Proxy:               http.ProxyURL(u),
				},
				Timeout: DefaultRequestTimeout,
			},
		}
		gptClients = append(gptClients, gpt)
	}
}

func NewChat() Chat {
	return new(chat)
}

func (c chat) SendMsg(ctx context.Context, request models.ReqChat) (*models.RespGPT3, error) {
	gptClient := gptClients.Get(request.UserID)
	gptReq := models.CreateReqGPT3(&request)
	if gptReq == nil {
		return nil, utils.ErrorParamsInvalid
	}
	// 发送请求
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/completions", gptReq)
	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"req":   request,
			"error": err,
		}).Errorln("make request to send msg error")
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+gptClient.APIKey)
	// req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15")
	// req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")

	resp, err := gptClient.Client.Do(req)
	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"req":   request,
			"error": err,
		}).Errorln("send msg to chat gpt error")
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"req":   request,
			"resp":  string(bodyBytes),
			"error": err,
		}).Errorln("send msg to chat gpt error")
		return nil, err
	}
	rspData, err := models.ToRespOpenApi(bodyBytes)
	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"req":   request,
			"error": err,
			"resp":  string(bodyBytes),
		}).Errorln("gpt respose data error")
		return nil, err
	}
	if rspData.Error.Message != "" {
		// fmt.Printf("ChatGPT Server error:%v\n", rspData.Error.Message)
		log.WithCtxFields(ctx, log.Fields{
			"req":   request,
			"error": rspData.Error.Message,
			"resp":  string(bodyBytes),
		}).Errorln("ChatGPT Server error")
		return nil, utils.ErrorChatGPTError.NewWithMsg(rspData.Error.Message)
	}
	return models.ToRespGPT3(request, *rspData), nil
	// line := bytes.Split(bodyBytes, []byte("\n\n"))
	// if len(line) < 2 {
	// 	log.WithCtxFields(ctx, log.Fields{
	// 		"req":   request,
	// 		"error": err,
	// 		"resp":  string(bodyBytes),
	// 	}).Errorln("gpt respose data error")
	// 	fmt.Println("============================")
	// 	fmt.Println(string(bodyBytes))
	// 	return nil, err
	// }
	// endBlock := line[len(line)-3][6:]
	// return models.ToRespChatGPT(endBlock)

}
