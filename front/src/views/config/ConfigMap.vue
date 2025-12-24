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
    <UnifiedDialog v-model="yamlAddDialogVisible" title="YAML添加" subtitle="通过 YAML 快速创建 ConfigMap" width="80%">
      <el-form :model="yamlAddForm" label-width="100px">
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

    <!-- 添加ConfigMap（参考 Kuboard 布局） -->
    <UnifiedDialog v-model="createDialogVisible" title="创建 ConfigMap" subtitle="填写元数据并配置数据键值" width="92%">
      <div class="cm-editor">
        <div class="cm-topbar">
          <el-form :model="createForm" inline class="cm-topbar-form">
            <el-form-item label="命名空间" required>
              <el-select v-model="createForm.namespace" placeholder="请选择命名空间" style="width: 220px">
                <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="名称" required>
              <el-input v-model="createForm.name" placeholder="请输入名称" style="width: 260px" />
            </el-form-item>
          </el-form>
          <div class="cm-topbar-right">
            <el-button text @click="toggleCreateYamlMode">{{ createYamlMode ? '返回表单' : '预览 YAML' }}</el-button>
          </div>
        </div>

        <div class="cm-workbench">
          <div v-if="createYamlMode" class="cm-yaml">
            <YamlEditor :model-value="createYaml" :readonly="true" height="520px" />
          </div>
          <div v-else class="cm-panel">
            <div class="cm-section-switch">
              <el-radio-group v-model="createSection" size="small">
                <el-radio-button label="data">数据</el-radio-button>
                <el-radio-button label="meta">元数据</el-radio-button>
              </el-radio-group>
            </div>
            <div v-if="createSection === 'data'" class="cm-data">
              <div class="kv-toolbar">
                <div class="kv-toolbar-left">
                  <el-button size="small" type="primary" @click="kvCreateRef?.addBlank()">添加条目</el-button>
                  <el-button size="small" @click="openCreateImport">从文件导入</el-button>
                  <input ref="createImportInputRef" class="cm-file-input" style="display: none" type="file" accept=".properties,.ini,.conf,.txt" @change="handleCreateImportChange" />
                  <el-button size="small" @click="resetCreateDataRows">清空</el-button>
                </div>
              </div>
              <KVEditorPane ref="kvCreateRef" v-model="createDataRows" height="520px" />
            </div>
            <div v-else class="cm-meta">
              <el-form :model="createForm" label-width="100px">
                <el-form-item label="别名">
                  <el-input v-model="createForm.alias" placeholder="可选" />
                </el-form-item>
                <el-form-item label="描述">
                  <el-input v-model="createForm.describe" type="textarea" :rows="3" placeholder="可选" />
                </el-form-item>
              </el-form>
              <div class="cm-meta-kv">
                <div class="cm-meta-kv__block">
                  <div class="cm-meta-kv__header">
                    <div class="cm-meta-kv__title">标签</div>
                    <el-button size="small" @click="addMetaRow(createLabelRows)">添加</el-button>
                  </div>
                  <el-table :data="createLabelRows" size="small" border style="width: 100%">
                    <el-table-column label="key" width="260">
                      <template #default="{ row }">
                        <el-input v-model="row.key" placeholder="key" />
                      </template>
                    </el-table-column>
                    <el-table-column label="value">
                      <template #default="{ row }">
                        <el-input v-model="row.value" placeholder="value" />
                      </template>
                    </el-table-column>
                    <el-table-column label="操作" width="90" align="center">
                      <template #default="{ $index }">
                        <el-button link type="danger" @click="removeMetaRow(createLabelRows, $index)">删除</el-button>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>

                <div class="cm-meta-kv__block">
                  <div class="cm-meta-kv__header">
                    <div class="cm-meta-kv__title">注解</div>
                    <el-button size="small" @click="addMetaRow(createAnnotationRows)">添加</el-button>
                  </div>
                  <el-table :data="createAnnotationRows" size="small" border style="width: 100%">
                    <el-table-column label="key" width="260">
                      <template #default="{ row }">
                        <el-input v-model="row.key" placeholder="key" />
                      </template>
                    </el-table-column>
                    <el-table-column label="value">
                      <template #default="{ row }">
                        <el-input v-model="row.value" placeholder="value" />
                      </template>
                    </el-table-column>
                    <el-table-column label="操作" width="90" align="center">
                      <template #default="{ $index }">
                        <el-button link type="danger" @click="removeMetaRow(createAnnotationRows, $index)">删除</el-button>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="cm-footer">
          <div class="cm-footer-right">
            <el-button @click="createDialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="createLoading" @click="confirmCreate">创建</el-button>
          </div>
        </div>
      </template>
    </UnifiedDialog>

    <!-- 修改别名/备注 -->
    <UnifiedDialog v-model="editInfoDialogVisible" title="编辑信息" subtitle="修改别名与备注" width="80%">
      <div class="group-box" style="margin-bottom: 12px;">
        <div class="group-title">基本信息</div>
        <el-form :model="editInfoForm" label-width="100px">
          <el-form-item label="名称">
            <el-input :model-value="currentRow?.name || ''" disabled />
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
      <template #footer>
        <el-button @click="editInfoDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editInfoLoading" @click="confirmEditInfo">确定</el-button>
      </template>
    </UnifiedDialog>

    <!-- 编辑ConfigMap（参考 Kuboard 布局） -->
    <UnifiedDialog v-model="editDialogVisible" title="编辑 ConfigMap" subtitle="修改元数据与 data 键值" width="92%">
      <div class="cm-editor">
        <div class="cm-topbar cm-topbar--edit">
          <div class="cm-topbar-left">
            <div class="cm-ident">{{ currentRow?.namespace || '-' }} / {{ currentRow?.name || '-' }}</div>
            <div class="cm-meta-line">创建时间：{{ formatDate(currentRow?.creationTimestamp) }}</div>
          </div>
          <div class="cm-topbar-right">
            <el-button text @click="toggleEditYamlMode">{{ editYamlMode ? '返回表单' : '预览 YAML' }}</el-button>
          </div>
        </div>

        <div class="cm-workbench">
          <div v-if="editYamlMode" class="cm-yaml">
            <YamlEditor :model-value="editYaml" :readonly="true" height="520px" />
          </div>
          <div v-else class="cm-panel">
            <div class="cm-section-switch">
              <el-radio-group v-model="editSection" size="small">
                <el-radio-button label="data">数据</el-radio-button>
                <el-radio-button label="meta">元数据</el-radio-button>
              </el-radio-group>
            </div>
            <div v-if="editSection === 'data'" class="cm-data">
              <div class="kv-toolbar">
                <div class="kv-toolbar-left">
                  <el-button size="small" type="primary" @click="kvEditRef?.addBlank()">添加条目</el-button>
                  <el-button size="small" @click="openEditImport">从文件导入</el-button>
                  <input ref="editImportInputRef" class="cm-file-input" style="display: none" type="file" accept=".properties,.ini,.conf,.txt" @change="handleEditImportChange" />
                  <el-button size="small" @click="resetEditDataRows">重置</el-button>
                </div>
              </div>
              <KVEditorPane ref="kvEditRef" v-model="editDataRows" height="520px" />
            </div>
            <div v-else class="cm-meta">
              <el-form :model="editForm" label-width="100px">
                <el-form-item label="别名">
                  <el-input v-model="editForm.alias" placeholder="可选" />
                </el-form-item>
                <el-form-item label="描述">
                  <el-input v-model="editForm.describe" type="textarea" :rows="3" placeholder="可选" />
                </el-form-item>
              </el-form>
              <div class="cm-meta-kv">
                <div class="cm-meta-kv__block">
                  <div class="cm-meta-kv__header">
                    <div class="cm-meta-kv__title">标签</div>
                    <el-button size="small" @click="addMetaRow(editLabelRows)">添加</el-button>
                  </div>
                  <el-table :data="editLabelRows" size="small" border style="width: 100%">
                    <el-table-column label="key" width="260">
                      <template #default="{ row }">
                        <el-input v-model="row.key" placeholder="key" />
                      </template>
                    </el-table-column>
                    <el-table-column label="value">
                      <template #default="{ row }">
                        <el-input v-model="row.value" placeholder="value" />
                      </template>
                    </el-table-column>
                    <el-table-column label="操作" width="90" align="center">
                      <template #default="{ $index }">
                        <el-button link type="danger" @click="removeMetaRow(editLabelRows, $index)">删除</el-button>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>

                <div class="cm-meta-kv__block">
                  <div class="cm-meta-kv__header">
                    <div class="cm-meta-kv__title">注解</div>
                    <el-button size="small" @click="addMetaRow(editAnnotationRows)">添加</el-button>
                  </div>
                  <el-table :data="editAnnotationRows" size="small" border style="width: 100%">
                    <el-table-column label="key" width="260">
                      <template #default="{ row }">
                        <el-input v-model="row.key" placeholder="key" />
                      </template>
                    </el-table-column>
                    <el-table-column label="value">
                      <template #default="{ row }">
                        <el-input v-model="row.value" placeholder="value" />
                      </template>
                    </el-table-column>
                    <el-table-column label="操作" width="90" align="center">
                      <template #default="{ $index }">
                        <el-button link type="danger" @click="removeMetaRow(editAnnotationRows, $index)">删除</el-button>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="cm-footer">
          <div class="cm-footer-right">
            <el-button @click="editDialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="editLoading" @click="confirmEdit">保存</el-button>
          </div>
        </div>
      </template>
    </UnifiedDialog>

    <!-- 查看/编辑YAML -->
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
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh, ArrowDown, Plus, Document } from '@element-plus/icons-vue'
import { configMapApi, type ConfigMapVO, type ConfigMapPageQueryDTO } from '@/api/configmap'
import { namespaceApi, type NamespaceVO } from '@/api/namespace'
import { useClusterStore } from '@/stores/cluster'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import UnifiedDialog from '@/components/UnifiedDialog.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import KVEditorPane from '@/components/KVEditorPane.vue'
import { ref as vRef } from 'vue'

