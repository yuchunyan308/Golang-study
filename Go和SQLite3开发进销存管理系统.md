# Go + SQLite3 进销存管理系统 API 开发指南

## 一、系统概述

进销存系统核心覆盖三大业务域：

- **进**：采购管理（供应商、采购单、入库）
- **销**：销售管理（客户、销售单、出库）
- **存**：库存管理（商品、仓库、库存变动、盘点）

---

## 二、技术选型

| 层次 | 推荐方案 | 说明 |
|------|----------|------|
| Web 框架 | **Gin** | 轻量、高性能、生态成熟 |
| ORM / DB | **GORM + go-sqlite3** | GORM 简化操作，sqlite3 零部署 |
| 认证 | **JWT (golang-jwt)** | 无状态 Token 鉴权 |
| 配置管理 | **Viper** | 支持多格式配置文件 |
| 日志 | **Zap / Logrus** | 结构化日志 |
| 参数校验 | **go-playground/validator** | Tag 式校验规则 |
| API 文档 | **Swaggo** | 自动生成 Swagger UI |

---

## 三、项目目录结构

```
ims-api/
├── cmd/
│   └── main.go                 # 程序入口
├── config/
│   └── config.yaml             # 配置文件
├── internal/
│   ├── bootstrap/              # 初始化：DB、路由、中间件
│   ├── middleware/             # JWT、日志、限流、CORS
│   ├── model/                  # GORM 数据模型
│   ├── dto/                    # 请求/响应数据传输对象
│   ├── repository/             # 数据访问层（DAL）
│   ├── service/                # 业务逻辑层
│   ├── handler/                # HTTP Handler 层
│   └── router/                 # 路由注册
├── pkg/
│   ├── response/               # 统一响应封装
│   ├── errors/                 # 业务错误码定义
│   ├── pagination/             # 分页工具
│   └── util/                   # 通用工具函数
└── docs/                       # Swagger 文档
```

> 核心原则：**Handler → Service → Repository → Model** 单向依赖，各层职责清晰。

---

## 四、数据库设计

### 核心表结构关系

```
用户/权限
  └── users, roles, permissions

基础档案
  ├── suppliers（供应商）
  ├── customers（客户）
  ├── categories（商品分类）
  ├── products（商品）
  └── warehouses（仓库）

采购模块
  ├── purchase_orders（采购单头）
  └── purchase_order_items（采购单明细）

销售模块
  ├── sale_orders（销售单头）
  └── sale_order_items（销售单明细）

库存模块
  ├── inventories（库存台账）       ← 核心表
  ├── inventory_transactions（库存流水）
  └── stocktakes（盘点单）
```

### 关键设计决策

**库存台账（inventories）** 是系统核心，记录每个商品在每个仓库的当前数量：
- `product_id + warehouse_id` 联合唯一索引
- 每次进/销/调/盘都写一条 `inventory_transactions` 流水
- 台账数量通过流水聚合或实时更新维护（推荐实时更新 + 定期对账）

**单据状态机** 每张单据都应有明确的状态流转：
```
采购单：草稿 → 已审核 → 入库中 → 已完成 → 已取消
销售单：草稿 → 已审核 → 出库中 → 已完成 → 已取消
```

---

## 五、API 模块规划

### 1. 认证模块 `/api/v1/auth`
- `POST /login` — 登录获取 Token
- `POST /logout` — 登出
- `POST /refresh` — 刷新 Token

### 2. 基础档案
- `/products` — 商品 CRUD + 搜索
- `/categories` — 分类管理
- `/suppliers` — 供应商管理
- `/customers` — 客户管理
- `/warehouses` — 仓库管理

### 3. 采购模块 `/purchase`
- 创建/编辑草稿采购单
- 审核采购单（改变状态）
- **入库确认**（核心：触发库存增加 + 写流水）
- 采购单列表/详情

