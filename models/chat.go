package models

import (
	"bytes"
	"encoding/json"
	"strings"
)

type ReqChat struct {
	UserID    int64  `json:"user_id"`
	Msg       string `json:"msg"`
	Prompt    string `json:"prompt"`
	RoleAsker string `json:"role_asker"`
	RoleAI    string `json:"role_ai"`
	N         int    `json:"n"`
}

type ReqChatGPT struct {
	Model             string    `json:"model"`
	Prompt            string    `json:"prompt"`
	Temperature       float64   `json:"temperature"`
	Max_tokens        int       `json:"max_tokens"`
	Top_p             int       `json:"top_p"`
	Frequency_penalty int       `json:"frequency_penalty"`
	Presence_penalty  float64   `json:"presence_penalty"`
	Stop              [2]string `json:"stop"`
	N                 int       `json:"n"`
}

type ChatReqMessage struct {
	Id      string            `json:"id"`
	Role    string            `json:"role"`
	Content ChatReqMsgContent `json:"content"`
}

type ChatReqMsgContent struct {
	ContentType string   `json:"content_type"`
	Parts       []string `json:"parts"`
}

func ToReqChatGPT(body []byte) *ReqChatGPT {
	var msg ReqChatGPT
	err := json.Unmarshal(body, &msg)
	if err != nil {
		panic(err)
	}
	return &msg
}

func (msg *ReqChatGPT) ToJson() []byte {
	body, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return body
}

func CreateReqChatGPTBody(req *ReqChat) *bytes.Buffer {
	if req == nil {
		return nil
	}
	if req.N < 1 {
		req.N = 1
	}

	req.RoleAsker = strings.TrimSpace(req.RoleAsker)
	req.RoleAI = strings.TrimSpace(req.RoleAI)

	if req.Msg == "" || strings.TrimSpace(req.Msg) == "" {
		return nil
	}
	reqGPT := &ReqChatGPT{
		Model:             "text-davinci-003",
		Temperature:       0.9,
		Max_tokens:        150,
		Top_p:             1,
		Frequency_penalty: 0,
		Presence_penalty:  0.6,
		Stop: [2]string{
			"You: ",
			"AI: ",
		},
		Prompt: req.Prompt,
		N:      req.N,
	}
	if req.RoleAI != "" {
		reqGPT.Stop[0] = req.RoleAsker + ": "
	}
	if req.RoleAI != "" {
		reqGPT.Stop[1] = req.RoleAI + ":"
	}
	reqGPT.Prompt += "\n" + reqGPT.Stop[0] + req.Msg + "\n" + reqGPT.Stop[1]
	req.Prompt = reqGPT.Prompt
	return bytes.NewBuffer(reqGPT.ToJson())
}

type RespChatGPT struct {
	// Message        ChatGPTMessage `json:"message"`
	// ConversationId string         `json:"conversation_id"`
	Msg       string `json:"message"`
	N         int    `json:"n"`
	Prompt    string `json:"prompt"`
	RoleAI    string `json:"role_ai"`
	RoleAsker string `json:"role_asker"`
}

type ChatGPTMessage struct {
	Id      string            `json:"id"`
	Content ChatResMsgContent `json:"content"`
}

type ChatResMsgContent struct {
	Parts []string `json:"parts"`
}

type OpenAiChoices struct {
	Text          string `json:"text"`
	Finish_reason string `json:"finish_reason"`
}

type OpenApiError struct {
	Message string `json:"message"`
}

type OpenAiRsp struct {
	Choices []OpenAiChoices `json:"choices"`
	Error   OpenApiError    `json:"error"`
}

func ToRespOpenApi(body []byte) (*OpenAiRsp, error) {
	var msg OpenAiRsp
	err := json.Unmarshal(body, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, err
}

func ToRespChatGPT(req ReqChat, aipRes OpenAiRsp) *RespChatGPT {
	res := new(RespChatGPT)
	res.Msg = aipRes.Choices[0].Text
	res.N = req.N + 1
	res.Prompt = req.Prompt + res.Msg
	// if aipRes.Choices[0].Finish_reason != "" {
	// 	res.N = 0
	// 	res.Prompt = ""
	// }
	res.RoleAsker = req.RoleAsker
	res.RoleAI = req.RoleAI
	return res
}
