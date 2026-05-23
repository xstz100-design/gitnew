package services

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"wholesale/config"
	"wholesale/database"
	"wholesale/models"
	"wholesale/utils"

	"gorm.io/gorm"
)

// BotLoginConfirmFunc 由 handlers 包在 main 初始化时注入，避免循环引用
var BotLoginConfirmFunc func(token string, userID int64) bool

// ict 柬埔寨时区 (UTC+7)
var ict = time.FixedZone("ICT", 7*3600)

// formatScheduledAt 格式化预约时间，显示为柬埔寨时区
func formatScheduledAt(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.In(ict).Format("2006-01-02 15:04")
}

// NotifyAdminsNewOrder 新订单通知所有管理员
func NotifyAdminsNewOrder(orderNo, merchantName string, totalUSD float64, scheduledAt *time.Time, items []map[string]interface{}) {
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

	scheduledLine := ""
	if scheduledAt != nil {
		scheduledLine = fmt.Sprintf("\n⏰ 预约配送: <b>%s</b>", formatScheduledAt(scheduledAt))
	}

	msg := fmt.Sprintf(
		"🛒 <b>新订单</b>\n订单号: %s\n商户: %s\n金额: <b>$%.2f</b>%s\n\n商品:\n%s",
		orderNo, utils.EscapeHTML(merchantName), totalUSD, scheduledLine, itemLines,
	)

	var admins []models.User
	database.DB.Where("role = ? AND notify_enabled = ? AND telegram_id IS NOT NULL", models.RoleAdmin, true).Find(&admins)
	for _, admin := range admins {
		if admin.TelegramID != nil {
			utils.SendTelegramMessage(token, fmt.Sprintf("%d", *admin.TelegramID), msg, nil)
		}
	}
}

// NotifyPickerNewOrder 通知配货员拣货（发到群组）
func NotifyPickerNewOrder(orderNo string, orderID int64, scheduledAt *time.Time, items []map[string]interface{}) {
	token := config.C.TGBotToken
	if token == "" {
		return
	}

	groupChatID := GetGroupChatID()
	if groupChatID == "" {
		return
	}

	lines := ""
	var imageURLs []string
	for _, item := range items {
		name, _ := item["name"].(string)
		nameKh, _ := item["name_kh"].(string)
		nameEn, _ := item["name_en"].(string)
		qty, _ := item["quantity"].(int)
		barcode := ""
		if b, ok := item["barcode"].(*string); ok && b != nil {
			barcode = *b
		}
		unit := ""
		if u, ok := item["unit"].(string); ok {
			unit = u
		}

		displayName := utils.EscapeHTML(name)
		if nameKh != "" {
			displayName += " / " + utils.EscapeHTML(nameKh)
		}
		if nameEn != "" {
			displayName += " / " + utils.EscapeHTML(nameEn)
		}
		lines += fmt.Sprintf("  • %s %s × %d", displayName, unit, qty)
		if barcode != "" {
			lines += fmt.Sprintf(" [%s]", barcode)
		}
		lines += "\n"

		if img, ok := item["image"].(string); ok && img != "" {
			var fullURL string
			if img[0] == '/' {
				fullURL = config.C.SiteURL + img
			} else {
				fullURL = img
			}
			imageURLs = append(imageURLs, fullURL)
		}
	}

	markup := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{{"text": "✅ 已配货 / បានបែងចែករួច", "callback_data": fmt.Sprintf("picked:%d", orderID)}},
		},
	}

	scheduledLine := ""
	if scheduledAt != nil {
		scheduledLine = fmt.Sprintf("\n⏰ <b>预约配送 / ពេលដឹកដែល:</b> %s", formatScheduledAt(scheduledAt))
	}

	msg := fmt.Sprintf(
		"📦 <b>配货单 / ការដឹកជញជូន</b> #%s%s\n\n<b>商品 / តំនិঞ:</b>\n%s",
		orderNo, scheduledLine, lines,
	)
	utils.SendTelegramMediaGroup(token, groupChatID, imageURLs, msg, markup)
}

