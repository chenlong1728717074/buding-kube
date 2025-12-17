<template>
  <div class="sa-page">
    <div class="page-header">
      <h1>ServiceAccount管理</h1>
      <div class="header-actions">
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>
          添加ServiceAccount
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
      <el-table v-loading="loading" :data="serviceAccounts" stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" header-align="center" align="center" />
        <el-table-column prop="name" label="名称" min-width="160" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="goDetail(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="140" header-align="center" align="center" />
        <el-table-column prop="automountServiceAccountToken" label="自动挂载Token" width="140" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="row.automountServiceAccountToken === false ? 'info' : 'success'">
              {{ row.automountServiceAccountToken === false ? '否' : '是' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="creationTimestamp" label="创建时间" min-width="180" :show-overflow-tooltip="true" header-align="center" align="center">
          <template #default="{ row }">{{ formatDate(row.creationTimestamp) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right" header-align="center" align="center">
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
                    <el-dropdown-item command="yaml">YAML</el-dropdown-item>
                    <el-dropdown-item command="editSetting">编辑设置</el-dropdown-item>
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

    <UnifiedDialog v-model="yamlAddDialogVisible" title="YAML添加" subtitle="通过 YAML 快速创建 ServiceAccount" width="80%">
      <el-form :model="yamlAddForm" label-width="100px">
        <el-form-item label="集群">
          <el-select v-model="yamlAddForm.clusterId" placeholder="请选择集群" style="width: 240px">
            <el-option v-for="c in clusterList" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="命名空间">
          <el-select v-model="yamlAddForm.namespace" placeholder="请选择命名空间" style="width: 240px">
            <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </el-form-item>
      </el-form>
      <div class="yaml-editor-wrapper">
        <YamlEditor :model-value="yamlAddContent" :readonly="false" height="400px" @update:modelValue="val => yamlAddContent = val" />
      </div>
      <template #footer>
        <el-button @click="yamlAddDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="yamlAddLoading" @click="confirmYamlAdd">应用</el-button>
      </template>
    </UnifiedDialog>

    <StepWizardDialog
      v-model="createDialogVisible"
      title="创建ServiceAccount"
      width="80%"
      :steps="createSteps"
      :active="createActive"
      :yaml="createYaml"
      :loading="createLoading"
      confirm-text="创建"
      @update:active="val => createActive = val"
      @update:yaml="val => createYaml = val"
      @confirm="confirmCreate"
    >
      <div v-if="createActive === 0">
        <el-form :model="createForm" label-width="120px" class="two-col">
          <el-form-item label="名称" required>
            <el-input v-model="createForm.name" />
          </el-form-item>
          <el-form-item label="集群" required>
            <el-select v-model="createForm.clusterId" placeholder="请选择集群" style="width: 240px">
              <el-option v-for="c in clusterList" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="命名空间" required>
            <el-select v-model="createForm.namespace" placeholder="请选择命名空间" style="width: 240px">
              <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
            </el-select>
          </el-form-item>
          <el-form-item label="自动挂载Token">
            <el-switch v-model="createForm.automountServiceAccountToken" />
          </el-form-item>
        </el-form>
      </div>
      <div v-else>
        <div class="group-box">
          <div class="group-title">关联设置</div>
          <el-form :model="createForm" label-width="120px">
            <el-form-item label="镜像拉取Secrets">
              <el-select v-model="createForm.imagePullSecrets" multiple filterable allow-create default-first-option style="width: 100%" placeholder="输入或选择Secret名称" />
            </el-form-item>
            <el-form-item label="挂载Secrets">
              <el-select v-model="createForm.secrets" multiple filterable allow-create default-first-option style="width: 100%" placeholder="输入或选择Secret名称" />
            </el-form-item>
          </el-form>
        </div>
      </div>
    </StepWizardDialog>

    <UnifiedDialog v-model="editInfoDialogVisible" title="编辑信息" subtitle="修改标签与注解" width="80%">
      <div class="group-box" style="margin-bottom: 12px;">
        <div class="group-title">基本信息</div>
        <el-form :model="editInfoForm" label-width="100px">
          <el-form-item label="名称">
            <el-input :model-value="currentRow?.name || ''" disabled />
          </el-form-item>
          <el-form-item label="标签(JSON)">
            <el-input v-model="editInfoForm.labelsJson" type="textarea" :rows="3" />
          </el-form-item>
          <el-form-item label="注解(JSON)">
            <el-input v-model="editInfoForm.annotationsJson" type="textarea" :rows="3" />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="editInfoDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editInfoLoading" @click="confirmEditInfo">确定</el-button>
      </template>
    </UnifiedDialog>

    <UnifiedDialog v-model="yamlDialogVisible" title="查看/编辑YAML" subtitle="仅应用到当前集群" width="90%">
      <div class="yaml-editor-wrapper">
        <YamlEditor :model-value="yamlContent" :readonly="yamlReadOnly" :loading="yamlLoading" height="500px" @update:modelValue="val => yamlContent = val" />
      </div>
      <template #footer>
        <el-button @click="yamlDialogVisible = false">关闭</el-button>
        <template v-if="!yamlReadOnly">
          <el-button type="primary" :loading="yamlApplyLoading" @click="confirmApplyYaml">应用</el-button>
        </template>
      </template>
    </UnifiedDialog>

    <DeleteConfirmDialog
      v-model="deleteDialogVisible"
      :item-name="currentRow?.name || ''"
      message="确定要删除该ServiceAccount吗？"
      :loading="deleteLoading"
      @confirm="confirmDelete"
      @cancel="() => deleteDialogVisible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh, ArrowDown, Plus, Document } from '@element-plus/icons-vue'
import { serviceAccountApi, type ServiceAccountVO, type ServiceAccountPageQueryDTO } from '@/api/serviceaccount'
import { clusterApi, type ClusterVO } from '@/api/cluster'
import { namespaceApi, type NamespaceVO } from '@/api/namespace'
import UnifiedDialog from '@/components/UnifiedDialog.vue'
import StepWizardDialog from '@/components/StepWizardDialog.vue'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import YamlEditor from '@/components/YamlEditor.vue'

const router = useRouter()

const loading = ref(false)
const serviceAccounts = ref<ServiceAccountVO[]>([])
const selectedRows = ref<ServiceAccountVO[]>([])

const searchForm = reactive<ServiceAccountPageQueryDTO>({
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

const goDetail = (row: ServiceAccountVO) => {
  const clusterId = searchForm.clusterId
  router.push({ path: '/config/serviceaccount/detail', query: { clusterId, namespace: row.namespace, name: row.name } })
}

const currentRow = ref<ServiceAccountVO | null>(null)

const editInfoDialogVisible = ref(false)
const editInfoLoading = ref(false)
const editInfoForm = reactive({
  labelsJson: '',
  annotationsJson: ''
})

const yamlDialogVisible = ref(false)
const yamlLoading = ref(false)
const yamlReadOnly = ref(false)
let yamlApplyLoading = ref(false)
let yamlContent = ref('')

const handleSelectionChange = (rows: ServiceAccountVO[]) => {
  selectedRows.value = rows
}

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
  } catch {
    ElMessage.error('获取集群列表失败')
  }
}

const fetchNamespaces = async () => {
  if (!searchForm.clusterId) {
    namespaceList.value = []
    return
  }
  try {
    const resp = await namespaceApi.getList({ clusterId: searchForm.clusterId, page: 1, pageSize: 10000 })
    if (resp.code === 200) {
      namespaceList.value = resp.data.items || []
    }
  } catch {
    ElMessage.error('获取命名空间失败')
  }
}

const fetchList = async () => {
  if (!searchForm.clusterId) {
    serviceAccounts.value = []
    pagination.total = 0
    return
  }
  try {
    loading.value = true
    const resp = await serviceAccountApi.getList({
      clusterId: searchForm.clusterId,
      namespace: searchForm.namespace || '',
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword || ''
    })
    if (resp.code === 200) {
      const items = (resp.data as any).items || (resp.data as any).list || []
      const total = (resp.data as any).total || (resp.data as any).totalCount || 0
      serviceAccounts.value = items
      pagination.total = total
    }
  } catch {
    ElMessage.error('获取ServiceAccount列表失败')
  } finally {
    loading.value = false
  }
}

const handleClusterChange = async () => {
  searchForm.namespace = ''
  pagination.page = 1
  await fetchNamespaces()
  serviceAccounts.value = []
  pagination.total = 0
}

const handleNamespaceChange = () => {
  pagination.page = 1
  fetchList()
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = async () => {
  searchForm.namespace = ''
  searchForm.keyword = ''
  pagination.page = 1
  await fetchNamespaces()
  fetchList()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchList()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchList()
}

const handleRowAction = (cmd: string, row: ServiceAccountVO) => {
  currentRow.value = row
  switch (cmd) {
    case 'editInfo':
      editInfoForm.labelsJson = JSON.stringify(row.labels || {}, null, 2)
      editInfoForm.annotationsJson = JSON.stringify(row.annotations || {}, null, 2)
      editInfoDialogVisible.value = true
      break
    case 'yaml':
      yamlLoading.value = true
      yamlDialogVisible.value = true
      yamlReadOnly.value = false
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

const openDelete = (row: ServiceAccountVO) => {
  currentRow.value = row
  deleteDialogVisible.value = true
}

const confirmEditInfo = async () => {
  if (!currentRow.value) return
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
      clusterId: searchForm.clusterId,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      automountServiceAccountToken: currentRow.value.automountServiceAccountToken,
      imagePullSecrets: (currentRow.value.imagePullSecrets || []).map(s => s.name),
      secrets: (currentRow.value.secrets || []).map(s => s.name),
      labels,
      annotations
    })
    ElMessage.success('更新成功')
    editInfoDialogVisible.value = false
    fetchList()
  } catch {
    ElMessage.error('更新失败，请检查JSON格式')
  } finally {
    editInfoLoading.value = false
  }
}

const confirmApplyYaml = async () => {
  if (!yamlContent.value || !searchForm.clusterId) {
    ElMessage.warning('YAML内容为空')
    return
  }
  if (!currentRow.value?.namespace) {
    ElMessage.warning('命名空间为空')
    return
  }
  yamlApplyLoading.value = true
  try {
    await serviceAccountApi.applyYaml({
      clusterId: searchForm.clusterId,
      namespace: currentRow.value.namespace,
      yaml: yamlContent.value
    })
    ElMessage.success('YAML应用成功')
    yamlDialogVisible.value = false
    fetchList()
  } catch {
    ElMessage.error('YAML应用失败')
  } finally {
    yamlApplyLoading.value = false
  }
}

const yamlAddDialogVisible = ref(false)
const yamlAddLoading = ref(false)
const yamlAddForm = reactive({ clusterId: '' as string, namespace: '' as string })
let yamlAddContent = ref(
  [
    'apiVersion: v1',
    'kind: ServiceAccount',
    'metadata:',
    '  name: example-sa',
    '  namespace: default'
  ].join('\n')
)

const openYamlAdd = () => {
  yamlAddForm.clusterId = searchForm.clusterId || ''
  yamlAddForm.namespace = searchForm.namespace || ''
  yamlAddDialogVisible.value = true
}

const confirmYamlAdd = async () => {
  if (!yamlAddForm.clusterId || !yamlAddContent.value) {
    ElMessage.warning('请选择集群并填写YAML')
    return
  }
  yamlAddLoading.value = true
  try {
    await serviceAccountApi.applyYaml({
      clusterId: yamlAddForm.clusterId,
      namespace: yamlAddForm.namespace || '',
      yaml: yamlAddContent.value
    })
    ElMessage.success('YAML添加成功')
    yamlAddDialogVisible.value = false
    fetchList()
  } catch {
    ElMessage.error('YAML添加失败')
  } finally {
    yamlAddLoading.value = false
  }
}

const createDialogVisible = ref(false)
const createLoading = ref(false)
const createForm = reactive({
  clusterId: '' as string,
  namespace: '' as string,
  name: '' as string,
  automountServiceAccountToken: true as boolean | undefined,
  imagePullSecrets: [] as string[],
  secrets: [] as string[]
})
const createSteps = [
  { key: 'base', label: '基本信息' },
  { key: 'settings', label: '关联设置' }
]
let createActive = ref(0)
let createYaml = ref('')

const openCreate = () => {
  createForm.clusterId = searchForm.clusterId || ''
  createForm.namespace = searchForm.namespace || ''
  createForm.name = ''
  createForm.automountServiceAccountToken = true
  createForm.imagePullSecrets = []
  createForm.secrets = []
  createActive.value = 0
  createYaml.value = ''
  createDialogVisible.value = true
}

const confirmCreate = async () => {
  if (!createForm.clusterId || !createForm.namespace || !createForm.name) {
    ElMessage.warning('请填写必填项')
    return
  }
  const lines: string[] = [
    'apiVersion: v1',
    'kind: ServiceAccount',
    'metadata:',
    `  name: ${createForm.name}`,
    `  namespace: ${createForm.namespace}`
  ]
  if (typeof createForm.automountServiceAccountToken === 'boolean') {
    lines.push(`automountServiceAccountToken: ${createForm.automountServiceAccountToken}`)
  }
  if (createForm.imagePullSecrets.length) {
    lines.push('imagePullSecrets:')
    for (const name of createForm.imagePullSecrets) {
      lines.push(`  - name: ${name}`)
    }
  }
  if (createForm.secrets.length) {
    lines.push('secrets:')
    for (const name of createForm.secrets) {
      lines.push(`  - name: ${name}`)
    }
  }
  const yaml = lines.join('\n') + '\n'
  createYaml.value = yaml
  createLoading.value = true
  try {
    const toApply = createYaml.value || yaml
    await serviceAccountApi.applyYaml({
      clusterId: createForm.clusterId,
      namespace: createForm.namespace,
      yaml: toApply
    })
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    fetchList()
  } catch {
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
    await serviceAccountApi.delete({
      clusterId: searchForm.clusterId,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name
    })
    ElMessage.success('删除成功')
    deleteDialogVisible.value = false
    fetchList()
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

onMounted(async () => {
  await fetchClusters()
  await fetchList()
})
</script>

<style scoped>
.sa-page {
  padding: 20px;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: var(--gap-4);
  background: #ffffff;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-card);
}
.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}
.header-actions {
  display: flex;
  gap: 12px;
}
.search-card {
  margin-bottom: 20px;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-card);
}
.table-card {
  margin-bottom: 20px;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-card);
}
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}
:deep(.el-dropdown-menu) {
  border-radius: 10px;
  padding: 6px;
}
:deep(.el-dropdown-menu__item) {
  border-radius: 12px;
  margin: 2px 4px;
}
:deep(.el-dropdown-menu__item:hover) {
  background-color: #eef2ff;
  color: #3b82f6;
}
</style>
