<template>
  <div class="detail-page">
    <!-- 头部卡片 -->
    <el-card class="ks-header-card" shadow="never">
      <div class="ks-breadcrumb-row">
        <el-link :underline="false" class="ks-back" @click="handleBack">
          <el-icon><ArrowLeft /></el-icon>
          <span>服务帐户</span>
        </el-link>
        <el-dropdown trigger="click" popper-class="ks-dark-menu" @command="handleMore">
          <el-button class="ks-action-btn">
            操作
            <el-icon class="el-icon--right"><ArrowUp /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="editInfo"><el-icon><EditPen /></el-icon>编辑信息</el-dropdown-item>
              <el-dropdown-item command="yaml"><el-icon><Document /></el-icon>查看/编辑YAML</el-dropdown-item>
              <el-dropdown-item command="delete" divided><el-icon><Delete /></el-icon>删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      <div class="ks-title-row">
        <div class="ks-icon-block"><el-icon :size="24"><User /></el-icon></div>
        <div class="ks-title-main">
          <div class="ks-name">{{ sa?.name || name }}</div>
          <div class="ks-desc">{{ sa?.annotations?.describe || '-' }}</div>
        </div>
      </div>
    </el-card>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="8" animated />
    </div>

    <template v-else-if="sa">
      <!-- 属性 -->
      <el-card class="ks-card" shadow="never">
        <template #header>
          <div class="ks-card-header"><span>属性</span></div>
        </template>
        <div class="props-grid">
          <div class="prop-item"><div class="prop-label">集群</div><div class="prop-value">{{ clusterId }}</div></div>
          <div class="prop-item"><div class="prop-label">命名空间</div><div class="prop-value">{{ sa.namespace }}</div></div>
          <div class="prop-item"><div class="prop-label">自动挂载令牌</div><div class="prop-value">{{ sa.automountServiceAccountToken === false ? '否' : '是' }}</div></div>
          <div class="prop-item"><div class="prop-label">创建时间</div><div class="prop-value">{{ formatDate(sa.creationTimestamp) }}</div></div>
        </div>
      </el-card>

      <!-- 挂载的 Secrets -->
      <el-card class="ks-card" shadow="never">
        <template #header><div class="ks-card-header"><span>挂载的 Secrets ({{ secretNames.length }})</span></div></template>
        <div v-if="secretNames.length" class="ks-chips">
          <span v-for="n in secretNames" :key="n" class="ks-chip">{{ n }}</span>
        </div>
        <el-empty v-else description="暂无" :image-size="60" />
      </el-card>

      <!-- 镜像拉取 Secrets -->
      <el-card class="ks-card" shadow="never">
        <template #header><div class="ks-card-header"><span>镜像拉取 Secrets ({{ imagePullSecretNames.length }})</span></div></template>
        <div v-if="imagePullSecretNames.length" class="ks-chips">
          <span v-for="n in imagePullSecretNames" :key="n" class="ks-chip">{{ n }}</span>
        </div>
        <el-empty v-else description="暂无" :image-size="60" />
      </el-card>

      <!-- 标签 -->
      <el-card class="ks-card" shadow="never">
        <template #header><div class="ks-card-header"><span>标签 ({{ labelRows.length }})</span></div></template>
        <div v-if="labelRows.length" class="ks-chips">
          <span v-for="row in labelRows" :key="row.key" class="ks-chip"><span class="ks-chip-key">{{ row.key }}</span><span class="ks-chip-value">= {{ row.value }}</span></span>
        </div>
        <el-empty v-else description="暂无标签" :image-size="60" />
      </el-card>

      <!-- 注解 -->
      <el-card class="ks-card" shadow="never">
        <template #header><div class="ks-card-header"><span>注解 ({{ annotationRows.length }})</span></div></template>
        <div v-if="annotationRows.length" class="ks-chips">
          <span v-for="row in annotationRows" :key="row.key" class="ks-chip"><span class="ks-chip-key">{{ row.key }}</span><span class="ks-chip-value">= {{ row.value }}</span></span>
        </div>
        <el-empty v-else description="暂无注解" :image-size="60" />
      </el-card>
    </template>

    <div v-else class="empty-container">
      <el-empty description="未找到ServiceAccount" />
    </div>

    <!-- 编辑信息 -->
    <el-dialog v-model="editInfoDialogVisible" title="编辑信息" width="760px" :close-on-click-modal="false" class="config-dialog">
      <template #header>
        <div class="dialog-header">
          <h3 class="dialog-title">编辑信息</h3>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <el-form :model="editInfoForm" label-width="130px">
            <el-form-item label="自动挂载令牌">
              <el-switch v-model="editInfoForm.automountServiceAccountToken" />
            </el-form-item>
            <el-form-item label="标签 (JSON)">
              <el-input v-model="editInfoForm.labelsJson" type="textarea" :rows="4" placeholder='{"key": "value"}' />
            </el-form-item>
            <el-form-item label="注解 (JSON)">
              <el-input v-model="editInfoForm.annotationsJson" type="textarea" :rows="4" placeholder='{"key": "value"}' />
            </el-form-item>
          </el-form>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="editInfoDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="editInfoLoading" @click="confirmEditInfo">确定</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 查看/编辑YAML -->
    <el-dialog v-model="yamlDialogVisible" title="查看/编辑YAML" width="90%" :close-on-click-modal="false" class="config-dialog yaml-dialog">
      <template #header>
        <div class="dialog-header">
          <h3 class="dialog-title">查看/编辑YAML</h3>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <div class="yaml-editor-wrapper">
            <YamlEditor :model-value="yamlContent" :readonly="false" height="100%" @update:model-value="val => yamlContent = val" />
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="yamlDialogVisible = false">关闭</el-button>
          <el-button type="primary" :loading="yamlApplyLoading" @click="confirmApplyYaml">应用</el-button>
          <span style="color:#909399;font-size:12px;margin-left:10px;">应用到当前集群：{{ clusterId }}</span>
        </div>
      </template>
    </el-dialog>

    <!-- 删除确认 -->
    <DeleteConfirmDialog
      v-model="deleteDialogVisible"
      :item-name="sa?.name || ''"
      message="确定要删除该ServiceAccount吗？"
      :loading="deleteLoading"
      @confirm="confirmDelete"
      @cancel="() => deleteDialogVisible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, ArrowUp, EditPen, Document, Delete, User } from '@element-plus/icons-vue'
