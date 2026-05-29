<template>
  <div class="page-container">
    <div class="page-header">
      <h2>{{ $t('admin.merchants') }}</h2>
      <div class="header-btns">
        <van-button type="primary" size="small" icon="plus" @click="handleAdd('merchant')">{{ $t('admin.addMerchant') }}</van-button>
        <van-button v-if="isSuperAdmin" type="warning" size="small" plain icon="plus" @click="handleAdd('admin')">{{ $t('admin.addAdmin') }}</van-button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stat-grid">
      <div class="stat-card">
        <div class="stat-icon" style="background:#2563EB"><van-icon name="manager-o" size="22" color="#fff" /></div>
        <div class="stat-content">
          <div class="stat-value">{{ merchants.length }}</div>
          <div class="stat-label">{{ $t('admin.totalUsers') }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon" style="background:#DC2626"><van-icon name="manager" size="22" color="#fff" /></div>
        <div class="stat-content">
          <div class="stat-value">{{ adminCount }}</div>
          <div class="stat-label">{{ $t('admin.adminUsers') }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon" style="background:#16A34A"><van-icon name="shop-o" size="22" color="#fff" /></div>
        <div class="stat-content">
          <div class="stat-value">{{ merchantCount }}</div>
          <div class="stat-label">{{ $t('admin.merchantUsers') }}</div>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon" style="background:#0891B2"><van-icon name="passed" size="22" color="#fff" /></div>
        <div class="stat-content">
          <div class="stat-value">{{ activeCount }}</div>
          <div class="stat-label">{{ $t('admin.activeUsers') }}</div>
        </div>
      </div>
    </div>

    <van-loading v-if="loading" size="24" vertical style="padding: 40px 0; text-align:center" />

    <div v-else class="card-list">
      <div v-for="row in filteredUsers" :key="row.id" class="user-card" @click="handleEdit(row)">
        <div class="ucard-top">
          <div class="ucard-meta">
            <div class="ucard-name">{{ row.full_name || row.username }}</div>
            <div class="ucard-sub">
              <span v-if="row.phone" class="ucard-phone">{{ row.phone }}</span>
              <span v-else class="ucard-phone">{{ row.username }}</span>
              <span v-if="row.telegram_id" class="ucard-tg">· TG: {{ row.telegram_id }}</span>
            </div>
          </div>
          <van-switch
            v-if="canToggleActive(row)"
            :model-value="row.is_active"
            size="20"
            @change="(val) => handleToggleActive(row, val)"
            @click.stop
          />
        </div>
        <div class="ucard-tags">
          <van-tag :type="getRoleType(row.role)" size="medium">{{ getRoleText(row.role) }}</van-tag>
          <van-tag v-if="row.is_super_admin" type="warning" size="medium">{{ $t('admin.superAdmin') }}</van-tag>
          <van-tag v-if="canToggleActive(row) && !row.is_active" type="primary" size="medium">{{ $t('common.disabled') }}</van-tag>
        </div>
        <div v-if="row.address" class="ucard-address">{{ row.address }}</div>
        <div class="ucard-footer" @click.stop>
          <van-button v-if="canEditRow(row)" type="primary" size="small" plain @click="handleEdit(row)">{{ $t('common.edit') }}</van-button>
          <van-button
            v-if="canManageSuperAdminRow(row)"
            :type="row.is_super_admin ? 'warning' : 'success'"
            size="small"
            plain
            @click="handleToggleSuperAdmin(row, !row.is_super_admin)"
          >{{ row.is_super_admin ? $t('admin.demoteSuperAdmin') : $t('admin.promoteSuperAdmin') }}</van-button>
          <van-button v-if="canDeleteRow(row)" type="danger" size="small" plain @click="handleDeleteUser(row)">{{ $t('common.delete') }}</van-button>
        </div>
      </div>
      <van-empty v-if="filteredUsers.length === 0" :description="$t('common.noData')" />
    </div>

    <!-- 添加/编辑用户弹窗 -->
    <van-popup v-model:show="dialogVisible" position="bottom" round :style="{ height: '92vh', overflowY: 'auto' }">
      <van-nav-bar
        :title="dialogTitle"
        :left-text="$t('common.cancel')"
        :right-text="$t('common.confirm')"
        @click-left="dialogVisible = false"
        @click-right="handleSubmit"
      />
      <div style="padding: 0 0 20px;">
        <van-notice-bar
          v-if="!isEdit && isSuperAdmin"
          :text="$t('admin.dialogRoleTip')"
          left-icon="info-o"
          color="#1989fa"
          background="#ecf9ff"
          wrapable
          :scrollable="false"
          style="margin: 8px 16px; border-radius: 6px;"
        />

        <van-cell-group inset style="margin-top: 8px;">
          <van-field v-if="isEdit" :label="$t('profile.phone') + '/' + $t('login.account')" :model-value="form.username" readonly />
          <van-field v-model="form.full_name" :label="$t('profile.name')" :placeholder="$t('profile.name')" required />

          <van-field :label="$t('product.status')" readonly>
            <template #input>
              <select v-model="form.role" :disabled="isEdit" style="border:none;outline:none;width:100%;font-size:14px;background:transparent;">
                <option value="merchant">{{ $t('role.merchant') }}</option>
                <option v-if="isSuperAdmin" value="admin">{{ $t('role.admin') }}</option>
              </select>
            </template>
          </van-field>

          <van-field v-if="!isEdit" v-model="form.phone" :label="$t('profile.phone')" :placeholder="form.role === 'merchant' ? $t('admin.phonePlaceholder') : ''" />
          <van-field v-model="form.address" type="textarea" rows="2" :label="$t('profile.address')" />

          <van-field :label="$t('profile.locationUrl')">
            <template #input>
              <input v-model="form.location_url" :placeholder="$t('profile.locationUrlPlaceholder')" style="flex:1;border:none;outline:none;font-size:14px;width:100%;" />
            </template>
            <template #button>
              <van-button size="small" plain @click="openAdminMapPicker">📍 {{ $t('profile.pickLocation') }}</van-button>
            </template>
          </van-field>
          <van-cell v-if="form.location_url">
            <template #title>
              <a :href="form.location_url" target="_blank" style="color:#1989fa;font-size:13px;">{{ $t('profile.viewMap') }}</a>
            </template>
          </van-cell>
        </van-cell-group>

        <template v-if="form.role === 'merchant'">
          <van-cell-group inset style="margin-top: 8px;">
            <van-cell :title="$t('admin.allowMonthlyBilling')">
              <template #right-icon>
                <van-switch v-model="form.allow_credit" size="20" />
              </template>
            </van-cell>
            <van-field v-model.number="form.credit_limit" type="number" :label="$t('admin.creditLimit')" readonly />
            <van-field v-model.number="form.billing_cycle_days" type="number" :label="$t('admin.billingDay')" :placeholder="$t('admin.billingDayPlaceholder')" />
          </van-cell-group>
        </template>

        <van-cell-group v-if="isEdit && !form.is_super_admin" inset style="margin-top: 8px;">
          <van-cell :title="$t('product.status')">
            <template #right-icon>
              <van-switch v-model="form.is_active" size="20" />
            </template>
          </van-cell>
        </van-cell-group>

        <template v-if="form.role === 'admin' && isEdit">
          <div class="form-section-title">{{ $t('admin.telegramSettings') }}</div>
          <van-cell-group inset>
            <van-field v-model="form.telegram_id" :label="$t('admin.telegramId')" :placeholder="$t('admin.telegramIdPlaceholder')" clearable />
          </van-cell-group>
          <van-notice-bar :text="$t('admin.telegramTip')" left-icon="info-o" color="#1989fa" background="#ecf9ff" wrapable :scrollable="false" style="margin: 8px 16px; border-radius: 6px;" />
        </template>

        <div style="padding: 16px;">
          <van-button block type="primary" :loading="submitting" @click="handleSubmit">{{ $t('common.confirm') }}</van-button>
        </div>
      </div>
    </van-popup>

    <!-- 新用户账号展示弹窗 -->
    <van-popup v-model:show="newAccountVisible" position="bottom" round :style="{ minHeight: '40vh' }" :close-on-click-overlay="false">
      <van-nav-bar :title="$t('admin.userAdded')" :right-text="$t('common.confirm')" @click-right="newAccountVisible = false" />
      <div class="new-account-info">
        <p class="account-tip">{{ $t('admin.userCreatedTip') }}</p>
        <div class="account-box">
          <span class="account-label">{{ $t('login.account') }}</span>
          <span class="account-number">{{ newAccountNumber }}</span>
        </div>
        <p class="account-pwd-tip">{{ $t('admin.temporaryPassword') }}: <b>{{ newAccountPassword }}</b></p>
      </div>
    </van-popup>

    <!-- 地图选点弹窗 -->
    <van-popup v-model:show="showAdminMapPicker" position="bottom" round :style="{ height: '85vh' }" :close-on-click-overlay="false">
      <van-nav-bar
        :title="$t('profile.pickLocation')"
        :left-text="$t('common.cancel')"
        :right-text="$t('common.confirm')"
        @click-left="showAdminMapPicker = false"
        @click-right="confirmAdminMapPick"
      />
      <div style="padding: 12px 16px;">
        <div style="display:flex;justify-content:flex-end;margin-bottom:8px">
          <van-button size="small" type="primary" plain :loading="adminLocating" @click="adminLocateMe">
            📍 {{ adminLocating ? $t('profile.locating') : $t('profile.locateMe') }}
          </van-button>
        </div>
        <div ref="adminMapContainerRef" style="width:100%;height:300px;background:#eee;border-radius:6px;margin-bottom:10px;"></div>
        <van-field v-model="adminPickedUrl" :placeholder="$t('profile.locationUrlPlaceholder')" clearable />
        <div style="font-size:12px;color:#999;margin-top:6px">{{ $t('profile.locationPickerHint') }}</div>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showFailToast, showConfirmDialog } from 'vant'
import { getUserList, register, updateUser, deleteUser, setUserSuperAdmin } from '@/api'
import { getRoleText } from '@/utils/format'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const isSuperAdmin = computed(() => userStore.isSuperAdmin)

// ──────────────── 地图选点 ────────────────
const GMAP_KEY = 'AIzaSyBLMXrpizbEE4f36sUCerOasHVWM8Doumc'
const showAdminMapPicker = ref(false)
const adminPickedUrl = ref('')
const adminLocating = ref(false)
const adminMapContainerRef = ref(null)
let adminMapInstance = null
let adminMapMarker = null

const loadGMapScript = () => {
  if (document.getElementById('google-maps-script')) {
    return window.google?.maps ? Promise.resolve() : new Promise(r => { const t = setInterval(() => { if (window.google?.maps) { clearInterval(t); r() } }, 100) })
  }
  return new Promise((resolve, reject) => {
    const s = document.createElement('script')
    s.id = 'google-maps-script'
    s.src = `https://maps.googleapis.com/maps/api/js?key=${GMAP_KEY}&language=zh-CN`
    s.async = true; s.defer = true; s.onload = resolve; s.onerror = reject
    document.head.appendChild(s)
  })
}

const placeAdminMarker = (latlng) => {
  if (!adminMapInstance) return
  if (adminMapMarker) adminMapMarker.setMap(null)
  adminMapMarker = new window.google.maps.Marker({ position: latlng, map: adminMapInstance })
  adminMapInstance.panTo(latlng)
  const lat = (typeof latlng.lat === 'function' ? latlng.lat() : latlng.lat).toFixed(6)
  const lng = (typeof latlng.lng === 'function' ? latlng.lng() : latlng.lng).toFixed(6)
  adminPickedUrl.value = `https://maps.google.com/?q=${lat},${lng}`
}

const adminLocateMe = () => {
  if (!navigator.geolocation) { showFailToast('浏览器不支持定位'); return }
  adminLocating.value = true
  navigator.geolocation.getCurrentPosition(
    (pos) => {
      adminLocating.value = false
      const ll = { lat: pos.coords.latitude, lng: pos.coords.longitude }
      if (adminMapInstance) { adminMapInstance.setCenter(ll); adminMapInstance.setZoom(17); placeAdminMarker(ll) }
    },
    () => { adminLocating.value = false; showFailToast('无法定位') },
    { enableHighAccuracy: true, timeout: 10000 }
  )
}

const openAdminMapPicker = async () => {
  adminPickedUrl.value = form.location_url || ''
  showAdminMapPicker.value = true
  await nextTick()
  try {
    await loadGMapScript()
    await nextTick()
    const container = adminMapContainerRef.value
    if (!container || !window.google?.maps) return
    const defaultCenter = { lat: 11.5564, lng: 104.9282 }
    let initCenter = defaultCenter
    const existing = adminPickedUrl.value
    if (existing) {
      const m = existing.match(/[?&]q=([\-\d.]+),([\-\d.]+)/) || existing.match(/@([\-\d.]+),([\-\d.]+)/)
      if (m) initCenter = { lat: parseFloat(m[1]), lng: parseFloat(m[2]) }
    }
    adminMapInstance = new window.google.maps.Map(container, {
      center: initCenter, zoom: 15,
      mapTypeControl: false, streetViewControl: false, fullscreenControl: false,
      gestureHandling: 'greedy',
    })
    if (existing && initCenter !== defaultCenter) {
      adminMapMarker = new window.google.maps.Marker({ position: initCenter, map: adminMapInstance })
    } else if (!existing && navigator.geolocation) {
      adminLocating.value = true
      navigator.geolocation.getCurrentPosition(
        (pos) => { adminLocating.value = false; adminMapInstance?.setCenter({ lat: pos.coords.latitude, lng: pos.coords.longitude }); adminMapInstance?.setZoom(17) },
        () => { adminLocating.value = false }, { enableHighAccuracy: true, timeout: 8000 }
      )
    }
    adminMapInstance.addListener('click', (e) => placeAdminMarker(e.latLng))
  } catch { showFailToast('地图加载失败') }
}

const confirmAdminMapPick = () => {
  form.location_url = adminPickedUrl.value
  showAdminMapPicker.value = false
  adminMapInstance = null; adminMapMarker = null
}
// ───────────────────────────────────────────────

const { t } = useI18n()

const loading = ref(false)
const merchants = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)

