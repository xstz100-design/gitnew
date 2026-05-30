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

	scheduledDeliveryLine := ""
	if order.ScheduledAt != nil {
		scheduledDeliveryLine = fmt.Sprintf("⏰ <b>预约配送 / ពេលដឹក:</b> %s\n", formatScheduledAt(order.ScheduledAt))
	}

	var paymentSection string
	if order.DeliveryFeeUSD > 0 {
		paymentSection = fmt.Sprintf(
			"💰 <b>货款 / ថ្លៃទំនិញ:</b> $%.2f (≈ %.0fR)\n"+
				"🚚 <b>派送费 / ថ្លៃដឹក:</b> $%.2f (≈ %.0fR)\n"+
				"💵 <b>应收合计 / សរុប:</b> <b>$%.2f</b> (≈ %.0fR)\n",
			order.GoodsTotalUSD, order.GoodsTotalUSD*khrRate,
			order.DeliveryFeeUSD, order.DeliveryFeeUSD*khrRate,
			order.TotalUSD, order.TotalUSD*khrRate,
		)
	} else {
		paymentSection = fmt.Sprintf(
			"💰 <b>货款 / ថ្លៃ:</b> $%.2f (≈ %.0fR)\n",
			order.TotalUSD, order.TotalUSD*khrRate,
		)
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
			"%s"+
			"━━━━━━━━━━━━━━━",
		order.OrderNo,
		utils.EscapeHTML(merchantName),
		phone,
		utils.EscapeHTML(address),
		mapLine,
		scheduledDeliveryLine,
		itemSummary,
		paymentSection,
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

	var amountLines string
	if order.DeliveryFeeUSD > 0 {
		amountLines = fmt.Sprintf(
			"💰 货款 / ថ្លៃទំនិញ: $%.2f (≈ %.0fR)\n"+
				"🚚 派送费 / ថ្លៃដឹក: $%.2f (≈ %.0fR)\n"+
				"💵 <b>应收合计 / សរុប: $%.2f</b> (≈ %.0fR)\n",
			order.GoodsTotalUSD, order.GoodsTotalUSD*khrRate,
			order.DeliveryFeeUSD, order.DeliveryFeeUSD*khrRate,
			order.TotalUSD, order.TotalUSD*khrRate,
		)
	} else {
		amountLines = fmt.Sprintf(
			"💵 <b>金额 / ចំនួនទឹកប្រាក់: $%.2f</b> (≈ %.0fR)\n",
			order.TotalUSD, order.TotalUSD*khrRate,
		)
	}

	msg := fmt.Sprintf(
		"💰 <b>请确认收款 / សូមបញ្ជាក់ការទូទាត់</b>\n"+
			"━━━━━━━━━━━━━━━\n"+
			"📋 订单 / បញ្ជា: #%s\n"+
			"🏪 商户 / ឈ្មោះ: %s\n"+
			"%s"+
			"━━━━━━━━━━━━━━━\n"+
			"请先将收款照片发送到群里，再点击下方按钮完成订单。\n"+
			"សូមផ្ញើរូបថតការទទួលប្រាក់ទៅក្នុងក្រុម រួចចុចប៊ូតុងខាងក្រោម។",
		order.OrderNo,
		utils.EscapeHTML(merchantName),
		amountLines,
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
		// 通知商户本人：订单已送达
		if order.Merchant != nil && order.Merchant.TelegramChatID != nil {
			merchantChatID := *order.Merchant.TelegramChatID
			merchantName := utils.EscapeHTML(order.Merchant.FullName)
			khrRate := config.C.USDToKHRRate
			if khrRate <= 0 {
				khrRate = 4000
			}
			merchantMsg := fmt.Sprintf(
				"✅ <b>订单已送达！</b>\nការបញ្ជាទិញបានបញ្ចប់！\n\n"+
					"📋 订单号 / លេខបញ្ជា: <b>#%s</b>\n"+
					"💵 金额 / ចំនួន: <b>$%.2f</b>（≈ %.0f R）\n\n"+
					"感谢 <b>%s</b> 的惠顾！\nអរគុណ <b>%s</b>！",
				order.OrderNo,
				order.TotalUSD, order.TotalUSD*khrRate,
				merchantName, merchantName,
			)
			go utils.SendTelegramMessage(token, merchantChatID, merchantMsg,
				miniAppMarkup(config.C.SiteURL))
		}

	default:
		utils.AnswerCallbackQuery(token, callbackID, "未知操作")
	}
}

