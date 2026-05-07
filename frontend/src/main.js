import { createApp } from 'vue'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import en from 'element-plus/es/locale/lang/en'

// Vant UI - 移动端
import Vant from 'vant'
import 'vant/lib/index.css'
import '@vant/touch-emulator'

// i18n
import i18n, { getCurrentLanguage } from './i18n'

// 全局样式
import '@/styles/global.scss'

// Telegram Mini App
import { isTelegramMiniApp, getInitData, initBrowserViewportVars, initTelegramWebApp } from '@/utils/telegram'

import App from './App.vue'
import router from './router'
import { useUserStore } from '@/stores/user'

const app = createApp(App)

const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)
app.use(pinia)
app.use(router)
app.use(i18n)
app.use(ElementPlus, { locale: getCurrentLanguage() === 'en' ? en : zhCn })
app.use(Vant)

initBrowserViewportVars()

// Telegram Mini App 自动登录
async function bootstrapTelegram() {
  if (!isTelegramMiniApp()) return false

  initTelegramWebApp()

  const userStore = useUserStore()
  // 已有 token 则验证 token 有效性
  if (userStore.isLoggedIn) {
    try {
      await userStore.fetchUserInfo()
      return true
    } catch {
      // token 失效，重新登录
      userStore.logout()
    }
  }

  try {
    await userStore.telegramLogin(getInitData())
    return true
  } catch (e) {
    console.error('Telegram 自动登录失败:', e)
    return false
  }
}

async function syncTelegramEntryRoute() {
  if (!isTelegramMiniApp()) return

  const userStore = useUserStore()
  const currentPath = router.currentRoute.value.path

  if (userStore.isAdmin) {
    if (currentPath !== '/admin/dashboard') {
      await router.replace('/admin/dashboard')
    }
    return
  }

  if (userStore.isMerchant && (currentPath === '/' || currentPath === '/login' || currentPath.startsWith('/admin'))) {
    await router.replace('/m/shop')
  }
}

bootstrapTelegram().finally(async () => {
  await syncTelegramEntryRoute()
  app.mount('#app')
})
