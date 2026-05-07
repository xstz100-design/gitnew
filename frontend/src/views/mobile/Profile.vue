<template>
  <div class="mobile-profile">
    <van-nav-bar :title="$t('profile.title')" fixed placeholder />
    
    <!-- 审核状态横幅 -->
    <div v-if="userStore.isMerchant && !userStore.isApproved" class="status-banner">
      <div v-if="userStore.approvalStatus === 'pending'" class="banner pending">
        <van-icon name="clock-o" size="20" />
        <div class="banner-text">
          <strong v-if="!userStore.profileCompleted">{{ $t('profile.pleaseCompleteProfile') }}</strong>
          <strong v-else>{{ $t('profile.pendingApproval') }}</strong>
          <p v-if="userStore.profileCompleted">{{ $t('profile.pendingApprovalTip') }}</p>
          <p v-else>{{ $t('profile.completeProfileTip') }}</p>
        </div>
      </div>
      <div v-else-if="userStore.approvalStatus === 'rejected'" class="banner rejected">
        <van-icon name="warning-o" size="20" />
        <div class="banner-text">
          <strong>{{ $t('profile.rejected') }}</strong>
          <p v-if="userStore.userInfo?.rejected_reason">{{ $t('profile.rejectedReason') }}: {{ userStore.userInfo.rejected_reason }}</p>
          <p>{{ $t('profile.rejectedTip') }}</p>
        </div>
      </div>
    </div>

    <!-- 用户信息卡片 -->
    <div class="profile-header">
      <div class="avatar">
        <van-icon name="user-o" size="32" color="#fff" />
      </div>
      <div class="user-info">
        <div class="user-name">{{ userStore.userInfo?.full_name || $t('profile.notLoggedIn') }}</div>
        <div class="user-phone">{{ userStore.userInfo?.phone || '' }}</div>
      </div>
    </div>
    
    <!-- 个人信息（可编辑） -->
    <van-cell-group inset :title="$t('profile.personalInfo')">
      <van-cell
        :title="$t('profile.name')"
        :value="userStore.userInfo?.full_name || $t('profile.notSet')"
        icon="contact"
        is-link
        @click="editField('full_name', $t('profile.name'), userStore.userInfo?.full_name)"
      />
      <van-cell
        :title="$t('profile.phone')"
        :value="userStore.userInfo?.phone || $t('profile.notSet')"
        icon="phone-o"
        is-link
        @click="handlePhoneEdit"
      />
      <van-cell
        :title="$t('profile.address')"
        :value="userStore.userInfo?.address || $t('profile.notSet')"
        icon="location-o"
        is-link
        @click="editField('address', $t('profile.address'), userStore.userInfo?.address)"
      />
      <van-cell
        :title="$t('profile.locationUrl')"
        icon="guide-o"
        is-link
        @click="editField('location_url', $t('profile.locationUrl'), userStore.userInfo?.location_url)"
      >
        <template #value>
          <a
            v-if="userStore.userInfo?.location_url"
            :href="userStore.userInfo.location_url"
            target="_blank"
            @click.stop
            class="map-link"
          >{{ $t('profile.viewMap') }}</a>
          <span v-else>{{ $t('profile.notSet') }}</span>
        </template>
      </van-cell>
      <van-cell :title="$t('profile.storePhoto')" icon="photograph">
        <template #value>
          <div class="store-photo-cell">
            <van-image
              v-if="userStore.userInfo?.store_photo"
              :src="userStore.userInfo.store_photo"
              width="60"
              height="45"
              fit="cover"
              radius="4"
              @click="previewStorePhoto"
            />
            <van-uploader
              v-else
              :after-read="onStorePhotoRead"
              :max-count="1"
              accept="image/*"
            >
              <van-button size="mini" type="primary" plain>{{ $t('profile.uploadStorePhoto') }}</van-button>
            </van-uploader>
            <van-icon v-if="userStore.userInfo?.store_photo" name="cross" class="remove-photo" @click="removeStorePhoto" />
          </div>
        </template>
      </van-cell>
    </van-cell-group>

    <!-- 资料未完成时显示提交按钮 -->
    <div v-if="userStore.isMerchant && !userStore.profileCompleted" class="submit-profile-section">
      <van-button type="primary" block round @click="submitProfileForReview" :loading="submittingProfile">
        {{ $t('profile.submitForReview') }}
      </van-button>
      <p class="submit-hint">{{ $t('profile.submitHint') }}</p>
    </div>

    <!-- 已提交待审核时显示刷新状态按钮 -->
    <div v-else-if="userStore.isMerchant && userStore.approvalStatus === 'pending'" class="submit-profile-section">
      <van-button type="primary" block round @click="checkApprovalStatus" :loading="checkingStatus" icon="replay">
        {{ $t('profile.checkStatus') }}
      </van-button>
    </div>

    <!-- 被拒绝时显示重新提交 -->
    <div v-else-if="userStore.isMerchant && userStore.approvalStatus === 'rejected'" class="submit-profile-section">
      <van-button type="warning" block round @click="submitProfileForReview" :loading="submittingProfile">
        {{ $t('profile.resubmitForReview') }}
      </van-button>
    </div>
    
    <van-cell-group inset :title="$t('profile.accountSecurity')" v-if="userStore.isAdmin">
      <van-cell
        :title="$t('profile.changePassword')"
        is-link
        icon="lock"
        @click="showPasswordDialog = true"
      />
    </van-cell-group>
    
    <!-- Telegram 通知设置 (仅管理员) -->
    <van-cell-group v-if="userStore.userInfo?.role === 'admin'" inset :title="$t('admin.telegramSettings')">
      <van-cell
        :title="$t('admin.telegramId')"
        :value="userStore.userInfo?.telegram_id || $t('profile.notSet')"
        icon="chat-o"
        is-link
        @click="editField('telegram_id', $t('admin.telegramId'), userStore.userInfo?.telegram_id)"
      />
      <van-cell
        v-if="telegramMiniAppAvailable"
        :title="userStore.userInfo?.telegram_id ? $t('admin.rebindCurrentTelegram') : $t('admin.bindCurrentTelegram')"
        icon="link-o"
        is-link
        @click="bindCurrentTelegram"
      />
      <div class="telegram-hint">
        <span>{{ telegramMiniAppAvailable ? $t('admin.telegramTipMiniApp') : $t('admin.telegramTipManual') }}</span>
      </div>
    </van-cell-group>

    <!-- 推送通知设置 -->
    <van-cell-group v-if="userStore.userInfo?.telegram_id" inset :title="$t('profile.notificationSettings')">
      <van-cell :title="$t('profile.notifyEnabled')" icon="bullhorn-o">
        <template #right-icon>
          <van-switch v-model="notifyEnabled" size="20" @change="toggleNotify" />
        </template>
      </van-cell>
      <div class="telegram-hint">
        <span>{{ $t('profile.notifyTip') }}</span>
      </div>
    </van-cell-group>
    
    <van-cell-group inset :title="$t('profile.helpSupport')">
      <van-cell
        :title="$t('profile.contactService')"
        is-link
        icon="service-o"
        @click="contactService"
        :label="$t('profile.contactHint')"
      />
      <van-cell
        :title="$t('profile.aboutSystem')"
        is-link
        icon="info-o"
        @click="showAbout = true"
      />
      <van-cell
        :title="$t('profile.clearCache')"
        is-link
        icon="delete-o"
        @click="clearCache"
      />
    </van-cell-group>
    
    <div class="logout-section">
      <van-button
        plain
        block
        type="default"
        @click="handleLogout"
      >
        {{ $t('profile.logout') }}
      </van-button>
    </div>
    
    <!-- 编辑个人信息弹窗 -->
    <van-dialog
      v-model:show="showEditDialog"
      :title="$t('profile.editPrefix') + editLabel"
      show-cancel-button
      :before-close="handleSaveProfile"
    >
      <div class="edit-form">
        <van-field
          v-model="editValue"
          :label="editLabel"
          :placeholder="$t('profile.inputPrefix') + editLabel"
          :type="editKey === 'phone' ? 'tel' : 'text'"
          clearable
        />
      </div>
    </van-dialog>

    <van-dialog
      v-model:show="showPhoneVerifyDialog"
      :title="$t('profile.verifyPhone')"
      show-cancel-button
      :confirm-button-text="$t('profile.verifyAndSave')"
      :before-close="handleVerifyPhone"
    >
      <div class="edit-form">
        <van-field
          v-model="phoneVerifyForm.phone"
          :label="$t('profile.phone')"
          :placeholder="$t('register.phonePlaceholder')"
          type="tel"
          clearable
        />
        <van-field
          v-if="!canDirectVerifyPhone()"
          v-model="phoneVerifyForm.code"
          :label="$t('register.smsCodePlaceholder')"
          :placeholder="$t('register.smsCodePlaceholder')"
          clearable
        >
          <template #button>
            <van-button size="small" type="primary" plain :loading="sendingPhoneCode" @click.stop="sendPhoneCode">
              {{ $t('profile.sendTelegramCode') }}
            </van-button>
          </template>
        </van-field>
        <div class="telegram-hint">
          <span>{{ canDirectVerifyPhone() ? $t('profile.verifyPhoneMiniAppTip') : $t('profile.verifyPhoneTip') }}</span>
        </div>
      </div>
    </van-dialog>
    
    <!-- 修改密码弹窗 -->
    <van-dialog
      v-model:show="showPasswordDialog"
      :title="$t('profile.changePassword')"
      show-cancel-button
      :before-close="handleChangePassword"
    >
      <div class="password-form">
        <van-field
          v-model="passwordForm.old_password"
          type="password"
          :label="$t('profile.oldPassword')"
          :placeholder="$t('profile.oldPasswordPlaceholder')"
        />
        <van-field
          v-model="passwordForm.new_password"
          type="password"
          :label="$t('profile.newPassword')"
          :placeholder="$t('profile.newPasswordPlaceholder')"
        />
        <van-field
          v-model="passwordForm.confirm_password"
          type="password"
          :label="$t('profile.confirmPassword')"
          :placeholder="$t('profile.confirmPasswordPlaceholder')"
        />
      </div>
    </van-dialog>
    
    <!-- 关于弹窗 -->
    <van-dialog
      v-model:show="showAbout"
      :title="$t('profile.aboutTitle')"
      :confirm-button-text="$t('common.confirm')"
    >
      <div class="about-content">
        <div class="about-logo">
          <img src="/images/logo1.svg" alt="Logo" />
        </div>
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
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onActivated } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showToast, showDialog, showImagePreview } from 'vant'
import { useUserStore } from '@/stores/user'
import { useCartStore } from '@/stores/cart'
import { changePassword, updateProfile, getPublicAnnouncements, uploadImage, submitForReview, updateAdminTelegram, bindCurrentAdminTelegram, sendPhoneVerificationCode, verifyPhoneVerificationCode, verifyPhoneWithTelegram } from '@/api'
import { hapticFeedback } from '@/utils/device'
import { getCurrentLanguage } from '@/i18n'
import { getInitData, isTelegramMiniApp } from '@/utils/telegram'

