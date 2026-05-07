<template>
  <div class="profile-page">
    <h2>{{ $t('profile.title') }}</h2>

    <!-- 审核状态提示 -->
    <el-alert
      v-if="userStore.isMerchant && !userStore.profileCompleted"
      type="warning"
      :title="$t('profile.pleaseCompleteProfile')"
      :description="$t('profile.completeProfileTip')"
      show-icon
      :closable="false"
      style="margin-bottom: 16px;"
    />
    <el-alert
      v-else-if="userStore.isMerchant && userStore.approvalStatus === 'pending'"
      type="info"
      :title="$t('profile.pendingApproval')"
      :description="$t('profile.pendingApprovalTip')"
      show-icon
      :closable="false"
      style="margin-bottom: 16px;"
    />
    <el-alert
      v-else-if="userStore.isMerchant && userStore.approvalStatus === 'rejected'"
      type="error"
      :title="$t('profile.rejected')"
      :description="(userStore.userInfo?.rejected_reason || '') + ' — ' + $t('profile.rejectedTip')"
      show-icon
      :closable="false"
      style="margin-bottom: 16px;"
    />
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('profile.personalInfo') }}</span>
        </div>
      </template>
      <el-form label-width="120px" class="profile-form">
        <el-form-item :label="$t('login.username')">
          <span class="info-value">{{ userStore.userInfo?.username }}</span>
        </el-form-item>
        <el-form-item :label="$t('profile.name')">
          <div class="editable-field">
            <span class="info-value">{{ userStore.userInfo?.full_name || $t('profile.notSet') }}</span>
            <el-button type="primary" link size="small" @click="editField('full_name', $t('profile.name'), userStore.userInfo?.full_name)">
              {{ $t('common.edit') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item :label="$t('profile.phone')">
          <div class="editable-field">
            <span class="info-value">{{ userStore.userInfo?.phone || $t('profile.notSet') }}</span>
            <el-button type="primary" link size="small" @click="handlePhoneEdit">
              {{ $t('common.edit') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item :label="$t('profile.address')">
          <div class="editable-field">
            <span class="info-value">{{ userStore.userInfo?.address || $t('profile.notSet') }}</span>
            <el-button type="primary" link size="small" @click="editField('address', $t('profile.address'), userStore.userInfo?.address)">
              {{ $t('common.edit') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item :label="$t('profile.locationUrl')">
          <div class="editable-field">
            <span v-if="userStore.userInfo?.location_url" class="info-value">
              <a :href="userStore.userInfo.location_url" target="_blank" class="link-text">
                <el-icon><Location /></el-icon>
                {{ $t('profile.viewMap') }}
              </a>
            </span>
            <span v-else class="info-value">{{ $t('profile.notSet') }}</span>
            <el-button type="primary" link size="small" @click="editField('location_url', $t('profile.locationUrl'), userStore.userInfo?.location_url)">
              {{ $t('common.edit') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item :label="$t('profile.storePhoto')">
          <div class="store-photo-field">
            <div v-if="userStore.userInfo?.store_photo" class="store-photo-preview">
              <el-image
                :src="userStore.userInfo.store_photo"
                fit="cover"
                style="width: 120px; height: 90px; border-radius: 8px;"
                :preview-src-list="[userStore.userInfo.store_photo]"
              />
              <el-button type="danger" link size="small" @click="removeStorePhoto">{{ $t('common.delete') }}</el-button>
            </div>
            <el-upload
              v-else
              class="store-photo-uploader"
              :show-file-list="false"
              :http-request="handleStorePhotoUpload"
              accept=".jpg,.jpeg,.png,.webp"
            >
              <div class="upload-area" @paste="onPasteStorePhoto">
                <el-icon :size="24"><Plus /></el-icon>
                <span>{{ $t('profile.uploadStorePhoto') }}</span>
              </div>
            </el-upload>
          </div>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 提交审核按钮 -->
    <div v-if="userStore.isMerchant && (!userStore.profileCompleted || userStore.approvalStatus === 'rejected')" style="margin-bottom: 16px;">
      <el-button type="primary" size="large" style="width: 100%;" :loading="submittingProfile" @click="submitProfileForReview">
        {{ userStore.approvalStatus === 'rejected' ? $t('profile.resubmitForReview') : $t('profile.submitForReview') }}
      </el-button>
    </div>

    <!-- 审核中：刷新状态按钮 -->
    <div v-else-if="userStore.isMerchant && userStore.approvalStatus === 'pending'" style="margin-bottom: 16px;">
      <el-button type="primary" size="large" style="width: 100%;" :loading="checkingStatus" @click="checkApprovalStatus">
        {{ $t('profile.checkStatus') }}
      </el-button>
    </div>

    <!-- 账户安全 (仅管理员) -->
    <el-card class="info-card" v-if="userStore.isAdmin">
      <template #header>
        <div class="card-header">
          <span>{{ $t('profile.accountSecurity') }}</span>
        </div>
      </template>
      <el-form label-width="120px">
        <el-form-item :label="$t('profile.changePassword')">
          <el-button type="primary" @click="showPasswordDialog = true">
            {{ $t('profile.changePassword') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="info-card" v-if="userStore.isAdmin">
      <template #header>
        <div class="card-header">
          <span>{{ $t('admin.telegramSettings') }}</span>
        </div>
      </template>
      <el-form label-width="120px">
        <el-form-item :label="$t('admin.telegramId')">
          <div class="editable-field">
            <span class="info-value">{{ userStore.userInfo?.telegram_id || $t('profile.notSet') }}</span>
            <el-button type="primary" link size="small" @click="editField('telegram_id', $t('admin.telegramId'), userStore.userInfo?.telegram_id)">
              {{ $t('common.edit') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item v-if="telegramMiniAppAvailable" :label="$t('admin.bindCurrentTelegram')">
          <el-button type="primary" :loading="bindingTelegram" @click="bindCurrentTelegram">
            {{ userStore.userInfo?.telegram_id ? $t('admin.rebindCurrentTelegram') : $t('admin.bindCurrentTelegram') }}
          </el-button>
        </el-form-item>
        <div class="notify-tip">{{ telegramMiniAppAvailable ? $t('admin.telegramTipMiniApp') : $t('admin.telegramTipManual') }}</div>
      </el-form>
    </el-card>

    <!-- 推送通知设置 -->
    <el-card class="info-card" v-if="userStore.userInfo?.telegram_id">
      <template #header>
        <div class="card-header">
          <span>{{ $t('profile.notificationSettings') }}</span>
        </div>
      </template>
      <el-form label-width="120px">
        <el-form-item :label="$t('profile.notifyEnabled')">
          <el-switch v-model="notifyEnabled" @change="toggleNotify" />
        </el-form-item>
        <div class="notify-tip">{{ $t('profile.notifyTip') }}</div>
      </el-form>
    </el-card>

    <!-- 帮助与支持 -->
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('profile.helpSupport') }}</span>
        </div>
      </template>
      <el-form label-width="120px">
        <el-form-item :label="$t('profile.contactService')">
          <el-button @click="contactService">{{ $t('profile.contactHint') }}</el-button>
        </el-form-item>
        <el-form-item :label="$t('profile.aboutSystem')">
          <el-button @click="showAbout = true">{{ $t('profile.aboutSystem') }}</el-button>
        </el-form-item>
        <el-form-item :label="$t('profile.clearCache')">
          <el-button type="warning" plain @click="clearCache">{{ $t('profile.clearCache') }}</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 编辑个人信息对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      :title="$t('profile.editPrefix') + editLabel"
      width="420px"
      destroy-on-close
    >
      <el-form label-width="80px">
        <el-form-item :label="editLabel">
          <el-input
            v-model="editValue"
            :placeholder="$t('profile.inputPrefix') + editLabel"
            :type="editKey === 'address' ? 'textarea' : 'text'"
            :rows="editKey === 'address' ? 3 : undefined"
            clearable
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="handleSaveProfile">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="showPhoneVerifyDialog"
      :title="$t('profile.verifyPhone')"
      width="420px"
      destroy-on-close
    >
      <el-form label-width="92px">
        <el-form-item :label="$t('profile.phone')">
          <el-input v-model="phoneVerifyForm.phone" :placeholder="$t('register.phonePlaceholder')" />
        </el-form-item>
        <el-form-item v-if="!canDirectVerifyPhone()" :label="$t('register.smsCodePlaceholder')">
          <div class="phone-verify-row">
            <el-input v-model="phoneVerifyForm.code" :placeholder="$t('register.smsCodePlaceholder')" />
            <el-button type="primary" plain :loading="sendingPhoneCode" @click="sendPhoneCode">
              {{ $t('profile.sendTelegramCode') }}
            </el-button>
          </div>
        </el-form-item>
        <div class="notify-tip">{{ canDirectVerifyPhone() ? $t('profile.verifyPhoneMiniAppTip') : $t('profile.verifyPhoneTip') }}</div>
      </el-form>
      <template #footer>
        <el-button @click="showPhoneVerifyDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="verifyingPhone" @click="confirmPhoneVerification">{{ $t('profile.verifyAndSave') }}</el-button>
      </template>
    </el-dialog>

    <!-- 修改密码对话框 -->
    <el-dialog
      v-model="showPasswordDialog"
      :title="$t('profile.changePassword')"
      width="420px"
      destroy-on-close
    >
      <el-form label-width="100px">
        <el-form-item :label="$t('profile.oldPassword')">
          <el-input v-model="passwordForm.old_password" type="password" show-password :placeholder="$t('profile.oldPasswordPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('profile.newPassword')">
          <el-input v-model="passwordForm.new_password" type="password" show-password :placeholder="$t('profile.newPasswordPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('profile.confirmPassword')">
          <el-input v-model="passwordForm.confirm_password" type="password" show-password :placeholder="$t('profile.confirmPasswordPlaceholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="handleChangePassword">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 关于弹窗 -->
    <el-dialog v-model="showAbout" :title="$t('profile.aboutTitle')" width="420px">
      <div class="about-content">
        <div class="about-logo"><img src="/images/logo1.svg" alt="Logo" /></div>
        <p class="about-version">{{ $t('profile.aboutVersion') }}</p>
        <div v-if="aboutInfo.length > 0" class="about-features">
          <p v-for="(item, i) in aboutInfo" :key="i">{{ currentLang === 'zh' ? item.content_zh : item.content_en }}</p>
        </div>
        <div v-else class="about-features">
          <p>{{ $t('profile.aboutFeature1') }}</p>
          <p>{{ $t('profile.aboutFeature2') }}</p>
          <p>{{ $t('profile.aboutFeature3') }}</p>
          <p>{{ $t('profile.aboutFeature4') }}</p>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="showAbout = false">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus/es/components/message/index'
import { ElMessageBox } from 'element-plus/es/components/message-box/index'
import { Location, Plus } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import { updateProfile, changePassword, getPublicAnnouncements, uploadImage, submitForReview, updateAdminTelegram, bindCurrentAdminTelegram, sendPhoneVerificationCode, verifyPhoneVerificationCode, verifyPhoneWithTelegram } from '@/api'
import { getCurrentLanguage } from '@/i18n'
import { getInitData, isTelegramMiniApp } from '@/utils/telegram'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()
const saving = ref(false)
const submittingProfile = ref(false)
const checkingStatus = ref(false)
const editDialogVisible = ref(false)
const showPhoneVerifyDialog = ref(false)
const showPasswordDialog = ref(false)
const showAbout = ref(false)
const currentLang = ref(getCurrentLanguage())
const contactInfo = ref([])
const aboutInfo = ref([])
const notifyEnabled = ref(userStore.userInfo?.notify_enabled !== false)
const telegramMiniAppAvailable = ref(isTelegramMiniApp())
const bindingTelegram = ref(false)
const sendingPhoneCode = ref(false)
const verifyingPhone = ref(false)
const phoneVerifyForm = reactive({ phone: '', code: '' })

const canDirectVerifyPhone = () => telegramMiniAppAvailable.value && !!getInitData()

onMounted(async () => {
  // 刷新用户信息，确保审核状态是最新的
  userStore.fetchUserInfo()
  try {
    const [c, a] = await Promise.all([getPublicAnnouncements('contact'), getPublicAnnouncements('about')])
    contactInfo.value = c
    aboutInfo.value = a
  } catch (e) { /* silent */ }
})

const editKey = ref('')
const editLabel = ref('')
const editValue = ref('')

const editField = (key, label, currentValue) => {
  editKey.value = key
  editLabel.value = label
  editValue.value = currentValue || ''
  editDialogVisible.value = true
}

const handlePhoneEdit = () => {
  if (!userStore.userInfo?.telegram_id) {
    editField('phone', t('profile.phone'), userStore.userInfo?.phone)
    return
  }
  phoneVerifyForm.phone = userStore.userInfo?.phone || ''
  phoneVerifyForm.code = ''
  showPhoneVerifyDialog.value = true
}

const sendPhoneCode = async () => {
  if (canDirectVerifyPhone()) {
    ElMessage.info(t('profile.verifyPhoneMiniAppTip'))
    return
  }
  if (!userStore.userInfo?.telegram_id) {
    ElMessage.warning(t('profile.telegramBindRequired'))
    return
  }
  if (!phoneVerifyForm.phone.trim()) {
    ElMessage.warning(t('profile.pleaseInputPhone'))
    return
  }
  sendingPhoneCode.value = true
  try {
    await sendPhoneVerificationCode({ phone: phoneVerifyForm.phone.trim() })
    ElMessage.success(t('profile.telegramCodeSent'))
  } finally {
    sendingPhoneCode.value = false
  }
}

const confirmPhoneVerification = async () => {
  if (!phoneVerifyForm.phone.trim()) {
    ElMessage.warning(t('profile.pleaseInputPhone'))
    return
  }
  if (canDirectVerifyPhone()) {
    verifyingPhone.value = true
    try {
      const updatedUser = await verifyPhoneWithTelegram({
        phone: phoneVerifyForm.phone.trim(),
        init_data: getInitData(),
      })
      userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
      notifyEnabled.value = updatedUser.notify_enabled !== false
      ElMessage.success(t('profile.phoneVerified'))
      showPhoneVerifyDialog.value = false
    } finally {
      verifyingPhone.value = false
    }
    return
  }
  if (!phoneVerifyForm.code.trim()) {
    ElMessage.warning(t('profile.phoneCodeRequired'))
    return
  }
  verifyingPhone.value = true
  try {
    const updatedUser = await verifyPhoneVerificationCode({
      phone: phoneVerifyForm.phone.trim(),
      code: phoneVerifyForm.code.trim(),
    })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    notifyEnabled.value = updatedUser.notify_enabled !== false
    ElMessage.success(t('profile.phoneVerified'))
    showPhoneVerifyDialog.value = false
  } finally {
    verifyingPhone.value = false
  }
}

const handleSaveProfile = async () => {
  saving.value = true
  try {
    const val = editValue.value.trim()
    let updatedUser
    if (editKey.value === 'telegram_id') {
      if (!val) {
        updatedUser = await updateAdminTelegram({ telegram_id: null })
      } else {
        const telegramId = parseInt(val, 10)
        if (Number.isNaN(telegramId)) {
          ElMessage.warning(t('admin.telegramIdInvalid'))
          return
        }
        updatedUser = await updateAdminTelegram({ telegram_id: telegramId })
      }
    } else {
      const data = { [editKey.value]: val }
      updatedUser = await updateProfile(data)
    }
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    ElMessage.success(t('profile.updateSuccess'))
    editDialogVisible.value = false
  } catch (error) {
    console.error('Update failed:', error)
  } finally {
    saving.value = false
  }
}

const bindCurrentTelegram = async () => {
  if (!telegramMiniAppAvailable.value) {
    ElMessage.warning(t('admin.telegramBindMiniAppOnly'))
    return
  }

  const initData = getInitData()
  if (!initData) {
    ElMessage.warning(t('admin.telegramBindMiniAppOnly'))
    return
  }

  bindingTelegram.value = true
  try {
    const updatedUser = await bindCurrentAdminTelegram(initData)
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    ElMessage.success(t('admin.telegramBindSuccess'))
  } finally {
    bindingTelegram.value = false
  }
}

const toggleNotify = async (val) => {
  try {
    const updatedUser = await updateProfile({ notify_enabled: val })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    ElMessage.success(t('profile.updateSuccess'))
  } catch (error) {
    notifyEnabled.value = !val
    console.error('Toggle notify failed:', error)
  }
}

// 提交资料等待审核
const submitProfileForReview = async () => {
  const info = userStore.userInfo
  if (!info?.full_name || info.full_name.startsWith('TG_')) {
    ElMessage.warning(t('profile.pleaseInputName'))
    return
  }
  if (!info?.phone) {
    ElMessage.warning(t('profile.pleaseInputPhone'))
    return
  }
  if (!info?.address) {
    ElMessage.warning(t('profile.pleaseInputAddress'))
    return
  }
  submittingProfile.value = true
  try {
    await submitForReview()
    await userStore.fetchUserInfo()
    ElMessage.success(t('profile.profileSubmitted'))
  } catch (error) {
    // request.js 拦截器已显示错误
  } finally {
    submittingProfile.value = false
  }
}

// 刷新审核状态
const checkApprovalStatus = async () => {
  checkingStatus.value = true
  try {
    await userStore.fetchUserInfo()
    if (userStore.isApproved) {
      ElMessage.success(t('profile.approved'))
      setTimeout(() => router.push('/merchant/products'), 800)
    } else {
      ElMessage.info(t('profile.pendingApproval'))
    }
  } catch {
    ElMessage.error(t('profile.updateFailed'))
  } finally {
    checkingStatus.value = false
  }
}

const handleStorePhotoUpload = async (options) => {
  try {
    const res = await uploadImage(options.file)
    const url = res.url || res
    const updatedUser = await updateProfile({ store_photo: url })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    ElMessage.success(t('profile.updateSuccess'))
  } catch (error) {
    ElMessage.error(t('product.uploadFailed'))
  }
}

const removeStorePhoto = async () => {
  try {
    await updateProfile({ store_photo: '' })
    userStore.userInfo = { ...userStore.userInfo, store_photo: '' }
    ElMessage.success(t('profile.updateSuccess'))
  } catch (error) { /* silent */ }
}

const onPasteStorePhoto = async (e) => {
  const items = e.clipboardData?.items
  if (!items) return
  for (const item of items) {
    if (item.type.startsWith('image/')) {
      e.preventDefault()
      const file = item.getAsFile()
      if (file) await handleStorePhotoUpload({ file })
      break
    }
  }
}

const passwordForm = reactive({ old_password: '', new_password: '', confirm_password: '' })

const handleChangePassword = async () => {
  if (!passwordForm.old_password || !passwordForm.new_password || !passwordForm.confirm_password) {
    ElMessage.warning(t('profile.fillRequired'))
    return
  }
  if (passwordForm.new_password.length < 6) {
    ElMessage.warning(t('profile.passwordMinLength'))
    return
  }
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    ElMessage.warning(t('profile.passwordMismatch'))
    return
  }
  saving.value = true
  try {
    await changePassword({ old_password: passwordForm.old_password, new_password: passwordForm.new_password })
    ElMessage.success(t('profile.passwordChanged'))
    showPasswordDialog.value = false
    Object.assign(passwordForm, { old_password: '', new_password: '', confirm_password: '' })
  } catch (error) { console.error(error) } finally { saving.value = false }
}

const contactService = () => {
  const lang = getCurrentLanguage()
  let msg = t('profile.contactMessage')
  if (contactInfo.value.length > 0) {
    msg = contactInfo.value.map(c => lang === 'zh' ? c.content_zh : c.content_en).join('\n\n')
  }
  ElMessageBox.alert(msg, t('profile.contactTitle'), { confirmButtonText: t('profile.contactOk') })
}

const clearCache = () => {
  ElMessageBox.confirm(t('profile.clearCacheMessage'), t('profile.clearCacheTitle'), {
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
    type: 'warning',
  }).then(() => {
    cartStore.clear()
    ElMessage.success(t('profile.cacheCleared'))
  }).catch(() => {})
}
</script>

<style scoped>
.profile-page { padding: 20px; max-width: 680px; }
.profile-page h2 { margin-bottom: 20px; }
.info-card { margin-bottom: 20px; }
.card-header { font-weight: 600; }
.profile-form .el-form-item { margin-bottom: 16px; }
.info-value { color: #333; font-size: 14px; }
.editable-field { display: flex; align-items: center; gap: 12px; }
.link-text { color: #409eff; text-decoration: none; display: flex; align-items: center; gap: 4px; }
.link-text:hover { text-decoration: underline; }
.store-photo-field { display: flex; align-items: flex-start; gap: 12px; }
.store-photo-preview { display: flex; flex-direction: column; align-items: center; gap: 6px; }
.store-photo-uploader .upload-area {
  width: 120px; height: 90px; border: 2px dashed #dcdfe6; border-radius: 8px;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 4px; cursor: pointer; transition: all 0.2s; color: #999; font-size: 12px;
}
.store-photo-uploader .upload-area:hover { border-color: #409eff; color: #409eff; }
.notify-tip { padding: 0 12px 12px; font-size: 12px; color: #909399; }
.phone-verify-row { display: flex; gap: 8px; width: 100%; }
.about-content { text-align: center; padding: 10px 0; }
.about-logo img { height: 48px; margin-bottom: 10px; }
.about-version { color: #999; font-size: 13px; margin-bottom: 16px; }
.about-features p { color: #666; font-size: 14px; margin: 6px 0; }

@media (max-width: 767px) {
  .profile-page { padding: 12px; padding-bottom: 70px; }
  .profile-page h2 { font-size: 18px; margin-bottom: 12px; }
  .el-dialog { width: 92vw !important; }
}
</style>
