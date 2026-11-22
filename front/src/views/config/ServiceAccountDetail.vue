<template>
  <div class="serviceaccount-detail detail-page">
    <PageHeader
      :title="sa?.name || name"
      :subtitle="sa ? `命名空间：${sa.namespace}` : 'ServiceAccount 详情'"
      icon="sa"
      back
      @back="handleBack"
    >
      <el-button @click="handleMore('editInfo')">
        <el-icon><EditPen /></el-icon>
        编辑
      </el-button>
      <el-button @click="handleMore('yaml')">
        <el-icon><Document /></el-icon>
        查看YAML
      </el-button>
      <el-button type="danger" @click="openDelete">
        <el-icon><Delete /></el-icon>
        删除
      </el-button>
    </PageHeader>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="8" animated />
    </div>

    <div v-else-if="sa" class="detail-content">
      <el-card class="info-card">
        <template #header>
          <div class="card-header"><span>基本信息</span></div>
        </template>
        <div class="info-grid">
          <div class="info-item"><label>名称</label><span>{{ sa.name }}</span></div>
          <div class="info-item"><label>命名空间</label><span>{{ sa.namespace }}</span></div>
          <div class="info-item"><label>自动挂载令牌</label><span>{{ sa.automountServiceAccountToken === false ? '否' : '是' }}</span></div>
          <div class="info-item"><label>创建时间</label><span>{{ formatDateTime(sa.creationTimestamp) }}</span></div>
        </div>
      </el-card>

      <el-row :gutter="20" style="margin-top: 20px;">
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>挂载的 Secrets</span>
                <el-tag size="small" type="info">{{ secretNames.length }}</el-tag>
              </div>
            </template>
            <div class="related-content">
              <div v-if="secretNames.length === 0" class="empty-state">
                <el-empty description="暂无" :image-size="60" />
              </div>
              <div v-else class="related-list">
                <div v-for="n in secretNames" :key="n" class="related-item">
                  <div class="related-name">{{ n }}</div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>镜像拉取 Secrets</span>
                <el-tag size="small" type="info">{{ imagePullSecretNames.length }}</el-tag>
              </div>
            </template>
            <div class="related-content">
              <div v-if="imagePullSecretNames.length === 0" class="empty-state">
                <el-empty description="暂无" :image-size="60" />
              </div>
              <div v-else class="related-list">
                <div v-for="n in imagePullSecretNames" :key="n" class="related-item">
                  <div class="related-name">{{ n }}</div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" style="margin-top: 20px;">
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>标签</span>
                <el-tag size="small" type="info">{{ labelRows.length }}</el-tag>
              </div>
            </template>
            <div class="labels-content">
              <div v-if="labelRows.length === 0" class="empty-state">
                <el-empty description="暂无标签" :image-size="60" />
              </div>
              <div v-else class="labels-list">
                <div v-for="row in labelRows" :key="row.key" class="label-item">
                  <div class="label-key">{{ row.key }}</div>
                  <div class="label-value">{{ row.value }}</div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>注解</span>
                <el-tag size="small" type="info">{{ annotationRows.length }}</el-tag>
              </div>
            </template>
            <div class="annotations-content">
              <div v-if="annotationRows.length === 0" class="empty-state">
                <el-empty description="暂无注解" :image-size="60" />
              </div>
              <div v-else class="annotations-list">
                <div v-for="row in annotationRows" :key="row.key" class="annotation-item">
                  <div class="annotation-key">{{ row.key }}</div>
                  <div class="annotation-value">{{ row.value }}</div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <div v-else class="empty-container">
      <el-empty description="未找到ServiceAccount" />
    </div>

    <!-- 编辑信息(与命名空间编辑弹窗同款) -->
    <el-dialog v-model="editInfoDialogVisible" title="编辑信息" width="800px" :close-on-click-modal="false" class="config-dialog form-dialog--compact">
      <template #header>
        <div class="dialog-header">
          <div>
            <h3 class="dialog-title">编辑信息</h3>
          </div>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <el-form :model="editInfoForm" label-width="96px" class="serviceaccount-form">
            <el-form-item label="名称">
              <el-input :model-value="sa?.name || ''" disabled />
            </el-form-item>
            <el-form-item label="标签(JSON)" prop="labelsJson">
              <el-input v-model="editInfoForm.labelsJson" type="textarea" :rows="3" placeholder='{"key": "value"}' />
            </el-form-item>
            <el-form-item label="注解(JSON)" prop="annotationsJson">
              <el-input v-model="editInfoForm.annotationsJson" type="textarea" :rows="3" placeholder='{"key": "value"}' />
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
    <YamlDialog
      v-model="yamlDialogVisible"
      title="查看/编辑YAML"
      :content="yamlContent"
      :apply-target="{ kind: 'serviceaccount', clusterId, namespace, name }"
      @applied="fetchDetailFromList()"
    />

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
import { formatDateTime } from '@/utils/format'
import { fetchResourceYaml } from '@/api/resource'
import PageHeader from '@/components/PageHeader.vue'
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { EditPen, Document, Delete } from '@element-plus/icons-vue'
import { serviceAccountApi, type ServiceAccountVO } from '@/api/serviceaccount'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import '@/assets/styles/config-editor.css'
import { useTabBack } from '@/composables/useTabBack'

