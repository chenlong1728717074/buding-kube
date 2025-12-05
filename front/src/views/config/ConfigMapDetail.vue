<template>
  <div class="configmap-detail">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="handleBack" style="margin-right: 16px;">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1>ConfigMap详情</h1>
      </div>
      <div class="header-right">
        <el-button size="small" @click="scrollToDetail">详情</el-button>
        <el-dropdown @command="handleMore">
          <el-button size="small">
            更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="editInfo">编辑信息</el-dropdown-item>
              <el-dropdown-item command="yaml">查看/编辑YAML</el-dropdown-item>
              <el-dropdown-item command="editData">编辑设置</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button size="small" type="danger" @click="openDelete">删除</el-button>
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="cm" class="detail-content" ref="detailRef">
      <el-card class="info-card">
        <template #header>
          <span>基本信息</span>
        </template>
        <div class="info-grid">
          <div class="info-item"><label>集群ID:</label><span>{{ clusterId }}</span></div>
          <div class="info-item"><label>命名空间:</label><span>{{ cm.namespace }}</span></div>
          <div class="info-item"><label>名称:</label><span>{{ cm.name }}</span></div>
          <div class="info-item"><label>别名:</label><span>{{ cm.alias || '-' }}</span></div>
          <div class="info-item"><label>备注:</label><span>{{ cm.describe || '-' }}</span></div>
          <div class="info-item"><label>创建时间:</label><span>{{ formatDate(cm.creationTimestamp) }}</span></div>
        </div>
      </el-card>

      <el-card class="info-card">
        <template #header>
          <span>数据</span>
        </template>
        <el-table :data="dataRows" size="small" style="width:100%">
          <el-table-column prop="key" label="键" width="220" />
          <el-table-column prop="value" label="值" />
        </el-table>
      </el-card>
    </div>

    <div v-else class="empty-container">
      <el-empty description="未找到ConfigMap" />
    </div>

    <!-- 编辑信息 -->
    <el-dialog v-model="editInfoDialogVisible" title="编辑信息" width="500px" :close-on-click-modal="false">
      <el-form :model="editInfoForm" label-width="100px">
        <el-form-item label="别名"><el-input v-model="editInfoForm.alias" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="editInfoForm.describe" type="textarea" :rows="3" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editInfoDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editInfoLoading" @click="confirmEditInfo">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑设置（data） -->
    <el-dialog v-model="editDataDialogVisible" title="编辑设置" width="700px" :close-on-click-modal="false">
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
      <template #footer>
        <el-button @click="editDataDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editDataLoading" @click="confirmEditData">保存</el-button>
      </template>
    </el-dialog>

    <!-- 查看/编辑YAML（修改时不允许选择集群） -->
    <el-dialog v-model="yamlDialogVisible" title="查看/编辑YAML" width="90%" :close-on-click-modal="false" destroy-on-close>
      <div class="yaml-editor-wrapper">
        <YamlEditor :model-value="yamlContent" height="500px" @update:modelValue="val => yamlContent = val" />
      </div>
      <template #footer>
        <el-button @click="yamlDialogVisible = false">关闭</el-button>
        <el-button type="primary" :loading="yamlApplyLoading" @click="confirmApplyYaml">应用</el-button>
        <span style="color:#909399;font-size:12px;margin-left:10px;">应用到当前集群：{{ clusterId }}</span>
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
import { ArrowLeft, ArrowDown } from '@element-plus/icons-vue'
import { configMapApi, type ConfigMapVO } from '@/api/configmap'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const clusterId = ref('')
const namespace = ref('')
const name = ref('')
const loading = ref(false)
const cm = ref<ConfigMapVO | null>(null)
const dataRows = ref<{ key:string; value:string }[]>([])
const detailRef = ref()

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
const scrollToDetail = () => { try { (detailRef.value as HTMLElement)?.scrollIntoView({ behavior: 'smooth' }) } catch {} }

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
    await configMapApi.applyYaml({ clusterId: clusterId.value, yaml: yamlContent.value })
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
.configmap-detail { padding: 20px; }
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
.header-left { display:flex; align-items:center; gap:16px; }
.page-header h1 { margin:0; font-size:24px; font-weight:600; color:#2c3e50; }
.header-right { display:flex; gap:8px; }
.loading-container { padding: 20px; }
.detail-content { display:flex; flex-direction:column; gap:16px; }
.info-card { border-radius:8px; }
.info-grid { display:grid; grid-template-columns: repeat(2, 1fr); gap:12px; }
.info-item label { color:#64748b; margin-right:8px; }
</style>
