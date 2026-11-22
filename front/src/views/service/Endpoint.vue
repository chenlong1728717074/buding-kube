<template>
  <div class="endpoint-list list-page">
    <PageHeader title="Endpoint管理" subtitle="Service 的后端地址列表(已过时)" icon="ep" deprecated>
      <el-button type="success" @click="openYamlAdd">
        <el-icon><Document /></el-icon>
        YAML添加
      </el-button>

      <template #extra>
        <el-alert
          title="功能已过时"
          type="warning"
          description="Endpoint功能已过时，建议使用EndpointSlice功能。"
          show-icon
          :closable="false"
        />
      </template>
    </PageHeader>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="命名空间">
          <InfiniteSelect
            v-model="searchForm.namespace"
            :fetch-data="namespaceFetcher"
            v-bind="namespaceSelectConfig"
            style="width: 200px"
            @change="handleNamespaceChange"
          />
        </el-form-item>
        <el-form-item label="名称">
          <el-input 
            v-model="searchForm.name" 
            placeholder="输入Endpoint名称" 
            clearable
            style="width: 200px"
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
      <el-table 
        v-loading="loading" 
        :data="endpointList"
        style="width: 100%"
        :header-cell-class-name="tableHeaderClass"
        :cell-class-name="tableCellClass"
      >
        <el-table-column prop="name" label="名称" min-width="120" header-align="center" align="center">
          <template #default="{ row }">
            <span class="deprecated-text">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" min-width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="handleNamespaceDetail(row)">
              {{ row.namespace }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" min-width="140" header-align="center" align="center">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" header-align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" class="btn-main" @click="handleViewDetail(row)">详情</el-button>
              <el-dropdown trigger="click" @command="(command: string) => handleMoreAction(command, row)">
                <el-button size="small" text circle class="more-btn">
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="yaml">查看YAML</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-button size="small" type="danger" plain class="btn-danger" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页器 -->
      <ListPagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>

    <!-- YAML添加对话框 -->
    <YamlDialog
      v-model="yamlAddDialogVisible"
      title="YAML添加Endpoint"
      subtitle="已内置示例，可直接修改后使用"
      v-model:content="yamlAddContent"
      :readonly="false"
      filename="endpoint.yaml"
      :submitting="yamlAddLoading"
      @confirm="confirmYamlAdd"
    />

    <!-- 查看YAML对话框 -->
    <YamlDialog
      v-model="viewYamlDialogVisible"
      title="编辑YAML"
      subtitle="Endpoint 配置"
      :info="[{ label: '集群', value: viewYamlForm.clusterName }, { label: '命名空间', value: viewYamlForm.namespace }, { label: '名称', value: currentEndpoint?.name }, { label: '类型', value: 'Endpoint' }]"
      v-model:content="viewYamlForm.yaml"
      :editor-title="`${currentEndpoint?.name} - Endpoint YAML`"
      :filename="`${currentEndpoint?.name}-endpoint.yaml`"
      :apply-target="currentEndpoint ? { clusterId, kind: 'endpoints', namespace: currentEndpoint.namespace, name: currentEndpoint.name } : null"
      @applied="fetchEndpointList"
    />
  </div>
</template>

<script setup lang="ts">
import { fetchResourceYaml } from '@/api/resource'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, MoreFilled, Document } from '@element-plus/icons-vue'
import { endpointApi, type EndpointVO, type EndpointQueryDTO } from '@/api/endpoint'

import { useClusterStore } from '@/stores/cluster'
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import { useNamespaceFetcher, namespaceSelectConfig } from '@/composables/useInfiniteSelect'
import '@/assets/styles/config-editor.css'

// 路由
const router = useRouter()

// 集群上下文
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)
const clusterName = computed(() => clusterStore.currentClusterName)

// 响应式数据
const loading = ref(false)
const endpointList = ref<EndpointVO[]>([])

// 渐进式加载的数据获取函数
const namespaceFetcher = computed(() => useNamespaceFetcher(clusterId.value))

// 对话框状态
const viewYamlDialogVisible = ref(false)
const currentEndpoint = ref<EndpointVO | null>(null)
const yamlAddDialogVisible = ref(false)
const yamlAddLoading = ref(false)
const yamlAddContent = ref('')

// 查看YAML表单
const viewYamlForm = reactive({
  clusterName: '',
  namespace: '',
  yaml: ''
})

// 搜索表单
const searchForm = reactive<Omit<EndpointQueryDTO, 'clusterId'>>({
  namespace: '',
  name: '',
  page: 1,
  pageSize: 10
})