const { t } = useI18n()
const router = useRouter()
const userStore = useUserStore()
const cartStore = useCartStore()

const showAbout = ref(false)
const showPasswordDialog = ref(false)
const contactInfo = ref([])
const aboutInfo = ref([])
const submittingProfile = ref(false)
const checkingStatus = ref(false)
const notifyEnabled = ref(userStore.userInfo?.notify_enabled !== false)
const telegramMiniAppAvailable = ref(isTelegramMiniApp())
const bindingTelegram = ref(false)
// 编辑个人信息
const showEditDialog = ref(false)
const showPhoneVerifyDialog = ref(false)
const editKey = ref('')
const editLabel = ref('')
const editValue = ref('')
const sendingPhoneCode = ref(false)
const phoneVerifyForm = reactive({ phone: '', code: '' })

const canDirectVerifyPhone = () => telegramMiniAppAvailable.value && !!getInitData()

const editField = (key, label, currentValue) => {
  editKey.value = key
  editLabel.value = label
  editValue.value = currentValue || ''
  showEditDialog.value = true
  hapticFeedback('light')
}

const handlePhoneEdit = () => {
  if (!userStore.userInfo?.telegram_id) {
    editField('phone', t('profile.phone'), userStore.userInfo?.phone)
    return
  }
  phoneVerifyForm.phone = userStore.userInfo?.phone || ''
  phoneVerifyForm.code = ''
  showPhoneVerifyDialog.value = true
  hapticFeedback('light')
}

