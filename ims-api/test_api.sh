#!/usr/bin/env bash
# ==========================================================
#  IMS API 完整功能测试脚本
#  用法：确保服务已运行后执行 bash test_api.sh
# ==========================================================

BASE="http://localhost:8080/api/v1"
PASS=0
FAIL=0
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok()   { echo -e "${GREEN}[PASS]${NC} $1"; ((PASS++)); }
fail() { echo -e "${RED}[FAIL]${NC} $1"; ((FAIL++)); }
info() { echo -e "${YELLOW}[INFO]${NC} $1"; }

check() {
  local desc="$1"
  local code="$2"
  local expected="${3:-0}"
  if [ "$code" = "$expected" ]; then
    ok "$desc"
  else
    fail "$desc (got code=$code, expected=$expected)"
  fi
}

# ---------- 健康检查 ----------
info "=== 健康检查 ==="
RES=$(curl -s "$BASE/../ping")
echo "$RES" | grep -q "pong" && ok "GET /ping" || fail "GET /ping"

# ---------- 认证 ----------
info "=== 认证模块 ==="
RES=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
TOKEN=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['token'])" 2>/dev/null)
check "POST /auth/login (admin)" "$CODE" "0"

# 错误密码
RES=$(curl -s -X POST "$BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"wrong"}')
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "POST /auth/login (wrong pwd)" "$CODE" "20002"

AUTH="Authorization: Bearer $TOKEN"

# ---------- 用户管理 ----------
info "=== 用户管理 ==="
RES=$(curl -s "$BASE/users" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "GET /users" "$CODE" "0"

RES=$(curl -s -X POST "$BASE/users" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"username":"operator1","password":"pass1234","real_name":"操作员1","role":"operator"}')
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "POST /users (创建操作员)" "$CODE" "0"

# ---------- 商品分类 ----------
info "=== 分类管理 ==="
RES=$(curl -s -X POST "$BASE/categories" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"name":"电子产品","remark":"手机、电脑等"}')
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "POST /categories" "$CODE" "0"

RES=$(curl -s "$BASE/categories" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
CAT_ID=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'][-1]['id'])" 2>/dev/null)
check "GET /categories" "$CODE" "0"

# ---------- 商品 ----------
info "=== 商品管理 ==="
RES=$(curl -s -X POST "$BASE/products" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d "{\"code\":\"P001\",\"name\":\"iPhone 15\",\"category_id\":$CAT_ID,\"unit\":\"台\",\"cost_price\":5000,\"sale_price\":7999,\"min_stock\":10}")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "POST /products" "$CODE" "0"

RES=$(curl -s "$BASE/products" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
PROD_ID=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['list'][-1]['id'])" 2>/dev/null)
check "GET /products" "$CODE" "0"

RES=$(curl -s "$BASE/products/$PROD_ID" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "GET /products/:id" "$CODE" "0"

RES=$(curl -s -X PUT "$BASE/products/$PROD_ID" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"sale_price":8499}')
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "PUT /products/:id" "$CODE" "0"

# ---------- 供应商 ----------
info "=== 供应商管理 ==="
RES=$(curl -s -X POST "$BASE/suppliers" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"code":"S001","name":"苹果供应商","contact":"张三","phone":"13800138000","address":"北京市"}')
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "POST /suppliers" "$CODE" "0"

RES=$(curl -s "$BASE/suppliers" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
SUP_ID=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['list'][-1]['id'])" 2>/dev/null)
check "GET /suppliers" "$CODE" "0"

# ---------- 客户 ----------
info "=== 客户管理 ==="
RES=$(curl -s -X POST "$BASE/customers" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"code":"C001","name":"小米公司","contact":"李四","phone":"13900139000","address":"上海市"}')
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "POST /customers" "$CODE" "0"

RES=$(curl -s "$BASE/customers" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
CUS_ID=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['list'][-1]['id'])" 2>/dev/null)
check "GET /customers" "$CODE" "0"

# ---------- 仓库 ----------
info "=== 仓库管理 ==="
RES=$(curl -s "$BASE/warehouses" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
WH_ID=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'][0]['id'])" 2>/dev/null)
check "GET /warehouses" "$CODE" "0"

RES=$(curl -s -X POST "$BASE/warehouses" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d '{"code":"WH002","name":"上海分仓","address":"上海市浦东新区","keeper":"王五"}')
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "POST /warehouses" "$CODE" "0"

# ---------- 采购流程 ----------
info "=== 采购完整流程：创建→审核→入库 ==="
RES=$(curl -s -X POST "$BASE/purchases" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d "{
    \"supplier_id\": $SUP_ID,
    \"warehouse_id\": $WH_ID,
    \"remark\": \"首批采购\",
    \"items\": [{
      \"product_id\": $PROD_ID,
      \"quantity\": 100,
      \"price\": 5000
    }]
  }")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
PO_ID=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['id'])" 2>/dev/null)
check "POST /purchases (创建采购单)" "$CODE" "0"

RES=$(curl -s "$BASE/purchases/$PO_ID" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "GET /purchases/:id" "$CODE" "0"

RES=$(curl -s -X PUT "$BASE/purchases/$PO_ID/approve" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "PUT /purchases/:id/approve (审核)" "$CODE" "0"

# 重复审核应失败
RES=$(curl -s -X PUT "$BASE/purchases/$PO_ID/approve" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "PUT /purchases/:id/approve (重复审核应拒绝)" "$CODE" "50002"

RES=$(curl -s -X PUT "$BASE/purchases/$PO_ID/receive" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "PUT /purchases/:id/receive (入库)" "$CODE" "0"

# ---------- 查看库存（入库后应有100台） ----------
info "=== 库存查询 ==="
RES=$(curl -s "$BASE/inventory?product_id=$PROD_ID&warehouse_id=$WH_ID" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
QTY=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['list'][0]['quantity'])" 2>/dev/null)
check "GET /inventory (库存应为100)" "$CODE" "0"
[ "$QTY" = "100" ] && ok "库存数量验证: quantity=$QTY" || fail "库存数量不对: quantity=$QTY (expected 100)"

# ---------- 库存流水 ----------
RES=$(curl -s "$BASE/inventory/transactions?product_id=$PROD_ID" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "GET /inventory/transactions" "$CODE" "0"

# ---------- 销售流程 ----------
info "=== 销售完整流程：创建→审核→出库 ==="
RES=$(curl -s -X POST "$BASE/sales" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d "{
    \"customer_id\": $CUS_ID,
    \"warehouse_id\": $WH_ID,
    \"remark\": \"首笔销售\",
    \"items\": [{
      \"product_id\": $PROD_ID,
      \"quantity\": 30,
      \"price\": 7999
    }]
  }")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
SO_ID=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['id'])" 2>/dev/null)
check "POST /sales (创建销售单)" "$CODE" "0"

RES=$(curl -s -X PUT "$BASE/sales/$SO_ID/approve" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "PUT /sales/:id/approve (审核)" "$CODE" "0"

RES=$(curl -s -X PUT "$BASE/sales/$SO_ID/ship" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "PUT /sales/:id/ship (出库)" "$CODE" "0"

# 出库后库存应为70
RES=$(curl -s "$BASE/inventory?product_id=$PROD_ID&warehouse_id=$WH_ID" -H "$AUTH")
QTY=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['list'][0]['quantity'])" 2>/dev/null)
[ "$QTY" = "70" ] && ok "出库后库存验证: quantity=$QTY" || fail "出库后库存不对: quantity=$QTY (expected 70)"

# ---------- 库存不足拦截 ----------
info "=== 库存不足拦截 ==="
RES=$(curl -s -X POST "$BASE/sales" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d "{
    \"customer_id\": $CUS_ID,
    \"warehouse_id\": $WH_ID,
    \"remark\": \"超量销售测试\",
    \"items\": [{
      \"product_id\": $PROD_ID,
      \"quantity\": 9999,
      \"price\": 7999
    }]
  }")
SO2_ID=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['id'])" 2>/dev/null)
curl -s -X PUT "$BASE/sales/$SO2_ID/approve" -H "$AUTH" > /dev/null
RES=$(curl -s -X PUT "$BASE/sales/$SO2_ID/ship" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "库存不足时出库应拒绝(40001)" "$CODE" "40001"

# ---------- 调拨 ----------
info "=== 库存调拨 ==="
RES=$(curl -s "$BASE/warehouses" -H "$AUTH")
WH2_ID=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data'][-1]['id'])" 2>/dev/null)

RES=$(curl -s -X POST "$BASE/inventory/transfer" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d "{
    \"product_id\": $PROD_ID,
    \"from_warehouse_id\": $WH_ID,
    \"to_warehouse_id\": $WH2_ID,
    \"quantity\": 20,
    \"remark\": \"调拨至上海分仓\"
  }")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "POST /inventory/transfer (调拨20台)" "$CODE" "0"

# 调拨后主仓应为50
RES=$(curl -s "$BASE/inventory?product_id=$PROD_ID&warehouse_id=$WH_ID" -H "$AUTH")
QTY=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['list'][0]['quantity'])" 2>/dev/null)
[ "$QTY" = "50" ] && ok "调拨后主仓库存: quantity=$QTY" || fail "调拨后主仓库存不对: quantity=$QTY (expected 50)"

# ---------- 手动调整 ----------
info "=== 手动库存调整 ==="
RES=$(curl -s -X POST "$BASE/inventory/adjust" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d "{\"product_id\":$PROD_ID,\"warehouse_id\":$WH_ID,\"quantity\":-5,\"remark\":\"损耗调整\"}")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "POST /inventory/adjust (扣减5台)" "$CODE" "0"

# ---------- 盘点流程 ----------
info "=== 盘点完整流程：创建→确认 ==="
RES=$(curl -s -X POST "$BASE/stocktakes" -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d "{
    \"warehouse_id\": $WH_ID,
    \"remark\": \"月末盘点\",
    \"items\": [{
      \"product_id\": $PROD_ID,
      \"actual_qty\": 48
    }]
  }")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
