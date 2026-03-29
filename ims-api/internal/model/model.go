package model

import (
	"time"
)

// ==================== 基础模型 ====================

// BaseModel 公共字段
type BaseModel struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ==================== 用户权限 ====================

// User 用户
type User struct {
	BaseModel
	Username string `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password string `gorm:"size:128;not null" json:"-"`
	RealName string `gorm:"size:64" json:"real_name"`
	Email    string `gorm:"size:128" json:"email"`
	Phone    string `gorm:"size:20" json:"phone"`
	Role     string `gorm:"size:32;default:'operator'" json:"role"` // admin | operator | viewer
	Status   int    `gorm:"default:1" json:"status"`                // 1=启用 0=禁用
}

// ==================== 基础档案 ====================

// Category 商品分类
type Category struct {
	BaseModel
	Name     string `gorm:"size:64;not null" json:"name"`
	ParentID *uint  `json:"parent_id"`
	Remark   string `gorm:"size:256" json:"remark"`
}

// Product 商品
type Product struct {
	BaseModel
	Code       string   `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Name       string   `gorm:"size:128;not null" json:"name"`
	CategoryID uint     `json:"category_id"`
	Category   Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Unit       string   `gorm:"size:16;not null" json:"unit"` // 件/箱/kg
	CostPrice  float64  `gorm:"type:decimal(12,4);default:0" json:"cost_price"`
	SalePrice  float64  `gorm:"type:decimal(12,4);default:0" json:"sale_price"`
	MinStock   float64  `gorm:"type:decimal(12,4);default:0" json:"min_stock"` // 安全库存
	Status     int      `gorm:"default:1" json:"status"`                       // 1=启用 0=停用
	Remark     string   `gorm:"size:512" json:"remark"`
}

// Supplier 供应商
type Supplier struct {
	BaseModel
	Code    string `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Name    string `gorm:"size:128;not null" json:"name"`
	Contact string `gorm:"size:64" json:"contact"`
	Phone   string `gorm:"size:20" json:"phone"`
	Email   string `gorm:"size:128" json:"email"`
	Address string `gorm:"size:256" json:"address"`
	Status  int    `gorm:"default:1" json:"status"`
	Remark  string `gorm:"size:512" json:"remark"`
}

// Customer 客户
type Customer struct {
	BaseModel
	Code    string `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Name    string `gorm:"size:128;not null" json:"name"`
	Contact string `gorm:"size:64" json:"contact"`
	Phone   string `gorm:"size:20" json:"phone"`
	Email   string `gorm:"size:128" json:"email"`
	Address string `gorm:"size:256" json:"address"`
	Status  int    `gorm:"default:1" json:"status"`
	Remark  string `gorm:"size:512" json:"remark"`
}

// Warehouse 仓库
type Warehouse struct {
	BaseModel
	Code    string `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Name    string `gorm:"size:128;not null" json:"name"`
	Address string `gorm:"size:256" json:"address"`
	Keeper  string `gorm:"size:64" json:"keeper"` // 保管员
	Status  int    `gorm:"default:1" json:"status"`
	Remark  string `gorm:"size:512" json:"remark"`
}

// ==================== 采购模块 ====================

// PurchaseOrder 采购单
type PurchaseOrder struct {
	BaseModel
	OrderNo      string               `gorm:"uniqueIndex;size:64;not null" json:"order_no"`
	SupplierID   uint                 `json:"supplier_id"`
	Supplier     Supplier             `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	WarehouseID  uint                 `json:"warehouse_id"`
	Warehouse    Warehouse            `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Status       string               `gorm:"size:32;default:'draft'" json:"status"` // draft|approved|received|cancelled
	TotalAmount  float64              `gorm:"type:decimal(12,4);default:0" json:"total_amount"`
	Remark       string               `gorm:"size:512" json:"remark"`
	OperatorID   uint                 `json:"operator_id"`
	Operator     User                 `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
	ApprovedAt   *time.Time           `json:"approved_at"`
	ReceivedAt   *time.Time           `json:"received_at"`
	Items        []PurchaseOrderItem  `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

// PurchaseOrderItem 采购单明细
type PurchaseOrderItem struct {
	BaseModel
	OrderID   uint    `json:"order_id"`
	ProductID uint    `json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  float64 `gorm:"type:decimal(12,4);not null" json:"quantity"`
	Price     float64 `gorm:"type:decimal(12,4);not null" json:"price"`
	Amount    float64 `gorm:"type:decimal(12,4);not null" json:"amount"`
	Remark    string  `gorm:"size:256" json:"remark"`
}

