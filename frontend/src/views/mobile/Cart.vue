<template>
  <div class="mobile-cart">
    <van-nav-bar left-arrow fixed placeholder @click-left="handleBack">
      <template #title>
        <img src="/images/logo2.svg" alt="logo" class="cart-logo" />
      </template>
      <template v-if="cartStore.items.length > 0" #right>
        <span class="clear-cart-btn" @click="clearCart">{{ $t('cart.clearCart') }}</span>
      </template>
    </van-nav-bar>
    
    <van-empty v-if="cartStore.items.length === 0" :description="$t('cart.empty')" class="cart-empty">
      <template #image>
        <div class="empty-cart-icon">
          <van-icon name="shopping-cart-o" size="64" color="#d9d9d9" />
        </div>
      </template>
      <van-button type="primary" round @click="$router.push('/m/shop')">
        {{ $t('cart.goShopping') }}
      </van-button>
    </van-empty>
    
    <template v-else>
      <!-- 商铺信息 -->
      <van-cell-group inset>
        <van-cell :title="$t('cart.wholesaleShop')" is-link />
      </van-cell-group>
      
      <!-- 商品列表 -->
      <van-checkbox-group v-model="checkedItems">
        <van-swipe-cell
          v-for="item in cartStore.items"
          :key="item.id"
        >
          <div class="cart-item">
            <van-checkbox :name="item.id" />
            
            <van-image
              :src="item.img1 || item.image_url || ''"
              width="80"
              height="80"
              fit="cover"
            />
            
            <div class="item-info">
              <div class="item-name">{{ item.name }}</div>
              <div v-if="item.name_kh" class="item-name-kh">{{ item.name_kh }}</div>
              <div v-if="item.purchase_mode && item.purchase_mode !== 'default'" class="item-mode-tag">
                {{ item.purchase_mode === 'piece' ? (item.display_unit || '件购') : (item.display_unit || '箱购') }}
              </div>
              
              <div class="item-price">
                <span class="price-usd">${{ item.price_usd }}</span>
                <span class="price-khr">{{ formatKHR(usdToKhr(item.price_usd)) }}</span>
              </div>
              
              <van-stepper
                :model-value="item.quantity"
                :min="1"
                :max="item.stock"
                @change="val => updateQuantity(item.id, val)"
                button-size="22"
                input-width="36"
              />
            </div>
          </div>
          
          <template #right>
            <van-button
              square
              type="danger"
              :text="$t('common.delete')"
              style="height: 100%"
              @click="removeItem(item.id)"
            />
          </template>
        </van-swipe-cell>
      </van-checkbox-group>
      
      <!-- 配送信息 -->
      <van-cell-group inset>
        <van-field
          v-model="orderForm.delivery_address"
          :label="$t('cart.address')"
          :placeholder="$t('cart.addressPlaceholder')"
          clearable
          @focus="handleInputFocus"
        />
        <van-field
          v-model="orderForm.delivery_phone"
          :label="$t('cart.phone')"
          type="tel"
          :placeholder="$t('cart.phonePlaceholder')"
          clearable
          @focus="handleInputFocus"
        />
        <van-field
          v-model="orderForm.note"
          :label="$t('cart.note')"
          type="textarea"
          :placeholder="$t('cart.notePlaceholder')"
          rows="2"
          autosize
          maxlength="200"
          show-word-limit
          @focus="handleInputFocus"
        />
        <van-cell :title="$t('cart.deliveryFee')">
          <template #value>
            <span v-if="estimatingFee" style="color:#999">{{ $t('cart.estimating') }}</span>
            <span v-else-if="deliveryFee !== null" style="color:#ee0a24;font-weight:600">${{ deliveryFee.toFixed(2) }}</span>
            <span v-else style="color:#999">--</span>
          </template>
          <template v-if="autoEstimated && !estimatingFee" #label>
            <span style="font-size:11px;color:#52c41a">📍 {{ $t('cart.autoEstimated') }}</span>
          </template>
        </van-cell>
      </van-cell-group>

      <!-- 预约配送时间 -->
      <van-cell-group inset>
        <van-field
          v-model="scheduledAtDisplay"
          :label="$t('cart.scheduledAt')"
          :placeholder="$t('cart.scheduledAtPlaceholder')"
          readonly
          is-link
          @click="showDatePicker = true"
        />
      </van-cell-group>

      <!-- 预约时间选择弹窗：第一步选日期 -->
      <van-popup v-model:show="showDatePicker" teleport="body" position="bottom" round>
        <van-date-picker
          v-model="pickerDate"
          :min-date="minDate"
          :title="$t('cart.scheduledAt')"
          @confirm="onDatePickerConfirm"
          @cancel="showDatePicker = false"
        />
      </van-popup>

      <!-- 预约时间选择弹窗：第二步选小时 -->
      <van-popup v-model:show="showTimePicker" teleport="body" position="bottom" round>
        <van-time-picker
          v-model="pickerTime"
          :columns-type="['hour']"
          :title="pickerDateValues.join('-')"
          @confirm="onTimePickerConfirm"
          @cancel="showTimePicker = false"
        />
      </van-popup>

      <!-- 支付方式 -->
      <van-cell-group inset>
        <van-cell :title="$t('cart.paymentMethod')" is-link @click="openPaymentPicker">
          <template #value>
            <span :style="{ color: orderForm.payment_status === 'monthly' ? '#1989fa' : '#333' }">
              {{ orderForm.payment_status === 'monthly' ? $t('cart.monthlyPayment') : $t('cart.cashPayment') }}
            </span>
          </template>
        </van-cell>
      </van-cell-group>

      <!-- 支付方式选择弹窗 -->
      <van-action-sheet
        v-model:show="showPaymentPicker"
        teleport="body"
        :actions="paymentActions"
        :cancel-text="$t('common.cancel')"
        @select="onPaymentSelect"
      />
      
      <!-- 底部结算栏 -->
      <van-submit-bar
        :price="totalWithFee * 100"
        :button-text="$t('cart.submitOrder')"
        @submit="handlePrimaryAction"
        :loading="submitting"
        :disabled="submitting"
      >
        <van-checkbox v-model="checkAll">{{ $t('cart.selectAll') }}</van-checkbox>
        <template #tip>
          <span v-if="deliveryFee !== null">
            {{ $t('cart.itemsAndFee', { count: checkedItems.length, fee: deliveryFee.toFixed(2) }) }}
            · {{ khrLabel(totalWithFee) }}
          </span>
          <span v-else>{{ $t('cart.itemCount', { count: checkedItems.length }) }} · {{ khrLabel(totalWithFee) }}</span>
        </template>
      </van-submit-bar>
    </template>

    <!-- 下单成功面板 -->
    <van-popup
      v-model:show="orderResultVisible"
      position="bottom"
      round
      :close-on-click-overlay="false"
      teleport="body"
      :style="{ maxHeight: '88vh' }"
    >
      <div class="order-result-sheet">
        <!-- 成功标识（居中，绿色图标） -->
        <div class="or-success">
          <van-icon name="passed" size="48" color="#07c160" />
          <div class="or-success-title">下单成功</div>
          <div v-if="orderResult" class="or-success-amount">
            ${{ Number(orderResult.total_usd).toFixed(2) }}
            <span class="or-success-khr">{{ khrLabel(orderResult.total_usd) }}</span>
          </div>
        </div>

        <!-- 付款流程说明 -->
        <div class="or-notice">
          <van-icon name="warn-o" size="16" color="#ff976a" class="or-notice-icon" />
          <span>付款后即可安排派送——请将以下订单信息发给客服，确认付款后将为您安排配送。</span>
        </div>

        <!-- 订单信息文本块（直接展示，点整块复制） -->
        <div v-if="orderText" class="or-text-block" @click="copyOrderInfo">
          <div class="or-text-header">
            <span class="or-text-label">订单信息</span>
            <span class="or-copy-tag">点击复制</span>
          </div>
          <pre class="or-text-content">{{ orderText }}</pre>
        </div>

        <!-- 操作按钮 -->
        <div class="or-actions">
          <van-button v-if="contactHref" block type="primary" @click="openContact">
            <van-icon name="service-o" style="margin-right:6px" />联系客服完成付款
          </van-button>
          <van-button block plain type="primary" @click="copyOrderInfo">
            <van-icon name="description" style="margin-right:6px" />复制订单信息
          </van-button>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showToast, showDialog } from 'vant'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import { createOrder, estimateDeliveryFeeByAddress, getContactInfo } from '@/api'
