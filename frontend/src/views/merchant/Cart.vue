<template>
  <div class="cart-page">
    <!-- 顶部导航 -->
    <div class="cart-header">
      <button class="back-btn" @click="$router.back()">
        <el-icon :size="20"><ArrowLeft /></el-icon>
      </button>
      <h2 class="cart-title">{{ $t('cart.title') }}</h2>
    </div>

    <!-- 空状态 -->
    <el-empty v-if="cartStore.items.length === 0" :description="$t('cart.empty')">
      <el-button type="primary" round @click="$router.push('/merchant/products')">
        {{ $t('cart.goShopping') }}
      </el-button>
    </el-empty>
    
    <template v-else>
      <div v-if="!userStore.canOrder" class="restriction-panel">
        <div class="restriction-copy">
          <div class="restriction-title">{{ restrictionTitle }}</div>
          <div class="restriction-desc">{{ restrictionDescription }}</div>
          <div class="restriction-note">{{ cartSavedTip }}</div>
        </div>
        <el-button type="warning" @click="goToRestrictionAction">
          {{ restrictionActionLabel }}
        </el-button>
      </div>

      <!-- 商品列表 (卡片式) -->
      <div class="cart-list">
        <div
          v-for="item in cartStore.items"
          :key="item.id"
          class="cart-item"
        >
          <div class="item-image">
            <img
              v-if="item.img1 || item.image_url"
              :src="item.img1 || item.image_url"
              :alt="item.name"
            />
            <div v-else class="image-placeholder">
              <el-icon :size="28" color="#ccc"><Picture /></el-icon>
            </div>
          </div>
          
          <div class="item-info">
            <div class="item-name">{{ item.name }}</div>
            <div v-if="item.name_kh" class="item-name-kh">{{ item.name_kh }}</div>
            <div v-if="item.purchase_mode && item.purchase_mode !== 'default'" class="item-name-kh">
              {{ item.purchase_mode === 'piece' ? $t('product.buyByPiece') : $t('product.buyByPackage') }}
            </div>
            <div class="item-price">
              <span class="price-usd">${{ item.price_usd }}</span>
              <span class="price-khr">{{ formatKHR(usdToKhr(item.price_usd)) }}</span>
            </div>
            <div class="item-bottom">
              <div class="qty-control">
                <button class="qty-btn" @click="changeQty(item, -1)">−</button>
                <span class="qty-num">{{ item.quantity }}</span>
                <button class="qty-btn" @click="changeQty(item, 1)" :disabled="item.quantity >= item.stock">+</button>
              </div>
              <span class="item-subtotal">${{ (item.price_usd * item.quantity).toFixed(2) }}</span>
            </div>
          </div>
          
          <button class="delete-btn" @click="removeItem(item.id)">
            <el-icon :size="16"><Delete /></el-icon>
          </button>
        </div>
      </div>
      
      <!-- 订单信息表单 -->
      <div class="order-section">
        <div class="section-title">{{ $t('cart.deliveryInfo') }}</div>
        <el-alert
          v-if="!userStore.canOrder"
          :title="cartFormTip"
          type="warning"
          :closable="false"
          class="checkout-disabled-tip"
        />
        <div class="form-group">
          <label>{{ $t('cart.address') }}</label>
          <input
            v-model="orderForm.delivery_address"
            type="text"
            :placeholder="$t('cart.addressPlaceholder')"
            :disabled="!formEditable"
            class="form-input"
          />
        </div>
        <div class="form-group">
          <label>{{ $t('cart.phone') }}</label>
          <input
            v-model="orderForm.delivery_phone"
            type="tel"
            :placeholder="$t('cart.phonePlaceholder')"
            :disabled="!formEditable"
            class="form-input"
          />
        </div>
        <div class="form-group">
          <label>{{ $t('cart.note') }}</label>
          <textarea
            v-model="orderForm.note"
            :placeholder="$t('cart.notePlaceholder')"
            rows="2"
            :disabled="!formEditable"
            class="form-input"
          ></textarea>
        </div>
        <div class="form-group">
          <label>{{ $t('cart.distanceKm') }}</label>
          <input
            v-model.number="orderForm.distance_km"
            type="number"
            min="0"
            step="0.1"
            :placeholder="$t('cart.distanceKmPlaceholder')"
            :disabled="!formEditable"
            class="form-input"
            @change="estimateFee"
          />
        </div>
        <div class="form-group">
          <label>{{ $t('cart.scheduledAt') }}</label>
          <input
            v-model="orderForm.scheduled_at"
            type="datetime-local"
            :disabled="!formEditable"
            class="form-input"
          />
        </div>
      </div>
      
      <!-- 支付方式 -->
      <div class="order-section">
        <div class="section-title">{{ $t('cart.paymentMethod') }}</div>
        <div class="payment-options">
          <div
            class="payment-option"
            :class="{ active: orderForm.payment_status === 'cash', disabled: !formEditable }"
            @click="setPaymentStatus('cash')"
          >
            <span class="payment-icon">💵</span>
            <span class="payment-label">{{ $t('cart.cashPayment') }}</span>
          </div>
          <div
            v-if="userStore.userInfo?.allow_credit"
            class="payment-option"
            :class="{ active: orderForm.payment_status === 'monthly', disabled: !formEditable }"
            @click="setPaymentStatus('monthly')"
          >
            <span class="payment-icon">📅</span>
            <span class="payment-label">{{ $t('cart.monthlyPayment') }}</span>
          </div>
        </div>
      </div>
      
      <!-- 底部结算栏 -->
      <div class="checkout-bar">
        <div class="bar-left">
          <button class="clear-btn" @click="handleClear">{{ $t('common.clear') }}</button>
          <div class="bar-total">
            <span class="bar-count">{{ $t('common.items', { count: cartStore.totalCount }) }}</span>
            <span class="bar-count">{{ $t('cart.deliveryFee') }}: ${{ deliveryFee.toFixed(2) }}</span>
            <span class="bar-price">${{ (cartStore.totalPrice + deliveryFee).toFixed(2) }}</span>
          </div>
        </div>
        <button
          class="submit-btn"
          :disabled="submitting"
          @click="handlePrimaryAction"
        >
          {{ submitting ? $t('common.submitting') : primaryButtonText }}
        </button>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus/es/components/message/index'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { Delete, Picture, ArrowLeft } from '@element-plus/icons-vue'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import { createOrder, estimateDeliveryFee } from '@/api'
