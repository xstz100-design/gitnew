package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"wholesale/config"
	"wholesale/services"

	"github.com/gin-gonic/gin"
)

type DeliveryFeeUpdateRequest struct {
	FreeDistance float64 `json:"free_distance_km"`
	FeePerKM     float64 `json:"fee_per_km_usd"`
}

type DeliveryFeeEstimateRequest struct {
	DistanceKM float64 `json:"distance_km" binding:"required"`
}

// GET /api/settings/delivery-fee
func GetDeliveryFee(c *gin.Context) {
	free, fee := services.GetDeliveryFeeSettings()
	c.JSON(http.StatusOK, gin.H{
		"free_distance_km": free,
		"fee_per_km_usd":   fee,
	})
}

// PATCH /api/settings/delivery-fee
func UpdateDeliveryFee(c *gin.Context) {
	var req DeliveryFeeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	services.SaveDeliveryFeeSettings(req.FreeDistance, req.FeePerKM)
	c.JSON(http.StatusOK, gin.H{
		"free_distance_km": req.FreeDistance,
		"fee_per_km_usd":   req.FeePerKM,
	})
}

// POST /api/settings/delivery-fee/estimate
func EstimateDeliveryFee(c *gin.Context) {
	var req DeliveryFeeEstimateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	free, fee := services.GetDeliveryFeeSettings()
	result := services.CalculateDeliveryFee(req.DistanceKM, free, fee)
	c.JSON(http.StatusOK, gin.H{
		"distance_km":      req.DistanceKM,
		"delivery_fee_usd": result,
	})
}

// GET /api/settings/role-chat-ids
func GetRoleChatIDs(c *gin.Context) {
	picker, delivery := services.GetRoleChatIDs("", "")
	c.JSON(http.StatusOK, gin.H{
		"picker_chat_id":   picker,
		"delivery_chat_id": delivery,
	})
}

// PUT /api/settings/role-chat-ids
func UpdateRoleChatIDs(c *gin.Context) {
	var req struct {
		PickerChatID   string `json:"picker_chat_id"`
		DeliveryChatID string `json:"delivery_chat_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	services.SaveRoleChatIDs(req.PickerChatID, req.DeliveryChatID)
	c.JSON(http.StatusOK, gin.H{
		"picker_chat_id":   req.PickerChatID,
		"delivery_chat_id": req.DeliveryChatID,
	})
}

// GET /api/settings/telegram-recent-chats — 调用 Telegram getUpdates 获取最近 chat 列表
func GetTelegramRecentChats(c *gin.Context) {
	token := config.C.TGBotToken
	if token == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "TG_BOT_TOKEN 未配置"})
		return
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?limit=100", token)
	resp, err := http.Get(apiURL) //nolint:gosec
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "无法连接 Telegram API"})
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "解析响应失败"})
		return
	}

	// 提取唯一的 chat 信息
	chats := map[string]interface{}{}
	if updates, ok := result["result"].([]interface{}); ok {
		for _, u := range updates {
			upd, ok := u.(map[string]interface{})
			if !ok {
				continue
			}
			// 检查 message 或 callback_query
			for _, key := range []string{"message", "callback_query"} {
				msg, ok := upd[key].(map[string]interface{})
				if !ok {
					continue
				}
				chatData, ok := msg["chat"].(map[string]interface{})
				if !ok {
					// callback_query 下是 message.chat
					if innerMsg, ok := msg["message"].(map[string]interface{}); ok {
						chatData, _ = innerMsg["chat"].(map[string]interface{})
					}
				}
				if chatData == nil {
					continue
				}
				if chatID, ok := chatData["id"]; ok {
					idStr := fmt.Sprintf("%v", chatID)
					if _, exists := chats[idStr]; !exists {
						title := ""
						if t, ok := chatData["title"].(string); ok {
							title = t
						} else if fn, ok := chatData["first_name"].(string); ok {
							title = fn
							if ln, ok := chatData["last_name"].(string); ok {
								title += " " + ln
							}
						}
						chats[idStr] = map[string]interface{}{
							"chat_id": idStr,
							"title":   title,
							"type":    chatData["type"],
						}
					}
				}
			}
		}
	}

	chatList := make([]interface{}, 0, len(chats))
	for _, v := range chats {
		chatList = append(chatList, v)
	}
	c.JSON(http.StatusOK, gin.H{"chats": chatList})
}

// GET /api/settings/contact-info — 公开
func GetContactInfo(c *gin.Context) {
	info := services.GetContactInfo()
	c.JSON(http.StatusOK, gin.H{
		"phone":    info.Phone,
		"telegram": info.Telegram,
		"whatsapp": info.Whatsapp,
	})
}

// PATCH /api/settings/contact-info — 管理员
func UpdateContactInfo(c *gin.Context) {
	var req struct {
		Phone    *string `json:"phone"`
		Telegram *string `json:"telegram"`
		Whatsapp *string `json:"whatsapp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}
	// 取当前值，类型为 string
	current := services.GetContactInfo()
	phone := current.Phone
	tg := current.Telegram
	wa := current.Whatsapp
	if req.Phone != nil {
		phone = *req.Phone
	}
	if req.Telegram != nil {
		tg = *req.Telegram
	}
	if req.Whatsapp != nil {
		wa = *req.Whatsapp
	}
	info := services.SaveContactInfo(phone, tg, wa)
	c.JSON(http.StatusOK, gin.H{
		"phone":    info.Phone,
		"telegram": info.Telegram,
		"whatsapp": info.Whatsapp,
	})
}
