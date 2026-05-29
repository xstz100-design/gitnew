<template>
  <div class="pd-page">
    <div v-if="loading" class="pd-loading">
      <van-loading size="32" vertical />
    </div>

    <template v-else-if="product">
      <!-- 图片区（含悬浮返回按钮） -->
      <div class="pd-img-area">
        <button class="pd-back-btn" @click="goBack">
          <van-icon name="arrow-left" size="20" color="#fff" />
        </button>
        <van-swipe :autoplay="0" indicator-color="rgba(255,255,255,0.85)">
          <van-swipe-item v-for="(img, idx) in productImages" :key="idx">
            <img :src="img" class="pd-img" @click="openViewer(idx)" />
          </van-swipe-item>
          <van-swipe-item v-if="productImages.length === 0">
            <div class="pd-img-empty"><van-icon name="photo-o" size="56" color="#ccc" /></div>
          </van-swipe-item>
        </van-swipe>
        <div class="pd-img-fade" />
      </div>

      <!-- 全屏图片查看 -->
      <div v-if="viewerVisible" class="img-viewer" @click="viewerVisible = false">
        <button class="img-viewer-close" @click.stop="viewerVisible = false">
          <van-icon name="cross" size="22" color="#fff" />
        </button>
        <van-swipe :autoplay="0" :initial-swipe="viewerIndex" style="height:100%;width:100%">
          <van-swipe-item v-for="(img, idx) in productImages" :key="idx"
            style="display:flex;align-items:center;justify-content:center;">
            <img :src="img" style="width:100%;max-height:100vh;object-fit:contain;" @click.stop />
          </van-swipe-item>
        </van-swipe>
      </div>

      <div class="pd-body">
        <!-- 标题 + 价格 -->
        <div class="pd-title-row">
          <h1 class="pd-name">{{ product.name }}</h1>
          <div class="pd-price-badge">
            ${{ Number(product.price_usd).toFixed(2) }}<span class="pd-price-unit">/{{ product.unit }}</span>
            <span class="pd-khr">{{ khrLabel(product.price_usd) }}</span>
          </div>
        </div>
        <p v-if="product.name_kh" class="pd-kh">{{ product.name_kh }}</p>

        <!-- 购买规格表 -->
        <div class="purchase-table">
          <div class="pt-header">
            <span>{{ $t('product.tableSpec') }}</span>
            <span>{{ $t('product.tableUnitPrice') }}</span>
            <span style="text-align:right">{{ $t('product.tableQty') }}</span>
          </div>
          <div class="pt-row">
            <span class="pt-spec">{{ product.unit_name || product.unit || $t('product.unitFallback') }}</span>
            <span class="pt-price">
              ${{ Number(product.price_usd).toFixed(2) }}<em>/{{ product.unit || $t('product.unitFallback') }}</em>
            </span>
            <div class="pt-stepper">
              <van-stepper v-model="qtyDefault" :min="0" :max="product.stock" integer button-size="26px" input-width="40px" />
            </div>
          </div>
          <div v-if="product.pieces_per_package" class="pt-row">
            <span class="pt-spec">
              {{ product.pack_name || $t('product.caseFallback') }}
              <span class="pt-spec-hint">({{ product.pieces_per_package }}{{ product.unit }}/{{ product.pack_name || $t('product.caseFallback') }})</span>
            </span>
            <span class="pt-price">
              <span v-if="product.price_per_package_usd" class="pt-orig">
                ${{ (Number(product.price_usd) * Number(product.pieces_per_package)).toFixed(2) }}
              </span>
              ${{ product.price_per_package_usd ? Number(product.price_per_package_usd).toFixed(2) : (Number(product.price_usd) * Number(product.pieces_per_package)).toFixed(2) }}<em>/{{ product.pack_name || $t('product.caseFallback') }}</em>
            </span>
            <div class="pt-stepper">
              <van-stepper v-model="qtyPackage" :min="0" integer button-size="26px" input-width="40px" />
            </div>
          </div>
          <div v-if="product.price_per_case_usd" class="pt-row">
            <span class="pt-spec">
              {{ $t('product.caseRow') }}
              <span v-if="product.unit_per_case" class="pt-spec-hint">({{ product.unit_per_case }}{{ product.unit }}/{{ $t('product.caseFallback') }})</span>
            </span>
            <span class="pt-price">
              ${{ Number(product.price_per_case_usd).toFixed(2) }}<em>/{{ $t('product.caseFallback') }}</em>
            </span>
            <div class="pt-stepper">
              <van-stepper v-model="qtyCase" :min="0" integer button-size="26px" input-width="40px" />
            </div>
          </div>
        </div>

        <!-- 商品属性 -->
        <div class="pd-meta">
          <span v-if="product.specs">{{ $t('product.specs') }}：{{ product.specs }}</span>
          <span v-if="product.unit_weight_value">重量：{{ product.unit_weight_value }}{{ product.unit_weight_unit || 'g' }}</span>
          <span>{{ $t('product.stock') }}：{{ product.stock }} {{ product.unit }}</span>
          <span>{{ $t('product.category') }}：{{ product.category || $t('product.uncategorized') }}</span>
          <span v-if="product.brand">{{ $t('product.brand') }}：{{ product.brand }}</span>
          <span v-if="product.country_of_origin">{{ $t('product.countryOfOrigin') }}：{{ product.country_of_origin }}</span>
          <span v-if="product.packing_format">{{ $t('product.packingFormat') }}：{{ product.packing_format }}</span>
          <span v-if="product.shelf_life_days">{{ $t('product.shelfLife') }}：{{ product.shelf_life_days }} {{ $t('product.daysUnit') }}</span>
          <span v-if="product.production_date">{{ $t('product.productionDate') }}：{{ product.production_date.substring(0, 10) }}</span>
          <span v-if="product.expiry_date" :class="isNearExpiry(product.expiry_date) ? 'meta-warn' : ''" style="width:100%;flex-basis:100%">
            {{ $t('product.expiryDate') }}：{{ product.expiry_date.substring(0, 10) }}
            <span v-if="isNearExpiry(product.expiry_date)">⚠️ {{ $t('product.nearExpiry') }}</span>
          </span>
        </div>

        <!-- 描述 -->
        <div v-if="product.description" class="pd-desc">
          <div class="pd-desc-title">{{ $t('product.description') }}</div>
          <div class="pd-desc-body">{{ product.description }}</div>
        </div>

        <!-- 最小订货量 -->
        <div v-if="product.min_order_qty > 1" class="pd-notice">
          {{ $t('product.minOrderQty') }}：{{ product.min_order_qty }} {{ product.unit }}
        </div>

      </div>

      <!-- 底部操作栏 -->
      <div class="pd-footer">
        <button class="pd-share-btn" :disabled="generatingShare" @click="shareProduct">
          <van-loading v-if="generatingShare" size="18" color="#555" />
          <van-icon v-else name="share-o" size="20" />
        </button>
        <button
          class="pd-add-btn"
          :disabled="product.stock <= 0 || (!qtyDefault && !qtyPackage && !qtyCase)"
          @click="handleAddToCart"
        >
          {{ $t('product.addToCart') }}
        </button>
      </div>
    </template>

    <van-empty v-else :description="$t('common.noData')" style="margin-top:60px" />
  </div>

