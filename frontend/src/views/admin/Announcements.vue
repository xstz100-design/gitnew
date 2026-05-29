<template>
  <div class="page-container">
    <div class="page-header">
      <h2>{{ $t('announcement.title') }}</h2>
    </div>

    <van-tabs v-model:active="activeTab" @change="loadData" style="margin-bottom: 12px;">
      <van-tab :title="$t('announcement.notice')" name="notice" />
      <van-tab :title="$t('announcement.contact')" name="contact" />
    </van-tabs>

    <div style="margin-bottom: 12px;">
      <van-button type="primary" size="small" icon="plus" @click="handleAdd">{{ $t('common.add') }}</van-button>
    </div>

    <van-loading v-if="loading" size="24" vertical style="padding: 40px 0; text-align:center" />

    <div v-else class="card-list">
      <div v-for="row in items" :key="row.id" class="list-card" @click="handleEdit(row)">
        <div class="card-top">
          <div class="card-content-preview">{{ row.content_zh || row.content_en }}</div>
          <van-switch :model-value="row.is_active" size="20" @change="(val) => toggleActive(row, val)" @click.stop />
        </div>
        <div v-if="row.content_en && row.content_zh" class="card-content-en">{{ row.content_en }}</div>
        <div class="card-bottom">
          <span class="card-sort-badge">{{ $t('product.sortOrder') }}: {{ row.sort_order }}</span>
          <van-button type="danger" size="mini" plain @click.stop="handleDelete(row)">{{ $t('common.delete') }}</van-button>
        </div>
      </div>
      <van-empty v-if="items.length === 0" :description="$t('common.noData')" />
    </div>

    <!-- 编辑弹窗 -->
    <van-popup v-model:show="dialogVisible" position="bottom" round :style="{ minHeight: '65vh' }" destroy-on-close>
      <van-nav-bar
        :title="isEdit ? $t('announcement.edit') : $t('announcement.add')"
        :left-text="$t('common.cancel')"
        :right-text="$t('common.confirm')"
        @click-left="dialogVisible = false"
        @click-right="handleSubmit"
      />
      <van-form ref="formRef" style="padding-top: 8px;">
        <van-field
          v-model="form.content_zh"
          :label="$t('announcement.contentZh')"
          type="textarea"
          rows="3"
          :placeholder="$t('announcement.contentZhPlaceholder')"
          :rules="[{ required: true, message: t('announcement.contentRequired') }]"
        />
        <van-field
          v-model="form.content_en"
          :label="$t('announcement.contentEn')"
          type="textarea"
          rows="3"
          :placeholder="$t('announcement.contentEnPlaceholder')"
        />
        <van-field v-model.number="form.sort_order" type="number" :label="$t('product.sortOrder')" placeholder="0" />
        <van-cell :title="$t('product.status')">
          <template #right-icon>
            <van-switch v-model="form.is_active" size="20" />
          </template>
        </van-cell>
      </van-form>
      <div style="padding: 16px;">
        <van-button block type="primary" :loading="submitting" @click="handleSubmit">{{ $t('common.confirm') }}</van-button>
      </div>
    </van-popup>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { showSuccessToast, showConfirmDialog } from 'vant'
import { getAnnouncements, createAnnouncement, updateAnnouncement, deleteAnnouncement } from '@/api'

const { t } = useI18n()
const loading = ref(false)
const items = ref([])
const activeTab = ref('notice')
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()

const form = reactive({ id: null, content_zh: '', content_en: '', sort_order: 0, is_active: true })

const loadData = async () => {
  loading.value = true
  try {
    items.value = await getAnnouncements(activeTab.value)
  } catch (error) {
    console.error('加载公告失败:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => { Object.assign(form, { id: null, content_zh: '', content_en: '', sort_order: 0, is_active: true }) }

const handleAdd = () => { resetForm(); isEdit.value = false; dialogVisible.value = true }

const handleEdit = (row) => {
  Object.assign(form, { id: row.id, content_zh: row.content_zh, content_en: row.content_en, sort_order: row.sort_order, is_active: row.is_active })
  isEdit.value = true
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (formRef.value) {
    try { await formRef.value.validate() } catch { return }
  }
  submitting.value = true
  try {
    const payload = { content_zh: form.content_zh, content_en: form.content_en, sort_order: form.sort_order, is_active: form.is_active }
    if (isEdit.value) {
      await updateAnnouncement(form.id, payload)
      showSuccessToast(t('announcement.updated'))
    } else {
      payload.type = activeTab.value
      await createAnnouncement(payload)
      showSuccessToast(t('announcement.added'))
    }
    dialogVisible.value = false
    loadData()
  } catch (error) {
    console.error('提交失败:', error)
  } finally {
    submitting.value = false
  }
}

const toggleActive = async (row, val) => {
  try {
    await updateAnnouncement(row.id, { is_active: val })
    loadData()
  } catch (error) {
    console.error('切换状态失败:', error)
  }
}

const handleDelete = async (row) => {
  try {
    await showConfirmDialog({
      title: t('common.confirmDelete'),
      message: t('announcement.deleteConfirm'),
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
    })
    await deleteAnnouncement(row.id)
    showSuccessToast(t('announcement.deleted'))
    loadData()
  } catch {
    // cancelled
  }
}

onMounted(() => { loadData() })
</script>

<style scoped>
.card-list { display: flex; flex-direction: column; gap: 10px; }

.list-card {
  background: #fff;
  border-radius: 10px;
  padding: 14px;
  box-shadow: 0 1px 4px rgba(0,0,0,0.06);
  cursor: pointer;
}

.list-card:active { opacity: 0.85; }

.card-top { display: flex; justify-content: space-between; align-items: flex-start; gap: 12px; margin-bottom: 6px; }

.card-content-preview {
  flex: 1;
  font-size: 14px;
  font-weight: 500;
  color: #1a1a1a;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-content-en {
  font-size: 12px;
  color: #8c8c8c;
  margin-bottom: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-bottom { display: flex; justify-content: space-between; align-items: center; padding-top: 8px; border-top: 1px solid #f5f5f5; }
.card-sort-badge { font-size: 12px; color: #999; }
</style>