// ==================== 销售模块 ====================

// SaleOrder 销售单
type SaleOrder struct {
	BaseModel
	OrderNo     string          `gorm:"uniqueIndex;size:64;not null" json:"order_no"`
	CustomerID  uint            `json:"customer_id"`
	Customer    Customer        `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	WarehouseID uint            `json:"warehouse_id"`
	Warehouse   Warehouse       `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Status      string          `gorm:"size:32;default:'draft'" json:"status"` // draft|approved|shipped|cancelled
	TotalAmount float64         `gorm:"type:decimal(12,4);default:0" json:"total_amount"`
	Remark      string          `gorm:"size:512" json:"remark"`
	OperatorID  uint            `json:"operator_id"`
	Operator    User            `gorm:"foreignKey:OperatorID" json:"operator,omitempty"`
	ApprovedAt  *time.Time      `json:"approved_at"`
	ShippedAt   *time.Time      `json:"shipped_at"`
	Items       []SaleOrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

// SaleOrderItem 销售单明细
type SaleOrderItem struct {
	BaseModel
	OrderID   uint    `json:"order_id"`
	ProductID uint    `json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  float64 `gorm:"type:decimal(12,4);not null" json:"quantity"`
	Price     float64 `gorm:"type:decimal(12,4);not null" json:"price"`
	Amount    float64 `gorm:"type:decimal(12,4);not null" json:"amount"`
	Remark    string  `gorm:"size:256" json:"remark"`
}

// ==================== 库存模块 ====================

// Inventory 库存台账
type Inventory struct {
	BaseModel
	ProductID   uint      `gorm:"uniqueIndex:idx_product_warehouse;not null" json:"product_id"`
	Product     Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	WarehouseID uint      `gorm:"uniqueIndex:idx_product_warehouse;not null" json:"warehouse_id"`
	Warehouse   Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Quantity    float64   `gorm:"type:decimal(12,4);default:0" json:"quantity"`
}

// TxType 库存流水类型
type TxType = string

const (
	TxTypePurchaseIn  TxType = "purchase_in"   // 采购入库
	TxTypeSaleOut     TxType = "sale_out"       // 销售出库
	TxTypeTransferIn  TxType = "transfer_in"    // 调拨入
	TxTypeTransferOut TxType = "transfer_out"   // 调拨出
	TxTypeStocktake   TxType = "stocktake_adj"  // 盘点调整
	TxTypeManualAdj   TxType = "manual_adj"     // 手动调整
)

// InventoryTransaction 库存流水
type InventoryTransaction struct {
	BaseModel
	ProductID   uint    `json:"product_id"`
	Product     Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	WarehouseID uint    `json:"warehouse_id"`
	Warehouse   Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	TxType      string  `gorm:"size:32;not null" json:"tx_type"`
	Quantity    float64 `gorm:"type:decimal(12,4);not null" json:"quantity"` // 正=入 负=出
	BeforeQty   float64 `gorm:"type:decimal(12,4)" json:"before_qty"`
	AfterQty    float64 `gorm:"type:decimal(12,4)" json:"after_qty"`
	RefID       uint    `json:"ref_id"`    // 关联单据ID
	RefType     string  `gorm:"size:32" json:"ref_type"` // purchase_order | sale_order | stocktake
	OperatorID  uint    `json:"operator_id"`
	Remark      string  `gorm:"size:256" json:"remark"`
}

// Stocktake 盘点单
type Stocktake struct {
	BaseModel
	OrderNo     string          `gorm:"uniqueIndex;size:64;not null" json:"order_no"`
	WarehouseID uint            `json:"warehouse_id"`
	Warehouse   Warehouse       `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	Status      string          `gorm:"size:32;default:'draft'" json:"status"` // draft|confirmed
	OperatorID  uint            `json:"operator_id"`
	Remark      string          `gorm:"size:512" json:"remark"`
	Items       []StocktakeItem `gorm:"foreignKey:StocktakeID" json:"items,omitempty"`
}

// StocktakeItem 盘点单明细
type StocktakeItem struct {
	BaseModel
	StocktakeID  uint    `json:"stocktake_id"`
	ProductID    uint    `json:"product_id"`
	Product      Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	SystemQty    float64 `gorm:"type:decimal(12,4)" json:"system_qty"`  // 系统数量
	ActualQty    float64 `gorm:"type:decimal(12,4)" json:"actual_qty"`  // 实盘数量
	DiffQty      float64 `gorm:"type:decimal(12,4)" json:"diff_qty"`    // 差异数量
}
