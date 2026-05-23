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
    
    <van-empty v-if="cartStore.items.length === 0" :description="$t('cart.empty')">
      <van-button type="primary" round @click="$router.push('/m/shop')">
        {{ $t('cart.goShopping') }}
      </van-button>
    </van-empty>
    
    <template v-else>
      <div v-if="!userStore.canOrder" class="restriction-panel">
        <div class="restriction-title">{{ restrictionTitle }}</div>
        <div class="restriction-desc">{{ restrictionDescription }}</div>
        <div class="restriction-note">{{ cartSavedTip }}</div>
        <van-button round block type="warning" class="restriction-btn" @click="goToRestrictionAction">
          {{ restrictionActionLabel }}
        </van-button>
      </div>

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
              :src="item.image_url || '/placeholder.png'"
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
        <van-cell
          v-if="!userStore.canOrder"
          :title="cartFormTip"
          icon="warning-o"
          class="checkout-disabled-tip"
        />
        <van-field
          v-model="orderForm.delivery_address"
          :label="$t('cart.address')"
          :placeholder="$t('cart.addressPlaceholder')"
          :readonly="!formEditable"
          clearable
          @focus="handleInputFocus"
        />
        <van-field
          v-model="orderForm.delivery_phone"
          :label="$t('cart.phone')"
          type="tel"
          :placeholder="$t('cart.phonePlaceholder')"
          :readonly="!formEditable"
          clearable
          @focus="handleInputFocus"
        />
        <van-field
          v-model="orderForm.note"
          :label="$t('cart.note')"
          type="textarea"
          :placeholder="$t('cart.notePlaceholder')"
          :readonly="!formEditable"
          rows="2"
          autosize
          maxlength="200"
          show-word-limit
          @focus="handleInputFocus"
        />
        <!-- 配送距离 & 运费（自动估算，只读） -->
        <van-field
          v-model="distanceInput"
          :label="$t('cart.distanceKm')"
          :placeholder="estimatingFee ? $t('cart.estimating') : '--'"
          type="number"
          readonly
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
          :disabled="!formEditable"
          @click="formEditable && (showDatePicker = true)"
        />
      </van-cell-group>

      <!-- 预约时间选择弹窗：第一步选日期 -->
      <van-popup v-model:show="showDatePicker" teleport="body" position="bottom" round>
        <van-date-picker
          v-model="pickerDate"
          :min-date="minDate"
          :title="$t('cart.scheduledAt') + ' — 1/2 选日期'"
          @confirm="onDatePickerConfirm"
          @cancel="showDatePicker = false"
        />
      </van-popup>

      <!-- 预约时间选择弹窗：第二步选小时 -->
      <van-popup v-model:show="showTimePicker" teleport="body" position="bottom" round>
        <van-time-picker
          v-model="pickerTime"
          :columns-type="['hour']"
          :title="pickerDateValues.join('-') + ' — 2/2 选时间'"
          @confirm="onTimePickerConfirm"
          @cancel="showTimePicker = false"
        />
      </van-popup>

      <!-- 支付方式 -->
      <van-cell-group inset>
        <van-cell :title="$t('cart.paymentMethod')" :is-link="formEditable" @click="openPaymentPicker">
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
        :button-text="primaryButtonText"
        @submit="handlePrimaryAction"
        :loading="submitting"
        :disabled="submitting"
      >
        <van-checkbox v-model="checkAll">{{ $t('cart.selectAll') }}</van-checkbox>
        <template #tip>
          <span v-if="deliveryFee !== null">{{ $t('cart.itemsAndFee', { count: checkedItems.length, fee: deliveryFee.toFixed(2) }) }}</span>
          <span v-else>{{ footerTipText }}</span>
        </template>
      </van-submit-bar>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showToast, showDialog } from 'vant'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import { createOrder, estimateDeliveryFee, estimateDeliveryFeeByAddress } from '@/api'
import { formatKHR, usdToKhr } from '@/utils/format'
import { hapticFeedback } from '@/utils/device'

const { t } = useI18n()
const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()

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

