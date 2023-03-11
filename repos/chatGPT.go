package repos

import (
	"context"
	"io"
	"net/http"

	"meipian.cn/meigo/v2/log"

	"chatgpt_server/models"
	"chatgpt_server/utils"
)

type ChatGPT interface {
	SendMsg(ctx context.Context, req models.ReqChatGPTFromCient) (*models.RespChatGPT, error)
}

type chatGPT struct {
}

func NewChatGPT() ChatGPT {
	return new(chatGPT)
}

func (c chatGPT) SendMsg(ctx context.Context, request models.ReqChatGPTFromCient) (*models.RespChatGPT, error) {
	chatgpt := gptClients.Get(request.UserID)
	gptReq := models.CreateReqChatGPT(&request)
	if gptReq == nil {
		return nil, utils.ErrorParamsInvalid
	}
	// 发送请求
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", gptReq)
	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"req":   request,
			"error": err,
		}).Errorln("make request to send msg error")
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+chatgpt.APIKey)
	// req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15")
	// req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")

	resp, err := chatgpt.Client.Do(req)
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
	rspData, err := models.ToRespChatGPT(bodyBytes)
	if err != nil {
		log.WithCtxFields(ctx, log.Fields{
			"req":   request,
			"error": err,
			"resp":  string(bodyBytes),
		}).Errorln("gpt respose data error")
		return nil, err
	}
	return rspData, nil
}
