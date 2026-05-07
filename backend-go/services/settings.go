package services

import (
	"strconv"
	"wholesale/database"
	"wholesale/models"
)

const (
	KeyDeliveryFreeDistance = "delivery.free_distance_km"
	KeyDeliveryFeePerKM     = "delivery.fee_per_extra_km_usd"
	KeyPickerChatID         = "telegram.picker_chat_id"
	KeyDeliveryChatID       = "telegram.delivery_chat_id"
	KeyContactPhone         = "contact.phone"
	KeyContactTelegram      = "contact.telegram"
	KeyContactWhatsapp      = "contact.whatsapp"

	DefaultFreeDistanceKM = 3.0
	DefaultFeePerKM       = 0.5
)

func GetSetting(key string) string {
	var s models.SystemSetting
	if err := database.DB.Where("key = ?", key).First(&s).Error; err != nil {
		return ""
	}
	return s.Value
}

func UpsertSetting(key, value string) {
	var s models.SystemSetting
	if database.DB.Where("key = ?", key).First(&s).Error == nil {
		database.DB.Model(&s).Update("value", value)
		database.DB.Model(&s).Update("updated_at", models.NowCambodia())
	} else {
		database.DB.Create(&models.SystemSetting{Key: key, Value: value, UpdatedAt: models.NowCambodia()})
	}
}

// ─────────────────── 配送费 ───────────────────

func GetDeliveryFeeSettings() (freeKM, feePerKM float64) {
	freeKM = parseFloatOrDefault(GetSetting(KeyDeliveryFreeDistance), DefaultFreeDistanceKM)
	feePerKM = parseFloatOrDefault(GetSetting(KeyDeliveryFeePerKM), DefaultFeePerKM)
	if freeKM < 0 {
		freeKM = 0
	}
	if feePerKM < 0 {
		feePerKM = 0
	}
	return
}

func SaveDeliveryFeeSettings(freeKM, feePerKM float64) {
	if freeKM < 0 {
		freeKM = 0
	}
	if feePerKM < 0 {
		feePerKM = 0
	}
	UpsertSetting(KeyDeliveryFreeDistance, strconv.FormatFloat(freeKM, 'f', 2, 64))
	UpsertSetting(KeyDeliveryFeePerKM, strconv.FormatFloat(feePerKM, 'f', 2, 64))
}

func CalculateDeliveryFee(distanceKM, freeKM, feePerKM float64) float64 {
	if distanceKM < 0 {
		distanceKM = 0
	}
	if distanceKM <= freeKM {
		return 0
	}
	extra := distanceKM - freeKM
	// 向上取整到整公里
	extraUnits := int(extra)
	if float64(extraUnits) < extra {
		extraUnits++
	}
	fee := float64(extraUnits) * feePerKM
	// 保留2位小数
	fee, _ = strconv.ParseFloat(strconv.FormatFloat(fee, 'f', 2, 64), 64)
	return fee
}

// ─────────────────── Telegram Chat IDs ───────────────────

func GetRoleChatIDs(pickerFallback, deliveryFallback string) (picker, delivery string) {
	picker = GetSetting(KeyPickerChatID)
	if picker == "" {
		picker = pickerFallback
	}
	delivery = GetSetting(KeyDeliveryChatID)
	if delivery == "" {
		delivery = deliveryFallback
	}
	return
}

func SaveRoleChatIDs(picker, delivery string) {
	UpsertSetting(KeyPickerChatID, picker)
	UpsertSetting(KeyDeliveryChatID, delivery)
}

// ─────────────────── 联系方式 ───────────────────

type ContactInfo struct {
	Phone    string `json:"phone"`
	Telegram string `json:"telegram"`
	Whatsapp string `json:"whatsapp"`
}

func GetContactInfo() ContactInfo {
	return ContactInfo{
		Phone:    GetSetting(KeyContactPhone),
		Telegram: GetSetting(KeyContactTelegram),
		Whatsapp: GetSetting(KeyContactWhatsapp),
	}
}

func SaveContactInfo(phone, telegram, whatsapp string) ContactInfo {
	UpsertSetting(KeyContactPhone, phone)
	UpsertSetting(KeyContactTelegram, telegram)
	UpsertSetting(KeyContactWhatsapp, whatsapp)
	return GetContactInfo()
}

// ─────────────────── 统计 ───────────────────

// TrackPageView 记录页面访问（日期粒度去重）
func TrackPageView() {
	today := models.NowCambodia().Format("2006-01-02")
	database.DB.Exec(`
		INSERT INTO daily_metrics (date, page_views, updated_at)
		VALUES (?, 1, ?)
		ON CONFLICT(date) DO UPDATE SET page_views = page_views + 1, updated_at = ?
	`, today, models.NowCambodia(), models.NowCambodia())
}

// ─────────────────── helpers ───────────────────

func parseFloatOrDefault(s string, def float64) float64 {
	if s == "" {
		return def
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return f
}
