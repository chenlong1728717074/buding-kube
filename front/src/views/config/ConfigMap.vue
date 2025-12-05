<template>
  <div class="configmap-page">
    <div class="page-header">
      <h1>ConfigMap管理</h1>
      <div class="header-actions">
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>
          添加ConfigMap
        </el-button>
        <el-button type="success" @click="openYamlAdd">
          <el-icon><Document /></el-icon>
          YAML添加
        </el-button>
      </div>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="集群">
          <el-select v-model="searchForm.clusterId" placeholder="请选择集群" style="width: 200px" clearable @change="handleClusterChange">
            <el-option v-for="cluster in clusterList" :key="cluster.id" :label="cluster.name" :value="cluster.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="命名空间">
          <el-select v-model="searchForm.namespace" placeholder="请选择命名空间" style="width: 200px" clearable @change="handleNamespaceChange">
            <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table v-loading="loading" :data="configMaps" stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" header-align="center" align="center" />
        <el-table-column prop="name" label="名称" min-width="160" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="goDetail(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="140" header-align="center" align="center" />
        <el-table-column prop="alias" label="别名" min-width="120" :show-overflow-tooltip="true" header-align="center" align="center">
          <template #default="{ row }">
            <el-text type="info">{{ row.alias || '-' }}</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="creationTimestamp" label="创建时间" min-width="160" :show-overflow-tooltip="true" header-align="center" align="center">
          <template #default="{ row }">{{ formatDate(row.creationTimestamp) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="320" fixed="right" header-align="center" align="center">
          <template #default="{ row }">
            <div style="display: flex; gap: 6px; align-items: center; justify-content: center; flex-wrap: nowrap;">
              <el-button size="small" @click="goDetail(row)">详情</el-button>
              <el-dropdown @command="(cmd) => handleRowAction(cmd, row)">
                <el-button size="small">
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="editInfo">编辑信息</el-dropdown-item>
                    <el-dropdown-item command="editSetting">编辑设置</el-dropdown-item>
                    <el-dropdown-item command="yaml">YAML</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-button size="small" type="danger" @click="openDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    
    <!-- YAML添加 -->
    <el-dialog v-model="yamlAddDialogVisible" title="YAML添加" width="80%" :close-on-click-modal="false" destroy-on-close>
      <el-form :model="yamlAddForm" label-width="100px">
        <el-form-item label="集群">
          <el-select v-model="yamlAddForm.clusterId" placeholder="请选择集群" style="width: 240px">
            <el-option v-for="c in clusterList" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <div class="yaml-editor-wrapper">
        <YamlEditor :model-value="yamlAddContent" height="400px" @update:modelValue="val => yamlAddContent = val" />
      </div>
      <template #footer>
        <el-button @click="yamlAddDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="yamlAddLoading" @click="confirmYamlAdd">应用</el-button>
      </template>
    </el-dialog>

    <!-- 添加ConfigMap（表单） -->
    <el-dialog v-model="createDialogVisible" title="添加ConfigMap" width="600px" :close-on-click-modal="false">
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="集群">
          <el-select v-model="createForm.clusterId" placeholder="请选择集群" style="width: 240px">
            <el-option v-for="c in clusterList" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="命名空间">
          <el-select v-model="createForm.namespace" placeholder="请选择命名空间" style="width: 240px">
            <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="createForm.name" placeholder="请输入名称" />
        </el-form-item>
        <el-form-item label="别名">
          <el-input v-model="createForm.alias" placeholder="可选" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="createForm.describe" type="textarea" :rows="2" placeholder="可选" />
        </el-form-item>
      </el-form>
      <div class="kv-editor">
        <div class="kv-toolbar">
          <el-button size="small" type="primary" @click="addCreateDataRow">新增键值</el-button>
          <el-button size="small" @click="resetCreateDataRows">重置</el-button>
        </div>
        <el-table :data="createDataRows" size="small" style="width: 100%">
          <el-table-column label="键" width="220">
            <template #default="{ row }">
              <el-input v-model="row.key" placeholder="例如: app.properties" />
            </template>
          </el-table-column>
          <el-table-column label="值">
            <template #default="{ row }">
              <el-input v-model="row.value" type="textarea" :rows="3" placeholder="内容" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100" align="center">
            <template #default="{ $index }">
              <el-button type="danger" size="small" @click="removeCreateDataRow($index)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="createLoading" @click="confirmCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 修改别名/备注 -->
    <el-dialog v-model="editInfoDialogVisible" title="修改别名/备注" width="500px" :close-on-click-modal="false">
      <el-form :model="editInfoForm" label-width="100px">
        <el-form-item label="别名">
          <el-input v-model="editInfoForm.alias" placeholder="请输入别名" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="editInfoForm.describe" type="textarea" :rows="3" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editInfoDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editInfoLoading" @click="confirmEditInfo">确定</el-button>
      </template>
    </el-dialog>

    <!-- 修改数据（仅 data） -->
    <el-dialog v-model="editDataDialogVisible" title="修改数据" width="700px" :close-on-click-modal="false">
      <div class="kv-editor">
        <div class="kv-toolbar">
          <el-button size="small" type="primary" @click="addDataRow">新增键值</el-button>
          <el-button size="small" @click="resetDataRows">重置</el-button>
        </div>
        <el-table :data="dataRows" size="small" style="width: 100%">
          <el-table-column label="键" width="220">
            <template #default="{ row }">
              <el-input v-model="row.key" placeholder="例如: app.properties" />
            </template>
          </el-table-column>
          <el-table-column label="值">
            <template #default="{ row }">
              <el-input v-model="row.value" type="textarea" :rows="3" placeholder="内容" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100" align="center">
            <template #default="{ $index }">
              <el-button type="danger" size="small" @click="removeDataRow($index)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="editDataDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editDataLoading" @click="confirmEditData">保存</el-button>
      </template>
    </el-dialog>

    <!-- 查看/编辑YAML -->
    <el-dialog v-model="yamlDialogVisible" title="查看/编辑YAML" width="90%" :close-on-click-modal="false" destroy-on-close>
      <div class="yaml-editor-wrapper">
        <YamlEditor :model-value="yamlContent" :loading="yamlLoading" height="500px" @update:modelValue="val => yamlContent = val" />
      </div>
      <template #footer>
        <el-button @click="yamlDialogVisible = false">关闭</el-button>
        <template v-if="!yamlReadOnly">
          <el-button type="primary" :loading="yamlApplyLoading" @click="confirmApplyYaml">应用</el-button>
          <span style="color: #909399; font-size: 12px; margin-left: 10px;">仅应用到当前集群</span>
        </template>
      </template>
    </el-dialog>

    <!-- 删除确认对话框 -->
    <DeleteConfirmDialog
      v-model="deleteDialogVisible"
      :item-name="currentRow?.name || ''"
      message="确定要删除该ConfigMap吗？"
      :loading="deleteLoading"
      @confirm="confirmDelete"
      @cancel="() => deleteDialogVisible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh, ArrowDown, Plus, Document } from '@element-plus/icons-vue'
