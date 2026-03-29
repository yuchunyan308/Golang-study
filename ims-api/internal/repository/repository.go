package repository

import (
	"gorm.io/gorm"

	"ims-api/internal/dto"
	"ims-api/internal/model"
	"ims-api/pkg/pagination"
)

// ==================== User Repository ====================

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db} }

func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	var u model.User
	return &u, r.db.Where("username = ?", username).First(&u).Error
}

func (r *UserRepo) FindByID(id uint) (*model.User, error) {
	var u model.User
	return &u, r.db.First(&u, id).Error
}

func (r *UserRepo) Create(u *model.User) error { return r.db.Create(u).Error }

func (r *UserRepo) Update(u *model.User) error { return r.db.Save(u).Error }

func (r *UserRepo) List(page *pagination.Pagination) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	q := r.db.Model(&model.User{})
	q.Count(&total)
	err := q.Offset(page.Offset()).Limit(page.PageSize).Find(&users).Error
	return users, total, err
}

// ==================== Category Repository ====================

type CategoryRepo struct{ db *gorm.DB }

func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db} }

func (r *CategoryRepo) Create(c *model.Category) error { return r.db.Create(c).Error }
func (r *CategoryRepo) FindByID(id uint) (*model.Category, error) {
	var c model.Category
	return &c, r.db.First(&c, id).Error
}
func (r *CategoryRepo) List() ([]model.Category, error) {
	var list []model.Category
	return list, r.db.Find(&list).Error
}
func (r *CategoryRepo) Update(c *model.Category) error { return r.db.Save(c).Error }
func (r *CategoryRepo) Delete(id uint) error           { return r.db.Delete(&model.Category{}, id).Error }

// ==================== Product Repository ====================

type ProductRepo struct{ db *gorm.DB }

func NewProductRepo(db *gorm.DB) *ProductRepo { return &ProductRepo{db} }

func (r *ProductRepo) Create(p *model.Product) error { return r.db.Create(p).Error }

func (r *ProductRepo) FindByID(id uint) (*model.Product, error) {
	var p model.Product
	return &p, r.db.Preload("Category").First(&p, id).Error
}

func (r *ProductRepo) FindByCode(code string) (*model.Product, error) {
	var p model.Product
	return &p, r.db.Where("code = ?", code).First(&p).Error
}

func (r *ProductRepo) Update(p *model.Product) error { return r.db.Save(p).Error }

func (r *ProductRepo) List(q *dto.ProductQuery, page *pagination.Pagination) ([]model.Product, int64, error) {
	var list []model.Product
	var total int64
	db := r.db.Model(&model.Product{}).Preload("Category")
	if q.Name != "" {
		db = db.Where("name LIKE ?", "%"+q.Name+"%")
	}
	if q.CategoryID > 0 {
		db = db.Where("category_id = ?", q.CategoryID)
	}
	if q.Status != nil {
		db = db.Where("status = ?", *q.Status)
	}
	db.Count(&total)
	err := db.Offset(page.Offset()).Limit(page.PageSize).Find(&list).Error
	return list, total, err
}

// ==================== Supplier Repository ====================

type SupplierRepo struct{ db *gorm.DB }

func NewSupplierRepo(db *gorm.DB) *SupplierRepo { return &SupplierRepo{db} }

func (r *SupplierRepo) Create(s *model.Supplier) error { return r.db.Create(s).Error }
func (r *SupplierRepo) FindByID(id uint) (*model.Supplier, error) {
	var s model.Supplier
	return &s, r.db.First(&s, id).Error
}
func (r *SupplierRepo) Update(s *model.Supplier) error { return r.db.Save(s).Error }
func (r *SupplierRepo) List(name string, page *pagination.Pagination) ([]model.Supplier, int64, error) {
	var list []model.Supplier
	var total int64
	db := r.db.Model(&model.Supplier{})
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	db.Count(&total)
	err := db.Offset(page.Offset()).Limit(page.PageSize).Find(&list).Error
	return list, total, err
}

// ==================== Customer Repository ====================

type CustomerRepo struct{ db *gorm.DB }

func NewCustomerRepo(db *gorm.DB) *CustomerRepo { return &CustomerRepo{db} }

