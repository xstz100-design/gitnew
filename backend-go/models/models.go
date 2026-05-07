package models

import "time"

// 柬埔寨时区 UTC+7
func NowCambodia() time.Time {
	loc := time.FixedZone("Asia/Phnom_Penh", 7*60*60)
	return time.Now().In(loc)
}

// ──────────────────────────── Enums ────────────────────────────

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleMerchant UserRole = "merchant"
)

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
)

type PaymentStatus string

const (
	PaymentUnpaid  PaymentStatus = "unpaid"
	PaymentCash    PaymentStatus = "cash"
	PaymentMonthly PaymentStatus = "monthly"
)

type DeliveryStatus string

const (
	DeliveryPending    DeliveryStatus = "pending"
	DeliveryDelivering DeliveryStatus = "delivering"
	DeliveryDelivered  DeliveryStatus = "delivered"
	DeliveryCancelled  DeliveryStatus = "cancelled"
)

type AnnouncementType string

const (
	AnnouncementNotice  AnnouncementType = "notice"
	AnnouncementContact AnnouncementType = "contact"
	AnnouncementAbout   AnnouncementType = "about"
)

type BillStatus string

const (
	BillUnpaid  BillStatus = "unpaid"
	BillPaid    BillStatus = "paid"
	BillPartial BillStatus = "partial"
)

// ──────────────────────────── User ────────────────────────────

type User struct {
	ID                 int64          `gorm:"primarykey;autoIncrement" json:"id"`
	Username           string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
	HashedPassword     string         `gorm:"not null" json:"-"`
	FullName           string         `gorm:"size:100;not null" json:"full_name"`
	Role               UserRole       `gorm:"size:20;default:'merchant'" json:"role"`
	IsSuperAdmin       bool           `gorm:"default:false" json:"is_super_admin"`
	Phone              *string        `gorm:"size:20;index" json:"phone"`
	Address            *string        `gorm:"size:200" json:"address"`
	CreditLimit        float64        `gorm:"default:0" json:"credit_limit"`
	BillingCycleDays   *int           `json:"billing_cycle_days"`
	AllowCredit        bool           `gorm:"default:false" json:"allow_credit"`
	LocationURL        *string        `gorm:"size:500" json:"location_url"`
	StorePhoto         *string        `gorm:"size:500" json:"store_photo"`
	TelegramID         *int64         `gorm:"uniqueIndex" json:"telegram_id"`
	TelegramBotToken   *string        `gorm:"size:200" json:"-"`
	TelegramChatID     *string        `gorm:"size:100" json:"-"`
	NotifyEnabled      bool           `gorm:"default:true" json:"notify_enabled"`
	ApprovalStatus     ApprovalStatus `gorm:"size:20;default:'approved'" json:"approval_status"`
	RejectedReason     *string        `gorm:"size:200" json:"rejected_reason"`
	ApprovedAt         *time.Time     `json:"approved_at"`
	MustChangePassword bool           `gorm:"default:true" json:"must_change_password"`
	IsActive           bool           `gorm:"default:true" json:"is_active"`
	CreatedAt          time.Time      `json:"created_at"`

	Orders []Order `gorm:"foreignKey:MerchantID" json:"-"`
}

func (User) TableName() string { return "users" }

// ──────────────────────────── Product ────────────────────────────

