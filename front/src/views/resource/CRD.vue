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
            <el-link type="primary" @click="openYaml(row)">{{ row.name }}</el-link>
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
        <el-table-column label="操作" width="160" fixed="right" header-align="center" align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" class="btn-main" @click="openYaml(row)">详情</el-button>
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

    <!-- CRD YAML 详情对话框 -->
    <el-dialog v-model="yamlVisible" :title="`CRD详情 - ${currentCrd?.name || ''}`" width="70%" top="5vh">
      <div v-loading="yamlLoading" class="yaml-wrapper">
        <pre class="yaml-pre">{{ yamlContent }}</pre>
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
import { Search, Refresh } from '@element-plus/icons-vue'
import { useClusterStore } from '@/stores/cluster'
import { crdApi, type CrdVO } from '@/api/crd'
import { jsonToYaml } from '@/utils/yaml'

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
  max-height: 65vh;
  overflow: auto;
}

.yaml-pre {
  margin: 0;
  padding: 16px;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
  background: #f8fafc;
  border-radius: 8px;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
