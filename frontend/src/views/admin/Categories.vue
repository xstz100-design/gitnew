<template>
  <div class="page-container">
    <div class="page-header">
      <h2>{{ $t('category.title') }}</h2>
      <van-button type="primary" size="small" icon="plus" @click="handleAdd">{{ $t('category.addCategory') }}</van-button>
    </div>

    <van-loading v-if="loading" size="24" vertical style="padding: 40px 0; text-align:center">加载中...</van-loading>

    <div v-else class="card-list">
      <div v-for="row in categories" :key="row.id" class="list-card" @click="handleEdit(row)">
        <div class="card-top">
          <div class="card-name-area">
            <span class="card-sort">{{ row.sort_order }}</span>
            <span class="card-name">{{ row.name }}</span>
          </div>
          <van-switch :model-value="row.is_active" size="20" @change="(val) => handleToggleActive(row, val)" @click.stop />
        </div>
        <div class="card-bottom">
          <span class="card-time">{{ formatDate(row.created_at) }}</span>
          <van-button type="danger" size="mini" plain @click.stop="handleDelete(row)">{{ $t('common.delete') }}</van-button>
        </div>
      </div>
      <van-empty v-if="categories.length === 0" :description="$t('common.noData')" />
    </div>

    <!-- 添加/编辑弹窗 -->
    <van-popup v-model:show="dialogVisible" position="bottom" round :style="{ minHeight: '50vh' }" destroy-on-close>
      <van-nav-bar
        :title="isEdit ? $t('category.editCategory') : $t('category.addCategory')"
        :left-text="$t('common.cancel')"
        :right-text="$t('common.confirm')"
        @click-left="dialogVisible = false"
        @click-right="handleSubmit"
      />
      <van-form ref="formRef" style="padding-top: 8px;">
        <van-field v-model="form.name" :label="$t('category.name')" :placeholder="$t('category.namePlaceholder')"
          :rules="[{ required: true, message: t('category.nameRequired') }]" />
        <van-field v-model.number="form.sort_order" type="number" :label="$t('category.sortOrder')" placeholder="0"
          :extra="$t('category.sortHint')" />
        <van-cell v-if="isEdit" :title="$t('common.status')">
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
import { showSuccessToast, showFailToast, showConfirmDialog } from 'vant'
import { getAllCategories, createCategory, updateCategory, deleteCategory } from '@/api'

const { t } = useI18n()
const loading = ref(false)
const categories = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()

const form = reactive({ id: null, name: '', sort_order: 0, is_active: true })

const formatDate = (str) => {
  if (!str) return ''
  const d = new Date(str)
  return d.toLocaleDateString('zh-CN') + ' ' + d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

const loadCategories = async () => {
  loading.value = true
  try {
    categories.value = await getAllCategories()
  } catch (error) {
    console.error('加载分类失败:', error)
  } finally {
    loading.value = false
  }
}

const resetForm = () => { Object.assign(form, { id: null, name: '', sort_order: 0, is_active: true }) }

const handleAdd = () => { resetForm(); isEdit.value = false; dialogVisible.value = true }

const handleEdit = (row) => {
  Object.assign(form, { id: row.id, name: row.name, sort_order: row.sort_order, is_active: row.is_active })
  isEdit.value = true
  dialogVisible.value = true
}

const handleToggleActive = async (row, val) => {
  try {
    await updateCategory(row.id, { is_active: val })
    row.is_active = val
    showSuccessToast(val ? t('common.enabled') : t('common.disabled'))
  } catch {
    showFailToast(t('common.operationFailed'))
  }
}

const handleSubmit = async () => {
  if (formRef.value) {
    try { await formRef.value.validate() } catch { return }
  }
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateCategory(form.id, { name: form.name, sort_order: form.sort_order, is_active: form.is_active })
      showSuccessToast(t('category.categoryUpdated'))
    } else {
      await createCategory({ name: form.name, sort_order: form.sort_order })
      showSuccessToast(t('category.categoryAdded'))
    }
    dialogVisible.value = false
    loadCategories()
  } catch (error) {
    console.error('提交失败:', error)
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (row) => {
  try {
    await showConfirmDialog({
      title: t('common.confirmDelete'),
      message: t('category.deleteConfirm', { name: row.name }),
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
    })
    await deleteCategory(row.id)
    showSuccessToast(t('category.categoryDeleted'))
    loadCategories()
  } catch {
    // cancelled
  }
}

onMounted(() => { loadCategories() })
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

.card-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }

.card-name-area { display: flex; align-items: center; gap: 10px; flex: 1; min-width: 0; }

.card-sort {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px; height: 24px;
  background: #1d4ed8;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  border-radius: 6px;
  flex-shrink: 0;
}

.card-name {
  font-size: 15px;
  font-weight: 600;
  color: #1a1a1a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-bottom { display: flex; justify-content: space-between; align-items: center; }
.card-time { font-size: 12px; color: #999; }
</style>