</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showToast } from 'vant'
import { getProduct } from '@/api'
import { useCartStore } from '@/stores/cart'
import { hapticFeedback } from '@/utils/device'
import { khrLabel } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const cartStore = useCartStore()
const { t } = useI18n()

const loading = ref(true)
const product = ref(null)
const qtyDefault = ref(0)
const qtyPackage = ref(0)
const qtyCase = ref(0)
const viewerVisible = ref(false)
const viewerIndex = ref(0)
const generatingShare = ref(false)

const productImages = computed(() => {
  if (!product.value) return []
  const p = product.value
  const imgs = [p.img1, p.img2, p.img3, p.img4, p.img5].filter(Boolean)
  if (!imgs.length && p.image_url) imgs.push(p.image_url)
  return imgs
})

const isNearExpiry = (d) => {
  if (!d) return false
  const diff = new Date(d).getTime() - Date.now()
  return diff > 0 && diff < 30 * 24 * 60 * 60 * 1000
}

const goBack = () => {
  if (window.history.length > 1) router.back()
  else router.push('/m/shop')
}

const openViewer = (idx) => { viewerIndex.value = idx; viewerVisible.value = true }

const handleAddToCart = () => {
  const p = product.value
  let added = false
  if (qtyDefault.value > 0) { cartStore.addItem(p, qtyDefault.value, 'default'); added = true }
  if (qtyPackage.value > 0 && p.pieces_per_package) { cartStore.addItem(p, qtyPackage.value, 'package'); added = true }
  if (qtyCase.value > 0 && p.price_per_case_usd) { cartStore.addItem(p, qtyCase.value, 'case'); added = true }
  if (added) { hapticFeedback('success'); showSuccessToast(t('product.addedToCart')); qtyDefault.value = 0; qtyPackage.value = 0; qtyCase.value = 0 }
  else showToast(t('product.selectQtyFirst'))
}

