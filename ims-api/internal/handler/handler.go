package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ims-api/internal/dto"
	"ims-api/internal/middleware"
	"ims-api/internal/service"
	bizErr "ims-api/pkg/errors"
	"ims-api/pkg/pagination"
	"ims-api/pkg/response"
)

// ==================== Auth Handler ====================

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc} }

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, bizErr.ErrParam)
		return
	}
	resp, err := h.svc.Login(&req)
	if err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, resp)
}

// ==================== User Handler ====================

type UserHandler struct{ svc *service.UserService }

func NewUserHandler(svc *service.UserService) *UserHandler { return &UserHandler{svc} }

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	if err := h.svc.Create(&req); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	if err := h.svc.Update(uint(id), &req); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	id := middleware.CurrentUserID(c)
	var req dto.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	if err := h.svc.ChangePassword(id, &req); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *UserHandler) List(c *gin.Context) {
	page := pagination.FromQuery(c)
	list, total, err := h.svc.List(page)
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OKPage(c, list, total, page.Page, page.PageSize)
}

// ==================== Category Handler ====================

type CategoryHandler struct{ svc *service.CategoryService }

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler { return &CategoryHandler{svc} }

func (h *CategoryHandler) Create(c *gin.Context) {
	var req dto.CreateCategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	if err := h.svc.Create(&req); err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OK(c, nil)
}

func (h *CategoryHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OK(c, list)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(uint(id)); err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OK(c, nil)
}

// ==================== Product Handler ====================

type ProductHandler struct{ svc *service.ProductService }

func NewProductHandler(svc *service.ProductService) *ProductHandler { return &ProductHandler{svc} }

func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	if err := h.svc.Create(&req); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, err := h.svc.GetByID(uint(id))
	if err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, p)
}

func (h *ProductHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateProductReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	if err := h.svc.Update(uint(id), &req); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *ProductHandler) List(c *gin.Context) {
	var q dto.ProductQuery
	c.ShouldBindQuery(&q)
	page := pagination.FromQuery(c)
	list, total, err := h.svc.List(&q, page)
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OKPage(c, list, total, page.Page, page.PageSize)
}

// ==================== Supplier Handler ====================

type SupplierHandler struct{ svc *service.SupplierService }

func NewSupplierHandler(svc *service.SupplierService) *SupplierHandler { return &SupplierHandler{svc} }

func (h *SupplierHandler) Create(c *gin.Context) {
	var req dto.CreateSupplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	if err := h.svc.Create(&req); err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OK(c, nil)
}

func (h *SupplierHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	s, err := h.svc.GetByID(uint(id))
	if err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, s)
}

func (h *SupplierHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateSupplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	if err := h.svc.Update(uint(id), &req); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *SupplierHandler) List(c *gin.Context) {
	name := c.Query("name")
	page := pagination.FromQuery(c)
	list, total, err := h.svc.List(name, page)
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OKPage(c, list, total, page.Page, page.PageSize)
}

// ==================== Customer Handler ====================

type CustomerHandler struct{ svc *service.CustomerService }

func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler { return &CustomerHandler{svc} }

func (h *CustomerHandler) Create(c *gin.Context) {
	var req dto.CreateCustomerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	if err := h.svc.Create(&req); err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OK(c, nil)
}

func (h *CustomerHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	cus, err := h.svc.GetByID(uint(id))
	if err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, cus)
}

func (h *CustomerHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateCustomerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	if err := h.svc.Update(uint(id), &req); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *CustomerHandler) List(c *gin.Context) {
	name := c.Query("name")
	page := pagination.FromQuery(c)
	list, total, err := h.svc.List(name, page)
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OKPage(c, list, total, page.Page, page.PageSize)
}

// ==================== Warehouse Handler ====================

type WarehouseHandler struct{ svc *service.WarehouseService }

func NewWarehouseHandler(svc *service.WarehouseService) *WarehouseHandler {
	return &WarehouseHandler{svc}
}

func (h *WarehouseHandler) Create(c *gin.Context) {
	var req dto.CreateWarehouseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	if err := h.svc.Create(&req); err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OK(c, nil)
}

func (h *WarehouseHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	w, err := h.svc.GetByID(uint(id))
	if err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, w)
}

func (h *WarehouseHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateWarehouseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	if err := h.svc.Update(uint(id), &req); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *WarehouseHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OK(c, list)
}

// ==================== Purchase Handler ====================

type PurchaseHandler struct{ svc *service.PurchaseService }

func NewPurchaseHandler(svc *service.PurchaseService) *PurchaseHandler { return &PurchaseHandler{svc} }

