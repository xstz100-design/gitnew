<template>
  <div class="mobile-shop">
    <!-- 公告栏（紧凑） -->
    <van-notice-bar
      v-if="notices.length > 0"
      :text="notices.map(n => currentLang === 'zh' ? n.content_zh : n.content_en).join('          ')"
      left-icon="volume-o"
      :speed="50"
      background="#fffbe6"
      color="#ad6800"
      :wrapable="false"
    />

    <!-- 搜索栏 -->
    <div class="shop-header">
      <div class="header-row">
        <div class="search-box">
          <van-icon name="search" class="search-icon" />
          <input
            v-model="searchKeyword"
            type="text"
            :placeholder="$t('product.searchPlaceholder')"
            class="search-input"
            @input="onSearch"
          />
          <van-icon
            v-if="searchKeyword"
            name="clear"
            class="clear-icon"
            @click="searchKeyword = ''; onSearch()"
          />
        </div>
        <button class="contact-btn" @click="contactVisible = true">
          <van-icon name="phone-o" />
          <span>{{ $t('common.contactUs') }}</span>
        </button>
        <button v-if="contactInfo.telegram" class="telegram-header-btn" @click="openTelegram">
          <van-icon name="send-gift-o" size="16" />
        </button>
        <button class="lang-btn" @click="toggleLang">{{ langLabel }}</button>
      </div>
    </div>

    <!-- 促销跑马灯合并区域（更紧凑） -->
    <div class="promo-section" v-if="newProducts.length > 0 || discountProducts.length > 0">
      <div class="promo-tabs">
        <span
          class="promo-tab"
          :class="{ active: promoActive === 'new' }"
          @click="promoActive = 'new'"
          v-if="newProducts.length > 0"
        >
          <span>🔥</span> {{ $t('product.newProducts') }}
        </span>
        <span
          class="promo-tab"
          :class="{ active: promoActive === 'discount' }"
          @click="promoActive = 'discount'"
          v-if="discountProducts.length > 0"
        >
          <span>🏷️</span> {{ $t('product.discountProducts') }}
        </span>
      </div>
      <div class="promo-marquee">
        <div
          class="marquee-track"
          :style="{ animationDuration: promoActive === 'new' ? newProductsDuration : discountProductsDuration }"
          v-if="promoActive === 'new' && newProducts.length > 0"
        >
          <div class="marquee-group">
            <div v-for="product in newProducts" :key="'na-'+product.id" class="promo-card" @click="showDetail(product)">
              <img v-if="getProductImage(product)" :src="getProductImage(product)" :alt="product.name" loading="lazy" />
              <div v-else class="promo-img-ph"><van-icon name="photo-o" size="12" color="#ccc" /></div>
              <span class="promo-badge">NEW</span>
              <div class="promo-card-price">${{ product.price_usd }}</div>
            </div>
          </div>
          <div v-if="newProducts.length > 1" class="marquee-group">
            <div v-for="product in newProducts" :key="'nb-'+product.id" class="promo-card" @click="showDetail(product)">
              <img v-if="getProductImage(product)" :src="getProductImage(product)" :alt="product.name" loading="lazy" />
              <div v-else class="promo-img-ph"><van-icon name="photo-o" size="12" color="#ccc" /></div>
              <span class="promo-badge">NEW</span>
              <div class="promo-card-price">${{ product.price_usd }}</div>
            </div>
          </div>
        </div>
        <div
          class="marquee-track"
          :style="{ animationDuration: promoActive === 'new' ? newProductsDuration : discountProductsDuration }"
          v-if="promoActive === 'discount' && discountProducts.length > 0"
        >
          <div class="marquee-group">
            <div v-for="product in discountProducts" :key="'da-'+product.id" class="promo-card" @click="showDetail(product)">
              <img v-if="getProductImage(product)" :src="getProductImage(product)" :alt="product.name" loading="lazy" />
              <div v-else class="promo-img-ph"><van-icon name="photo-o" size="12" color="#ccc" /></div>
              <span class="promo-badge dc-badge">{{ $t('product.discountBadge') }}</span>
              <div class="promo-card-price">${{ product.price_usd }}</div>
            </div>
          </div>
          <div v-if="discountProducts.length > 1" class="marquee-group">
            <div v-for="product in discountProducts" :key="'db-'+product.id" class="promo-card" @click="showDetail(product)">
              <img v-if="getProductImage(product)" :src="getProductImage(product)" :alt="product.name" loading="lazy" />
              <div v-else class="promo-img-ph"><van-icon name="photo-o" size="12" color="#ccc" /></div>
              <span class="promo-badge dc-badge">{{ $t('product.discountBadge') }}</span>
              <div class="promo-card-price">${{ product.price_usd }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 主体：左侧分类 + 右侧商品 -->
    <div class="main-body">
      <!-- 左侧分类侧边栏 -->
      <div class="cat-sidebar" ref="catSidebarRef">
        <div
          v-for="cat in categories"
          :key="cat.id"
          class="cat-item"
          :class="{ active: activeCat === cat.id }"
          @click="switchCategory(cat.id)"
        >
          <span class="cat-dot" :class="{ active: activeCat === cat.id }"></span>
          <span class="cat-name">{{ cat.name }}</span>
        </div>
      </div>

      <!-- 右侧商品区域 -->
      <div class="product-area" ref="scrollContainer">
        <!-- 骨架屏 -->
        <div v-if="loading" class="skeleton-grid">
          <div v-for="i in 6" :key="i" class="skeleton-card">
            <div class="skeleton-img shimmer"></div>
            <div class="skeleton-name shimmer"></div>
            <div class="skeleton-price shimmer"></div>
          </div>
        </div>

        <!-- 空状态 -->
        <van-empty v-else-if="filteredProducts.length === 0" :description="$t('common.noData')" />

        <!-- 商品网格 -->
        <div v-else class="product-grid">
          <div
            v-for="product in filteredProducts"
            :key="product.id"
            class="product-card"
            @click="showDetail(product)"
          >
            <!-- 图片 -->
            <div class="card-image">
              <img
                v-if="getProductImage(product)"
                :src="getProductImage(product)"
                :alt="product.name"
                loading="lazy"
              />
              <div v-else class="image-placeholder">
                <van-icon name="photo-o" size="24" color="#ccc" />
              </div>
              <!-- 标记 -->
              <span v-if="product.stock <= 0" class="badge-out">{{ $t('product.outOfStock') }}</span>
              <span v-else-if="product.is_low_stock" class="badge-low">{{ $t('product.lowStock') }}</span>
              <span v-if="product.is_featured" class="badge-featured">{{ $t('product.recommended') }}</span>
            </div>

            <!-- 名称 -->
            <div class="card-name">{{ product.name }}</div>

            <!-- 单价 -->
            <div class="card-price-row">
              <span class="price-val">${{ product.price_usd }}</span>
              <span class="price-unit">/{{ product.unit || $t('product.unitFallback') }}</span>
            </div>

            <!-- 阶梯/批发价格（突出显示） -->
            <div v-if="product.price_per_piece_usd || product.price_per_package_usd" class="card-tiers">
              <div v-if="product.price_per_piece_usd" class="tier-row">
                <span class="tier-label">{{ $t('product.pricePerPiece') }}</span>
                <span class="tier-value">${{ Number(product.price_per_piece_usd).toFixed(2) }}<em v-if="product.unit_name">/{{ product.unit_name }}</em></span>
              </div>
              <div v-if="product.price_per_package_usd" class="tier-row tier-bulk">
                <span class="tier-label">{{ $t('product.pricePerPackage') }}</span>
                <span class="tier-value">${{ Number(product.price_per_package_usd).toFixed(2) }}<em v-if="product.pack_name">/{{ product.pack_name }}</em></span>
              </div>
            </div>

            <!-- 加购按钮 -->
            <div class="card-action" @click.stop>
              <div v-if="getCartQty(product.id) > 0" class="qty-control">
                <button class="qty-btn" @click="decrementCart(product)">−</button>
                <input
                  v-if="editingQtyId === product.id"
                  class="qty-input"
                  type="number"
                  :value="getCartQty(product.id)"
                  @blur="onQtyInputBlur(product, $event)"
                  @keyup.enter="$event.target.blur()"
                  ref="qtyInputRef"
                  autofocus
                />
                <span v-else class="qty-num" @click="startEditQty(product)">{{ getCartQty(product.id) }}</span>
                <button class="qty-btn" @click="incrementCart(product)">+</button>
              </div>
              <button v-else class="add-btn" :disabled="product.stock <= 0" @click="incrementCart(product)">
                {{ $t('product.addToCart') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部结算栏 -->
    <div v-if="cartStore.totalCount > 0" class="checkout-bar">
      <div class="cart-info" @click="goToCart">
        <div class="cart-icon-wrap">
          <van-icon name="shopping-cart-o" size="22" />
          <span class="cart-badge">{{ cartStore.totalCount }}</span>
        </div>
        <div class="cart-total">
          <span class="total-label">{{ $t('common.total') }}</span>
          <span class="total-price">${{ cartStore.totalPrice.toFixed(2) }}</span>
        </div>
      </div>
      <button class="checkout-btn" @click="goToCart">{{ $t('product.checkout') }}</button>
    </div>

    <!-- 联系弹窗 -->
    <van-popup
      v-model:show="contactVisible"
      position="bottom"
      teleport="body"
      round
      :style="{ padding: '24px 20px 32px' }"
    >
      <div class="contact-sheet">
        <div class="contact-title">{{ $t('common.contactUs') }}</div>
        <div v-if="!contactInfo.phone && !contactInfo.telegram && !contactInfo.whatsapp && !contactInfo.wechat" class="contact-empty">
          {{ $t('common.contactNotSet') }}
        </div>
        <a v-if="contactInfo.phone" :href="'tel:' + contactInfo.phone" class="contact-item">
          <van-icon name="phone-o" size="20" />
          <span>{{ contactInfo.phone }}</span>
        </a>
        <a v-if="contactInfo.whatsapp" :href="'https://wa.me/' + contactInfo.whatsapp.replace(/\D/g, '')" target="_blank" class="contact-item">
          <van-icon name="chat-o" size="20" />
          <span>WhatsApp: {{ contactInfo.whatsapp }}</span>
        </a>
        <a v-if="contactInfo.telegram" :href="'https://t.me/' + contactInfo.telegram.replace(/^@/, '')" target="_blank" class="contact-item">
          <van-icon name="send-gift-o" size="20" />
          <span>Telegram: {{ contactInfo.telegram }}</span>
        </a>
        <div v-if="contactInfo.wechat" class="contact-item" @click="copyWechat">
          <van-icon name="wechat-o" size="20" />
          <span>WeChat: {{ contactInfo.wechat }}</span>
        </div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showToast } from 'vant'
import { getProducts, getCategories, getPublicAnnouncements, getContactInfo } from '@/api'
import { useCartStore } from '@/stores/cart'
import { hapticFeedback } from '@/utils/device'
import { setLanguage, getCurrentLanguage } from '@/i18n'

const router = useRouter()
const cartStore = useCartStore()
const { t } = useI18n()

const loading = ref(true)
const products = ref([])
const searchKeyword = ref('')
const activeCat = ref('all')
const categories = ref([{ id: 'all', name: t('product.all') }])
const notices = ref([])
const currentLang = ref(getCurrentLanguage())
const promoActive = ref('new')

const toggleLang = () => {
  const order = ['zh', 'en', 'kh']
  const idx = order.indexOf(currentLang.value)
  const newLang = order[(idx + 1) % order.length]
  setLanguage(newLang)
  currentLang.value = newLang
  hapticFeedback('light')
}

const langLabel = computed(() => {
  const next = { zh: 'EN', en: 'ខ្មែរ', kh: '中' }
  return next[currentLang.value] || 'EN'
})

// 联系我
const contactVisible = ref(false)
const contactInfo = ref({ phone: '', telegram: '', whatsapp: '', wechat: '' })

const loadContactInfo = async () => {
  try {
    contactInfo.value = await getContactInfo()
  } catch {
    // 静默失败
  }
}

const copyWechat = () => {
  if (!contactInfo.value.wechat) return
  navigator.clipboard?.writeText(contactInfo.value.wechat).then(() => {
    showSuccessToast(t('common.wechatCopied'))
  }).catch(() => {
    showToast('WeChat: ' + contactInfo.value.wechat)
  })
}

const openTelegram = () => {
  if (!contactInfo.value.telegram) return
  hapticFeedback('light')
  const tg = contactInfo.value.telegram.replace(/^@/, '')
  window.open('https://t.me/' + tg, '_blank')
}

// 加载分类
const loadCategories = async () => {
  try {
    const data = await getCategories()
    categories.value = [
      { id: 'all', name: t('product.all') },
      ...data.map(c => ({ id: c.name, name: c.name })),
    ]
  } catch {
    // 回退：从商品中提取分类
  }
}

// 加载商品
const loadProducts = async () => {
  loading.value = true
  try {
    const data = await getProducts({ is_active: true })
    products.value = data
  } catch {
    showToast(t('product.loadFailed'))
  } finally {
    loading.value = false
  }
}

// 过滤商品
const filteredProducts = computed(() => {
  let list = products.value
  if (activeCat.value !== 'all') {
    list = list.filter(p => p.category === activeCat.value)
  }
  if (searchKeyword.value.trim()) {
    const kw = searchKeyword.value.toLowerCase()
    list = list.filter(p =>
      p.name.toLowerCase().includes(kw) ||
      (p.name_kh && p.name_kh.toLowerCase().includes(kw)) ||
      (p.name_en && p.name_en.toLowerCase().includes(kw)) ||
      (p.brand && p.brand.toLowerCase().includes(kw))
    )
  }
  // 推荐商品置顶
  return [...list].sort((a, b) => {
    if (a.is_featured && !b.is_featured) return -1
    if (!a.is_featured && b.is_featured) return 1
    return (a.sort_order || 0) - (b.sort_order || 0)
  })
})

// 新品推荐 — 仅管理员勾选推荐的商品
const newProducts = computed(() => {
  if (!products.value.length) return []
  return [...products.value]
    .filter(p => p.is_featured)
    .sort((a, b) => b.id - a.id)
    .slice(0, 10)
})

// 折扣专区 — 管理员勾选 is_discounted 的商品
const discountProducts = computed(() => {
  if (!products.value.length) return []
  return products.value.filter(p => p.is_discounted).slice(0, 20)
})

// 跑马灯速度统一：每件商品约 83px（75px卡片+8px间距），目标速度 40px/s
const MARQUEE_SPEED = 40
const CARD_PX = 83
const marqueeDuration = (count) => count <= 1 ? '0s' : `${Math.max(6, Math.round(count * CARD_PX / MARQUEE_SPEED))}s`
const newProductsDuration = computed(() => marqueeDuration(newProducts.value.length))
const discountProductsDuration = computed(() => marqueeDuration(discountProducts.value.length))

// 获取商品首图
const getProductImage = (product) => {
  return product.img1 || product.image_url || ''
}

// 购物车数量
const getCartQty = (id) => {
  const item = cartStore.items.find(i => i.id === id)
  return item ? item.quantity : 0
}

const incrementCart = (product) => {
  if (product.stock <= 0) return
  hapticFeedback('light')
  cartStore.addItem(product, 1)
}

const decrementCart = (product) => {
  const current = getCartQty(product.id)
  hapticFeedback('light')
  if (current <= 1) {
    cartStore.removeItem(product.id)
  } else {
    cartStore.updateQuantity(product.id, current - 1)
  }
}

// 点击数量进入编辑模式
const editingQtyId = ref(null)

const startEditQty = (product) => {
  editingQtyId.value = product.id
}

const onQtyInputBlur = (product, event) => {
  const val = parseInt(event.target.value)
  if (!isNaN(val) && val > 0) {
    const qty = Math.min(val, product.stock)
    cartStore.updateQuantity(product.id, qty)
  } else if (val === 0 || event.target.value === '') {
    cartStore.removeItem(product.id)
  }
  editingQtyId.value = null
}

// 品类切换
const switchCategory = (id) => {
  activeCat.value = id
  hapticFeedback('light')
}

const onSearch = () => {
  // 搜索时重置分类到全部，确保搜索结果不受分类过滤影响
  if (searchKeyword.value.trim()) {
    activeCat.value = 'all'
  }
}

// 跳转商品详情页
const showDetail = (product) => {
  hapticFeedback('medium')
  router.push({ name: 'ProductDetail', params: { id: product.id }, state: { product: JSON.stringify(product) } })
}

// 加载公告
const loadNotices = async () => {
  try {
    const data = await getPublicAnnouncements('notice')
    notices.value = data
  } catch {
    // 静默处理
  }
}

const goToCart = () => {
  hapticFeedback('medium')
  router.push('/m/cart')
}

onMounted(async () => {
  await Promise.all([loadCategories(), loadProducts(), loadNotices(), loadContactInfo()])
  // 从商品中补充分类
  if (categories.value.length <= 1) {
    const cats = new Set()
    products.value.forEach(p => { if (p.category) cats.add(p.category) })
    cats.forEach(c => categories.value.push({ id: c, name: c }))
  }
})
</script>

<style scoped>
.mobile-shop {
  min-height: var(--tg-viewport-height, 100vh);
  background: #f5f5f7;
  padding-bottom: calc(60px + var(--tg-content-safe-area-inset-bottom, env(safe-area-inset-bottom, 0px)));
  overflow: visible;
}

/* ===== 搜索栏 ===== */
.shop-header {
  background: #fff;
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
}

.header-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.search-box {
  flex: 1;
  display: flex;
  align-items: center;
  background: #f5f5f7;
  border-radius: 8px;
  padding: 0 10px;
  height: 36px;
  border: 1px solid transparent;
  transition: border-color 0.2s;
}
.search-box:focus-within {
  border-color: #c76b35;
  background: #fff;
}

.search-icon {
  color: #999;
  font-size: 15px;
  margin-right: 6px;
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 14px;
  color: #1a1a1a;
  min-width: 0;
}
.search-input::placeholder {
  color: #bfbfbf;
}

.clear-icon {
  color: #999;
  font-size: 15px;
  margin-left: 4px;
  flex-shrink: 0;
}

.contact-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  flex-shrink: 0;
  height: 36px;
  border: 1.5px solid #ddb7a0;
  border-radius: 8px;
  background: #fff6f0;
  color: #c76b35;
  font-size: 12px;
  font-weight: 600;
  padding: 0 12px;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.2s;
}

.telegram-header-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border: 1.5px solid #a8c4e2;
  border-radius: 8px;
  background: #eef4fb;
  color: #2b6cb0;
  cursor: pointer;
  transition: background 0.2s;
}
.telegram-header-btn:active {
  background: #d8e5f5;
}
.contact-btn:active {
  background: #ffede3;
}

