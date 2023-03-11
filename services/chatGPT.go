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
	return c.repo.SendMsg(ctx, req)
}
