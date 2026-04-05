# 电商订单中心 · 微服务架构设计文档

> **版本**：v1.0 · **技术栈**：Go 1.21+ / MySQL 8.0 / Redis 7.0 / Kafka 3.0 / Nacos 2.0  
> **适用场景**：日均百万级订单量的电商平台  
> **阶段**：架构设计（不含代码实现）

---

## 目录

1. [服务拆分与领域建模](#一服务拆分与领域建模)
2. [技术架构与组件选型](#二技术架构与组件选型)
3. [核心业务流程设计](#三核心业务流程设计)
4. [基础设施与高可用设计](#四基础设施与高可用设计)
5. [可观测性与运维](#五可观测性与运维)

---

## 一、服务拆分与领域建模

### 1.1 核心子域划分（DDD视角）

基于领域驱动设计，将电商订单中心划分为以下子域：

| 子域类型 | 子域名称 | 核心职责 |
|---------|---------|---------|
| **核心域** | 订单域（Order Context） | 订单生命周期管理，是业务核心竞争力所在 |
| **核心域** | 支付域（Payment Context） | 支付渠道对接、支付状态流转 |
| **支撑域** | 库存域（Inventory Context） | 库存预占、实扣、释放 |
| **支撑域** | 物流域（Logistics Context） | 发货、运单、轨迹查询 |
| **通用域** | 用户域（User Context） | 用户鉴权、地址管理 |
| **通用域** | 商品域（Product Context） | 商品信息、价格快照 |
| **通用域** | 通知域（Notification Context） | 短信/Push/邮件通知 |

### 1.2 限界上下文图

```mermaid
graph TB
    subgraph 核心域
        OC["🛒 Order Context<br/>订单聚合根<br/>OrderItem / OrderStatus<br/>PriceSnapshot"]
        PC["💳 Payment Context<br/>支付单聚合根<br/>PayChannel / RefundRecord"]
    end

    subgraph 支撑域
        IC["📦 Inventory Context<br/>库存聚合根<br/>StockReservation / StockLog"]
        LC["🚚 Logistics Context<br/>运单聚合根<br/>Shipment / TrackingEvent"]
    end

    subgraph 通用域
        UC["👤 User Context<br/>用户 / 收货地址"]
        PRC["🏷️ Product Context<br/>商品 / SKU / 价格"]
        NC["🔔 Notification Context<br/>消息模板 / 发送记录"]
    end

    subgraph 基础设施
        GW["🌐 API Gateway<br/>Kong / Nginx"]
        MQ["📨 Kafka<br/>消息总线"]
        REG["🗂️ Nacos<br/>注册中心 / 配置中心"]
    end

    GW --> OC
    GW --> UC
    OC -->|"同步gRPC: 价格快照"| PRC
    OC -->|"同步gRPC: 预占库存"| IC
    OC -->|"同步gRPC: 创建支付单"| PC
    PC -->|"异步Kafka: order-paid"| OC
    OC -->|"异步Kafka: order-confirmed"| IC
    OC -->|"异步Kafka: order-confirmed"| LC
    OC -->|"异步Kafka: order-created"| NC
    PC -->|"异步Kafka: payment-success"| NC
    LC -->|"异步Kafka: shipment-updated"| NC
    OC -.->|"服务发现"| REG
    IC -.->|"服务发现"| REG
    PC -.->|"服务发现"| REG
```

### 1.3 微服务拆分与职责边界

#### order-service（订单服务）— 核心服务

- **职责**：订单创建、状态机流转（待支付→已支付→已发货→已完成→已取消）、订单查询、退款申请
- **数据所有权**：`orders` 表、`order_items` 表、`order_status_logs` 表
- **对外接口**：HTTP REST（面向前端）+ gRPC（面向内部服务）
- **不负责**：支付处理、库存计算、物流跟踪

#### inventory-service（库存服务）

- **职责**：SKU 库存管理、预占（下单锁定）、实扣（支付成功）、释放（取消/超时）
- **数据所有权**：`inventory` 表、`stock_reservations` 表、`stock_logs` 表
- **对外接口**：gRPC（仅内部调用）
- **核心约束**：库存数据最终一致性 ≤ 3s，不允许超卖

#### payment-service（支付服务）

- **职责**：支付单创建、对接第三方支付（支付宝/微信）、支付回调处理、退款
- **数据所有权**：`payment_orders` 表、`refund_records` 表
- **对外接口**：HTTP（接收第三方回调）+ gRPC（内部查询）
- **核心约束**：支付回调必须幂等处理

#### logistics-service（物流服务）

- **职责**：创建运单、对接物流供应商、轨迹同步
- **数据所有权**：`shipments` 表、`tracking_events` 表
- **触发方式**：消费 `order-confirmed` Kafka 事件，异步创建运单

#### user-service（用户服务）

- **职责**：用户信息、JWT Token 鉴权、收货地址管理
- **数据所有权**：`users` 表、`addresses` 表
- **说明**：通用域服务，由 API Gateway 直接调用进行鉴权

#### notification-service（通知服务）

- **职责**：消费业务事件，发送短信/Push/邮件通知
- **数据所有权**：`notification_logs` 表
- **说明**：纯消费者，无对外 gRPC 接口，与业务服务完全解耦

### 1.4 服务依赖拓扑图

```mermaid
graph LR
    Client(["📱 Client<br/>App / Web"])
    GW["🌐 API Gateway"]

    subgraph 同步调用链 gRPC
        OS["order-service"]
        IS["inventory-service"]
        PS["payment-service"]
        US["user-service"]
        PRS["product-service"]
    end

    subgraph 异步事件驱动 Kafka
        LS["logistics-service"]
        NS["notification-service"]
    end

    Client --> GW
    GW -->|"JWT验证"| US
    GW -->|"HTTP REST"| OS
    GW -->|"HTTP REST"| PS

    OS -->|"gRPC: LockStock"| IS
    OS -->|"gRPC: GetSKUSnapshot"| PRS
    OS -->|"gRPC: CreatePayOrder"| PS

    PS -.->|"Kafka: payment-success"| OS
    OS -.->|"Kafka: order-confirmed"| IS
    OS -.->|"Kafka: order-confirmed"| LS
    OS -.->|"Kafka: order-created"| NS
    PS -.->|"Kafka: payment-success"| NS
    LS -.->|"Kafka: shipment-updated"| NS

    style OS fill:#ff6b6b,color:#fff
    style IS fill:#4ecdc4,color:#fff
    style PS fill:#45b7d1,color:#fff
    style NS fill:#96ceb4,color:#fff
    style LS fill:#96ceb4,color:#fff
```

**循环依赖规避原则**：

- 所有同步调用方向单向流动：`order → inventory/payment/product`，下游服务不反向调用 order-service
- 跨域数据同步一律走 Kafka 异步事件，彻底解耦双向依赖
- payment-service 通过发布事件通知 order-service，而非直接 gRPC 回调

---

## 二、技术架构与组件选型

### 2.1 技术栈选型决策

#### Web 框架：选型 **Gin** ✅

| 维度 | Gin | Echo |
|------|-----|------|
| 性能 | ✅ httprouter，极低内存占用 | ✅ 相当，略逊 |
| 社区生态 | ✅ GitHub 77k+ Stars，生态最大 | 🟡 较成熟 |
| 中间件 | 🟡 需自行整合 | ✅ 内置更丰富 |
| 学习曲线 | ✅ 极平缓 | ✅ 平缓 |
| **适用场景** | **高并发 API 服务，团队规模大** | **快速原型，中间件重度依赖** |

**决策**：选 **Gin**。订单系统为高并发核心链路，Gin 的极简设计和 httprouter 路由树性能更可控；团队成员对 Gin 的熟悉程度普遍更高，降低协作摩擦。

#### ORM：选型 **GORM** ✅

| 维度 | GORM | Ent |
|------|------|-----|
| 代码生成 | ❌ 无 Schema 强约束 | ✅ Schema 即代码，类型安全 |
| 灵活性 | ✅ Raw SQL 随意穿插 | 🟡 复杂查询需适配 |
| 复杂查询 | ✅ 链式调用灵活 | 🟡 图遍历场景更佳 |
| 学习成本 | ✅ 低 | 🟡 中等 |
| **适用场景** | **关系型复杂查询、已有团队经验** | **实体关系复杂、强类型约束** |

**决策**：选 **GORM**。订单系统核心表结构相对稳定，涉及大量复杂 JOIN 查询和分片场景，GORM 的 Raw SQL 能力更灵活；分库分表场景下，GORM + ShardingSphere-Proxy 的组合更成熟。

#### 服务注册/配置中心：选型 **Nacos** ✅（题目约束已指定）

**补充说明 Trade-off**：

| 维度 | Nacos | Consul |
|------|-------|--------|
| 配置中心 | ✅ 原生支持，动态配置 | ❌ 需搭配 Vault |
| 服务发现 | ✅ AP/CP 可切换 | ✅ CP 强一致 |
| 健康检查 | ✅ 心跳/HTTP/TCP | ✅ 更丰富 |
| 生态 | ✅ 阿里系，与 Go SDK 配合好 | ✅ HashiCorp 生态 |

Nacos 在国内中间件生态（Sentinel、Dubbo）整合更顺畅，且统一提供配置中心能力，减少组件数量，符合当前技术栈约束。

#### 链路追踪：选型 **Jaeger + OpenTelemetry SDK** ✅

| 维度 | Jaeger | SkyWalking |
|------|--------|------------|
| Go 支持 | ✅ 原生 OTel SDK | 🟡 Java 优先，Go agent 较弱 |
| 数据存储 | Elasticsearch / Cassandra | ES / MySQL / TiDB |
| UI 能力 | ✅ 简洁实用 | ✅ 更丰富，拓扑图好用 |
| 协议 | OTLP 标准 | 私有协议为主 |

**决策**：选 **Jaeger**，通过 **OpenTelemetry SDK** 埋点（标准化，未来可无缝切换后端）。Go 生态对 OTel 的支持远优于 SkyWalking agent 方式。

#### 完整技术栈清单

```
Web 框架:       Gin 1.9+
gRPC:           google.golang.org/grpc
ORM:            GORM v2 + gen（代码生成）
数据库迁移:      golang-migrate/migrate
配置管理:       Nacos SDK + viper（本地降级）
日志:           uber-go/zap（结构化）
链路追踪:       OpenTelemetry Go SDK + Jaeger
熔断限流:       alibaba/sentinel-golang
消息队列:       IBM/sarama（Kafka）
缓存:           redis/go-redis/v9
本地缓存:       dgraph-io/ristretto（高性能本地缓存）
依赖注入:       google/wire（编译期注入）
错误处理:       pkg/errors（堆栈追踪）
参数校验:       go-playground/validator/v10
Mock 测试:      golang/mock + testify
```

### 2.2 标准项目目录结构

以 `order-service` 为例，所有微服务遵循统一目录规范：

```
order-service/
├── api/                        # 接口定义层（对外契约）
│   ├── proto/                  # .proto 文件（gRPC 接口定义）
│   │   └── order/v1/
│   │       └── order.proto
│   └── http/                   # OpenAPI / Swagger 定义
│       └── swagger.yaml
│
├── cmd/                        # 程序入口（每个可执行文件一个子目录）
│   └── server/
│       └── main.go             # 仅做依赖装配，禁止写业务逻辑
│
├── configs/                    # 配置文件模板（非敏感信息）
│   ├── config.yaml             # 本地开发配置
│   └── config.prod.yaml        # 生产环境配置模板（敏感值从 Nacos 注入）
│
├── internal/                   # 私有业务逻辑（禁止外部包导入）
│   ├── domain/                 # 领域层：聚合根、实体、值对象、领域事件
│   │   ├── order/
│   │   │   ├── aggregate.go    # Order 聚合根
│   │   │   ├── entity.go       # OrderItem 等实体
│   │   │   ├── value_object.go # Money、Address 值对象
│   │   │   ├── event.go        # OrderCreated、OrderPaid 领域事件
│   │   │   └── repository.go   # Repository 接口定义（依赖倒置）
│   │   └── ...
│   │
│   ├── application/            # 应用层：用例编排（Use Case），不含领域逻辑
│   │   ├── command/            # 写命令（CreateOrder、CancelOrder）
│   │   │   └── create_order.go
│   │   └── query/              # 读查询（GetOrderDetail、ListOrders）
│   │       └── get_order.go
│   │
│   ├── infrastructure/         # 基础设施层：技术实现细节
│   │   ├── persistence/        # 数据库实现（GORM）
│   │   │   ├── model/          # GORM 数据模型（PO）
│   │   │   ├── repository/     # Repository 接口实现
│   │   │   └── migration/      # SQL 迁移文件
│   │   ├── cache/              # Redis 缓存实现
│   │   ├── mq/                 # Kafka Producer / Consumer
│   │   └── rpc/                # gRPC Client（调用其他服务）
│   │       ├── inventory/
│   │       └── payment/
│   │
│   └── interfaces/             # 接口适配层（Adapter）
│       ├── http/               # HTTP Handler（Gin Router）
│       │   ├── handler/
│       │   ├── middleware/
│       │   └── dto/            # HTTP 请求/响应 DTO
│       └── grpc/               # gRPC Server Handler
│           └── handler/
│
├── pkg/                        # 可复用的公共库（允许其他服务导入）
│   ├── errors/                 # 业务错误码定义
│   ├── idgen/                  # 分布式 ID 生成（Snowflake）
│   ├── pagination/             # 分页工具
│   └── timeutil/               # 时间处理工具
│
├── scripts/                    # 构建、部署、数据库脚本
│   ├── build.sh
│   └── migrate.sh
│
├── deployments/                # K8s / Docker 部署配置
│   ├── Dockerfile
│   ├── docker-compose.yaml     # 本地联调
│   └── k8s/
│       ├── deployment.yaml
│       ├── service.yaml
│       └── configmap.yaml
│
├── test/                       # 集成测试 / E2E 测试
│   └── integration/
│
├── go.mod
├── go.sum
└── Makefile                    # 统一构建命令
```

**目录职责说明**：

- `internal/`：Go 语言原生访问控制，强制隔离。`domain` 层零依赖任何框架，`application` 层依赖 `domain` 接口而非实现，确保可测试性。
- `pkg/`：跨服务共享工具包，通过 Go Module 方式共享，避免将业务逻辑放入此目录。
- `api/`：接口契约优先（API First），`proto` 文件是 gRPC 服务的唯一真相来源，需纳入版本控制。

### 2.3 服务内部分层架构

```mermaid
graph TB
    subgraph interfaces["接口适配层 (Interfaces)"]
        H["HTTP Handler<br/>• 参数绑定与校验<br/>• DTO ↔ Command 转换<br/>• HTTP 状态码映射"]
        G["gRPC Handler<br/>• Proto Message 转换<br/>• 错误码映射"]
    end

    subgraph application["应用层 (Application)"]
        CMD["Command Handler<br/>• 创建订单<br/>• 取消订单<br/>• 确认收货"]
        QRY["Query Handler<br/>• 查询订单详情<br/>• 分页查询订单列表"]
    end

    subgraph domain["领域层 (Domain)"]
        AGG["Order 聚合根<br/>• 状态机转换逻辑<br/>• 业务规则校验<br/>• 领域事件发布"]
        REPO["Repository 接口<br/>（抽象，不含实现）"]
    end

    subgraph infrastructure["基础设施层 (Infrastructure)"]
        REPOIM["Repository 实现<br/>（GORM）"]
        CACHE["Cache 实现<br/>（Redis）"]
        MQPUB["MQ Publisher<br/>（Kafka）"]
        RPCCLI["RPC Client<br/>（gRPC Client）"]
    end

    H --> CMD
    H --> QRY
    G --> CMD
    G --> QRY
    CMD --> AGG
    CMD --> REPO
    QRY --> REPO
    QRY --> CACHE
    AGG --> REPO
    REPO -.->|"依赖倒置<br/>接口 → 实现"| REPOIM
    CMD --> MQPUB
    CMD --> RPCCLI

    style domain fill:#fff3cd,stroke:#ffc107
    style application fill:#d1ecf1,stroke:#17a2b8
    style interfaces fill:#d4edda,stroke:#28a745
    style infrastructure fill:#f8d7da,stroke:#dc3545
```

**分层约束**：

- 依赖方向严格向下：`interfaces → application → domain ← infrastructure`
- `domain` 层不得导入任何框架包（Gin/GORM/Redis），保证纯粹性与可测试性
- `application` 层通过接口依赖 `domain.Repository`，具体实现由 `wire` 在启动时注入

---

## 三、核心业务流程设计

### 3.1 下单完整时序图

```mermaid
sequenceDiagram
    autonumber
    actor User as 用户 App
    participant GW as API Gateway
    participant OS as order-service
    participant PS as product-service
    participant IS as inventory-service
    participant PMS as payment-service
    participant DB as Order DB (MySQL)
    participant RD as Redis
    participant KF as Kafka
    participant NS as notification-service

    User->>GW: POST /v1/orders (下单请求)
    GW->>GW: JWT 鉴权 + 限流检查 (Sentinel)
    GW->>OS: 转发请求

    Note over OS,PS: ① 同步：获取商品价格快照（防止下单后改价）
    OS->>PS: gRPC GetSKUSnapshot(sku_ids)
    PS-->>OS: 返回价格快照列表

    Note over OS,IS: ② 同步：库存预占（Redis Lua 原子扣减）
    OS->>IS: gRPC LockStock(sku_id, quantity)
    IS->>RD: EVAL Lua预扣减脚本
    RD-->>IS: 扣减成功 / 库存不足
    IS-->>OS: 返回预占结果 + reservation_id

    Note over OS,DB: ③ 同步：创建订单（本地事务 + 本地消息表）
    OS->>DB: BEGIN TRANSACTION
    OS->>DB: INSERT orders (status=PENDING_PAYMENT)
    OS->>DB: INSERT order_items
    OS->>DB: INSERT local_messages (order-created事件, status=PENDING)
    OS->>DB: COMMIT

    Note over OS,PMS: ④ 同步：创建支付单
    OS->>PMS: gRPC CreatePayOrder(order_id, amount)
    PMS-->>OS: 返回 pay_url / 支付参数

    OS-->>GW: 返回 order_id + pay_url
    GW-->>User: 200 OK，跳转支付页

    Note over OS,KF: ⑤ 异步：定时任务扫描本地消息表，发布事件
    OS->>KF: Produce order-created (order_id, user_id)
    KF->>NS: Consume order-created
    NS-->>User: 推送"订单创建成功"通知

    Note over User,PMS: ⑥ 用户完成支付（第三方回调）
    User->>PMS: 完成支付（支付宝/微信收银台）
    PMS->>PMS: 接收第三方回调，幂等校验
    PMS->>DB: UPDATE payment_orders SET status=PAID

    Note over PMS,OS: ⑦ 异步：支付结果通知订单服务
    PMS->>KF: Produce payment-success (order_id, pay_order_id)
    KF->>OS: Consume payment-success
    OS->>DB: UPDATE orders SET status=PAID
    OS->>DB: INSERT local_messages (order-paid事件, status=PENDING)

    Note over OS,IS: ⑧ 异步：订单确认后实扣库存
    OS->>KF: Produce order-confirmed (order_id, reservation_ids)
    KF->>IS: Consume order-confirmed
    IS->>DB: 本地事务：实扣库存 + 释放预占记录
    IS->>RD: 同步更新 Redis 实际库存

    KF->>NS: Consume payment-success
    NS-->>User: 推送"支付成功"通知
```

### 3.2 分布式事务方案

#### 方案对比：Seata AT 模式 vs Saga 模式

| 维度 | Seata AT 模式 | Saga 模式 |
|------|--------------|-----------|
| 一致性保证 | 强一致（阶段提交） | 最终一致 |
| 侵入性 | 低（代理数据源） | 中（需实现补偿逻辑） |
| 性能 | 🟡 有锁，并发受限 | ✅ 无全局锁 |
| 适用场景 | 短事务、简单场景 | 长事务、高并发 |
| 回滚复杂度 | ✅ 自动回滚 | ❌ 需手动编写补偿 |
| 日均百万量级 | ❌ 全局锁影响吞吐量 | ✅ 推荐 |

**决策**：订单场景选择 **Saga 模式 + 本地消息表**实现最终一致性。

原因：下单流程涉及 order-service、inventory-service、payment-service 三方，事务跨度长（含第三方支付回调，可能延迟数秒），Seata AT 的全局锁在百万级并发下会成为瓶颈。

#### 本地消息表 + 定时补偿方案

```mermaid
flowchart LR
    subgraph order-service
        BIZ["业务操作<br/>INSERT orders"]
        LM["本地消息表<br/>local_messages<br/>status=PENDING"]
        BIZ -->|"同一事务"| LM
    end

    subgraph 消息投递
        TASK["定时任务<br/>（每5s扫描）"]
        KF["Kafka"]
        LM -->|"扫描PENDING消息"| TASK
        TASK -->|"Produce消息"| KF
        TASK -->|"投递成功<br/>UPDATE status=DONE"| LM
    end

    subgraph 下游消费
        IS["inventory-service<br/>幂等消费"]
        NS["notification-service<br/>幂等消费"]
        KF --> IS
        KF --> NS
    end

    subgraph 补偿机制
        RETRY["超时补偿<br/>（消息PENDING超30min<br/>重新投递）"]
        LM -.->|"监控超时"| RETRY
        RETRY --> KF
    end
```

**幂等性保障**：每条消息携带全局唯一 `message_id`（Snowflake），消费者在处理前先查 Redis 是否已消费，已消费则直接返回 ACK，保证"至少一次投递 + 幂等消费 = 恰好一次效果"。

#### 超时订单自动取消

```mermaid
stateDiagram-v2
    [*] --> PENDING_PAYMENT: 下单成功
    PENDING_PAYMENT --> PAID: 支付成功
    PENDING_PAYMENT --> CANCELLED: 30分钟未支付（定时任务扫描）
    PAID --> CONFIRMED: 用户确认收货 / 超时自动确认
    PAID --> REFUNDING: 申请退款
    CONFIRMED --> [*]: 完成
    CANCELLED --> [*]: 释放库存预占
    REFUNDING --> REFUNDED: 退款成功
    REFUNDED --> [*]: 完成
```

### 3.3 库存扣减策略：防超卖机制

```mermaid
flowchart TB
    REQ["下单请求<br/>sku_id=X, qty=3"]

    subgraph Redis预扣减层
        LUA["Lua 原子脚本<br/>GET stock:X<br/>IF stock >= qty<br/>  DECRBY stock:X qty<br/>  HSET reservations order_id qty<br/>ELSE RETURN -1"]
        SUCC{"扣减成功?"}
    end

    subgraph MySQL持久化层
        DB1["stock_reservations<br/>INSERT 预占记录<br/>（异步，通过MQ）"]
        DB2["inventory<br/>UPDATE stock = stock - qty<br/>（支付成功后实扣）"]
    end

    subgraph 补偿机制
        COMP["库存补偿对账<br/>每小时对比<br/>Redis vs MySQL<br/>不一致则修正"]
    end

    REQ --> LUA
    LUA --> SUCC
    SUCC -->|"Yes: 返回reservation_id"| DB1
    SUCC -->|"No: 返回库存不足错误"| FAIL["返回下单失败"]
    DB1 -.->|"支付成功事件"| DB2
    DB2 -.->|"定时对账"| COMP
    COMP -.->|"Redis库存修正"| LUA
```

**防超卖三重保障**：

1. **Redis Lua 原子扣减**：利用 Redis 单线程特性，Lua 脚本保证"查询+扣减"原子性，防并发超卖
2. **MySQL 乐观锁兜底**：实扣时 `UPDATE inventory SET stock = stock - qty WHERE sku_id = X AND stock >= qty`，失败则触发告警
3. **定时对账**：每小时对比 Redis 与 MySQL 库存数量，差异超过阈值触发告警并修正

---

## 四、基础设施与高可用设计

### 4.1 数据库设计策略

#### 分库分表方案

| 服务 | 分片键 | 分片规则 | 分库数 | 分表数 |
|------|-------|---------|-------|-------|
| order-service | `user_id` | 哈希取模 | 4库 | 64表（每库16表） |
| inventory-service | `sku_id` | 哈希取模 | 2库 | 16表 |
| payment-service | `order_id` | 哈希取模 | 2库 | 16表 |

**选择 `user_id` 作为订单分片键的理由**：
- 80% 的查询场景为"查询我的订单"，按 user_id 分片可保证同一用户订单落在同一分片，避免跨片查询
- 订单号（`order_no`）查询通过路由表或 `order_no` 中编码 `user_id` 后4位来定位分片

**实现工具**：ShardingSphere-Proxy 5.x（透明代理模式，应用无感知）

#### 核心表索引设计

```sql
-- orders 表核心索引设计
CREATE TABLE orders_00 (          -- 按 user_id % 64 路由
    id            BIGINT PRIMARY KEY,         -- Snowflake ID
    order_no      VARCHAR(32) NOT NULL,       -- 业务订单号（编码分片信息）
    user_id       BIGINT NOT NULL,
    status        TINYINT NOT NULL,           -- 枚举：1待付款/2已付款/...
    total_amount  DECIMAL(12,2) NOT NULL,
    created_at    DATETIME(3) NOT NULL,
    updated_at    DATETIME(3) NOT NULL,
    deleted_at    DATETIME(3),                -- 软删除

    -- 索引策略
    UNIQUE KEY uk_order_no (order_no),                        -- 全局唯一，防重复下单
    KEY idx_user_status (user_id, status, created_at),       -- 覆盖索引：用户订单列表
    KEY idx_created_at (created_at)                          -- 范围查询：数据归档
) ENGINE=InnoDB;
```

#### 读写分离

- 架构：1主2从（MySQL Group Replication，半同步复制）
- 路由：GORM 配置 `dbresolver` 插件，写操作（INSERT/UPDATE/DELETE）走主库，读操作走从库
- 延迟容忍：复制延迟 ≤ 100ms；订单详情查询对强一致性有要求时（如支付后立即查询），强制走主库

#### 数据库迁移策略

- 工具：`golang-migrate/migrate`，SQL 文件纳入 Git 版本控制
- 规范：每次变更新增迁移文件（`V{version}__{description}.up.sql` / `.down.sql`），禁止修改历史文件
- 执行：CI/CD 流水线在部署前自动执行 `migrate up`，失败则阻断发布

### 4.2 缓存策略

```mermaid
flowchart TB
    REQ["查询请求<br/>GET /orders/{order_id}"]

    LC["本地缓存<br/>ristretto<br/>TTL: 30s<br/>容量: 1万条<br/>（防缓存击穿）"]

    RC["Redis 集群<br/>order:detail:{order_id}<br/>TTL: 10min<br/>序列化: Protobuf"]

    DB["MySQL<br/>（主/从）"]

    REQ --> LC
    LC -->|"命中"| RESP1["返回结果"]
    LC -->|"未命中"| RC
    RC -->|"命中"| LC2["回填本地缓存"]
    LC2 --> RESP2["返回结果"]
    RC -->|"未命中"| LOCK["分布式锁<br/>Redis SETNX<br/>防缓存击穿"]
    LOCK -->|"获取锁"| DB
    DB --> FILL["回填 Redis<br/>回填本地缓存"]
    FILL --> RESP3["返回结果"]
    LOCK -->|"未获取锁，等待重试"| RC
```

#### 缓存一致性方案

采用 **Cache-Aside + 延迟双删** 策略：

1. **写操作**：先更新 MySQL → 删除 Redis 缓存（非更新，防止并发写导致旧值覆盖）
2. **延迟二次删除**：主库写入后，延迟 500ms 再次删除 Redis 缓存（应对主从复制延迟期间从库读到旧数据的场景）
3. **TTL 兜底**：所有缓存设置 TTL（10min），即使二次删除失败，最终也会过期

#### Redis 数据结构选型

| 场景 | Key 设计 | 数据结构 | TTL |
|------|---------|---------|-----|
| 订单详情缓存 | `order:detail:{order_id}` | String（Protobuf序列化） | 10min |
| 库存预占计数 | `stock:{sku_id}` | String（整数） | 永久（业务控制） |
| 库存预占明细 | `reservations:{sku_id}` | Hash（order_id → qty） | 1h |
| 幂等去重 | `idempotent:{message_id}` | String | 24h |
| 用户购物车 | `cart:{user_id}` | Hash（sku_id → qty） | 7天 |
| 限流计数器 | `ratelimit:{user_id}:{window}` | String（incr） | 60s |

### 4.3 Kafka Topic 设计

```mermaid
graph LR
    subgraph Producers
        OS["order-service"]
        PS["payment-service"]
        IS["inventory-service"]
        LS["logistics-service"]
    end

    subgraph Kafka Topics
        T1["order-created<br/>分区: 16<br/>副本: 3<br/>保留: 7天"]
        T2["order-confirmed<br/>分区: 16<br/>副本: 3"]
        T3["order-cancelled<br/>分区: 8<br/>副本: 3"]
        T4["payment-success<br/>分区: 16<br/>副本: 3"]
        T5["payment-failed<br/>分区: 8<br/>副本: 3"]
        T6["inventory-deducted<br/>分区: 8<br/>副本: 3"]
        T7["shipment-updated<br/>分区: 8<br/>副本: 3"]
    end

    subgraph Consumers
        OS2["order-service<br/>CG: order-payment-result"]
        IS2["inventory-service<br/>CG: inventory-order-event"]
        LS2["logistics-service<br/>CG: logistics-order-event"]
        NS["notification-service<br/>CG: notification-all-events"]
    end

    OS --> T1
    OS --> T2
    OS --> T3
    PS --> T4
    PS --> T5
    IS --> T6
    LS --> T7

    T4 --> OS2
    T2 --> IS2
    T3 --> IS2
    T2 --> LS2
    T4 --> NS
    T1 --> NS
    T7 --> NS
```

**Topic 设计原则**：

- 分区数按峰值 TPS / 单消费者处理能力估算：订单峰值 TPS ≈ 5000，单 Consumer 处理 500/s，故 `order-created` 设 16 分区
- 分区键（Partition Key）：使用 `order_id` 作为分区键，保证同一订单的事件有序
- 消费者组隔离：不同业务消费者使用独立的 Consumer Group，互不影响进度
- 死信队列（DLQ）：消费失败超过 3 次后，消息转入 `{topic}.DLQ`，人工介入处理

### 4.4 API 网关策略

```mermaid
flowchart TB
    CLIENT["客户端请求"]

    subgraph Kong API Gateway
        direction TB
        SSL["TLS 终止<br/>证书统一管理"]
        AUTH["JWT 鉴权插件<br/>校验 Token 有效性<br/>注入 user_id Header"]
        RL["限流插件<br/>全局: 10000 QPS<br/>单用户: 100 QPS/分"]
        CIRCUIT["熔断插件<br/>上游错误率>50%<br/>熔断30s"]
        ROUTE["路由插件<br/>按路径转发到对应服务"]
        LOG["请求日志插件<br/>记录 trace_id / latency"]
    end

    subgraph 后端服务
        OS["order-service"]
        US["user-service"]
        PRS["product-service"]
    end

    CLIENT --> SSL --> AUTH --> RL --> CIRCUIT --> ROUTE --> LOG
    LOG --> OS
    LOG --> US
    LOG --> PRS
```

### 4.5 容错机制设计

#### Sentinel 熔断降级规则

| 接口 | 规则类型 | 阈值 | 降级策略 |
|------|---------|------|---------|
| 下单接口 | QPS 限流 | 5000 QPS | 排队等待 → 拒绝并返回"系统繁忙" |
| 库存预占（gRPC） | 慢调用比例 | RT>200ms 比例>50% | 熔断30s，返回"库存查询失败"，触发告警 |
| 商品快照（gRPC） | 异常数 | 10次/20s | 熔断，使用缓存商品信息降级 |
| 支付创建（gRPC） | 异常比例 | >30% | 熔断，返回"支付服务暂时不可用" |

#### gRPC 重试策略

- **幂等接口**（查询类）：自动重试 3 次，初始间隔 100ms，指数退避，最大间隔 1s
- **非幂等接口**（下单、扣库存）：禁止自动重试，上层通过幂等 Key 处理重复请求
- **超时配置**：gRPC Deadline 传递，下单总链路超时 3s，单个 RPC 调用超时 500ms

---

## 五、可观测性与运维

### 5.1 结构化日志规范

#### 日志格式（JSON）

```json
{
  "timestamp": "2024-01-15T10:30:00.123Z",
  "level": "INFO",
  "service": "order-service",
  "version": "v1.2.3",
  "env": "production",

  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",

  "user_id": 10086,
  "order_id": "1234567890123456789",
  "order_no": "ORD202401151030001234",

  "msg": "订单创建成功",
  "duration_ms": 45,

  "caller": "internal/application/command/create_order.go:87",
  "method": "POST /v1/orders",
  "status_code": 200
}
```

#### 日志级别规范

| 级别 | 使用场景 | 示例 |
|------|---------|------|
| **ERROR** | 影响业务的错误，需立即告警 | 数据库连接失败、支付回调处理异常 |
| **WARN** | 潜在问题，需关注 | 库存不足（业务正常）、缓存击穿 |
| **INFO** | 关键业务节点 | 订单创建、状态流转、支付成功 |
| **DEBUG** | 调试信息（生产禁用） | SQL 详情、RPC 请求体 |

#### ELK 采集方案

```mermaid
flowchart LR
    SVC["Go 服务<br/>zap 写入 stdout"]
    FK["Filebeat<br/>（DaemonSet）<br/>采集容器日志"]
    LS["Logstash<br/>解析JSON<br/>添加K8s元数据"]
    ES["Elasticsearch<br/>索引：order-logs-{YYYY.MM.DD}<br/>保留：30天热数据<br/>180天冷数据"]
    KB["Kibana<br/>日志搜索<br/>告警规则"]

    SVC -->|"JSON日志"| FK
    FK --> LS
    LS --> ES
    ES --> KB
```

### 5.2 监控指标体系

#### Prometheus 核心指标

```
# 业务指标（RED方法：Rate / Errors / Duration）
order_create_total{status="success|failed"}          # 订单创建总数
order_create_duration_seconds{quantile="0.99"}       # 下单 P99 耗时
inventory_deduct_duration_seconds{quantile="0.95"}  # 库存扣减 P95 耗时
payment_callback_total{channel="alipay|wechat"}      # 支付回调次数
payment_callback_duration_seconds                    # 支付回调处理耗时

# 系统指标（USE方法：Utilization / Saturation / Errors）
go_goroutines                                        # Goroutine 数量
go_gc_duration_seconds                               # GC 暂停时间
process_cpu_seconds_total                            # CPU 使用率
process_resident_memory_bytes                        # 内存占用

# 依赖健康指标
db_connection_pool_idle                              # 数据库连接池空闲数
redis_connected_slaves                               # Redis 从节点数
kafka_consumer_group_lag                             # 消费者积压量
```

#### Grafana 看板设计

```mermaid
graph TB
    subgraph Grafana Dashboard - 订单系统总览
        subgraph 业务大盘
            B1["📊 订单创建QPS<br/>实时折线图"]
            B2["✅ 订单成功率<br/>目标: >99.9%"]
            B3["💰 支付成功率<br/>目标: >99.5%"]
            B4["⚡ 下单P99耗时<br/>告警阈值: >500ms"]
        end

        subgraph 服务健康
            S1["🔴 各服务错误率<br/>告警: >1%"]
            S2["⏱️ gRPC P95耗时<br/>热力图"]
            S3["📦 Kafka消费积压<br/>告警: >10000条"]
            S4["🔄 熔断触发次数"]
        end

        subgraph 基础资源
            R1["💾 MySQL 慢查询数<br/>阈值: >100ms"]
            R2["🗃️ Redis 内存使用率<br/>告警: >80%"]
            R3["🖥️ Pod CPU/内存<br/>使用率"]
            R4["🌐 数据库连接池<br/>使用率"]
        end
    end
```

#### 告警规则（PagerDuty / 钉钉）

| 告警名称 | 条件 | 级别 | 响应时间 |
|---------|------|------|---------|
| 订单创建成功率下降 | 成功率 < 99% 持续2分钟 | P0 紧急 | 5分钟 |
| 下单P99耗时超标 | P99 > 1s 持续3分钟 | P1 严重 | 15分钟 |
| Kafka 消费积压 | 积压 > 50000条 持续5分钟 | P1 严重 | 15分钟 |
| 数据库连接耗尽 | 连接池使用率 > 90% | P1 严重 | 15分钟 |
| 库存扣减失败率 | 失败率 > 5% 持续1分钟 | P0 紧急 | 5分钟 |
| Redis内存告警 | 使用率 > 85% | P2 警告 | 60分钟 |

### 5.3 分布式链路追踪

#### OpenTelemetry 埋点方案

```mermaid
flowchart LR
    subgraph 埋点层（自动 + 手动）
        GIN["Gin HTTP 中间件<br/>（自动埋点）<br/>生成 Root Span"]
        GRPC["gRPC 拦截器<br/>（自动埋点）<br/>传播 TraceContext"]
        GORM_P["GORM Plugin<br/>（自动埋点）<br/>记录 SQL 语句"]
        KAFKA_P["Kafka Producer/Consumer<br/>（手动埋点）<br/>Header传播TraceID"]
        CUSTOM["业务关键节点<br/>（手动埋点）<br/>订单状态流转、库存扣减"]
    end

    subgraph 数据上报
        OC["OTel Collector<br/>（DaemonSet）<br/>采样率: 100%（错误）<br/>          1%（成功）"]
    end

    subgraph 存储与可视化
        JG["Jaeger<br/>Elasticsearch 后端<br/>数据保留: 7天"]
    end

    GIN --> OC
    GRPC --> OC
    GORM_P --> OC
    KAFKA_P --> OC
    CUSTOM --> OC
    OC --> JG
```

#### 关键链路追踪示例（下单→支付→发货）

```
TraceID: 4bf92f3577b34da6a3ce929d0e0e4736
│
├─ [HTTP] POST /v1/orders                              0ms ~ 210ms (Root Span)
│   ├─ [gRPC] product-service.GetSKUSnapshot           2ms ~ 15ms
│   ├─ [gRPC] inventory-service.LockStock              16ms ~ 45ms
│   │   └─ [Redis] EVAL lua_stock_deduct               1ms ~ 3ms
│   ├─ [DB] INSERT orders (order_shard_03)             46ms ~ 80ms
│   ├─ [DB] INSERT local_messages                      81ms ~ 90ms
│   └─ [gRPC] payment-service.CreatePayOrder           91ms ~ 200ms
│
├─ [Kafka] Consume: payment-success                    T+3s（支付完成）
│   ├─ [DB] UPDATE orders SET status=PAID
│   └─ [Kafka] Produce: order-confirmed
│
└─ [Kafka] Consume: order-confirmed (logistics)        T+3.5s
    ├─ [HTTP] 调用物流供应商API创建运单
    └─ [DB] INSERT shipments
```

**Span 标签（Tags）规范**：

每个 Span 必须携带以下业务标签，便于问题定位：

```
order.id       = "1234567890123456789"
order.no       = "ORD202401151030001234"
user.id        = "10086"
db.shard       = "order_shard_03"
mq.topic       = "payment-success"
mq.partition   = "5"
mq.offset      = "1234567"
```

---

## 附录：架构全景图

```mermaid
graph TB
    subgraph 客户端层
        APP["📱 移动端 App"]
        WEB["🌐 Web 端"]
    end

    subgraph 网关层
        CDN["CDN / WAF"]
        GW["Kong API Gateway<br/>鉴权 / 限流 / 路由 / 熔断"]
    end

    subgraph 微服务层
        OS["🛒 order-service<br/>Gin + gRPC"]
        IS["📦 inventory-service<br/>gRPC"]
        PS["💳 payment-service<br/>Gin + gRPC"]
        US["👤 user-service<br/>Gin + gRPC"]
        LS["🚚 logistics-service<br/>gRPC"]
        NS["🔔 notification-service<br/>Kafka Consumer"]
        PRS["🏷️ product-service<br/>Gin + gRPC"]
    end

    subgraph 数据层
        MySQL["🗄️ MySQL 8.0<br/>ShardingSphere分片<br/>主从复制"]
        Redis["⚡ Redis 7.0 Cluster<br/>库存预占 / 缓存 / 限流"]
        Kafka["📨 Kafka 3.0<br/>异步事件总线"]
    end

    subgraph 基础设施层
        Nacos["🗂️ Nacos<br/>注册中心 + 配置中心"]
        Sentinel["🛡️ Sentinel<br/>熔断限流"]
        K8s["☸️ Kubernetes<br/>容器编排"]
    end

    subgraph 可观测性
        Jaeger["🔍 Jaeger<br/>链路追踪"]
        Prom["📊 Prometheus + Grafana<br/>监控告警"]
        ELK["📋 ELK Stack<br/>日志采集分析"]
    end

    APP --> CDN --> GW
    WEB --> CDN
    GW --> OS & US & PRS & PS
    OS <-->|"gRPC"| IS
    OS <-->|"gRPC"| PS
    OS <-->|"gRPC"| PRS
    OS & IS & PS & LS & NS & US & PRS --> MySQL
    OS & IS & PS & US --> Redis
    OS & PS & IS & LS --> Kafka
    Kafka --> IS & LS & NS & OS
    OS & IS & PS & LS & NS --> Nacos
    OS & IS & PS --> Sentinel
    OS & IS & PS & LS & NS --> Jaeger
    OS & IS & PS & LS & NS --> Prom
    OS & IS & PS & LS & NS --> ELK
    K8s -.->|"编排"| OS & IS & PS & LS & NS & US & PRS
```

---

*文档版本：v1.0 | 最后更新：2024-01 | 维护人：架构组*
