# 用户与订单管理系统 — 完整实现

---

## 📁 项目目录结构

```
user-order-api/
├── go.mod
├── go.sum
├── main.go
├── models/
│   ├── user.go
│   └── order.go
├── repository/
│   ├── db.go
│   ├── user_repo.go
│   └── order_repo.go
├── service/
│   ├── user_service.go
│   └── order_service.go
├── handler/
│   ├── user_handler.go
│   └── order_handler.go
└── response/
    └── response.go
```

---

## 📦 go.mod

```go
module user-order-api

go 1.22

require (
    github.com/go-chi/chi/v5 v5.1.0
    modernc.org/sqlite v1.33.1
)
```

---

## 📄 models/user.go

```go
package models

import "time"

// User 对应 users 主表
type User struct {
    ID        int64     `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
    Orders    []Order   `json:"orders,omitempty"` // 级联查询时填充
}

// CreateUserRequest 创建用户的请求体
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Validate 基础校验
func (r *CreateUserRequest) Validate() error {
    if r.Name == "" {
        return ErrFieldRequired("name")
    }
    if r.Email == "" {
        return ErrFieldRequired("email")
    }
    return nil
}
```

---

## 📄 models/order.go

```go
package models

import (
    "errors"
    "fmt"
    "time"
)

// OrderStatus 订单状态枚举
type OrderStatus string

const (
    StatusPending   OrderStatus = "pending"
    StatusPaid      OrderStatus = "paid"
    StatusCancelled OrderStatus = "cancelled"
)

// Order 对应 orders 从表
type Order struct {
    ID        int64       `json:"id"`
    UserID    int64       `json:"user_id"`
    Amount    float64     `json:"amount"`
    Status    OrderStatus `json:"status"`
    CreatedAt time.Time   `json:"created_at"`
}

// CreateOrderRequest 创建订单的请求体
type CreateOrderRequest struct {
    UserID int64       `json:"user_id"`
    Amount float64     `json:"amount"`
    Status OrderStatus `json:"status"` // 可选，默认 pending
}

// Validate 基础校验
func (r *CreateOrderRequest) Validate() error {
    if r.UserID <= 0 {
        return ErrFieldRequired("user_id")
    }
    if r.Amount <= 0 {
        return errors.New("amount must be greater than 0")
    }
    // 状态校验：为空则使用默认值
    if r.Status == "" {
        r.Status = StatusPending
    }
    switch r.Status {
    case StatusPending, StatusPaid, StatusCancelled:
    default:
        return fmt.Errorf("invalid status: %q, must be one of [pending, paid, cancelled]", r.Status)
    }
    return nil
}

// — 公共错误工具 —

// ValidationError 代表业务校验错误（区别于系统错误）
type ValidationError struct {
    Msg string
}

func (e *ValidationError) Error() string { return e.Msg }

// ErrFieldRequired 快速构造必填错误
func ErrFieldRequired(field string) *ValidationError {
    return &ValidationError{Msg: fmt.Sprintf("field %q is required", field)}
}

// ErrNotFound 资源不存在
var ErrNotFound = errors.New("resource not found")

// ErrDuplicateEmail 邮箱重复
var ErrDuplicateEmail = errors.New("email already exists")

// ErrUserNotFound 用户不存在
var ErrUserNotFound = errors.New("user not found")
```

---

## 📄 response/response.go

```go
package response

import (
    "encoding/json"
    "log"
    "net/http"
)

// Envelope 统一响应封装
type Envelope struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    any    `json:"data"` // Go 1.18+ any 别名
}

// write 是内部写入方法
func write(w http.ResponseWriter, httpStatus, code int, message string, data any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(httpStatus)
    if err := json.NewEncoder(w).Encode(Envelope{
        Code:    code,
        Message: message,
        Data:    data,
    }); err != nil {
        log.Printf("response encoding error: %v", err)
    }
}

func Success(w http.ResponseWriter, data any) {
    write(w, http.StatusOK, 200, "success", data)
}