// ── Canvas 商品卡片生成 ──
const loadImgEl = (url) => new Promise((resolve) => {
  if (!url) { resolve(null); return }
  const img = new Image()
  img.crossOrigin = 'anonymous'
  const timer = setTimeout(() => resolve(null), 4000)
  img.onload = () => { clearTimeout(timer); resolve(img) }
  img.onerror = () => { clearTimeout(timer); resolve(null) }
  img.src = url + (url.includes('?') ? '&' : '?') + '_c=1'
})

const rr = (ctx, x, y, w, h, r) => {
  ctx.beginPath()
  ctx.moveTo(x + r, y); ctx.lineTo(x + w - r, y); ctx.arcTo(x + w, y, x + w, y + r, r)
  ctx.lineTo(x + w, y + h - r); ctx.arcTo(x + w, y + h, x + w - r, y + h, r)
  ctx.lineTo(x + r, y + h); ctx.arcTo(x, y + h, x, y + h - r, r)
  ctx.lineTo(x, y + r); ctx.arcTo(x, y, x + r, y, r)
  ctx.closePath()
}

const generateCard = async () => {
  const W = 600, H = 840
  const c = document.createElement('canvas')
  c.width = W; c.height = H
  const ctx = c.getContext('2d')
  const font = '"PingFang SC","Microsoft YaHei",Arial,sans-serif'
  const IMG_H = 380

  // 白底
  ctx.fillStyle = '#fff'; ctx.fillRect(0, 0, W, H)

  // 商品图片
  const img = await loadImgEl(productImages.value[0] || '')
  if (img) {
    const sa = img.width / img.height, da = W / IMG_H
    let sx = 0, sy = 0, sw = img.width, sh = img.height
    if (sa > da) { sw = img.height * da; sx = (img.width - sw) / 2 }
    else { sh = img.width / da; sy = (img.height - sh) / 2 }
    try { ctx.drawImage(img, sx, sy, sw, sh, 0, 0, W, IMG_H) } catch {}
  } else {
    ctx.fillStyle = '#f0f0f0'; ctx.fillRect(0, 0, W, IMG_H)
    ctx.font = `80px ${font}`; ctx.textAlign = 'center'; ctx.fillStyle = '#ccc'
    ctx.fillText('🛒', W / 2, IMG_H / 2 + 28)
  }

  // 顶部深色渐变（品牌区）
  const tg = ctx.createLinearGradient(0, 0, 0, 90)
  tg.addColorStop(0, 'rgba(26,26,46,0.92)'); tg.addColorStop(1, 'rgba(26,26,46,0)')
  ctx.fillStyle = tg; ctx.fillRect(0, 0, W, 90)

  // 品牌名
  ctx.fillStyle = '#fff'; ctx.font = `bold 28px ${font}`; ctx.textAlign = 'left'
  ctx.fillText('东方优选', 22, 48)

  // 底部白色渐变过渡
  const bg = ctx.createLinearGradient(0, IMG_H - 80, 0, IMG_H)
  bg.addColorStop(0, 'rgba(255,255,255,0)'); bg.addColorStop(1, '#fff')
  ctx.fillStyle = bg; ctx.fillRect(0, IMG_H - 80, W, 80)

  let y = IMG_H + 8
  const p = product.value

  // 商品名
  ctx.fillStyle = '#1a1a1a'; ctx.font = `bold 30px ${font}`; ctx.textAlign = 'left'
  const nm = p.name.length > 20 ? p.name.slice(0, 20) + '…' : p.name
  ctx.fillText(nm, 22, y + 36); y += 50

  // 高棉文
  if (p.name_kh) {
    ctx.fillStyle = '#888'; ctx.font = `22px ${font}`
    const kh = p.name_kh.length > 24 ? p.name_kh.slice(0, 24) + '…' : p.name_kh
    ctx.fillText(kh, 22, y + 26); y += 40
  }
  y += 6

  // 价格盒子
  rr(ctx, 22, y, 230, 62, 10); ctx.fillStyle = '#fff0f0'; ctx.fill()
  ctx.strokeStyle = '#fcc'; ctx.lineWidth = 1.5; rr(ctx, 22, y, 230, 62, 10); ctx.stroke()
  const ps = `$${Number(p.price_usd).toFixed(2)}`
  ctx.fillStyle = '#d44e4e'; ctx.font = `bold 36px ${font}`; ctx.textAlign = 'left'
  ctx.fillText(ps, 34, y + 44)
  const pw = ctx.measureText(ps).width
  ctx.fillStyle = '#aaa'; ctx.font = `20px ${font}`; ctx.fillText(`/${p.unit}`, 34 + pw + 4, y + 44)
  y += 80

  // 属性
  const attrs = []
  if (p.specs) attrs.push(`规格: ${p.specs}`)
  attrs.push(`库存: ${p.stock} ${p.unit}`)
  if (p.category) attrs.push(`分类: ${p.category}`)
  if (p.brand) attrs.push(`品牌: ${p.brand}`)
  ctx.fillStyle = '#666'; ctx.font = `21px ${font}`; ctx.textAlign = 'left'
  for (const a of attrs.slice(0, 3)) {
    const at = a.length > 28 ? a.slice(0, 28) + '…' : a
    ctx.fillText(at, 22, y + 26); y += 36
  }

  // 底部分割线 + CTA
  ctx.strokeStyle = '#eee'; ctx.lineWidth = 1
  ctx.beginPath(); ctx.moveTo(22, H - 85); ctx.lineTo(W - 22, H - 85); ctx.stroke()
  ctx.fillStyle = '#555'; ctx.font = `22px ${font}`; ctx.textAlign = 'center'
  ctx.fillText('点击下单 · Click to Order', W / 2, H - 52)
  ctx.fillStyle = '#1d4ed8'; ctx.font = `19px ${font}`
  ctx.fillText(window.location.origin + '/m/shop', W / 2, H - 24)

  return c
}