const route = useRoute()
// 返回当前标签内的上一页
const goBack = useTabBack()

const clusterId = ref('')
const namespace = ref('')
const name = ref('')

const loading = ref(false)
const sa = ref<ServiceAccountVO | null>(null)

const labelRows = ref<{ key: string; value: string }[]>([])
const annotationRows = ref<{ key: string; value: string }[]>([])

const imagePullSecretNames = computed(() => (sa.value?.imagePullSecrets || []).map(x => x.name).filter(Boolean))
const secretNames = computed(() => (sa.value?.secrets || []).map(x => x.name).filter(Boolean))

const handleBack = () => goBack()

const openYaml = async () => {
  try {
    yamlContent.value = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'serviceaccount', namespace: namespace.value, name: name.value })
    yamlDialogVisible.value = true
  } catch {
    // 请求拦截器已提示错误
  }
}

const handleMore = (cmd: string) => {
  switch (cmd) {
    case 'editInfo':
      if (!sa.value) return
      editInfoForm.labelsJson = JSON.stringify(sa.value.labels || {}, null, 2)
      editInfoForm.annotationsJson = JSON.stringify(sa.value.annotations || {}, null, 2)
      editInfoDialogVisible.value = true
      break
    case 'yaml':
      openYaml()
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
    sa.value = null
  } finally {
    loading.value = false
  }
}

const editInfoDialogVisible = ref(false)
const editInfoLoading = ref(false)
const editInfoForm = reactive({
  labelsJson: '',
  annotationsJson: ''
})

const confirmEditInfo = async () => {
  if (!sa.value) return
  // 先校验 JSON, 格式错误时单独提示; 接口错误由请求拦截器提示
  let labels: Record<string, string> = {}
  let annotations: Record<string, string> = {}
  try {
    if (editInfoForm.labelsJson.trim()) labels = JSON.parse(editInfoForm.labelsJson)
    if (editInfoForm.annotationsJson.trim()) annotations = JSON.parse(editInfoForm.annotationsJson)
  } catch {
    ElMessage.error('标签或注解不是合法的 JSON')
    return
  }
  editInfoLoading.value = true
  try {
    await serviceAccountApi.update({
      clusterId: clusterId.value,
      namespace: namespace.value,
      name: name.value,
      automountServiceAccountToken: sa.value.automountServiceAccountToken,
      imagePullSecrets: imagePullSecretNames.value,
      secrets: secretNames.value,
      labels,
      annotations
    })
    ElMessage.success('更新成功')
    editInfoDialogVisible.value = false
    fetchDetailFromList()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    editInfoLoading.value = false
  }
}

const yamlDialogVisible = ref(false)
const yamlContent = ref('')

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
    // 请求拦截器已提示错误
  } finally {
    deleteLoading.value = false
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
.serviceaccount-detail {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-left h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
  word-break: break-all;
}

.header-right {
  display: flex;
  gap: 8px;
}

.loading-container {
  padding: 20px;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
}

.info-item {
  display: flex;
  align-items: center;
}

.info-item label {
  font-weight: 500;
  color: #606266;
  min-width: 120px;
}

.related-content,
.labels-content,
.annotations-content {
  max-height: 280px;
  overflow-y: auto;
}

.empty-state {
  text-align: center;
  padding: 20px;
}

.related-list,
.labels-list,
.annotations-list {
  padding: 0;
}

.related-item {
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  margin-bottom: 8px;
  background-color: #fafafa;
}

.related-name {
  font-weight: 500;
  color: #409eff;
  font-size: 13px;
  word-break: break-all;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
}

.label-item,
.annotation-item {
  display: flex;
  flex-direction: column;
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  margin-bottom: 8px;
  background-color: #fafafa;
}

.label-key,
.annotation-key {
  font-weight: 500;
  color: #409eff;
  font-size: 12px;
  margin-bottom: 4px;
  word-break: break-all;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
}

.label-value,
.annotation-value {
  color: #606266;
  font-size: 13px;
  word-break: break-all;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
  white-space: pre-wrap;
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

:deep(.el-card__header) {
  background-color: #f8f9fa;
  border-bottom: 1px solid #ebeef5;
}

:deep(.el-card__body) {
  padding: 20px;
}

@media (max-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
}
</style>