type KvRow = { key: string; value: string }

// 集群上下文
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)

const loading = ref(false)
const configMaps = ref<ConfigMapVO[]>([])
const selectedRows = ref<ConfigMapVO[]>([])

const searchForm = reactive<ConfigMapPageQueryDTO>({
  namespace: '',
  page: 1,
  pageSize: 20,
  keyword: ''
})

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

const editDialogVisible = ref(false)
const editLoading = ref(false)
const editSection = ref<'data' | 'meta'>('data')
const editYamlMode = ref(false)
const editYaml = ref('')
const editForm = reactive({ alias: '', describe: '' })
const editDataRows = ref<KvRow[]>([])
const editLabelRows = ref<KvRow[]>([])
const editAnnotationRows = ref<KvRow[]>([])
const kvEditRef = vRef<any>()
const editImportInputRef = vRef<HTMLInputElement | null>(null)

const yamlDialogVisible = ref(false)
const yamlLoading = ref(false)
const yamlReadOnly = ref(false)
let yamlApplyLoading = ref(false)
let yamlContent = ref('')

const handleSelectionChange = (rows: ConfigMapVO[]) => { selectedRows.value = rows }

const fetchNamespaces = async () => {
  if (!clusterId.value) { namespaceList.value = []; return }
  try {
    const resp = await namespaceApi.getList({ clusterId: clusterId.value, page: 1, pageSize: 10000 })
    if (resp.code === 200) {
      namespaceList.value = resp.data.items || []
    }
  } catch (e) {
    ElMessage.error('获取命名空间失败')
  }
}