// 分页信息
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 获取Endpoint列表
const fetchEndpointList = async () => {
  if (!clusterId.value) {
    // 没有集群时清空列表
    endpointList.value = []
    pagination.total = 0
    return
  }
  
  loading.value = true
  try {
    const params = {
      ...searchForm,
      clusterId: clusterId.value,
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    const response = await endpointApi.getEndpointList(params)
    if (response.code === 200) {
      endpointList.value = response.data.items || []
      pagination.total = response.data.total || 0
    }
  } catch (error) {
    console.error('获取Endpoint列表失败:', error)
    ElMessage.error('获取Endpoint列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.page = 1
  fetchEndpointList()
}

// 重置处理
const handleReset = () => {
  searchForm.namespace = ''
  searchForm.name = ''
  pagination.page = 1
  fetchEndpointList()
}

// 分页大小变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchEndpointList()
}

// 当前页变化
const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchEndpointList()
}

// 跳转到命名空间详情
const handleNamespaceDetail = (row: EndpointVO) => {
  router.push({
    path: `/cluster/${clusterId.value}/namespace/detail`,
    query: {
      clusterId: clusterId.value,
      namespace: row.namespace
    }
  })
}

// 格式化时间
const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 处理命名空间变化
const handleNamespaceChange = (namespace: string) => {
  searchForm.namespace = namespace
  if (clusterId.value) {
    fetchEndpointList()
  }
}

// 监听集群变化
watch(clusterId, (newClusterId) => {
  if (newClusterId) {
    searchForm.namespace = ''
    pagination.page = 1
    fetchEndpointList()
  } else {
    endpointList.value = []
    pagination.total = 0
  }
}, { immediate: true })

// 页面加载时获取数据
onMounted(() => {
  if (clusterId.value) {
    fetchEndpointList()
  }
})

// 更多操作处理
const handleMoreAction = (command: string, row: EndpointVO) => {
  switch (command) {
    case 'yaml':
      handleViewYaml(row)
      break
  }
}

// 详情(暂未实现, 与全站统一)
const handleViewDetail = (row: EndpointVO) => {
  ElMessage.info('详情功能暂未实现')
}

// 删除Endpoint(Endpoint 由控制器管理, 删除后可能被自动重建)
const handleDelete = (row: EndpointVO) => {
  ElMessageBox.confirm(
    `确定要删除 Endpoint "${row.name}" 吗？该对象通常由控制器自动管理，删除后可能被重建。`,
    '确认删除',
    {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    }
  ).then(async () => {
    try {
      await endpointApi.deleteEndpoint({
        clusterId: clusterId.value,
        namespace: row.namespace,
        name: row.name
      })
      ElMessage.success('删除成功')
      fetchEndpointList()
    } catch (error) {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

const buildEndpointYamlExample = () => {
  const namespace = searchForm.namespace || 'default'
  return [
    'apiVersion: v1',
    'kind: Endpoints',
    'metadata:',
    '  name: example-endpoint',
    `  namespace: ${namespace}`,
    'subsets:',
    '  - addresses:',
    '      - ip: 10.0.0.10',
    '    ports:',
    '      - port: 80',
    '        protocol: TCP'
  ].join('\n')
}

const openYamlAdd = () => {
  yamlAddContent.value = buildEndpointYamlExample()
  yamlAddDialogVisible.value = true
}

const confirmYamlAdd = async () => {
  if (!yamlAddContent.value.trim()) {
    ElMessage.warning('请输入YAML内容')
    return
  }
  yamlAddLoading.value = true
  try {
    ElMessage.warning('当前版本后端暂未开放Endpoint YAML应用接口，请先复制示例使用kubectl apply执行')
  } finally {
    yamlAddLoading.value = false
  }
}

// 查看YAML
const handleViewYaml = async (row: EndpointVO) => {
  let yamlText = ''
  try {
    yamlText = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'endpoints', namespace: row.namespace, name: row.name })
  } catch {
    return
  }
  currentEndpoint.value = row
  viewYamlForm.clusterName = clusterName.value
  viewYamlForm.namespace = row.namespace
  viewYamlForm.yaml = yamlText
  viewYamlDialogVisible.value = true
}
</script>

<style scoped>
.endpoint-list {
  padding: 20px;
}

.header-card {
  margin-bottom: 20px;
  text-align: left;
}

.header-card :deep(.el-card__body) {
  display: flex;
  justify-content: flex-start;
  text-align: left;
}

.page-header {
  display: flex;
  flex-direction: column;
  width: 100%;
}

.page-header-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  width: 100%;
  gap: 12px;
}

.deprecated-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  text-decoration: line-through;
  color: #909399;
  text-align: left;
}

.deprecated-text {
  text-decoration: line-through;
  color: #909399;
}

.deprecated-button {
  text-decoration: line-through;
}

.header-actions {
  display: flex;
  gap: 12px;
  flex-shrink: 0;
}

.search-card {
  margin-bottom: 20px;
}

.table-card {
  min-height: 400px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.yaml-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.yaml-info {
  margin-bottom: 16px;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.yaml-editor-wrapper {
  flex: 1;
}

.dialog-footer {
  text-align: right;
}
.action-buttons {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  flex-wrap: nowrap;
}

.btn-more {
  color: #475569;
  border-color: #d5deea;
  background: #fff;
}

.btn-more:hover,
.btn-more:focus {
  color: #334155;
  border-color: #b8c4d5;
  background: #f8fafc;
}

:deep(.action-buttons .el-button) {
  height: 28px;
  padding: 0 10px;
  border-radius: 8px;
}

</style>
