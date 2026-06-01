import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { login as apiLogin, getCurrentUser, telegramAuth as apiTelegramAuth, guestLogin as apiGuestLogin } from '@/api'

let addressIdCounter = Date.now()

/**
 * 创建一个收货地址对象
 */
function createAddress(label, fullAddress, phone, opts = {}) {
  return {
    id: ++addressIdCounter,
    label: label || '',
    full_address: fullAddress || '',
    full_name: opts.full_name || opts.name || '',
    phone: phone || '',
    location_url: opts.location_url || '',
    isDefault: !!opts.isDefault,
    distance_km: opts.distance_km || 0,
    note: opts.note || '',
  }
}

export const useUserStore = defineStore(
  'user',
  () => {
    const token = ref('')
    const userInfo = ref(null)
    /** @type {import('vue').Ref<Array>} 多地址列表 */
    const addresses = ref([])

    // ---------- 计算属性 ----------

    // 是否已登录
    const isLoggedIn = computed(() => !!token.value && !!userInfo.value)

    // 用户角色
    const userRole = computed(() => userInfo.value?.role || '')

    // 是否是管理员
    const isAdmin = computed(() => userRole.value === 'admin')

    const isSuperAdmin = computed(() => !!userInfo.value?.is_super_admin)

    // 是否是商户
    const isMerchant = computed(() => ['merchant', 'picker', 'delivery'].includes(userRole.value))

    // 资料是否完整
    const profileCompleted = computed(() => !!userInfo.value?.profile_completed)

    // 审核状态
    const approvalStatus = computed(() => userInfo.value?.approval_status || '')

    // 是否已审核通过
    const isApproved = computed(() => approvalStatus.value === 'approved')

    // 下单受限状态
    const orderAccessState = computed(() => {
      if (!isMerchant.value) return 'approved'
      if (!profileCompleted.value) return 'incomplete'
      if (approvalStatus.value === 'rejected') return 'rejected'
      if (approvalStatus.value === 'pending') return 'pending'
      if (approvalStatus.value === 'approved') return 'approved'
      return 'pending'
    })

    // 能否下单
    const canOrder = computed(() => profileCompleted.value && isApproved.value)

    // 默认地址
    const defaultAddress = computed(() => addresses.value.find(a => a.isDefault) || null)

    // ---------- 登录 ----------

    async function login(username, password) {
      const data = await apiLogin(username, password)
      token.value = data.access_token
      userInfo.value = data.user
      return data
    }

    async function telegramLogin(initData) {
      const data = await apiTelegramAuth(initData)
      token.value = data.access_token
      userInfo.value = data.user
      return data
    }

    async function guestLogin() {
      // 获取或生成本地设备 ID（持久化到 localStorage）
      let deviceId = localStorage.getItem('_guest_device_id')
      if (!deviceId) {
        deviceId = crypto.randomUUID ? crypto.randomUUID() : Math.random().toString(36).slice(2) + Date.now().toString(36)
        localStorage.setItem('_guest_device_id', deviceId)
      }
      const data = await apiGuestLogin(deviceId)
      token.value = data.access_token
      userInfo.value = data.user
      return data
    }

    function logout() {
      token.value = ''
      userInfo.value = null
      addresses.value = []
    }

    async function fetchUserInfo() {
      if (!token.value) return
      try {
        const data = await getCurrentUser()
        userInfo.value = data
      } catch (error) {
        console.error('获取用户信息失败:', error)
        logout()
      }
    }

    // ---------- 地址管理 ----------

    /**
     * 添加地址
     */
    function addAddress(label, fullAddress, phone, opts = {}) {
      const addr = createAddress(label, fullAddress, phone, opts)
      // 如果是第一个地址，自动设为默认
      if (addresses.value.length === 0) {
        addr.isDefault = true
      } else if (addr.isDefault) {
        // 将其他地址的非默认
        addresses.value.forEach(a => (a.isDefault = false))
      }
      addresses.value.push(addr)
      return addr
    }

    /**
     * 更新地址
     */
    function updateAddress(id, data) {
      const idx = addresses.value.findIndex(a => a.id === id)
      if (idx === -1) return
      if (data.isDefault) {
        addresses.value.forEach(a => (a.isDefault = a.id === id))
      }
      Object.assign(addresses.value[idx], data)
    }

    /**
     * 删除地址
     */
    function removeAddress(id) {
      const idx = addresses.value.findIndex(a => a.id === id)
      if (idx === -1) return
      const wasDefault = addresses.value[idx].isDefault
      addresses.value.splice(idx, 1)
      // 删的是默认地址，把第一个设为默认
      if (wasDefault && addresses.value.length > 0) {
        addresses.value[0].isDefault = true
      }
    }

    /**
     * 设为默认地址
     */
    function setDefaultAddress(id) {
      addresses.value.forEach(a => (a.isDefault = a.id === id))
    }

    /**
     * 从 userInfo 上迁移旧地址到多地址列表
     *（如果用户之前只在 userInfo.address 存了单个地址）
     */
    function migrateLegacyAddress() {
      if (addresses.value.length > 0) return
      const info = userInfo.value
      if (!info) return
      const legacyAddr = info.address || ''
      const legacyPhone = info.phone || ''
      if (legacyAddr.trim()) {
        addAddress(
          legacyAddr.length > 10 ? legacyAddr.slice(0, 10) + '…' : legacyAddr,
          legacyAddr,
          legacyPhone,
          { isDefault: true },
        )
      }
    }

    return {
      token,
      userInfo,
      addresses,
      // computed
      isLoggedIn,
      userRole,
      isAdmin,
      isSuperAdmin,
      isMerchant,
      profileCompleted,
      approvalStatus,
      isApproved,
      orderAccessState,
      canOrder,
      defaultAddress,
      // auth
      login,
      telegramLogin,
      guestLogin,
      logout,
      fetchUserInfo,
      // addresses
      addAddress,
      updateAddress,
      removeAddress,
      setDefaultAddress,
      migrateLegacyAddress,
    }
  },
  {
    persist: {
      key: 'user',
      storage: localStorage,
      pick: ['token', 'userInfo', 'addresses'],
    },
  }
)
