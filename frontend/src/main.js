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

// --real-vh: 始终用 visualViewport.height（浏览器所有 UI 都已排除的可见高度）
// 布局高度 = 可见高度，tabbar 作为 flex 末尾自然贴底，无需任何 bottom 偏移
function syncViewportVars() {
  const h = Math.round(window.visualViewport ? window.visualViewport.height : window.innerHeight)
  document.documentElement.style.setProperty('--real-vh', h + 'px')
}
syncViewportVars()
if (window.visualViewport) {
  window.visualViewport.addEventListener('resize', syncViewportVars, { passive: true })
}
window.addEventListener('resize', syncViewportVars, { passive: true })

// --browser-nav-offset: iOS 浏览器工具栏高度
// position:fixed 在 iOS 上相对 layout viewport 定位，工具栏显示时元素被遮挡
// 用 visualViewport 计算实际补偿值
function syncBrowserNavOffset() {
  const offset = window.visualViewport
    ? Math.max(0, window.innerHeight - window.visualViewport.height)
    : 0
  document.documentElement.style.setProperty('--browser-nav-offset', offset + 'px')
}
syncBrowserNavOffset()
if (window.visualViewport) {
  window.visualViewport.addEventListener('resize', syncBrowserNavOffset, { passive: true })
}

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

  // 管理员和商户在 Telegram miniapp 中均使用移动端界面
  if (userStore.isLoggedIn) {
    if (currentPath === '/' || currentPath.startsWith('/admin')) {
      await router.replace('/m/shop')
    }
  }
}

async function bootstrapGuest() {
  // 非 Telegram 环境且未登录时，自动以游客身份登录
  if (isTelegramMiniApp()) return
  const userStore = useUserStore()
  if (userStore.isLoggedIn) return
  try {
    await userStore.guestLogin()
  } catch (e) {
    console.warn('游客登录失败:', e)
  }
}

bootstrapTelegram()
  .then(bootstrapGuest)
  .finally(async () => {
    await syncTelegramEntryRoute()
    app.mount('#app')
  })