// 运费估算
const distanceInput = ref('')
const deliveryFee = ref(null)
const estimatingFee = ref(false)
const autoEstimated = ref(false)
let estimateDebounceTimer = null

const doEstimateFee = async () => {
  const km = parseFloat(distanceInput.value)
  if (isNaN(km) || km < 0) return
  estimatingFee.value = true
  try {
    const res = await estimateDeliveryFee(km)
    deliveryFee.value = res.fee ?? res.delivery_fee ?? res.amount ?? null
  } catch {
    deliveryFee.value = null
  } finally {
    estimatingFee.value = false
  }
}

const handleDistanceBlur = () => {
  if (distanceInput.value !== '') {
    autoEstimated.value = false
    doEstimateFee()
  }
}

// 根据用户保存的 Google 定位自动估算运费
const autoEstimateFromLocation = async () => {
  const locationUrl = userStore.userInfo?.location_url
  if (!locationUrl) return
  const m = locationUrl.match(/[?&]q=([-\d.]+),([-\d.]+)/)
  if (!m) return
  const destination = `${m[1]},${m[2]}`
  estimatingFee.value = true
  try {
    const res = await estimateDeliveryFeeByAddress('', destination)
    if (res.warning) {
      // API 无法获取实际距离，不覆盖现有值
      return
    }
    if (res.distance_km !== undefined) distanceInput.value = String(Number(res.distance_km).toFixed(1))
    deliveryFee.value = res.delivery_fee_usd ?? null
    autoEstimated.value = true
  } catch {
    // 静默失败，用户可手动填写距离
  } finally {
    estimatingFee.value = false
  }
}

