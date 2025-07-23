package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ninja0404/meme-signal/pkg/logger"
)

// larkTextMessageContent 飞书文本消息内容结构
type larkTextMessageContent struct {
	Text string `json:"text"`
}

// larkMessage 飞书机器人消息结构
type larkMessage struct {
	MsgType string                 `json:"msg_type"`
	Content larkTextMessageContent `json:"content"`
}

// larkResponse 飞书机器人响应结构 (用于检查错误)
type larkResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// SendToLark 发送格式化后的文本消息到指定的飞书 Webhook URL
func SendToLark(messageText string, webhookURL string) error {
	if webhookURL == "" {
		return fmt.Errorf("飞书 Webhook URL 为空")
	}
	if messageText == "" {
		logger.Warn("尝试发送空消息到飞书，已跳过")
		return nil // 不视为错误，但记录警告
	}

	msg := larkMessage{
		MsgType: "text",
		Content: larkTextMessageContent{
			Text: messageText,
		},
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		logger.Error(fmt.Sprintf("序列化飞书消息失败: %v", err))
		return fmt.Errorf("序列化飞书消息失败: %w", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		logger.Error(fmt.Sprintf("创建飞书请求失败: %v", err))
		return fmt.Errorf("创建飞书请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("发送飞书消息失败: %v", err))
		return fmt.Errorf("发送飞书消息失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查飞书返回的状态码和内容
	if resp.StatusCode != http.StatusOK {
		var larkResp larkResponse
		if err := json.NewDecoder(resp.Body).Decode(&larkResp); err == nil {
			errMsg := fmt.Sprintf("发送飞书消息返回错误状态码 %d, Code: %d, Msg: %s", resp.StatusCode, larkResp.Code, larkResp.Msg)
			logger.Error(errMsg)
			return fmt.Errorf(errMsg)
		}
		errMsg := fmt.Sprintf("发送飞书消息返回错误状态码 %d, 无法解析响应体", resp.StatusCode)
		logger.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	// 检查响应体中的 code 是否为 0 (成功)
	var larkResp larkResponse
	if err := json.NewDecoder(resp.Body).Decode(&larkResp); err == nil {
		if larkResp.Code != 0 {
			errMsg := fmt.Sprintf("发送飞书消息成功，但飞书API返回错误 Code: %d, Msg: %s", larkResp.Code, larkResp.Msg)
			logger.Error(errMsg)
			// 根据需要决定是否将其视为 Go 函数的错误
			// return fmt.Errorf(errMsg)
		} else {
			logger.Info("成功发送消息到飞书")
		}
	} else {
		logger.Warn("发送飞书消息成功，但无法解析响应体")
		// 即使无法解析响应体，请求本身是成功的 (status 200)
	}

	return nil
}
