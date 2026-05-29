<template>
  <div class="admin-login-container">
    <div class="admin-login-card">
      <div class="admin-logo-area">
        <img src="/images/logo1.svg" alt="Logo" class="admin-logo-img" />
        <div class="admin-login-subtitle">管理员后台</div>
      </div>

      <van-form @submit="handleLogin" ref="formRef">
        <van-cell-group inset>
          <van-field
            v-model="form.username"
            name="username"
            label="账号"
            placeholder="请输入账号"
            left-icon="contact"
            :rules="[{ required: true, message: '请输入账号' }]"
          />
          <van-field
            v-model="form.password"
            type="password"
            name="password"
            label="密码"
            placeholder="请输入密码"
            left-icon="lock"
            :rules="[{ required: true, message: '请输入密码' }]"
            @keyup.enter="handleLogin"
          />
        </van-cell-group>
        <div style="margin: 20px 16px 0;">
          <van-button round block type="primary" native-type="submit" :loading="loading" loading-text="登录中...">
            登 录
          </van-button>
        </div>
      </van-form>

      <div class="admin-back-link" @click="$router.push('/login')">← 返回用户登录</div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { showSuccessToast, showFailToast } from 'vant'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref()
const loading = ref(false)

const form = reactive({ username: '', password: '' })

const handleLogin = async () => {
  loading.value = true
  try {
    await userStore.login(form.username, form.password)
    await new Promise(resolve => setTimeout(resolve, 100))
    if (!userStore.isAdmin) {
      showFailToast('无权限，仅管理员可登录此后台')
      userStore.logout()
      return
    }
    showSuccessToast('登录成功')
    router.push('/admin/dashboard')
  } catch {
    // 错误由 store 处理
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.admin-login-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #1a2a4a 0%, #1d4ed8 60%, #0ea5e9 100%);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 24px;
}

.admin-login-card {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: 12px;
  padding: 40px 0 32px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.25);
}

.admin-logo-area {
  text-align: center;
  margin-bottom: 32px;
  padding: 0 36px;
}

.admin-logo-img {
  width: 200px;
  max-width: 100%;
  height: auto;
  object-fit: contain;
}

.admin-login-subtitle {
  margin-top: 10px;
  font-size: 15px;
  color: #64748b;
  font-weight: 500;
  letter-spacing: 2px;
}

.admin-back-link {
  text-align: center;
  margin-top: 20px;
  font-size: 13px;
  color: #94a3b8;
  cursor: pointer;
  transition: color 0.2s;
}

.admin-back-link:hover {
  color: #1d4ed8;
}
</style>
