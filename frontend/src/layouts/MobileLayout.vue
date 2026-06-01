<template>
  <!-- 桌面端外层容器：根据环境自适应 -->
  <div class="desktop-wrapper" :class="{ 'tg-frame-mode': isTgContext }">
    <div class="mobile-layout"
      :class="{ 'tg-frame-mode': isTgContext }"
      :style="isTgContext ? { paddingTop: tgTopPadding + 'px' } : {}"
    >
      <!-- 全局顶部 Logo 栏 -->
      <div class="global-header">
        <img src="/images/logo-main.png" class="global-logo" alt="logo" />
      </div>

      <!-- 内容区（可滚动） -->
      <div class="mobile-content">
        <router-view v-slot="{ Component }">
          <keep-alive>
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </div>

      <!-- fixed 定位：无 transform 干扰，bottom:0 = 视口底部（已排除浏览器导航栏） -->
      <van-tabbar v-model="activeTab" @change="handleTabChange" :fixed="true" :placeholder="true" safe-area-inset-bottom>
        <van-tabbar-item icon="shop-o" to="/m/shop">{{ $t('nav.shop') }}</van-tabbar-item>
        <van-tabbar-item icon="shopping-cart-o" :badge="cartStore.totalCount || ''">
          {{ $t('nav.cart') }}
        </van-tabbar-item>
        <van-tabbar-item icon="user-o" to="/m/profile">{{ $t('nav.profile') }}</van-tabbar-item>
      </van-tabbar>
    </div>

  </div>
</template>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useCartStore } from '@/stores/cart'
import { useUserStore } from '@/stores/user'
import { hapticFeedback } from '@/utils/device'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const cartStore = useCartStore()

const activeTab = ref(0)

// 真正的 Telegram Mini App：必须有 initData（内置浏览器没有 initData）
const isTgContext = ref(
  typeof window !== 'undefined' &&
  !!window.Telegram?.WebApp?.initData &&
  window.Telegram.WebApp.initData.length > 0
)

// Telegram 顶部安全区高度，初始为 0，由 API 精确赋值
const tgTopPadding = ref(0)

onMounted(() => {
  document.body.classList.add('mobile-layout-active')

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
    // 桌面端 TG 标题栏在 WebView 外侧，不需要顶部内边距
    if (window.innerWidth >= 600) {
      tgTopPadding.value = 0
      return
    }
    const content = tg.contentSafeAreaInset?.top ?? 0
    const safe = tg.safeAreaInset?.top ?? 0
    tgTopPadding.value = content + safe
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
  else if (newPath.includes('/m/profile')) activeTab.value = 2
}, { immediate: true })

const handleTabChange = (index) => {
  hapticFeedback('light')
  const routes = ['/m/shop', '/m/cart', '/m/profile']
  router.push(routes[index])
}

onUnmounted(() => {
  document.body.classList.remove('mobile-layout-active')
})
</script>

<style scoped>
/* ── 外层容器 ── */
.desktop-wrapper {
  height: var(--real-vh, 100dvh);
  background: #f0f2f5;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  /* 不用 overflow:hidden，避免裁掉视口级 fixed 元素（tabbar 等） */
}

/* ── 主容器（所有设备） ── */
/* 注意：
   1. 不加 transform，否则 position:fixed 子元素会以此为基准而非视口
   2. 不在全局设 overflow:hidden，否则 position:fixed 子元素（tabbar 等）会被裁剪。
      桌面端需要圆角裁剪时在相应 media query 中单独设置。 */
.mobile-layout {
  width: 100%;
  max-width: 520px;
  height: var(--real-vh, 100dvh);
  background: #f5f5f5;
  display: flex;
  flex-direction: column;
  box-shadow: 0 0 24px rgba(0, 0, 0, 0.12);
}

/* ── 全局顶部 Logo 栏 ── */
.global-header {
  flex-shrink: 0;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px 16px;
  height: 56px;
  z-index: 10;
}

.global-logo {
  height: 40px;
  max-width: 200px;
  object-fit: contain;
}

