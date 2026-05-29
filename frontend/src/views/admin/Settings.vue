<template>
  <div class="page-container">
    <div class="page-header">
      <h2>{{ $t('settings.title') }}</h2>
    </div>

    <!-- Telegram 设置 -->
    <div class="setting-section">
      <div class="section-title">{{ $t('settings.roleChatIds') }}</div>
      <van-notice-bar :text="$t('settings.roleChatIdsTip')" left-icon="info-o" color="#1989fa" background="#ecf9ff" wrapable :scrollable="false" style="margin-bottom:12px;border-radius:6px;" />
      <van-loading v-if="loadingRoles" size="20" style="padding: 12px 0;" />
      <van-cell-group v-else inset>
        <van-field v-model="rolesForm.group_chat_id" :label="$t('settings.groupChatId')" :placeholder="$t('settings.chatIdPlaceholder')" clearable>
          <template #button>
            <van-button size="small" type="primary" plain @click="openChatPicker">{{ $t('settings.detectChat') }}</van-button>
          </template>
        </van-field>
        <van-field v-model="rolesForm.delivery_group_link" :label="$t('settings.deliveryGroupLink')" placeholder="https://t.me/+..." clearable />
      </van-cell-group>
      <div style="padding: 12px 16px 0;">
        <van-button type="primary" size="small" :loading="savingRoles" @click="saveRoles">{{ $t('common.save') }}</van-button>
      </div>
    </div>

    <!-- 配送费设置 -->
    <div class="setting-section">
      <div class="section-title">{{ $t('settings.deliveryFee') }}</div>
      <van-loading v-if="loadingFee" size="20" style="padding: 12px 0;" />
      <van-cell-group v-else inset>
        <van-field v-model.number="feeForm.free_distance_km" type="number" :label="$t('settings.freeDistanceKm')" placeholder="3">
          <template #extra><span style="color:#999;font-size:12px">km</span></template>
        </van-field>
        <van-field v-model.number="feeForm.fee_per_extra_km_usd" type="number" :label="$t('settings.feePerExtraKm')" placeholder="0.5">
          <template #extra><span style="color:#999;font-size:12px">USD/km</span></template>
        </van-field>
      </van-cell-group>
      <div style="padding: 12px 16px 0;">
        <van-button type="primary" size="small" :loading="savingFee" @click="saveFee">{{ $t('common.save') }}</van-button>
      </div>
    </div>

    <!-- Google Maps 配置 -->
    <div class="setting-section">
      <div class="section-title">{{ $t('settings.googleMaps') }}</div>
      <van-notice-bar :text="$t('settings.googleMapsTip')" left-icon="info-o" color="#1989fa" background="#ecf9ff" wrapable :scrollable="false" style="margin-bottom:12px;border-radius:6px;" />
      <van-loading v-if="loadingMaps" size="20" style="padding: 12px 0;" />
      <van-cell-group v-else inset>
        <van-field v-model="mapsForm.api_key" :label="$t('settings.googleMapsKey')" :placeholder="$t('settings.googleMapsKeyPlaceholder')" clearable />
        <van-cell v-if="mapsForm.warehouse_lat && mapsForm.warehouse_lng" :title="$t('settings.warehouseLocation')"
          :value="`${mapsForm.warehouse_lat}, ${mapsForm.warehouse_lng}`">
          <template #right-icon>
            <a :href="`https://maps.google.com/?q=${mapsForm.warehouse_lat},${mapsForm.warehouse_lng}`" target="_blank" style="color:#1989fa;font-size:12px;white-space:nowrap;">查看</a>
          </template>
        </van-cell>
        <div v-if="mapsForm.api_key" ref="mapPickerEl" style="width:100%;height:240px;border-radius:6px;overflow:hidden;margin:8px 0;" />
      </van-cell-group>
      <div style="padding: 12px 16px 0;">
        <van-button type="primary" size="small" :loading="savingMaps" @click="saveMaps">{{ $t('common.save') }}</van-button>
      </div>
    </div>

    <!-- 联系方式设置 -->
    <div class="setting-section">
      <div class="section-title">{{ $t('settings.contactInfo') }}</div>
      <van-loading v-if="loadingContact" size="20" style="padding: 12px 0;" />
      <van-cell-group v-else inset>
        <van-field v-model="contactForm.phone" :label="$t('settings.contactPhone')" placeholder="+855 12 345 678" clearable />
        <van-field v-model="contactForm.whatsapp" :label="$t('settings.contactWhatsapp')" placeholder="+855 12 345 678" clearable />
        <van-field v-model="contactForm.telegram" :label="$t('settings.contactTelegram')" placeholder="@username" clearable />
        <van-field v-model="contactForm.wechat" :label="$t('settings.contactWechat')" placeholder="WeChat ID" clearable />
      </van-cell-group>
      <div style="padding: 12px 16px 0;">
        <van-button type="primary" size="small" :loading="savingContact" @click="saveContact">{{ $t('common.save') }}</van-button>
      </div>
    </div>

    <!-- Telegram chat 选择弹窗 -->
    <van-popup v-model:show="chatPickerVisible" position="bottom" round :style="{ height: '70vh' }">
      <van-nav-bar :title="$t('settings.detectChatTitle')" :left-text="$t('common.cancel')" @click-left="chatPickerVisible = false" />
      <van-notice-bar :text="$t('settings.detectChatHelp')" left-icon="warning-o" color="#ff976a" background="#fff7cc" wrapable :scrollable="false" style="margin:8px 16px;border-radius:6px;" />
      <div style="padding: 8px 16px;">
        <van-button size="small" type="primary" plain :loading="loadingChats" @click="loadChats">{{ $t('settings.refreshList') }}</van-button>
      </div>
      <van-loading v-if="loadingChats" size="24" vertical style="padding: 20px 0; text-align:center" />
      <van-cell-group v-else inset>
        <van-cell v-for="c in chats" :key="c.id" :title="c.title" :label="`${c.type} · ${c.id}`" is-link @click="selectChat(c)">
          <template #icon>
            <van-tag :type="chatTagType(c.type)" style="margin-right:8px;align-self:center">{{ c.type }}</van-tag>
          </template>
        </van-cell>
        <van-empty v-if="chats.length === 0" :description="$t('settings.noChats')" />
      </van-cell-group>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, onMounted, watch, nextTick } from 'vue'
