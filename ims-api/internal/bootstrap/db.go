package bootstrap

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"ims-api/internal/model"
)

var DB *gorm.DB

// InitDB 初始化数据库
func InitDB(dsn string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}

	sqlDB, _ := DB.DB()
	sqlDB.SetMaxOpenConns(1) // SQLite 单写
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 开启 WAL 模式 & 外键
	DB.Exec("PRAGMA journal_mode=WAL")
	DB.Exec("PRAGMA foreign_keys=ON")

	// 自动迁移
	if err = autoMigrate(); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	// 初始化种子数据
	if err = seed(); err != nil {
		return fmt.Errorf("seed: %w", err)
	}

	return nil
}

func autoMigrate() error {
	return DB.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Product{},
		&model.Supplier{},
		&model.Customer{},
		&model.Warehouse{},
		&model.PurchaseOrder{},
		&model.PurchaseOrderItem{},
		&model.SaleOrder{},
		&model.SaleOrderItem{},
		&model.Inventory{},
		&model.InventoryTransaction{},
		&model.Stocktake{},
		&model.StocktakeItem{},
	)
}

func seed() error {
	// 创建默认管理员（只在首次）
	var count int64
	DB.Model(&model.User{}).Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		admin := model.User{
			Username: "admin",
			Password: string(hash),
			RealName: "系统管理员",
			Role:     "admin",
			Status:   1,
		}
		if err := DB.Create(&admin).Error; err != nil {
			return err
		}

		// 默认仓库
		warehouse := model.Warehouse{
			Code:   "WH001",
			Name:   "主仓库",
			Keeper: "管理员",
			Status: 1,
		}
		DB.Create(&warehouse)

		// 默认分类
		category := model.Category{Name: "默认分类"}
		DB.Create(&category)

		zap.L().Info("seed data initialized, admin/admin123")
	}
	return nil
}

// InitLogger 初始化日志
func InitLogger(logPath string) {
	os.MkdirAll("./logs", 0755)

	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout", logPath}
	cfg.ErrorOutputPaths = []string{"stderr"}

	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(l)
}