// 根据配送地址文本自动估算（无 location_url 时的兜底）
const autoEstimateFromAddress = async (address) => {
  if (!address || address.trim().length < 5) return
  estimatingFee.value = true
  try {
    const res = await estimateDeliveryFeeByAddress('', address.trim())
    if (res.warning) {
      // API 无法解析地址，不覆盖现有值
      return
    }
    if (res.distance_km !== undefined) distanceInput.value = String(Number(res.distance_km).toFixed(1))
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

const formEditable = computed(() => userStore.canOrder || userStore.orderAccessState === 'incomplete')

const handleInputFocus = (e) => {
  setTimeout(() => {
    if (e?.target) e.target.scrollIntoView({ behavior: 'smooth', block: 'center' })
  }, 300)
}

const restrictionTitle = computed(() => {
  switch (userStore.orderAccessState) {
    case 'incomplete':
      return t('profile.orderGuideIncompleteTitle')
    case 'pending':
      return t('profile.orderGuidePendingTitle')
    case 'rejected':
      return t('profile.orderGuideRejectedTitle')
    default:
      return t('cart.checkoutDisabled')
  }
})

const restrictionDescription = computed(() => {
  if (userStore.orderAccessState === 'rejected' && userStore.userInfo?.rejected_reason) {
    return `${t('profile.rejectedReason')}：${userStore.userInfo.rejected_reason}。${t('profile.orderGuideRejectedDesc')}`
  }
  switch (userStore.orderAccessState) {
    case 'incomplete':
      return t('profile.orderGuideIncompleteDesc')
    case 'pending':
      return t('profile.orderGuidePendingDesc')
    case 'rejected':
      return t('profile.orderGuideRejectedDesc')
    default:
      return t('cart.checkoutDisabled')
  }
})

const restrictionActionLabel = computed(() => {
  switch (userStore.orderAccessState) {
    case 'incomplete':
      return t('profile.completeProfileAction')
    case 'pending':
      return t('profile.viewApprovalStatusAction')
    case 'rejected':
      return t('profile.resubmitForReview')
    default:
      return t('cart.submitOrder')
  }
})

const cartSavedTip = computed(() => {
  switch (userStore.orderAccessState) {
    case 'incomplete':
      return t('cart.draftInfoTip')
    case 'pending':
      return t('cart.savedForApprovalTip')
    case 'rejected':
      return t('cart.savedForResubmitTip')
    default:
      return t('cart.itemCount', { count: checkedItems.value.length })
  }
})

const cartFormTip = computed(() => {
  return formEditable.value ? t('cart.draftInfoTip') : cartSavedTip.value
})

const primaryButtonText = computed(() => {
  return userStore.canOrder ? t('cart.submitOrder') : restrictionActionLabel.value
})

const footerTipText = computed(() => {
  return userStore.canOrder ? t('cart.itemCount', { count: checkedItems.value.length }) : cartSavedTip.value
})

const paymentActions = computed(() => {
  const actions = [
    { name: t('cart.cashPayment'), value: 'cash' },
  ]
  if (userStore.userInfo?.allow_credit) {
    actions.push({ name: t('cart.monthlyPayment'), value: 'monthly' })
  }
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
  if (!formEditable.value) {
    goToRestrictionAction()
    return
  }
  showPaymentPicker.value = true
}

const orderForm = ref({
  delivery_address: '',
  delivery_phone: '',
  payment_status: 'cash',
  note: '',
  scheduled_at: null,
})

// 自动填充个人资料中的默认地址和电话，并根据定位自动估算运费
onMounted(() => {
  if (userStore.userInfo?.address) {
    orderForm.value.delivery_address = userStore.userInfo.address
  }
  if (userStore.userInfo?.phone) {
    orderForm.value.delivery_phone = userStore.userInfo.phone
  }
  // 优先用精确定位，没有则用地址文本
  if (userStore.userInfo?.location_url) {
    autoEstimateFromLocation()
  } else if (orderForm.value.delivery_address) {
    autoEstimateFromAddress(orderForm.value.delivery_address)
  }
})

// 地址变更时防抖自动估算
watch(() => orderForm.value.delivery_address, (newAddr) => {
  if (estimateDebounceTimer) clearTimeout(estimateDebounceTimer)
  if (!newAddr || newAddr.trim().length < 5) return
  estimateDebounceTimer = setTimeout(() => {
    if (userStore.userInfo?.location_url) {
      autoEstimateFromLocation()
    } else {
      autoEstimateFromAddress(newAddr)
    }
  }, 1500)
})

// 选中商品的总价
const totalPrice = computed(() => {
  return cartStore.items
    .filter(item => checkedItems.value.includes(item.id))
    .reduce((sum, item) => sum + item.price_usd * item.quantity, 0)
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
const goToRestrictionAction = () => {
  showDialog({ message: restrictionDescription.value })
  router.push('/m/profile')
}

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

  if (!userStore.canOrder) {
    goToRestrictionAction()
    return
  }

  if (checkedItems.value.length === 0) {
    showDialog({ message: t('cart.selectItems') })
    return
  }
  
  // 地址校验：表单地址和个人资料默认地址都没有时才拦截
  const address = orderForm.value.delivery_address?.trim() || userStore.userInfo?.address?.trim()
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
    
    await createOrder({
      items,
      ...orderForm.value,
      client_request_id: clientRequestId.value,
    })
    
    // 清除已下单的商品
    checkedItems.value.forEach(id => {
      cartStore.removeItem(id)
    })
    
    hapticFeedback('success')
    showSuccessToast(t('cart.orderSuccess'))
    clientRequestId.value = ''
    router.push('/m/orders')
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
  if (!userStore.canOrder) {
    goToRestrictionAction()
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

.restriction-panel {
  margin: 10px 12px 0;
  padding: 14px;
  background: #fff7e6;
  border: 1px solid #ffd591;
  border-radius: 12px;
}

.restriction-title {
  font-size: 15px;
  font-weight: 700;
  color: #ad6800;
}

.restriction-desc,
.restriction-note {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.6;
  color: #8c5a00;
}

.restriction-btn {
  margin-top: 12px;
}

.cart-item {
  display: flex;
  align-items: center;
  padding: 12px;
  background: #fff;
  gap: 12px;
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

.clear-cart-btn {
  font-size: 13px;
  color: #ee0a24;
  padding: 4px 4px;
  cursor: pointer;
}

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