import { showSuccessToast, showFailToast } from 'vant'
import { useI18n } from 'vue-i18n'
import {
  getRoleChatIds, updateRoleChatIds,
  getDeliveryFeeSettings, updateDeliveryFeeSettings,
  getTelegramRecentChats,
  getContactInfo, updateContactInfo,
  getGoogleMapsSettings, updateGoogleMapsSettings,
} from '@/api'

const { t } = useI18n()

const loadingRoles = ref(false)
const savingRoles = ref(false)
const rolesForm = ref({ group_chat_id: '', delivery_group_link: '' })

const loadingFee = ref(false)
const savingFee = ref(false)
const feeForm = ref({ free_distance_km: 3, fee_per_extra_km_usd: 0.5 })

const chatPickerVisible = ref(false)
const loadingChats = ref(false)
const chats = ref([])

const chatTagType = (type) => {
  if (type === 'private') return 'primary'
  if (type === 'group' || type === 'supergroup') return 'success'
  if (type === 'channel') return 'warning'
  return 'default'
}

const openChatPicker = () => {
  chatPickerVisible.value = true
  if (chats.value.length === 0) loadChats()
}

const loadChats = async () => {
  loadingChats.value = true
  try {
    const data = await getTelegramRecentChats()
    chats.value = data.chats || []
  } catch (e) {
    showFailToast(e?.message || 'Load failed')
  } finally {
    loadingChats.value = false
  }
}

const selectChat = (c) => {
  rolesForm.value.group_chat_id = c.id
  chatPickerVisible.value = false
  showSuccessToast(t('settings.chatSelected') || 'Selected')
}

const loadRoles = async () => {
  loadingRoles.value = true
  try {
    const data = await getRoleChatIds()
    rolesForm.value = { group_chat_id: data.group_chat_id || '', delivery_group_link: data.delivery_group_link || '' }
  } catch (e) {
    showFailToast(e?.message || 'Load failed')
  } finally {
    loadingRoles.value = false
  }
}

const saveRoles = async () => {
  savingRoles.value = true
  try {
    const data = await updateRoleChatIds({ group_chat_id: rolesForm.value.group_chat_id || '', delivery_group_link: rolesForm.value.delivery_group_link || '' })
    rolesForm.value = { group_chat_id: data.group_chat_id || '', delivery_group_link: data.delivery_group_link || '' }
    showSuccessToast(t('common.saveSuccess') || 'Saved')
  } catch (e) {
    showFailToast(e?.message || 'Save failed')
  } finally {
    savingRoles.value = false
  }
}

const loadFee = async () => {
  loadingFee.value = true
  try {
    const data = await getDeliveryFeeSettings()
    feeForm.value = { free_distance_km: data.free_distance_km ?? 3, fee_per_extra_km_usd: data.fee_per_extra_km_usd ?? 0.5 }
  } catch (e) {
    showFailToast(e?.message || 'Load failed')
  } finally {
    loadingFee.value = false
  }
}