// miniAppMarkup 生成打开 Miniapp 商城的内联键盘按钮
func miniAppMarkup(siteURL string) map[string]interface{} {
	return map[string]interface{}{
		"inline_keyboard": [][]map[string]interface{}{
			{{"text": "🛒 打开商城 / ហាងលក់ដុំ", "web_app": map[string]string{"url": siteURL + "/m/shop"}}},
		},
	}
}

// HandlePrivateMessage 处理用户私聊 Bot 的消息
func HandlePrivateMessage(db *gorm.DB, tgID int64, text string, firstName string, lastName string) {
	token := config.C.TGBotToken
	if token == "" {
		return
	}
	chatIDStr := fmt.Sprintf("%d", tgID)
	trimmed := strings.TrimSpace(text)
	siteURL := config.C.SiteURL
	shopBtn := miniAppMarkup(siteURL)

	// ── 深链登录：/start login_TOKEN ──────────────────────────────
	if strings.HasPrefix(trimmed, "/start login_") {
		loginToken := strings.TrimSpace(strings.TrimPrefix(trimmed, "/start login_"))
		if loginToken == "" {
			return
		}

		var user models.User
		if db.Where("telegram_id = ? AND is_active = ?", tgID, true).First(&user).Error == nil {
			// 已有绑定账号
			if BotLoginConfirmFunc != nil && BotLoginConfirmFunc(loginToken, user.ID) {
				msg := fmt.Sprintf(
					"✅ 您好，<b>%s</b>！登录成功，请返回商城继续使用。\n"+
						"ចូលជោគជ័យ <b>%s</b>！",
					utils.EscapeHTML(user.FullName), utils.EscapeHTML(user.FullName),
				)
				go utils.SendTelegramMessage(token, chatIDStr, msg, shopBtn)
			} else {
				go utils.SendTelegramMessage(token, chatIDStr,
					"❌ 登录链接无效或已过期，请重新点击网页上的登录按钮。\nតំណអស់សុពលភាព។", nil)
			}
			return
		}

		// 账号不存在 — 检查是否被禁用
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
		tgIDVal := tgID
		chatIDStrVal := chatIDStr
		newUser := models.User{
			Username:       username,
			HashedPassword: hashed,
			FullName:       fullName,
			Role:           models.RoleMerchant,
			IsActive:       true,
			ApprovalStatus: models.ApprovalApproved,
			TelegramID:     &tgIDVal,
			TelegramChatID: &chatIDStrVal,
		}
		if db.Create(&newUser).Error != nil {
			go utils.SendTelegramMessage(token, chatIDStr, "❌ 注册失败，请重试。", nil)
			return
		}
		if BotLoginConfirmFunc != nil && BotLoginConfirmFunc(loginToken, newUser.ID) {
			msg := fmt.Sprintf(
				"🎉 欢迎加入<b>东方优选</b>！\n" +
					"ស្វាគមន៍！\n\n" +
					"已为您自动创建账号，点击下方按钮立即开始批发下单 👇",
			)
			go utils.SendTelegramMessage(token, chatIDStr, msg, shopBtn)
		} else {
			go utils.SendTelegramMessage(token, chatIDStr,
				"❌ 登录链接无效或已过期，请重新点击网页上的登录按钮。", nil)
		}
		return
	}

	// ── 查找已绑定账号 ────────────────────────────────────────────
	var user models.User
	bound := db.Where("telegram_id = ? AND is_active = ?", tgID, true).First(&user).Error == nil

	if !bound {
		// 未绑定 — 仅在 /start 或 /help 时回复引导语，其余静默
		if trimmed == "/start" || strings.HasPrefix(trimmed, "/help") || trimmed == "帮助" {
			msg := "👋 欢迎使用 <b>东方优选</b> 批发商城！\n" +
				"ស្វាគមន៍ <b>Dongfang Youxuan</b>！\n\n" +
				"点击下方按钮打开商城，注册账号后即可批发下单。\n" +
				"ចុចប៊ូតុងខាងក្រោម ដើម្បីបើកហាង និងចុះឈ្មោះ។"
			go utils.SendTelegramMessage(token, chatIDStr, msg, shopBtn)
		}
		return
	}

	// ── 已绑定账号，按指令分流 ─────────────────────────────────────
	switch {
	case trimmed == "/start":
		go botSendWelcome(db, &user, token, chatIDStr, shopBtn)

	case trimmed == "/orders" || trimmed == "我的订单" || trimmed == "订单" || trimmed == "order":
		go botSendOrders(db, &user, token, chatIDStr, shopBtn)

	case trimmed == "/help" || trimmed == "帮助" || trimmed == "help":
		go botSendHelp(token, chatIDStr, shopBtn, user.Role == models.RoleAdmin)

	default:
		// 其他任何消息 — 回复商城入口
		name := utils.EscapeHTML(user.FullName)
		msg := fmt.Sprintf("您好 <b>%s</b>！点击下方进入商城 👇\nសួស្ដី <b>%s</b>！", name, name)
		go utils.SendTelegramMessage(token, chatIDStr, msg, shopBtn)
	}
}

