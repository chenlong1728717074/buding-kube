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

    <el-dialog v-model="yamlAddDialogVisible" title="YAML添加" width="1200px" :close-on-click-modal="false" class="config-dialog yaml-dialog">
      <template #header>
        <div class="dialog-header">
          <h3 class="dialog-title">YAML添加</h3>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <el-form :model="yamlAddForm" label-width="100px">
            <el-form-item label="命名空间">
              <el-select v-model="yamlAddForm.namespace" placeholder="请选择命名空间" style="width: 240px">
                <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
              </el-select>
            </el-form-item>
          </el-form>
          <div class="yaml-editor-wrapper">
            <YamlEditor :model-value="yamlAddContent" :readonly="false" height="100%" @update:modelValue="val => yamlAddContent = val" />
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="yamlAddDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="yamlAddLoading" @click="confirmYamlAdd">应用</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="createDialogVisible" title="创建ServiceAccount" width="1600px" :close-on-click-modal="false" class="config-dialog">
      <template #header>
        <div class="dialog-header">
          <h3 class="dialog-title">创建ServiceAccount</h3>
          <div class="dialog-actions">
            <el-switch v-model="createYamlMode" active-text="编辑 YAML" @change="(v) => v && buildCreateYaml()" />
          </div>
        </div>
      </template>

      <div class="config-editor">
        <div v-if="!createYamlMode" class="step-tabs">
          <div :class="['step-tab', { active: createSection === 'basic', completed: createForm.name }]" @click="createSection = 'basic'">
            <div class="step-icon">
              <span v-if="!createForm.name">1</span>
              <el-icon v-else><Check /></el-icon>
            </div>
            <div class="step-info">
              <div class="step-title">基本信息</div>
              <div class="step-desc">已设置</div>
            </div>
          </div>
          <div :class="['step-tab', { active: createSection === 'settings' }]" @click="createSection = 'settings'">
            <div class="step-icon">
              <span>2</span>
            </div>
            <div class="step-info">
              <div class="step-title">关联设置</div>
              <div class="step-desc">当前</div>
            </div>
          </div>
        </div>

        <div class="config-content">
          <div v-if="createYamlMode" class="yaml-view">
            <YamlEditor :model-value="createYaml" :readonly="true" height="550px" />
          </div>

          <div v-else-if="createSection === 'basic'" class="basic-form">
            <el-form :model="createForm" label-width="100px" label-position="right">
              <el-form-item label="命名空间" required>
                <el-select v-model="createForm.namespace" placeholder="请选择命名空间" style="width: 100%">
                  <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
                </el-select>
              </el-form-item>
              <el-form-item label="名称" required>
                <el-input v-model="createForm.name" placeholder="请输入名称" />
              </el-form-item>
              <el-form-item label="自动挂载Token">
                <el-switch v-model="createForm.automountServiceAccountToken" />
              </el-form-item>
            </el-form>

            <div class="form-section">
              <div class="section-header-row">
                <div class="section-title">标签</div>
                <el-button size="small" @click="addMetaRow(createLabelRows)">添加标签</el-button>
              </div>
              <el-table :data="createLabelRows" size="default" style="width: 100%">
                <el-table-column label="键" width="300">
                  <template #default="{ row }">
                    <el-input v-model="row.key" placeholder="key" />
                  </template>
                </el-table-column>
                <el-table-column label="值">
                  <template #default="{ row }">
                    <el-input v-model="row.value" placeholder="value" />
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="100" align="center">
                  <template #default="{ $index }">
                    <el-button link type="danger" @click="removeMetaRow(createLabelRows, $index)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <div class="form-section">
              <div class="section-header-row">
                <div class="section-title">注解</div>
                <el-button size="small" @click="addMetaRow(createAnnotationRows)">添加注解</el-button>
              </div>
              <el-table :data="createAnnotationRows" size="default" style="width: 100%">
                <el-table-column label="键" width="300">
                  <template #default="{ row }">
                    <el-input v-model="row.key" placeholder="key" />
                  </template>
                </el-table-column>
                <el-table-column label="值">
                  <template #default="{ row }">
                    <el-input v-model="row.value" placeholder="value" />
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="100" align="center">
                  <template #default="{ $index }">
                    <el-button link type="danger" @click="removeMetaRow(createAnnotationRows, $index)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </div>

          <div v-else class="data-form">
            <div class="form-section">
              <div class="section-header-row">
                <div class="section-title">关联设置</div>
              </div>
              <el-form :model="createForm" label-width="120px" label-position="right" style="max-width: 900px">
                <el-form-item label="镜像拉取Secrets">
                  <el-select v-model="createForm.imagePullSecrets" multiple filterable allow-create default-first-option style="width: 100%" placeholder="输入或选择Secret名称" />
                </el-form-item>
                <el-form-item label="挂载Secrets">
                  <el-select v-model="createForm.secrets" multiple filterable allow-create default-first-option style="width: 100%" placeholder="输入或选择Secret名称" />
                </el-form-item>
              </el-form>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="createDialogVisible = false">取消</el-button>
          <template v-if="!createYamlMode">
            <el-button v-if="createSection === 'settings'" @click="createSection = 'basic'">上一步</el-button>
            <el-button v-if="createSection === 'basic'" type="primary" @click="createSection = 'settings'">下一步</el-button>
            <el-button v-if="createSection === 'settings'" type="primary" :loading="createLoading" @click="confirmCreate">创建</el-button>
          </template>
          <el-button v-else type="primary" :loading="createLoading" @click="confirmCreate">创建</el-button>
        </div>
      </template>
    </el-dialog>

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
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="editInfoDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="editInfoLoading" @click="confirmEditInfo">确定</el-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="yamlDialogVisible" title="查看/编辑YAML" width="90%" :close-on-click-modal="false" class="config-dialog yaml-dialog">
      <template #header>
        <div class="dialog-header">
          <h3 class="dialog-title">查看/编辑YAML</h3>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <div class="yaml-editor-wrapper">
            <YamlEditor :model-value="yamlContent" :readonly="yamlReadOnly" :loading="yamlLoading" height="100%" @update:modelValue="val => yamlContent = val" />
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="yamlDialogVisible = false">关闭</el-button>
          <el-button v-if="!yamlReadOnly" type="primary" :loading="yamlApplyLoading" @click="confirmApplyYaml">应用</el-button>
        </div>
      </template>
    </el-dialog>

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
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh, ArrowDown, Plus, Document, Check } from '@element-plus/icons-vue'
import { serviceAccountApi, type ServiceAccountVO, type ServiceAccountPageQueryDTO } from '@/api/serviceaccount'
import { secretApi, type SecretVO } from '@/api/secret'
import { namespaceApi, type NamespaceVO } from '@/api/namespace'
import { useClusterStore } from '@/stores/cluster'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import '@/assets/styles/config-editor.css'

