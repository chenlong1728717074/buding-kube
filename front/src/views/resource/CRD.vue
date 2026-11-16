<template>
  <div class="crd-page">
    <div class="page-header">
      <h1>CRD管理</h1>
      <div class="header-actions">
        <el-button @click="loadCrds">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="关键词">
          <el-input
            v-model="searchForm.keyword"
            placeholder="搜索名称 / Kind / 组"
            clearable
            style="width: 260px"
            @clear="handleSearch"
            @keyup.enter="handleSearch"
          />
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
      <el-table v-loading="loading" :data="crds" stripe style="width: 100%">
        <el-table-column prop="name" label="名称" min-width="280" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="openDetail(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="group" label="组" min-width="180" :show-overflow-tooltip="true" header-align="center" align="center" />
        <el-table-column prop="kind" label="Kind" min-width="140" header-align="center" align="center" />
        <el-table-column prop="version" label="存储版本" min-width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag size="small" type="info">{{ row.version || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="scope" label="作用域" min-width="110" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag size="small" :type="row.scope === 'Cluster' ? 'warning' : 'success'">
              {{ row.scope === 'Cluster' ? '集群级' : '命名空间级' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" min-width="160" header-align="center" align="center">
          <template #default="{ row }">{{ row.createTime || '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" header-align="center" align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" class="btn-main" @click="openDetail(row)">详情</el-button>
              <el-dropdown trigger="click" @command="(cmd) => handleRowAction(cmd, row)">
                <el-button size="small" text circle class="more-btn">
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="yaml">查看YAML</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-button size="small" type="danger" plain class="btn-danger" @click="openDelete(row)">删除</el-button>
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

    <!-- CRD 详情对话框(结构化信息) -->
    <el-dialog v-model="detailVisible" :title="`CRD详情 - ${currentCrd?.name || ''}`" width="720px" top="8vh">
      <div v-loading="detailLoading" class="detail-wrapper">
        <div class="detail-grid">
          <div class="detail-item"><span class="label">名称</span><span class="value">{{ detailData.name }}</span></div>
          <div class="detail-item"><span class="label">组</span><span class="value">{{ detailData.group }}</span></div>
          <div class="detail-item"><span class="label">Kind</span><span class="value">{{ detailData.kind }}</span></div>
          <div class="detail-item"><span class="label">List Kind</span><span class="value">{{ detailData.listKind }}</span></div>
          <div class="detail-item"><span class="label">存储版本</span><span class="value">{{ detailData.version }}</span></div>
          <div class="detail-item"><span class="label">作用域</span><span class="value">{{ detailData.scope === 'Cluster' ? '集群级' : '命名空间级' }}</span></div>
          <div class="detail-item"><span class="label">复数名</span><span class="value">{{ detailData.plural }}</span></div>
          <div class="detail-item"><span class="label">单数名</span><span class="value">{{ detailData.singular }}</span></div>
          <div class="detail-item"><span class="label">短名</span><span class="value">{{ detailData.shortNames?.length ? detailData.shortNames.join(', ') : '-' }}</span></div>
          <div class="detail-item"><span class="label">创建时间</span><span class="value">{{ detailData.createTime }}</span></div>
        </div>
        <div v-if="detailData.versions?.length" class="verbs-section">
          <span class="label">全部版本</span>
          <div class="verb-tags">
            <el-tag v-for="ver in detailData.versions" :key="ver" size="small" type="info">{{ ver }}</el-tag>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button v-if="currentCrd" type="primary" @click="openYaml(currentCrd)">查看YAML</el-button>
      </template>
    </el-dialog>

    <!-- CRD YAML 对话框 -->
    <el-dialog v-model="yamlVisible" :title="`YAML - ${currentCrd?.name || ''}`" width="70%" top="5vh">
      <div v-loading="yamlLoading" class="yaml-wrapper">
        <YamlEditor :model-value="yamlContent" :readonly="true" height="100%" class="yaml-editor" />
      </div>
      <template #footer>
        <el-button @click="yamlVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, MoreFilled } from '@element-plus/icons-vue'
import { useClusterStore } from '@/stores/cluster'
import { crdApi, type CrdVO } from '@/api/crd'
import { jsonToYaml } from '@/utils/yaml'
import YamlEditor from '@/components/YamlEditor.vue'
import '@/assets/styles/config-editor.css'

const clusterStore = useClusterStore()

const loading = ref(false)
const crds = ref<CrdVO[]>([])
const searchForm = reactive({ keyword: '' })
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// YAML 详情
const yamlVisible = ref(false)
const yamlLoading = ref(false)
const yamlContent = ref('')
const currentCrd = ref<CrdVO | null>(null)

// 结构化详情
const detailVisible = ref(false)
const detailLoading = ref(false)
const detailData = ref<Partial<CrdVO> & { listKind?: string; plural?: string; singular?: string; shortNames?: string[]; versions?: string[] }>({})

// 行操作分发
const handleRowAction = (command: string, row: CrdVO) => {
  if (command === 'yaml') {
    openYaml(row)
  }
}

// 打开结构化详情
const openDetail = async (row: CrdVO) => {
  const clusterId = clusterStore.currentClusterId
  if (!clusterId) return
  currentCrd.value = row
  detailVisible.value = true
  detailLoading.value = true
  detailData.value = {
    name: row.name,
    group: row.group,
    kind: row.kind,
    version: row.version,
    scope: row.scope,
    createTime: row.createTime
  }
  try {
    const response = await crdApi.getCrd(clusterId, row.name)
    const data = response.data as any
    const meta = data.metadata || {}
    detailData.value = {
      ...detailData.value,
      plural: data.spec?.names?.plural || '-',
      singular: data.spec?.names?.singular || '-',
      listKind: data.spec?.names?.listKind || '-',
      shortNames: data.spec?.names?.shortNames || [],
      versions: data.spec?.versions?.map((v: any) => v.name).filter(Boolean) || [],
      createTime: meta.creationTimestamp ? String(meta.creationTimestamp).replace('T', ' ').slice(0, 19) : row.createTime
    }
  } catch (error) {
    console.error('Failed to load CRD detail:', error)
    // 获取失败时仍展示列表已知信息
  } finally {
    detailLoading.value = false
  }
}

// 加载 CRD 列表
const loadCrds = async () => {
  const clusterId = clusterStore.currentClusterId
  if (!clusterId) {
    ElMessage.warning('请先选择集群')
    return
  }
  loading.value = true
  try {
    const response = await crdApi.getCrds({
      clusterId,
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword || undefined
    })
    crds.value = response.data?.items || []
    pagination.total = response.data?.total || 0
  } catch (error) {
    console.error('Failed to load CRDs:', error)
    ElMessage.error('加载CRD列表失败')
    crds.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadCrds()
}

const handleReset = () => {
  searchForm.keyword = ''
  pagination.page = 1
  loadCrds()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  loadCrds()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  loadCrds()
}

// 查看 CRD YAML 详情
const openYaml = async (row: CrdVO) => {
  const clusterId = clusterStore.currentClusterId
  if (!clusterId) return
  currentCrd.value = row
  yamlVisible.value = true
  yamlLoading.value = true
  yamlContent.value = ''
  try {
    const response = await crdApi.getCrd(clusterId, row.name)
    const data = response.data as Record<string, unknown>
    // managedFields 为 k8s 追踪字段, 体积大且无展示价值
    const meta = data.metadata as Record<string, unknown> | undefined
    if (meta) delete meta.managedFields
    yamlContent.value = jsonToYaml(data)
  } catch (error) {
    console.error('Failed to load CRD detail:', error)
    ElMessage.error('获取CRD详情失败')
    yamlVisible.value = false
  } finally {
    yamlLoading.value = false
  }
}

// 删除 CRD
const openDelete = (row: CrdVO) => {
  ElMessageBox.confirm(
    `确定要删除 CRD "${row.name}" 吗？该 CRD 下的所有自定义资源将一并删除，此操作不可恢复！`,
    '危险操作',
    {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      confirmButtonClass: 'el-button--danger'
    }
  ).then(async () => {
    try {
      await crdApi.deleteCrd(clusterStore.currentClusterId, row.name)
      ElMessage.success('删除成功')
      loadCrds()
    } catch (error) {
      console.error('Failed to delete CRD:', error)
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

onMounted(() => {
  loadCrds()
})
</script>

<style scoped>
.crd-page {
  padding: 20px;
}

.search-card {
  margin-bottom: 16px;
}

.yaml-wrapper {
  height: 60vh;
  overflow: auto;
}

.yaml-editor {
  height: 100%;
}

.detail-wrapper {
  min-height: 200px;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px 24px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.detail-item .label,
.verbs-section .label {
  font-size: 12px;
  color: #6b7280;
}

.detail-item .value {
  font-size: 14px;
  color: #1f2937;
  word-break: break-all;
}

.verbs-section {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}

.verb-tags {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
