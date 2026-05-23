<template>
  <!-- 桌面端外层容器：根据环境自适应 -->
  <div class="desktop-wrapper" :class="{ 'tg-frame-mode': isTgContext }">
    <div class="mobile-layout"
      :class="{ 'tg-frame-mode': isTgContext }"
      :style="isTgContext ? { paddingTop: tgTopPadding + 'px' } : {}"
    >
      <!-- 内容区（可滚动） -->
      <div class="mobile-content">
        <router-view v-slot="{ Component }">
          <keep-alive>
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </div>

      <!-- 底部导航栏（始终固定在底部） -->
      <van-tabbar v-model="activeTab" @change="handleTabChange" :fixed="false" :placeholder="false">
        <van-tabbar-item icon="shop-o" to="/m/shop">{{ $t('nav.shop') }}</van-tabbar-item>
        <van-tabbar-item icon="shopping-cart-o" :badge="cartStore.totalCount || ''">
          {{ $t('nav.cart') }}
        </van-tabbar-item>
        <van-tabbar-item icon="orders-o" to="/m/orders">{{ $t('nav.orders') }}</van-tabbar-item>
        <van-tabbar-item icon="user-o" to="/m/profile">{{ $t('nav.profile') }}</van-tabbar-item>
      </van-tabbar>
    </div>

    <!-- 强制完善账号弹窗（商户首次登录必须设置手机号+密码） -->
    <div v-if="needsSetup" class="setup-overlay">
      <div class="setup-card">
        <h3 class="setup-title">{{ $t('setup.title') }}</h3>
        <p class="setup-tip">{{ $t('setup.tip') }}</p>
        <van-cell-group inset>
          <van-field
            v-model="setupForm.phone"
            :placeholder="$t('setup.phonePlaceholder')"
            :disabled="setupLoading"
            clearable
          />
          <van-field
            v-model="setupForm.password"
            type="password"
            :placeholder="$t('setup.passwordPlaceholder')"
            :disabled="setupLoading"
          />
          <van-field
            v-model="setupForm.confirmPassword"
            type="password"
            :placeholder="$t('setup.confirmPasswordPlaceholder')"
            :disabled="setupLoading"
            @keyup.enter="handleSetup"
          />
        </van-cell-group>
        <p v-if="setupError" class="setup-error">{{ setupError }}</p>
        <van-button type="primary" block :loading="setupLoading" style="margin-top:16px" @click="handleSetup">
          {{ $t('common.confirm') }}
        </van-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { showToast } from 'vant'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import { hapticFeedback } from '@/utils/device'
import { useI18n } from 'vue-i18n'
import { setupCredentials } from '@/api'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const cartStore = useCartStore()
const { t } = useI18n()

const activeTab = ref(0)

// 商户首次登录 —— 强制设置手机号+密码
const needsSetup = computed(() => {
  const u = userStore.userInfo
  return u?.role === 'merchant' && !u?.phone
})
const setupForm = reactive({ phone: '', password: '', confirmPassword: '' })
const setupLoading = ref(false)
const setupError = ref('')

const handleSetup = async () => {
  setupError.value = ''
  if (!setupForm.phone.trim()) { setupError.value = t('setup.phoneRequired'); return }
  if (!setupForm.password || setupForm.password.length < 6) { setupError.value = t('setup.passwordMinLength'); return }
  if (setupForm.password !== setupForm.confirmPassword) { setupError.value = t('setup.passwordMismatch'); return }
  setupLoading.value = true
  try {
    const user = await setupCredentials({ phone: setupForm.phone.trim(), password: setupForm.password })
    userStore.userInfo = { ...userStore.userInfo, ...user }
    showToast(t('setup.success'))
  } catch (e) {
    setupError.value = e?.response?.data?.detail || t('common.requestFailed')
  } finally {
    setupLoading.value = false
  }
}

// 只要 window.Telegram.WebApp 对象存在就算 Telegram 环境
// 不用 initData（可能为空字符串）做判断
const isTgContext = ref(typeof window !== 'undefined' && !!window.Telegram?.WebApp)

// Telegram 标题栏/浮动按钮占用的顶部高度
// 初始值 72px（状态栏~24px + Telegram操作栏~48px），onMounted 后用 API 精确值替换
const tgTopPadding = ref(isTgContext.value ? 72 : 0)

