package handlers

import (
	"net/http"
	"strconv"
	"time"
	"wholesale/database"
	"wholesale/middleware"
	"wholesale/models"

	"github.com/gin-gonic/gin"
)

// GET /api/billing
func ListBills(c *gin.Context) {
	user := middleware.CurrentUser(c)

	q := database.DB.Model(&models.MonthlyBill{})
	if user.Role == models.RoleMerchant {
		q = q.Where("merchant_id = ?", user.ID)
	} else {
		if mid := c.Query("merchant_id"); mid != "" {
			q = q.Where("merchant_id = ?", mid)
		}
	}
	if year := c.Query("year"); year != "" {
		q = q.Where("year = ?", year)
	}
	if month := c.Query("month"); month != "" {
		q = q.Where("month = ?", month)
	}

	var bills []models.MonthlyBill
	q.Order("year DESC, month DESC").Find(&bills)
	c.JSON(http.StatusOK, bills)
}

// POST /api/billing/generate — 管理员生成账单
func GenerateBills(c *gin.Context) {
	loc := time.FixedZone("Asia/Phnom_Penh", 7*60*60)
	now := time.Now().In(loc)
	year := now.Year()
	month := int(now.Month())

	if y := c.Query("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			year = v
		}
	}
	if m := c.Query("month"); m != "" {
		if v, err := strconv.Atoi(m); err == nil {
			month = v
		}
	}

	// 取所有 allow_credit=true 的商户
	var merchants []models.User
	database.DB.Where("role = ? AND allow_credit = ?", models.RoleMerchant, true).Find(&merchants)

	var generated, skipped int
	for _, merchant := range merchants {
		// 避免重复生成
		var existing models.MonthlyBill
		if database.DB.Where("merchant_id = ? AND year = ? AND month = ?",
			merchant.ID, year, month).First(&existing).Error == nil {
			skipped++
			continue
		}

		// 统计该月未结算订单金额
		startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
		periodEnd := startDate.AddDate(0, 1, 0)

		var totalAmount float64
		database.DB.Model(&models.Order{}).
			Where("merchant_id = ? AND payment_status = ? AND is_deleted = ? AND created_at >= ? AND created_at < ?",
				merchant.ID, models.PaymentMonthly, false, startDate, periodEnd).
			Select("COALESCE(SUM(total_usd), 0)").
			Scan(&totalAmount)

		bill := models.MonthlyBill{
			MerchantID:  merchant.ID,
			Year:        &year,
			Month:       &month,
			PeriodStart: &startDate,
			PeriodEnd:   &periodEnd,
			TotalAmount: totalAmount,
			PaidAmount:  0,
			Status:      models.BillUnpaid,
			CreatedAt:   models.NowCambodia(),
			UpdatedAt:   models.NowCambodia(),
		}
		database.DB.Create(&bill)
		generated++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "账单生成完成",
		"generated": generated,
		"skipped":   skipped,
	})
}

// PATCH /api/billing/:id
func UpdateBill(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var bill models.MonthlyBill
	if err := database.DB.First(&bill, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "账单不存在"})
		return
	}

	var req struct {
		Status     *string  `json:"status"`
		PaidAmount *float64 `json:"paid_amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	updates := map[string]interface{}{"updated_at": models.NowCambodia()}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.PaidAmount != nil {
		updates["paid_amount"] = *req.PaidAmount
	}

	database.DB.Model(&bill).Updates(updates)
	database.DB.First(&bill, id)
	c.JSON(http.StatusOK, bill)
}
