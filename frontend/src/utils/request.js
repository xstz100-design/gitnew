import axios from 'axios'
import { ElMessage } from 'element-plus/es/components/message/index'
import { useUserStore } from '@/stores/user'
import i18n from '@/i18n'

const request = axios.create({
  baseURL: '',
  timeout: 15000,
})

// 修复: 移除全局单例缓存，避免登出后状态不一致
// 请求拦截器

request.interceptors.request.use(
  (config) => {
    const store = useUserStore()
    if (store.token) {
      config.headers.Authorization = `Bearer ${store.token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 响应拦截器
request.interceptors.response.use(
  (response) => response.data,
  (error) => {
    const message = error.code === 'ECONNABORTED'
      ? i18n.global.t('common.requestTimeout')
      : error.response?.data?.detail || error.message || i18n.global.t('common.requestFailed')

    if (error.response?.status === 401 && !error.config?.url?.includes('/api/auth/login')) {
      useUserStore().logout()
      if (window.location.pathname !== '/login') {
        window.location.assign('/login')
      }
    }

    ElMessage.error(message)
    return Promise.reject(error)
  }
)

export default request