// botSendWelcome 发送欢迎消息（根据角色与审核状态）
func botSendWelcome(db *gorm.DB, user *models.User, token, chatIDStr string, shopBtn interface{}) {
	name := utils.EscapeHTML(user.FullName)

	if user.Role == models.RoleAdmin {
		msg := fmt.Sprintf(
			"👋 管理员 <b>%s</b>，您好！\n\n"+
				"您已绑定此 Bot，将收到新订单、库存预警等推送通知。\n\n"+
				"/orders — 查看待处理订单\n"+
				"/help — 指令帮助",
			name,
		)
		utils.SendTelegramMessage(token, chatIDStr, msg, shopBtn)
		return
	}

	// 商户 — 根据审核状态
	switch user.ApprovalStatus {
	case models.ApprovalPending:
		msg := fmt.Sprintf(
			"👋 您好，<b>%s</b>！\n\n"+
				"⏳ 您的账号正在等待管理员审核，审核通过后即可下单。\n"+
				"គណនីរបស់អ្នកកំពុងរង់ចាំការពិនិត្យ។\n\n"+
				"您可以先浏览商品，购物车商品会为您保留。",
			name,
		)
		utils.SendTelegramMessage(token, chatIDStr, msg, shopBtn)

	case models.ApprovalRejected:
		reason := ""
		if user.RejectedReason != nil && *user.RejectedReason != "" {
			reason = "\n原因：" + utils.EscapeHTML(*user.RejectedReason)
		}
		msg := fmt.Sprintf(
			"⚠️ 您好，<b>%s</b>！\n\n"+
				"您的账号审核未通过。%s\n\n"+
				"请在商城中更新资料后重新提交审核。\n"+
				"សូមធ្វើបច្ចុប្បន្នភាពព័ត៌មានរបស់អ្នក។",
			name, reason,
		)
		utils.SendTelegramMessage(token, chatIDStr, msg, shopBtn)

	default: // approved
		var pendingCount int64
		db.Model(&models.Order{}).
			Where("merchant_id = ? AND delivery_status = ? AND is_deleted = ?",
				user.ID, models.DeliveryPending, false).
			Count(&pendingCount)

		pendingLine := ""
		if pendingCount > 0 {
			pendingLine = fmt.Sprintf("\n📦 当前有 <b>%d</b> 个订单等待配货。", pendingCount)
		}

		msg := fmt.Sprintf(
			"👋 欢迎回来，<b>%s</b>！\nស្វាគមន៍ <b>%s</b>！%s\n\n"+
				"发送 /orders 查看最近订单，或点击下方打开商城 👇",
			name, name, pendingLine,
		)
		utils.SendTelegramMessage(token, chatIDStr, msg, shopBtn)
	}
}

