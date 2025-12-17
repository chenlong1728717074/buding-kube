<template>
  <div class="secret-page">
    <div class="page-header">
      <h1>Secret管理</h1>
      <div class="header-actions">
        <el-button type="primary" @click="openCreate">
          <el-icon><Plus /></el-icon>
          添加Secret
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
      <el-table v-loading="loading" :data="secrets" stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="55" header-align="center" align="center" />
        <el-table-column prop="name" label="名称" min-width="160" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="goDetail(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="140" header-align="center" align="center" />
        <el-table-column prop="type" label="类型" width="160" header-align="center" align="center" />
        <el-table-column prop="creationTimestamp" label="创建时间" min-width="180" :show-overflow-tooltip="true" header-align="center" align="center">
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
                    <el-dropdown-item command="editData">编辑数据</el-dropdown-item>
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

    <UnifiedDialog v-model="yamlAddDialogVisible" title="YAML添加" subtitle="通过 YAML 快速创建 Secret" width="80%">
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

    <UnifiedDialog v-model="createDialogVisible" title="创建 Secret" subtitle="填写元数据并配置数据键值" width="92%">
      <div class="cm-editor">
        <div class="cm-topbar">
          <el-form :model="createForm" inline class="cm-topbar-form">
            <el-form-item label="集群" required>
              <el-select v-model="createForm.clusterId" placeholder="请选择集群" style="width: 220px">
                <el-option v-for="c in clusterList" :key="c.id" :label="c.name" :value="c.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="命名空间" required>
              <el-select v-model="createForm.namespace" placeholder="请选择命名空间" style="width: 220px">
                <el-option v-for="ns in namespaceList" :key="ns.name" :label="ns.name" :value="ns.name" />
              </el-select>
            </el-form-item>
            <el-form-item label="名称" required>
              <el-input v-model="createForm.name" placeholder="请输入名称" style="width: 260px" />
            </el-form-item>
            <el-form-item label="类型">
              <el-select v-model="createForm.type" placeholder="请选择类型" style="width: 260px">
                <el-option label="Opaque" value="Opaque" />
                <el-option label="kubernetes.io/dockerconfigjson" value="kubernetes.io/dockerconfigjson" />
                <el-option label="kubernetes.io/tls" value="kubernetes.io/tls" />
              </el-select>
            </el-form-item>
          </el-form>
        </div>

        <div class="cm-workbench">
          <div v-if="createYamlMode" class="cm-yaml">
            <YamlEditor :model-value="createYaml" :readonly="true" height="520px" />
          </div>
          <div v-else class="cm-panel">
            <el-tabs v-model="createSection" class="cm-tabs">
              <el-tab-pane label="数据" name="data">
                <div class="cm-data">
                  <template v-if="createForm.type === 'Opaque'">
                    <div class="kv-toolbar">
                      <div class="kv-toolbar-left">
                        <el-button size="small" type="primary" @click="kvCreateRef?.addBlank()">添加条目</el-button>
                        <el-button size="small" @click="openCreateImport">从文件导入</el-button>
                        <input ref="createImportInputRef" class="cm-file-input" type="file" accept=".properties,.ini,.conf,.txt" @change="handleCreateImportChange" />
                      </div>
                      <div class="kv-toolbar-right">
                        <el-button size="small" @click="resetCreateDataRows">清空</el-button>
                      </div>
                    </div>
                    <KVEditorPane ref="kvCreateRef" v-model="createDataRows" height="420px" />
                  </template>

                  <template v-else-if="createForm.type === 'kubernetes.io/dockerconfigjson'">
                    <el-form :model="createDockerForm" label-width="140px" style="max-width: 720px">
                      <el-form-item label="docker server" required>
                        <el-input v-model="createDockerForm.server" placeholder="例如 registry.example.com" />
                      </el-form-item>
                      <el-form-item label="docker username" required>
                        <el-input v-model="createDockerForm.username" placeholder="用户名" />
                      </el-form-item>
                      <el-form-item label="docker password" required>
                        <el-input v-model="createDockerForm.password" type="password" show-password placeholder="密码" />
                      </el-form-item>
                    </el-form>
                  </template>

                  <template v-else-if="createForm.type === 'kubernetes.io/tls'">
                    <el-form :model="createTlsForm" label-width="100px">
                      <el-form-item label="tls.crt" required>
                        <el-input v-model="createTlsForm.crt" type="textarea" :rows="8" placeholder="粘贴证书内容" />
                      </el-form-item>
                      <el-form-item label="tls.key" required>
                        <el-input v-model="createTlsForm.key" type="textarea" :rows="8" placeholder="粘贴私钥内容" />
                      </el-form-item>
                    </el-form>
                  </template>
                </div>
              </el-tab-pane>
              <el-tab-pane label="元数据" name="meta">
                <div class="cm-meta">
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
              </el-tab-pane>
            </el-tabs>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="cm-footer">
          <el-button text @click="toggleCreateYamlMode">{{ createYamlMode ? '返回表单' : '预览 YAML' }}</el-button>
          <div class="cm-footer-right">
            <el-button @click="createDialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="createLoading" @click="confirmCreate">创建</el-button>
          </div>
        </div>
      </template>
    </UnifiedDialog>

    <UnifiedDialog v-model="editDialogVisible" title="编辑 Secret" subtitle="修改元数据与数据键值" width="92%">
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
            <el-tabs v-model="editSection" class="cm-tabs">
              <el-tab-pane label="数据" name="data">
                <div class="cm-data">
                  <div class="kv-toolbar">
                    <div class="kv-toolbar-left">
                      <el-button size="small" type="primary" @click="kvEditSettingRef?.addBlank()">添加条目</el-button>
                      <el-button size="small" @click="openEditImport">从文件导入</el-button>
                      <input ref="editImportInputRef" class="cm-file-input" type="file" accept=".properties,.ini,.conf,.txt" @change="handleEditImportChange" />
                    </div>
                    <div class="kv-toolbar-right">
                      <el-button size="small" @click="resetEditDataRows">重置</el-button>
                    </div>
                  </div>
                  <KVEditorPane ref="kvEditSettingRef" v-model="editDataRows" height="420px" />
                </div>
              </el-tab-pane>
              <el-tab-pane label="元数据" name="meta">
                <div class="cm-meta">
                  <el-form :model="editForm" label-width="100px">
                    <el-form-item label="类型">
                      <el-select v-model="editForm.type" placeholder="请选择类型" style="width: 320px">
                        <el-option label="Opaque" value="Opaque" />
                        <el-option label="kubernetes.io/dockerconfigjson" value="kubernetes.io/dockerconfigjson" />
                        <el-option label="kubernetes.io/tls" value="kubernetes.io/tls" />
                      </el-select>
                    </el-form-item>
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
              </el-tab-pane>
            </el-tabs>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="cm-footer">
          <div class="cm-footer-right">
            <el-button @click="editDialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="editLoading" @click="confirmEditSetting">保存</el-button>
          </div>
        </div>
      </template>
    </UnifiedDialog>

    <UnifiedDialog v-model="editInfoDialogVisible" title="编辑信息" subtitle="修改别名与备注" width="80%">
      <div class="group-box" style="margin-bottom: 12px;">
        <div class="group-title">基本信息</div>
        <el-form :model="editInfoForm" label-width="100px">
          <el-form-item label="名称">
            <el-input :model-value="currentRow?.name || ''" disabled />
          </el-form-item>
          <el-form-item label="别名">
            <el-input v-model="editInfoForm.alias" />
            <div class="helper">别名只能包含中文、字母、数字和连字符（-），不得以连字符（-）开头或结尾，最长 63 个字符。</div>
          </el-form-item>
          <el-form-item label="备注">
            <el-input v-model="editInfoForm.describe" type="textarea" :rows="3" />
            <div class="helper">描述可包含任意字符，最长 256 个字符。</div>
          </el-form-item>
          <el-form-item label="快捷操作">
            <el-button type="primary" @click="openEditDataQuick">添加数据</el-button>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <el-button @click="editInfoDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editInfoLoading" @click="confirmEditInfo">确定</el-button>
      </template>
    </UnifiedDialog>

    <UnifiedDialog v-model="editDataDialogVisible" title="编辑数据" subtitle="仅修改数据键值" width="80%">
      <div class="kv-toolbar">
        <el-button size="small" type="primary" @click="kvEditRef?.addBlank()">添加数据</el-button>
        <el-button size="small" @click="resetDataRows()">重置</el-button>
      </div>
      <KVEditorPane ref="kvEditRef" v-model="dataRows" height="420px" />
      <template #footer>
        <el-button @click="editDataDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editDataLoading" @click="confirmEditData">保存</el-button>
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
      message="确定要删除该Secret吗？"
      :loading="deleteLoading"
      @confirm="confirmDelete"
      @cancel="() => deleteDialogVisible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh, ArrowDown, Plus, Document } from '@element-plus/icons-vue'
