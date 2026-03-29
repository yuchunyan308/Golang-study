package service

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"ims-api/internal/dto"
	"ims-api/internal/middleware"
	"ims-api/internal/model"
	"ims-api/internal/repository"
	bizErr "ims-api/pkg/errors"
	"ims-api/pkg/pagination"
	"ims-api/pkg/util"
)

// ==================== Auth Service ====================

type AuthService struct {
	userRepo    *repository.UserRepo
	jwtSecret   string
	expireHours int
}

func NewAuthService(userRepo *repository.UserRepo, secret string, expire int) *AuthService {
	return &AuthService{userRepo, secret, expire}
}

func (s *AuthService) Login(req *dto.LoginReq) (*dto.LoginResp, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizErr.ErrUserNotFound
		}
		return nil, bizErr.ErrServer
	}
	if user.Status == 0 {
		return nil, bizErr.New(20005, "账号已禁用")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, bizErr.ErrWrongPassword
	}
	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role, s.expireHours)
	if err != nil {
		return nil, bizErr.ErrServer
	}
	return &dto.LoginResp{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
		RealName: user.RealName,
		Role:     user.Role,
	}, nil
}

// ==================== User Service ====================

type UserService struct{ repo *repository.UserRepo }

func NewUserService(repo *repository.UserRepo) *UserService { return &UserService{repo} }

func (s *UserService) Create(req *dto.CreateUserReq) error {
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	u := &model.User{
		Username: req.Username,
		Password: string(hash),
		RealName: req.RealName,
		Email:    req.Email,
		Phone:    req.Phone,
		Role:     req.Role,
		Status:   1,
	}
	if err := s.repo.Create(u); err != nil {
		return bizErr.ErrDuplicate
	}
	return nil
}

func (s *UserService) Update(id uint, req *dto.UpdateUserReq) error {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrUserNotFound
	}
	if req.RealName != "" {
		u.RealName = req.RealName
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.Phone != "" {
		u.Phone = req.Phone
	}
	if req.Role != "" {
		u.Role = req.Role
	}
	if req.Status != nil {
		u.Status = *req.Status
	}
	return s.repo.Update(u)
}

func (s *UserService) ChangePassword(id uint, req *dto.ChangePasswordReq) error {
	u, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrUserNotFound
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.OldPassword)); err != nil {
		return bizErr.ErrWrongPassword
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	u.Password = string(hash)
	return s.repo.Update(u)
}

func (s *UserService) List(page *pagination.Pagination) ([]model.User, int64, error) {
	return s.repo.List(page)
}

// ==================== Category Service ====================

type CategoryService struct{ repo *repository.CategoryRepo }

func NewCategoryService(repo *repository.CategoryRepo) *CategoryService { return &CategoryService{repo} }

func (s *CategoryService) Create(req *dto.CreateCategoryReq) error {
	return s.repo.Create(&model.Category{Name: req.Name, ParentID: req.ParentID, Remark: req.Remark})
}

func (s *CategoryService) List() ([]model.Category, error) { return s.repo.List() }

func (s *CategoryService) Delete(id uint) error { return s.repo.Delete(id) }

// ==================== Product Service ====================

type ProductService struct{ repo *repository.ProductRepo }

func NewProductService(repo *repository.ProductRepo) *ProductService { return &ProductService{repo} }

func (s *ProductService) Create(req *dto.CreateProductReq) error {
	return s.repo.Create(&model.Product{
		Code: req.Code, Name: req.Name, CategoryID: req.CategoryID,
		Unit: req.Unit, CostPrice: req.CostPrice, SalePrice: req.SalePrice,
		MinStock: req.MinStock, Remark: req.Remark, Status: 1,
	})
}

func (s *ProductService) GetByID(id uint) (*model.Product, error) {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return nil, bizErr.ErrProductNotFound
	}
	return p, nil
}

func (s *ProductService) Update(id uint, req *dto.UpdateProductReq) error {
	p, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrProductNotFound
	}
	if req.Name != "" {
		p.Name = req.Name
	}
	if req.CategoryID > 0 {
		p.CategoryID = req.CategoryID
	}
	if req.Unit != "" {
		p.Unit = req.Unit
	}
	if req.CostPrice != nil {
		p.CostPrice = *req.CostPrice
	}
	if req.SalePrice != nil {
		p.SalePrice = *req.SalePrice
	}
	if req.MinStock != nil {
		p.MinStock = *req.MinStock
	}
	if req.Status != nil {
		p.Status = *req.Status
	}
	if req.Remark != "" {
		p.Remark = req.Remark
	}
	return s.repo.Update(p)
}

