package services

import (
	"context"

	"chatgpt_server/models"
	"chatgpt_server/repos"
)

type Chat interface {
	SendMsg(ctx context.Context, req models.ReqChat) (*models.RespGPT3, error)
}

type chat struct {
	repo repos.Chat
}

func NewChat() Chat {
	return &chat{
		repos.NewChat(),
	}
}

func (c chat) SendMsg(ctx context.Context, req models.ReqChat) (*models.RespGPT3, error) {
	return c.repo.SendMsg(ctx, req)
}
