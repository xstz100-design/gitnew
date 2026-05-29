<template>
  <div class="page-container">
    <div class="page-header">
      <h2>{{ $t('admin.dashboard') }}</h2>
      <van-button size="small" icon="replay" @click="loadStats">{{ $t('admin.refreshData') }}</van-button>
    </div>

    <!-- 库存预警横条 -->
    <div v-if="stats.lowStockProducts > 0" class="alert-bar alert-danger" @click="stockFilter = 'low'">
      <van-icon name="warning-o" class="alert-icon" />
      <span class="alert-text">{{ $t('admin.lowStockAlert', { count: stats.lowStockProducts }) }}</span>
    </div>

    <!-- 统计卡片 -->
    <div class="stat-grid mb-lg">
      <div class="stat-card clickable" :class="{ active: stockFilter === 'low' }" @click="stockFilter = stockFilter === 'low' ? '' : 'low'">
        <div class="stat-icon" style="background:#DC2626"><van-icon name="warning-o" size="20" /></div>
        <div class="stat-content">
          <div class="stat-value" style="color:#DC2626">{{ stats.lowStockProducts }}</div>
          <div class="stat-label">{{ $t('admin.stockUrgent') }}</div>
        </div>
      </div>
      <div class="stat-card clickable" :class="{ active: stockFilter === 'slow' }" @click="stockFilter = stockFilter === 'slow' ? '' : 'slow'">
        <div class="stat-icon" style="background:#EA580C"><van-icon name="clock-o" size="20" /></div>
        <div class="stat-content">
          <div class="stat-value" style="color:#EA580C">{{ stats.slowMoving }}</div>
          <div class="stat-label">{{ $t('admin.slowMoving') }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon" style="background:#16A34A"><van-icon name="bag-o" size="20" /></div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.totalProducts }}</div>
          <div class="stat-label">{{ $t('admin.totalProducts') }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon" style="background:#0891B2"><van-icon name="orders-o" size="20" /></div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.pendingOrders }}</div>
          <div class="stat-label">{{ $t('admin.pendingOrders') }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon" style="background:#0F766E"><van-icon name="manager-o" size="20" /></div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.activeUsers }}</div>
          <div class="stat-label">{{ $t('admin.activeUsers') }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon" style="background:#2563EB"><van-icon name="eye-o" size="20" /></div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.pageViews }}</div>
          <div class="stat-label">{{ $t('admin.pageViews') }}</div>
        </div>
      </div>
    </div>

    <!-- 7 日 PV/AU 趋势 -->
    <div class="content-card mb-lg">
      <div class="table-header">
        <span class="table-title">{{ $t('admin.pvAuTrend') }}</span>
        <span class="table-subtitle">{{ $t('admin.last7Days') }}</span>
      </div>
      <div class="dash-list">
        <div class="dash-item dash-header">
          <span class="dash-item-name" style="font-weight:600">{{ $t('common.date') }}</span>
          <span class="dash-item-val" style="font-weight:600">PV</span>
          <span class="dash-item-val" style="font-weight:600">AU</span>
        </div>
        <div v-for="m in metricsData" :key="m.date" class="dash-item">
          <span class="dash-item-name">{{ m.date.slice(5) }}</span>
          <span class="dash-item-val">{{ m.page_views }}</span>
          <span class="dash-item-val" style="color:#0891B2">{{ m.active_users }}</span>
        </div>
        <div v-if="metricsData.length === 0" class="dash-empty">暂无数据</div>
      </div>
    </div>

    <div class="two-col-grid">
      <div class="content-card">
        <div class="table-header">
          <span class="table-title">
            {{ stockFilter === 'slow' ? $t('admin.slowMovingTitle') : $t('admin.lowStockTitle') }}
          </span>
          <span class="link-btn" @click="$router.push('/admin/products')">{{ $t('common.viewAll') }} →</span>
        </div>
        <div class="dash-list">
          <div v-for="row in displayProducts" :key="row.id" class="dash-item">
            <div class="dash-item-left">
              <span class="stock-dot" :class="getStockColor(row)"></span>
              <span class="dash-item-name">{{ row.name }}</span>
            </div>
            <span class="dash-item-val" :class="getStockColor(row)">{{ row.stock }}<span class="dash-item-warn"> / {{ row.stock_warning }}</span></span>
          </div>
          <div v-if="displayProducts.length === 0" class="dash-empty">—</div>
        </div>
      </div>

      <div class="content-card">
        <div class="table-header">
          <span class="table-title">{{ $t('admin.pendingOrdersTitle') }}</span>
          <span class="link-btn" @click="$router.push('/admin/orders')">{{ $t('common.viewAll') }} →</span>
        </div>
        <div class="dash-list">
          <div v-for="row in pendingOrders" :key="row.id" class="dash-item">
            <span class="dash-item-name">{{ row.merchant_name }}</span>
            <span class="dash-item-val amount">{{ formatUSD(row.total_usd) }}</span>
          </div>
          <div v-if="pendingOrders.length === 0" class="dash-empty">—</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { getProducts, getOrders, getUserList, getDashboardMetrics } from '@/api'
import { formatUSD } from '@/utils/format'

const stockFilter = ref('')
const stats = ref({ totalProducts: 0, lowStockProducts: 0, slowMoving: 0, pendingOrders: 0, totalMerchants: 0, activeUsers: 0, pageViews: 0 })
const allProducts = ref([])
const lowStockProducts = ref([])
const slowMovingProducts = ref([])
const pendingOrders = ref([])
const metricsData = ref([])

const displayProducts = computed(() => {
  if (stockFilter.value === 'slow') return slowMovingProducts.value
  return lowStockProducts.value
})

const getStockColor = (product) => {
  if (product.stock <= product.stock_warning) return 'text-danger'
  return 'text-normal'
}

const loadStats = async () => {
  try {
    const products = await getProducts()
    allProducts.value = products
    stats.value.totalProducts = products.length
    lowStockProducts.value = products.filter((p) => p.is_low_stock).slice(0, 8)
    stats.value.lowStockProducts = products.filter((p) => p.is_low_stock).length

    const now = new Date()
    const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
    const slowItems = products.filter(p => {
      const updatedAt = new Date(p.updated_at)
      return p.stock > p.stock_warning && p.is_active && updatedAt < thirtyDaysAgo
    })
    slowMovingProducts.value = slowItems.slice(0, 8)
    stats.value.slowMoving = slowItems.length

    const orders = await getOrders({ delivery_status: 'pending' })
    stats.value.pendingOrders = orders.length
    pendingOrders.value = orders.slice(0, 5)

    const merchants = await getUserList('merchant')
    stats.value.totalMerchants = merchants.length

    const metric = await getDashboardMetrics(7)
    const metrics = metric.metrics || []
    metricsData.value = [...metrics].reverse()
    stats.value.activeUsers = metrics.reduce((sum, m) => sum + (m.active_users || 0), 0)
    stats.value.pageViews = metrics.reduce((sum, m) => sum + (m.page_views || 0), 0)
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

onMounted(() => { loadStats() })
</script>

<style scoped>
.stat-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

@media (max-width: 767px) {
  .stat-grid { grid-template-columns: repeat(2, 1fr); gap: 8px; }
}

.stat-card.clickable { cursor: pointer; border: 2px solid transparent; }
.stat-card.clickable:hover { transform: translateY(-2px); box-shadow: 0 6px 16px rgba(0,0,0,0.12); }
.stat-card.clickable.active { border-color: #409eff; }

.two-col-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

@media (max-width: 767px) {
  .two-col-grid { grid-template-columns: 1fr; gap: 12px; }
}

.dash-list { display: flex; flex-direction: column; }
.dash-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #f0f0f0;
  font-size: 13px;
}
.dash-item:last-child { border-bottom: none; }
.dash-item-left { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 0; }
.dash-item-name { color: #303133; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; margin-right: 12px; }
.dash-item-val { font-weight: 600; white-space: nowrap; min-width: 44px; text-align: right; }
.dash-item-warn { font-weight: 400; color: #909399; font-size: 12px; }
.dash-empty { padding: 16px 0; text-align: center; color: #c0c4cc; }
.dash-header { border-bottom: 2px solid #e6e8eb; padding-bottom: 6px; }
.table-subtitle { font-size: 12px; color: #909399; margin-left: 8px; }
.link-btn { font-size: 13px; color: #1d4ed8; cursor: pointer; }
.link-btn:hover { text-decoration: underline; }

.stock-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.stock-dot.text-danger { background: #DC2626; }
.stock-dot.text-normal { background: #16A34A; }
.text-danger { color: #DC2626; }
.text-normal { color: #16A34A; }
.amount { color: #1D4ED8; }
</style>