import { serviceAccountApi, type ServiceAccountVO } from '@/api/serviceaccount'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import '@/assets/styles/config-editor.css'

const route = useRoute()
const router = useRouter()

const clusterId = ref('')
const namespace = ref('')
const name = ref('')

const loading = ref(false)
const sa = ref<ServiceAccountVO | null>(null)

const labelRows = ref<{ key: string; value: string }[]>([])
const annotationRows = ref<{ key: string; value: string }[]>([])

const imagePullSecretNames = computed(() => (sa.value?.imagePullSecrets || []).map(x => x.name).filter(Boolean))
const secretNames = computed(() => (sa.value?.secrets || []).map(x => x.name).filter(Boolean))

const handleBack = () => router.back()

const handleMore = (cmd: string) => {
  switch (cmd) {
    case 'editInfo':
      if (!sa.value) return
      editInfoForm.automountServiceAccountToken = sa.value.automountServiceAccountToken !== false
      editInfoForm.labelsJson = JSON.stringify(sa.value.labels || {}, null, 2)
      editInfoForm.annotationsJson = JSON.stringify(sa.value.annotations || {}, null, 2)
      editInfoDialogVisible.value = true
      break
    case 'yaml':
      yamlContent.value = sa.value?.yaml || ''
      yamlDialogVisible.value = true
      break
    case 'delete':
      openDelete()
      break
  }
}

const fetchDetailFromList = async () => {
  if (!clusterId.value || !namespace.value || !name.value) {
    sa.value = null
    return
  }
  loading.value = true
  try {
    const resp = await serviceAccountApi.getList({ clusterId: clusterId.value, namespace: namespace.value, page: 1, pageSize: 10000 })
    if (resp.code === 200) {
      const items = resp.data.items || []
      const found = items.find((it: ServiceAccountVO) => it.name === name.value) || null
      sa.value = found
      labelRows.value = Object.entries(found?.labels || {}).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
      annotationRows.value = Object.entries(found?.annotations || {}).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
      return
    }
    sa.value = null
  } catch {
    ElMessage.error('获取详情失败')
    sa.value = null
  } finally {
    loading.value = false
  }
}

const editInfoDialogVisible = ref(false)
const editInfoLoading = ref(false)
const editInfoForm = reactive({
  automountServiceAccountToken: true,
  labelsJson: '',
  annotationsJson: ''
})