onMounted(() => {
  const tg = window.Telegram?.WebApp
  if (!tg) return

  tg.ready()
  tg.expand()
  tg.setHeaderColor?.('#ffffff')
  tg.setBackgroundColor?.('#f5f5f5')

  // 直接从 JS API 读取 Telegram 占用的顶部高度
  // contentSafeAreaInset.top = Telegram 标题/按钮覆盖内容的高度
  // safeAreaInset.top = 系统状态栏高度
  // 两者相加才是内容需要下移的总量
  const updatePadding = () => {
    const content = tg.contentSafeAreaInset?.top ?? 0
    const safe = tg.safeAreaInset?.top ?? 0
    // 若 API 返回 0（旧版 Telegram），保持 72px 兜底值不变
    const apiVal = content + safe
    if (apiVal > 0) tgTopPadding.value = apiVal
  }
  updatePadding()
  // 窗口大小/方向变化时重新计算
  tg.onEvent?.('safe_area_changed', updatePadding)
  tg.onEvent?.('content_safe_area_changed', updatePadding)
  tg.onEvent?.('viewport_changed', updatePadding)
})

// 根据路由更新选中状态
watch(() => route.path, (newPath) => {
  if (newPath.includes('/m/shop')) activeTab.value = 0
  else if (newPath.includes('/m/cart')) activeTab.value = 1
  else if (newPath.includes('/m/orders')) activeTab.value = 2
  else if (newPath.includes('/m/profile')) activeTab.value = 3
}, { immediate: true })

const handleTabChange = (index) => {
  hapticFeedback('light')
  const routes = ['/m/shop', '/m/cart', '/m/orders', '/m/profile']
  router.push(routes[index])
}
</script>

<style scoped>
/* ── 外层容器 ── */
.desktop-wrapper {
  height: 100vh;
  height: 100dvh;
  background: #f0f2f5;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  overflow: hidden;
}

/* ── 主容器（所有设备） ── */
.mobile-layout {
  width: 100%;
  max-width: 520px;
  height: 100vh;
  height: 100dvh;
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 0 24px rgba(0, 0, 0, 0.12);
}

/* ── 内容滚动区 ── */
.mobile-content {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  overscroll-behavior-y: contain;
  /* 必须有 position + z-index，否则 iOS/Telegram 中 van-tabbar(z-index:1)
     会压盖内容区的所有固定定位子元素（弹窗/结算栏等） */
  position: relative;
  z-index: 2;
}

/* ════════════════════════════════════════
   桌面端 (≥600px) — 分两种模式
   ════════════════════════════════════════ */

/* 模式1：电脑 Telegram → 仿手机框 */
@media (min-width: 600px) {
  .desktop-wrapper.tg-frame-mode {
    background: linear-gradient(160deg, #0d1b2a 0%, #1b2d3e 45%, #0f3460 100%);
  }
  .mobile-layout.tg-frame-mode {
    max-width: 520px;
    height: calc(100vh - 20px);
    height: calc(100dvh - 20px);
    margin-top: 20px;
    border-radius: 32px 32px 0 0;
    overflow: hidden;
    box-shadow:
      0 0 0 6px rgba(255, 255, 255, 0.07),
      0 0 0 9px rgba(255, 255, 255, 0.04),
      0 32px 80px rgba(0, 0, 0, 0.55);
  }
}

/* 模式2：电脑浏览器 → 居中仿手机框 */
@media (min-width: 600px) {
  .desktop-wrapper:not(.tg-frame-mode) {
    background: linear-gradient(160deg, #1a1a2e 0%, #16213e 45%, #0f3460 100%);
  }
  .mobile-layout:not(.tg-frame-mode) {
    max-width: 520px;
    height: calc(100vh - 20px);
    height: calc(100dvh - 20px);
    margin-top: 20px;
    border-radius: 32px 32px 0 0;
    box-shadow:
      0 0 0 6px rgba(255, 255, 255, 0.07),
      0 0 0 9px rgba(255, 255, 255, 0.04),
      0 32px 80px rgba(0, 0, 0, 0.55);
  }
}

/* ══ 安全区域底边距（TG + 底部Tabbar） ══ */
.mobile-content {
  padding-bottom: env(safe-area-inset-bottom, 0px);
}

/* ── 底部导航栏（深色渐变，与 Shop 头部一致） ── */
:deep(.van-tabbar) {
  flex-shrink: 0;
  width: 100%;
  position: relative;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%) !important;
}
:deep(.van-tabbar-item) {
  color: rgba(255,255,255,0.5) !important;
  transition: color 0.2s;
}
:deep(.van-tabbar-item--active) {
  color: #fff !important;
}
:deep(.van-tabbar-item__icon) {
  font-size: 20px;
}

/* ── 强制完善账号遮罩 ── */
.setup-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}
.setup-card {
  background: #fff;
  border-radius: 16px;
  padding: 28px 20px 20px;
  width: 100%;
  max-width: 360px;
}
.setup-title {
  font-size: 17px;
  font-weight: 600;
  text-align: center;
  margin: 0 0 10px;
  color: #1a1a1a;
}
.setup-tip {
  font-size: 13px;
  color: #888;
  text-align: center;
  margin: 0 0 16px;
  line-height: 1.6;
}
.setup-error {
  color: #ee0a24;
  font-size: 13px;
  text-align: center;
  margin: 10px 0 0;
}
</style>