const form = reactive({
  id: null, username: '', password: '', full_name: '', role: 'merchant',
  phone: '', address: '', location_url: '', credit_limit: 0,
  billing_cycle_days: null, allow_credit: false, is_active: true,
  telegram_id: '', is_super_admin: false,
})

const newAccountVisible = ref(false)
const newAccountNumber = ref('')
const newAccountPassword = ref('')

const getRoleType = (role) => {
  const map = { admin: 'danger', merchant: 'success' }
  return map[role] || 'primary'
}

const isProtectedUser = (row) => !!row?.is_super_admin
const canEditRow = (row) => row.role !== 'admin' || isSuperAdmin.value || row.id === userStore.userInfo?.id
const canToggleActive = (row) => !isProtectedUser(row) && (row.role !== 'admin' || isSuperAdmin.value)
const canDeleteRow = (row) => isSuperAdmin.value && !isProtectedUser(row)
const canManageSuperAdminRow = (row) => isSuperAdmin.value && row.role === 'admin' && row.id !== userStore.userInfo?.id

const adminCount = computed(() => merchants.value.filter(r => r.role === 'admin').length)
const merchantCount = computed(() => merchants.value.filter(r => r.role === 'merchant').length)
const activeCount = computed(() => merchants.value.filter(r => r.is_active).length)
const filteredUsers = computed(() => merchants.value)