import { configMapApi, type ConfigMapVO, type ConfigMapPageQueryDTO } from '@/api/configmap'
import { clusterApi, type ClusterVO } from '@/api/cluster'
import { namespaceApi, type NamespaceVO } from '@/api/namespace'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import YamlEditor from '@/components/YamlEditor.vue'

const loading = ref(false)
const configMaps = ref<ConfigMapVO[]>([])
const selectedRows = ref<ConfigMapVO[]>([])

const searchForm = reactive<ConfigMapPageQueryDTO>({
  clusterId: '',
  namespace: '',
  page: 1,
  pageSize: 20,
  keyword: ''
})

const clusterList = ref<ClusterVO[]>([])
const namespaceList = ref<NamespaceVO[]>([])

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const currentRow = ref<ConfigMapVO | null>(null)
// 详情改为新页面，不再使用弹窗

const editInfoDialogVisible = ref(false)
const editInfoLoading = ref(false)
const editInfoForm = reactive({ alias: '', describe: '' })

const editDataDialogVisible = ref(false)
const editDataLoading = ref(false)
const dataRows = ref<{ key: string; value: string }[]>([])

const yamlDialogVisible = ref(false)
const yamlLoading = ref(false)
const yamlReadOnly = ref(false)
let yamlApplyLoading = ref(false)
let yamlContent = ref('')

const handleSelectionChange = (rows: ConfigMapVO[]) => { selectedRows.value = rows }

const fetchClusters = async () => {
  try {
    const resp = await clusterApi.getClusters({ page: 1, pageSize: 10000 })
    if ((resp as any)?.data?.items) {
      clusterList.value = (resp as any).data.items
      if (!searchForm.clusterId && clusterList.value.length > 0) {
        searchForm.clusterId = clusterList.value[0].id || clusterList.value[0].name
        await fetchNamespaces()
      }
    }
  } catch (e) {
    ElMessage.error('获取集群列表失败')
  }
}

