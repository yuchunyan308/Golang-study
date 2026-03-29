package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"ims-api/internal/bootstrap"
	"ims-api/internal/handler"
	"ims-api/internal/middleware"
	"ims-api/internal/repository"
	"ims-api/internal/router"
	"ims-api/internal/service"
)

func main() {
	// ---- 加载配置 ----
	viper.SetConfigFile("config/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic("load config: " + err.Error())
	}

	// ---- 初始化日志 ----
	bootstrap.InitLogger(viper.GetString("log.filename"))
	defer zap.L().Sync()

	// ---- 初始化数据库 ----
	if err := bootstrap.InitDB(viper.GetString("database.dsn")); err != nil {
		zap.L().Fatal("init db", zap.Error(err))
	}
	db := bootstrap.DB

	// ---- 设置JWT密钥 ----
	middleware.SetJWTSecret(viper.GetString("app.jwt_secret"))

	// ---- 依赖注入：Repository → Service → Handler ----
	userRepo := repository.NewUserRepo(db)
	catRepo := repository.NewCategoryRepo(db)
	prodRepo := repository.NewProductRepo(db)
	supRepo := repository.NewSupplierRepo(db)
	cusRepo := repository.NewCustomerRepo(db)
	whRepo := repository.NewWarehouseRepo(db)
	purchRepo := repository.NewPurchaseRepo(db)
	saleRepo := repository.NewSaleRepo(db)
	invRepo := repository.NewInventoryRepo(db)
	stRepo := repository.NewStocktakeRepo(db)

	authSvc := service.NewAuthService(userRepo, viper.GetString("app.jwt_secret"), viper.GetInt("app.jwt_expire_hours"))
	userSvc := service.NewUserService(userRepo)
	catSvc := service.NewCategoryService(catRepo)
	prodSvc := service.NewProductService(prodRepo)
	supSvc := service.NewSupplierService(supRepo)
	cusSvc := service.NewCustomerService(cusRepo)
	whSvc := service.NewWarehouseService(whRepo)
	purchSvc := service.NewPurchaseService(db, purchRepo, invRepo)
	saleSvc := service.NewSaleService(db, saleRepo, invRepo)
	invSvc := service.NewInventoryService(db, invRepo)
	stSvc := service.NewStocktakeService(db, stRepo, invRepo)

	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc)
	catH := handler.NewCategoryHandler(catSvc)
	prodH := handler.NewProductHandler(prodSvc)
	supH := handler.NewSupplierHandler(supSvc)
	cusH := handler.NewCustomerHandler(cusSvc)
	whH := handler.NewWarehouseHandler(whSvc)
	purchH := handler.NewPurchaseHandler(purchSvc)
	saleH := handler.NewSaleHandler(saleSvc)
	invH := handler.NewInventoryHandler(invSvc)
	stH := handler.NewStocktakeHandler(stSvc)

	// ---- 创建Gin引擎 ----
	gin.SetMode(viper.GetString("app.mode"))
	engine := gin.New()

	// ---- 注册路由 ----
	router.Setup(engine, authH, userH, catH, prodH, supH, cusH, whH, purchH, saleH, invH, stH)

	// ---- 启动服务 ----
	addr := fmt.Sprintf(":%d", viper.GetInt("app.port"))
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		zap.L().Info("server starting", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("server error", zap.Error(err))
		}
	}()

	// ---- 优雅停机 ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.L().Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Error("shutdown error", zap.Error(err))
	}
	zap.L().Info("server exited")
}
