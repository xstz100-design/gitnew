package database

import (
	"log"
	"strings"
	"time"
	"wholesale/models"
	"wholesale/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect 初始化数据库连接并自动迁移表结构
func Connect(dsn string) {
	// 标准化 DSN: 兼容 Python 的 sqlite:///./xxx.db 格式
	dsn = normalizeDSN(dsn)

	gormCfg := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
		NowFunc: func() time.Time {
			return models.NowCambodia()
		},
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dsn+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)&_pragma=cache_size(-8000)"), gormCfg)
	if err != nil {
		log.Fatalf("[DB] 连接失败: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("[DB] 获取底层连接失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(5) // WAL 模式支持并发读，多连接可并行处理请求；写冲突由 busy_timeout 处理
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(3600 * time.Second)

	autoMigrate()
	log.Println("[DB] 数据库连接与迁移完成")
}

// autoMigrate 自动建表/补列
func autoMigrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.Product{},
		&models.Order{},
		&models.OrderItem{},
		&models.Category{},
		&models.Announcement{},
		&models.SystemSetting{},
		&models.DailyMetric{},
		&models.PhoneVerification{},
		&models.StockLedger{},
	)
	if err != nil {
		log.Fatalf("[DB] 自动迁移失败: %v", err)
	}

	// 创建额外索引（忽略已存在的索引错误）
	extraIndexes := []string{
		`CREATE INDEX IF NOT EXISTS ix_orders_merchant_created_at ON orders(merchant_id, created_at)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_orders_merchant_client_request_id ON orders(merchant_id, client_request_id)`,
	}
	for _, sql := range extraIndexes {
		if err := DB.Exec(sql).Error; err != nil {
			log.Printf("[DB] 创建索引警告: %v", err)
		}
	}

	// 确保超级管理员字段已存在（兼容旧数据库）
	ensureSuperAdmin()
	// 首次启动若无管理员账号则自动创建
	seedDefaultAdmin()
	// 首次启动若无仓库坐标则设置默认（金边中心）
	seedDefaultWarehouseCoords()
}

// ensureSuperAdmin 如果存在旧 admin 账号(100001)，设置 is_super_admin=1
func ensureSuperAdmin() {
	DB.Exec(`
		UPDATE users
		SET is_super_admin = 1
		WHERE role = 'admin'
		  AND username = '100001'
		  AND COALESCE(is_super_admin, 0) = 0
	`)
}

// normalizeDSN 将 Python 风格的 sqlite URL 转成文件路径
func normalizeDSN(dsn string) string {
	// sqlite:///./xxx.db → ./xxx.db
	if strings.HasPrefix(dsn, "sqlite:///") {
		return strings.TrimPrefix(dsn, "sqlite:///")
	}
	return dsn
}

// seedDefaultAdmin 首次启动时若无任何管理员账号则自动创建 admin/admin123
func seedDefaultAdmin() {
	var count int64
	DB.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return
	}
	hashed, err := utils.HashPassword("admin123")
	if err != nil {
		log.Printf("[DB] seedDefaultAdmin: 密码哈希失败: %v", err)
		return
	}
	admin := models.User{
		Username:       "admin",
		HashedPassword: hashed,
		FullName:       "Admin",
		Role:           "admin",
		IsActive:       true,
		IsSuperAdmin:   true,
	}
	if err := DB.Create(&admin).Error; err != nil {
		log.Printf("[DB] seedDefaultAdmin: 创建管理员失败: %v", err)
		return
	}
	log.Println("[DB] 已自动创建默认管理员账号: admin / admin123（请登录后立即修改密码）")
}

// seedDefaultWarehouseCoords 首次启动时若无仓库坐标则设置默认（金边市中心）
func seedDefaultWarehouseCoords() {
	// 直接查询，避免引入 services 包（避免循环导入）
	var count int64
	DB.Model(&models.SystemSetting{}).Where("key = ?", "delivery.warehouse_lat").Where("value != ?", "").Count(&count)
	if count > 0 {
		return
	}
	upsert := func(key, value string) {
		var s models.SystemSetting
		if DB.Where("key = ?", key).First(&s).Error == nil {
			DB.Model(&s).Updates(map[string]interface{}{"value": value, "updated_at": models.NowCambodia()})
		} else {
			DB.Create(&models.SystemSetting{Key: key, Value: value, UpdatedAt: models.NowCambodia()})
		}
	}
	upsert("delivery.warehouse_lat", "11.5564")
	upsert("delivery.warehouse_lng", "104.9282")
	log.Println("[DB] 已设置默认仓库坐标：金边市中心 (11.5564, 104.9282)，可在管理后台→设置→Google Maps中修改")
}