type KvRow = { key: string; value: string }

const router = useRouter()

// 集群上下文
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)

const loading = ref(false)
const serviceAccounts = ref<ServiceAccountVO[]>([])
const selectedRows = ref<ServiceAccountVO[]>([])

const searchForm = reactive<ServiceAccountPageQueryDTO>({
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
  router.push({ path: '/config/serviceaccount/detail', query: { clusterId: clusterId.value, namespace: row.namespace, name: row.name } })
}

const currentRow = ref<ServiceAccountVO | null>(null)

const editInfoDialogVisible = ref(false)
const editInfoLoading = ref(false)
const editInfoForm = reactive({
  labelsJson: '',
  annotationsJson: ''
})

const addMetaRow = (rows: KvRow[]) => {
  rows.push({ key: '', value: '' })
}
const removeMetaRow = (rows: KvRow[], idx: number) => {
  rows.splice(idx, 1)
}
const rowsToRecord = (rows: KvRow[]) => {
  const out: Record<string, string> = {}
  for (const r of rows) {
    const k = (r.key || '').trim()
    if (!k) continue
    out[k] = String(r.value ?? '')
  }
  return out
}

const yamlDialogVisible = ref(false)
const yamlLoading = ref(false)
const yamlReadOnly = ref(false)
let yamlApplyLoading = ref(false)
let yamlContent = ref('')

const buildServiceAccountYaml = (opts: {
  name: string
  namespace: string
  automountServiceAccountToken?: boolean
  imagePullSecrets?: string[]
  secrets?: string[]
  labels?: Record<string, string>
  annotations?: Record<string, string>
}) => {
  const lines: string[] = [
    'apiVersion: v1',
    'kind: ServiceAccount',
    'metadata:',
    `  name: ${opts.name}`,
    `  namespace: ${opts.namespace}`
  ]

  const labels = opts.labels || {}
  const annotations = opts.annotations || {}

  if (Object.keys(labels).length > 0) {
    lines.push('  labels:')
    for (const [k, v] of Object.entries(labels)) {
      lines.push(`    ${k}: ${JSON.stringify(String(v ?? ''))}`)
    }
  }
  if (Object.keys(annotations).length > 0) {
    lines.push('  annotations:')
    for (const [k, v] of Object.entries(annotations)) {
      lines.push(`    ${k}: ${JSON.stringify(String(v ?? ''))}`)
    }
  }

  if (typeof opts.automountServiceAccountToken === 'boolean') {
    lines.push(`automountServiceAccountToken: ${opts.automountServiceAccountToken}`)
  }
  if (opts.imagePullSecrets?.length) {
    lines.push('imagePullSecrets:')
    for (const name of opts.imagePullSecrets) {
      if (!name) continue
      lines.push(`  - name: ${name}`)
    }
  }
  if (opts.secrets?.length) {
    lines.push('secrets:')
    for (const name of opts.secrets) {
      if (!name) continue
      lines.push(`  - name: ${name}`)
    }
  }

  return lines.join('\n') + '\n'
}

const handleSelectionChange = (rows: ServiceAccountVO[]) => {
  selectedRows.value = rows
}

const fetchNamespaces = async () => {
  if (!clusterId.value) {
    namespaceList.value = []
    return
  }
  try {
    const resp = await namespaceApi.getList({ clusterId: clusterId.value, page: 1, pageSize: 10000 })
    if (resp.code === 200) {
      namespaceList.value = resp.data.items || []
    }
  } catch {
    ElMessage.error('获取命名空间失败')
  }
}

const fetchList = async () => {
  if (!clusterId.value) {
    serviceAccounts.value = []
    pagination.total = 0
    return
  }
  try {
    loading.value = true
    const resp = await serviceAccountApi.getList({
      clusterId: clusterId.value,
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
      clusterId: clusterId.value,
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
  if (!yamlContent.value || !clusterId.value) {
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
      clusterId: clusterId.value,
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
const yamlAddForm = reactive({ namespace: '' as string })
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
  yamlAddForm.namespace = searchForm.namespace || ''
  yamlAddDialogVisible.value = true
}

const confirmYamlAdd = async () => {
  if (!clusterId.value || !yamlAddContent.value) {
    ElMessage.warning('请先选择集群并填写YAML')
    return
  }
  yamlAddLoading.value = true
  try {
    await serviceAccountApi.applyYaml({
      clusterId: clusterId.value,
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
const createSection = ref<'basic' | 'settings'>('basic')
const createYamlMode = ref(false)
const createYaml = ref('')
const createForm = reactive({
  clusterId: '' as string,
  namespace: '' as string,
  name: '' as string,
  automountServiceAccountToken: true as boolean | undefined,
  imagePullSecrets: [] as string[],
  secrets: [] as string[]
})
const createLabelRows = ref<KvRow[]>([])
const createAnnotationRows = ref<KvRow[]>([])

const buildCreateYaml = () => {
  createYaml.value = buildServiceAccountYaml({
    name: createForm.name,
    namespace: createForm.namespace,
    automountServiceAccountToken: createForm.automountServiceAccountToken,
    imagePullSecrets: createForm.imagePullSecrets,
    secrets: createForm.secrets,
    labels: rowsToRecord(createLabelRows.value),
    annotations: rowsToRecord(createAnnotationRows.value)
  })
  return createYaml.value
}

const openCreate = () => {
  createForm.clusterId = clusterId.value || ''
  createForm.namespace = searchForm.namespace || ''
  createForm.name = ''
  createForm.automountServiceAccountToken = true
  createForm.imagePullSecrets = []
  createForm.secrets = []
  createLabelRows.value = []
  createAnnotationRows.value = []
  createSection.value = 'basic'
  createYamlMode.value = false
  createYaml.value = ''
  createDialogVisible.value = true
}

watch(
  () => [createForm.name, createForm.namespace, createForm.automountServiceAccountToken, createForm.imagePullSecrets, createForm.secrets, createLabelRows.value, createAnnotationRows.value],
  () => {
    if (createYamlMode.value) buildCreateYaml()
  },
  { deep: true }
)

const confirmCreate = async () => {
  if (!createForm.namespace || !createForm.name) {
    ElMessage.warning('请填写必填项')
    return
  }
  if (!clusterId.value) {
    ElMessage.warning('请先选择集群')
    return
  }
  const yaml = buildCreateYaml()
  createLoading.value = true
  try {
    const toApply = createYamlMode.value ? createYaml.value : yaml
    await serviceAccountApi.applyYaml({
      clusterId: clusterId.value,
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
      clusterId: clusterId.value,
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

// 监听集群变化
watch(clusterId, async (newClusterId) => {
  if (newClusterId) {
    searchForm.namespace = ''
    pagination.page = 1
    await fetchNamespaces()
    await fetchList()
  } else {
    serviceAccounts.value = []
    pagination.total = 0
  }
}, { immediate: true })

onMounted(async () => {
  if (clusterId.value) {
    await fetchNamespaces()
    await fetchList()
  }
})
</script>

<style scoped>
.sa-page {
  padding: 20px;
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