.lang-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 44px;
  height: 36px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background: #fff;
  color: #666;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.lang-btn:active {
  background: #f5f5f5;
}

/* ===== 促销跑马灯（合并） ===== */
.promo-section {
  background: #fff;
  margin: 0;
  border-bottom: 1px solid #f0f0f0;
  padding: 6px 0;
}

.promo-tabs {
  display: flex;
  gap: 2px;
  padding: 0 12px 6px;
}

.promo-tab {
  font-size: 11px;
  color: #999;
  background: #f5f5f7;
  padding: 3px 10px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
}
.promo-tab.active {
  background: #2b2b2b;
  color: #fff;
}

.promo-marquee {
  overflow: hidden;
  padding: 0 0 2px 12px;
  -webkit-mask-image: linear-gradient(to right, #000 82%, transparent 100%);
  mask-image: linear-gradient(to right, #000 82%, transparent 100%);
}

@keyframes marqueeLeft {
  0%   { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}

.marquee-track {
  display: flex;
  width: max-content;
  gap: 0;
  animation-name: marqueeLeft;
  animation-timing-function: linear;
  animation-iteration-count: infinite;
}

.marquee-group {
  display: flex;
  gap: 6px;
}

.promo-card {
  flex-shrink: 0;
  width: 64px;
  border-radius: 6px;
  overflow: hidden;
  position: relative;
  background: #fafafa;
  cursor: pointer;
}
.promo-card img {
  width: 100%;
  height: 48px;
  object-fit: cover;
  display: block;
}
.promo-img-ph {
  width: 100%;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.promo-badge {
  position: absolute;
  top: 2px;
  left: 2px;
  background: linear-gradient(135deg, #ff4d4f, #ff7a45);
  color: #fff;
  font-size: 7px;
  font-weight: 700;
  padding: 1px 3px;
  border-radius: 2px;
  line-height: 1.2;
}
.dc-badge {
  background: linear-gradient(135deg, #f5a623, #e8222e);
}
.promo-card-price {
  font-size: 9px;
  font-weight: 700;
  color: #e8222e;
  padding: 1px 4px 3px;
  line-height: 1.2;
}

/* ===== 主体：左侧分类 + 右侧商品 ===== */
.main-body {
  display: flex;
  min-height: calc(100vh - 260px);
}

/* 左侧侧边栏 */
.cat-sidebar {
  width: 76px;
  flex-shrink: 0;
  background: #fafafa;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  border-right: 1px solid #f0f0f0;
  padding: 4px 0 12px;
  position: sticky;
  top: 0;
  align-self: flex-start;
  max-height: calc(100vh - 180px);
}

.cat-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 8px;
  font-size: 12px;
  color: #666;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
  border-left: 3px solid transparent;
}
.cat-item.active {
  color: #1a1a1a;
  font-weight: 600;
  background: #fff;
  border-left-color: #c76b35;
}

.cat-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: #ddd;
  flex-shrink: 0;
  transition: all 0.2s;
}
.cat-dot.active {
  background: #c76b35;
  width: 6px;
  height: 6px;
}

.cat-name {
  line-height: 1.3;
  word-break: break-all;
}

/* 右侧商品区域 */
.product-area {
  flex: 1;
  min-width: 0;
  padding: 6px;
}

/* ===== 商品网格 ===== */
.product-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 6px;
}

.product-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
  border: 1px solid #f0f0f0;
  transition: box-shadow 0.2s;
}
.product-card:active {
  box-shadow: 0 1px 6px rgba(0,0,0,0.08);
}

.card-image {
  position: relative;
  width: 100%;
  padding-top: 100%;
  background: #fafafa;
  overflow: hidden;
}
.card-image img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s;
}
.product-card:active .card-image img {
  transform: scale(1.05);
}
.image-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fafafa;
}