const fetchNamespaces = async () => {
  if (!searchForm.clusterId) { namespaceList.value = []; return }
  try {
    const resp = await namespaceApi.getList({ clusterId: searchForm.clusterId, page: 1, pageSize: 10000 })
    if (resp.code === 200) {
      namespaceList.value = resp.data.items || []
    }
  } catch (e) {
    ElMessage.error('获取命名空间失败')
  }
}

const fetchList = async () => {
  if (!searchForm.clusterId) { configMaps.value = []; pagination.total = 0; return }
  try {
    loading.value = true
    const resp = await configMapApi.getList({ clusterId: searchForm.clusterId, namespace: searchForm.namespace || '', page: pagination.page, pageSize: pagination.pageSize, keyword: searchForm.keyword || '' })
    if (resp.code === 200) {
      configMaps.value = resp.data.items || []
      pagination.total = resp.data.total || 0
    }
  } catch (e) {
    ElMessage.error('获取ConfigMap列表失败')
  } finally {
    loading.value = false
  }
}

const handleClusterChange = async () => {
  searchForm.namespace = ''
  pagination.page = 1
  await fetchNamespaces()
  configMaps.value = []
  pagination.total = 0
}

const handleNamespaceChange = () => {
  pagination.page = 1
  fetchList()
}

const handleSearch = () => { pagination.page = 1; fetchList() }
const handleReset = async () => {
  searchForm.namespace = ''
  searchForm.keyword = ''
  pagination.page = 1
  await fetchNamespaces()
  fetchList()
}
// 刷新通过“搜索”进行，不再提供单独刷新按钮

const handleSizeChange = (size: number) => { pagination.pageSize = size; pagination.page = 1; fetchList() }
const handleCurrentChange = (page: number) => { pagination.page = page; fetchList() }

const router = useRouter()
const goDetail = (row: ConfigMapVO) => {
  const clusterId = searchForm.clusterId
  router.push({ path: '/config/configmap/detail', query: { clusterId, namespace: row.namespace, name: row.name } })
}

const handleRowAction = (cmd: string, row: ConfigMapVO) => {
  currentRow.value = row
  switch (cmd) {
    case 'editInfo':
      editInfoForm.alias = row.alias || ''
      editInfoForm.describe = row.describe || ''
      editInfoDialogVisible.value = true
      break
    case 'editData':
      dataRows.value = Object.entries(row.data || {}).map(([k, v]) => ({ key: k, value: v }))
      editDataDialogVisible.value = true
      break
    case 'yaml':
      yamlLoading.value = true
      yamlDialogVisible.value = true
      yamlReadOnly.value = true
      yamlContent.value = row.yaml || ''
      yamlLoading.value = false
      break
    case 'editSetting':
      yamlLoading.value = true
      yamlDialogVisible.value = true
      yamlReadOnly.value = false
      yamlContent.value = row.yaml || ''
      yamlLoading.value = false
      break
  }
}

const openDelete = (row: ConfigMapVO) => {
  currentRow.value = row
  deleteDialogVisible.value = true
}

const confirmEditInfo = async () => {
  if (!currentRow.value) return
  editInfoLoading.value = true
  try {
    const payload = { clusterId: searchForm.clusterId, namespace: currentRow.value.namespace, name: currentRow.value.name, alias: editInfoForm.alias, describe: editInfoForm.describe }
    await configMapApi.updateInfo(payload)
    ElMessage.success('更新成功')
    editInfoDialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error('更新失败')
  } finally {
    editInfoLoading.value = false
  }
}

const addDataRow = () => { dataRows.value.push({ key: '', value: '' }) }
const removeDataRow = (index: number) => { dataRows.value.splice(index, 1) }
const resetDataRows = () => { dataRows.value = Object.entries(currentRow.value?.data || {}).map(([k, v]) => ({ key: k, value: v })) }

const confirmEditData = async () => {
  if (!currentRow.value) return
  editDataLoading.value = true
  try {
    const data: Record<string, string> = {}
    for (const { key, value } of dataRows.value) {
      if (!key) continue
      data[key] = value
    }
    const payload = { clusterId: searchForm.clusterId, namespace: currentRow.value.namespace, name: currentRow.value.name, data }
    await configMapApi.updateData(payload)
    ElMessage.success('数据已保存')
    editDataDialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error('保存失败')
  } finally {
    editDataLoading.value = false
  }
}

const confirmApplyYaml = async () => {
  if (!yamlContent.value || !searchForm.clusterId) { ElMessage.warning('YAML内容为空'); return }
  yamlApplyLoading.value = true
  try {
    await configMapApi.applyYaml({ clusterId: searchForm.clusterId, yaml: yamlContent.value })
    ElMessage.success('YAML应用成功')
    yamlDialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error('YAML应用失败')
  } finally {
    yamlApplyLoading.value = false
  }
}