func (s *ProductService) List(q *dto.ProductQuery, page *pagination.Pagination) ([]model.Product, int64, error) {
	return s.repo.List(q, page)
}

// ==================== Supplier Service ====================

type SupplierService struct{ repo *repository.SupplierRepo }

func NewSupplierService(repo *repository.SupplierRepo) *SupplierService {
	return &SupplierService{repo}
}

func (s *SupplierService) Create(req *dto.CreateSupplierReq) error {
	return s.repo.Create(&model.Supplier{
		Code: req.Code, Name: req.Name, Contact: req.Contact,
		Phone: req.Phone, Email: req.Email, Address: req.Address,
		Remark: req.Remark, Status: 1,
	})
}

func (s *SupplierService) GetByID(id uint) (*model.Supplier, error) {
	sup, err := s.repo.FindByID(id)
	if err != nil {
		return nil, bizErr.ErrNotFound
	}
	return sup, nil
}

func (s *SupplierService) Update(id uint, req *dto.UpdateSupplierReq) error {
	sup, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrNotFound
	}
	if req.Name != "" {
		sup.Name = req.Name
	}
	sup.Contact = req.Contact
	sup.Phone = req.Phone
	sup.Email = req.Email
	sup.Address = req.Address
	sup.Remark = req.Remark
	if req.Status != nil {
		sup.Status = *req.Status
	}
	return s.repo.Update(sup)
}

func (s *SupplierService) List(name string, page *pagination.Pagination) ([]model.Supplier, int64, error) {
	return s.repo.List(name, page)
}

// ==================== Customer Service ====================

type CustomerService struct{ repo *repository.CustomerRepo }

func NewCustomerService(repo *repository.CustomerRepo) *CustomerService {
	return &CustomerService{repo}
}

func (s *CustomerService) Create(req *dto.CreateCustomerReq) error {
	return s.repo.Create(&model.Customer{
		Code: req.Code, Name: req.Name, Contact: req.Contact,
		Phone: req.Phone, Email: req.Email, Address: req.Address,
		Remark: req.Remark, Status: 1,
	})
}

func (s *CustomerService) GetByID(id uint) (*model.Customer, error) {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return nil, bizErr.ErrNotFound
	}
	return c, nil
}

func (s *CustomerService) Update(id uint, req *dto.UpdateCustomerReq) error {
	c, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrNotFound
	}
	if req.Name != "" {
		c.Name = req.Name
	}
	c.Contact = req.Contact
	c.Phone = req.Phone
	c.Email = req.Email
	c.Address = req.Address
	c.Remark = req.Remark
	if req.Status != nil {
		c.Status = *req.Status
	}
	return s.repo.Update(c)
}

func (s *CustomerService) List(name string, page *pagination.Pagination) ([]model.Customer, int64, error) {
	return s.repo.List(name, page)
}

// ==================== Warehouse Service ====================

type WarehouseService struct{ repo *repository.WarehouseRepo }

func NewWarehouseService(repo *repository.WarehouseRepo) *WarehouseService {
	return &WarehouseService{repo}
}

func (s *WarehouseService) Create(req *dto.CreateWarehouseReq) error {
	return s.repo.Create(&model.Warehouse{
		Code: req.Code, Name: req.Name, Address: req.Address,
		Keeper: req.Keeper, Remark: req.Remark, Status: 1,
	})
}

func (s *WarehouseService) GetByID(id uint) (*model.Warehouse, error) {
	w, err := s.repo.FindByID(id)
	if err != nil {
		return nil, bizErr.ErrNotFound
	}
	return w, nil
}

func (s *WarehouseService) Update(id uint, req *dto.UpdateWarehouseReq) error {
	w, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrNotFound
	}
	if req.Name != "" {
		w.Name = req.Name
	}
	w.Address = req.Address
	w.Keeper = req.Keeper
	w.Remark = req.Remark
	if req.Status != nil {
		w.Status = *req.Status
	}
	return s.repo.Update(w)
}

func (s *WarehouseService) List() ([]model.Warehouse, error) { return s.repo.List() }

// ==================== Purchase Service ====================

type PurchaseService struct {
	db          *gorm.DB
	repo        *repository.PurchaseRepo
	invRepo     *repository.InventoryRepo
}

func NewPurchaseService(db *gorm.DB, repo *repository.PurchaseRepo, invRepo *repository.InventoryRepo) *PurchaseService {
	return &PurchaseService{db, repo, invRepo}
}

