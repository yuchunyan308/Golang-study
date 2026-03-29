package dto

// ==================== Auth ====================

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	Token    string `json:"token"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Role     string `json:"role"`
}

// ==================== User ====================

type CreateUserReq struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6"`
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role" binding:"required,oneof=admin operator viewer"`
}

type UpdateUserReq struct {
	RealName string `json:"real_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role" binding:"omitempty,oneof=admin operator viewer"`
	Status   *int   `json:"status" binding:"omitempty,oneof=0 1"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ==================== Category ====================

type CreateCategoryReq struct {
	Name     string `json:"name" binding:"required,max=64"`
	ParentID *uint  `json:"parent_id"`
	Remark   string `json:"remark"`
}

// ==================== Product ====================

type CreateProductReq struct {
	Code       string  `json:"code" binding:"required,max=64"`
	Name       string  `json:"name" binding:"required,max=128"`
	CategoryID uint    `json:"category_id" binding:"required"`
	Unit       string  `json:"unit" binding:"required,max=16"`
	CostPrice  float64 `json:"cost_price" binding:"gte=0"`
	SalePrice  float64 `json:"sale_price" binding:"gte=0"`
	MinStock   float64 `json:"min_stock" binding:"gte=0"`
	Remark     string  `json:"remark"`
}

type UpdateProductReq struct {
	Name       string   `json:"name" binding:"omitempty,max=128"`
	CategoryID uint     `json:"category_id"`
	Unit       string   `json:"unit" binding:"omitempty,max=16"`
	CostPrice  *float64 `json:"cost_price" binding:"omitempty,gte=0"`
	SalePrice  *float64 `json:"sale_price" binding:"omitempty,gte=0"`
	MinStock   *float64 `json:"min_stock" binding:"omitempty,gte=0"`
	Status     *int     `json:"status" binding:"omitempty,oneof=0 1"`
	Remark     string   `json:"remark"`
}

// ==================== Supplier ====================

type CreateSupplierReq struct {
	Code    string `json:"code" binding:"required,max=64"`
	Name    string `json:"name" binding:"required,max=128"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
	Remark  string `json:"remark"`
}

type UpdateSupplierReq struct {
	Name    string `json:"name" binding:"omitempty,max=128"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
	Status  *int   `json:"status" binding:"omitempty,oneof=0 1"`
	Remark  string `json:"remark"`
}

// ==================== Customer ====================

type CreateCustomerReq struct {
	Code    string `json:"code" binding:"required,max=64"`
	Name    string `json:"name" binding:"required,max=128"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
	Remark  string `json:"remark"`
}

type UpdateCustomerReq struct {
	Name    string `json:"name" binding:"omitempty,max=128"`
	Contact string `json:"contact"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
	Status  *int   `json:"status" binding:"omitempty,oneof=0 1"`
	Remark  string `json:"remark"`
}

// ==================== Warehouse ====================

type CreateWarehouseReq struct {
	Code    string `json:"code" binding:"required,max=64"`
	Name    string `json:"name" binding:"required,max=128"`
	Address string `json:"address"`
	Keeper  string `json:"keeper"`
	Remark  string `json:"remark"`
}

type UpdateWarehouseReq struct {
	Name    string `json:"name" binding:"omitempty,max=128"`
	Address string `json:"address"`
	Keeper  string `json:"keeper"`
	Status  *int   `json:"status" binding:"omitempty,oneof=0 1"`
	Remark  string `json:"remark"`
}

// ==================== Purchase ====================

type PurchaseItemReq struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price" binding:"required,gt=0"`
	Remark    string  `json:"remark"`
}

type CreatePurchaseReq struct {
	SupplierID  uint              `json:"supplier_id" binding:"required"`
	WarehouseID uint              `json:"warehouse_id" binding:"required"`
	Remark      string            `json:"remark"`
	Items       []PurchaseItemReq `json:"items" binding:"required,min=1,dive"`
}

type UpdatePurchaseReq struct {
	SupplierID  uint              `json:"supplier_id"`
	WarehouseID uint              `json:"warehouse_id"`
	Remark      string            `json:"remark"`
	Items       []PurchaseItemReq `json:"items" binding:"omitempty,min=1,dive"`
}

// ==================== Sale ====================

type SaleItemReq struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  float64 `json:"quantity" binding:"required,gt=0"`
	Price     float64 `json:"price" binding:"required,gt=0"`
	Remark    string  `json:"remark"`
}

type CreateSaleReq struct {
	CustomerID  uint          `json:"customer_id" binding:"required"`
	WarehouseID uint          `json:"warehouse_id" binding:"required"`
	Remark      string        `json:"remark"`
	Items       []SaleItemReq `json:"items" binding:"required,min=1,dive"`
}

type UpdateSaleReq struct {
	CustomerID  uint          `json:"customer_id"`
	WarehouseID uint          `json:"warehouse_id"`
	Remark      string        `json:"remark"`
	Items       []SaleItemReq `json:"items" binding:"omitempty,min=1,dive"`
}

// ==================== Inventory ====================

type TransferReq struct {
	ProductID      uint    `json:"product_id" binding:"required"`
	FromWarehouseID uint   `json:"from_warehouse_id" binding:"required"`
	ToWarehouseID  uint    `json:"to_warehouse_id" binding:"required"`
	Quantity       float64 `json:"quantity" binding:"required,gt=0"`
	Remark         string  `json:"remark"`
}

type ManualAdjReq struct {
	ProductID   uint    `json:"product_id" binding:"required"`
	WarehouseID uint    `json:"warehouse_id" binding:"required"`
	Quantity    float64 `json:"quantity" binding:"required"` // 可正可负
	Remark      string  `json:"remark"`
}

// ==================== Stocktake ====================

type StocktakeItemReq struct {
	ProductID uint    `json:"product_id" binding:"required"`
	ActualQty float64 `json:"actual_qty" binding:"gte=0"`
}

type CreateStocktakeReq struct {
	WarehouseID uint               `json:"warehouse_id" binding:"required"`
	Remark      string             `json:"remark"`
	Items       []StocktakeItemReq `json:"items" binding:"required,min=1,dive"`
}

// ==================== Query Filters ====================

type ProductQuery struct {
	Name       string `form:"name"`
	CategoryID uint   `form:"category_id"`
	Status     *int   `form:"status"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

type PurchaseQuery struct {
	OrderNo    string `form:"order_no"`
	SupplierID uint   `form:"supplier_id"`
	Status     string `form:"status"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

type SaleQuery struct {
	OrderNo    string `form:"order_no"`
	CustomerID uint   `form:"customer_id"`
	Status     string `form:"status"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

type InventoryQuery struct {
	ProductID   uint   `form:"product_id"`
	WarehouseID uint   `form:"warehouse_id"`
	LowStock    bool   `form:"low_stock"` // 只看低于安全库存的
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
}

type TxQuery struct {
	ProductID   uint   `form:"product_id"`
	WarehouseID uint   `form:"warehouse_id"`
	TxType      string `form:"tx_type"`
	StartDate   string `form:"start_date"`
	EndDate     string `form:"end_date"`
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
}