/* ── 内容滚动区 ── */
.mobile-content {
  flex: 1;
  min-height: 0;
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
    height: calc(var(--real-vh, 100dvh) - 20px);
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
    height: calc(var(--real-vh, 100dvh) - 20px);
    margin-top: 20px;
    border-radius: 32px 32px 0 0;
    overflow: hidden;
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

/* ── 底部导航栏 ── */
:deep(.van-tabbar) {
  background: #fff !important;
  border-top: 1px solid #f0f0f0;
  /* 保证在 mobile-content (z-index:2) 之上绘制，避免被内容区覆盖 */
  z-index: 10 !important;
}
:deep(.van-tabbar-item) {
  color: #aaa !important;
  transition: color 0.2s;
}
:deep(.van-tabbar-item--active) {
  color: #1a1a2e !important;
}
:deep(.van-tabbar-item__icon) {
  font-size: 20px;
}

</style>

<!-- 非 scoped：桌面端将 teleport="body" 的 Vant 弹窗约束在 520px 移动容器内 -->
<style>
/* ── iOS 浏览器工具栏补偿 ── */
/* iOS Chrome/Safari 上 position:fixed 相对 layout viewport 定位，
   浏览器工具栏显示时会被遮挡。--browser-nav-offset 动态补偿工具栏高度 */
@media (max-width: 599px) {
  body.mobile-layout-active .van-tabbar--fixed,
  body.mobile-layout-active .checkout-bar,
  body.mobile-layout-active .back-top-btn,
  body.mobile-layout-active .pd-footer,
  body.mobile-layout-active .van-submit-bar--fixed {
    bottom: calc(env(safe-area-inset-bottom, 0px) + var(--browser-nav-offset, 0px)) !important;
  }
}

/* 桌面端把 position:fixed 的元素约束在 520px 框内 */
@media (min-width: 600px) {
  /* tabbar */
  body.mobile-layout-active .van-tabbar--fixed {
    left: 50% !important;
    right: auto !important;
    max-width: 520px !important;
    width: 520px !important;
    bottom: calc(env(safe-area-inset-bottom, 0px) + var(--browser-nav-offset, 0px)) !important;
    transform: translateX(-50%) !important;
  }
  /* 结算栏（Shop.vue checkout-bar） */
  body.mobile-layout-active .checkout-bar {
    left: 50% !important;
    right: auto !important;
    max-width: 520px !important;
    width: 520px !important;
    bottom: calc(env(safe-area-inset-bottom, 0px) + var(--browser-nav-offset, 0px)) !important;
    transform: translateX(-50%) !important;
  }
  /* 回到顶部按钮 */
  body.mobile-layout-active .back-top-btn {
    right: calc(50% - 244px) !important;
    bottom: calc(env(safe-area-inset-bottom, 0px) + var(--browser-nav-offset, 0px)) !important;
  }
  /* ProductDetail 底部操作栏 */
  body.mobile-layout-active .pd-footer {
    left: 50% !important;
    right: auto !important;
    max-width: 520px !important;
    width: 520px !important;
    bottom: calc(env(safe-area-inset-bottom, 0px) + var(--browser-nav-offset, 0px)) !important;
    transform: translateX(-50%) !important;
  }
  /* Cart 提交栏 */
  body.mobile-layout-active .van-submit-bar--fixed {
    left: 50% !important;
    right: auto !important;
    max-width: 520px !important;
    width: 520px !important;
    bottom: calc(env(safe-area-inset-bottom, 0px) + var(--browser-nav-offset, 0px)) !important;
    transform: translateX(-50%) !important;
  }
}

@media (min-width: 600px) {
  body.mobile-layout-active .van-popup--bottom,
  body.mobile-layout-active .van-popup--top {
    left: 50% !important;
    right: auto !important;
    max-width: 520px !important;
    width: 520px !important;
    transform: translateX(-50%) !important;
  }

  body.mobile-layout-active .van-action-sheet {
    left: 50% !important;
    right: auto !important;
    max-width: 520px !important;
    width: 520px !important;
    transform: translateX(-50%) !important;
  }

  /* 购物车结算栏 */
  body.mobile-layout-active .van-submit-bar {
    left: 50% !important;
    right: auto !important;
    max-width: 520px !important;
    width: 520px !important;
    transform: translateX(-50%) !important;
  }
}
</style>