func (s *PurchaseService) Create(req *dto.CreatePurchaseReq, operatorID uint) (*model.PurchaseOrder, error) {
	order := &model.PurchaseOrder{
		OrderNo:     util.GenerateOrderNo("PO"),
		SupplierID:  req.SupplierID,
		WarehouseID: req.WarehouseID,
		Status:      "draft",
		Remark:      req.Remark,
		OperatorID:  operatorID,
	}
	var total float64
	for _, item := range req.Items {
		amount := item.Quantity * item.Price
		total += amount
		order.Items = append(order.Items, model.PurchaseOrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Amount:    amount,
			Remark:    item.Remark,
		})
	}
	order.TotalAmount = total

	var err error
	s.db.Transaction(func(tx *gorm.DB) error {
		err = s.repo.Create(tx, order)
		return err
	})
	return order, err
}

func (s *PurchaseService) GetByID(id uint) (*model.PurchaseOrder, error) {
	o, err := s.repo.FindByID(id)
	if err != nil {
		return nil, bizErr.ErrPurchaseNotFound
	}
	return o, nil
}

func (s *PurchaseService) List(q *dto.PurchaseQuery, page *pagination.Pagination) ([]model.PurchaseOrder, int64, error) {
	return s.repo.List(q, page)
}

// Approve 审核采购单 draft → approved
func (s *PurchaseService) Approve(id, operatorID uint) error {
	o, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrPurchaseNotFound
	}
	if o.Status != "draft" {
		return bizErr.ErrPurchaseStatusFlow
	}
	now := time.Now()
	o.Status = "approved"
	o.ApprovedAt = &now
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.repo.Update(tx, o)
	})
}

// Receive 入库确认 approved → received（核心：事务操作库存）
func (s *PurchaseService) Receive(id, operatorID uint) error {
	o, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrPurchaseNotFound
	}
	if o.Status != "approved" {
		return bizErr.ErrPurchaseStatusFlow
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range o.Items {
			// 获取或创建库存台账
			inv, err := s.invRepo.GetOrCreate(tx, item.ProductID, o.WarehouseID)
			if err != nil {
				return err
			}
			before := inv.Quantity
			// 增加库存
			if err := s.invRepo.UpdateQuantity(tx, inv.ID, item.Quantity); err != nil {
				return err
			}
			// 写流水
			txRecord := &model.InventoryTransaction{
				ProductID:   item.ProductID,
				WarehouseID: o.WarehouseID,
				TxType:      model.TxTypePurchaseIn,
				Quantity:    item.Quantity,
				BeforeQty:   before,
				AfterQty:    before + item.Quantity,
				RefID:       o.ID,
				RefType:     "purchase_order",
				OperatorID:  operatorID,
				Remark:      "采购入库 " + o.OrderNo,
			}
			if err := s.invRepo.AddTransaction(tx, txRecord); err != nil {
				return err
			}
		}
		now := time.Now()
		o.Status = "received"
		o.ReceivedAt = &now
		return s.repo.Update(tx, o)
	})
}

// Cancel 取消采购单
func (s *PurchaseService) Cancel(id uint) error {
	o, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrPurchaseNotFound
	}
	if o.Status == "received" || o.Status == "cancelled" {
		return bizErr.ErrPurchaseStatusFlow
	}
	o.Status = "cancelled"
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.repo.Update(tx, o)
	})
}

// ==================== Sale Service ====================

type SaleService struct {
	db      *gorm.DB
	repo    *repository.SaleRepo
	invRepo *repository.InventoryRepo
}

func NewSaleService(db *gorm.DB, repo *repository.SaleRepo, invRepo *repository.InventoryRepo) *SaleService {
	return &SaleService{db, repo, invRepo}
}

func (s *SaleService) Create(req *dto.CreateSaleReq, operatorID uint) (*model.SaleOrder, error) {
	order := &model.SaleOrder{
		OrderNo:     util.GenerateOrderNo("SO"),
		CustomerID:  req.CustomerID,
		WarehouseID: req.WarehouseID,
		Status:      "draft",
		Remark:      req.Remark,
		OperatorID:  operatorID,
	}
	var total float64
	for _, item := range req.Items {
		amount := item.Quantity * item.Price
		total += amount
		order.Items = append(order.Items, model.SaleOrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Amount:    amount,
			Remark:    item.Remark,
		})
	}
	order.TotalAmount = total

	var err error
	s.db.Transaction(func(tx *gorm.DB) error {
		err = s.repo.Create(tx, order)
		return err
	})
	return order, err
}