ST_ID=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['id'])" 2>/dev/null)
check "POST /stocktakes (创建盘点单)" "$CODE" "0"

RES=$(curl -s "$BASE/stocktakes/$ST_ID" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
DIFF=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['items'][0]['diff_qty'])" 2>/dev/null)
check "GET /stocktakes/:id" "$CODE" "0"
info "盘点差异: diff_qty=$DIFF (系统45→实盘48=+3)"

RES=$(curl -s -X PUT "$BASE/stocktakes/$ST_ID/confirm" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "PUT /stocktakes/:id/confirm (确认盘点)" "$CODE" "0"

# ---------- 低库存预警 ----------
info "=== 低库存预警 ==="
RES=$(curl -s "$BASE/inventory?low_stock=true" -H "$AUTH")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "GET /inventory?low_stock=true" "$CODE" "0"

# ---------- 权限检查 ----------
info "=== 权限验证 ==="
RES=$(curl -s "$BASE/warehouses" -H "Authorization: Bearer invalid_token")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "无效Token应拒绝(20004)" "$CODE" "20004"

RES=$(curl -s "$BASE/users")
CODE=$(echo "$RES" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['code'])" 2>/dev/null)
check "无Token应拒绝(10002)" "$CODE" "10002"

# ---------- 汇总 ----------
echo ""
echo "========================================"
echo -e "  测试完成：${GREEN}PASS=$PASS${NC}  ${RED}FAIL=$FAIL${NC}"
echo "========================================"
[ $FAIL -eq 0 ] && exit 0 || exit 1