### 4. 销售模块 `/sales`
- 创建/编辑销售单
- 审核销售单
- **出库确认**（核心：触发库存减少 + 写流水，需检查库存是否充足）
- 销售单列表/详情

### 5. 库存模块 `/inventory`
- 查询各仓库库存现量
- 库存流水查询（按商品/时间/类型筛选）
- 库存调拨（A仓 → B仓，原子操作）
- 库存盘点（创建盘点单 → 录入实际数量 → 生成差异 → 确认调整）
- 库存预警（低于安全库存的商品列表）

### 6. 报表模块 `/reports`
- 采购统计（按时间/供应商/商品）
- 销售统计（按时间/客户/商品）
- 库存周转率分析
- 出入库汇总报表

---

## 六、核心业务逻辑要点

### 库存操作必须使用数据库事务

入库、出库、调拨任何一步失败，整个操作回滚，防止数据不一致：

```
BEGIN TRANSACTION
  1. 更新采购单状态 → "已入库"
  2. 更新 inventories 数量 += 入库数量
  3. 插入 inventory_transactions 流水记录
COMMIT / ROLLBACK
```

### 出库前库存校验

销售出库时必须先校验库存充足，且要考虑**并发场景**（SQLite 使用 WAL 模式 + 行级锁或乐观锁）：
```
SELECT quantity FROM inventories WHERE product_id=? AND warehouse_id=? FOR UPDATE
IF quantity < 出库数量 → 返回库存不足错误
```

### 库存流水类型枚举
```
purchase_in    采购入库
sale_out       销售出库
transfer_in    调拨入
transfer_out   调拨出
stocktake_adj  盘点调整
manual_adj     手动调整
```

---

## 七、中间件层设计

| 中间件 | 作用 |
|--------|------|
| JWT Auth | 验证 Token，注入用户上下文 |
| RBAC 权限 | 基于角色控制接口访问 |
| 请求日志 | 记录请求/响应/耗时 |
| Panic Recovery | 捕获 panic，返回 500 |
| CORS | 跨域处理 |
| 请求 ID | 每个请求注入唯一 TraceID |

---

## 八、统一响应规范

所有接口返回统一结构：
```json
{
  "code": 0,          // 0=成功，非0=业务错误码
  "message": "ok",
  "data": { ... },
  "trace_id": "xxx"   // 便于日志追踪
}
```

分页响应：
```json
{
  "code": 0,
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

---

## 九、SQLite 生产注意事项

1. **开启 WAL 模式**：`PRAGMA journal_mode=WAL` — 提升并发读性能
2. **开启外键约束**：`PRAGMA foreign_keys=ON` — SQLite 默认关闭
3. **连接池配置**：SQLite 写操作串行，`SetMaxOpenConns(1)` 防止写冲突
4. **定期 VACUUM**：防止数据库文件膨胀
5. **备份策略**：利用 SQLite 的 `.backup` API 做热备份

---

## 十、开发推进路线

```
阶段一（基础骨架）
  ✦ 项目初始化、目录结构、配置加载
  ✦ 数据库连接、GORM AutoMigrate
  ✦ 统一响应、错误码、日志中间件
  ✦ JWT 认证 + 用户登录接口

阶段二（基础档案）
  ✦ 商品、分类、供应商、客户、仓库 CRUD

阶段三（核心业务）
  ✦ 采购单完整流程（含事务入库）
  ✦ 销售单完整流程（含库存校验出库）
  ✦ 库存查询与流水

阶段四（高级功能）
  ✦ 库存调拨、盘点
  ✦ 报表统计接口
  ✦ RBAC 权限控制

阶段五（工程完善）
  ✦ Swagger 文档
  ✦ 单元测试（service 层重点覆盖）
  ✦ Docker 打包部署
```

---

## 总结

整个系统的**设计核心**只有一句话：**所有库存变动必须经过流水表，绝不直接改库存数字**。做到这一点，数据可审计、可回溯、可对账，系统就有了坚实的业务基础。
