<template>
  <div class="page-container">
    <div class="page-header">
      <h2>{{ $t('admin.myProfile') }}</h2>
    </div>

    <!-- 个人信息 -->
    <div class="setting-section">
      <div class="section-title-row">
        <span class="section-title">{{ $t('profile.personalInfo') }}</span>
        <van-button type="primary" size="small" plain @click="openEditName">{{ $t('common.edit') }}</van-button>
      </div>
      <van-cell-group inset>
        <van-cell :title="$t('login.username')" :value="userStore.userInfo?.username" />
        <van-cell :title="$t('profile.name')" :value="userStore.userInfo?.full_name || $t('profile.notSet')" />
        <van-cell :title="$t('profile.phone')" :value="userStore.userInfo?.phone || $t('profile.notSet')" />
      </van-cell-group>
    </div>

    <!-- 账户安全 -->
    <div class="setting-section">
      <div class="section-title">{{ $t('profile.accountSecurity') }}</div>
      <van-cell-group inset>
        <van-cell :title="$t('profile.changePassword')" is-link @click="showPasswordDialog = true" />
      </van-cell-group>
    </div>

    <!-- Telegram 通知设置 -->
    <div class="setting-section">
      <div class="section-title">{{ $t('admin.telegramSettings') }}</div>
      <van-cell-group inset>
        <van-cell :title="$t('admin.telegramId')" :value="userStore.userInfo?.telegram_id || $t('profile.notSet')" is-link @click="openEditTelegram" />
        <van-cell v-if="telegramMiniAppAvailable">
          <template #title>
            <van-button type="primary" size="small" plain :loading="bindingTelegram" @click="bindCurrentTelegram">
              {{ userStore.userInfo?.telegram_id ? $t('admin.rebindCurrentTelegram') : $t('admin.bindCurrentTelegram') }}
            </van-button>
          </template>
        </van-cell>
      </van-cell-group>
      <div class="notify-tip">{{ telegramMiniAppAvailable ? $t('admin.telegramTipMiniApp') : $t('admin.telegramTipManual') }}</div>
    </div>

    <!-- 推送通知 -->
    <div v-if="userStore.userInfo?.telegram_id" class="setting-section">
      <div class="section-title">{{ $t('profile.notificationSettings') }}</div>
      <van-cell-group inset>
        <van-cell :title="$t('profile.notifyEnabled')">
          <template #right-icon>
            <van-switch v-model="notifyEnabled" size="20" @change="toggleNotify" />
          </template>
        </van-cell>
      </van-cell-group>
      <div class="notify-tip">{{ $t('profile.notifyTip') }}</div>
    </div>

    <!-- 帮助与支持 -->
    <div class="setting-section">
      <div class="section-title">{{ $t('profile.helpSupport') }}</div>
      <van-cell-group inset>
        <van-cell :title="$t('profile.contactService')" is-link @click="contactService">
          <template #value>
            <span style="color: #1989fa;">{{ contactPhone || $t('profile.contactHint') }}</span>
          </template>
        </van-cell>
        <van-cell :title="$t('profile.clearCache')" is-link @click="clearCache" />
      </van-cell-group>
    </div>

    <!-- 退出登录 -->
    <div style="padding: 16px 0 40px;">
      <van-button block type="danger" plain @click="handleLogout">{{ $t('profile.logout') }}</van-button>
    </div>

    <!-- 修改密码弹窗 -->
    <van-popup v-model:show="showPasswordDialog" position="bottom" round :style="{ minHeight: '55vh' }" destroy-on-close>
      <van-nav-bar
        :title="$t('profile.changePassword')"
        :left-text="$t('common.cancel')"
        :right-text="$t('common.confirm')"
        @click-left="showPasswordDialog = false"
        @click-right="handleChangePassword"
      />
      <van-cell-group inset style="margin-top: 8px;">
        <van-field v-model="passwordForm.old_password" type="password" :label="$t('profile.oldPassword')" :placeholder="$t('profile.oldPasswordPlaceholder')" />
        <van-field v-model="passwordForm.new_password" type="password" :label="$t('profile.newPassword')" :placeholder="$t('profile.newPasswordPlaceholder')" />
        <van-field v-model="passwordForm.confirm_password" type="password" :label="$t('profile.confirmPassword')" :placeholder="$t('profile.confirmPasswordPlaceholder')" />
      </van-cell-group>
      <div style="padding: 16px;">
        <van-button block type="primary" :loading="saving" @click="handleChangePassword">{{ $t('common.confirm') }}</van-button>
      </div>
    </van-popup>

    <!-- 编辑姓名弹窗 -->
    <van-popup v-model:show="editNameVisible" position="bottom" round :style="{ minHeight: '35vh' }" destroy-on-close>
      <van-nav-bar
        :title="$t('profile.editPrefix') + $t('profile.name')"
        :left-text="$t('common.cancel')"
        :right-text="$t('common.confirm')"
        @click-left="editNameVisible = false"
        @click-right="handleSaveName"
      />
      <van-cell-group inset style="margin-top: 8px;">
        <van-field v-model="editNameVal" :label="$t('profile.name')" :placeholder="$t('profile.inputPrefix') + $t('profile.name')" clearable />
      </van-cell-group>
      <div style="padding: 16px;">
        <van-button block type="primary" :loading="savingProfile" @click="handleSaveName">{{ $t('common.confirm') }}</van-button>
      </div>
    </van-popup>

    <!-- 编辑 Telegram ID 弹窗 -->
    <van-popup v-model:show="editTgVisible" position="bottom" round :style="{ minHeight: '35vh' }" destroy-on-close>
      <van-nav-bar
        :title="$t('admin.telegramId')"
        :left-text="$t('common.cancel')"
        :right-text="$t('common.confirm')"
        @click-left="editTgVisible = false"
        @click-right="handleSaveTelegram"
      />
      <van-cell-group inset style="margin-top: 8px;">
        <van-field v-model="editTgVal" :label="$t('admin.telegramId')" placeholder="123456789" clearable />
      </van-cell-group>
      <div style="padding: 16px;">
        <van-button block type="primary" :loading="savingTg" @click="handleSaveTelegram">{{ $t('common.confirm') }}</van-button>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { showSuccessToast, showFailToast, showConfirmDialog } from 'vant'
