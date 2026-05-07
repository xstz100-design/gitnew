<template>
  <div class="profile-page">
    <h2>{{ $t('admin.myProfile') }}</h2>

    <!-- 个人信息卡片 -->
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
      </el-form>
    </el-card>

    <!-- Telegram 绑定 -->
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>{{ $t('admin.telegramSettings') }}</span>
        </div>
      </template>
      <el-form label-width="120px" class="profile-form">
        <el-form-item :label="$t('admin.telegramId')">
          <div class="editable-field">
            <span class="info-value">{{ userStore.userInfo?.telegram_id || $t('profile.notSet') }}</span>
            <el-button type="primary" link size="small" @click="showTelegramDialog = true">
              {{ $t('common.edit') }}
            </el-button>
          </div>
        </el-form-item>
        <div class="telegram-tip">
          <el-icon><InfoFilled /></el-icon>
          <span>{{ $t('admin.telegramTip') }}</span>
        </div>
      </el-form>
    </el-card>

    <!-- 账户安全 -->
    <el-card class="info-card">
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

    <!-- 编辑单字段弹窗 -->
    <el-dialog
      v-model="editDialogVisible"
      :title="$t('profile.editPrefix') + editLabel"
      :width="mobile ? '92vw' : '420px'"
    >
      <el-form label-width="80px">
        <el-form-item :label="editLabel">
          <el-input
            v-model="editValue"
            :placeholder="$t('profile.inputPrefix') + editLabel"
            clearable
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="handleSaveProfile">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- Telegram ID 绑定弹窗 -->
    <el-dialog
      v-model="showTelegramDialog"
      :title="$t('admin.telegramSettings')"
      :width="mobile ? '92vw' : '420px'"
    >
      <el-form label-width="120px">
        <el-form-item :label="$t('admin.telegramId')">
          <el-input
            v-model="telegramIdInput"
            :placeholder="$t('admin.telegramIdPlaceholder')"
            clearable
          />
        </el-form-item>
      </el-form>
      <div class="telegram-tip" style="margin-top: 8px;">
        <el-icon><InfoFilled /></el-icon>
        <span>{{ $t('admin.telegramTip') }}</span>
      </div>
      <template #footer>
        <el-button @click="showTelegramDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="handleSaveTelegram">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <!-- 修改密码对话框 -->
    <el-dialog
      v-model="showPasswordDialog"
      :title="$t('profile.changePassword')"
      :width="mobile ? '92vw' : '420px'"
    >
      <el-form label-width="100px">
        <el-form-item :label="$t('profile.oldPassword')">
          <el-input v-model="passwordForm.old_password" type="password" show-password :placeholder="$t('profile.oldPasswordPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('profile.newPassword')">
          <el-input v-model="passwordForm.new_password" type="password" show-password :placeholder="$t('profile.newPasswordPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('profile.confirmPassword')">
          <el-input v-model="passwordForm.confirm_password" type="password" show-password :placeholder="$t('profile.confirmPasswordPlaceholder')" @keyup.enter="handleChangePassword" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="handleChangePassword">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus/es/components/message/index'
import { InfoFilled } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { updateProfile, changePassword, updateAdminTelegram } from '@/api'

const { t } = useI18n()
const userStore = useUserStore()
const saving = ref(false)
const editDialogVisible = ref(false)
const showPasswordDialog = ref(false)
const showTelegramDialog = ref(false)
const telegramIdInput = ref('')

const mobile = ref(window.innerWidth < 768)
const onResize = () => { mobile.value = window.innerWidth < 768 }
onMounted(() => window.addEventListener('resize', onResize))
onBeforeUnmount(() => window.removeEventListener('resize', onResize))

const editKey = ref('')
const editLabel = ref('')
const editValue = ref('')

const editField = (key, label, currentValue) => {
  editKey.value = key
  editLabel.value = label
  editValue.value = currentValue || ''
  editDialogVisible.value = true
}

const handleSaveProfile = async () => {
  saving.value = true
  try {
    const val = editValue.value.trim()
    const data = { [editKey.value]: val }
    const updatedUser = await updateProfile(data)
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    ElMessage.success(t('profile.updateSuccess'))
    editDialogVisible.value = false
  } catch (error) {
    console.error('Update failed:', error)
  } finally {
    saving.value = false
  }
}

const handleSaveTelegram = async () => {
  saving.value = true
  try {
    const val = telegramIdInput.value.trim()
    const telegram_id = val ? parseInt(val) : null
    if (val && isNaN(telegram_id)) {
      ElMessage.warning(t('admin.telegramIdInvalid'))
      saving.value = false
      return
    }
    const updatedUser = await updateAdminTelegram({ telegram_id })
    userStore.userInfo = { ...userStore.userInfo, ...updatedUser }
    ElMessage.success(t('profile.updateSuccess'))
    showTelegramDialog.value = false
  } catch (error) {
    console.error('Telegram bind failed:', error)
  } finally {
    saving.value = false
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
  } catch (error) {
    console.error(error)
  } finally {
    saving.value = false
  }
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
.telegram-tip {
  display: flex; align-items: flex-start; gap: 6px;
  padding: 10px 12px; background: #f0f9ff; border-radius: 6px;
  font-size: 12px; color: #909399; line-height: 1.5;
}
.telegram-tip .el-icon { color: #409eff; margin-top: 2px; flex-shrink: 0; }

@media (max-width: 767px) {
  .profile-page { padding: 12px; padding-bottom: 70px; }
  .profile-page h2 { font-size: 18px; margin-bottom: 12px; }
}
</style>