const fetchList = async () => {
  if (!clusterId.value) { configMaps.value = []; pagination.total = 0; return }
  try {
    loading.value = true
    const resp = await configMapApi.getList({ clusterId: clusterId.value, namespace: searchForm.namespace || '', page: pagination.page, pageSize: pagination.pageSize, keyword: searchForm.keyword || '' })
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
  router.push({ path: '/config/configmap/detail', query: { clusterId: clusterId.value, namespace: row.namespace, name: row.name } })
}

const handleRowAction = (cmd: string, row: ConfigMapVO) => {
  currentRow.value = row
  switch (cmd) {
    case 'editInfo':
      editInfoForm.alias = row.alias || ''
      editInfoForm.describe = row.describe || ''
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
      openEdit(row)
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
    const payload = { clusterId: clusterId.value, namespace: currentRow.value.namespace, name: currentRow.value.name, alias: editInfoForm.alias, describe: editInfoForm.describe }
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

const confirmApplyYaml = async () => {
  if (!yamlContent.value || !clusterId.value) { ElMessage.warning('YAML内容为空'); return }
  if (!currentRow.value?.namespace) { ElMessage.warning('命名空间为空'); return }
  yamlApplyLoading.value = true
  try {
    await configMapApi.applyYaml({ clusterId: clusterId.value, namespace: currentRow.value.namespace, yaml: yamlContent.value })
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
const yamlAddForm = reactive({ namespace: '' as string })
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
  yamlAddForm.namespace = searchForm.namespace || ''
  yamlAddDialogVisible.value = true
}
const confirmYamlAdd = async () => {
  if (!clusterId.value || !yamlAddContent.value) { ElMessage.warning('请选择集群并填写YAML'); return }
  yamlAddLoading.value = true
  try {
    await configMapApi.applyYaml({ clusterId: clusterId.value, namespace: yamlAddForm.namespace || '', yaml: yamlAddContent.value })
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
const createSection = ref<'data' | 'meta'>('data')
const createYamlMode = ref(false)
const kvCreateRef = vRef<any>()
const createImportInputRef = vRef<HTMLInputElement | null>(null)
const createForm = reactive({ namespace: '' as string, name: '' as string, alias: '' as string, describe: '' as string })
const createDataRows = ref<KvRow[]>([])
const createLabelRows = ref<KvRow[]>([])
const createAnnotationRows = ref<KvRow[]>([])
const createYaml = ref('')
const openCreate = () => {
  createForm.namespace = searchForm.namespace || ''
  createForm.name = ''
  createForm.alias = ''
  createForm.describe = ''
  createDataRows.value = []
  createLabelRows.value = []
  createAnnotationRows.value = []
  createYaml.value = ''
  createYamlMode.value = false
  createSection.value = 'data'
  createDialogVisible.value = true
}
const resetCreateDataRows = () => { createDataRows.value = [] }

const buildYamlFromRows = (name: string, namespace: string, labels: Record<string, string>, ann: Record<string, string>, data: Record<string, string>) => {
  const lines: string[] = [
    'apiVersion: v1',
    'kind: ConfigMap',
    'metadata:',
    `  name: ${name}`,
    `  namespace: ${namespace}`
  ]
  if (Object.keys(labels).length) {
    lines.push('  labels:')
    for (const [k, v] of Object.entries(labels)) {
      lines.push(`    ${k}: ${v}`)
    }
  }
  if (Object.keys(ann).length) {
    lines.push('  annotations:')
    for (const [k, v] of Object.entries(ann)) {
      lines.push(`    ${k}: ${v}`)
    }
  }
  if (Object.keys(data).length) {
    lines.push('data:')
    for (const [k, v] of Object.entries(data)) {
      lines.push(`  ${k}: |`)
      const vv = String(v ?? '')
      const rowLines = vv.split('\n')
      for (const l of rowLines) {
        lines.push(`    ${l}`)
      }
    }
  }
  return lines.join('\n') + '\n'
}

const rowsToRecord = (rows: KvRow[], { excludeKeys }: { excludeKeys?: string[] } = {}) => {
  const result: Record<string, string> = {}
  const exclude = new Set((excludeKeys || []).map(k => String(k)))
  for (const { key, value } of rows) {
    const k = (key || '').trim()
    if (!k) continue
    if (exclude.has(k)) continue
    result[k] = String(value ?? '')
  }
  return result
}

const collectCreateLabels = () => rowsToRecord(createLabelRows.value)

const collectCreateAnnotations = () => {
  const ann: Record<string, string> = {}
  if (createForm.alias) ann.alias = createForm.alias
  if (createForm.describe) ann.describe = createForm.describe
  Object.assign(ann, rowsToRecord(createAnnotationRows.value, { excludeKeys: ['alias', 'describe'] }))
  return ann
}

const collectRowsData = (rows: KvRow[]) => {
  const data: Record<string, string> = {}
  for (const { key, value } of rows) {
    const k = (key || '').trim()
    if (!k) continue
    data[k] = value ?? ''
  }
  return data
}

const toggleCreateYamlMode = () => {
  if (!createYamlMode.value) {
    const labels = collectCreateLabels()
    const ann = collectCreateAnnotations()
    const data = collectRowsData(createDataRows.value)
    createYaml.value = buildYamlFromRows(createForm.name, createForm.namespace, labels, ann, data)
    createYamlMode.value = true
    return
  }
  createYamlMode.value = false
}

const openCreateImport = () => {
  createImportInputRef.value?.click()
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

const handleCreateImportChange = async (e: Event) => {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  try {
    const text = await file.text()
    const parsed = parseKvFromText(text)
    const before = createDataRows.value.length
    mergeDataRows(createDataRows.value, parsed)
    const after = createDataRows.value.length
    ElMessage.success(`已导入 ${Object.keys(parsed).length} 项，新增 ${Math.max(after - before, 0)} 项`)
  } catch {
    ElMessage.error('导入失败')
  } finally {
    input.value = ''
  }
}

const confirmCreate = async () => {
  if (!clusterId.value || !createForm.namespace || !createForm.name) { ElMessage.warning('请填写必填项'); return }
  const labels = collectCreateLabels()
  const data = collectRowsData(createDataRows.value)
  const annotations = rowsToRecord(createAnnotationRows.value, { excludeKeys: ['alias', 'describe'] })
  createYaml.value = buildYamlFromRows(createForm.name, createForm.namespace, { ...labels }, collectCreateAnnotations(), data)
  createLoading.value = true
  try {
    await configMapApi.add({
      clusterId: clusterId.value,
      namespace: createForm.namespace,
      name: createForm.name,
      alias: createForm.alias,
      describe: createForm.describe,
      data,
      labels,
      annotations
    })
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    fetchList()
  } catch (e) {
    ElMessage.error('创建失败')
  } finally {
    createLoading.value = false
  }
}

const openEdit = (row: ConfigMapVO) => {
  editForm.alias = row.alias || ''
  editForm.describe = row.describe || ''
  editDataRows.value = Object.entries(row.data || {}).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
  editLabelRows.value = Object.entries(row.labels || {}).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
  editAnnotationRows.value = Object.entries(row.annotations || {})
    .filter(([k]) => k !== 'alias' && k !== 'describe')
    .map(([k, v]) => ({ key: k, value: String(v ?? '') }))
  editSection.value = 'data'
  editYamlMode.value = false
  editYaml.value = ''
  editDialogVisible.value = true
}

const resetEditDataRows = () => {
  editDataRows.value = Object.entries(currentRow.value?.data || {}).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
}

const collectEditLabels = () => rowsToRecord(editLabelRows.value)
const collectEditAnnotations = () => {
  const ann: Record<string, string> = {}
  if (editForm.alias) ann.alias = editForm.alias
  if (editForm.describe) ann.describe = editForm.describe
  Object.assign(ann, rowsToRecord(editAnnotationRows.value, { excludeKeys: ['alias', 'describe'] }))
  return ann
}

const toggleEditYamlMode = () => {
  if (!currentRow.value) return
  if (!editYamlMode.value) {
    const data = collectRowsData(editDataRows.value)
    const labels = collectEditLabels()
    const ann = collectEditAnnotations()
    editYaml.value = buildYamlFromRows(currentRow.value.name, currentRow.value.namespace, labels, ann, data)
    editYamlMode.value = true
    return
  }
  editYamlMode.value = false
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

const confirmEdit = async () => {
  if (!currentRow.value) return
  editLoading.value = true
  try {
    const data = collectRowsData(editDataRows.value)
    const labels = collectEditLabels()
    const annotations = rowsToRecord(editAnnotationRows.value, { excludeKeys: ['alias', 'describe'] })
    await configMapApi.updateInfo({
      clusterId: clusterId.value,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      alias: editForm.alias,
      describe: editForm.describe
    })
    await configMapApi.updateSetting({
      clusterId: clusterId.value,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      labels,
      annotations
    })
    await configMapApi.updateData({
      clusterId: clusterId.value,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      data
    })
    ElMessage.success('保存成功')
    editDialogVisible.value = false
    fetchList()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    editLoading.value = false
  }
}

const deleteDialogVisible = ref(false)
const deleteLoading = ref(false)
const confirmDelete = async () => {
  if (!currentRow.value) return
  deleteLoading.value = true
  try {
    await configMapApi.delete({ clusterId: clusterId.value, namespace: currentRow.value.namespace, name: currentRow.value.name })
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

// 监听集群变化
watch(clusterId, async (newClusterId) => {
  if (newClusterId) {
    searchForm.namespace = ''
    pagination.page = 1
    await fetchNamespaces()
    await fetchList()
  } else {
    configMaps.value = []
    pagination.total = 0
  }
})

onMounted(async () => {
  if (clusterId.value) {
    await fetchNamespaces()
    await fetchList()
  }
})

const addMetaRow = (rows: KvRow[]) => {
  rows.push({ key: '', value: '' })
}
const removeMetaRow = (rows: KvRow[], idx: number) => {
  rows.splice(idx, 1)
}
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
.header-actions { display: flex; gap: 12px; }
.search-card { margin-bottom: 20px; border-radius: var(--radius-md); box-shadow: var(--shadow-card); }
.table-card { margin-bottom: 20px; border-radius: var(--radius-md); box-shadow: var(--shadow-card); }
.pagination-wrapper { display: flex; justify-content: flex-end; margin-top: 20px; }
.kv-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: 8px; }
.kv-toolbar-left, .kv-toolbar-right { display: flex; align-items: center; gap: 8px; }
.cm-file-input { display: none; }
.cm-topbar { display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; }
.cm-topbar-right { padding-top: 2px; }
.cm-footer { display: flex; justify-content: flex-end; align-items: center; }
.cm-section-switch { margin-bottom: 10px; }
.cm-meta-kv { display: grid; gap: 12px; margin-top: 12px; }
.cm-meta-kv__header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.cm-meta-kv__title { font-weight: 600; color: #334155; }
.cm-topbar--edit { display: flex; justify-content: space-between; gap: 12px; }

/* 下拉菜单圆润样式 */
:deep(.el-dropdown-menu) { border-radius: 10px; padding: 6px; }
:deep(.el-dropdown-menu__item) { border-radius: 12px; margin: 2px 4px; }
:deep(.el-dropdown-menu__item:hover) { background-color: #eef2ff; color: #3b82f6; }
</style>