func (s *SaleService) GetByID(id uint) (*model.SaleOrder, error) {
	o, err := s.repo.FindByID(id)
	if err != nil {
		return nil, bizErr.ErrSaleNotFound
	}
	return o, nil
}

func (s *SaleService) List(q *dto.SaleQuery, page *pagination.Pagination) ([]model.SaleOrder, int64, error) {
	return s.repo.List(q, page)
}

// Approve 审核销售单
func (s *SaleService) Approve(id, operatorID uint) error {
	o, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrSaleNotFound
	}
	if o.Status != "draft" {
		return bizErr.ErrSaleStatusFlow
	}
	now := time.Now()
	o.Status = "approved"
	o.ApprovedAt = &now
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.repo.Update(tx, o)
	})
}

// Ship 出库确认 approved → shipped（核心：校验库存 + 事务扣减）
func (s *SaleService) Ship(id, operatorID uint) error {
	o, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrSaleNotFound
	}
	if o.Status != "approved" {
		return bizErr.ErrSaleStatusFlow
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range o.Items {
			inv, err := s.invRepo.GetOrCreate(tx, item.ProductID, o.WarehouseID)
			if err != nil {
				return err
			}
			// 库存充足校验
			if inv.Quantity < item.Quantity {
				return bizErr.ErrStockInsufficient
			}
			before := inv.Quantity
			// 扣减库存（负数）
			if err := s.invRepo.UpdateQuantity(tx, inv.ID, -item.Quantity); err != nil {
				return err
			}
			// 写流水
			txRecord := &model.InventoryTransaction{
				ProductID:   item.ProductID,
				WarehouseID: o.WarehouseID,
				TxType:      model.TxTypeSaleOut,
				Quantity:    -item.Quantity,
				BeforeQty:   before,
				AfterQty:    before - item.Quantity,
				RefID:       o.ID,
				RefType:     "sale_order",
				OperatorID:  operatorID,
				Remark:      "销售出库 " + o.OrderNo,
			}
			if err := s.invRepo.AddTransaction(tx, txRecord); err != nil {
				return err
			}
		}
		now := time.Now()
		o.Status = "shipped"
		o.ShippedAt = &now
		return s.repo.Update(tx, o)
	})
}

// Cancel 取消销售单
func (s *SaleService) Cancel(id uint) error {
	o, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrSaleNotFound
	}
	if o.Status == "shipped" || o.Status == "cancelled" {
		return bizErr.ErrSaleStatusFlow
	}
	o.Status = "cancelled"
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.repo.Update(tx, o)
	})
}

// ==================== Inventory Service ====================

type InventoryService struct {
	db      *gorm.DB
	invRepo *repository.InventoryRepo
}

func NewInventoryService(db *gorm.DB, invRepo *repository.InventoryRepo) *InventoryService {
	return &InventoryService{db, invRepo}
}

func (s *InventoryService) List(q *dto.InventoryQuery, page *pagination.Pagination) ([]model.Inventory, int64, error) {
	return s.invRepo.List(q, page)
}

func (s *InventoryService) ListTx(q *dto.TxQuery, page *pagination.Pagination) ([]model.InventoryTransaction, int64, error) {
	return s.invRepo.ListTx(q, page)
}

// Transfer 库存调拨（原子操作）
func (s *InventoryService) Transfer(req *dto.TransferReq, operatorID uint) error {
	if req.FromWarehouseID == req.ToWarehouseID {
		return bizErr.New(40003, "源仓库和目标仓库不能相同")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 出库仓库
		from, err := s.invRepo.GetOrCreate(tx, req.ProductID, req.FromWarehouseID)
		if err != nil {
			return err
		}
		if from.Quantity < req.Quantity {
			return bizErr.ErrStockInsufficient
		}
		fromBefore := from.Quantity

		// 入库仓库
		to, err := s.invRepo.GetOrCreate(tx, req.ProductID, req.ToWarehouseID)
		if err != nil {
			return err
		}
		toBefore := to.Quantity

		// 更新数量
		if err := s.invRepo.UpdateQuantity(tx, from.ID, -req.Quantity); err != nil {
			return err
		}
		if err := s.invRepo.UpdateQuantity(tx, to.ID, req.Quantity); err != nil {
			return err
		}

		// 写两条流水
		records := []*model.InventoryTransaction{
			{
				ProductID: req.ProductID, WarehouseID: req.FromWarehouseID,
				TxType: model.TxTypeTransferOut, Quantity: -req.Quantity,
				BeforeQty: fromBefore, AfterQty: fromBefore - req.Quantity,
				OperatorID: operatorID, Remark: req.Remark,
			},
			{
				ProductID: req.ProductID, WarehouseID: req.ToWarehouseID,
				TxType: model.TxTypeTransferIn, Quantity: req.Quantity,
				BeforeQty: toBefore, AfterQty: toBefore + req.Quantity,
				OperatorID: operatorID, Remark: req.Remark,
			},
		}
		for _, r := range records {
			if err := s.invRepo.AddTransaction(tx, r); err != nil {
				return err
			}
		}
		return nil
	})
}