import { secretApi, type SecretVO, type SecretPageQueryDTO } from '@/api/secret'
import { clusterApi, type ClusterVO } from '@/api/cluster'
import { namespaceApi, type NamespaceVO } from '@/api/namespace'
import UnifiedDialog from '@/components/UnifiedDialog.vue'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import KVEditorPane from '@/components/KVEditorPane.vue'

type KvRow = { key: string; value: string }

const router = useRouter()

const loading = ref(false)
const secrets = ref<SecretVO[]>([])
const selectedRows = ref<SecretVO[]>([])

const searchForm = reactive<SecretPageQueryDTO>({
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

const goDetail = (row: SecretVO) => {
  const clusterId = searchForm.clusterId
  router.push({ path: '/config/secret/detail', query: { clusterId, namespace: row.namespace, name: row.name } })
}

const currentRow = ref<SecretVO | null>(null)

const editInfoDialogVisible = ref(false)
const editInfoLoading = ref(false)
const editInfoForm = reactive({ alias: '', describe: '' })

const editDataDialogVisible = ref(false)
const editDataLoading = ref(false)
const dataRows = ref<{ key: string; value: string }[]>([])
const kvEditRef = ref<any>()

const yamlDialogVisible = ref(false)
const yamlLoading = ref(false)
const yamlReadOnly = ref(false)
let yamlApplyLoading = ref(false)
let yamlContent = ref('')

const handleSelectionChange = (rows: SecretVO[]) => {
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
    secrets.value = []
    pagination.total = 0
    return
  }
  try {
    loading.value = true
    const resp = await secretApi.getList({
      clusterId: searchForm.clusterId,
      namespace: searchForm.namespace || '',
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword || ''
    })
    if (resp.code === 200) {
      const items = (resp.data as any).items || (resp.data as any).list || []
      const total = (resp.data as any).total || (resp.data as any).totalCount || 0
      secrets.value = items
      pagination.total = total
    }
  } catch {
    ElMessage.error('获取Secret列表失败')
  } finally {
    loading.value = false
  }
}

const handleClusterChange = async () => {
  searchForm.namespace = ''
  pagination.page = 1
  await fetchNamespaces()
  secrets.value = []
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

const rowsFromSecret = (row?: SecretVO | null) => {
  const sd = row?.stringData
  if (sd && Object.keys(sd).length > 0) {
    return Object.entries(sd).map(([k, v]) => ({ key: k, value: v }))
  }
  const data = row?.data || {}
  return Object.entries(data).map(([k, v]) => ({ key: k, value: decodeBase64ToUtf8(v) }))
}

const fetchSecretDetail = async (row: SecretVO) => {
  if (!searchForm.clusterId) return row
  const resp = await secretApi.getInfo({ clusterId: searchForm.clusterId, namespace: row.namespace, name: row.name })
  return resp.data
}

const handleRowAction = async (cmd: string, row: SecretVO) => {
  currentRow.value = row
  switch (cmd) {
    case 'editInfo':
      editInfoForm.alias = row.alias || ''
      editInfoForm.describe = row.describe || ''
      editInfoDialogVisible.value = true
      break
    case 'editData':
      try {
        const detail = await fetchSecretDetail(row)
        currentRow.value = { ...row, ...detail }
      } catch {}
      dataRows.value = rowsFromSecret(currentRow.value)
      editDataDialogVisible.value = true
      break
    case 'yaml':
      yamlLoading.value = true
      yamlDialogVisible.value = true
      yamlReadOnly.value = false
      try {
        const detail = await fetchSecretDetail(row)
        currentRow.value = { ...row, ...detail }
        yamlContent.value = currentRow.value.yaml || ''
      } catch {
        yamlContent.value = row.yaml || ''
      } finally {
        yamlLoading.value = false
      }
      break
    case 'editSetting':
      await openEditSetting(row)
      break
  }
}

const openDelete = (row: SecretVO) => {
  currentRow.value = row
  deleteDialogVisible.value = true
}
const openEditDataQuick = async () => {
  editInfoDialogVisible.value = false
  if (currentRow.value) {
    try {
      const detail = await fetchSecretDetail(currentRow.value)
      currentRow.value = { ...currentRow.value, ...detail }
    } catch {}
  }
  dataRows.value = rowsFromSecret(currentRow.value)
  editDataDialogVisible.value = true
}

const confirmEditInfo = async () => {
  if (!currentRow.value) return
  editInfoLoading.value = true
  try {
    const payload = {
      clusterId: searchForm.clusterId,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      alias: editInfoForm.alias,
      describe: editInfoForm.describe
    }
    await secretApi.updateInfo(payload)
    ElMessage.success('更新成功')
    editInfoDialogVisible.value = false
    fetchList()
  } catch {
    ElMessage.error('更新失败')
  } finally {
    editInfoLoading.value = false
  }
}

const resetDataRows = () => {
  dataRows.value = rowsFromSecret(currentRow.value)
}

const confirmEditData = async () => {
  if (!currentRow.value) return
  editDataLoading.value = true
  try {
    const stringData: Record<string, string> = {}
    for (const { key, value } of dataRows.value) {
      if (!key) continue
      stringData[key] = value
    }
    const payload = {
      clusterId: searchForm.clusterId,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      stringData
    }
    await secretApi.updateData(payload)
    ElMessage.success('数据已保存')
    editDataDialogVisible.value = false
    fetchList()
  } catch {
    ElMessage.error('保存失败')
  } finally {
    editDataLoading.value = false
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
    await secretApi.applyYaml({
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
    'kind: Secret',
    'metadata:',
    '  name: example-secret',
    '  namespace: default',
    'type: Opaque',
    'stringData:',
    '  username: admin',
    '  password: password'
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
    await secretApi.applyYaml({
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

const rowsToStringData = (rows: KvRow[]) => {
  const result: Record<string, string> = {}
  for (const { key, value } of rows) {
    const k = (key || '').trim()
    if (!k) continue
    result[k] = String(value ?? '')
  }
  return result
}

const encodeUtf8ToBase64 = (text: string) => {
  try {
    const bytes = new TextEncoder().encode(String(text ?? ''))
    let bin = ''
    for (let i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i])
    return btoa(bin)
  } catch {
    try {
      return btoa(String(text ?? ''))
    } catch {
      return ''
    }
  }
}

const buildDockerConfigJson = (server: string, username: string, password: string) => {
  const s = (server || '').trim()
  const u = (username || '').trim()
  const p = String(password ?? '')
  const auth = encodeUtf8ToBase64(`${u}:${p}`)
  const obj = {
    auths: {
      [s]: {
        username: u,
        password: p,
        auth
      }
    }
  }
  return JSON.stringify(obj, null, 2)
}

const buildSecretYamlFromRows = (
  name: string,
  namespace: string,
  type: string,
  labels: Record<string, string>,
  annotations: Record<string, string>,
  stringData: Record<string, string>
) => {
  const lines: string[] = ['apiVersion: v1', 'kind: Secret', 'metadata:', `  name: ${name}`, `  namespace: ${namespace}`]
  if (Object.keys(labels).length) {
    lines.push('  labels:')
    for (const [k, v] of Object.entries(labels)) {
      lines.push(`    ${k}: ${v}`)
    }
  }
  if (Object.keys(annotations).length) {
    lines.push('  annotations:')
    for (const [k, v] of Object.entries(annotations)) {
      lines.push(`    ${k}: ${v}`)
    }
  }
  if (type) {
    lines.push(`type: ${type}`)
  }
  if (Object.keys(stringData).length) {
    lines.push('stringData:')
    for (const [k, v] of Object.entries(stringData)) {
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

const mergeDataRows = (rows: KvRow[], incoming: Record<string, string>) => {
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

const addMetaRow = (rows: KvRow[]) => {
  rows.push({ key: '', value: '' })
}
const removeMetaRow = (rows: KvRow[], idx: number) => {
  rows.splice(idx, 1)
}

const createDialogVisible = ref(false)
const createLoading = ref(false)
const createSection = ref<'data' | 'meta'>('data')
const createYamlMode = ref(false)
const createYaml = ref('')
const kvCreateRef = ref<any>()
const createImportInputRef = ref<HTMLInputElement | null>(null)
const createForm = reactive({
  clusterId: '' as string,
  namespace: '' as string,
  name: '' as string,
  type: '' as string,
  alias: '' as string,
  describe: '' as string
})
const createDockerForm = reactive({ server: '', username: '', password: '' })
const createTlsForm = reactive({ crt: '', key: '' })
const createDataRows = ref<KvRow[]>([])
const createLabelRows = ref<KvRow[]>([])
const createAnnotationRows = ref<KvRow[]>([])

const getCreateStringData = () => {
  if (createForm.type === 'kubernetes.io/dockerconfigjson') {
    const server = (createDockerForm.server || '').trim()
    const username = (createDockerForm.username || '').trim()
    const password = String(createDockerForm.password ?? '')
    if (!server || !username || !password) return {}
    return { '.dockerconfigjson': buildDockerConfigJson(server, username, password) }
  }
  if (createForm.type === 'kubernetes.io/tls') {
    const crt = String(createTlsForm.crt ?? '')
    const key = String(createTlsForm.key ?? '')
    const result: Record<string, string> = {}
    if (crt.trim()) result['tls.crt'] = crt
    if (key.trim()) result['tls.key'] = key
    return result
  }
  return rowsToStringData(createDataRows.value)
}

const openCreate = () => {
  createForm.clusterId = searchForm.clusterId || ''
  createForm.namespace = searchForm.namespace || ''
  createForm.name = ''
  createForm.type = 'Opaque'
  createForm.alias = ''
  createForm.describe = ''
  createDockerForm.server = ''
  createDockerForm.username = ''
  createDockerForm.password = ''
  createTlsForm.crt = ''
  createTlsForm.key = ''
  createDataRows.value = []
  createLabelRows.value = []
  createAnnotationRows.value = []
  createSection.value = 'data'
  createYamlMode.value = false
  createYaml.value = ''
  createDialogVisible.value = true
}

const resetCreateDataRows = () => {
  createDataRows.value = []
}

const buildCreateYaml = () => {
  const labels = rowsToRecord(createLabelRows.value)
  const annotations: Record<string, string> = {}
  if (createForm.alias) annotations.alias = createForm.alias
  if (createForm.describe) annotations.describe = createForm.describe
  Object.assign(annotations, rowsToRecord(createAnnotationRows.value, { excludeKeys: ['alias', 'describe'] }))
  const stringData = getCreateStringData()
  createYaml.value = buildSecretYamlFromRows(createForm.name, createForm.namespace, createForm.type, labels, annotations, stringData)
  return { labels, annotations, stringData }
}

const toggleCreateYamlMode = () => {
  if (!createYamlMode.value) {
    buildCreateYaml()
    createYamlMode.value = true
    return
  }
  createYamlMode.value = false
}

const openCreateImport = () => {
  createImportInputRef.value?.click()
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
  if (!createForm.clusterId || !createForm.namespace || !createForm.name) {
    ElMessage.warning('请填写必填项')
    return
  }

  if (createForm.type === 'kubernetes.io/dockerconfigjson') {
    if (!(createDockerForm.server || '').trim() || !(createDockerForm.username || '').trim() || !String(createDockerForm.password ?? '').trim()) {
      ElMessage.warning('请填写 docker server、docker username、docker password')
      return
    }
  }
  if (createForm.type === 'kubernetes.io/tls') {
    if (!String(createTlsForm.crt ?? '').trim() || !String(createTlsForm.key ?? '').trim()) {
      ElMessage.warning('请填写 tls.crt 与 tls.key')
      return
    }
  }

  const labels = rowsToRecord(createLabelRows.value)
  const annotations = rowsToRecord(createAnnotationRows.value, { excludeKeys: ['alias', 'describe'] })
  const stringData = getCreateStringData()
  buildCreateYaml()
  createLoading.value = true
  try {
    await secretApi.add({
      clusterId: createForm.clusterId,
      namespace: createForm.namespace,
      name: createForm.name,
      type: createForm.type,
      alias: createForm.alias,
      describe: createForm.describe,
      stringData,
      labels,
      annotations
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

watch(
  () => createForm.type,
  (t) => {
    if (t !== 'Opaque') {
      createDataRows.value = []
    }
    if (t !== 'kubernetes.io/dockerconfigjson') {
      createDockerForm.server = ''
      createDockerForm.username = ''
      createDockerForm.password = ''
    }
    if (t !== 'kubernetes.io/tls') {
      createTlsForm.crt = ''
      createTlsForm.key = ''
    }
    if (createYamlMode.value) buildCreateYaml()
  }
)

const editDialogVisible = ref(false)
const editLoading = ref(false)
const editSection = ref<'data' | 'meta'>('data')
const editYamlMode = ref(false)
const editYaml = ref('')
const kvEditSettingRef = ref<any>()
const editImportInputRef = ref<HTMLInputElement | null>(null)
const editForm = reactive({ type: '' as string, alias: '' as string, describe: '' as string })
const editDataRows = ref<KvRow[]>([])
const editLabelRows = ref<KvRow[]>([])
const editAnnotationRows = ref<KvRow[]>([])

const buildEditYaml = () => {
  if (!currentRow.value) return
  const labels = rowsToRecord(editLabelRows.value)
  const annotations: Record<string, string> = {}
  if (editForm.alias) annotations.alias = editForm.alias
  if (editForm.describe) annotations.describe = editForm.describe
  Object.assign(annotations, rowsToRecord(editAnnotationRows.value, { excludeKeys: ['alias', 'describe'] }))
  const stringData = rowsToStringData(editDataRows.value)
  editYaml.value = buildSecretYamlFromRows(currentRow.value.name, currentRow.value.namespace, editForm.type, labels, annotations, stringData)
}

const toggleEditYamlMode = () => {
  if (!currentRow.value) return
  if (!editYamlMode.value) {
    buildEditYaml()
    editYamlMode.value = true
    return
  }
  editYamlMode.value = false
}

const resetEditDataRows = () => {
  editDataRows.value = rowsFromSecret(currentRow.value)
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

const openEditSetting = async (row: SecretVO) => {
  editLoading.value = true
  try {
    const detail = await fetchSecretDetail(row)
    currentRow.value = { ...row, ...detail }
  } catch {
    currentRow.value = row
  } finally {
    editLoading.value = false
  }

  editForm.type = currentRow.value?.type || 'Opaque'
  editForm.alias = currentRow.value?.alias || ''
  editForm.describe = currentRow.value?.describe || ''

  editDataRows.value = rowsFromSecret(currentRow.value)
  editLabelRows.value = Object.entries(currentRow.value?.labels || {}).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
  editAnnotationRows.value = Object.entries(currentRow.value?.annotations || {})
    .filter(([k]) => k !== 'alias' && k !== 'describe')
    .map(([k, v]) => ({ key: k, value: String(v ?? '') }))

  editSection.value = 'data'
  editYamlMode.value = false
  editYaml.value = ''
  editDialogVisible.value = true
}

const confirmEditSetting = async () => {
  if (!currentRow.value) return
  editLoading.value = true
  try {
    const labels = rowsToRecord(editLabelRows.value)
    const annotations = rowsToRecord(editAnnotationRows.value, { excludeKeys: ['alias', 'describe'] })
    const stringData = rowsToStringData(editDataRows.value)
    buildEditYaml()
    await secretApi.updateInfo({
      clusterId: searchForm.clusterId,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      alias: editForm.alias,
      describe: editForm.describe
    })
    await secretApi.updateSetting({
      clusterId: searchForm.clusterId,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      type: editForm.type,
      labels,
      annotations
    })
    await secretApi.updateData({
      clusterId: searchForm.clusterId,
      namespace: currentRow.value.namespace,
      name: currentRow.value.name,
      stringData
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
    await secretApi.delete({
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
.secret-page {
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
.kv-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
}
.kv-toolbar-left,
.kv-toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.cm-file-input {
  display: none;
}
.cm-meta-kv {
  display: grid;
  gap: 12px;
  margin-top: 12px;
}
.cm-meta-kv__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}
.cm-meta-kv__title {
  font-weight: 600;
  color: #334155;
}
.cm-topbar--edit {
  display: flex;
  justify-content: space-between;
  gap: 12px;
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