func (h *PurchaseHandler) Create(c *gin.Context) {
	var req dto.CreatePurchaseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	operatorID := middleware.CurrentUserID(c)
	o, err := h.svc.Create(&req, operatorID)
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OK(c, o)
}

func (h *PurchaseHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	o, err := h.svc.GetByID(uint(id))
	if err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, o)
}

func (h *PurchaseHandler) List(c *gin.Context) {
	var q dto.PurchaseQuery
	c.ShouldBindQuery(&q)
	page := pagination.FromQuery(c)
	list, total, err := h.svc.List(&q, page)
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OKPage(c, list, total, page.Page, page.PageSize)
}

func (h *PurchaseHandler) Approve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	operatorID := middleware.CurrentUserID(c)
	if err := h.svc.Approve(uint(id), operatorID); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *PurchaseHandler) Receive(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	operatorID := middleware.CurrentUserID(c)
	if err := h.svc.Receive(uint(id), operatorID); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *PurchaseHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Cancel(uint(id)); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

// ==================== Sale Handler ====================

type SaleHandler struct{ svc *service.SaleService }

func NewSaleHandler(svc *service.SaleService) *SaleHandler { return &SaleHandler{svc} }

func (h *SaleHandler) Create(c *gin.Context) {
	var req dto.CreateSaleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	operatorID := middleware.CurrentUserID(c)
	o, err := h.svc.Create(&req, operatorID)
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OK(c, o)
}

func (h *SaleHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	o, err := h.svc.GetByID(uint(id))
	if err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, o)
}

func (h *SaleHandler) List(c *gin.Context) {
	var q dto.SaleQuery
	c.ShouldBindQuery(&q)
	page := pagination.FromQuery(c)
	list, total, err := h.svc.List(&q, page)
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OKPage(c, list, total, page.Page, page.PageSize)
}

func (h *SaleHandler) Approve(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	operatorID := middleware.CurrentUserID(c)
	if err := h.svc.Approve(uint(id), operatorID); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *SaleHandler) Ship(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	operatorID := middleware.CurrentUserID(c)
	if err := h.svc.Ship(uint(id), operatorID); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *SaleHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Cancel(uint(id)); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

// ==================== Inventory Handler ====================

type InventoryHandler struct{ svc *service.InventoryService }

func NewInventoryHandler(svc *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{svc}
}

func (h *InventoryHandler) List(c *gin.Context) {
	var q dto.InventoryQuery
	c.ShouldBindQuery(&q)
	page := pagination.FromQuery(c)
	list, total, err := h.svc.List(&q, page)
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OKPage(c, list, total, page.Page, page.PageSize)
}

func (h *InventoryHandler) ListTx(c *gin.Context) {
	var q dto.TxQuery
	c.ShouldBindQuery(&q)
	page := pagination.FromQuery(c)
	list, total, err := h.svc.ListTx(&q, page)
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OKPage(c, list, total, page.Page, page.PageSize)
}

func (h *InventoryHandler) Transfer(c *gin.Context) {
	var req dto.TransferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	operatorID := middleware.CurrentUserID(c)
	if err := h.svc.Transfer(&req, operatorID); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

func (h *InventoryHandler) ManualAdj(c *gin.Context) {
	var req dto.ManualAdjReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	operatorID := middleware.CurrentUserID(c)
	if err := h.svc.ManualAdj(&req, operatorID); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}

// ==================== Stocktake Handler ====================

type StocktakeHandler struct{ svc *service.StocktakeService }

func NewStocktakeHandler(svc *service.StocktakeService) *StocktakeHandler {
	return &StocktakeHandler{svc}
}

func (h *StocktakeHandler) Create(c *gin.Context) {
	var req dto.CreateStocktakeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailMsg(c, bizErr.ErrParam.Code, err.Error())
		return
	}
	operatorID := middleware.CurrentUserID(c)
	st, err := h.svc.Create(&req, operatorID)
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OK(c, st)
}

func (h *StocktakeHandler) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	st, err := h.svc.GetByID(uint(id))
	if err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, st)
}

func (h *StocktakeHandler) List(c *gin.Context) {
	page := pagination.FromQuery(c)
	list, total, err := h.svc.List(page)
	if err != nil {
		response.Fail(c, bizErr.ErrServer)
		return
	}
	response.OKPage(c, list, total, page.Page, page.PageSize)
}

func (h *StocktakeHandler) Confirm(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	operatorID := middleware.CurrentUserID(c)
	if err := h.svc.Confirm(uint(id), operatorID); err != nil {
		if be, ok := err.(*bizErr.BizError); ok {
			response.Fail(c, be)
		} else {
			response.Fail(c, bizErr.ErrServer)
		}
		return
	}
	response.OK(c, nil)
}