const dialogTitle = computed(() => {
  const roleText = form.role === 'admin' ? t('role.admin') : t('role.merchant')
  return isEdit.value ? t('admin.editUserWithRole', { role: roleText }) : t('admin.addUserWithRole', { role: roleText })
})

const loadMerchants = async () => {
  loading.value = true
  try {
    merchants.value = await getUserList()
  } catch (error) {
    console.error('加载用户失败:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  Object.assign(form, {
    id: null, username: '', password: '', full_name: '', role: 'merchant',
    phone: '', address: '', location_url: '', credit_limit: 0,
    billing_cycle_days: null, allow_credit: false, is_active: true,
    telegram_id: '', is_super_admin: false,
  })
  newAccountPassword.value = ''
}

const handleAdd = (role = 'merchant') => {
  resetForm()
  form.role = role
  isEdit.value = false
  dialogVisible.value = true
}

const handleEdit = (row) => {
  Object.assign(form, {
    id: row.id, username: row.username, password: '',
    full_name: row.full_name || '', role: row.role,
    phone: row.phone || '', address: row.address || '',
    location_url: row.location_url || '', credit_limit: row.credit_limit || 0,
    billing_cycle_days: row.billing_cycle_days || null, allow_credit: row.allow_credit || false,
    is_active: row.is_active, telegram_id: row.telegram_id || '',
    is_super_admin: !!row.is_super_admin,
  })
  isEdit.value = true
  dialogVisible.value = true
}