// ManualAdj 手动调整库存
func (s *InventoryService) ManualAdj(req *dto.ManualAdjReq, operatorID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		inv, err := s.invRepo.GetOrCreate(tx, req.ProductID, req.WarehouseID)
		if err != nil {
			return err
		}
		if inv.Quantity+req.Quantity < 0 {
			return bizErr.ErrStockInsufficient
		}
		before := inv.Quantity
		if err := s.invRepo.UpdateQuantity(tx, inv.ID, req.Quantity); err != nil {
			return err
		}
		return s.invRepo.AddTransaction(tx, &model.InventoryTransaction{
			ProductID: req.ProductID, WarehouseID: req.WarehouseID,
			TxType: model.TxTypeManualAdj, Quantity: req.Quantity,
			BeforeQty: before, AfterQty: before + req.Quantity,
			OperatorID: operatorID, Remark: req.Remark,
		})
	})
}

// ==================== Stocktake Service ====================

type StocktakeService struct {
	db          *gorm.DB
	repo        *repository.StocktakeRepo
	invRepo     *repository.InventoryRepo
}

func NewStocktakeService(db *gorm.DB, repo *repository.StocktakeRepo, invRepo *repository.InventoryRepo) *StocktakeService {
	return &StocktakeService{db, repo, invRepo}
}

func (s *StocktakeService) Create(req *dto.CreateStocktakeReq, operatorID uint) (*model.Stocktake, error) {
	st := &model.Stocktake{
		OrderNo:     util.GenerateOrderNo("ST"),
		WarehouseID: req.WarehouseID,
		Status:      "draft",
		OperatorID:  operatorID,
		Remark:      req.Remark,
	}
	// 查询系统库存作为基准
	for _, item := range req.Items {
		inv, _ := s.invRepo.GetOrCreate(s.db, item.ProductID, req.WarehouseID)
		diff := item.ActualQty - inv.Quantity
		st.Items = append(st.Items, model.StocktakeItem{
			ProductID: item.ProductID,
			SystemQty: inv.Quantity,
			ActualQty: item.ActualQty,
			DiffQty:   diff,
		})
	}

	var err error
	s.db.Transaction(func(tx *gorm.DB) error {
		err = s.repo.Create(tx, st)
		return err
	})
	return st, err
}

func (s *StocktakeService) GetByID(id uint) (*model.Stocktake, error) {
	st, err := s.repo.FindByID(id)
	if err != nil {
		return nil, bizErr.ErrNotFound
	}
	return st, nil
}

func (s *StocktakeService) List(page *pagination.Pagination) ([]model.Stocktake, int64, error) {
	return s.repo.List(page)
}

// Confirm 确认盘点（生成调整流水）
func (s *StocktakeService) Confirm(id, operatorID uint) error {
	st, err := s.repo.FindByID(id)
	if err != nil {
		return bizErr.ErrNotFound
	}
	if st.Status != "draft" {
		return bizErr.New(70001, "盘点单已确认")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range st.Items {
			if item.DiffQty == 0 {
				continue
			}
			inv, err := s.invRepo.GetOrCreate(tx, item.ProductID, st.WarehouseID)
			if err != nil {
				return err
			}
			before := inv.Quantity
			if err := s.invRepo.UpdateQuantity(tx, inv.ID, item.DiffQty); err != nil {
				return err
			}
			if err := s.invRepo.AddTransaction(tx, &model.InventoryTransaction{
				ProductID: item.ProductID, WarehouseID: st.WarehouseID,
				TxType: model.TxTypeStocktake, Quantity: item.DiffQty,
				BeforeQty: before, AfterQty: before + item.DiffQty,
				RefID: st.ID, RefType: "stocktake",
				OperatorID: operatorID, Remark: "盘点调整 " + st.OrderNo,
			}); err != nil {
				return err
			}
		}
		st.Status = "confirmed"
		return s.repo.Update(tx, st)
	})
}
