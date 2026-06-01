import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { isMobile } from '@/utils/device'
import { isTelegramMiniApp, getInitData } from '@/utils/telegram'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    // /login 已废弃，直接重定向到商城
    { path: '/login', redirect: '/m/shop' },
    // ── 管理员后台独立登录
    {
      path: '/admin/login',
      name: 'AdminLogin',
      component: () => import('@/views/admin/AdminLogin.vue'),
      meta: { requiresAuth: false, adminLoginPage: true },
    },
    // ── 移动端商户路由 (Vant UI) — 公开可浏览
    {
      path: '/m',
      component: () => import('@/layouts/MobileLayout.vue'),
      meta: { mobile: true },
      children: [
        { path: '', redirect: '/m/shop' },
        { path: 'shop', name: 'MobileShop', component: () => import('@/views/mobile/Shop.vue') },
        { path: 'product/:id', name: 'ProductDetail', component: () => import('@/views/mobile/ProductDetail.vue') },
        { path: 'cart', name: 'MobileCart', component: () => import('@/views/mobile/Cart.vue') },
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

// ─── 路由守卫 — 仅保护管理端路由
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  const inTelegram = isTelegramMiniApp()

  const isAdminRoute = to.path.startsWith('/admin') && !to.meta.adminLoginPage

  // ── 未登录访问管理端受保护路由 → 去管理员登录页
  if (isAdminRoute && to.meta.requiresAuth && !userStore.isLoggedIn) {
    next('/admin/login')
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

  next()
})

export default router