import { formatKHR, usdToKhr, khrLabel } from '@/utils/format'
import { hapticFeedback } from '@/utils/device'

const { t } = useI18n()
const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()

// 下单成功面板
const orderResult = ref(null)
const orderResultVisible = ref(false)
const contactInfoData = ref({})

const contactHref = computed(() => {
  const c = contactInfoData.value
  if (c.telegram) return 'https://t.me/' + c.telegram.replace(/^@/, '')
  if (c.whatsapp) return 'https://wa.me/' + c.whatsapp.replace(/\D/g, '')
  if (c.phone) return 'tel:' + c.phone
  return ''
})

const openContact = () => {
  if (!contactHref.value) return
  const tg = window.Telegram?.WebApp
  if (tg?.openLink) tg.openLink(contactHref.value)
  else window.open(contactHref.value, '_blank')
}

const orderText = computed(() => {
  if (!orderResult.value) return ''
  const o = orderResult.value
  let text = `订单号: ${o.order_no}\n`
  if (o.delivery_address) text += `地址: ${o.delivery_address}\n`
  if (o.delivery_phone) text += `电话: ${o.delivery_phone}\n`
  text += `商品:\n`
  for (const item of o.items || []) {
    text += `  ${item.product_name} × ${item.quantity} = $${Number(item.subtotal_usd).toFixed(2)}\n`
  }
  text += `合计: $${Number(o.total_usd).toFixed(2)}`
  return text
})

