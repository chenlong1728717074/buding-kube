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

    <!-- 编辑信息 -->
    <el-dialog v-model="editInfoDialogVisible" title="编辑信息" width="900px" :close-on-click-modal="false" class="config-dialog">
      <template #header>
        <div class="dialog-header">
          <h3 class="dialog-title">编辑信息</h3>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <el-form :model="editInfoForm" label-width="100px">
            <el-form-item label="别名"><el-input v-model="editInfoForm.alias" /></el-form-item>
            <el-form-item label="备注"><el-input v-model="editInfoForm.describe" type="textarea" :rows="3" /></el-form-item>
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

    <!-- 编辑设置（data） -->
    <el-dialog v-model="editDataDialogVisible" title="编辑设置" width="80%" :close-on-click-modal="false" class="config-dialog">
      <template #header>
        <div class="dialog-header">
          <div>
            <h3 class="dialog-title">编辑设置</h3>
            <div style="margin-top:4px;color:#6b7280;font-size:12px;">仅修改 data 键值</div>
          </div>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <div class="kv-editor">
            <div class="kv-toolbar">
              <el-button size="small" type="primary" @click="addDataRow">新增键值</el-button>
              <el-button size="small" @click="resetDataRows">重置</el-button>
            </div>
            <el-table :data="editRows" size="small" style="width: 100%">
              <el-table-column label="键" width="220">
                <template #default="{ row }"><el-input v-model="row.key" /></template>
              </el-table-column>
              <el-table-column label="值">
                <template #default="{ row }"><el-input v-model="row.value" type="textarea" :rows="3" /></template>
              </el-table-column>
              <el-table-column label="操作" width="100" align="center">
                <template #default="{ $index }"><el-button type="danger" size="small" @click="removeDataRow($index)">删除</el-button></template>
              </el-table-column>
            </el-table>
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
import '@/assets/styles/config-editor.css'

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
const editRows = ref<{ key:string; value:string }[]>([])

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
      editRows.value = Object.entries(cm.value?.data || {}).map(([k,v]) => ({ key: k, value: v }))
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

const addDataRow = () => { editRows.value.push({ key: '', value: '' }) }
const removeDataRow = (idx: number) => { editRows.value.splice(idx, 1) }
const resetDataRows = () => { editRows.value = Object.entries(cm.value?.data || {}).map(([k,v]) => ({ key: k, value: v })) }
const confirmEditData = async () => {
  if (!cm.value) return
  editDataLoading.value = true
  try {
    const data: Record<string, string> = {}
    for (const { key, value } of editRows.value) { if (key) data[key] = value }
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