func (r *CustomerRepo) Create(c *model.Customer) error { return r.db.Create(c).Error }
func (r *CustomerRepo) FindByID(id uint) (*model.Customer, error) {
	var c model.Customer
	return &c, r.db.First(&c, id).Error
}
func (r *CustomerRepo) Update(c *model.Customer) error { return r.db.Save(c).Error }
func (r *CustomerRepo) List(name string, page *pagination.Pagination) ([]model.Customer, int64, error) {
	var list []model.Customer
	var total int64
	db := r.db.Model(&model.Customer{})
	if name != "" {
		db = db.Where("name LIKE ?", "%"+name+"%")
	}
	db.Count(&total)
	err := db.Offset(page.Offset()).Limit(page.PageSize).Find(&list).Error
	return list, total, err
}

// ==================== Warehouse Repository ====================

type WarehouseRepo struct{ db *gorm.DB }

func NewWarehouseRepo(db *gorm.DB) *WarehouseRepo { return &WarehouseRepo{db} }

func (r *WarehouseRepo) Create(w *model.Warehouse) error { return r.db.Create(w).Error }
func (r *WarehouseRepo) FindByID(id uint) (*model.Warehouse, error) {
	var w model.Warehouse
	return &w, r.db.First(&w, id).Error
}
func (r *WarehouseRepo) Update(w *model.Warehouse) error { return r.db.Save(w).Error }
func (r *WarehouseRepo) List() ([]model.Warehouse, error) {
	var list []model.Warehouse
	return list, r.db.Find(&list).Error
}

// ==================== Purchase Repository ====================

type PurchaseRepo struct{ db *gorm.DB }

func NewPurchaseRepo(db *gorm.DB) *PurchaseRepo { return &PurchaseRepo{db} }

func (r *PurchaseRepo) Create(tx *gorm.DB, o *model.PurchaseOrder) error {
	return tx.Create(o).Error
}

func (r *PurchaseRepo) FindByID(id uint) (*model.PurchaseOrder, error) {
	var o model.PurchaseOrder
	err := r.db.Preload("Supplier").Preload("Warehouse").Preload("Operator").
		Preload("Items.Product").First(&o, id).Error
	return &o, err
}

func (r *PurchaseRepo) Update(tx *gorm.DB, o *model.PurchaseOrder) error {
	return tx.Save(o).Error
}

func (r *PurchaseRepo) DeleteItems(tx *gorm.DB, orderID uint) error {
	return tx.Where("order_id = ?", orderID).Delete(&model.PurchaseOrderItem{}).Error
}