const copyOrderInfo = () => {
  if (!orderText.value) return
  navigator.clipboard?.writeText(orderText.value)
    .then(() => showSuccessToast('已复制，发给客服即可'))
    .catch(() => showToast(orderText.value))
}

const checkedItems = ref(cartStore.items.map(item => item.id))
const checkAll = computed({
  get: () => cartStore.items.length > 0 && checkedItems.value.length === cartStore.items.length,
  set: (val) => { checkedItems.value = val ? cartStore.items.map(item => item.id) : [] },
})
const submitting = ref(false)
const clientRequestId = ref('')
const showPaymentPicker = ref(false)
const showDatePicker = ref(false)
const showTimePicker = ref(false)
const pickerDate = ref([])
const pickerTime = ref([])
const pickerDateValues = ref([]) // 存储第一步选中的 [y, m, d]
const minDate = new Date()
const scheduledAtDisplay = ref('')

// 第一步：选完日期后自动弹出小时选择器
const onDatePickerConfirm = ({ selectedValues }) => {
  pickerDateValues.value = selectedValues // [year, month, day]
  showDatePicker.value = false
  showTimePicker.value = true
}

// 第二步：选完小时后合并并提交
const onTimePickerConfirm = ({ selectedValues }) => {
  const [y, m, d] = pickerDateValues.value
  const [h] = selectedValues
  const hPadded = String(h).padStart(2, '0')
  scheduledAtDisplay.value = `${y}-${m}-${d} ${hPadded}:00`
  orderForm.value.scheduled_at = `${y}-${m}-${d}T${hPadded}:00:00+07:00`
  showTimePicker.value = false
}

// 运费估算（根据地址自动计算，不需要用户手动输入距离）
const deliveryFee = ref(null)
const estimatingFee = ref(false)
const autoEstimated = ref(false)
let estimateDebounceTimer = null

// 根据配送地址文本自动估算
const autoEstimateFromAddress = async (address) => {
  if (!address || address.trim().length < 5) return
  estimatingFee.value = true
  try {
    const res = await estimateDeliveryFeeByAddress('', address.trim())
    if (res.warning) return
    deliveryFee.value = res.delivery_fee_usd ?? null
    autoEstimated.value = true
  } catch {
    // 静默失败
  } finally {
    estimatingFee.value = false
  }
}

const totalWithFee = computed(() => {
  const base = cartStore.items
    .filter(item => checkedItems.value.includes(item.id))
    .reduce((sum, item) => sum + item.price_usd * item.quantity, 0)
  return base + (deliveryFee.value ?? 0)
})

const handleInputFocus = (e) => {
  setTimeout(() => {
    if (e?.target) e.target.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }, 300)
}

const paymentActions = computed(() => {
  const actions = [
    { name: t('cart.cashPayment'), value: 'cash' },
  ]
  // 月结支付需要后端支持，默认展示现金支付
  return actions
})