const confirmEditInfo = async () => {
  if (!sa.value) return
  editInfoLoading.value = true
  try {
    let labels: Record<string, string> = {}
    let annotations: Record<string, string> = {}
    if (editInfoForm.labelsJson.trim()) {
      labels = JSON.parse(editInfoForm.labelsJson)
    }
    if (editInfoForm.annotationsJson.trim()) {
      annotations = JSON.parse(editInfoForm.annotationsJson)
    }
    await serviceAccountApi.update({
      clusterId: clusterId.value,
      namespace: namespace.value,
      name: name.value,
      automountServiceAccountToken: editInfoForm.automountServiceAccountToken,
      imagePullSecrets: imagePullSecretNames.value,
      secrets: secretNames.value,
      labels,
      annotations
    })
    ElMessage.success('更新成功')
    editInfoDialogVisible.value = false
    fetchDetailFromList()
  } catch {
    ElMessage.error('更新失败(请检查 JSON 格式)')
  } finally {
    editInfoLoading.value = false
  }
}

const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlApplyLoading = ref(false)

const confirmApplyYaml = async () => {
  if (!yamlContent.value) { ElMessage.warning('YAML内容为空'); return }
  yamlApplyLoading.value = true
  try {
    await serviceAccountApi.applyYaml({ clusterId: clusterId.value, yaml: yamlContent.value })
    ElMessage.success('YAML应用成功')
    yamlDialogVisible.value = false
    fetchDetailFromList()
  } catch {
    ElMessage.error('YAML应用失败')
  } finally {
    yamlApplyLoading.value = false
  }
}

const deleteDialogVisible = ref(false)
const deleteLoading = ref(false)

const openDelete = () => { deleteDialogVisible.value = true }
const confirmDelete = async () => {
  deleteLoading.value = true
  try {
    await serviceAccountApi.delete({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })
    ElMessage.success('删除成功')
    deleteDialogVisible.value = false
    handleBack()
  } catch {
    ElMessage.error('删除失败')
  } finally {
    deleteLoading.value = false
  }
}

const formatDate = (dateString?: string) => {
  if (!dateString) return '-'
  try {
    const d = new Date(dateString)
    if (isNaN(d.getTime())) return '-'
    return d.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  } catch {
    return '-'
  }
}

onMounted(() => {
  clusterId.value = (route.query.clusterId as string) || ''
  namespace.value = (route.query.namespace as string) || ''
  name.value = (route.query.name as string) || ''
  fetchDetailFromList()
})
</script>

<style scoped>
.detail-page {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.loading-container {
  padding: 20px;
  background: #fff;
  border-radius: 8px;
}

.ks-breadcrumb-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.ks-back {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #1f2937;
}

.ks-back:hover {
  color: var(--el-color-primary);
}

.ks-title-row {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-top: 16px;
}

.ks-icon-block {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  background: linear-gradient(135deg, #10b981, #34d399);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.ks-title-main {
  min-width: 0;
}

.ks-name {
  font-size: 22px;
  font-weight: 600;
  color: #1f2937;
  word-break: break-all;
  line-height: 1.3;
}

.ks-desc {
  margin-top: 4px;
  color: #9ca3af;
  font-size: 13px;
}

.ks-card {
  border-radius: 8px;
}

.ks-card-header {
  font-weight: 600;
  color: #1f2937;
}

.props-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 18px 32px;
}

.prop-label {
  font-size: 12px;
  color: #9ca3af;
  margin-bottom: 6px;
}

.prop-value {
  font-size: 14px;
  color: #1f2937;
  font-weight: 500;
  word-break: break-all;
}

.ks-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.ks-chip {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  padding: 4px 10px;
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  font-size: 12px;
  color: #475569;
}

.ks-chip-key {
  color: #1f2937;
  margin-right: 4px;
}

.ks-chip-value {
  color: #6b7280;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.config-dialog :deep(.el-dialog__header) {
  padding: 0;
  margin: 0;
}

.config-dialog :deep(.el-dialog__body) {
  padding: 0;
}

.config-dialog :deep(.el-dialog__footer) {
  padding: 0;
}
</style>

<!-- 操作下拉深色菜单(KubeSphere 风格), popper 挂在 body 上需全局样式 -->
<style>
.ks-dark-menu.el-popper {
  background: #1f2937;
  border: 1px solid #1f2937;
}

.ks-dark-menu .el-dropdown-menu {
  background: #1f2937;
  padding: 6px;
}

.ks-dark-menu .el-dropdown-menu__item {
  color: #e5e7eb;
  padding: 9px 16px;
  border-radius: 4px;
}

.ks-dark-menu .el-dropdown-menu__item:hover {
  background: #374151;
  color: #fff;
}

.ks-dark-menu .el-popper__arrow::before {
  background: #1f2937;
  border: 1px solid #1f2937;
}
</style>