// botSendOrders 发送最近订单列表
func botSendOrders(db *gorm.DB, user *models.User, token, chatIDStr string, shopBtn interface{}) {
	dsText := map[models.DeliveryStatus]string{
		models.DeliveryPending:    "⏳待配货",
		models.DeliveryDelivering: "🚚派送中",
		models.DeliveryDelivered:  "✅已送达",
		models.DeliveryCancelled:  "❌已取消",
	}
	psText := map[models.PaymentStatus]string{
		models.PaymentUnpaid: "未结款",
		models.PaymentCash:   "现结",
	}

	if user.Role == models.RoleAdmin {
		// 管理员查待处理订单
		var orders []models.Order
		db.Preload("Merchant").
			Where("delivery_status = ? AND is_deleted = ?", models.DeliveryPending, false).
			Order("created_at DESC").Limit(8).Find(&orders)
		if len(orders) == 0 {
			utils.SendTelegramMessage(token, chatIDStr, "✅ 当前没有待配货订单。\nគ្មានការបញ្ជាទិញ។", shopBtn)
			return
		}
		msg := fmt.Sprintf("📋 <b>待配货订单（%d条）</b>\n\n", len(orders))
		for _, o := range orders {
			mn := ""
			if o.Merchant != nil {
				mn = utils.EscapeHTML(o.Merchant.FullName)
			}
			msg += fmt.Sprintf("• <b>#%s</b> %s $%.2f\n", o.OrderNo, mn, o.TotalUSD)
		}
		utils.SendTelegramMessage(token, chatIDStr, msg, shopBtn)
		return
	}

	// 商户查自己的订单
	var orders []models.Order
	db.Where("merchant_id = ? AND is_deleted = ?", user.ID, false).
		Order("created_at DESC").Limit(5).Find(&orders)
	if len(orders) == 0 {
		utils.SendTelegramMessage(token, chatIDStr,
			"您还没有订单记录。\nអ្នកមិនទាន់មានបញ្ជាទិញ។", shopBtn)
		return
	}

	msg := "<b>您的最近订单</b> / បញ្ជាទិញ\n\n"
	for _, o := range orders {
		ds := dsText[o.DeliveryStatus]
		if ds == "" {
			ds = string(o.DeliveryStatus)
		}
		ps := psText[o.PaymentStatus]
		if ps == "" {
			ps = string(o.PaymentStatus)
		}
		msg += fmt.Sprintf("• <b>#%s</b> $%.2f | %s | %s\n",
			o.OrderNo, o.TotalUSD, ds, ps)
	}
	utils.SendTelegramMessage(token, chatIDStr, msg, shopBtn)
}

// botSendHelp 发送帮助消息
func botSendHelp(token, chatIDStr string, shopBtn interface{}, isAdmin bool) {
	msg := "📖 <b>指令帮助</b>\n\n" +
		"/start — 欢迎页\n" +
		"/orders — 查看最近订单\n" +
		"/help — 帮助\n\n"
	if isAdmin {
		msg += "您是管理员，将自动收到订单通知与库存预警推送。"
	} else {
		msg += "点击下方按钮打开商城批发下单 👇\nចុចប៊ូតុង ដើម្បីបើកហាង។"
	}
	utils.SendTelegramMessage(token, chatIDStr, msg, shopBtn)
}