const sendPhoneCode = async () => {
  if (canDirectVerifyPhone()) {
    showToast(t('profile.verifyPhoneMiniAppTip'))
    return
  }
  if (!userStore.userInfo?.telegram_id) {
    showToast(t('profile.telegramBindRequired'))
    return
  }
  if (!phoneVerifyForm.phone.trim()) {
    showToast(t('profile.pleaseInputPhone'))
    return
  }
  sendingPhoneCode.value = true
  try {
    await sendPhoneVerificationCode({ phone: phoneVerifyForm.phone.trim() })
    hapticFeedback('success')
    showSuccessToast(t('profile.telegramCodeSent'))
  } finally {
    sendingPhoneCode.value = false
  }
}

const handleVerifyPhone = async (action) => {
  if (action !== 'confirm') return true
  if (!phoneVerifyForm.phone.trim()) {
    showToast(t('profile.pleaseInputPhone'))
    return false
  }
  if (canDirectVerifyPhone()) {
    try {
      const updatedUser = await verifyPhoneWithTelegram({
        phone: phoneVerifyForm.phone.trim(),
        init_data: getInitData(),
      })
      userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
      notifyEnabled.value = updatedUser.notify_enabled !== false
      hapticFeedback('success')
      showSuccessToast(t('profile.phoneVerified'))
      return true
    } catch {
      return false
    }
  }
  if (!phoneVerifyForm.code.trim()) {
    showToast(t('profile.phoneCodeRequired'))
    return false
  }
  try {
    const updatedUser = await verifyPhoneVerificationCode({
      phone: phoneVerifyForm.phone.trim(),
      code: phoneVerifyForm.code.trim(),
    })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    notifyEnabled.value = updatedUser.notify_enabled !== false
    hapticFeedback('success')
    showSuccessToast(t('profile.phoneVerified'))
    return true
  } catch {
    return false
  }
}

