<template>
  <div class="page-container">
    <div class="page-header">
      <h2>{{ $t('picker.title') }}</h2>
    </div>

    <van-tabs v-model:active="filterStatus" type="card" style="margin-bottom: 12px;" @change="() => {}">
      <van-tab :title="$t('picker.pending')" name="pending" />
      <van-tab :title="$t('picker.picked')" name="picked" />
      <van-tab :title="$t('picker.all')" name="all" />
    </van-tabs>

    <div style="margin-bottom: 12px;">
      <van-button type="primary" size="small" plain icon="replay" @click="loadOrders">{{ $t('common.refresh') }}</van-button>
    </div>

    <van-loading v-if="loading" size="24" vertical style="padding: 40px 0; text-align:center" />

    <div v-else class="card-list">
      <div v-for="row in filteredOrders" :key="row.id" class="pick-card" @click="openItems(row)">
        <div class="pick-card-header">
          <span class="pick-order-no">{{ row.order_no }}</span>
          <van-tag :type="row.picked_at ? 'success' : 'warning'" size="medium">
            {{ row.picked_at ? $t('picker.picked') : $t('picker.pending') }}
          </van-tag>
        </div>
        <div class="pick-card-body">
          <span class="pick-meta">{{ formatDateTime(row.created_at) }}</span>
          <span class="pick-items-count">{{ $t('picker.items') }}: {{ countItems(row) }}</span>
        </div>
        <div v-if="row.picked_at" class="pick-card-time">
          {{ $t('picker.pickedAt') }}: {{ formatDateTime(row.picked_at) }}
        </div>
        <div class="pick-card-footer" @click.stop>
          <van-button type="primary" size="small" plain @click="openItems(row)">{{ $t('picker.confirmPick') }}</van-button>
        </div>
      </div>
      <van-empty v-if="filteredOrders.length === 0" :description="$t('picker.noOrders')" />
    </div>

    <!-- 配货清单弹窗 -->
    <van-popup v-model:show="dialogVisible" position="bottom" round :style="{ height: '85vh', overflowY: 'auto' }">
      <van-nav-bar
        :title="dialogTitle"
        :left-text="$t('common.cancel')"
        @click-left="dialogVisible = false"
      />
      <van-loading v-if="dialogLoading" size="24" vertical style="padding: 40px 0; text-align:center" />
      <div v-else-if="currentDetail" style="padding: 12px 16px;">
        <div class="picker-list">
          <div v-for="(it, idx) in currentDetail.items" :key="idx" class="picker-item">
            <img v-if="it.image" :src="it.image" class="picker-img" />
            <div v-else class="picker-img picker-img-empty" />
            <div class="picker-info">
              <div class="picker-name">{{ it.name }}</div>
              <div v-if="it.barcode" class="picker-barcode">{{ it.barcode }}</div>
              <div v-if="it.specs" class="picker-specs">{{ it.specs }}</div>
            </div>
            <div class="picker-qty">
              × <b>{{ it.quantity }}</b>
              <span v-if="it.unit" class="picker-unit">{{ it.unit }}</span>
            </div>
          </div>
        </div>
        <div class="picker-total">
          {{ $t('picker.totalItems', { n: currentDetail.items.length, q: totalQty }) }}
        </div>
        <div v-if="!currentDetail.picked_at" style="padding-top: 16px;">
          <van-button block type="primary" :loading="confirming" @click="confirmPick">{{ $t('picker.confirmPick') }}</van-button>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { showSuccessToast, showFailToast, showConfirmDialog } from 'vant'
import { useI18n } from 'vue-i18n'
import { getOrders, getPickerItems, markOrderPicked } from '@/api'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()

const loading = ref(false)
const orders = ref([])
const filterStatus = ref('pending')

const dialogVisible = ref(false)
const dialogLoading = ref(false)
const confirming = ref(false)
const currentDetail = ref(null)
const currentOrderId = ref(null)

const filteredOrders = computed(() => {
  if (filterStatus.value === 'pending') return orders.value.filter(o => !o.picked_at && o.delivery_status !== 'cancelled')
  if (filterStatus.value === 'picked') return orders.value.filter(o => !!o.picked_at)
  return orders.value
})

const dialogTitle = computed(() => {
  if (!currentDetail.value) return t('picker.title')
  return `${t('picker.orderNo')}: ${currentDetail.value.order_no}`
})

const totalQty = computed(() => {
  if (!currentDetail.value) return 0
  return currentDetail.value.items.reduce((s, it) => s + (it.quantity || 0), 0)
})

function countItems(row) {
  return (row.items || []).length
}

async function loadOrders() {
  loading.value = true
  try {
    const data = await getOrders({})
    orders.value = data || []
  } catch (e) {
    showFailToast(e?.message || 'Load failed')
  } finally {
    loading.value = false
  }
}

async function openItems(row) {
  currentOrderId.value = row.id
  dialogVisible.value = true
  dialogLoading.value = true
  currentDetail.value = null
  try {
    currentDetail.value = await getPickerItems(row.id)
  } catch (e) {
    showFailToast(e?.message || 'Load failed')
    dialogVisible.value = false
  } finally {
    dialogLoading.value = false
  }
}

async function confirmPick() {
  try {
    await showConfirmDialog({
      title: t('picker.confirmPick'),
      message: t('picker.confirmPickTip'),
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
    })
  } catch { return }
  confirming.value = true
  try {
    await markOrderPicked(currentOrderId.value)
    showSuccessToast(t('picker.pickedSuccess'))
    dialogVisible.value = false
    await loadOrders()
  } catch (e) {
    showFailToast(e?.message || 'Failed')
  } finally {
    confirming.value = false
  }
}

onMounted(loadOrders)
</script>

<style scoped>
.card-list { display: flex; flex-direction: column; gap: 10px; }

.pick-card {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  padding: 14px;
  cursor: pointer;
}
.pick-card:active { background: #f5f7fa; }

.pick-card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.pick-order-no { font-size: 14px; font-weight: 600; color: #303133; font-family: monospace; }
.pick-card-body { display: flex; justify-content: space-between; font-size: 13px; color: #909399; margin-bottom: 6px; }
.pick-card-time { font-size: 12px; color: #67c23a; margin-bottom: 6px; }

.pick-card-footer {
  display: flex;
  justify-content: flex-end;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
}

.picker-list { display: flex; flex-direction: column; gap: 12px; }

.picker-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px;
  border: 1px solid #eee;
  border-radius: 8px;
}

.picker-img {
  width: 64px;
  height: 64px;
  border-radius: 6px;
  flex-shrink: 0;
  object-fit: cover;
}
.picker-img-empty {
  background: #f0f0f0;
}

.picker-info { flex: 1; min-width: 0; }
.picker-name { font-weight: 600; font-size: 15px; color: #2b2b2b; margin-bottom: 4px; }
.picker-barcode { font-family: monospace; font-size: 13px; color: #666; }
.picker-specs { font-size: 12px; color: #999; }

.picker-qty { font-size: 18px; color: #d44e4e; white-space: nowrap; }
.picker-unit { margin-left: 4px; font-size: 12px; color: #999; }

.picker-total {
  text-align: right;
  padding: 8px 4px;
  color: #666;
  border-top: 1px dashed #ddd;
  margin-top: 8px;
}
</style>
