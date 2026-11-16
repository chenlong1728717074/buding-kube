<template>
  <div class="secret-detail">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="handleBack">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1>{{ secret?.name || name }}</h1>
      </div>
      <div class="header-right">
        <el-button @click="handleMore('editInfo')">
          <el-icon><EditPen /></el-icon>
          编辑
        </el-button>
        <el-button @click="handleMore('editData')">
          <el-icon><Setting /></el-icon>
          编辑设置
        </el-button>
        <el-button @click="handleMore('yaml')">
          <el-icon><Document /></el-icon>
          查看YAML
        </el-button>
        <el-button type="danger" @click="openDelete">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="8" animated />
    </div>

    <div v-else-if="secret" class="detail-content">
      <el-card class="info-card">
        <template #header>
          <div class="card-header"><span>基本信息</span></div>
        </template>
        <div class="info-grid">
          <div class="info-item"><label>名称：</label><span>{{ secret.name }}</span></div>
          <div class="info-item"><label>命名空间：</label><span>{{ secret.namespace }}</span></div>
          <div class="info-item"><label>类型：</label><span>{{ secret.type || '-' }}</span></div>
          <div class="info-item"><label>别名：</label><span>{{ secret.alias || '-' }}</span></div>
          <div class="info-item"><label>备注：</label><span>{{ secret.describe || '-' }}</span></div>
          <div class="info-item"><label>创建时间：</label><span>{{ formatDate(secret.creationTimestamp) }}</span></div>
        </div>
      </el-card>

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

      <el-row style="margin-top: 20px;">
        <el-col :span="24">
          <el-card class="info-card">
            <template #header>
              <div class="card-header"><span>数据 ({{ dataRows.length }})</span></div>
            </template>
            <div class="data-hint">以下值已自动解码显示</div>
            <div class="data-content">
              <div v-if="dataRows.length === 0" class="empty-state">
                <el-empty description="暂无数据" :image-size="60" />
              </div>
              <div v-else class="data-list">
                <div v-for="row in dataRows" :key="row.key" class="data-item">
                  <div class="data-key">{{ row.key }}</div>
                  <div class="data-value">{{ row.value }}</div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <div v-else class="empty-container">
      <el-empty description="未找到Secret" />
    </div>

    <!-- 编辑信息(与列表页同款) -->
    <el-dialog v-model="editInfoDialogVisible" title="编辑信息" width="900px" :close-on-click-modal="false" class="config-dialog">
      <template #header>
        <div class="dialog-header">
          <h3 class="dialog-title">编辑信息</h3>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <div class="group-box" style="margin-bottom: 12px;">
            <div class="group-title">基本信息</div>
            <el-form :model="editInfoForm" label-width="120px">
              <el-form-item label="名称">
                <el-input :model-value="secret?.name || ''" disabled />
              </el-form-item>
              <el-form-item label="别名">
                <el-input v-model="editInfoForm.alias" placeholder="请输入别名" maxlength="63" show-word-limit />
                <div class="helper">别名只能包含中文、字母、数字和连字符（-），不得以连字符（-）开头或结尾，最长 63 个字符。</div>
              </el-form-item>
              <el-form-item label="备注">
                <el-input v-model="editInfoForm.describe" type="textarea" :rows="4" placeholder="请输入备注" maxlength="200" show-word-limit />
                <div class="helper">描述可包含任意字符，最长 256 个字符。</div>
              </el-form-item>
            </el-form>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="editInfoDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="editInfoLoading" @click="confirmEditInfo">确定</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 编辑数据设置(与列表页同款) -->
    <el-dialog v-model="editDataDialogVisible" title="编辑数据设置" width="1600px" :close-on-click-modal="false" class="config-dialog">
      <template #header>
        <div class="dialog-header">
          <h3 class="dialog-title">编辑数据设置</h3>
        </div>
      </template>

      <div class="config-editor">
        <div class="resource-info-bar">
          <div class="resource-info-item">
            <span class="resource-info-label">命名空间:</span>
            <span class="resource-info-value">{{ secret?.namespace || '-' }}</span>
          </div>
          <div class="resource-info-item">
            <span class="resource-info-label">名称:</span>
            <span class="resource-info-value">{{ secret?.name || '-' }}</span>
          </div>
        </div>

        <div class="config-content">
          <div class="data-form">
            <div class="form-section">
              <div class="section-header-row">
                <div class="section-title">数据 *(以明文 stringData 提交)</div>
                <div class="section-actions">
                  <el-button size="small" @click="kvEditRef?.addBlank()">添加数据</el-button>
                  <el-button size="small" @click="openEditImport">从文件导入</el-button>
                  <input ref="editImportInputRef" style="display: none" type="file" accept=".properties,.ini,.conf,.txt" @change="handleEditImportChange" />
                  <el-button size="small" @click="resetEditDataRows">重置</el-button>
                </div>
              </div>
              <div class="data-editor">
                <KVEditorPane ref="kvEditRef" v-model="editDataRows" height="500px" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="editDataDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="editDataLoading" @click="confirmEditData">保存</el-button>
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
      :item-name="secret?.name || ''"
      message="确定要删除该Secret吗？"
      :loading="deleteLoading"
      @confirm="confirmDelete"
      @cancel="() => deleteDialogVisible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, EditPen, Document, Setting, Delete, Key } from '@element-plus/icons-vue'