const handleSaveProfile = async (action) => {
  if (action !== 'confirm') return true
  
  if (!editValue.value.trim()) {
    showToast(t('profile.inputRequired'))
    return false
  }
  
  try {
    let updatedUser
    if (editKey.value === 'telegram_id') {
      // 管理员绑定 Telegram ID 使用专用接口
      const telegram_id = parseInt(editValue.value.trim())
      if (isNaN(telegram_id)) {
        showToast(t('admin.telegramIdInvalid') || 'Telegram ID 格式错误')
        return false
      }
      updatedUser = await updateAdminTelegram({ telegram_id })
    } else {
      const data = { [editKey.value]: editValue.value.trim() }
      updatedUser = await updateProfile(data)
    }
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    hapticFeedback('success')
    showSuccessToast(t('profile.updateSuccess'))
    return true
  } catch (error) {
    showToast(t('profile.updateFailed'))
    return false
  }
}

const bindCurrentTelegram = async () => {
  if (!telegramMiniAppAvailable.value) {
    showToast(t('admin.telegramBindMiniAppOnly'))
    return
  }

  const initData = getInitData()
  if (!initData) {
    showToast(t('admin.telegramBindMiniAppOnly'))
    return
  }

  bindingTelegram.value = true
  try {
    const updatedUser = await bindCurrentAdminTelegram(initData)
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    hapticFeedback('success')
    showSuccessToast(t('admin.telegramBindSuccess'))
  } finally {
    bindingTelegram.value = false
  }
}