const buildClientRequestId = () => {
  return `mobile-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

const onPaymentSelect = (action) => {
  orderForm.value.payment_status = action.value
  showPaymentPicker.value = false
}

const openPaymentPicker = () => {
  showPaymentPicker.value = true
}

const orderForm = ref({
  delivery_address: '',
  delivery_phone: '',
  payment_status: 'cash',
  note: '',
  scheduled_at: null,
})

// 自动根据定位估算运费 + 加载联系方式
onMounted(async () => {
  const info = userStore.userInfo
  if (!orderForm.value.delivery_address && info?.address) {
    orderForm.value.delivery_address = info.address
  }
  if (!orderForm.value.delivery_phone && info?.phone) {
    orderForm.value.delivery_phone = info.phone
  }
  if (orderForm.value.delivery_address) {
    autoEstimateFromAddress(orderForm.value.delivery_address)
  }
  try {
    contactInfoData.value = await getContactInfo()
  } catch { /* 静默 */ }
})

// 地址变更时防抖自动估算
watch(() => orderForm.value.delivery_address, (newAddr) => {
  if (estimateDebounceTimer) clearTimeout(estimateDebounceTimer)
  if (!newAddr || newAddr.trim().length < 5) return
  estimateDebounceTimer = setTimeout(() => {
    autoEstimateFromAddress(newAddr)
  }, 1500)
})

// 清空购物车
const clearCart = async () => {
  const confirmed = await showDialog({
    title: t('cart.clearCart'),
    message: t('cart.deleteMessage'),
    showCancelButton: true,
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
  }).catch(() => false)
  if (confirmed) {
    cartStore.clear()
    checkedItems.value = []
    hapticFeedback('medium')
  }
}

// 更新数量
const updateQuantity = (id, quantity) => {
  hapticFeedback('light')
  cartStore.updateQuantity(id, quantity)
}

// 删除商品
const removeItem = async (id) => {
  hapticFeedback('heavy')
  const confirmed = await showDialog({
    title: t('cart.deleteConfirm'),
    message: t('cart.deleteMessage'),
    showCancelButton: true,
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
  }).catch(() => false)
  
  if (confirmed) {
    cartStore.removeItem(id)
    checkedItems.value = checkedItems.value.filter(itemId => itemId !== id)
  }
}

// 提交订单

const handleBack = () => {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.replace('/m/shop')
}

const handleSubmit = async () => {
  if (submitting.value) {
    return
  }

  if (checkedItems.value.length === 0) {
    showDialog({ message: t('cart.selectItems') })
    return
  }
  
  // 地址校验
  const address = orderForm.value.delivery_address?.trim()
  if (!address) {
    showDialog({ message: t('cart.addressRequired') })
    return
  }
  
  submitting.value = true
  clientRequestId.value = clientRequestId.value || buildClientRequestId()
  hapticFeedback('medium')
  
  try {
    const items = cartStore.items
      .filter(item => checkedItems.value.includes(item.id))
      .map(item => ({
        product_id: item.id,
        quantity: item.quantity,
        purchase_mode: item.purchase_mode || 'default',
      }))

    if (items.length === 0) {
      showToast(t('cart.selectItems'))
      submitting.value = false
      return
    }
    
    const result = await createOrder({
      items,
      ...orderForm.value,
      client_request_id: clientRequestId.value,
    })

    // 清除已下单的商品
    checkedItems.value.forEach(id => {
      cartStore.removeItem(id)
    })

    hapticFeedback('success')
    clientRequestId.value = ''
    orderResult.value = result
    orderResultVisible.value = true
  } catch (error) {
    hapticFeedback('error')
    const msg = error?.message || ''
    if (msg.includes('Items') && msg.includes('min')) {
      showToast(t('cart.selectItems'))
    } else {
      showToast(msg || t('common.requestFailed'))
    }
  } finally {
    submitting.value = false
  }
}

const handlePrimaryAction = async () => {
  if (submitting.value) {
    return
  }
  await handleSubmit()
}
</script>

<style scoped>
.mobile-cart {
  min-height: var(--tg-viewport-height, 100vh);
  background: #f5f5f5;
  padding-bottom: calc(100px + var(--tg-content-safe-area-inset-bottom, env(safe-area-inset-bottom, 0px)));
}

/* van-submit-bar 默认 bottom:0，会被导航栏遮挡，需偏移导航栏高度 */
:deep(.van-submit-bar) {
  bottom: calc(var(--van-tabbar-height, 50px) + env(safe-area-inset-bottom, 0px));
}

.cart-logo {
  height: 22px;
  display: block;
  margin: 0 auto;
}

.cart-item {
  display: flex;
  align-items: center;
  padding: 14px 12px;
  background: #fff;
  gap: 12px;
}

/* 商品卡片之间的间距和圆角 */
.van-swipe-cell + .van-swipe-cell .cart-item {
  border-top: 1px solid #f0f0f0;
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-name {
  font-size: 15px;
  font-weight: 500;
  color: #262626;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
}

.item-name-kh {
  font-size: 12px;
  color: #8c8c8c;
  margin-bottom: 4px;
}

.item-mode-tag {
  display: inline-block;
  font-size: 11px;
  color: #1D4ED8;
  background: #EFF6FF;
  border-radius: 4px;
  padding: 1px 6px;
  margin-bottom: 6px;
}

.item-price {
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.price-usd {
  font-size: 16px;
  font-weight: 700;
  color: #f5222d;
}

.price-khr {
  font-size: 12px;
  color: #8c8c8c;
}

/* ===== 增强空状态 ===== */
.cart-empty {
  padding-top: 80px;
}
.cart-empty :deep(.van-empty__image) {
  margin-bottom: 8px;
}
.empty-cart-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100px;
  height: 100px;
  border-radius: 50%;
  background: #fafafa;
  margin: 0 auto;
}
.cart-empty :deep(.van-empty__description) {
  font-size: 15px;
  color: #999;
  margin-top: 12px;
}

/* ===== 商品列表卡片容器 ===== */
:deep(.van-checkbox-group) {
  margin: 8px 12px;
  border-radius: 12px;
  overflow: hidden;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

/* 商品卡片内部 */
.cart-item {
  display: flex;
  align-items: center;
  padding: 14px 12px;
  background: #fff;
  gap: 12px;
}

.van-swipe-cell + .van-swipe-cell .cart-item {
  border-top: 1px solid #f0f0f0;
}

/* 商品首项增加圆角 */
:deep(.van-swipe-cell:first-child .cart-item) {
  border-radius: 12px 12px 0 0;
}
:deep(.van-swipe-cell:last-child .cart-item) {
  border-radius: 0 0 12px 12px;
}
:deep(.van-swipe-cell:only-child .cart-item) {
  border-radius: 12px;
}

.clear-cart-btn {
  font-size: 13px;
  color: #ee0a24;
  padding: 4px 4px;
  cursor: pointer;
}

/* ===== 下单成功面板 ===== */
.order-result-sheet {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 24px 16px calc(20px + env(safe-area-inset-bottom, 0px));
  overflow-y: auto;
  max-height: 88vh;
}

/* 成功标识 */
.or-success {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}
.or-success-title {
  font-size: 20px;
  font-weight: 700;
  color: #323233;
}
.or-success-amount {
  font-size: 28px;
  font-weight: 800;
  color: #323233;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}
.or-success-khr {
  font-size: 13px;
  font-weight: 400;
  color: #969799;
}

/* 付款说明条（Vant warning 色调） */
.or-notice {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  background: #fffbe8;
  border: 1px solid #f5d87a;
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 13px;
  color: #7d4e00;
  line-height: 1.6;
}
.or-notice-icon { margin-top: 1px; flex-shrink: 0; }

/* 订单信息文本块 */
.or-text-block {
  border: 1px solid #ebedf0;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
}
.or-text-block:active { background: #f7f8fa; }
.or-text-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f7f8fa;
  border-bottom: 1px solid #ebedf0;
  font-size: 13px;
  font-weight: 600;
  color: #323233;
}
.or-copy-tag {
  font-size: 11px;
  color: #1989fa;
  font-weight: 400;
}
.or-text-content {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  font-size: 13px;
  line-height: 1.8;
  color: #323233;
  padding: 12px;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: min(200px, 30vh);
  overflow-y: auto;
  background: #fff;
}

/* 操作按钮 */
.or-actions { display: flex; flex-direction: column; gap: 10px; }

.checkout-disabled-tip {
  margin-bottom: 8px;
}

/* 紧凑表单布局 */
:deep(.van-cell-group) {
  margin-top: 6px !important;
}
:deep(.van-cell-group .van-cell) {
  padding-top: 8px;
  padding-bottom: 8px;
}
:deep(.van-field__label) {
  width: auto;
  min-width: 54px;
  margin-right: 6px;
}
</style>