const handleToggleActive = async (row, val) => {
  try {
    await updateUser(row.id, { is_active: val })
    row.is_active = val
    showSuccessToast(val ? t('common.enabled') : t('common.disabled'))
  } catch {
    showFailToast(t('common.operationFailed'))
  }
}

const handleSubmit = async () => {
  if (!form.full_name?.trim()) {
    showFailToast(t('admin.fullNameRequired')); return
  }
  if (form.role === 'merchant' && !isEdit.value && !form.phone?.trim()) {
    showFailToast(t('admin.phoneRequired')); return
  }
  submitting.value = true
  try {
    if (isEdit.value) {
      const payload = {
        full_name: form.full_name, phone: form.phone, address: form.address,
        location_url: form.location_url, credit_limit: form.credit_limit,
        billing_cycle_days: form.billing_cycle_days, allow_credit: form.allow_credit,
        is_active: form.is_active, telegram_id: form.telegram_id || null,
      }
      await updateUser(form.id, payload)
      showSuccessToast(t('admin.userUpdated'))
      if (form.id === userStore.userInfo?.id) await userStore.fetchUserInfo()
    } else {
      const res = await register({
        full_name: form.full_name, role: form.role, phone: form.phone,
        address: form.address, location_url: form.location_url,
        credit_limit: form.credit_limit, billing_cycle_days: form.billing_cycle_days,
        allow_credit: form.allow_credit,
      })
      newAccountNumber.value = res.user.username
      newAccountPassword.value = res.temporary_password
      newAccountVisible.value = true
    }
    dialogVisible.value = false
    loadMerchants()
  } catch (error) {
    console.error('提交失败:', error)
  } finally {
    submitting.value = false
  }
}