import { secretApi, type SecretVO } from '@/api/secret'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import KVEditorPane from '@/components/KVEditorPane.vue'
import '@/assets/styles/config-editor.css'

type KvRow = { key: string; value: string }

const rowsFromSecret = (row?: SecretVO | null) => {
  const sd = row?.stringData
  if (sd && Object.keys(sd).length > 0) {
    return Object.entries(sd).map(([k, v]) => ({ key: k, value: v }))
  }
  const data = row?.data || {}
  return Object.entries(data).map(([k, v]) => ({ key: k, value: decodeBase64ToUtf8(v) }))
}

const route = useRoute()
const router = useRouter()

const clusterId = ref('')
const namespace = ref('')
const name = ref('')

const loading = ref(false)
const secret = ref<SecretVO | null>(null)

const labelRows = ref<{ key: string; value: string }[]>([])
const annotationRows = ref<{ key: string; value: string }[]>([])
const dataRows = ref<{ key: string; value: string }[]>([])

const handleBack = () => router.back()

const handleMore = (cmd: string) => {
  switch (cmd) {
    case 'editInfo':
      if (!secret.value) return
      editInfoForm.alias = secret.value.alias || ''
      editInfoForm.describe = secret.value.describe || ''
      editInfoDialogVisible.value = true
      break
    case 'yaml':
      yamlContent.value = secret.value?.yaml || ''
      yamlDialogVisible.value = true
      break
    case 'editData':
      editDataRows.value = rowsFromSecret(secret.value)
      editDataDialogVisible.value = true
      break
    case 'delete':
      openDelete()
      break
  }
}

const decodeBase64ToUtf8 = (b64: string) => {
  try {
    const bin = atob(b64)
    const bytes = new Uint8Array(bin.length)
    for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i)
    return new TextDecoder().decode(bytes)
  } catch {
    return b64
  }
}

const fetchDetail = async () => {
  if (!clusterId.value || !namespace.value || !name.value) {
    secret.value = null
    return
  }
  loading.value = true
  try {
    const resp = await secretApi.getInfo({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })
    if (resp.code === 200) {
      secret.value = resp.data || null
      labelRows.value = Object.entries(resp.data?.labels || {}).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
      annotationRows.value = Object.entries(resp.data?.annotations || {})
        .filter(([k]) => k !== 'alias' && k !== 'describe')
        .map(([k, v]) => ({ key: k, value: String(v ?? '') }))
      const displayData = resp.data?.stringData && Object.keys(resp.data.stringData).length > 0 ? resp.data.stringData : undefined
      if (displayData) {
        dataRows.value = Object.entries(displayData).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
      } else {
        const data = (resp.data?.data || {}) as Record<string, string>
        dataRows.value = Object.entries(data).map(([k, v]) => ({ key: k, value: decodeBase64ToUtf8(String(v ?? '')) }))
      }
      return
    }
    secret.value = null
  } catch {
    ElMessage.error('获取详情失败')
    secret.value = null
  } finally {
    loading.value = false
  }
}

const editInfoDialogVisible = ref(false)
const editInfoLoading = ref(false)
const editInfoForm = reactive({ alias: '', describe: '' })

const editDataDialogVisible = ref(false)
const editDataLoading = ref(false)
const editDataRows = ref<KvRow[]>([])
const kvEditRef = ref()
const editImportInputRef = ref()

const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlApplyLoading = ref(false)

const deleteDialogVisible = ref(false)
const deleteLoading = ref(false)