func Created(w http.ResponseWriter, data any) {
    write(w, http.StatusCreated, 201, "created", data)
}

func BadRequest(w http.ResponseWriter, message string) {
    write(w, http.StatusBadRequest, 400, message, nil)
}

func NotFound(w http.ResponseWriter, message string) {
    write(w, http.StatusNotFound, 404, message, nil)
}

func Conflict(w http.ResponseWriter, message string) {
    write(w, http.StatusConflict, 409, message, nil)
}

func InternalError(w http.ResponseWriter, message string) {
    // 生产环境不暴露内部错误细节，这里为调试方便保留
    write(w, http.StatusInternalServerError, 500, message, nil)
}
```

---

## 📄 repository/db.go

```go
package repository

import (
    "database/sql"
    "log"
    "sync"
    "time"

    _ "modernc.org/sqlite" // 纯 Go SQLite 驱动，无需 CGO
)

var (
    globalDB *sql.DB
    once     sync.Once
)

// InitDB 单例初始化数据库，幂等安全
func InitDB(dsn string) (*sql.DB, error) {
    var initErr error
    once.Do(func() {
        db, err := sql.Open("sqlite", dsn)
        if err != nil {
            initErr = err
            return
        }

        // SQLite 为单写模型，限制连接数防止锁竞争
        db.SetMaxOpenConns(1)
        db.SetMaxIdleConns(1)
        db.SetConnMaxLifetime(0)

        if err = db.Ping(); err != nil {
            initErr = err
            return
        }

        if err = runMigrations(db); err != nil {
            initErr = err
            return
        }

        globalDB = db
        log.Println("[DB] initialized successfully")
    })
    return globalDB, initErr
}

// GetDB 获取全局 DB 实例（已初始化后调用）
func GetDB() *sql.DB {
    return globalDB
}

// CloseDB 优雅关闭连接
func CloseDB() {
    if globalDB != nil {
        if err := globalDB.Close(); err != nil {
            log.Printf("[DB] close error: %v", err)
        } else {
            log.Println("[DB] connection closed")
        }
    }
}

// runMigrations 自动建表，保持幂等
func runMigrations(db *sql.DB) error {
    schema := `
    -- 开启外键约束（SQLite 默认关闭）
    PRAGMA foreign_keys = ON;
    PRAGMA journal_mode = WAL;   -- 提升并发读性能

    -- 用户主表
    CREATE TABLE IF NOT EXISTS users (
        id         INTEGER  PRIMARY KEY AUTOINCREMENT,
        name       TEXT     NOT NULL,
        email      TEXT     NOT NULL UNIQUE,
        created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
    );

    -- 订单从表（外键关联 users）
    CREATE TABLE IF NOT EXISTS orders (
        id         INTEGER  PRIMARY KEY AUTOINCREMENT,
        user_id    INTEGER  NOT NULL,
        amount     REAL     NOT NULL CHECK(amount > 0),
        status     TEXT     NOT NULL DEFAULT 'pending',
        created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );

    -- 加速按 user_id 查询订单
    CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
    `
    _, err := db.Exec(schema)
    return err
}

// parseTime 统一解析 SQLite 返回的时间字符串
func parseTime(s string) time.Time {
    formats := []string{
        time.RFC3339,
        "2006-01-02T15:04:05Z",
        "2006-01-02 15:04:05",
    }
    for _, f := range formats {
        if t, err := time.Parse(f, s); err == nil {
            return t
        }
    }
    log.Printf("[DB] cannot parse time: %q", s)
    return time.Time{}
}
```

---

## 📄 repository/user_repo.go

```go
package repository

import (
    "database/sql"
    "strings"

    "user-order-api/models"
)

