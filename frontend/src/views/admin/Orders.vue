<template>
  <div class="page-container">
    <div class="page-header">
      <h2>{{ $t('order.management') }}</h2>
    </div>

    <!-- 过滤器 -->
    <div class="filter-row">
      <select v-model="filters.payment_status" class="filter-select">
        <option value="">{{ $t('order.paymentStatus') }}</option>
        <option value="unpaid">{{ $t('order.unpaid') }}</option>
        <option value="cash">{{ $t('order.cash') }}</option>
        <option value="monthly">{{ $t('order.monthly') }}</option>
      </select>
      <select v-model="filters.delivery_status" class="filter-select">
        <option value="">{{ $t('order.deliveryStatus') }}</option>
        <option value="pending">{{ $t('order.deliveryPending') }}</option>
        <option value="delivering">{{ $t('order.delivering') }}</option>
        <option value="delivered">{{ $t('order.delivered') }}</option>
        <option value="cancelled">{{ $t('order.cancelled') }}</option>
      </select>
      <van-button type="primary" size="small" @click="loadOrders">{{ $t('common.query') }}</van-button>
    </div>

    <van-loading v-if="loading" size="24" vertical style="padding: 40px 0; text-align:center" />

    <!-- 订单卡片列表 -->
    <div v-else class="card-list">
      <div v-for="row in paginatedOrders" :key="row.id" class="order-card" @click="viewOrder(row)">
        <div class="ocard-header">
          <span class="ocard-no">{{ row.order_no }}</span>
          <span class="ocard-time">{{ formatDateTime(row.created_at) }}</span>
        </div>
        <div class="ocard-body">
          <div class="ocard-row">
            <span class="ocard-merchant">{{ row.merchant_name }}</span>
            <span class="ocard-amount">{{ formatUSD(row.total_usd) }}</span>
          </div>
          <div class="ocard-tags">
            <van-tag :type="getPaymentTagType(row.payment_status)" size="medium">{{ getPaymentStatusText(row.payment_status) }}</van-tag>
            <van-tag :type="getDeliveryTagType(row.delivery_status)" size="medium">{{ getDeliveryStatusText(row.delivery_status) }}</van-tag>
            <van-tag v-if="row.unpaid_days > 0" type="danger" size="medium">{{ $t('order.unpaidDaysShort', { days: row.unpaid_days }) }}</van-tag>
            <van-tag v-if="row.days_to_billing != null && row.payment_status === 'monthly'" type="warning" size="medium">{{ $t('order.daysToBillingShort', { days: row.days_to_billing }) }}</van-tag>
          </div>
        </div>
        <div class="ocard-footer" @click.stop>
          <van-button type="primary" size="small" plain @click="viewOrder(row)">{{ $t('common.detail') }}</van-button>
          <van-button type="success" size="small" plain @click="handleUpdateStatus(row)">{{ $t('common.updateStatus') }}</van-button>
        </div>
      </div>
      <van-empty v-if="orders.length === 0" :description="$t('common.noData')" />
    </div>

    <van-pagination
      v-if="orders.length > pageSize"
      v-model="currentPage"
      :total-items="orders.length"
      :items-per-page="pageSize"
      :show-page-size="5"
      style="padding: 16px 0;"
    />

    <!-- 订单详情弹窗 -->
    <van-popup v-model:show="detailVisible" position="bottom" round :style="{ height: '85vh', overflowY: 'auto' }">
      <van-nav-bar :title="$t('order.orderDetail')" :left-text="$t('common.cancel')" @click-left="detailVisible = false" />
      <div v-if="currentOrder" class="order-detail">
        <van-cell-group inset style="margin-bottom: 12px;">
          <van-cell :title="$t('order.orderNo')" :value="currentOrder.order_no" />
          <van-cell :title="$t('order.merchant')" :value="currentOrder.merchant_name" />
          <van-cell :title="$t('order.orderTime')" :value="formatDateTime(currentOrder.created_at)" />
          <van-cell :title="$t('order.deliveryAddress')" :value="currentOrder.delivery_address || '-'" />
          <van-cell :title="$t('order.paymentStatus')">
            <template #right-icon>
              <van-tag :type="getPaymentTagType(currentOrder.payment_status)">{{ getPaymentStatusText(currentOrder.payment_status) }}</van-tag>
            </template>
          </van-cell>
          <van-cell :title="$t('order.deliveryStatus')">
            <template #right-icon>
              <van-tag :type="getDeliveryTagType(currentOrder.delivery_status)">{{ getDeliveryStatusText(currentOrder.delivery_status) }}</van-tag>
            </template>
          </van-cell>
          <van-cell v-if="currentOrder.note" :title="$t('order.note')" :value="currentOrder.note" />
          <van-cell v-if="currentOrder.scheduled_at" :title="$t('cart.scheduledAt')" :value="formatDateTime(currentOrder.scheduled_at)" />
        </van-cell-group>

        <div style="padding: 0 16px 8px; font-size: 14px; font-weight: 600; color: #303133;">{{ $t('order.orderItems') }}</div>
        <van-cell-group inset style="margin-bottom: 12px;">
          <van-cell v-for="item in currentOrder.items" :key="item.id"
            :title="item.product_name"
            :label="`${formatUSD(item.unit_price_usd)} × ${item.quantity}`"
            :value="formatUSD(item.subtotal_usd)"
          />
        </van-cell-group>

        <div class="order-total">
          <span>{{ $t('order.orderTotal') }}:</span>
          <strong>{{ formatUSD(currentOrder.total_usd) }}</strong>
        </div>

        <div style="padding: 16px; display: flex; flex-direction: column; gap: 8px;">
          <van-button v-if="['delivering','delivered'].includes(currentOrder.delivery_status)" block type="primary" @click="handleCompleteOrder(currentOrder)">
            确认完成&扣货
          </van-button>
          <van-button block type="success" @click="detailVisible = false; handleUpdateStatus(currentOrder)">
            {{ $t('common.updateStatus') }}
          </van-button>
        </div>
      </div>
    </van-popup>

    <!-- 更新状态弹窗 -->
    <van-popup v-model:show="statusVisible" position="bottom" round :style="{ minHeight: '60vh' }">
      <van-nav-bar
        :title="$t('order.updateOrderStatus')"
        :left-text="$t('common.cancel')"
        :right-text="$t('common.confirm')"
        @click-left="statusVisible = false"
        @click-right="submitStatusUpdate"
      />
      <div v-if="currentOrder" style="padding: 8px 0;">
        <van-cell-group inset>
          <van-cell :title="$t('order.paymentStatus')">
            <template #value>
              <van-radio-group v-model="statusForm.payment_status" direction="horizontal">
                <van-radio name="unpaid">{{ $t('order.unpaid') }}</van-radio>
                <van-radio name="cash">{{ $t('order.cash') }}</van-radio>
                <van-radio name="monthly">{{ $t('order.monthly') }}</van-radio>
              </van-radio-group>
            </template>
          </van-cell>
          <van-cell :title="$t('order.deliveryStatus')">
            <template #value>
              <van-radio-group v-model="statusForm.delivery_status" direction="horizontal" style="flex-wrap: wrap; gap: 4px;">
                <van-radio name="pending">{{ $t('order.deliveryPending') }}</van-radio>
                <van-radio name="delivering">{{ $t('order.delivering') }}</van-radio>
                <van-radio name="delivered">{{ $t('order.delivered') }}</van-radio>
                <van-radio name="cancelled">{{ $t('order.cancelled') }}</van-radio>
              </van-radio-group>
            </template>
          </van-cell>
        </van-cell-group>
        <van-cell-group inset style="margin-top: 8px;">
          <van-field
            :label="$t('cart.scheduledAt')"
            :model-value="statusForm.scheduled_at"
            type="datetime-local"
            @update:model-value="statusForm.scheduled_at = $event"
          >
            <template #input>
              <input type="datetime-local" v-model="statusForm.scheduled_at" style="border:none;outline:none;width:100%;font-size:14px;" />
            </template>
          </van-field>
          <van-field v-model="statusForm.note" type="textarea" rows="2" :label="$t('order.note')" />
        </van-cell-group>
        <div style="padding: 16px;">
          <van-button block type="primary" :loading="submitting" @click="submitStatusUpdate">{{ $t('common.confirm') }}</van-button>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showConfirmDialog } from 'vant'