const toggleNotify = async (val) => {
  try {
    const updatedUser = await updateProfile({ notify_enabled: val })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    hapticFeedback('success')
    showSuccessToast(t('profile.updateSuccess'))
  } catch (error) {
    notifyEnabled.value = !val
    showToast(t('profile.updateFailed'))
  }
}

// 提交资料等待审核
const submitProfileForReview = async () => {
  const info = userStore.userInfo
  if (!info?.full_name || info.full_name.startsWith('TG_')) {
    showToast(t('profile.pleaseInputName'))
    return
  }
  if (!info?.phone) {
    showToast(t('profile.pleaseInputPhone'))
    return
  }
  if (!info?.address) {
    showToast(t('profile.pleaseInputAddress'))
    return
  }
  submittingProfile.value = true
  try {
    await submitForReview()
    await userStore.fetchUserInfo()
    hapticFeedback('success')
    showSuccessToast(t('profile.profileSubmitted'))
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
    hapticFeedback('light')
    if (userStore.isApproved) {
      hapticFeedback('success')
      showSuccessToast(t('profile.approved'))
      // 审核通过，延迟后跳转到商城
      setTimeout(() => router.push('/m/shop'), 800)
    } else {
      showToast(t('profile.pendingApproval'))
    }
  } catch {
    showToast(t('profile.updateFailed'))
  } finally {
    checkingStatus.value = false
  }
}

// 门面照片上传
const onStorePhotoRead = async (file) => {
  try {
    const res = await uploadImage(file.file)
    const url = res.url || res
    const updatedUser = await updateProfile({ store_photo: url })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    hapticFeedback('success')
    showSuccessToast(t('profile.updateSuccess'))
  } catch (error) {
    showToast(t('product.uploadFailed'))
  }
}

const previewStorePhoto = () => {
  if (userStore.userInfo?.store_photo) {
    showImagePreview([userStore.userInfo.store_photo])
  }
}

const removeStorePhoto = async () => {
  try {
    await updateProfile({ store_photo: '' })
    userStore.userInfo = { ...userStore.userInfo, store_photo: '' }
    hapticFeedback('success')
    showSuccessToast(t('profile.updateSuccess'))
  } catch (error) { /* silent */ }
}

const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

// 修改密码
const handleChangePassword = async (action) => {
  if (action !== 'confirm') return true
  
  if (!passwordForm.old_password || !passwordForm.new_password) {
    showToast(t('profile.fillRequired'))
    return false
  }
  
  if (passwordForm.new_password.length < 6) {
    showToast(t('profile.passwordMinLength'))
    return false
  }
  
  if (passwordForm.new_password !== passwordForm.confirm_password) {
    showToast(t('profile.passwordMismatch'))
    return false
  }
  
  try {
    await changePassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password,
    })
    hapticFeedback('success')
    showSuccessToast(t('profile.passwordChanged'))
    passwordForm.old_password = ''
    passwordForm.new_password = ''
    passwordForm.confirm_password = ''
    return true
  } catch (error) {
    return false
  }
}

// 联系客服
const contactService = () => {
  hapticFeedback('light')
  const lang = getCurrentLanguage()
  let msg = t('profile.contactMessage')
  if (contactInfo.value.length > 0) {
    msg = contactInfo.value.map(c => lang === 'zh' ? c.content_zh : c.content_en).join('\n\n')
  }
  showDialog({
    title: t('profile.contactTitle'),
    message: msg,
    confirmButtonText: t('profile.contactOk'),
  })
}