// UserRepository 封装用户表所有 DB 操作
type UserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Create 插入新用户，返回完整记录
func (r *UserRepository) Create(req *models.CreateUserRequest) (*models.User, error) {
    const q = `INSERT INTO users (name, email) VALUES (?, ?)`
    result, err := r.db.Exec(q, req.Name, req.Email)
    if err != nil {
        // 检测唯一约束冲突
        if strings.Contains(err.Error(), "UNIQUE constraint failed") {
            return nil, models.ErrDuplicateEmail
        }
        return nil, err
    }
    id, err := result.LastInsertId()
    if err != nil {
        return nil, err
    }
    return r.FindByID(id)
}

// FindAll 查询所有用户列表
func (r *UserRepository) FindAll() ([]models.User, error) {
    const q = `SELECT id, name, email, created_at FROM users ORDER BY id DESC`
    rows, err := r.db.Query(q)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []models.User
    for rows.Next() {
        u, err := scanUser(rows)
        if err != nil {
            return nil, err
        }
        users = append(users, u)
    }
    return users, rows.Err()
}

// FindByID 按主键查找用户，不存在返回 (nil, nil)
func (r *UserRepository) FindByID(id int64) (*models.User, error) {
    const q = `SELECT id, name, email, created_at FROM users WHERE id = ? LIMIT 1`
    row := r.db.QueryRow(q, id)
    u, err := scanUser(row)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return &u, nil
}

// ExistsInTx 在事务中校验用户是否存在（供 Order 事务使用）
func (r *UserRepository) ExistsInTx(tx *sql.Tx, userID int64) (bool, error) {
    var count int
    err := tx.QueryRow(`SELECT COUNT(1) FROM users WHERE id = ?`, userID).Scan(&count)
    return count > 0, err
}

// — 内部扫描辅助 —

// scanner 抽象 *sql.Row 和 *sql.Rows，统一扫描逻辑
type scanner interface {
    Scan(dest ...any) error
}

func scanUser(s scanner) (models.User, error) {
    var u models.User
    var createdAt string
    err := s.Scan(&u.ID, &u.Name, &u.Email, &createdAt)
    if err != nil {
        return models.User{}, err
    }
    u.CreatedAt = parseTime(createdAt)
    return u, nil
}
```

---

## 📄 repository/order_repo.go

```go
package repository

import (
    "database/sql"

    "user-order-api/models"
)

// OrderRepository 封装订单表所有 DB 操作
type OrderRepository struct {
    db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
    return &OrderRepository{db: db}
}

// CreateWithTx 在事务内插入订单（由 OrderService 调用）
func (r *OrderRepository) CreateWithTx(tx *sql.Tx, req *models.CreateOrderRequest) (*models.Order, error) {
    const q = `INSERT INTO orders (user_id, amount, status) VALUES (?, ?, ?)`
    result, err := tx.Exec(q, req.UserID, req.Amount, req.Status)
    if err != nil {
        return nil, err
    }
    id, err := result.LastInsertId()
    if err != nil {
        return nil, err
    }
    // 在同一事务内查询刚插入的记录
    return r.findByIDInTx(tx, id)
}

// FindAll 查询订单列表，可选按 userID 筛选
func (r *OrderRepository) FindAll(userID *int64) ([]models.Order, error) {
    q := `SELECT id, user_id, amount, status, created_at FROM orders`
    args := []any{}
    if userID != nil {
        q += ` WHERE user_id = ?`
        args = append(args, *userID)
    }
    q += ` ORDER BY id DESC`

    rows, err := r.db.Query(q, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orders []models.Order
    for rows.Next() {
        o, err := scanOrder(rows)
        if err != nil {
            return nil, err
        }
        orders = append(orders, o)
    }
    return orders, rows.Err()
}

// FindByUserID 查询指定用户的所有订单
func (r *OrderRepository) FindByUserID(userID int64) ([]models.Order, error) {
    return r.FindAll(&userID)
}

// — 内部方法 —

func (r *OrderRepository) findByIDInTx(tx *sql.Tx, id int64) (*models.Order, error) {
    const q = `SELECT id, user_id, amount, status, created_at FROM orders WHERE id = ?`
    o, err := scanOrder(tx.QueryRow(q, id))
    if err != nil {
        return nil, err
    }
    return &o, nil
}

func scanOrder(s scanner) (models.Order, error) {
    var o models.Order
    var createdAt string
    var status string
    err := s.Scan(&o.ID, &o.UserID, &o.Amount, &status, &createdAt)
    if err != nil {
        return models.Order{}, err
    }
    o.Status = models.OrderStatus(status)
    o.CreatedAt = parseTime(createdAt)
    return o, nil
}
```

---

## 📄 service/user_service.go

```go
package service

import (
    "user-order-api/models"
    "user-order-api/repository"
)

// UserService 用户业务逻辑层
type UserService struct {
    userRepo  *repository.UserRepository
    orderRepo *repository.OrderRepository
}

func NewUserService(
    userRepo *repository.UserRepository,
    orderRepo *repository.OrderRepository,
) *UserService {
    return &UserService{userRepo: userRepo, orderRepo: orderRepo}
}

// CreateUser 创建用户（含参数校验）
func (s *UserService) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
    if err := req.Validate(); err != nil {
        return nil, err
    }
    return s.userRepo.Create(req)
}