func (r *PurchaseRepo) List(q *dto.PurchaseQuery, page *pagination.Pagination) ([]model.PurchaseOrder, int64, error) {
	var list []model.PurchaseOrder
	var total int64
	db := r.db.Model(&model.PurchaseOrder{}).Preload("Supplier").Preload("Warehouse")
	if q.OrderNo != "" {
		db = db.Where("order_no LIKE ?", "%"+q.OrderNo+"%")
	}
	if q.SupplierID > 0 {
		db = db.Where("supplier_id = ?", q.SupplierID)
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if q.StartDate != "" {
		db = db.Where("created_at >= ?", q.StartDate)
	}
	if q.EndDate != "" {
		db = db.Where("created_at <= ?", q.EndDate+" 23:59:59")
	}
	db.Count(&total)
	err := db.Order("created_at DESC").Offset(page.Offset()).Limit(page.PageSize).Find(&list).Error
	return list, total, err
}

// ==================== Sale Repository ====================

type SaleRepo struct{ db *gorm.DB }

func NewSaleRepo(db *gorm.DB) *SaleRepo { return &SaleRepo{db} }

func (r *SaleRepo) Create(tx *gorm.DB, o *model.SaleOrder) error { return tx.Create(o).Error }

func (r *SaleRepo) FindByID(id uint) (*model.SaleOrder, error) {
	var o model.SaleOrder
	err := r.db.Preload("Customer").Preload("Warehouse").Preload("Operator").
		Preload("Items.Product").First(&o, id).Error
	return &o, err
}

func (r *SaleRepo) Update(tx *gorm.DB, o *model.SaleOrder) error { return tx.Save(o).Error }

func (r *SaleRepo) DeleteItems(tx *gorm.DB, orderID uint) error {
	return tx.Where("order_id = ?", orderID).Delete(&model.SaleOrderItem{}).Error
}

func (r *SaleRepo) List(q *dto.SaleQuery, page *pagination.Pagination) ([]model.SaleOrder, int64, error) {
	var list []model.SaleOrder
	var total int64
	db := r.db.Model(&model.SaleOrder{}).Preload("Customer").Preload("Warehouse")
	if q.OrderNo != "" {
		db = db.Where("order_no LIKE ?", "%"+q.OrderNo+"%")
	}
	if q.CustomerID > 0 {
		db = db.Where("customer_id = ?", q.CustomerID)
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if q.StartDate != "" {
		db = db.Where("created_at >= ?", q.StartDate)
	}
	if q.EndDate != "" {
		db = db.Where("created_at <= ?", q.EndDate+" 23:59:59")
	}
	db.Count(&total)
	err := db.Order("created_at DESC").Offset(page.Offset()).Limit(page.PageSize).Find(&list).Error
	return list, total, err
}

// ==================== Inventory Repository ====================

type InventoryRepo struct{ db *gorm.DB }

func NewInventoryRepo(db *gorm.DB) *InventoryRepo { return &InventoryRepo{db} }

// GetOrCreate 获取或创建库存记录
func (r *InventoryRepo) GetOrCreate(tx *gorm.DB, productID, warehouseID uint) (*model.Inventory, error) {
	var inv model.Inventory
	err := tx.Where("product_id = ? AND warehouse_id = ?", productID, warehouseID).
		FirstOrCreate(&inv, model.Inventory{
			ProductID:   productID,
			WarehouseID: warehouseID,
			Quantity:    0,
		}).Error
	return &inv, err
}

// UpdateQuantity 更新库存数量
func (r *InventoryRepo) UpdateQuantity(tx *gorm.DB, id uint, delta float64) error {
	return tx.Model(&model.Inventory{}).Where("id = ?", id).
		UpdateColumn("quantity", gorm.Expr("quantity + ?", delta)).Error
}

// AddTransaction 添加流水记录
func (r *InventoryRepo) AddTransaction(tx *gorm.DB, t *model.InventoryTransaction) error {
	return tx.Create(t).Error
}

func (r *InventoryRepo) List(q *dto.InventoryQuery, page *pagination.Pagination) ([]model.Inventory, int64, error) {
	var list []model.Inventory
	var total int64
	db := r.db.Model(&model.Inventory{}).Preload("Product.Category").Preload("Warehouse")
	if q.ProductID > 0 {
		db = db.Where("product_id = ?", q.ProductID)
	}
	if q.WarehouseID > 0 {
		db = db.Where("warehouse_id = ?", q.WarehouseID)
	}
	if q.LowStock {
		db = db.Joins("JOIN products ON products.id = inventories.product_id").
			Where("inventories.quantity < products.min_stock")
	}
	db.Count(&total)
	err := db.Offset(page.Offset()).Limit(page.PageSize).Find(&list).Error
	return list, total, err
}

func (r *InventoryRepo) ListTx(q *dto.TxQuery, page *pagination.Pagination) ([]model.InventoryTransaction, int64, error) {
	var list []model.InventoryTransaction
	var total int64
	db := r.db.Model(&model.InventoryTransaction{}).Preload("Product").Preload("Warehouse")
	if q.ProductID > 0 {
		db = db.Where("product_id = ?", q.ProductID)
	}
	if q.WarehouseID > 0 {
		db = db.Where("warehouse_id = ?", q.WarehouseID)
	}
	if q.TxType != "" {
		db = db.Where("tx_type = ?", q.TxType)
	}
	if q.StartDate != "" {
		db = db.Where("created_at >= ?", q.StartDate)
	}
	if q.EndDate != "" {
		db = db.Where("created_at <= ?", q.EndDate+" 23:59:59")
	}
	db.Count(&total)
	err := db.Order("created_at DESC").Offset(page.Offset()).Limit(page.PageSize).Find(&list).Error
	return list, total, err
}

// ==================== Stocktake Repository ====================

type StocktakeRepo struct{ db *gorm.DB }

func NewStocktakeRepo(db *gorm.DB) *StocktakeRepo { return &StocktakeRepo{db} }

func (r *StocktakeRepo) Create(tx *gorm.DB, s *model.Stocktake) error { return tx.Create(s).Error }

func (r *StocktakeRepo) FindByID(id uint) (*model.Stocktake, error) {
	var s model.Stocktake
	err := r.db.Preload("Warehouse").Preload("Items.Product").First(&s, id).Error
	return &s, err
}

func (r *StocktakeRepo) Update(tx *gorm.DB, s *model.Stocktake) error { return tx.Save(s).Error }

func (r *StocktakeRepo) List(page *pagination.Pagination) ([]model.Stocktake, int64, error) {
	var list []model.Stocktake
	var total int64
	r.db.Model(&model.Stocktake{}).Count(&total)
	err := r.db.Preload("Warehouse").Order("created_at DESC").
		Offset(page.Offset()).Limit(page.PageSize).Find(&list).Error
	return list, total, err
}