import { getOrders, updateOrder, confirmOrderComplete } from '@/api'
import { formatUSD, formatDateTime, getPaymentStatusText, getDeliveryStatusText } from '@/utils/format'

const { t } = useI18n()
const loading = ref(false)
const orders = ref([])
const detailVisible = ref(false)
const statusVisible = ref(false)
const submitting = ref(false)
const currentOrder = ref(null)

const filters = reactive({ payment_status: '', delivery_status: '' })
const currentPage = ref(1)
const pageSize = ref(20)

const paginatedOrders = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return orders.value.slice(start, start + pageSize.value)
})

const statusForm = reactive({ payment_status: 'unpaid', delivery_status: 'pending', scheduled_at: '', note: '' })

const getPaymentTagType = (status) => {
  const map = { unpaid: 'warning', cash: 'success', monthly: 'primary' }
  return map[(status || '').toLowerCase()] || 'primary'
}

const getDeliveryTagType = (status) => {
  const map = { pending: 'primary', delivering: 'warning', delivered: 'success', cancelled: 'danger' }
  return map[(status || '').toLowerCase()] || 'primary'
}

const loadOrders = async () => {
  loading.value = true
  try {
    orders.value = await getOrders(filters)
  } catch (error) {
    console.error('加载订单失败:', error)
  } finally {
    loading.value = false
  }
}