// GetUsers 获取所有用户列表
func (s *UserService) GetUsers() ([]models.User, error) {
    return s.userRepo.FindAll()
}

// GetUserWithOrders 获取用户详情 + 级联查询其订单（两步查询）
func (s *UserService) GetUserWithOrders(id int64) (*models.User, error) {
    user, err := s.userRepo.FindByID(id)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, nil // 上层处理 404
    }

    orders, err := s.orderRepo.FindByUserID(id)
    if err != nil {
        return nil, err
    }
    // 保证 JSON 序列化时为 [] 而非 null
    if orders == nil {
        orders = []models.Order{}
    }
    user.Orders = orders
    return user, nil
}
```

---

## 📄 service/order_service.go

```go
package service

import (
    "database/sql"

    "user-order-api/models"
    "user-order-api/repository"
)

// OrderService 订单业务逻辑层
type OrderService struct {
    db        *sql.DB
    userRepo  *repository.UserRepository
    orderRepo *repository.OrderRepository
}

func NewOrderService(
    db *sql.DB,
    userRepo *repository.UserRepository,
    orderRepo *repository.OrderRepository,
) *OrderService {
    return &OrderService{db: db, userRepo: userRepo, orderRepo: orderRepo}
}

// CreateOrder 在事务中创建订单：先校验用户，再插入订单
func (s *OrderService) CreateOrder(req *models.CreateOrderRequest) (*models.Order, error) {
    // 1. 业务参数校验（在事务外，快速失败）
    if err := req.Validate(); err != nil {
        return nil, err
    }

    // 2. 开启事务
    tx, err := s.db.Begin()
    if err != nil {
        return nil, err
    }
    // defer 确保 panic 时也能回滚
    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback()
        }
    }()

    // 3. 事务内确认用户存在（避免脏读）
    exists, err := s.userRepo.ExistsInTx(tx, req.UserID)
    if err != nil {
        _ = tx.Rollback()
        return nil, err
    }
    if !exists {
        _ = tx.Rollback()
        return nil, models.ErrUserNotFound
    }

    // 4. 事务内插入订单
    order, err := s.orderRepo.CreateWithTx(tx, req)
    if err != nil {
        _ = tx.Rollback()
        return nil, err
    }

    // 5. 提交事务
    if err = tx.Commit(); err != nil {
        return nil, err
    }
    return order, nil
}

// GetOrders 查询订单列表，支持按 user_id 筛选
func (s *OrderService) GetOrders(userID *int64) ([]models.Order, error) {
    return s.orderRepo.FindAll(userID)
}
```

---

## 📄 handler/user_handler.go

```go
package handler

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "github.com/go-chi/chi/v5"

    "user-order-api/models"
    "user-order-api/response"
    "user-order-api/service"
)

// UserHandler HTTP 层，只负责解析请求 / 返回响应
type UserHandler struct {
    svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
    return &UserHandler{svc: svc}
}

