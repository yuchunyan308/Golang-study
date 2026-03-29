package router

import (
	"github.com/gin-gonic/gin"

	"ims-api/internal/handler"
	"ims-api/internal/middleware"
)

// Setup 注册所有路由
func Setup(
	engine *gin.Engine,
	authH *handler.AuthHandler,
	userH *handler.UserHandler,
	catH *handler.CategoryHandler,
	prodH *handler.ProductHandler,
	supH *handler.SupplierHandler,
	cusH *handler.CustomerHandler,
	whH *handler.WarehouseHandler,
	purchH *handler.PurchaseHandler,
	saleH *handler.SaleHandler,
	invH *handler.InventoryHandler,
	stH *handler.StocktakeHandler,
) {
	// 全局中间件
	engine.Use(middleware.Recovery(), middleware.TraceID(), middleware.Logger(), middleware.CORS())

	// 健康检查
	engine.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"message": "pong"}) })

	api := engine.Group("/api/v1")

	// 公开路由
	api.POST("/auth/login", authH.Login)

	// 需要认证的路由
	auth := api.Group("", middleware.JWT())

	// ---- 用户管理 ----
	users := auth.Group("/users")
	users.GET("", userH.List)
	users.POST("", middleware.RequireRole("admin"), userH.Create)
	users.PUT("/:id", middleware.RequireRole("admin"), userH.Update)
	users.POST("/change-password", userH.ChangePassword)

	// ---- 分类管理 ----
	cats := auth.Group("/categories")
	cats.GET("", catH.List)
	cats.POST("", middleware.RequireRole("admin", "operator"), catH.Create)
	cats.DELETE("/:id", middleware.RequireRole("admin"), catH.Delete)

	// ---- 商品管理 ----
	prods := auth.Group("/products")
	prods.GET("", prodH.List)
	prods.GET("/:id", prodH.GetByID)
	prods.POST("", middleware.RequireRole("admin", "operator"), prodH.Create)
	prods.PUT("/:id", middleware.RequireRole("admin", "operator"), prodH.Update)

	// ---- 供应商管理 ----
	sups := auth.Group("/suppliers")
	sups.GET("", supH.List)
	sups.GET("/:id", supH.GetByID)
	sups.POST("", middleware.RequireRole("admin", "operator"), supH.Create)
	sups.PUT("/:id", middleware.RequireRole("admin", "operator"), supH.Update)

	// ---- 客户管理 ----
	cuss := auth.Group("/customers")
	cuss.GET("", cusH.List)
	cuss.GET("/:id", cusH.GetByID)
	cuss.POST("", middleware.RequireRole("admin", "operator"), cusH.Create)
	cuss.PUT("/:id", middleware.RequireRole("admin", "operator"), cusH.Update)

	// ---- 仓库管理 ----
	whs := auth.Group("/warehouses")
	whs.GET("", whH.List)
	whs.GET("/:id", whH.GetByID)
	whs.POST("", middleware.RequireRole("admin"), whH.Create)
	whs.PUT("/:id", middleware.RequireRole("admin"), whH.Update)

	// ---- 采购管理 ----
	purch := auth.Group("/purchases")
	purch.GET("", purchH.List)
	purch.GET("/:id", purchH.GetByID)
	purch.POST("", middleware.RequireRole("admin", "operator"), purchH.Create)
	purch.PUT("/:id/approve", middleware.RequireRole("admin"), purchH.Approve)
	purch.PUT("/:id/receive", middleware.RequireRole("admin", "operator"), purchH.Receive)
	purch.PUT("/:id/cancel", middleware.RequireRole("admin"), purchH.Cancel)

	// ---- 销售管理 ----
	sales := auth.Group("/sales")
	sales.GET("", saleH.List)
	sales.GET("/:id", saleH.GetByID)
	sales.POST("", middleware.RequireRole("admin", "operator"), saleH.Create)
	sales.PUT("/:id/approve", middleware.RequireRole("admin"), saleH.Approve)
	sales.PUT("/:id/ship", middleware.RequireRole("admin", "operator"), saleH.Ship)
	sales.PUT("/:id/cancel", middleware.RequireRole("admin"), saleH.Cancel)

	// ---- 库存管理 ----
	inv := auth.Group("/inventory")
	inv.GET("", invH.List)
	inv.GET("/transactions", invH.ListTx)
	inv.POST("/transfer", middleware.RequireRole("admin", "operator"), invH.Transfer)
	inv.POST("/adjust", middleware.RequireRole("admin"), invH.ManualAdj)

	// ---- 盘点管理 ----
	st := auth.Group("/stocktakes")
	st.GET("", stH.List)
	st.GET("/:id", stH.GetByID)
	st.POST("", middleware.RequireRole("admin", "operator"), stH.Create)
	st.PUT("/:id/confirm", middleware.RequireRole("admin"), stH.Confirm)
}
