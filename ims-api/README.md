# IMS API — Go + SQLite3 进销存管理系统

## 技术栈

| 组件 | 版本 |
|------|------|
| Go | 1.21+ |
| Gin | v1.9 |
| GORM + SQLite | gorm v1.25 |
| JWT | golang-jwt/v5 |
| Zap | v1.27 |
| Viper | v1.18 |

---

## 快速启动

```bash
# 1. 安装依赖（需要 CGO，确保系统有 gcc）
go mod tidy

# 2. 运行服务（默认端口 8080）
go run cmd/main.go

# 3. 或编译后运行
make build && ./bin/ims-api
```

> 首次启动自动创建数据库 `ims.db`，并初始化默认管理员账号。

**默认账号**
```
用户名: admin
密  码: admin123
```

---

## 目录结构

```
ims-api/
├── cmd/main.go                  # 程序入口，依赖注入
├── config/config.yaml           # 配置文件
├── internal/
│   ├── bootstrap/db.go          # 数据库初始化 & 种子数据
│   ├── middleware/middleware.go  # JWT / CORS / 日志 / Recovery
│   ├── model/model.go           # GORM 数据模型（14张表）
│   ├── dto/dto.go               # 请求/响应 DTO
│   ├── repository/repository.go # 数据访问层
│   ├── service/service.go       # 业务逻辑层（含事务）
│   ├── handler/handler.go       # HTTP Handler 层
│   └── router/router.go         # 路由注册 & RBAC
├── pkg/
│   ├── errors/errors.go         # 业务错误码
│   ├── response/response.go     # 统一响应封装
│   ├── pagination/pagination.go # 分页工具
│   └── util/util.go             # 工具函数
├── test_api.sh                  # 全流程 Shell 测试脚本
└── Makefile
```

---

## API 接口总览

### 认证
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/login` | 登录获取 Token |

### 用户管理
| 方法 | 路径 | 权限 |
|------|------|------|
| GET | `/api/v1/users` | 登录 |
| POST | `/api/v1/users` | admin |
| PUT | `/api/v1/users/:id` | admin |
| POST | `/api/v1/users/change-password` | 登录（改自己） |

### 基础档案（分类/商品/供应商/客户/仓库）
```
GET/POST   /api/v1/categories
DELETE     /api/v1/categories/:id

GET/POST   /api/v1/products
GET/PUT    /api/v1/products/:id

GET/POST   /api/v1/suppliers
GET/PUT    /api/v1/suppliers/:id

GET/POST   /api/v1/customers
GET/PUT    /api/v1/customers/:id

GET/POST   /api/v1/warehouses
GET/PUT    /api/v1/warehouses/:id
```

### 采购管理
```
GET/POST           /api/v1/purchases
GET                /api/v1/purchases/:id
PUT                /api/v1/purchases/:id/approve   # 审核
PUT                /api/v1/purchases/:id/receive   # 入库（触发库存）
PUT                /api/v1/purchases/:id/cancel    # 取消
```

### 销售管理
```
GET/POST           /api/v1/sales
GET                /api/v1/sales/:id
PUT                /api/v1/sales/:id/approve       # 审核
PUT                /api/v1/sales/:id/ship          # 出库（校验+扣减库存）
PUT                /api/v1/sales/:id/cancel        # 取消
```

### 库存管理
```
GET    /api/v1/inventory                    # 库存台账（支持低库存筛选）
GET    /api/v1/inventory/transactions       # 库存流水
POST   /api/v1/inventory/transfer          # 仓间调拨
POST   /api/v1/inventory/adjust            # 手动调整
```

### 盘点管理
```
GET/POST   /api/v1/stocktakes
GET        /api/v1/stocktakes/:id
PUT        /api/v1/stocktakes/:id/confirm  # 确认盘点（自动调整库存）
```

---

## 统一响应格式

```json
{
  "code": 0,
  "message": "ok",
  "data": { ... },
  "trace_id": "a1b2c3d4"
}
```

分页响应：
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 错误码说明

| 错误码 | 含义 |
|--------|------|
| 0 | 成功 |
| 10001 | 参数错误 |
| 10002 | 未授权 |
| 10003 | 无权限 |
| 10004 | 资源不存在 |
| 20002 | 密码错误 |
| 40001 | **库存不足**（出库/调拨时触发） |
| 50002 | 采购单状态流转错误 |
| 60002 | 销售单状态流转错误 |

---

## 角色权限

| 角色 | 权限 |
|------|------|
| `admin` | 全部操作 |
| `operator` | 创建档案/单据、入库、出库 |
| `viewer` | 只读查询 |

---

## 运行测试

```bash
# 确保服务已启动后
bash test_api.sh
```

测试覆盖：登录鉴权 → 档案CRUD → 采购完整流程 → 销售完整流程 → 库存校验 → 调拨 → 盘点 → 权限拦截，共 **30+** 个测试用例。

---

## 核心设计要点

1. **所有库存变动走事务**：入库/出库/调拨/盘点均在 `db.Transaction()` 内完成，保证原子性。
2. **库存流水完整性**：每次库存变动写 `inventory_transactions`，包含变前/变后数量，可完整审计。
3. **出库前校验库存**：`Ship()` 在事务内先查库存再判断，防止并发超卖。
4. **SQLite WAL 模式**：启动时自动设置，提升并发读性能，写操作 `SetMaxOpenConns(1)` 串行化。
5. **单向依赖**：Handler → Service → Repository → Model，禁止跨层调用。