// POST /users — 创建用户
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req models.CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "invalid request body: "+err.Error())
        return
    }

    user, err := h.svc.CreateUser(&req)
    if err != nil {
        switch {
        case errors.As(err, &models.ValidationError{}):
            response.BadRequest(w, err.Error())
        case errors.Is(err, models.ErrDuplicateEmail):
            response.Conflict(w, err.Error())
        default:
            response.InternalError(w, err.Error())
        }
        return
    }
    response.Created(w, user)
}

// GET /users — 获取用户列表
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
    users, err := h.svc.GetUsers()
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }
    if users == nil {
        users = []models.User{}
    }
    response.Success(w, users)
}

// GET /users/{id} — 获取用户详情（含订单）
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
    if err != nil || id <= 0 {
        response.BadRequest(w, "invalid user id")
        return
    }

    user, err := h.svc.GetUserWithOrders(id)
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }
    if user == nil {
        response.NotFound(w, "user not found")
        return
    }
    response.Success(w, user)
}
```

---

## 📄 handler/order_handler.go

```go
package handler

import (
    "encoding/json"
    "errors"
    "net/http"
    "strconv"

    "user-order-api/models"
    "user-order-api/response"
    "user-order-api/service"
)

// OrderHandler HTTP 层，只负责解析请求 / 返回响应
type OrderHandler struct {
    svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
    return &OrderHandler{svc: svc}
}

// POST /orders — 创建订单
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req models.CreateOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "invalid request body: "+err.Error())
        return
    }

    order, err := h.svc.CreateOrder(&req)
    if err != nil {
        var ve *models.ValidationError
        switch {
        case errors.As(err, &ve):
            response.BadRequest(w, err.Error())
        case errors.Is(err, models.ErrUserNotFound):
            response.NotFound(w, err.Error())
        default:
            response.InternalError(w, err.Error())
        }
        return
    }
    response.Created(w, order)
}

// GET /orders?user_id=1 — 查询订单（可选用户过滤）
func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
    var userID *int64
    if uidStr := r.URL.Query().Get("user_id"); uidStr != "" {
        uid, err := strconv.ParseInt(uidStr, 10, 64)
        if err != nil || uid <= 0 {
            response.BadRequest(w, "invalid user_id parameter")
            return
        }
        userID = &uid
    }

    orders, err := h.svc.GetOrders(userID)
    if err != nil {
        response.InternalError(w, err.Error())
        return
    }
    if orders == nil {
        orders = []models.Order{}
    }
    response.Success(w, orders)
}
```

---

## 📄 main.go

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"

    "user-order-api/handler"
    "user-order-api/repository"
    "user-order-api/service"
)

func main() {
    // ── 1. 初始化数据库（单例）──────────────────────────────
    db, err := repository.InitDB("./data.db")
    if err != nil {
        log.Fatalf("DB init failed: %v", err)
    }
    defer repository.CloseDB()

    // ── 2. 依赖注入（Repository → Service → Handler）───────
    userRepo  := repository.NewUserRepository(db)
    orderRepo := repository.NewOrderRepository(db)

    userSvc  := service.NewUserService(userRepo, orderRepo)
    orderSvc := service.NewOrderService(db, userRepo, orderRepo)

    userH  := handler.NewUserHandler(userSvc)
    orderH := handler.NewOrderHandler(orderSvc)

    // ── 3. 路由注册 ─────────────────────────────────────────
    r := chi.NewRouter()

    // 中间件
    r.Use(middleware.RequestID)   // 每个请求注入唯一 ID
    r.Use(middleware.RealIP)      // 解析真实客户端 IP
    r.Use(middleware.Logger)      // 结构化访问日志
    r.Use(middleware.Recoverer)   // panic → 500，不崩进程
    r.Use(middleware.Timeout(30 * time.Second)) // 超时保护

    // 用户路由
    r.Route("/users", func(r chi.Router) {
        r.Post("/", userH.CreateUser)
        r.Get("/", userH.GetUsers)
        r.Get("/{id}", userH.GetUser)
    })

    // 订单路由
    r.Route("/orders", func(r chi.Router) {
        r.Post("/", orderH.CreateOrder)
        r.Get("/", orderH.GetOrders)
    })

    // 健康检查
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"ok"}`))
    })

    // ── 4. 启动服务 + 优雅关闭 ──────────────────────────────
    srv := &http.Server{
        Addr:         ":8080",
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    // 非阻塞启动
    go func() {
        log.Println("Server listening on :8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server error: %v", err)
        }
    }()

    // 等待 SIGINT / SIGTERM 信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutdown signal received, draining connections...")

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Forced shutdown: %v", err)
    }
    log.Println("Server exited cleanly ✓")
}
```

---

## 🚀 启动项目

```bash
# 1. 初始化模块并下载依赖
go mod tidy

