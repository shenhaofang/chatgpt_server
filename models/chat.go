package models

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type ReqGPT3 struct {
	Model             string    `json:"model"`
	Prompt            string    `json:"prompt"`
	Temperature       float64   `json:"temperature"`
	Max_tokens        int       `json:"max_tokens"`
	Top_p             float64   `json:"top_p"`
	Frequency_penalty int       `json:"frequency_penalty"`
	Presence_penalty  float64   `json:"presence_penalty"`
	Stop              [2]string `json:"stop"`
	N                 int       `json:"n"`
}

func (msg *ReqGPT3) ToJson() []byte {
	body, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return body
}

func CreateReqGPT3(req *ReqChat) *bytes.Buffer {
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
	reqGPT := &ReqGPT3{
		Model:             "text-davinci-003",
		Temperature:       0.9,
		Max_tokens:        150,
		Top_p:             1,
		Frequency_penalty: 0,
		Presence_penalty:  0.6,
		Stop: [2]string{
			"You: Bye",
			"AI: Bye",
		},
		Prompt: req.Prompt,
		N:      req.N,
	}
	if req.RoleAsker != "" {
		reqGPT.Stop[0] = req.RoleAsker + ": Bye"
	}
	if req.RoleAI != "" {
		reqGPT.Stop[1] = req.RoleAI + ": Bye"
	}
	reqGPT.Prompt += "\n" + reqGPT.Stop[0] + req.Msg + "\n" + reqGPT.Stop[1]
	req.Prompt = reqGPT.Prompt
	return bytes.NewBuffer(reqGPT.ToJson())
}

type RespGPT3 struct {
	Msg       string `json:"message"`
	Prompt    string `json:"prompt"`
	RoleAI    string `json:"role_ai"`
	RoleAsker string `json:"role_asker"`
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

func ToRespGPT3(req ReqChat, aipRes OpenAiRsp) *RespGPT3 {
	res := new(RespGPT3)
	res.Msg = aipRes.Choices[0].Text
	res.Prompt = req.Prompt + res.Msg
	// if aipRes.Choices[0].Finish_reason != "" {
	// 	res.N = 0
	// 	res.Prompt = ""
	// }
	res.RoleAsker = req.RoleAsker
	res.RoleAI = req.RoleAI
	return res
}

type ChatGPTMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// https://platform.openai.com/docs/api-reference/chat/create
type ReqChatGPT struct {
	Model            string           `json:"model"`
	Message          []ChatGPTMessage `json:"messages"`
	Temperature      float64          `json:"temperature"`
	TopP             float64          `json:"top_p"`
	N                int              `json:"n"`
	Stream           bool             `json:"stream"`
	Stop             []string         `json:"stop"`
	MaxTokens        int              `json:"max_tokens"`
	FrequencyPenalty int              `json:"frequency_penalty"`
	PresencePenalty  float64          `json:"presence_penalty"`
	User             string           `json:"user"`
}

type ReqChatGPTFromCient struct {
	ReqChatGPT
	UserID int64 `json:"user_id"`
}

func (msg ReqChatGPT) ToJson() []byte {
	body, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return body
}

func CreateReqChatGPT(req *ReqChatGPTFromCient) *bytes.Buffer {
	if req == nil {
		return nil
	}
	if req.Model == "" {
		req.Model = "gpt-3.5-turbo-0301" //gpt-3.5-turbo or gpt-3.5-turbo-0301
	}
	if req.UserID > 0 {
		req.User = fmt.Sprintf("client_user_%d", req.UserID)
	}
	if req.N < 1 {
		req.N = 1
	}
	if req.N > 5 {
		req.N = 5
	}
	if len(req.Message) == 0 || strings.TrimSpace(req.Message[0].Content) == "" {
		return nil
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 200
	}
	if req.MaxTokens > 4096 {
		req.MaxTokens = 4096
	}
	if req.Temperature < 0 || req.Temperature > 2 {
		req.Temperature = 0.9
	}
	if req.TopP > 0 {
		req.Temperature = 0
		if req.TopP > 1 {
			req.TopP = 1
		}
	}
	if req.Temperature == 0 && req.TopP == 0 {
		req.TopP = 1
	}
	if req.FrequencyPenalty > 2 || req.FrequencyPenalty < (-2) {
		req.FrequencyPenalty = 0
	}
	if req.PresencePenalty > 2 || req.PresencePenalty < (-2) {
		req.PresencePenalty = 0.6
	}

	return bytes.NewBuffer(req.ToJson())
}

type ChatChoice struct {
	Index        int `json:"index"`
	Message      ChatGPTMessage
	FinishReason string `json:"finish_reason"`
}

type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type RespChatGPT struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Choices []ChatChoice `json:"choices"`
	Usage   ChatUsage
}

func ToRespChatGPT(body []byte) (*RespChatGPT, error) {
	msg := new(RespChatGPT)
	err := json.Unmarshal(body, msg)
	if err != nil {
		return nil, err
	}
	return msg, err
}
