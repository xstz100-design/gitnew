<template>
  <div class="admin-login-container">
    <div class="admin-login-card">
      <div class="admin-logo-area">
        <img src="/images/logo1.svg" alt="Logo" class="admin-logo-img" />
        <div class="admin-login-subtitle">管理员后台</div>
      </div>

      <el-form :model="form" :rules="rules" ref="formRef" class="admin-login-form">
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="请输入账号"
            size="large"
            clearable
            :prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            size="large"
            show-password
            :prefix-icon="Lock"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            size="large"
            style="width: 100%"
            @click="handleLogin"
          >
            登 录
          </el-button>
        </el-form-item>
      </el-form>

      <div class="admin-back-link" @click="$router.push('/login')">← 返回用户登录</div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus/es/components/message/index'
import { User, Lock } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { changePassword } from '@/api'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref()
const loading = ref(false)

const form = reactive({ username: '', password: '' })

const rules = {
  username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

// 强制改密
const changePasswordVisible = ref(false)
const changingPassword = ref(false)
const pwdFormRef = ref()
const pwdForm = reactive({ new_password: '', confirm_password: '' })
let tempOldPassword = ''

const handleLogin = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
  } catch {
    return
  }
  loading.value = true
  try {
    const data = await userStore.login(form.username, form.password)
    await new Promise(resolve => setTimeout(resolve, 100))
    if (!userStore.isAdmin) {
      ElMessage.error('无权限，仅管理员可登录此后台')
      userStore.logout()
      return
    }
    if (data.user?.must_change_password) {
      tempOldPassword = form.password
      changePasswordVisible.value = true
      return
    }
    ElMessage.success('登录成功')
    router.push('/admin/dashboard')
  } catch {
    // 错误由 store 处理
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
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
  padding: 40px 36px 32px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.25);
}

.admin-logo-area {
  text-align: center;
  margin-bottom: 32px;
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

.admin-login-form {
  :deep(.el-form-item) {
    margin-bottom: 20px;
    &:last-child { margin-bottom: 0; margin-top: 12px; }
  }
  :deep(.el-input__wrapper) {
    height: 46px;
    border-radius: 8px;
  }
  :deep(.el-button--large) {
    height: 46px;
    font-size: 15px;
    font-weight: 600;
    border-radius: 8px;
    letter-spacing: 4px;
  }
}

.admin-back-link {
  text-align: center;
  margin-top: 20px;
  font-size: 13px;
  color: #94a3b8;
  cursor: pointer;
  transition: color 0.2s;
  &:hover { color: #1d4ed8; }
}
</style>