// 清除缓存
const clearCache = async () => {
  const confirmed = await showDialog({
    title: t('profile.clearCacheTitle'),
    message: t('profile.clearCacheMessage'),
    showCancelButton: true,
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
  }).catch(() => false)
  
  if (confirmed !== false) {
    cartStore.clear()
    hapticFeedback('success')
    showSuccessToast(t('profile.cacheCleared'))
  }
}

// 退出登录
const handleLogout = async () => {
  const confirmed = await showDialog({
    title: t('profile.logoutTitle'),
    message: t('profile.logoutMessage'),
    showCancelButton: true,
    confirmButtonText: t('common.confirm'),
    cancelButtonText: t('common.cancel'),
  }).catch(() => false)
  
  if (confirmed !== false) {
    userStore.logout()
    cartStore.clear()
    hapticFeedback('success')
    router.push('/login')
  }
}

const currentLang = ref(getCurrentLanguage())

// 加载客服和关于信息
const loadContactAbout = async () => {
  try {
    const [contact, about] = await Promise.all([
      getPublicAnnouncements('contact'),
      getPublicAnnouncements('about'),
    ])
    contactInfo.value = contact
    aboutInfo.value = about
  } catch {
    // 静默处理
  }
}

onMounted(() => {
  loadContactAbout()
  // 刷新用户信息，确保审核状态是最新的
  userStore.fetchUserInfo()
})

// keep-alive 激活时也刷新用户信息
onActivated(() => {
  userStore.fetchUserInfo()
})
</script>

<style scoped>
.mobile-profile {
  min-height: var(--tg-viewport-height, 100vh);
  background: var(--bg-gray, #f7f7f7);
  padding-bottom: calc(60px + var(--tg-content-safe-area-inset-bottom, env(safe-area-inset-bottom, 0px)));
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 28px 20px;
  background: var(--primary-color, #2b2b2b);
  color: #fff;
}

.avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: rgba(255,255,255,0.15);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-name {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 4px;
}

.user-phone {
  font-size: 13px;
  opacity: 0.8;
}

.logout-section {
  padding: 24px 16px;
}

.password-form {
  padding: 8px 0;
}

.about-content {
  padding: 20px;
  text-align: center;
}

.about-logo {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
  
  img {
    width: 200px;
    max-width: 100%;
    height: auto;
    object-fit: contain;
  }
}

.about-version {
  font-size: 13px;
  color: #999;
  margin: 0 0 16px;
}

.about-features {
  text-align: left;
  padding: 12px 16px;
  background: #f5f5f5;
  border-radius: 4px;
}

.about-features p {
  font-size: 13px;
  color: #666;
  margin: 0;
  padding: 4px 0;
}

.about-features p::before {
  content: '• ';
  color: #999;
}

.map-link {
  color: #1989fa;
  text-decoration: none;
}

.store-photo-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.remove-photo {
  color: #ee0a24;
  font-size: 16px;
  cursor: pointer;
}

.telegram-hint {
  padding: 8px 16px 12px;
  font-size: 12px;
  color: #999;
  line-height: 1.5;
}

/* 审核状态横幅 */
.status-banner {
  margin: 0 12px 8px;
}
.banner {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 8px;
  margin-top: 8px;
}
.banner.pending {
  background: #FFF7E6;
  border: 1px solid #FFD591;
  color: #AD6800;
}
.banner.rejected {
  background: #FFF1F0;
  border: 1px solid #FFA39E;
  color: #CF1322;
}
.banner-text strong {
  display: block;
  font-size: 14px;
  margin-bottom: 4px;
}
.banner-text p {
  font-size: 12px;
  margin: 0;
  opacity: 0.85;
}

/* 提交审核按钮 */
.submit-profile-section {
  padding: 16px;
}
.submit-hint {
  text-align: center;
  font-size: 12px;
  color: #999;
  margin: 8px 0 0;
}
</style>
