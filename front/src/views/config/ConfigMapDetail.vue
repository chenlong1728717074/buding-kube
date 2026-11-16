<template>
  <div class="detail-page">
    <!-- 头部卡片 -->
    <el-card class="ks-header-card" shadow="never">
      <div class="ks-breadcrumb-row">
        <el-link :underline="false" class="ks-back" @click="handleBack">
          <el-icon><ArrowLeft /></el-icon>
          <span>ConfigMap</span>
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
              <el-dropdown-item command="editData"><el-icon><Setting /></el-icon>编辑设置</el-dropdown-item>
              <el-dropdown-item command="delete" divided><el-icon><Delete /></el-icon>删除</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
      <div class="ks-title-row">
        <div class="ks-icon-block"><el-icon :size="24"><Document /></el-icon></div>
        <div class="ks-title-main">
          <div class="ks-name">{{ cm?.name || name }}</div>
          <div class="ks-desc">{{ cm?.describe || '-' }}</div>
        </div>
      </div>
    </el-card>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="8" animated />
    </div>

    <template v-else-if="cm">
      <!-- 属性 -->
      <el-card class="ks-card" shadow="never">
        <template #header>
          <div class="ks-card-header"><span>属性</span></div>
        </template>
        <div class="props-grid">
          <div class="prop-item"><div class="prop-label">集群</div><div class="prop-value">{{ clusterId }}</div></div>
          <div class="prop-item"><div class="prop-label">命名空间</div><div class="prop-value">{{ cm.namespace }}</div></div>
          <div class="prop-item"><div class="prop-label">别名</div><div class="prop-value">{{ cm.alias || '-' }}</div></div>
          <div class="prop-item"><div class="prop-label">创建时间</div><div class="prop-value">{{ formatDate(cm.creationTimestamp) }}</div></div>
          <div class="prop-item"><div class="prop-label">备注</div><div class="prop-value">{{ cm.describe || '-' }}</div></div>
        </div>
      </el-card>

      <!-- 数据 -->
      <el-card class="ks-card" shadow="never">
        <template #header><div class="ks-card-header"><span>数据 ({{ dataRows.length }})</span></div></template>
        <div v-if="dataRows.length" class="ks-data-list">
          <div v-for="row in dataRows" :key="row.key" class="ks-data-row" @click="toggleExpand(row.key)">
            <span class="ks-data-key">{{ row.key }}</span>
            <span class="ks-data-value" :class="{ expanded: !!expandedKeys[row.key] }">{{ expandedKeys[row.key] ? row.value : truncateValue(row.value) }}</span>
          </div>
        </div>
        <el-empty v-else description="暂无数据" :image-size="60" />
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
      <el-empty description="未找到ConfigMap" />
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
            <el-form :model="editInfoForm" label-width="100px">
              <el-form-item label="名称">
                <el-input :model-value="cm?.name || ''" disabled />
              </el-form-item>
              <el-form-item label="别名">
                <el-input v-model="editInfoForm.alias" placeholder="请输入别名" />
                <div class="helper">别名只能包含中文、字母、数字和连字符（-），不得以连字符（-）开头或结尾，最长 63 个字符。</div>
              </el-form-item>
              <el-form-item label="备注">
                <el-input v-model="editInfoForm.describe" type="textarea" :rows="3" placeholder="请输入备注" />
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
            <span class="resource-info-value">{{ cm?.namespace || '-' }}</span>
          </div>
          <div class="resource-info-item">
            <span class="resource-info-label">名称:</span>
            <span class="resource-info-value">{{ cm?.name || '-' }}</span>
          </div>
        </div>

        <div class="config-content">
          <div class="data-form">
            <div class="form-section">
              <div class="section-header-row">
                <div class="section-title">数据 *</div>
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

    <!-- 查看/编辑YAML（修改时不允许选择集群） -->
    <el-dialog v-model="yamlDialogVisible" title="查看/编辑YAML" width="90%" :close-on-click-modal="false" class="config-dialog yaml-dialog">
      <template #header>
        <div class="dialog-header">
          <h3 class="dialog-title">查看/编辑YAML</h3>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <div class="yaml-editor-wrapper">
            <YamlEditor :model-value="yamlContent" :readonly="false" height="100%" @update:modelValue="val => yamlContent = val" />
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
      :item-name="cm?.name || ''"
      message="确定要删除该ConfigMap吗？"
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
import { ArrowLeft, ArrowUp, EditPen, Document, Setting, Delete } from '@element-plus/icons-vue'
import { configMapApi, type ConfigMapVO } from '@/api/configmap'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import KVEditorPane from '@/components/KVEditorPane.vue'
import '@/assets/styles/config-editor.css'

type KvRow = { key: string; value: string }

const route = useRoute()
const router = useRouter()
const clusterId = ref('')
const namespace = ref('')
const name = ref('')
const loading = ref(false)
const cm = ref<ConfigMapVO | null>(null)
const dataRows = ref<{ key:string; value:string }[]>([])
const labelRows = ref<{ key:string; value:string }[]>([])
const annotationRows = ref<{ key:string; value:string }[]>([])

const editInfoDialogVisible = ref(false)
const editInfoLoading = ref(false)
const editInfoForm = reactive({ alias: '', describe: '' })

