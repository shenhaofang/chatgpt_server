package controllers

import (
	"github.com/gin-gonic/gin"

	"meipian.cn/meigo/v2/util"

	"chatgpt_server/models"
	"chatgpt_server/services"
	"chatgpt_server/utils"
)

type Chat struct {
	Service services.Chat
}

func NewChat() *Chat {
	return &Chat{Service: services.NewChat()}
}

func (chat *Chat) SendMsg(c *gin.Context) {
	req := new(models.ReqChat)
	if err := c.ShouldBindJSON(req); err != nil {
		util.OutJsonErrMsg(c, utils.GetErrorCode(utils.ErrorParamsInvalid), utils.GetErrorMsg(utils.ErrorParamsInvalid))
		return
	}

	ctx := c.Request.Context()

	resp, err := chat.Service.SendMsg(ctx, *req)
	if err != nil {
		util.OutJsonErrMsg(c, utils.GetErrorCode(utils.ErrorSystemError), utils.GetErrorMsg(utils.ErrorSystemError))
	}
	util.OutJsonOk(c, resp)
}