// YAML添加
const yamlAddDialogVisible = ref(false)
const yamlAddLoading = ref(false)
const yamlAddForm = reactive({ clusterId: '' as string })
let yamlAddContent = ref([
  'apiVersion: v1',
  'kind: ConfigMap',
  'metadata:',
  '  name: example-config',
  '  namespace: default',
  'data:',
  '  app.properties: |',
  '    key=value'
].join('\n'))
const openYamlAdd = () => {
  yamlAddForm.clusterId = searchForm.clusterId || ''
  yamlAddDialogVisible.value = true
}
const confirmYamlAdd = async () => {
  if (!yamlAddForm.clusterId || !yamlAddContent.value) { ElMessage.warning('请选择集群并填写YAML'); return }
  yamlAddLoading.value = true
  try {
    await configMapApi.applyYaml({ clusterId: yamlAddForm.clusterId, yaml: yamlAddContent.value })
    ElMessage.success('YAML添加成功')
    yamlAddDialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error('YAML添加失败')
  } finally {
    yamlAddLoading.value = false
  }
}

// 添加ConfigMap（表单生成YAML应用）
const createDialogVisible = ref(false)
const createLoading = ref(false)
const createForm = reactive({ clusterId: '' as string, namespace: '' as string, name: '' as string, alias: '' as string, describe: '' as string })
const createDataRows = ref<{ key: string; value: string }[]>([])
const openCreate = () => {
  createForm.clusterId = searchForm.clusterId || ''
  createForm.namespace = searchForm.namespace || ''
  createForm.name = ''
  createForm.alias = ''
  createForm.describe = ''
  createDataRows.value = []
  createDialogVisible.value = true
}
const addCreateDataRow = () => { createDataRows.value.push({ key: '', value: '' }) }
const removeCreateDataRow = (idx: number) => { createDataRows.value.splice(idx, 1) }
const resetCreateDataRows = () => { createDataRows.value = [] }
const confirmCreate = async () => {
  if (!createForm.clusterId || !createForm.namespace || !createForm.name) { ElMessage.warning('请填写集群、命名空间与名称'); return }
  const ann: Record<string, string> = {}
  if (createForm.alias) ann['alias'] = createForm.alias
  if (createForm.describe) ann['describe'] = createForm.describe
  const data: Record<string, string> = {}
  for (const { key, value } of createDataRows.value) { if (key) data[key] = value }
  const yaml = `apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: ${createForm.name}\n  namespace: ${createForm.namespace}${Object.keys(ann).length ? '\n  annotations:' : ''}${Object.entries(ann).map(([k,v]) => '\n    ' + k + ': ' + v).join('')}\n${Object.keys(data).length ? 'data:' : ''}${Object.entries(data).map(([k,v]) => '\n  ' + k + ': |\n' + v.split('\n').map(l => '    ' + l).join('\n')).join('')}\n`
  createLoading.value = true
  try {
    await configMapApi.applyYaml({ clusterId: createForm.clusterId, yaml })
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error('创建失败')
  } finally {
    createLoading.value = false
  }
}

const deleteDialogVisible = ref(false)
const deleteLoading = ref(false)
const confirmDelete = async () => {
  if (!currentRow.value) return
  deleteLoading.value = true
  try {
    await configMapApi.delete({ clusterId: searchForm.clusterId, namespace: currentRow.value.namespace, name: currentRow.value.name })
    ElMessage.success('删除成功')
    deleteDialogVisible.value = false
    fetchList()
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

onMounted(async () => {
  await fetchClusters()
  await fetchList()
})
</script>

<style scoped>
.configmap-page {
  padding: 20px;
}
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
.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}
.header-actions { display: flex; gap: 12px; }
.header-actions :deep(.el-button) { border-radius: 20px; }
.search-card { margin-bottom: 20px; }
.table-card { margin-bottom: 20px; }
.pagination-wrapper { display: flex; justify-content: flex-end; margin-top: 20px; }
.kv-toolbar { display: flex; gap: 8px; margin-bottom: 8px; }

/* 下拉菜单圆润样式 */
:deep(.el-dropdown-menu) { border-radius: 10px; padding: 6px; }
:deep(.el-dropdown-menu__item) { border-radius: 12px; margin: 2px 4px; }
:deep(.el-dropdown-menu__item:hover) { background-color: #eef2ff; color: #3b82f6; }
</style>