const viewOrder = (order) => {
  currentOrder.value = order
  detailVisible.value = true
}

const handleUpdateStatus = (order) => {
  currentOrder.value = order
  statusForm.payment_status = order.payment_status
  statusForm.delivery_status = order.delivery_status
  statusForm.scheduled_at = order.scheduled_at ? order.scheduled_at.slice(0, 16) : ''
  statusForm.note = order.note || ''
  statusVisible.value = true
}

const submitStatusUpdate = async () => {
  submitting.value = true
  try {
    const payload = { ...statusForm }
    if (!payload.scheduled_at) {
      payload.scheduled_at = null
    } else {
      payload.scheduled_at = new Date(payload.scheduled_at).toISOString()
    }
    await updateOrder(currentOrder.value.id, payload)
    showSuccessToast(t('order.statusUpdateSuccess'))
    statusVisible.value = false
    loadOrders()
  } catch (error) {
    console.error('更新失败:', error)
  } finally {
    submitting.value = false
  }
}

const handleCompleteOrder = async (order) => {
  try {
    await showConfirmDialog({
      title: '确认完成',
      message: `确认订单「${order.order_no}」已完成并扣减库存？此操作不可撤销。`,
    })
    await confirmOrderComplete(order.id)
    showSuccessToast('确认完成 & 扣货成功')
    detailVisible.value = false
    loadOrders()
  } catch (error) {
    // 用户取消或接口异常均不做处理
    if (error?.code !== 'cancel') {
      console.error('扣货失败:', error)
    }
  }
}

onMounted(() => { loadOrders() })
</script>

<style scoped>
.filter-row {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.filter-select {
  height: 34px;
  padding: 0 8px;
  border: 1px solid #ebedf0;
  border-radius: 4px;
  font-size: 14px;
  background: #fff;
  color: #323233;
  flex: 1;
  min-width: 0;
}

.card-list { display: flex; flex-direction: column; gap: 10px; }

.order-card {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  padding: 14px;
  cursor: pointer;
}

.order-card:active { background: #f5f7fa; }

.ocard-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; gap: 8px; overflow: hidden; }
.ocard-no { font-size: 14px; font-weight: 600; color: #303133; font-family: monospace; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }
.ocard-time { font-size: 11px; color: #909399; white-space: nowrap; }
.ocard-body { display: flex; flex-direction: column; gap: 8px; }
.ocard-row { display: flex; justify-content: space-between; align-items: center; }
.ocard-merchant { font-size: 14px; color: #303133; }
.ocard-amount { font-size: 16px; font-weight: 700; color: #1D4ED8; }
.ocard-tags { display: flex; gap: 6px; flex-wrap: wrap; }
.ocard-footer { display: flex; justify-content: flex-end; gap: 10px; margin-top: 10px; padding-top: 8px; border-top: 1px solid #f0f0f0; }

.order-detail { padding-bottom: 20px; }

.order-total {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-top: 1px solid #ebedf0;
  font-size: 16px;
}
</style>