import { useUserStore } from '@/stores/user'
import { changePassword, updateProfile, updateAdminTelegram, bindCurrentAdminTelegram, getContactInfo } from '@/api'
import { isTelegramMiniApp, getInitData } from '@/utils/telegram'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()

const contactPhone = ref('')
onMounted(async () => {
  try {
    const info = await getContactInfo()
    contactPhone.value = info.phone || ''
  } catch {}
})

const saving = ref(false)
const showPasswordDialog = ref(false)
const passwordForm = reactive({ old_password: '', new_password: '', confirm_password: '' })

const handleChangePassword = async () => {
  if (!passwordForm.old_password || !passwordForm.new_password || !passwordForm.confirm_password) {
    showFailToast(t('profile.fillRequired')); return
  }
  if (passwordForm.new_password.length < 6) {
    showFailToast(t('profile.passwordMinLength')); return
  }
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    showFailToast(t('profile.passwordMismatch')); return
  }
  saving.value = true
  try {
    await changePassword({ old_password: passwordForm.old_password, new_password: passwordForm.new_password })
    showSuccessToast(t('profile.passwordChanged'))
    showPasswordDialog.value = false
    Object.assign(passwordForm, { old_password: '', new_password: '', confirm_password: '' })
  } catch {}
  finally { saving.value = false }
}

const editNameVisible = ref(false)
const editNameVal = ref('')
const savingProfile = ref(false)

const openEditName = () => {
  editNameVal.value = userStore.userInfo?.full_name || ''
  editNameVisible.value = true
}

const handleSaveName = async () => {
  savingProfile.value = true
  try {
    const updated = await updateProfile({ full_name: editNameVal.value.trim() })
    userStore.userInfo = { ...userStore.userInfo, ...updated }
    showSuccessToast(t('profile.updateSuccess'))
    editNameVisible.value = false
  } catch {}
  finally { savingProfile.value = false }
}

const editTgVisible = ref(false)
const editTgVal = ref('')
const savingTg = ref(false)
const bindingTelegram = ref(false)
const telegramMiniAppAvailable = ref(isTelegramMiniApp())
const notifyEnabled = ref(userStore.userInfo?.notify_enabled !== false)

const openEditTelegram = () => {
  editTgVal.value = userStore.userInfo?.telegram_id || ''
  editTgVisible.value = true
}

const handleSaveTelegram = async () => {
  savingTg.value = true
  try {
    const updated = await updateAdminTelegram({ telegram_id: editTgVal.value.trim() })
    userStore.userInfo = { ...userStore.userInfo, ...updated }
    showSuccessToast(t('profile.updateSuccess'))
    editTgVisible.value = false
  } catch {}
  finally { savingTg.value = false }
}

const bindCurrentTelegram = async () => {
  const initData = getInitData()
  if (!initData) { showFailToast(t('admin.notInTelegram')); return }
  bindingTelegram.value = true
  try {
    const updated = await bindCurrentAdminTelegram(initData)
    userStore.userInfo = { ...userStore.userInfo, ...updated }
    showSuccessToast(t('admin.telegramBound'))
  } catch {}
  finally { bindingTelegram.value = false }
}

const toggleNotify = async (val) => {
  try {
    const updated = await updateProfile({ notify_enabled: val })
    userStore.userInfo = { ...userStore.userInfo, ...updated }
    showSuccessToast(t('profile.updateSuccess'))
  } catch {
    notifyEnabled.value = !val
    showFailToast(t('profile.updateFailed'))
  }
}

const contactService = () => {
  if (contactPhone.value) {
    window.open('tel:' + contactPhone.value)
  } else {
    showFailToast(t('profile.contactHint'))
  }
}

const clearCache = () => {
  localStorage.clear()
  sessionStorage.clear()
  showSuccessToast(t('profile.cacheCleared'))
}

const handleLogout = async () => {
  try {
    await showConfirmDialog({
      title: t('admin.hint'),
      message: t('admin.logoutConfirm'),
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
    })
    userStore.logout()
    router.push('/login')
  } catch {}
}
</script>

<style scoped>
.setting-section {
  background: #fff;
  border-radius: 10px;
  padding: 16px;
  margin-bottom: 16px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
  margin-bottom: 12px;
}

.section-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.section-title-row .section-title {
  margin-bottom: 0;
}

.notify-tip {
  font-size: 12px;
  color: #999;
  margin-top: 8px;
  padding: 0 4px;
}
</style>