const shareProduct = async () => {
  if (!product.value || generatingShare.value) return
  generatingShare.value = true
  try {
    const canvas = await generateCard()
    const blob = await new Promise((res) => canvas.toBlob(res, 'image/png'))
    const file = new File([blob], `${product.value.name}.png`, { type: 'image/png' })

    if (navigator.share && navigator.canShare?.({ files: [file] })) {
      await navigator.share({ files: [file], title: product.value.name, text: `$${product.value.price_usd}/${product.value.unit}` })
    } else if (navigator.share) {
      await navigator.share({ title: product.value.name, text: `${product.value.name}\n$${product.value.price_usd}/${product.value.unit}`, url: window.location.href })
    } else {
      // 桌面端：直接下载图片
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob); a.download = `${product.value.name}.png`; a.click()
      URL.revokeObjectURL(a.href)
      showSuccessToast(t('product.cardSaved'))
    }
  } catch (e) {
    if (e?.name === 'AbortError') return
    // canvas 被 CORS 污染，退回纯文本分享
    try {
      if (navigator.share) await navigator.share({ title: product.value.name, text: `${product.value.name}\n$${product.value.price_usd}/${product.value.unit}`, url: window.location.href })
      else { await navigator.clipboard?.writeText(`${product.value.name}\n$${product.value.price_usd}/${product.value.unit}\n${window.location.href}`); showSuccessToast(t('product.linkCopied')) }
    } catch {}
  } finally {
    generatingShare.value = false
  }
}

