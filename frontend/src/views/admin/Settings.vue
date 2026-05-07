<template>
  <div class="settings-page">
    <div class="page-header">
      <h2>{{ $t('settings.title') }}</h2>
    </div>

    <el-card class="setting-card" shadow="never">
      <template #header>
        <span>{{ $t('settings.roleChatIds') }}</span>
      </template>
      <el-alert
        type="info"
        :closable="false"
        show-icon
        :title="$t('settings.roleChatIdsTip')"
        style="margin-bottom: 16px;"
      />
      <el-form
        v-loading="loadingRoles"
        :model="rolesForm"
        label-width="160px"
        label-position="left"
      >
        <el-form-item :label="$t('settings.pickerChatId')">
          <el-input
            v-model="rolesForm.picker_chat_id"
            :placeholder="$t('settings.chatIdPlaceholder')"
            clearable
            style="max-width: 300px;"
          />
          <el-button
            link
            type="primary"
            style="margin-left: 8px;"
            @click="openChatPicker('picker')"
          >
            {{ $t('settings.detectChat') }}
          </el-button>
        </el-form-item>
        <el-form-item :label="$t('settings.deliveryChatId')">
          <el-input
            v-model="rolesForm.delivery_chat_id"
            :placeholder="$t('settings.chatIdPlaceholder')"
            clearable
            style="max-width: 300px;"
          />
          <el-button
            link
            type="primary"
            style="margin-left: 8px;"
            @click="openChatPicker('delivery')"
          >
            {{ $t('settings.detectChat') }}
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="savingRoles"
            @click="saveRoles"
          >
            {{ $t('common.save') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="setting-card" shadow="never">
      <template #header>
        <span>{{ $t('settings.deliveryFee') }}</span>
      </template>
      <el-form
        v-loading="loadingFee"
        :model="feeForm"
        label-width="160px"
        label-position="left"
      >
        <el-form-item :label="$t('settings.freeDistanceKm')">
          <el-input-number
            v-model="feeForm.free_distance_km"
            :min="0"
            :step="0.5"
            :precision="2"
          />
          <span class="unit">km</span>
        </el-form-item>
        <el-form-item :label="$t('settings.feePerExtraKm')">
          <el-input-number
            v-model="feeForm.fee_per_extra_km_usd"
            :min="0"
            :step="0.1"
            :precision="2"
          />
          <span class="unit">USD / km</span>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="savingFee"
            @click="saveFee"
          >
            {{ $t('common.save') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 联系方式设置 -->
    <el-card class="setting-card" shadow="never">
      <template #header>
        <span>{{ $t('settings.contactInfo') }}</span>
      </template>
      <el-form
        v-loading="loadingContact"
        :model="contactForm"
        label-width="160px"
        label-position="left"
      >
        <el-form-item :label="$t('settings.contactPhone')">
          <el-input
            v-model="contactForm.phone"
            placeholder="+855 12 345 678"
            clearable
            style="max-width: 300px;"
          />
        </el-form-item>
        <el-form-item :label="$t('settings.contactWhatsapp')">
          <el-input
            v-model="contactForm.whatsapp"
            placeholder="+855 12 345 678"
            clearable
            style="max-width: 300px;"
          />
        </el-form-item>
        <el-form-item :label="$t('settings.contactTelegram')">
          <el-input
            v-model="contactForm.telegram"
            placeholder="@username"
            clearable
            style="max-width: 300px;"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            :loading="savingContact"
            @click="saveContact"
          >
            {{ $t('common.save') }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 选择 Telegram 群/用户对话框 -->
    <el-dialog
      v-model="chatPickerVisible"
      :title="$t('settings.detectChatTitle')"
      width="600px"
    >
      <el-alert
        type="warning"
        :closable="false"
        show-icon
        :title="$t('settings.detectChatHelp')"
        style="margin-bottom: 12px;"
      />
      <div style="margin-bottom: 12px;">
        <el-button
          :loading="loadingChats"
          type="primary"
          plain
          @click="loadChats"
        >
          {{ $t('settings.refreshList') }}
        </el-button>
      </div>
      <div v-loading="loadingChats" class="chat-list">
        <div
          v-for="c in chats"
          :key="c.id"
          class="chat-item"
          @click="selectChat(c)"
        >
          <div class="chat-main">
            <div class="chat-title">
              <el-tag size="small" :type="chatTagType(c.type)">{{ c.type }}</el-tag>
              <span class="chat-name">{{ c.title }}</span>
              <span v-if="c.username" class="chat-username">@{{ c.username }}</span>
            </div>
            <div class="chat-id">{{ c.id }}</div>
            <div v-if="c.last_message" class="chat-msg">{{ c.last_message }}</div>
          </div>
          <el-button size="small" type="primary">
            {{ $t('settings.useThis') }}
          </el-button>
        </div>
        <div v-if="!loadingChats && chats.length === 0" class="empty">
          {{ $t('settings.noChats') }}
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'
import {
  getRoleChatIds,
  updateRoleChatIds,
  getDeliveryFeeSettings,
  updateDeliveryFeeSettings,
  getTelegramRecentChats,
  getContactInfo,
  updateContactInfo,
} from '@/api'

const { t } = useI18n()

const loadingRoles = ref(false)
const savingRoles = ref(false)
const rolesForm = ref({ picker_chat_id: '', delivery_chat_id: '' })

const loadingFee = ref(false)
const savingFee = ref(false)
const feeForm = ref({ free_distance_km: 3, fee_per_extra_km_usd: 0.5 })

// Telegram chat 检测
const chatPickerVisible = ref(false)
const loadingChats = ref(false)
const chats = ref([])
const pickerTarget = ref('picker') // 'picker' | 'delivery'

function chatTagType(t) {
  if (t === 'private') return 'info'
  if (t === 'group' || t === 'supergroup') return 'success'
  if (t === 'channel') return 'warning'
  return ''
}

function openChatPicker(target) {
  pickerTarget.value = target
  chatPickerVisible.value = true
  if (chats.value.length === 0) loadChats()
}

async function loadChats() {
  loadingChats.value = true
  try {
    const data = await getTelegramRecentChats()
    chats.value = data.chats || []
  } catch (e) {
    ElMessage.error(e?.message || 'Load failed')
  } finally {
    loadingChats.value = false
  }
}

function selectChat(c) {
  if (pickerTarget.value === 'picker') {
    rolesForm.value.picker_chat_id = c.id
  } else {
    rolesForm.value.delivery_chat_id = c.id
  }
  chatPickerVisible.value = false
  ElMessage.success(t('settings.chatSelected') || 'Selected')
}

async function loadRoles() {
  loadingRoles.value = true
  try {
    const data = await getRoleChatIds()
    rolesForm.value = {
      picker_chat_id: data.picker_chat_id || '',
      delivery_chat_id: data.delivery_chat_id || '',
    }
  } catch (e) {
    ElMessage.error(e?.message || 'Load failed')
  } finally {
    loadingRoles.value = false
  }
}

async function saveRoles() {
  savingRoles.value = true
  try {
    const data = await updateRoleChatIds({
      picker_chat_id: rolesForm.value.picker_chat_id || '',
      delivery_chat_id: rolesForm.value.delivery_chat_id || '',
    })
    rolesForm.value = {
      picker_chat_id: data.picker_chat_id || '',
      delivery_chat_id: data.delivery_chat_id || '',
    }
    ElMessage.success(t('common.saveSuccess') || 'Saved')
  } catch (e) {
    ElMessage.error(e?.message || 'Save failed')
  } finally {
    savingRoles.value = false
  }
}

async function loadFee() {
  loadingFee.value = true
  try {
    const data = await getDeliveryFeeSettings()
    feeForm.value = {
      free_distance_km: data.free_distance_km ?? 3,
      fee_per_extra_km_usd: data.fee_per_extra_km_usd ?? 0.5,
    }
  } catch (e) {
    ElMessage.error(e?.message || 'Load failed')
  } finally {
    loadingFee.value = false
  }
}

async function saveFee() {
  savingFee.value = true
  try {
    const data = await updateDeliveryFeeSettings({
      free_distance_km: feeForm.value.free_distance_km,
      fee_per_extra_km_usd: feeForm.value.fee_per_extra_km_usd,
    })
    feeForm.value = {
      free_distance_km: data.free_distance_km,
      fee_per_extra_km_usd: data.fee_per_extra_km_usd,
    }
    ElMessage.success(t('common.saveSuccess') || 'Saved')
  } catch (e) {
    ElMessage.error(e?.message || 'Save failed')
  } finally {
    savingFee.value = false
  }
}

onMounted(() => {
  loadRoles()
  loadFee()
  loadContact()
})

const loadingContact = ref(false)
const savingContact = ref(false)
const contactForm = ref({ phone: '', telegram: '', whatsapp: '' })

async function loadContact() {
  loadingContact.value = true
  try {
    const data = await getContactInfo()
    contactForm.value = { phone: data.phone || '', telegram: data.telegram || '', whatsapp: data.whatsapp || '' }
  } catch (e) {
    ElMessage.error(e?.message || 'Load failed')
  } finally {
    loadingContact.value = false
  }
}

async function saveContact() {
  savingContact.value = true
  try {
    const data = await updateContactInfo(contactForm.value)
    contactForm.value = { phone: data.phone || '', telegram: data.telegram || '', whatsapp: data.whatsapp || '' }
    ElMessage.success(t('settings.contactInfoSaved'))
  } catch (e) {
    ElMessage.error(e?.message || 'Save failed')
  } finally {
    savingContact.value = false
  }
}
</script>

<style scoped lang="scss">
.settings-page {
  padding: 16px;
}
.page-header {
  margin-bottom: 16px;
  h2 {
    margin: 0;
    color: #2b2b2b;
  }
}
.setting-card {
  margin-bottom: 16px;
  border-radius: 8px;
}
.unit {
  margin-left: 8px;
  color: #666;
}
.chat-list {
  max-height: 400px;
  overflow-y: auto;
}
.chat-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid #eee;
  border-radius: 6px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.2s;
  &:hover {
    border-color: #2b2b2b;
    background: #fafafa;
  }
}
.chat-main {
  flex: 1;
  min-width: 0;
}
.chat-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 600;
  color: #2b2b2b;
}
.chat-name {
  margin-left: 4px;
}
.chat-username {
  color: #888;
  font-weight: normal;
  font-size: 12px;
}
.chat-id {
  font-family: monospace;
  font-size: 12px;
  color: #666;
  margin-top: 2px;
}
.chat-msg {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.empty {
  text-align: center;
  color: #999;
  padding: 24px 0;
}
</style>