const saveFee = async () => {
  savingFee.value = true
  try {
    const data = await updateDeliveryFeeSettings({ free_distance_km: feeForm.value.free_distance_km, fee_per_extra_km_usd: feeForm.value.fee_per_extra_km_usd })
    feeForm.value = { free_distance_km: data.free_distance_km, fee_per_extra_km_usd: data.fee_per_extra_km_usd }
    showSuccessToast(t('common.saveSuccess') || 'Saved')
  } catch (e) {
    showFailToast(e?.message || 'Save failed')
  } finally {
    savingFee.value = false
  }
}

const loadingMaps = ref(false)
const savingMaps = ref(false)
const mapsForm = ref({ api_key: '', warehouse_lat: '', warehouse_lng: '' })
const mapPickerEl = ref(null)
let _googleMap = null
let _warehouseMarker = null

const initMapPicker = async () => {
  await nextTick()
  if (!mapPickerEl.value || !mapsForm.value.api_key) return
  if (!window.google?.maps) {
    await new Promise((resolve, reject) => {
      if (document.querySelector('#gmap-script')) { resolve(); return }
      const s = document.createElement('script')
      s.id = 'gmap-script'
      s.src = `https://maps.googleapis.com/maps/api/js?key=${mapsForm.value.api_key}`
      s.onload = resolve; s.onerror = reject
      document.head.appendChild(s)
    })
  }
  const lat = parseFloat(mapsForm.value.warehouse_lat) || 11.5564
  const lng = parseFloat(mapsForm.value.warehouse_lng) || 104.9282
  _googleMap = new window.google.maps.Map(mapPickerEl.value, { center: { lat, lng }, zoom: mapsForm.value.warehouse_lat ? 16 : 13, mapTypeControl: false, streetViewControl: false, fullscreenControl: false })
  if (mapsForm.value.warehouse_lat && mapsForm.value.warehouse_lng) {
    _warehouseMarker = new window.google.maps.Marker({ position: { lat, lng }, map: _googleMap })
  }
  _googleMap.addListener('click', (e) => {
    const newLat = e.latLng.lat().toFixed(6)
    const newLng = e.latLng.lng().toFixed(6)
    mapsForm.value.warehouse_lat = newLat
    mapsForm.value.warehouse_lng = newLng
    const pos = { lat: parseFloat(newLat), lng: parseFloat(newLng) }
    if (_warehouseMarker) { _warehouseMarker.setPosition(pos) } else { _warehouseMarker = new window.google.maps.Marker({ position: pos, map: _googleMap }) }
  })
}

watch(() => mapsForm.value.api_key, (val) => { if (val) initMapPicker() })

const loadMaps = async () => {
  loadingMaps.value = true
  try {
    const data = await getGoogleMapsSettings()
    mapsForm.value = { api_key: data.api_key || '', warehouse_lat: data.warehouse_lat || '', warehouse_lng: data.warehouse_lng || '' }
    if (mapsForm.value.api_key) initMapPicker()
  } catch (e) {
    showFailToast(e?.message || 'Load failed')
  } finally {
    loadingMaps.value = false
  }
}

const saveMaps = async () => {
  savingMaps.value = true
  try {
    await updateGoogleMapsSettings(mapsForm.value)
    showSuccessToast(t('common.saveSuccess') || 'Saved')
  } catch (e) {
    showFailToast(e?.message || 'Save failed')
  } finally {
    savingMaps.value = false
  }
}

const loadingContact = ref(false)
const savingContact = ref(false)
const contactForm = ref({ phone: '', telegram: '', whatsapp: '', wechat: '' })

const loadContact = async () => {
  loadingContact.value = true
  try {
    const data = await getContactInfo()
    contactForm.value = { phone: data.phone || '', telegram: data.telegram || '', whatsapp: data.whatsapp || '', wechat: data.wechat || '' }
  } catch (e) {
    showFailToast(e?.message || 'Load failed')
  } finally {
    loadingContact.value = false
  }
}

const saveContact = async () => {
  savingContact.value = true
  try {
    const data = await updateContactInfo(contactForm.value)
    contactForm.value = { phone: data.phone || '', telegram: data.telegram || '', whatsapp: data.whatsapp || '', wechat: data.wechat || '' }
    showSuccessToast(t('settings.contactInfoSaved'))
  } catch (e) {
    showFailToast(e?.message || 'Save failed')
  } finally {
    savingContact.value = false
  }
}

onMounted(() => { loadRoles(); loadFee(); loadContact(); loadMaps() })
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
</style>