onMounted(async () => {
  const st = history.state?.product
  if (st) {
    try { product.value = typeof st === 'string' ? JSON.parse(st) : st; loading.value = false; return } catch {}
  }
  try { product.value = await getProduct(route.params.id) }
  catch { showToast(t('product.loadFailed')) }
  finally { loading.value = false }
})
</script>

<style scoped>
.pd-page {
  min-height: 100%;
  background: #fff;
  padding-bottom: calc(80px + 50px + env(safe-area-inset-bottom, 0px));
}

.pd-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 60vh;
}

/* ── 图片 ── */
.pd-img-area {
  position: relative;
  background: #f0f0f0;
}

/* 悬浮返回按钮（叠在图片左上角） */
.pd-back-btn {
  position: absolute;
  top: 10px;
  left: 12px;
  z-index: 10;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.38);
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

:deep(.pd-img-area .van-swipe) {
  height: clamp(260px, 55vw, 380px);
}

.pd-img {
  width: 100%;
  height: clamp(260px, 55vw, 380px);
  object-fit: cover;
  cursor: pointer;
  display: block;
}

.pd-img-empty {
  height: clamp(260px, 55vw, 380px);
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
}

.pd-img-fade {
  position: absolute;
  bottom: 0; left: 0; right: 0;
  height: 60px;
  background: linear-gradient(to bottom, transparent, rgba(247,247,247,0.96));
  pointer-events: none;
}

/* ── 全屏图片查看 ── */
.img-viewer {
  position: fixed;
  inset: 0;
  background: #000;
  z-index: 9999;
}

.img-viewer-close {
  position: absolute;
  top: 16px; right: 16px;
  z-index: 10001;
  background: rgba(255,255,255,0.15);
  border: none;
  border-radius: 50%;
  width: 40px; height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

/* ── 正文 ── */
.pd-body {
  padding: 14px 16px 0;
  background: #fff;
}

.pd-title-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 4px;
}

.pd-name {
  flex: 1;
  font-size: 18px;
  font-weight: 700;
  color: #1a1a1a;
  margin: 0;
  line-height: 1.4;
}

.pd-price-badge {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
  font-size: 20px;
  font-weight: 800;
  color: #d44e4e;
  white-space: nowrap;
  padding: 4px 10px;
  background: #fff0f0;
  border-radius: 6px;
  border: 1px solid #fdd;
}