const confirmEditInfo = async () => {
  if (!secret.value) return
  editInfoLoading.value = true
  try {
    await secretApi.updateInfo({
      clusterId: clusterId.value,
      namespace: namespace.value,
      name: name.value,
      alias: editInfoForm.alias,
      describe: editInfoForm.describe
    })
    ElMessage.success('更新成功')
    editInfoDialogVisible.value = false
    fetchDetail()
  } catch {
    ElMessage.error('更新失败')
  } finally {
    editInfoLoading.value = false
  }
}

// KV 工具函数(与列表页一致)
const collectRowsData = (rows: KvRow[]) => {
  const data: Record<string, string> = {}
  for (const { key, value } of rows) {
    const k = (key || '').trim()
    if (!k) continue
    data[k] = value ?? ''
  }
  return data
}

const parseKvFromText = (text: string) => {
  const result: Record<string, string> = {}
  const lines = String(text || '').split(/\r?\n/)
  let section = ''
  for (const raw of lines) {
    const line = raw.trim()
    if (!line) continue
    if (line.startsWith('#') || line.startsWith(';')) continue
    if (line.startsWith('[') && line.endsWith(']')) {
      section = line.slice(1, -1).trim()
      continue
    }
    const idx = line.indexOf('=')
    if (idx <= 0) continue
    const key = line.slice(0, idx).trim()
    const value = line.slice(idx + 1).trim()
    const fullKey = section ? `${section}.${key}` : key
    if (!fullKey) continue
    result[fullKey] = value
  }
  return result
}

const mergeDataRows = (rows: Array<{ key: string; value: string }>, incoming: Record<string, string>) => {
  const map = new Map(rows.map(r => [r.key, r]))
  for (const [k, v] of Object.entries(incoming)) {
    const existed = map.get(k)
    if (existed) {
      existed.value = v
    } else {
      rows.push({ key: k, value: v })
      map.set(k, rows[rows.length - 1])
    }
  }
}

const openEditImport = () => {
  editImportInputRef.value?.click()
}

const handleEditImportChange = async (e: Event) => {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    const text = await file.text()
    const parsed = parseKvFromText(text)
    const before = editDataRows.value.length
    mergeDataRows(editDataRows.value, parsed)
    const after = editDataRows.value.length
    ElMessage.success(`已导入 ${Object.keys(parsed).length} 项，新增 ${Math.max(after - before, 0)} 项`)
  } catch {
    ElMessage.error('导入失败')
  } finally {
    input.value = ''
  }
}

const resetEditDataRows = () => {
  editDataRows.value = rowsFromSecret(secret.value)
}

const confirmEditData = async () => {
  if (!secret.value) return
  editDataLoading.value = true
  try {
    const stringData = collectRowsData(editDataRows.value)
    await secretApi.updateData({
      clusterId: clusterId.value,
      namespace: namespace.value,
      name: name.value,
      stringData
    })
    ElMessage.success('保存成功')
    editDataDialogVisible.value = false
    fetchDetail()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    editDataLoading.value = false
  }
}

const confirmApplyYaml = async () => {
  if (!yamlContent.value) { ElMessage.warning('YAML内容为空'); return }
  yamlApplyLoading.value = true
  try {
    await secretApi.apply({ clusterId: clusterId.value, namespace: namespace.value, yaml: yamlContent.value })
    ElMessage.success('YAML应用成功')
    yamlDialogVisible.value = false
    fetchDetail()
  } catch {
    ElMessage.error('YAML应用失败')
  } finally {
    yamlApplyLoading.value = false
  }
}

const openDelete = () => { deleteDialogVisible.value = true }
const confirmDelete = async () => {
  deleteLoading.value = true
  try {
    await secretApi.delete({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })
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
  fetchDetail()
})
</script>


<style scoped>
.secret-detail {
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

.data-hint {
  font-size: 12px;
  color: #9ca3af;
  margin-bottom: 12px;
}

.labels-content,
.annotations-content,
.data-content {
  max-height: 320px;
  overflow-y: auto;
}

.empty-state {
  text-align: center;
  padding: 20px;
}

.labels-list,
.annotations-list,
.data-list {
  padding: 0;
}

.label-item,
.annotation-item,
.data-item {
  display: flex;
  flex-direction: column;
  padding: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  margin-bottom: 8px;
  background-color: #fafafa;
}

.label-key,
.annotation-key,
.data-key {
  font-weight: 500;
  color: #409eff;
  font-size: 12px;
  margin-bottom: 4px;
  word-break: break-all;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
}

.label-value,
.annotation-value,
.data-value {
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
