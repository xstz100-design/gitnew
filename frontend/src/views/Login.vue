<template>
  <div class="login-container" :class="{ 'tg-fullscreen': isTgContext }">

    <!-- ══ Telegram 内：全屏自动登录 ══ -->
    <div v-if="isTgContext" class="tg-autoscreen">
      <div v-if="tgAutoLoading" class="tg-auto-body">
        <div class="tg-spinner"></div>
        <p>{{ $t('login.loginSuccess') }}…</p>
      </div>
      <div v-else-if="tgAutoError" class="tg-auto-body tg-auto-err">
        <p>{{ tgAutoError }}</p>
        <button class="tg-retry-btn" @click="handleTgAutoLogin">{{ $t('common.refresh') }}</button>
      </div>
    </div>

    <!-- ══ 外部浏览器：登录卡片 ══ -->
    <template v-else>
      <button class="login-lang-btn" @click="toggleLang">{{ langLabel }}</button>
      <div class="login-card">
        <div class="login-header">
          <div class="logo-container">
            <img src="/images/logo1.svg" alt="Logo" class="logo-image" />
          </div>
        </div>

        <!-- Telegram Bot 深链登录 -->
        <div class="tg-login-section">
          <div v-if="!tgPolling">
            <p class="tg-browser-tip">{{ $t('login.openInTelegram') }}</p>
            <div style="display:flex;justify-content:center;margin-bottom:8px">
              <button class="tg-btn" :disabled="tgLoading" @click="handleTgLogin">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="#fff"><path d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0zm5.894 8.221l-1.97 9.28c-.145.658-.537.818-1.084.508l-3-2.21-1.447 1.394c-.16.16-.295.295-.605.295l.213-3.053 5.56-5.023c.242-.213-.054-.333-.373-.12L7.17 13.857l-2.96-.924c-.643-.204-.657-.643.136-.953l11.57-4.461c.537-.194 1.006.131.978.702z"/></svg>
                <span>{{ tgLoading ? $t('login.generatingLink') : $t('login.loginWithTelegram') }}</span>
              </button>
            </div>
          </div>
          <div v-else class="tg-waiting">
            <div class="tg-spinner"></div>
            <p>{{ $t('login.tgWaiting') }}</p>
            <p class="tg-countdown">{{ $t('login.tgExpiry', { sec: tgCountdown }) }}</p>
            <button class="tg-cancel-btn" @click="cancelTgLogin">{{ $t('common.cancel') }}</button>
          </div>
          <div v-if="tgError" class="tg-error">{{ tgError }}</div>
        </div>

        <!-- 账号密码登录 -->
        <div class="pwd-login-section">
          <div class="pwd-login-divider"><span>{{ $t('login.orPhoneLogin') }}</span></div>
          <el-input v-model="pwdForm.username" :placeholder="$t('login.phonePlaceholder')"
            size="large" style="margin-bottom:10px" :disabled="pwdLoading" />
          <el-input v-model="pwdForm.password" type="password" :placeholder="$t('login.passwordPlaceholder')"
            size="large" show-password :disabled="pwdLoading" @keyup.enter="handlePwdLogin" />
          <div v-if="pwdError" class="form-error" style="margin-top:6px">{{ pwdError }}</div>
          <el-button type="primary" size="large" style="width:100%;margin-top:12px"
            :loading="pwdLoading" @click="handlePwdLogin">
            {{ $t('login.login') }}
          </el-button>
        </div>

        <div class="login-footer">
          <span class="forgot-link" @click="showForgotPassword">{{ $t('login.forgotPassword') }}</span>
        </div>
      </div>
    </template>

    <!-- 完善账号弹窗（Telegram 注册后首次设置手机号+密码） -->
    <el-dialog v-model="setupVisible" :title="$t('setup.title')" width="400px"
      :close-on-click-modal="false" :close-on-press-escape="false" :show-close="false" center>
      <p class="setup-tip">{{ $t('setup.tip') }}</p>
      <el-input v-model="setupForm.phone" :placeholder="$t('setup.phonePlaceholder')"
        size="large" style="margin-bottom:12px" :disabled="setupLoading" />
      <el-input v-model="setupForm.password" type="password" :placeholder="$t('setup.passwordPlaceholder')"
        size="large" show-password style="margin-bottom:12px" :disabled="setupLoading" />
      <el-input v-model="setupForm.confirmPassword" type="password" :placeholder="$t('setup.confirmPasswordPlaceholder')"
        size="large" show-password :disabled="setupLoading" @keyup.enter="handleSetup" />
      <div v-if="setupError" class="form-error" style="margin-top:8px">{{ setupError }}</div>
      <template #footer>
        <el-button type="primary" size="large" style="width:100%" :loading="setupLoading" @click="handleSetup">
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 忘记密码弹窗 -->
    <el-dialog v-model="forgotVisible" :title="$t('login.forgotPassword')" width="380px" center>
      <div class="forgot-content">
        <p class="forgot-tip">{{ $t('login.forgotTip') }}</p>
        <div class="forgot-contact">
          <p v-if="contactInfo" v-for="(line, idx) in contactLines" :key="idx">{{ line }}</p>
          <p v-else>{{ $t('login.contactAdmin') }}</p>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="forgotVisible = false">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 首次登录强制改密弹窗（管理员临时密码） -->
    <el-dialog v-model="changePwdVisible" :title="$t('login.mustChangePassword')" width="400px"
      :close-on-click-modal="false" :close-on-press-escape="false" :show-close="false" center>
      <p class="change-pwd-tip">{{ $t('login.mustChangePasswordTip') }}</p>
      <el-form :model="changePwdForm" :rules="changePwdRules" ref="changePwdRef">
        <el-form-item prop="new_password">
          <el-input v-model="changePwdForm.new_password" type="password"
            :placeholder="$t('profile.newPasswordPlaceholder')" size="large" show-password />
        </el-form-item>
        <el-form-item prop="confirm_password">
          <el-input v-model="changePwdForm.confirm_password" type="password"
            :placeholder="$t('profile.confirmPasswordPlaceholder')" size="large" show-password
            @keyup.enter="handleChangePwd" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button type="primary" :loading="changePwdLoading" style="width:100%" @click="handleChangePwd">
          {{ $t('common.confirm') }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus/es/components/message/index'
import { useUserStore } from '@/stores/user'
import { useI18n } from 'vue-i18n'
import { setLanguage, getCurrentLanguage } from '@/i18n'
import { changePassword, getPublicAnnouncements, telegramAuth, botLoginCreate, botLoginVerify, login, setupCredentials } from '@/api'

const router = useRouter()
const userStore = useUserStore()
const { t } = useI18n()
const currentLang = ref(getCurrentLanguage())

const toggleLang = () => {
  const order = ['zh', 'en', 'kh']
  const newLang = order[(order.indexOf(currentLang.value) + 1) % order.length]
  setLanguage(newLang)
  currentLang.value = newLang
}
const langLabel = computed(() => ({ zh: 'English', en: 'ខ្មែរ', kh: '中文' })[currentLang.value] || 'English')

const isTgContext = typeof window !== 'undefined' && (
  !!(window.Telegram?.WebApp?.initData) ||
  !!(window.TelegramWebviewProxy) ||
  /Telegram/i.test(navigator.userAgent)
)

// ── 登录后统一检查：是否需要完善资料 ──
const afterLogin = (user) => {
  if (user.role === 'merchant' && !user.phone) {
    setupVisible.value = true
  } else if (user.must_change_password) {
    changePwdVisible.value = true
  } else {
    router.push('/m/shop')
  }
}

// ── Telegram 自动登录（Mini App） ──
const tgAutoLoading = ref(false)
const tgAutoError = ref('')

const handleTgAutoLogin = async () => {
  const initData = window.Telegram?.WebApp?.initData
  if (!initData) { tgAutoError.value = t('login.tgNoInitData'); return }
  tgAutoLoading.value = true
  tgAutoError.value = ''
  try {
    const data = await telegramAuth(initData)
    userStore.token = data.access_token
    userStore.userInfo = data.user
    afterLogin(data.user)
  } catch (e) {
    tgAutoError.value = e?.response?.status === 403
      ? t('login.accountDisabled')
      : (e?.response?.data?.detail || t('login.tgAutoFailed'))
  } finally {
    tgAutoLoading.value = false
  }
}

// ── Bot 深链登录（浏览器） ──
const tgLoading = ref(false)
const tgPolling = ref(false)
const tgCountdown = ref(300)
const tgError = ref('')
let tgPollTimer = null
let tgCountdownTimer = null
let tgCurrentToken = ''

const stopPolling = () => {
  clearInterval(tgPollTimer); clearInterval(tgCountdownTimer)
  tgPollTimer = null; tgCountdownTimer = null
}
const cancelTgLogin = () => {
  stopPolling(); tgPolling.value = false; tgCurrentToken = ''; tgCountdown.value = 300
}
const handleTgLogin = async () => {
  tgError.value = ''
  tgLoading.value = true
  try {
    const data = await botLoginCreate()
    tgCurrentToken = data.token
    window.open(data.bot_url, '_blank')
    tgPolling.value = true
    tgCountdown.value = 300
    tgPollTimer = setInterval(async () => {
      try {
        const result = await botLoginVerify(tgCurrentToken)
        if (result?.access_token) {
          stopPolling(); tgPolling.value = false
          userStore.token = result.access_token
          userStore.userInfo = result.user
          ElMessage.success(t('login.loginSuccess'))
          afterLogin(result.user)
        }
      } catch (e) {
        const s = e?.response?.status
        if (s === 410 || s === 404) { stopPolling(); tgPolling.value = false; tgError.value = t('login.tgLinkExpired') }
        else if (s === 403) { stopPolling(); tgPolling.value = false; tgError.value = t('login.accountDisabled') }
      }
    }, 2000)
    tgCountdownTimer = setInterval(() => {
      if (--tgCountdown.value <= 0) { stopPolling(); tgPolling.value = false; tgError.value = t('login.tgTimeout') }
    }, 1000)
  } catch {
    tgError.value = t('login.tgCreateFailed')
  } finally {
    tgLoading.value = false
  }
}

// ── 手机号+密码备用登录 ──
const pwdLoginExpanded = ref(false)
const pwdForm = reactive({ username: '', password: '' })
const pwdLoading = ref(false)
const pwdError = ref('')

const handlePwdLogin = async () => {
  pwdError.value = ''
  if (!pwdForm.username.trim() || !pwdForm.password) return
  pwdLoading.value = true
  try {
    const data = await login(pwdForm.username.trim(), pwdForm.password)
    userStore.token = data.access_token
    userStore.userInfo = data.user
    ElMessage.success(t('login.loginSuccess'))
    afterLogin(data.user)
  } catch (e) {
    pwdError.value = e?.response?.data?.detail || t('common.requestFailed')
  } finally {
    pwdLoading.value = false
  }
}

// ── 完善账号（Telegram 注册后设置手机号+密码） ──
const setupVisible = ref(false)
const setupForm = reactive({ phone: '', password: '', confirmPassword: '' })
const setupLoading = ref(false)
const setupError = ref('')

const handleSetup = async () => {
  setupError.value = ''
  if (!setupForm.phone.trim()) { setupError.value = t('setup.phoneRequired'); return }
  if (!setupForm.password || setupForm.password.length < 6) { setupError.value = t('setup.passwordMinLength'); return }
  if (setupForm.password !== setupForm.confirmPassword) { setupError.value = t('setup.passwordMismatch'); return }
  setupLoading.value = true
  try {
    const user = await setupCredentials({ phone: setupForm.phone.trim(), password: setupForm.password })
    userStore.userInfo = user
    setupVisible.value = false
    ElMessage.success(t('setup.success'))
    router.push('/m/shop')
  } catch (e) {
    setupError.value = e?.response?.data?.detail || t('common.requestFailed')
  } finally {
    setupLoading.value = false
  }
}

// ── 首次改密（管理员） ──
const changePwdVisible = ref(false)
const changePwdLoading = ref(false)
const changePwdRef = ref()
const changePwdForm = reactive({ new_password: '', confirm_password: '' })
const changePwdRules = {
  new_password: [
    { required: true, message: () => t('profile.newPasswordPlaceholder'), trigger: 'blur' },
    { min: 6, message: () => t('profile.passwordMinLength'), trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: () => t('profile.confirmPasswordPlaceholder'), trigger: 'blur' },
    { validator: (_, v, cb) => v !== changePwdForm.new_password ? cb(new Error(t('profile.passwordMismatch'))) : cb(), trigger: 'blur' },
  ],
}
const handleChangePwd = async () => {
  try { await changePwdRef.value?.validate() } catch { return }
  changePwdLoading.value = true
  try {
    await changePassword({ old_password: '', new_password: changePwdForm.new_password })
    changePwdVisible.value = false
    if (userStore.userInfo) userStore.userInfo.must_change_password = false
    ElMessage.success(t('profile.passwordChanged'))
    router.push('/m/shop')
  } catch {
    ElMessage.error(t('profile.updateFailed'))
  } finally {
    changePwdLoading.value = false
  }
}

// ── 忘记密码 ──
const forgotVisible = ref(false)
const contactInfo = ref('')
const contactLines = computed(() => contactInfo.value ? contactInfo.value.split('\n') : [])
const showForgotPassword = async () => {
  try {
    const data = await getPublicAnnouncements('contact')
    const item = data?.[0]
    contactInfo.value = item ? (currentLang.value === 'zh' ? item.content_zh : item.content_en) : ''
  } catch { contactInfo.value = '' }
  forgotVisible.value = true
}

onMounted(() => { if (isTgContext) handleTgAutoLogin() })
onUnmounted(() => stopPolling())
</script>

<style scoped lang="scss">
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #F8F9FA;
  padding: 24px;
  position: relative;
}

