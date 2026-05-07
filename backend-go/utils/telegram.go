package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// SendTelegramMessage 同步发送 Telegram 消息，返回 message_id（失败返回 0）
func SendTelegramMessage(botToken, chatID, text string, replyMarkup interface{}) int64 {
	if botToken == "" || chatID == "" {
		return 0
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}

	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[Telegram] 发送失败: %v", err)
		return 0
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBytes, &result); err != nil || !result.OK {
		log.Printf("[Telegram] 响应异常: %s", string(respBytes))
		return 0
	}
	return result.Result.MessageID
}

// AnswerCallbackQuery 响应 inline button 点击
func AnswerCallbackQuery(botToken, callbackQueryID, text string) {
	if botToken == "" {
		return
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", botToken)
	payload := map[string]interface{}{
		"callback_query_id": callbackQueryID,
		"text":              text,
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[Telegram] answerCallbackQuery 失败: %v", err)
		return
	}
	resp.Body.Close()
}

// EscapeHTML HTML 转义
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
