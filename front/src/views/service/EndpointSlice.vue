<template>
  <div class="endpointslice-list">
    <el-card class="header-card">
      <div class="page-header">
        <h1>EndpointSlice管理</h1>
      </div>
    </el-card>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="集群">
          <InfiniteSelect
            v-model="searchForm.clusterId"
            :fetch-data="clusterFetcher"
            v-bind="clusterSelectConfig"
            style="width: 200px"
            @change="handleClusterChange"
          />
        </el-form-item>
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
            placeholder="输入EndpointSlice名称" 
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
        :data="endpointSliceList" 
        stripe
        style="width: 100%"
      >
        <el-table-column prop="name" label="名称" min-width="120" header-align="center" align="center" />
        <el-table-column prop="namespace" label="命名空间" min-width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="handleNamespaceDetail(row)">
              {{ row.namespace }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="addressType" label="地址类型" min-width="100" header-align="center" align="center" />
        <el-table-column prop="createTime" label="创建时间" min-width="140" header-align="center" align="center">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" header-align="center">
          <template #default="{ row }">
            <div style="display: flex; gap: 6px; align-items: center; justify-content: center; flex-wrap: nowrap;">
              <el-dropdown @command="(command) => handleMoreAction(command, row)">
                <el-button size="small">
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="yaml">查看YAML</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页器 -->
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

    <!-- 查看YAML对话框 -->
    <UnifiedDialog 
      v-model="viewYamlDialogVisible" 
      title="查看YAML" 
      subtitle="EndpointSlice 配置"
      width="90%"
    >
      <div class="yaml-dialog-content">
        <div class="yaml-info">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="集群">{{ viewYamlForm.clusterName }}</el-descriptions-item>
            <el-descriptions-item label="命名空间">{{ viewYamlForm.namespace }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ currentEndpointSlice?.name }}</el-descriptions-item>
            <el-descriptions-item label="类型">EndpointSlice</el-descriptions-item>
          </el-descriptions>
        </div>
        
        <div class="yaml-editor-wrapper">
          <YamlEditor
            v-model="viewYamlForm.yaml"
            :title="`${currentEndpointSlice?.name} - EndpointSlice YAML`"
            :readonly="true"
            height="500px"
            :filename="`${currentEndpointSlice?.name}-endpointslice.yaml`"
          />
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="viewYamlDialogVisible = false">关闭</el-button>
        </div>
      </template>
    </UnifiedDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh, ArrowDown } from '@element-plus/icons-vue'
import { endpointSliceApi, type EndpointSliceVO, type EndpointSliceQueryDTO } from '@/api/endpointslice'
import { clusterApi } from '@/api/cluster'
import { namespaceApi } from '@/api/namespace'
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import UnifiedDialog from '@/components/UnifiedDialog.vue'
import { useClusterFetcher, useNamespaceFetcher, clusterSelectConfig, namespaceSelectConfig } from '@/composables/useInfiniteSelect'

// 路由
const router = useRouter()

// 响应式数据
const loading = ref(false)
const endpointSliceList = ref<EndpointSliceVO[]>([])

// 渐进式加载的数据获取函数
const clusterFetcher = useClusterFetcher()
const namespaceFetcher = computed(() => useNamespaceFetcher(searchForm.clusterId))

// 对话框状态
const viewYamlDialogVisible = ref(false)
const currentEndpointSlice = ref<EndpointSliceVO | null>(null)

// 查看YAML表单
const viewYamlForm = reactive({
  clusterName: '',
  namespace: '',
  yaml: ''
})

// 搜索表单
const searchForm = reactive<EndpointSliceQueryDTO>({
  clusterId: '',
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

// 获取EndpointSlice列表
const fetchEndpointSliceList = async () => {
  if (!searchForm.clusterId) {
    // 没有集群时清空列表
    endpointSliceList.value = []
    pagination.total = 0
    return
  }
  
  loading.value = true
  try {
    const params = {
      ...searchForm,
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    const response = await endpointSliceApi.getEndpointSliceList(params)
    if (response.code === 200) {
      endpointSliceList.value = response.data.items || []
      pagination.total = response.data.total || 0
    }
  } catch (error) {
    console.error('获取EndpointSlice列表失败:', error)
    ElMessage.error('获取EndpointSlice列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.page = 1
  fetchEndpointSliceList()
}

// 重置处理
const handleReset = () => {
  searchForm.namespace = ''
  searchForm.name = ''
  pagination.page = 1
  fetchEndpointSliceList()
}

// 分页大小变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchEndpointSliceList()
}

// 当前页变化
const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchEndpointSliceList()
}

// 跳转到命名空间详情
const handleNamespaceDetail = (row: EndpointSliceVO) => {
  router.push({
    path: '/namespace/detail',
    query: {
      clusterId: searchForm.clusterId,
      namespace: row.namespace
    }
  })
}

// 格式化时间
const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 处理集群变化
const handleClusterChange = (clusterId: string) => {
  searchForm.clusterId = clusterId
  searchForm.namespace = '' // 重置命名空间选择
  if (clusterId) {
    // 命名空间下拉框会自动重新加载数据
    // 如果有集群选择，立即加载EndpointSlice列表
    fetchEndpointSliceList()
  } else {
    // 清空列表
    endpointSliceList.value = []
    pagination.total = 0
  }
}

// 处理命名空间变化
const handleNamespaceChange = (namespace: string) => {
  searchForm.namespace = namespace
  // 只要有集群选择，就重新查询（无论命名空间是否为空）
  if (searchForm.clusterId) {
    fetchEndpointSliceList()
  }
}

// 页面加载时获取数据
onMounted(async () => {
  // 自动选择第一个集群并查询数据
  try {
    // 获取第一个集群
    const clusterResponse = await clusterApi.getClusters({ page: 1, pageSize: 1 })
    if (clusterResponse.code === 200 && clusterResponse.data.items && clusterResponse.data.items.length > 0) {
      const firstCluster = clusterResponse.data.items[0]
      searchForm.clusterId = firstCluster.id
      
      // 自动加载EndpointSlice列表
      fetchEndpointSliceList()
    }
  } catch (error) {
    console.error('自动选择集群失败:', error)
    // 如果自动选择失败，用户仍可以手动选择
  }
})

// 更多操作处理
const handleMoreAction = (command: string, row: EndpointSliceVO) => {
  switch (command) {
    case 'yaml':
      handleViewYaml(row)
      break
  }
}

// 查看YAML
const handleViewYaml = async (row: EndpointSliceVO) => {
  currentEndpointSlice.value = row
  // 获取集群名称
  try {
    const response = await clusterApi.getCluster(searchForm.clusterId)
    viewYamlForm.clusterName = response.data?.name || ''
  } catch (error) {
    viewYamlForm.clusterName = searchForm.clusterId
  }
  viewYamlForm.namespace = row.namespace
  viewYamlForm.yaml = row.yaml || '# YAML内容加载中...'
  viewYamlDialogVisible.value = true
}
</script>

<style scoped>
.endpointslice-list {
  padding: 20px;
}

.header-card {
  margin-bottom: 20px;
}

.page-header {
  display: flex;
  flex-direction: column;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
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
}

.yaml-editor-wrapper {
  flex: 1;
}

.dialog-footer {
  text-align: right;
}
</style>
