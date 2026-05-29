<template>
  <div class="admin-layout" :class="{ 'is-mobile': mobile }">
    <!-- 桌面端 header -->
    <div v-if="!mobile" class="admin-header">
      <div class="header-content">
        <div class="header-left">
          <img src="/images/logo1.svg" alt="Logo" class="header-logo" />
        </div>
        <div class="header-right">
          <button class="lang-btn" @click="toggleLang">{{ langLabel }}</button>
          <div class="user-info" @click="handleLogout">
            {{ userStore.userInfo?.full_name }} ({{ $t('admin.admin') }})
            <van-icon name="arrow-down" size="12" />
          </div>
        </div>
      </div>
    </div>

    <!-- 移动端 header -->
    <div v-if="mobile" class="mobile-header">
      <img src="/images/logo1.svg" alt="Logo" class="header-logo" />
      <button class="lang-btn" @click="toggleLang">{{ langLabel }}</button>
    </div>

    <div class="admin-body">
      <!-- 桌面端侧边栏 -->
      <div v-if="!mobile" class="admin-sidebar">
        <div class="sidebar-title">管理后台</div>
        <nav class="admin-nav">
          <router-link to="/admin/dashboard" class="nav-item" active-class="active">
            <van-icon name="bar-chart-o" /><span>{{ $t('admin.dashboard') }}</span>
          </router-link>
          <router-link to="/admin/products" class="nav-item" active-class="active">
            <van-icon name="bag-o" /><span>{{ $t('admin.products') }}</span>
          </router-link>
          <router-link to="/admin/orders" class="nav-item" active-class="active">
            <van-icon name="orders-o" /><span>{{ $t('admin.orders') }}</span>
          </router-link>
          <router-link to="/admin/merchants" class="nav-item" active-class="active">
            <van-icon name="manager-o" /><span>{{ $t('admin.merchants') }}</span>
          </router-link>
          <router-link to="/admin/categories" class="nav-item" active-class="active">
            <van-icon name="apps-o" /><span>{{ $t('admin.categories') }}</span>
          </router-link>
          <router-link to="/admin/announcements" class="nav-item" active-class="active">
            <van-icon name="bell" /><span>{{ $t('admin.announcements') }}</span>
          </router-link>
          <router-link to="/admin/settings" class="nav-item" active-class="active">
            <van-icon name="setting-o" /><span>{{ $t('settings.title') }}</span>
          </router-link>
          <div class="nav-item nav-logout" @click="handleLogout">
            <van-icon name="exchange" /><span>{{ $t('profile.logout') }}</span>
          </div>
        </nav>
      </div>

      <main class="admin-main" :style="mobile ? mobileMainStyle : {}">
        <router-view />
      </main>
    </div>

    <!-- 移动端底部 TabBar -->
    <van-tabbar v-if="mobile" route active-color="#1d4ed8">
      <van-tabbar-item to="/admin/dashboard" icon="bar-chart-o">{{ $t('admin.dashboard') }}</van-tabbar-item>
      <van-tabbar-item to="/admin/products" icon="bag-o">{{ $t('admin.products') }}</van-tabbar-item>
      <van-tabbar-item to="/admin/orders" icon="orders-o">{{ $t('admin.orders') }}</van-tabbar-item>
      <van-tabbar-item to="/admin/merchants" icon="manager-o">{{ $t('admin.merchants') }}</van-tabbar-item>
      <van-tabbar-item icon="more-o" @click.prevent="showMore = true">更多</van-tabbar-item>
    </van-tabbar>

    <!-- 更多菜单 -->
    <van-action-sheet v-if="mobile" v-model:show="showMore" :cancel-text="$t('common.cancel')">
      <van-cell-group inset style="margin-bottom: 8px;">
        <van-cell :title="$t('admin.categories')" icon="apps-o" is-link @click="goTo('/admin/categories')" />
        <van-cell :title="$t('admin.announcements')" icon="bell" is-link @click="goTo('/admin/announcements')" />
        <van-cell :title="$t('settings.title')" icon="setting-o" is-link @click="goTo('/admin/settings')" />
        <van-cell :title="langLabel" icon="gem-o" @click="toggleLang" />
      </van-cell-group>
      <div style="padding: 0 16px 16px;">
        <van-button block type="danger" plain @click="handleLogout">{{ $t('profile.logout') }}</van-button>
      </div>
    </van-action-sheet>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { showConfirmDialog } from 'vant'
import { useUserStore } from '@/stores/user'
import { useI18n } from 'vue-i18n'
import { setLanguage, getCurrentLanguage } from '@/i18n'

const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()
const currentLang = ref(getCurrentLanguage())
const mobile = ref(window.innerWidth < 768)
const showMore = ref(false)