const editDataDialogVisible = ref(false)
const editDataLoading = ref(false)
const editDataRows = ref<KvRow[]>([])
const kvEditRef = ref()
const editImportInputRef = ref()

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
  editDataRows.value = Object.entries(cm.value?.data || {}).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
}

const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlApplyLoading = ref(false)

const deleteDialogVisible = ref(false)
const deleteLoading = ref(false)

const handleBack = () => router.back()

// 数据值展开/收起
const expandedKeys = ref<Record<string, boolean>>({})
const toggleExpand = (key: string) => {
  expandedKeys.value[key] = !expandedKeys.value[key]
}
const truncateValue = (value: string) => {
  const flat = String(value ?? '').replace(/\s+/g, ' ').trim()
  return flat.length > 120 ? flat.slice(0, 120) + '…' : flat
}

const handleMore = (cmd: string) => {
  switch (cmd) {
    case 'editInfo':
      if (!cm.value) return
      editInfoForm.alias = cm.value.alias || ''
      editInfoForm.describe = cm.value.describe || ''
      editInfoDialogVisible.value = true
      break
    case 'yaml':
      yamlContent.value = cm.value?.yaml || ''
      yamlDialogVisible.value = true
      break
    case 'editData':
      editDataRows.value = Object.entries(cm.value?.data || {}).map(([k,v]) => ({ key: k, value: v }))
      editDataDialogVisible.value = true
      break
    case 'delete':
      openDelete()
      break
  }
}

const fetchDetailFromList = async () => {
  if (!clusterId.value) return
  loading.value = true
  try {
    const resp = await configMapApi.getList({ clusterId: clusterId.value, namespace: namespace.value, page: 1, pageSize: 10000 })
    if (resp.code === 200) {
      const items = resp.data.items || []
      const found = items.find((it: ConfigMapVO) => it.name === name.value)
      cm.value = found || null
      dataRows.value = Object.entries(found?.data || {}).map(([k,v]) => ({ key: k, value: v }))
      labelRows.value = Object.entries(found?.labels || {}).map(([k,v]) => ({ key: k, value: String(v ?? '') }))
      annotationRows.value = Object.entries(found?.annotations || {})
        .filter(([k]) => k !== 'alias' && k !== 'describe')
        .map(([k,v]) => ({ key: k, value: String(v ?? '') }))
    }
  } catch (e) {
    ElMessage.error('获取详情失败')
  } finally {
    loading.value = false
  }
}

const confirmEditInfo = async () => {
  if (!cm.value) return
  editInfoLoading.value = true
  try {
    await configMapApi.updateInfo({ clusterId: clusterId.value, namespace: namespace.value, name: name.value, alias: editInfoForm.alias, describe: editInfoForm.describe })
    ElMessage.success('更新成功')
    editInfoDialogVisible.value = false
    fetchDetailFromList()
  } catch (e) {
    ElMessage.error('更新失败')
  } finally {
    editInfoLoading.value = false
  }
}

const confirmEditData = async () => {
  if (!cm.value) return
  editDataLoading.value = true
  try {
    const data = collectRowsData(editDataRows.value)
    await configMapApi.updateData({ clusterId: clusterId.value, namespace: namespace.value, name: name.value, data })
    ElMessage.success('保存成功')
    editDataDialogVisible.value = false
    fetchDetailFromList()
  } catch (e) {
    ElMessage.error('保存失败')
  } finally {
    editDataLoading.value = false
  }
}

const confirmApplyYaml = async () => {
  if (!yamlContent.value) { ElMessage.warning('YAML内容为空'); return }
  yamlApplyLoading.value = true
  try {
    await configMapApi.applyYaml({ clusterId: clusterId.value, namespace: namespace.value, yaml: yamlContent.value })
    ElMessage.success('YAML应用成功')
    yamlDialogVisible.value = false
    fetchDetailFromList()
  } catch (e) {
    ElMessage.error('YAML应用失败')
  } finally {
    yamlApplyLoading.value = false
  }
}

const openDelete = () => { deleteDialogVisible.value = true }
const confirmDelete = async () => {
  deleteLoading.value = true
  try {
    await configMapApi.delete({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })
    ElMessage.success('删除成功')
    deleteDialogVisible.value = false
    router.back()
  } catch (e) {
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
    return d.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
  } catch { return '-' }
}

onMounted(() => {
  clusterId.value = route.query.clusterId as string || ''
  namespace.value = route.query.namespace as string || ''
  name.value = route.query.name as string || ''
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

/* 头部卡片 */
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

/* 卡片与属性网格 */
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

/* 数据行(KubeSphere 胶囊样式) */
.ks-data-row {
  display: flex;
  align-items: center;
  gap: 16px;
  background: #f7f8fa;
  border-radius: 24px;
  padding: 12px 24px;
  margin-bottom: 12px;
  cursor: pointer;
  transition: background 0.2s ease;
}

.ks-data-row:hover {
  background: #f0f2f5;
}

.ks-data-key {
  width: 220px;
  flex-shrink: 0;
  color: #6b7280;
  font-size: 13px;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ks-data-value {
  flex: 1;
  min-width: 0;
  color: #1f2937;
  font-size: 13px;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: right;
}

.ks-data-value.expanded {
  white-space: pre-wrap;
  overflow: auto;
  max-height: 300px;
  text-align: left;
}

/* 标签/注解 芯片 */
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
