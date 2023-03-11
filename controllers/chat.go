package controllers

import (
	"github.com/gin-gonic/gin"

	"meipian.cn/meigo/v2/util"

	"chatgpt_server/models"
	"chatgpt_server/services"
	"chatgpt_server/utils"
)

type Chat struct {
	Srv        services.Chat
	ChatGPTSrv services.ChatGPT
}

func NewChat() *Chat {
	return &Chat{
		Srv:        services.NewChat(),
		ChatGPTSrv: services.NewChatGPT(),
	}
}

func (chat *Chat) SendMsg(c *gin.Context) {
	req := new(models.ReqChat)
	if err := c.ShouldBindJSON(req); err != nil {
		util.OutJsonErrMsg(c, utils.GetErrorCode(utils.ErrorParamsInvalid), utils.GetErrorMsg(utils.ErrorParamsInvalid))
		return
	}

	ctx := c.Request.Context()

	resp, err := chat.Srv.SendMsg(ctx, *req)
	if err != nil {
		util.OutJsonErrMsg(c, utils.GetErrorCode(utils.ErrorSystemError), utils.GetErrorMsg(utils.ErrorSystemError))
		return
	}
	util.OutJsonOk(c, resp)
}

func (chat *Chat) SendChatGPTMsg(c *gin.Context) {
	req := new(models.ReqChatGPTFromCient)
	if err := c.ShouldBindJSON(req); err != nil {
		util.OutJsonErrMsg(c, utils.GetErrorCode(utils.ErrorParamsInvalid), utils.GetErrorMsg(utils.ErrorParamsInvalid))
		return
	}

	ctx := c.Request.Context()

	resp, err := chat.ChatGPTSrv.SendMsg(ctx, *req)
	if err != nil {
		util.OutJsonErrMsg(c, utils.GetErrorCode(utils.ErrorSystemError), utils.GetErrorMsg(utils.ErrorSystemError))
		return
	}
	util.OutJsonOk(c, resp)
}