const mobileMainStyle = computed(() => ({
  paddingTop: `calc(46px + var(--tg-content-safe-area-inset-top, 0px))`,
  paddingBottom: `calc(60px + var(--tg-content-safe-area-inset-bottom, env(safe-area-inset-bottom, 0px)))`,
  paddingLeft: 'var(--tg-content-safe-area-inset-left, 0px)',
  paddingRight: 'var(--tg-content-safe-area-inset-right, 0px)',
}))

const onResize = () => { mobile.value = window.innerWidth < 768 }
onMounted(() => window.addEventListener('resize', onResize))
onUnmounted(() => window.removeEventListener('resize', onResize))

const toggleLang = () => {
  const order = ['zh', 'en', 'kh']
  const idx = order.indexOf(currentLang.value)
  const newLang = order[(idx + 1) % order.length]
  setLanguage(newLang)
  currentLang.value = newLang
  showMore.value = false
}

const langLabel = computed(() => {
  const next = { zh: 'EN', en: 'ខ្មែរ', kh: '中文' }
  return next[currentLang.value] || 'EN'
})

const goTo = (path) => {
  showMore.value = false
  router.push(path)
}

const handleLogout = () => {
  showConfirmDialog({
    title: t('admin.hint'),
    message: t('admin.logoutConfirm'),
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
  }).then(() => {
    userStore.logout()
    router.push('/login')
  }).catch(() => {})
}
</script>

<style scoped>
.admin-layout {
  min-height: var(--tg-viewport-stable-height, 100vh);
  display: flex;
  flex-direction: column;
}

/* ===== 桌面端 header ===== */
.admin-header {
  background: #fff;
  border-bottom: 1px solid #eef0f3;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
  position: relative;
  flex-shrink: 0;
  z-index: 10;
}

.admin-header::after {
  content: '';
  position: absolute;
  left: 0; right: 0; bottom: 0;
  height: 2px;
  background: linear-gradient(90deg, #1e40af 0%, #2563eb 50%, #0891b2 100%);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  height: 56px;
}

.header-left, .header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-logo {
  height: 32px;
  width: auto;
  object-fit: contain;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  color: #1a1a1a;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.2s;
}

.user-info:hover {
  background: #f5f7fa;
}

/* ===== 桌面端布局 ===== */
.admin-body {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.admin-sidebar {
  width: 200px;
  flex-shrink: 0;
  background: #fff;
  border-right: 1px solid #eef0f3;
  box-shadow: 2px 0 8px rgba(0,0,0,0.04);
  display: flex;
  flex-direction: column;
  z-index: 5;
}

.sidebar-title {
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  color: #fff;
  letter-spacing: 1.5px;
  flex-shrink: 0;
  background: linear-gradient(135deg, #1e40af 0%, #2563eb 100%);
}

.admin-nav {
  padding: 6px 0 12px;
  flex: 1;
  overflow-y: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 11px 18px;
  font-size: 13px;
  color: #4a5568;
  text-decoration: none;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  border-left: 3px solid transparent;
  margin: 1px 6px;
  border-radius: 6px;
}

.nav-item:hover {
  background: #eff6ff;
  color: #1d4ed8;
  border-left-color: transparent;
}

.nav-item.active {
  background: #eff6ff;
  color: #1d4ed8;
  border-left-color: #1d4ed8;
  font-weight: 600;
  border-radius: 0 6px 6px 0;
  margin-left: 0;
  padding-left: 15px;
}

.nav-logout {
  color: #6b7280;
  margin-top: 6px;
  border-top: 1px solid #f0f0f0;
  padding-top: 4px;
}

.nav-logout:hover {
  background: #fef2f2;
  color: #dc2626;
}

.admin-main {
  flex: 1;
  overflow: auto;
  background: #f5f7fa;
}

/* ===== 移动端 header ===== */
.mobile-header {
  height: 46px;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  position: fixed;
  top: var(--tg-content-safe-area-inset-top, 0px);
  left: 0;
  right: 0;
  z-index: 100;
  border-bottom: 1px solid #eef0f3;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
}

.mobile-header::after {
  content: '';
  position: absolute;
  left: 0; right: 0; bottom: 0;
  height: 2px;
  background: linear-gradient(90deg, #1e40af 0%, #2563eb 50%, #0891b2 100%);
}

.mobile-header .header-logo { height: 28px; }

/* ===== 通用 ===== */
.lang-btn {
  padding: 4px 12px;
  border: 1px solid #d9dde3;
  border-radius: 4px;
  background: #fff;
  color: #1a1a1a;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.lang-btn:hover {
  background: #f0f6ff;
  border-color: #1d4ed8;
  color: #1d4ed8;
}

/* 移动端适配 */
.is-mobile .admin-body {
  flex-direction: column;
}
</style>