// NotifyDeliveryOrder 通知派送员（发到群组）——双语 + 地图链接 + 商户照片
func NotifyDeliveryOrder(order *models.Order) {
	token := config.C.TGBotToken
	if token == "" {
		return
	}
	groupChatID := GetGroupChatID()
	if groupChatID == "" {
		return
	}

	merchantName := ""
	address := ""
	phone := ""
	locationURL := ""
	storePhoto := ""
	if order.Merchant != nil {
		merchantName = order.Merchant.FullName
		if order.Merchant.LocationURL != nil {
			locationURL = *order.Merchant.LocationURL
		}
		if order.Merchant.StorePhoto != nil {
			sp := *order.Merchant.StorePhoto
			if len(sp) > 0 && sp[0] == '/' {
				storePhoto = config.C.SiteURL + sp
			} else {
				storePhoto = sp
			}
		}
	}
	if order.DeliveryAddress != nil {
		address = *order.DeliveryAddress
	}
	if order.DeliveryPhone != nil {
		phone = *order.DeliveryPhone
	}

	mapLine := ""
	if locationURL != "" {
		mapLine = fmt.Sprintf("\n🗺️ <b>地图 / ផែនទី:</b> <a href=\"%s\">点击导航 / ចុចនៅទីនេះ</a>", locationURL)
	}

	// 商品摘要
	itemSummary := ""
	for _, item := range order.Items {
		productName := "未知"
		if item.Product != nil {
			productName = item.Product.Name
		}
		itemSummary += fmt.Sprintf("  • %s × %d\n", utils.EscapeHTML(productName), item.Quantity)
	}

	khrRate := config.C.USDToKHRRate
	if khrRate <= 0 {
		khrRate = 4000
	}
	khr := order.TotalUSD * khrRate

	scheduledDeliveryLine := ""
	if order.ScheduledAt != nil {
		scheduledDeliveryLine = fmt.Sprintf("⏰ <b>预约配送 / ពេលដឹក:</b> %s\n", formatScheduledAt(order.ScheduledAt))
	}

	msg := fmt.Sprintf(
		"🚚 <b>派送任务 / ភារកិច្ចដឹកជញ្ជូន</b>\n"+
			"━━━━━━━━━━━━━━━\n"+
			"📋 <b>订单 / បញ្ជា:</b> #%s\n"+
			"🏪 <b>商户 / ឈ្មោះ:</b> %s\n"+
			"📞 <b>电话 / ទូរសព្ទ:</b> %s\n"+
			"📍 <b>地址 / អាសយដ្ឋាន:</b> %s%s\n"+
			"━━━━━━━━━━━━━━━\n"+
			"%s"+
			"%s"+
			"💰 <b>货款 / ថ្លៃ:</b> $%.2f (≈ %.0fR)\n"+
			"━━━━━━━━━━━━━━━",
		order.OrderNo,
		utils.EscapeHTML(merchantName),
		phone,
		utils.EscapeHTML(address),
		mapLine,
		scheduledDeliveryLine,
		itemSummary,
		order.TotalUSD,
		khr,
	)

	markup := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{{"text": "✅ 已送达 / បានដឹកជញ្ជូនរួច", "callback_data": fmt.Sprintf("delivered:%d", order.ID)}},
		},
	}
	utils.SendTelegramPhoto(token, groupChatID, storePhoto, msg, markup)
}