const handleToggleSuperAdmin = async (row, enabled) => {
  try {
    await showConfirmDialog({
      title: t('admin.hint'),
      message: enabled
        ? t('admin.promoteSuperAdminConfirm', { name: row.full_name || row.username })
        : t('admin.demoteSuperAdminConfirm', { name: row.full_name || row.username }),
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
    })
    await setUserSuperAdmin(row.id, enabled)
    showSuccessToast(t('admin.superAdminUpdated'))
    loadMerchants()
  } catch {}
}

const handleDeleteUser = async (row) => {
  try {
    await showConfirmDialog({
      title: t('admin.hint'),
      message: t('admin.deleteUserConfirm', { name: row.full_name || row.username }),
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
    })
    await deleteUser(row.id)
    showSuccessToast(t('admin.userDeleted'))
    loadMerchants()
  } catch {}
}

onMounted(() => { loadMerchants() })
</script>

<style scoped>
.header-btns { display: flex; gap: 8px; flex-wrap: wrap; }

.stat-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 16px;
}

.stat-card {
  background: #fff;
  border-radius: 10px;
  padding: 12px;
  display: flex;
  align-items: center;
  gap: 10px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
}

.stat-icon {
  width: 40px; height: 40px;
  border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}

.stat-value { font-size: 22px; font-weight: 700; color: #1a1a1a; }
.stat-label { font-size: 12px; color: #909399; margin-top: 2px; }

.card-list { display: flex; flex-direction: column; gap: 10px; }

.user-card {
  background: #fff;
  border-radius: 10px;
  padding: 12px 14px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
  cursor: pointer;
}
.user-card:active { opacity: 0.85; }

.ucard-top { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.ucard-meta { flex: 1; min-width: 0; }
.ucard-name { font-size: 15px; font-weight: 700; color: #111827; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.ucard-sub { font-size: 12px; color: #6b7280; margin-top: 2px; display: flex; align-items: center; gap: 4px; flex-wrap: wrap; }
.ucard-phone { font-size: 13px; color: #1d4ed8; font-weight: 600; }
.ucard-tg { font-size: 11px; color: #9ca3af; }
.ucard-tags { display: flex; flex-wrap: wrap; gap: 4px; margin-top: 8px; }
.ucard-address { font-size: 12px; color: #6b7280; margin-top: 6px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.ucard-footer {
  display: flex;
  gap: 8px;
  margin-top: 10px;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
  flex-wrap: wrap;
}

.form-section-title {
  font-size: 14px;
  font-weight: 600;
  color: #1a1a1a;
  padding: 12px 16px 4px;
}

.new-account-info { text-align: center; padding: 24px 16px; }
.account-tip { color: #67c23a; font-size: 14px; margin-bottom: 16px; }
.account-box {
  display: flex; align-items: center; justify-content: center; gap: 12px;
  padding: 16px; background: #f0f9ff; border-radius: 8px; margin-bottom: 12px;
}
.account-label { color: #909399; font-size: 14px; }
.account-number { font-size: 24px; font-weight: 700; color: #1D4ED8; letter-spacing: 2px; }
.account-pwd-tip { color: #909399; font-size: 13px; }
</style>
