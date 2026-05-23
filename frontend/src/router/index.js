import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { isMobile } from '@/utils/device'
import { isTelegramMiniApp, getInitData } from '@/utils/telegram'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // ── 用户端登录（Telegram 登录，无管理员入口）
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false },
    },
    // ── 管理员后台独立登录
    {
      path: '/admin/login',
      name: 'AdminLogin',
      component: () => import('@/views/admin/AdminLogin.vue'),
      meta: { requiresAuth: false, adminLoginPage: true },
    },
    // ── 移动端商户路由 (Vant UI)
    {
      path: '/m',
      component: () => import('@/layouts/MobileLayout.vue'),
      meta: { requiresAuth: true, roles: ['merchant', 'picker', 'delivery'], mobile: true },
      children: [
        { path: '', redirect: '/m/shop' },
        { path: 'shop', name: 'MobileShop', component: () => import('@/views/mobile/Shop.vue') },
        { path: 'cart', name: 'MobileCart', component: () => import('@/views/mobile/Cart.vue') },
        { path: 'orders', name: 'MobileOrders', component: () => import('@/views/mobile/Orders.vue') },
        { path: 'profile', name: 'MobileProfile', component: () => import('@/views/mobile/Profile.vue') },
      ],
    },
    // ── 管理端路由 (Element Plus) - 仅 admin 角色
    {
      path: '/admin',
      component: () => import('@/layouts/AdminLayout.vue'),
      meta: { requiresAuth: true, roles: ['admin'] },
      children: [
        { path: '', redirect: '/admin/dashboard' },
        { path: 'dashboard', name: 'AdminDashboard', component: () => import('@/views/admin/Dashboard.vue') },
        { path: 'products', name: 'AdminProducts', component: () => import('@/views/admin/Products.vue') },
        { path: 'orders', name: 'AdminOrders', component: () => import('@/views/admin/Orders.vue') },
        { path: 'settings', name: 'AdminSettings', component: () => import('@/views/admin/Settings.vue') },
        { path: 'merchants', name: 'AdminMerchants', component: () => import('@/views/admin/Merchants.vue') },
        { path: 'approvals', redirect: '/admin/merchants' },
        { path: 'categories', name: 'AdminCategories', component: () => import('@/views/admin/Categories.vue') },
        { path: 'announcements', name: 'AdminAnnouncements', component: () => import('@/views/admin/Announcements.vue') },
        { path: 'profile', name: 'AdminProfile', component: () => import('@/views/admin/Profile.vue') },
        { path: 'picking', name: 'AdminPicking', component: () => import('@/views/admin/Picking.vue') },
      ],
    },
    // 根路径
    {
      path: '/',
      redirect: '/m/shop',
    },
  ],
})

// ─── 路由守卫
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  const inTelegram = isTelegramMiniApp()

  const isAdminRoute = to.path.startsWith('/admin') && !to.meta.adminLoginPage
  const isUserRoute = to.path.startsWith('/m') || to.path === '/'

  // ── 未登录
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    if (isAdminRoute) {
      next('/admin/login')
      return
    }
    // 在 Telegram 内，尝试自动登录
    if (inTelegram) {
      try {
        await userStore.telegramLogin(getInitData())
        // 登录成功后继续导航
      } catch (e) {
        console.error('TG 自动登录失败:', e)
        next('/login')
        return
      }
    } else {
      next('/login')
      return
    }
  }

  // ── 已登录访问 /login → 去用户端
  if (to.path === '/login' && userStore.isLoggedIn) {
    next('/m/shop')
    return
  }

  // ── 已登录访问 /admin/login → 去管理后台
  if (to.meta.adminLoginPage && userStore.isLoggedIn) {
    if (userStore.isAdmin) {
      next('/admin/dashboard')
    } else {
      next('/m/shop')
    }
    return
  }

  // ── Telegram 内访问 /login → 自动登录后去用户端
  if (to.path === '/login' && inTelegram && !userStore.isLoggedIn) {
    try {
      await userStore.telegramLogin(getInitData())
      next('/m/shop')
      return
    } catch {
      // 登录失败，继续显示登录页
    }
  }

  // ── 角色不匹配
  const allowedRoles = to.meta.roles || (to.meta.role ? [to.meta.role] : null)
  if (allowedRoles && !allowedRoles.includes(userStore.userRole)) {
    // 管理员在 Telegram miniapp 中可访问 /m/* 移动端路由
    if (userStore.isAdmin && inTelegram && (to.path.startsWith('/m') || to.path === '/')) {
      next()
      return
    }
    if (isAdminRoute) {
      next('/admin/login')
    } else if (userStore.isAdmin) {
      next('/admin/dashboard')
    } else {
      next('/login')
    }
    return
  }

  next()
})

export default router