// notifyPaymentRequest 已送达后在群里提示派送员上传收款照片并确认
func notifyPaymentRequest(order *models.Order, token, groupChatID string) {
	merchantName := ""
	if order.Merchant != nil {
		merchantName = order.Merchant.FullName
	}
	khrRate := config.C.USDToKHRRate
	if khrRate <= 0 {
		khrRate = 4000
	}
	khr := order.TotalUSD * khrRate

	msg := fmt.Sprintf(
		"💰 <b>请确认收款 / សូមបញ្ជាក់ការទូទាត់</b>\n"+
			"━━━━━━━━━━━━━━━\n"+
			"📋 订单 / បញ្ជា: #%s\n"+
			"🏪 商户 / ឈ្មោះ: %s\n"+
			"💵 金额 / ចំនួនទឹកប្រាក់: <b>$%.2f</b> (≈ %.0fR)\n"+
			"━━━━━━━━━━━━━━━\n"+
			"请先将收款照片发送到群里，再点击下方按钮完成订单。\n"+
			"សូមផ្ញើរូបថតការទទួលប្រាក់ទៅក្នុងក្រុម រួចចុចប៊ូតុងខាងក្រោម។",
		order.OrderNo,
		utils.EscapeHTML(merchantName),
		order.TotalUSD,
		khr,
	)
	markup := map[string]interface{}{
		"inline_keyboard": [][]map[string]string{
			{{"text": "💰 确认收款完成 / បញ្ជាក់ការទូទាត់", "callback_data": fmt.Sprintf("confirm_pay:%d", order.ID)}},
		},
	}
	utils.SendTelegramMessage(token, groupChatID, msg, markup)
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

// CheckAndNotifyExpiringProducts 检查 30 天内过期商品并推送给管理员
func CheckAndNotifyExpiringProducts() {
	token := config.C.TGBotToken
	if token == "" {
		return
	}
	threshold := time.Now().AddDate(0, 0, 30)
	var products []models.Product
	database.DB.Where("is_deleted = ? AND expiry_date IS NOT NULL AND expiry_date <= ?", false, threshold).
		Order("expiry_date ASC").Find(&products)
	if len(products) == 0 {
		return
	}

	msg := "⚠️ <b>过期预警</b>（30天内到期）：\n"
	for i, p := range products {
		daysLeft := int(time.Until(*p.ExpiryDate).Hours() / 24)
		msg += fmt.Sprintf("%d. <b>%s</b> — 到期 %s（剩 %d 天）\n",
			i+1, utils.EscapeHTML(p.Name), p.ExpiryDate.Format("2006-01-02"), daysLeft)
	}

	var admins []models.User
	database.DB.Where("role = ? AND notify_enabled = ? AND telegram_id IS NOT NULL", models.RoleAdmin, true).Find(&admins)
	for _, admin := range admins {
		if admin.TelegramID != nil {
			utils.SendTelegramMessage(token, fmt.Sprintf("%d", *admin.TelegramID), msg, nil)
		}
	}
	log.Printf("[Expiry] 已推送过期预警，涉及 %d 个商品", len(products))
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

	// 解析 "action:orderID" —— Go 不支持 scanf 字符类，用 SplitN
	parts := strings.SplitN(data, ":", 2)
	if len(parts) != 2 {
		log.Printf("[Bot] 未知 callback_data: %s", data)
		return
	}
	action := parts[0]
	orderID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		log.Printf("[Bot] callback_data 解析失败: %s", data)
		return
	}

	var order models.Order
	if err := db.Preload("Merchant").Preload("Items.Product").First(&order, orderID).Error; err != nil {
		utils.AnswerCallbackQuery(token, callbackID, "订单不存在")
		return
	}

	groupChatID := GetGroupChatID()

	switch action {
	case "picked":
		// 原子 CAS：只允许 pending → delivering，防止重复回调竞争
		now := models.NowCambodia()
		result := db.Model(&models.Order{}).
			Where("id = ? AND delivery_status = ?", orderID, models.DeliveryPending).
			Updates(map[string]interface{}{
				"delivery_status": models.DeliveryDelivering,
				"picked_at":       now,
				"updated_at":      now,
			})
		if result.RowsAffected == 0 {
			utils.AnswerCallbackQuery(token, callbackID, "该订单已处理 / ការបញ្ជាទិញត្រូវបានដំណើរការ")
			return
		}
		utils.AnswerCallbackQuery(token, callbackID, "✅ 配货完成，通知派送员！")
		// 通知派送员（重新加载完整数据）
		if groupChatID != "" {
			var fullOrder models.Order
			if db.Preload("Merchant").Preload("Items.Product").First(&fullOrder, orderID).Error == nil {
				go NotifyDeliveryOrder(&fullOrder)
			}
		}

	case "delivered":
		// 只有 delivering 状态才能触发收款确认
		var checkOrder models.Order
		if db.Select("id, delivery_status").First(&checkOrder, orderID).Error != nil {
			utils.AnswerCallbackQuery(token, callbackID, "订单不存在")
			return
		}
		if checkOrder.DeliveryStatus == models.DeliveryDelivered {
			utils.AnswerCallbackQuery(token, callbackID, "订单已完成 / ការបញ្ជាទិញបានបញ្ចប់")
			return
		}
		// 提示派送员上传收款照片并点击确认按钮
		utils.AnswerCallbackQuery(token, callbackID, "💰 请拍照发到群里，再点击确认收款！")
		if groupChatID != "" {
			go notifyPaymentRequest(&order, token, groupChatID)
		}

	case "confirm_pay":
		// 原子 CAS：只允许 delivering → delivered，防止重复点击
		now := models.NowCambodia()
		result := db.Model(&models.Order{}).
			Where("id = ? AND delivery_status = ?", orderID, models.DeliveryDelivering).
			Updates(map[string]interface{}{
				"delivery_status": models.DeliveryDelivered,
				"payment_status":  models.PaymentCash,
				"delivered_at":    now,
				"updated_at":      now,
			})
		if result.RowsAffected == 0 {
			utils.AnswerCallbackQuery(token, callbackID, "订单已完成 / ការបញ្ជាទិញបានបញ្ចប់")
			return
		}
		utils.AnswerCallbackQuery(token, callbackID, "✅ 订单完成！收款已确认 / ការបញ្ជាទិញបានបញ្ចប់")

	default:
		utils.AnswerCallbackQuery(token, callbackID, "未知操作")
	}
}

