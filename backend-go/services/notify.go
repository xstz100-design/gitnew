package services

import (
	"fmt"
	"log"
	"wholesale/config"
	"wholesale/database"
	"wholesale/models"
	"wholesale/utils"

	"gorm.io/gorm"
)

// NotifyAdminsNewOrder 新订单通知所有管理员
func NotifyAdminsNewOrder(orderNo, merchantName string, totalUSD float64, items []map[string]interface{}) {
	token := config.C.TGBotToken
	if token == "" {
		return
	}

	itemLines := ""
	for _, item := range items {
		name, _ := item["product_name"].(string)
		qty, _ := item["quantity"].(int)
		sub, _ := item["subtotal_usd"].(float64)
		itemLines += fmt.Sprintf("  • %s × %d = $%.2f\n", utils.EscapeHTML(name), qty, sub)
	}

	msg := fmt.Sprintf(
		"🛒 <b>新订单</b>\n订单号: %s\n商户: %s\n金额: <b>$%.2f</b>\n\n商品:\n%s",
		orderNo, utils.EscapeHTML(merchantName), totalUSD, itemLines,
	)

	var admins []models.User
	database.DB.Where("role = ? AND notify_enabled = ? AND telegram_id IS NOT NULL", models.RoleAdmin, true).Find(&admins)
	for _, admin := range admins {
		if admin.TelegramID != nil {
			utils.SendTelegramMessage(token, fmt.Sprintf("%d", *admin.TelegramID), msg, nil)
		}
	}
}

// NotifyPickerNewOrder 通知配货员拣货
func NotifyPickerNewOrder(orderNo string, orderID int64, items []map[string]interface{}) {
	token := config.C.TGBotToken
	if token == "" {
		return
	}

	pickerChatID, _ := GetRoleChatIDs(config.C.TelegramPickerChatID, config.C.TelegramDeliveryChatID)
	if pickerChatID == "" {
		return
	}

	lines := ""
	for _, item := range items {
		name, _ := item["name"].(string)
		qty, _ := item["quantity"].(int)
		barcode := ""
		if b, ok := item["barcode"].(string); ok {
			barcode = b
		}
		unit := ""
		if u, ok := item["unit"].(string); ok {
			unit = u
		}
		lines += fmt.Sprintf("  • %s %s × %d", utils.EscapeHTML(name), unit, qty)
		if barcode != "" {
			lines += fmt.Sprintf(" [%s]", barcode)
		}
		lines += "\n"
	}

	// 添加 inline button: 已配货确认
	markup := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{{"text": "✅ 已配货", "callback_data": fmt.Sprintf("picked:%d", orderID)}},
		},
	}

	msg := fmt.Sprintf("📦 <b>配货单</b> #%s\n\n%s", orderNo, lines)
	utils.SendTelegramMessage(token, pickerChatID, msg, markup)
}

// NotifyDeliveryOrder 通知配送员
func NotifyDeliveryOrder(orderNo string, orderID int64, merchantName, address, phone string) {
	token := config.C.TGBotToken
	if token == "" {
		return
	}
	_, deliveryChatID := GetRoleChatIDs(config.C.TelegramPickerChatID, config.C.TelegramDeliveryChatID)
	if deliveryChatID == "" {
		return
	}

	markup := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{{"text": "✅ 已送达", "callback_data": fmt.Sprintf("delivered:%d", orderID)}},
		},
	}

	msg := fmt.Sprintf(
		"🚚 <b>配送单</b> #%s\n商户: %s\n地址: %s\n电话: %s",
		orderNo, utils.EscapeHTML(merchantName), utils.EscapeHTML(address), phone,
	)
	utils.SendTelegramMessage(token, deliveryChatID, msg, markup)
}

// NotifyLowStock 库存预警通知
func NotifyLowStock(productName string, stock int) {
	token := config.C.TGBotToken
	if token == "" {
		return
	}
	msg := fmt.Sprintf("⚠️ 库存预警: <b>%s</b> 当前库存 %d，请及时补货！", utils.EscapeHTML(productName), stock)

	var admins []models.User
	database.DB.Where("role = ? AND notify_enabled = ? AND telegram_id IS NOT NULL", models.RoleAdmin, true).Find(&admins)
	for _, admin := range admins {
		if admin.TelegramID != nil {
			utils.SendTelegramMessage(token, fmt.Sprintf("%d", *admin.TelegramID), msg, nil)
		}
	}
}

// HandleBotCallback 处理 Telegram inline button 回调
func HandleBotCallback(db *gorm.DB, update map[string]interface{}) {
	token := config.C.TGBotToken
	cq, ok := update["callback_query"].(map[string]interface{})
	if !ok {
		return
	}
	callbackID, _ := cq["id"].(string)
	data, _ := cq["data"].(string)
	if data == "" {
		return
	}

	var orderID int64
	var action string
	n, err := fmt.Sscanf(data, "%[^:]:%d", &action, &orderID)
	if err != nil || n != 2 {
		log.Printf("[Bot] 未知 callback_data: %s", data)
		return
	}

	var order models.Order
	if err := db.First(&order, orderID).Error; err != nil {
		utils.AnswerCallbackQuery(token, callbackID, "订单不存在")
		return
	}

	switch action {
	case "picked":
		if order.DeliveryStatus != models.DeliveryPending {
			utils.AnswerCallbackQuery(token, callbackID, "该订单已处理")
			return
		}
		now := models.NowCambodia()
		db.Model(&order).Updates(map[string]interface{}{
			"delivery_status": models.DeliveryDelivering,
			"picked_at":       now,
		})
		utils.AnswerCallbackQuery(token, callbackID, "✅ 已标记为配货中")

	case "delivered":
		if order.DeliveryStatus == models.DeliveryDelivered {
			utils.AnswerCallbackQuery(token, callbackID, "该订单已送达")
			return
		}
		now := models.NowCambodia()
		db.Model(&order).Updates(map[string]interface{}{
			"delivery_status": models.DeliveryDelivered,
			"delivered_at":    now,
		})
		utils.AnswerCallbackQuery(token, callbackID, "✅ 已标记为已送达")

	default:
		utils.AnswerCallbackQuery(token, callbackID, "未知操作")
	}
}