type Product struct {
	ID                 int64      `gorm:"primarykey;autoIncrement" json:"id"`
	Name               string     `gorm:"size:100;index;not null" json:"name"`
	NameKh             *string    `gorm:"size:100;column:name_kh" json:"name_kh"`
	NameEn             *string    `gorm:"size:100;column:name_en" json:"name_en"`
	Brand              *string    `gorm:"size:100" json:"brand"`
	CountryOfOrigin    *string    `gorm:"size:100;column:country_of_origin" json:"country_of_origin"`
	Unit               string     `gorm:"size:20;default:'件'" json:"unit"`
	Specs              *string    `gorm:"size:100" json:"specs"`
	Barcode            *string    `gorm:"size:100" json:"barcode"`
	PriceUSD           float64    `gorm:"not null" json:"price_usd"`
	RetailPriceUSD     *float64   `gorm:"column:retail_price_usd" json:"retail_price_usd"`
	PricePerPieceUSD   *float64   `gorm:"column:price_per_piece_usd" json:"price_per_piece_usd"`
	PricePerPackageUSD *float64   `gorm:"column:price_per_package_usd" json:"price_per_package_usd"`
	PiecesPerPackage   *int       `gorm:"column:pieces_per_package" json:"pieces_per_package"`
	Stock              int        `gorm:"default:0" json:"stock"`
	StockWarning       int        `gorm:"default:10;column:stock_warning" json:"stock_warning"`
	Description        *string    `gorm:"size:500" json:"description"`
	ImageURL           *string    `gorm:"size:500;column:image_url" json:"image_url"`
	Img1               *string    `gorm:"size:500" json:"img1"`
	Img2               *string    `gorm:"size:500" json:"img2"`
	Img3               *string    `gorm:"size:500" json:"img3"`
	Img4               *string    `gorm:"size:500" json:"img4"`
	Img5               *string    `gorm:"size:500" json:"img5"`
	Category           *string    `gorm:"size:50" json:"category"`
	SortOrder          int        `gorm:"default:0;column:sort_order" json:"sort_order"`
	IsFeatured         bool       `gorm:"default:false;column:is_featured" json:"is_featured"`
	IsActive           bool       `gorm:"default:true;column:is_active" json:"is_active"`
	IsDeleted          bool       `gorm:"default:false;index;column:is_deleted" json:"is_deleted"`
	ProductionDate     *time.Time `gorm:"column:production_date" json:"production_date"`
	ExpiryDate         *time.Time `gorm:"column:expiry_date" json:"expiry_date"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (Product) TableName() string { return "products" }

func (p *Product) IsLowStock() bool {
	return p.Stock <= p.StockWarning
}

// ──────────────────────────── Order ────────────────────────────

type Order struct {
	ID               int64          `gorm:"primarykey;autoIncrement" json:"id"`
	OrderNo          string         `gorm:"uniqueIndex;size:50;column:order_no" json:"order_no"`
	MerchantID       int64          `gorm:"index;column:merchant_id" json:"merchant_id"`
	TotalUSD         float64        `gorm:"default:0;column:total_usd" json:"total_usd"`
	GoodsTotalUSD    float64        `gorm:"default:0;column:goods_total_usd" json:"goods_total_usd"`
	DeliveryFeeUSD   float64        `gorm:"default:0;column:delivery_fee_usd" json:"delivery_fee_usd"`
	DistanceKM       *float64       `gorm:"column:distance_km" json:"distance_km"`
	PaymentStatus    PaymentStatus  `gorm:"size:20;default:'unpaid';column:payment_status" json:"payment_status"`
	DeliveryStatus   DeliveryStatus `gorm:"size:20;default:'pending';column:delivery_status" json:"delivery_status"`
	DeliveryAddress  *string        `gorm:"size:200;column:delivery_address" json:"delivery_address"`
	DeliveryPhone    *string        `gorm:"size:20;column:delivery_phone" json:"delivery_phone"`
	DeliveryPersonID *int64         `gorm:"column:delivery_person_id" json:"delivery_person_id"`
	Note             *string        `gorm:"size:500" json:"note"`
	ClientRequestID  *string        `gorm:"size:64;index;column:client_request_id" json:"client_request_id"`
	IsDeleted        bool           `gorm:"default:false;index;column:is_deleted" json:"is_deleted"`
	ScheduledAt      *time.Time     `gorm:"index;column:scheduled_at" json:"scheduled_at"`
	PickedAt         *time.Time     `gorm:"index;column:picked_at" json:"picked_at"`
	PickedByID       *int64         `gorm:"column:picked_by_id" json:"picked_by_id"`
	CreatedAt        time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeliveredAt      *time.Time     `gorm:"column:delivered_at" json:"delivered_at"`

	Merchant *User       `gorm:"foreignKey:MerchantID" json:"merchant,omitempty"`
	Items    []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

func (Order) TableName() string { return "orders" }

func (o *Order) TotalKHR(rate float64) float64 {
	return o.TotalUSD * rate
}

// ──────────────────────────── OrderItem ────────────────────────────

type OrderItem struct {
	ID           int64     `gorm:"primarykey;autoIncrement" json:"id"`
	OrderID      int64     `gorm:"index;column:order_id" json:"order_id"`
	ProductID    int64     `gorm:"index;column:product_id" json:"product_id"`
	Quantity     int       `gorm:"default:1" json:"quantity"`
	UnitPriceUSD float64   `gorm:"default:0;column:unit_price_usd" json:"unit_price_usd"`
	SubtotalUSD  float64   `gorm:"default:0;column:subtotal_usd" json:"subtotal_usd"`
	PurchaseMode string    `gorm:"size:20;default:'default';column:purchase_mode" json:"purchase_mode"`
	CreatedAt    time.Time `json:"created_at"`

	Product *Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (OrderItem) TableName() string { return "order_items" }

// ──────────────────────────── Category ────────────────────────────

type Category struct {
	ID        int64     `gorm:"primarykey;autoIncrement" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:50;not null" json:"name"`
	SortOrder int       `gorm:"default:0;column:sort_order" json:"sort_order"`
	IsActive  bool      `gorm:"default:true;column:is_active" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

func (Category) TableName() string { return "categories" }

// ──────────────────────────── Announcement ────────────────────────────

type Announcement struct {
	ID        int64            `gorm:"primarykey;autoIncrement" json:"id"`
	Type      AnnouncementType `gorm:"size:20;default:'notice'" json:"type"`
	ContentZh string           `gorm:"size:2000;column:content_zh;not null" json:"content_zh"`
	ContentEn string           `gorm:"size:2000;column:content_en;default:''" json:"content_en"`
	IsActive  bool             `gorm:"default:true;column:is_active" json:"is_active"`
	SortOrder int              `gorm:"default:0;column:sort_order" json:"sort_order"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

func (Announcement) TableName() string { return "announcements" }

// ──────────────────────────── MonthlyBill ────────────────────────────

type MonthlyBill struct {
	ID          int64      `gorm:"primarykey;autoIncrement" json:"id"`
	MerchantID  int64      `gorm:"index;column:merchant_id" json:"merchant_id"`
	Year        *int       `json:"year"`
	Month       *int       `json:"month"`
	PeriodStart *time.Time `gorm:"column:period_start" json:"period_start"`
	PeriodEnd   *time.Time `gorm:"column:period_end" json:"period_end"`
	TotalAmount float64    `gorm:"default:0;column:total_amount" json:"total_amount"`
	PaidAmount  float64    `gorm:"default:0;column:paid_amount" json:"paid_amount"`
	Status      BillStatus `gorm:"size:20;default:'unpaid'" json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (MonthlyBill) TableName() string { return "monthly_bills" }

// ──────────────────────────── SystemSetting ────────────────────────────

type SystemSetting struct {
	ID        int64     `gorm:"primarykey;autoIncrement" json:"id"`
	Key       string    `gorm:"uniqueIndex;size:100;not null" json:"key"`
	Value     string    `gorm:"size:500;default:''" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (SystemSetting) TableName() string { return "system_settings" }

// ──────────────────────────── DailyMetric ────────────────────────────

type DailyMetric struct {
	ID        int64     `gorm:"primarykey;autoIncrement" json:"id"`
	Date      string    `gorm:"uniqueIndex;size:10;not null" json:"date"`
	PageViews int       `gorm:"default:0;column:page_views" json:"page_views"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (DailyMetric) TableName() string { return "daily_metrics" }

// ──────────────────────────── PhoneVerification ────────────────────────────

type PhoneVerification struct {
	ID        int64     `gorm:"primarykey;autoIncrement" json:"id"`
	Phone     string    `gorm:"size:20;index;not null" json:"phone"`
	Code      string    `gorm:"size:10;not null" json:"code"`
	ExpiresAt time.Time `gorm:"column:expires_at" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

func (PhoneVerification) TableName() string { return "phone_verifications" }