// HandlePrivateMessage 处理用户私聊 Bot 的消息
// 用于：1) 收到 /start 时记录 chat_id；2) Bot 深链登录（自动注册商户）；3) 未绑定时引导关联
func HandlePrivateMessage(db *gorm.DB, tgID int64, text string, firstName string, lastName string) {
	token := config.C.TGBotToken
	if token == "" {
		return
	}
	chatIDStr := fmt.Sprintf("%d", tgID)
	trimmed := strings.TrimSpace(text)

	// 处理深链登录：/start login_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
	if strings.HasPrefix(trimmed, "/start login_") {
		loginToken := strings.TrimPrefix(trimmed, "/start login_")
		loginToken = strings.TrimSpace(loginToken)
		if loginToken != "" {
			// 查找该 TG ID 对应的账号
			var user models.User
			if err := db.Where("telegram_id = ? AND is_active = ?", tgID, true).First(&user).Error; err == nil {
				// 已有账号：调用 handlers 包的确认函数（通过接口避免循环引用）
				if BotLoginConfirmFunc != nil && BotLoginConfirmFunc(loginToken, user.ID) {
					msg := fmt.Sprintf("✅ 您好，<b>%s</b>！\n\n登录验证成功，请返回网页，已自动登录。", user.FullName)
					go utils.SendTelegramMessage(token, chatIDStr, msg, nil)
				} else {
					go utils.SendTelegramMessage(token, chatIDStr, "❌ 登录链接无效或已过期，请重新点击网页上的登录按钮。", nil)
				}
			} else {
				// 账号不存在——检查是否被禁用
				var disabledUser models.User
				if db.Where("telegram_id = ? AND is_active = ?", tgID, false).First(&disabledUser).Error == nil {
					go utils.SendTelegramMessage(token, chatIDStr, "❌ 您的账号已被禁用，请联系管理员。", nil)
					return
				}
				// 自动注册商户账号
				username := fmt.Sprintf("tg_%d", tgID)
				fullName := strings.TrimSpace(firstName + " " + lastName)
				if fullName == "" {
					fullName = username
				}
				hashed, hashErr := utils.HashPassword(utils.GenerateTemporaryPassword(16))
				if hashErr != nil {
					go utils.SendTelegramMessage(token, chatIDStr, "❌ 注册失败，请重试。", nil)
					return
				}
				newUser := models.User{
					Username:       username,
					HashedPassword: hashed,
					FullName:       fullName,
					Role:           "merchant",
					IsActive:       true,
					ApprovalStatus: "approved",
				}
				tgIDVal := tgID
				newUser.TelegramID = &tgIDVal
				chatIDStrVal := chatIDStr
				newUser.TelegramChatID = &chatIDStrVal
				if createErr := db.Create(&newUser).Error; createErr != nil {
					go utils.SendTelegramMessage(token, chatIDStr, "❌ 注册失败，请重试。", nil)
					return
				}
				// 完成登录
				if BotLoginConfirmFunc != nil && BotLoginConfirmFunc(loginToken, newUser.ID) {
					msg := fmt.Sprintf("✅ 欢迎！已为您自动创建账号并完成登录。\n\n用户名：<code>%s</code>\n请返回网页继续使用。", username)
					go utils.SendTelegramMessage(token, chatIDStr, msg, nil)
				} else {
					go utils.SendTelegramMessage(token, chatIDStr, "❌ 登录链接无效或已过期，请重新点击网页上的登录按钮。", nil)
				}
			}
			return
		}
	}

	// 检查该 TG ID 是否已绑定账号
	var user models.User
	if db.Where("telegram_id = ? AND is_active = ?", tgID, true).First(&user).Error == nil {
		// 已绑定，回复欢迎语（仅 /start 时）
		if trimmed == "/start" {
			msg := fmt.Sprintf("👋 您好，<b>%s</b>！\n\n您的账号已绑定 Telegram，可在网页登录页面点击「Telegram 登录」按钮快速登录。", user.FullName)
			go utils.SendTelegramMessage(token, chatIDStr, msg, nil)
		}
		return
	}

	// 未绑定，发送引导说明
	siteURL := config.C.SiteURL
	msg := fmt.Sprintf(
		"👋 您好！\n\n请联系管理员在 <b>%s</b> 为您创建账号，绑定您的 Telegram 后即可使用一键登录。",
		siteURL,
	)
	go utils.SendTelegramMessage(token, chatIDStr, msg, nil)
}