.badge-out,
.badge-low,
.badge-featured {
  position: absolute;
  font-size: 9px;
  padding: 1px 5px;
  border-radius: 2px;
  font-weight: 500;
  line-height: 1.4;
}
.badge-out {
  top: 5px;
  right: 5px;
  background: #fef2f2;
  color: #dc2626;
  border: 1px solid #dc2626;
}
.badge-low {
  top: 5px;
  right: 5px;
  background: #fff7ed;
  color: #ea580c;
  border: 1px solid #ea580c;
}
.badge-featured {
  top: 5px;
  left: 5px;
  background: #fff0f0;
  color: #e8222e;
  border: 1px solid #e8222e;
}

.card-name {
  padding: 6px 6px 0;
  font-size: 12px;
  font-weight: 600;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.3;
  min-height: 31px;
}

.card-price-row {
  padding: 2px 6px 0;
  display: flex;
  align-items: baseline;
  gap: 2px;
}
.price-val {
  font-size: 15px;
  font-weight: 700;
  color: #d44e4e;
}
.price-unit {
  font-size: 10px;
  color: #bbb;
}

/* 阶梯价 */
.card-tiers {
  padding: 2px 6px 0;
  display: flex;
  flex-direction: column;
  gap: 1px;
}
.tier-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 10px;
  background: #fafafa;
  border-radius: 3px;
  padding: 2px 5px;
}
.tier-label {
  color: #888;
}
.tier-value {
  font-weight: 600;
  color: #555;
}
.tier-value em {
  font-style: normal;
  color: #bbb;
}
.tier-bulk .tier-value {
  color: #1a4ed8;
}