.pd-price-unit { font-size: 13px; font-weight: 400; color: #999; }
.pd-khr { font-size: 11px; font-weight: 400; color: #bbb; }

.pd-kh { font-size: 13px; color: #999; margin: 0 0 14px; }

/* ── 规格表 ── */
.purchase-table {
  margin: 12px 0;
  border: 1px solid #eef0f3;
  border-radius: 10px;
  overflow: hidden;
}

.pt-header {
  display: grid;
  grid-template-columns: minmax(60px, 1fr) minmax(80px, 1.2fr) minmax(90px, 1fr);
  background: #f8fafc;
  padding: 8px 10px;
  font-size: 12px;
  color: #8c8c8c;
  font-weight: 500;
}

.pt-row {
  display: grid;
  grid-template-columns: minmax(60px, 1fr) minmax(80px, 1.2fr) minmax(90px, 1fr);
  align-items: center;
  padding: 10px 10px;
  border-top: 1px solid #eef0f3;
}

.pt-spec { font-size: 13px; color: #333; font-weight: 500; }
.pt-spec-hint { display: block; font-size: 11px; color: #999; margin-top: 2px; }

.pt-price { font-size: 14px; font-weight: 700; color: #d44e4e; }
.pt-price em { font-style: normal; font-size: 11px; font-weight: 400; color: #999; }
.pt-orig { font-size: 11px; color: #bbb; text-decoration: line-through; margin-right: 3px; }

.pt-stepper { display: flex; justify-content: flex-end; }

:deep(.purchase-table .van-stepper__input) {
  background: transparent;
  font-size: 15px;
  font-weight: 600;
  min-width: 32px !important;
  width: 40px !important;
}

:deep(.purchase-table .van-stepper__minus),
:deep(.purchase-table .van-stepper__plus) {
  width: 24px !important;
  height: 24px !important;
}

/* ── 属性 ── */
.pd-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 16px;
  font-size: 13px;
  color: #666;
  margin-bottom: 14px;
  padding-bottom: 14px;
  border-bottom: 1px solid #f0f0f0;
  line-height: 1.7;
}

.meta-warn { color: #d97706; font-weight: 600; }

/* ── 描述 ── */
.pd-desc {
  margin-bottom: 14px;
  padding-bottom: 14px;
  border-bottom: 1px solid #f0f0f0;
}

.pd-desc-title { font-size: 14px; font-weight: 600; color: #1a1a1a; margin-bottom: 6px; }
.pd-desc-body { font-size: 13px; color: #666; line-height: 1.6; white-space: pre-wrap; }

.pd-notice {
  font-size: 13px;
  color: #ea580c;
  background: #fff7ed;
  padding: 8px 12px;
  border-radius: 6px;
  margin-bottom: 14px;
}

/* ── 底部操作栏（坐在 tabbar 上方） ── */
.pd-footer {
  position: fixed;
  /* tabbar 高度约 50px，van-tabbar 在 MobileLayout 内是 relative 定位，
     但 fixed 元素以视口为基准，所以 bottom=50px 即可贴住 tabbar 上沿 */
  bottom: calc(50px + env(safe-area-inset-bottom, 0px));
  left: 50%;
  transform: translateX(-50%);
  width: 100%;
  max-width: 520px;
  z-index: 100;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 16px;
  background: #fff;
  border-top: 1px solid #f0f0f0;
  box-sizing: border-box;
}

.pd-share-btn {
  flex-shrink: 0;
  width: 48px; height: 48px;
  border: 1.5px solid #e0e0e0;
  border-radius: 10px;
  background: #f8f8f8;
  color: #555;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.15s;
}

.pd-share-btn:active,
.pd-share-btn:disabled { background: #efefef; }

.pd-add-btn {
  flex: 1;
  height: 48px;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
  color: #fff;
  font-size: 16px;
  font-weight: 500;
  border: none;
  border-radius: 10px;
  cursor: pointer;
}

.pd-add-btn:active { opacity: 0.85; }
.pd-add-btn:disabled { opacity: 0.4; }

.qp-popup { padding-bottom: env(safe-area-inset-bottom, 0px); }
.qp-title {
  font-size: 16px;
  font-weight: 600;
  text-align: center;
  padding: 20px 16px 12px;
  color: #1a1a1a;
}
</style>
