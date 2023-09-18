package services

import (
	"context"

	"chatgpt_server/models"
	"chatgpt_server/repos"
)

type ChatGPT interface {
	SendMsg(ctx context.Context, req models.ReqChatGPTFromCient) (*models.RespChatGPT, error)
}

type chatGPT struct {
	repo repos.ChatGPT
}

func NewChatGPT() ChatGPT {
	return &chatGPT{
		repos.NewChatGPT(),
	}
}

func (c chatGPT) SendMsg(ctx context.Context, req models.ReqChatGPTFromCient) (*models.RespChatGPT, error) {
	res, err := c.repo.SendMsg(ctx, req)
	if err != nil {
		return res, err
	}
	if len(res.Choices) == 0 {
		return res, err
	}
	for res.Choices[0].FinishReason == "length" {
		req.Message = append(req.Message, res.Choices[0].Message)
		nextRes, err := c.repo.SendMsg(ctx, req)
		if err != nil {
			return res, err
		}
		if len(res.Choices) == 0 {
			break
		}
		res.Choices[0].Message.Content += nextRes.Choices[0].Message.Content
		res.Choices[0].FinishReason = nextRes.Choices[0].FinishReason
		res.Usage = nextRes.Usage
	}
	return res, err
}