.card-action {
  padding: 3px 6px 6px;
}

.add-btn {
  width: 100%;
  height: 28px;
  border: 1px solid #e0e0e0;
  background: #fff;
  color: #333;
  font-size: 12px;
  font-weight: 500;
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.15s;
}
.add-btn:active {
  background: #f5f5f5;
  transform: scale(0.97);
}
.add-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.qty-control {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0;
}
.qty-btn {
  width: 28px;
  height: 28px;
  border: 1px solid #e0e0e0;
  background: #fff;
  font-size: 15px;
  font-weight: 500;
  color: #333;
  border-radius: 5px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}
.qty-btn:active {
  background: #f0f0f0;
}
.qty-num {
  width: 34px;
  text-align: center;
  font-size: 13px;
  font-weight: 600;
  color: #1a1a1a;
  cursor: pointer;
}
.qty-input {
  width: 38px;
  text-align: center;
  font-size: 13px;
  font-weight: 600;
  border: 1px solid #c76b35;
  border-radius: 4px;
  outline: none;
  padding: 2px 0;
  -moz-appearance: textfield;
  appearance: textfield;
}
.qty-input::-webkit-inner-spin-button,
.qty-input::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

/* ===== 底部结算栏 ===== */
.checkout-bar {
  position: fixed;
  bottom: calc(50px + env(safe-area-inset-bottom, 0px));
  left: 0;
  right: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  background: #fff;
  border-top: 1px solid #f0f0f0;
  padding: 8px 12px;
  box-shadow: 0 -2px 8px rgba(0,0,0,0.04);
}