# 2. 运行服务（自动建表，生成 data.db）
go run .

# 或编译后运行
go build -o server . && ./server
```

---

## 🧪 完整 curl 测试流程

### Step 1 — 创建用户

```bash
# 创建用户 Alice
curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}' | jq .

# 创建用户 Bob
curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Bob","email":"bob@example.com"}' | jq .

# 测试邮箱重复（应返回 409）
curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice2","email":"alice@example.com"}' | jq .
```

### Step 2 — 查询用户列表

```bash
curl -s http://localhost:8080/users | jq .
```

### Step 3 — 创建订单

```bash
# 为 Alice (user_id=1) 创建两笔订单
curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"amount":199.99,"status":"paid"}' | jq .

curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"amount":88.00}' | jq .

# 为 Bob (user_id=2) 创建一笔订单
curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":2,"amount":350.00,"status":"pending"}' | jq .

# 测试用户不存在（应返回 404）
curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":999,"amount":100}' | jq .

# 测试金额非法（应返回 400）
curl -s -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id":1,"amount":-50}' | jq .
```

### Step 4 — 查询用户详情（级联查询订单）

```bash
# 获取 Alice 详情 + 她的所有订单
curl -s http://localhost:8080/users/1 | jq .

# 获取 Bob 详情 + 他的所有订单
curl -s http://localhost:8080/users/2 | jq .

# 不存在的用户（应返回 404）
curl -s http://localhost:8080/users/999 | jq .
```

### Step 5 — 查询订单列表（含筛选）

```bash
# 所有订单
curl -s http://localhost:8080/orders | jq .

# 只看 Alice 的订单
curl -s "http://localhost:8080/orders?user_id=1" | jq .
```

---

## 📊 期望响应示例

**`GET /users/1`** 的响应（级联含订单）：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com",
    "created_at": "2024-01-15T10:00:00Z",
    "orders": [
      {
        "id": 2,
        "user_id": 1,
        "amount": 88,
        "status": "pending",
        "created_at": "2024-01-15T10:01:20Z"
      },
      {
        "id": 1,
        "user_id": 1,
        "amount": 199.99,
        "status": "paid",
        "created_at": "2024-01-15T10:01:10Z"
      }
    ]
  }
}
```

---

## 🏗️ 架构总结

```
HTTP Request
    │
    ▼
┌─────────────┐     解析/校验 JSON
│   Handler   │──── 参数类型检查
└──────┬──────┘     统一错误响应
       │ 调用
       ▼
┌─────────────┐     业务规则校验
│   Service   │──── 事务编排
└──────┬──────┘     跨 Repo 协调
       │ 调用
       ▼
┌─────────────┐     SQL 构造
│ Repository  │──── 结果扫描
└──────┬──────┘     错误语义转换
       │
       ▼
   SQLite DB
```

| 层 | 职责 | 不做什么 |
|---|---|---|
| **Handler** | HTTP 解码/编码，入参校验格式 | 不写 SQL，不含业务逻辑 |
| **Service** | 业务规则、事务、跨资源编排 | 不接触 HTTP，不写 SQL |
| **Repository** | 单一资源的 CRUD | 不含业务逻辑，不跨表联动 |