.login-lang-btn {
  position: absolute;
  top: 20px; right: 20px;
  padding: 6px 16px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background: #fff;
  color: #333;
  font-size: 13px;
  cursor: pointer;
  z-index: 10;
  &:hover { border-color: #1D4ED8; color: #1D4ED8; }
}

/* ── Telegram 全屏自动登录 ── */
.tg-fullscreen {
  background: linear-gradient(160deg, #0d1b2a 0%, #1b2d3e 45%, #0f3460 100%);
}
.tg-autoscreen {
  display: flex; flex-direction: column; align-items: center;
  justify-content: center; gap: 24px; padding: 40px 20px;
}
.tg-auto-body {
  display: flex; flex-direction: column; align-items: center;
  gap: 14px; color: rgba(255,255,255,0.85); font-size: 15px;
}
.tg-auto-err { color: #ff7875; }
.tg-retry-btn {
  padding: 6px 20px; background: #409eff; color: #fff;
  border: none; border-radius: 6px; font-size: 14px; cursor: pointer;
  &:hover { opacity: 0.88; }
}

/* ── 登录卡片 ── */
.login-card {
  width: 100%; max-width: 420px;
  background: #fff; border: 1px solid #E8E8E8;
  border-radius: 4px; padding: 48px 40px;
}
.login-header {
  text-align: center; margin-bottom: 40px;
  .logo-container { display: flex; justify-content: center; margin-bottom: 20px; }
  .logo-image { width: 280px; max-width: 100%; height: auto; }
}

/* ── Telegram 登录区 ── */
.tg-login-section { padding: 12px 0 8px; }
.tg-browser-tip { text-align: center; font-size: 13px; color: #666; margin-bottom: 14px; }
.tg-btn {
  display: inline-flex; align-items: center; gap: 10px;
  background: #229ED9; color: #fff; border: none;
  border-radius: 10px; padding: 12px 28px;
  font-size: 15px; font-weight: 600; cursor: pointer;
  &:hover { opacity: 0.88; }
  &:disabled { opacity: 0.6; cursor: not-allowed; }
}
.tg-waiting { text-align: center; color: #555; font-size: 14px; line-height: 1.8; }
.tg-countdown { font-size: 12px; color: #aaa; margin-top: 2px; }
.tg-cancel-btn {
  margin-top: 8px; background: none; border: 1px solid #ddd;
  border-radius: 6px; padding: 4px 16px; font-size: 13px; color: #999; cursor: pointer;
  &:hover { color: #666; }
}
.tg-error { color: #f56c6c; font-size: 13px; text-align: center; margin-top: 10px; }
.tg-spinner {
  display: inline-block; width: 30px; height: 30px;
  border: 3px solid #e0e0e0; border-top-color: #229ED9;
  border-radius: 50%; animation: spin 0.8s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* ── 手机号+密码备用 ── */
.pwd-login-section { margin-top: 16px; }
.pwd-login-divider {
  display: flex; align-items: center; gap: 10px;
  color: #bbb; font-size: 12px; margin-bottom: 12px;
  &::before, &::after { content: ''; flex: 1; height: 1px; background: #e8e8e8; }
}
.pwd-login-toggle {
  text-align: center; font-size: 13px; color: #409eff; cursor: pointer;
  &:hover { text-decoration: underline; }
}

/* ── 完善账号弹窗 ── */
.setup-tip { font-size: 13px; color: #666; margin-bottom: 16px; text-align: center; }

/* ── 公共 ── */
.form-error { color: #f56c6c; font-size: 13px; text-align: center; }
.login-footer { margin-top: 24px; text-align: center; padding-top: 16px; border-top: 1px solid #f0f0f0; }
.forgot-link { color: #409eff; font-size: 13px; cursor: pointer; &:hover { text-decoration: underline; } }
.forgot-content { text-align: center; }
.forgot-tip { color: #606266; font-size: 14px; margin-bottom: 16px; }
.forgot-contact { padding: 16px; background: #f5f7fa; border-radius: 6px; font-size: 14px; line-height: 1.8; }
.change-pwd-tip { color: #E6A23C; font-size: 14px; margin-bottom: 20px; text-align: center; }

@media (max-width: 768px) {
  .login-card { padding: 32px 24px; }
}
</style>
