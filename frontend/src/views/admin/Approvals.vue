<template>
  <div class="approval-page">
    <div class="page-header">
      <h2>{{ $t('admin.userApproval') }}</h2>
      <el-radio-group v-model="statusFilter" size="small" @change="loadUsers">
        <el-radio-button value="pending">{{ $t('admin.statusPending') }} <span v-if="pendingCount > 0" class="count-badge">{{ pendingCount }}</span></el-radio-button>
        <el-radio-button value="approved">{{ $t('admin.statusApproved') }}</el-radio-button>
        <el-radio-button value="rejected">{{ $t('admin.statusRejected') }}</el-radio-button>
        <el-radio-button value="">{{ $t('admin.allRegistrations') }}</el-radio-button>
      </el-radio-group>
    </div>

    <!-- 桌面端: 表格视图 -->
    <el-table v-if="!mobile" v-loading="loading" :data="users" border>
      <el-table-column :label="$t('profile.phone')" prop="username" width="140" />
      <el-table-column :label="$t('profile.name')" prop="full_name" min-width="120" />
      <el-table-column :label="$t('profile.address')" prop="address" min-width="180" show-overflow-tooltip />
      <el-table-column :label="$t('admin.registrationTime')" prop="created_at" width="150" />
      <el-table-column :label="$t('admin.approvalStatus')" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.approval_status)" size="small">
            {{ getStatusText(row.approval_status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="$t('admin.rejectReason')" prop="rejected_reason" min-width="150" show-overflow-tooltip />
      <el-table-column :label="$t('common.operation')" width="160" fixed="right">
        <template #default="{ row }">
          <template v-if="row.approval_status === 'pending'">
            <el-button type="success" size="small" @click="handleApprove(row)">{{ $t('admin.approve') }}</el-button>
            <el-button type="danger" size="small" @click="handleReject(row)">{{ $t('admin.reject') }}</el-button>
          </template>
          <span v-else style="color: #909399;">{{ getStatusText(row.approval_status) }}</span>
        </template>
      </el-table-column>
    </el-table>

    <!-- 移动端: 卡片列表 -->
    <div v-else v-loading="loading" class="mobile-card-list">
      <div v-for="row in users" :key="row.id" class="user-card">
        <div class="card-top">
          <div class="card-user-info">
            <span class="card-username">{{ row.username }}</span>
            <el-tag :type="getStatusType(row.approval_status)" size="small">
              {{ getStatusText(row.approval_status) }}
            </el-tag>
          </div>
        </div>
        <div class="card-body">
          <div class="card-row">
            <span class="card-label">{{ $t('profile.name') }}</span>
            <span class="card-value">{{ row.full_name }}</span>
          </div>
          <div v-if="row.address" class="card-row">
            <span class="card-label">{{ $t('profile.address') }}</span>
            <span class="card-value text-ellipsis">{{ row.address }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">{{ $t('admin.registrationTime') }}</span>
            <span class="card-value">{{ row.created_at }}</span>
          </div>
          <div v-if="row.rejected_reason" class="card-row">
            <span class="card-label">{{ $t('admin.rejectReason') }}</span>
            <span class="card-value text-danger">{{ row.rejected_reason }}</span>
          </div>
        </div>
        <div v-if="row.approval_status === 'pending'" class="card-footer">
          <el-button type="success" size="small" @click="handleApprove(row)">{{ $t('admin.approve') }}</el-button>
          <el-button type="danger" size="small" @click="handleReject(row)">{{ $t('admin.reject') }}</el-button>
        </div>
      </div>
      <el-empty v-if="!loading && users.length === 0" :description="$t('admin.noPendingUsers')" />
    </div>

    <!-- 拒绝原因对话框 -->
    <el-dialog
      v-model="rejectDialogVisible"
      :title="$t('admin.reject')"
      width="400px"
    >
      <el-form ref="rejectFormRef" :model="rejectForm" :rules="rejectRules" label-width="100px">
        <el-form-item :label="$t('admin.rejectReason')" prop="reason">
          <el-input
            v-model="rejectForm.reason"
            type="textarea"
            :rows="3"
            :placeholder="$t('admin.rejectReasonPlaceholder')"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialogVisible = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="danger" :loading="submitting" @click="confirmReject">{{ $t('admin.reject') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAllRegistrations, getPendingCount, approveUser } from '@/api'

const { t } = useI18n()

// 移动端检测
const mobile = ref(window.innerWidth < 768)
const onResize = () => { mobile.value = window.innerWidth < 768 }
onMounted(() => {
  window.addEventListener('resize', onResize)
  loadUsers()
  loadPendingCount()
})
onBeforeUnmount(() => window.removeEventListener('resize', onResize))

const loading = ref(false)
const users = ref([])
const statusFilter = ref('pending')
const pendingCount = ref(0)

const rejectDialogVisible = ref(false)
const rejectFormRef = ref()
const submitting = ref(false)
const currentUser = ref(null)

const rejectForm = reactive({
  reason: '',
})

const rejectRules = {
  reason: [
    { required: true, message: () => t('admin.rejectReasonRequired'), trigger: 'blur' },
  ],
}

const getStatusType = (status) => {
  switch (status) {
    case 'pending': return 'warning'
    case 'approved': return 'success'
    case 'rejected': return 'danger'
    default: return 'info'
  }
}

const getStatusText = (status) => {
  switch (status) {
    case 'pending': return t('admin.statusPending')
    case 'approved': return t('admin.statusApproved')
    case 'rejected': return t('admin.statusRejected')
    default: return status
  }
}

const loadUsers = async () => {
  loading.value = true
  try {
    const data = await getAllRegistrations(statusFilter.value || null)
    users.value = data
  } catch (error) {
    console.error('加载用户列表失败:', error)
  } finally {
    loading.value = false
  }
}

const loadPendingCount = async () => {
  try {
    const data = await getPendingCount()
    pendingCount.value = data.count
  } catch (error) {
    console.error('获取待审核数量失败:', error)
  }
}

const handleApprove = async (row) => {
  try {
    await ElMessageBox.confirm(
      t('admin.approveConfirm', { name: row.full_name }),
      t('admin.hint'),
      { type: 'success', confirmButtonText: t('admin.approve'), cancelButtonText: t('common.cancel') }
    )
    
    await approveUser(row.id, true)
    ElMessage.success(t('admin.userApproved'))
    loadUsers()
    loadPendingCount()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('审核失败:', error)
    }
  }
}

const handleReject = (row) => {
  currentUser.value = row
  rejectForm.reason = ''
  rejectDialogVisible.value = true
}

const confirmReject = async () => {
  if (!rejectFormRef.value) return
  try {
    await rejectFormRef.value.validate()
  } catch {
    return
  }
  
  submitting.value = true
  try {
    await approveUser(currentUser.value.id, false, rejectForm.reason)
    ElMessage.success(t('admin.userRejected'))
    rejectDialogVisible.value = false
    loadUsers()
    loadPendingCount()
  } catch (error) {
    console.error('拒绝失败:', error)
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped lang="scss">
.approval-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;

  h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }
}

.count-badge {
  display: inline-block;
  min-width: 18px;
  height: 18px;
  line-height: 18px;
  padding: 0 5px;
  font-size: 12px;
  background: #f56c6c;
  color: #fff;
  border-radius: 9px;
  margin-left: 4px;
}

.mobile-card-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.user-card {
  background: #fff;
  border-radius: 8px;
  border: 1px solid #ebeef5;
  overflow: hidden;

  .card-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    background: #fafafa;
    border-bottom: 1px solid #ebeef5;
  }

  .card-user-info {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .card-username {
    font-weight: 600;
    font-size: 15px;
    color: #303133;
  }

  .card-body {
    padding: 12px 16px;
  }

  .card-row {
    display: flex;
    justify-content: space-between;
    padding: 6px 0;
    font-size: 14px;
  }

  .card-label {
    color: #909399;
    flex-shrink: 0;
  }

  .card-value {
    color: #303133;
    text-align: right;
    margin-left: 12px;

    &.text-ellipsis {
      max-width: 200px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    &.text-danger {
      color: #f56c6c;
    }
  }

  .card-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 12px 16px;
    border-top: 1px solid #ebeef5;
    background: #fafafa;
  }
}

@media (max-width: 768px) {
  .approval-page {
    padding: 12px;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