.cart-info {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
}

.cart-icon-wrap {
  position: relative;
}

.cart-badge {
  position: absolute;
  top: -6px;
  right: -8px;
  background: #d44e4e;
  color: #fff;
  font-size: 10px;
  min-width: 16px;
  height: 16px;
  line-height: 16px;
  text-align: center;
  border-radius: 8px;
  padding: 0 4px;
  font-weight: 600;
}

.total-label {
  font-size: 12px;
  color: #999;
}

.total-price {
  font-size: 17px;
  font-weight: 700;
  color: #d44e4e;
  margin-left: 4px;
}

.checkout-btn {
  padding: 0 24px;
  height: 36px;
  background: #2b2b2b;
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: opacity 0.2s;
}
.checkout-btn:active {
  opacity: 0.85;
  transform: scale(0.97);
}

/* ===== 联系弹窗 ===== */
.contact-sheet {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.contact-title {
  font-size: 17px;
  font-weight: 700;
  color: #1a1a1a;
  margin-bottom: 4px;
  text-align: center;
}
.contact-empty {
  text-align: center;
  color: #aaa;
  font-size: 14px;
  padding: 12px 0;
}
.contact-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  background: #f8f8f8;
  border-radius: 10px;
  color: #1a1a1a;
  font-size: 15px;
  text-decoration: none;
  border: 1px solid #eee;
}

/* ===== 骨架屏 ===== */
.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 6px;
}
.skeleton-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #f0f0f0;
}
.skeleton-img {
  width: 100%;
  padding-top: 100%;
}
.skeleton-name {
  height: 14px;
  margin: 8px;
  width: 60%;
  border-radius: 2px;
}
.skeleton-price {
  height: 16px;
  margin: 0 8px 8px;
  width: 40%;
  border-radius: 2px;
}
.shimmer {
  background: linear-gradient(90deg, #f0f0f0 0%, #f8f8f8 50%, #f0f0f0 100%);
  background-size: 200% 100%;
  animation: shimmer 1.5s ease-in-out infinite;
}
@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}
</style>