import { formatKHR, usdToKhr } from '@/utils/format'

const router = useRouter()
const cartStore = useCartStore()
const userStore = useUserStore()
const { t } = useI18n()

const submitting = ref(false)
const clientRequestId = ref('')
const formEditable = computed(() => userStore.canOrder || userStore.orderAccessState === 'incomplete')
const orderForm = reactive({
  delivery_address: '',
  delivery_phone: '',
  payment_status: 'cash',
  note: '',
  distance_km: 0,
  scheduled_at: '',
})
const deliveryFee = ref(0)

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
    return `${t('profile.rejectedReason')}: ${userStore.userInfo.rejected_reason}. ${t('profile.orderGuideRejectedDesc')}`
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
      return t('common.items', { count: cartStore.totalCount })
  }
})

const cartFormTip = computed(() => {
  return formEditable.value ? t('cart.draftInfoTip') : cartSavedTip.value
})

const primaryButtonText = computed(() => {
  return userStore.canOrder ? t('cart.submitOrder') : restrictionActionLabel.value
})

const buildClientRequestId = () => {
  return `web-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

// 自动填充个人资料中的默认地址和电话
onMounted(() => {
  if (userStore.userInfo?.address) {
    orderForm.delivery_address = userStore.userInfo.address
  }
  if (userStore.userInfo?.phone) {
    orderForm.delivery_phone = userStore.userInfo.phone
  }
  estimateFee()
})

const estimateFee = async () => {
  try {
    const distance = Number(orderForm.distance_km || 0)
    const data = await estimateDeliveryFee(distance)
    deliveryFee.value = Number(data.delivery_fee_usd || 0)
  } catch {
    deliveryFee.value = 0
  }
}

// 数量变化
const changeQty = (item, delta) => {
  const newQty = item.quantity + delta
  if (newQty < 1) {
    removeItem(item.id)
    return
  }
  if (newQty > item.stock) return
  cartStore.updateQuantity(item.id, newQty)
}

const setPaymentStatus = (status) => {
  if (!formEditable.value) {
    goToRestrictionAction()
    return
  }
  orderForm.payment_status = status
}

// 删除商品
const removeItem = async (id) => {
  try {
    await ElMessageBox.confirm(t('cart.deleteMessage'), t('cart.deleteConfirm'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    })
    cartStore.removeItem(id)
  } catch {}
}

// 清空购物车
const handleClear = async () => {
  try {
    await ElMessageBox.confirm(t('cart.clearConfirm'), t('cart.deleteConfirm'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    })
    cartStore.clear()
  } catch {}
}

// 提交订单
const handleCheckout = async () => {
  if (submitting.value) {
    return
  }

  if (!userStore.canOrder) {
    goToRestrictionAction()
    return
  }

  if (cartStore.items.length === 0) {
    ElMessage.warning(t('cart.empty'))
    return
  }
  
  // 地址校验：表单地址和个人资料默认地址都没有时才拦截
  const address = orderForm.delivery_address?.trim() || userStore.userInfo?.address?.trim()
  if (!address) {
    ElMessage.warning(t('cart.addressRequired'))
    return
  }
  
  try {
    await ElMessageBox.confirm(t('cart.submitConfirm'), t('common.confirm'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'info',
    })
  } catch {
    return
  }
  
  submitting.value = true
  clientRequestId.value = clientRequestId.value || buildClientRequestId()
  try {
    const items = cartStore.items.map((item) => ({
      product_id: item.id,
      quantity: item.quantity,
      purchase_mode: item.purchase_mode || 'default',
    }))
    
    await createOrder({
      items,
      ...orderForm,
      distance_km: Number(orderForm.distance_km || 0),
      scheduled_at: orderForm.scheduled_at ? new Date(orderForm.scheduled_at).toISOString() : null,
      client_request_id: clientRequestId.value,
    })
    
    ElMessage.success(t('cart.orderSuccess'))
    cartStore.clear()
    clientRequestId.value = ''
    router.push('/merchant/orders')
  } catch (error) {
    // 修复: 错误已被 request.js 拦截器提示，此处仅打印日志，不清空购物车
    console.error('[订单] 提交失败:', error)
  } finally {
    submitting.value = false
  }
}

const goToRestrictionAction = () => {
  ElMessage.warning(restrictionDescription.value)
  router.push('/merchant/profile')
}

const handlePrimaryAction = async () => {
  if (submitting.value) {
    return
  }
  if (!userStore.canOrder) {
    goToRestrictionAction()
    return
  }
  await handleCheckout()
}
</script>

<style scoped>
.cart-page {
  min-height: 100%;
  background: #f7f7f7;
  padding-bottom: 80px;
}

.restriction-panel {
  margin: 12px;
  padding: 16px;
  border-radius: 12px;
  background: #fff7e6;
  border: 1px solid #ffd591;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.restriction-copy {
  min-width: 0;
}

.restriction-title {
  font-size: 16px;
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

.cart-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 12px 8px;
}

.back-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  background: #fff;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 1px 4px rgba(0,0,0,0.08);
  transition: all 0.2s;
  flex-shrink: 0;
}

.back-btn:hover {
  background: #ecf5ff;
  color: #409eff;
}

.cart-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
}

/* 商品列表 */
.cart-list {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.cart-item {
  display: flex;
  background: #fff;
  border-radius: 12px;
  padding: 14px;
  gap: 14px;
  position: relative;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.item-image {
  width: 90px;
  height: 90px;
  border-radius: 8px;
  overflow: hidden;
  flex-shrink: 0;
  background: #f5f5f5;
}

.item-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.item-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.item-name {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-name-kh {
  font-size: 12px;
  color: #999;
}

.item-price {
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
  color: #999;
}

.item-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: auto;
}

.qty-control {
  display: flex;
  align-items: center;
  background: #f5f5f5;
  border-radius: 20px;
  overflow: hidden;
}

.qty-btn {
  width: 30px;
  height: 30px;
  border: none;
  background: #409eff;
  color: #fff;
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.qty-btn:hover { background: #66b1ff; }
.qty-btn:disabled { background: #c0c4cc; cursor: not-allowed; }

.qty-num {
  width: 36px;
  text-align: center;
  font-size: 14px;
  font-weight: 600;
}

.item-subtotal {
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
}

.delete-btn {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 28px;
  height: 28px;
  border: none;
  background: #fef0f0;
  color: #f56c6c;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.delete-btn:hover {
  background: #f56c6c;
  color: #fff;
}

/* 订单信息 */
.order-section {
  margin: 12px;
  background: #fff;
  border-radius: 12px;
  padding: 16px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 14px;
}

.checkout-disabled-tip {
  margin-bottom: 14px;
}

.form-group {
  margin-bottom: 14px;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  font-size: 13px;
  color: #999;
  margin-bottom: 6px;
}

.form-input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  font-size: 14px;
  color: #1a1a1a;
  outline: none;
  transition: border-color 0.2s;
  font-family: inherit;
  resize: vertical;
  box-sizing: border-box;
}

.form-input:focus {
  border-color: #409eff;
}

.form-input::placeholder {
  color: #c0c0c0;
}

/* 支付方式 */
.payment-options {
  display: flex;
  gap: 12px;
}

.payment-option {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 14px 16px;
  border: 2px solid #e8e8e8;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;
}

.payment-option:hover {
  border-color: #409eff;
}

.payment-option.disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.payment-option.active {
  border-color: #409eff;
  background: #f0f7ff;
}

.payment-icon {
  font-size: 20px;
}

.payment-label {
  font-weight: 600;
  font-size: 14px;
  color: #1a1a1a;
}

/* 底部结算栏 */
.checkout-bar {
  position: fixed;
  bottom: 0;
  left: 200px;
  right: 0;
  height: 64px;
  background: #fff;
  border-top: 1px solid #eee;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  z-index: 100;
  box-shadow: 0 -2px 12px rgba(0, 0, 0, 0.06);
}

.bar-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.clear-btn {
  padding: 6px 16px;
  border: 1px solid #ddd;
  background: #fff;
  border-radius: 20px;
  font-size: 13px;
  color: #666;
  cursor: pointer;
  transition: all 0.2s;
}

.clear-btn:hover {
  border-color: #f56c6c;
  color: #f56c6c;
}

.bar-total {
  display: flex;
  flex-direction: column;
}

.bar-count {
  font-size: 12px;
  color: #999;
}

.bar-price {
  font-size: 22px;
  font-weight: 700;
  color: #f5222d;
}

.submit-btn {
  padding: 0 40px;
  height: 42px;
  background: #409eff;
  color: #fff;
  border: none;
  border-radius: 21px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.submit-btn:hover {
  background: #66b1ff;
}

.submit-btn:disabled {
  background: #a0cfff;
  cursor: not-allowed;
}

/* === 移动端适配 === */
@media (max-width: 767px) {
  .cart-page {
    padding: 12px;
    padding-bottom: 130px;
  }

  .checkout-bar {
    left: 0 !important;
    bottom: 50px;
  }

  .cart-item {
    padding: 12px;
    flex-wrap: wrap;
  }

  .item-image {
    width: 60px;
    height: 60px;
  }

  .order-form {
    padding: 12px;
  }

  .submit-btn {
    padding: 0 24px;
    font-size: 14px;
  }

  .bar-price {
    font-size: 18px;
  }
}
</style>
