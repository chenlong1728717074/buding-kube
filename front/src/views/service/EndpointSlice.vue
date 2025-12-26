<template>
  <div class="endpointslice-list">
    <el-card class="header-card">
      <div class="page-header">
        <h1>EndpointSlice管理</h1>
      </div>
    </el-card>

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
    <el-dialog v-model="viewYamlDialogVisible" title="查看YAML" width="90%" :close-on-click-modal="false" class="config-dialog yaml-dialog">
      <template #header>
        <div class="dialog-header">
          <div>
            <h3 class="dialog-title">查看YAML</h3>
            <div style="margin-top:4px;color:#6b7280;font-size:12px;">EndpointSlice 配置</div>
          </div>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <div class="yaml-dialog-content">
            <div class="yaml-info">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="集群">{{ viewYamlForm.clusterName }}</el-descriptions-item>
                <el-descriptions-item label="命名空间">{{ viewYamlForm.namespace }}</el-descriptions-item>
                <el-descriptions-item label="名称">{{ currentEndpointSlice?.name }}</el-descriptions-item>
                <el-descriptions-item label="类型">EndpointSlice</el-descriptions-item>
              </el-descriptions>
            </div>

            <div class="yaml-view">
              <YamlEditor
                v-model="viewYamlForm.yaml"
                :title="`${currentEndpointSlice?.name} - EndpointSlice YAML`"
                :readonly="true"
                height="100%"
                :filename="`${currentEndpointSlice?.name}-endpointslice.yaml`"
              />
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="viewYamlDialogVisible = false">关闭</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh, ArrowDown } from '@element-plus/icons-vue'
import { endpointSliceApi, type EndpointSliceVO, type EndpointSliceQueryDTO } from '@/api/endpointslice'
import { clusterApi } from '@/api/cluster'
import { namespaceApi } from '@/api/namespace'
import { useClusterStore } from '@/stores/cluster'
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import YamlEditor from '@/components/YamlEditor.vue'
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
const endpointSliceList = ref<EndpointSliceVO[]>([])

// 渐进式加载的数据获取函数
const namespaceFetcher = computed(() => useNamespaceFetcher(clusterId.value))

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
  if (!clusterId.value) {
    endpointSliceList.value = []
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
    fetchEndpointSliceList()
  }
}

// 监听集群变化
watch(clusterId, (newClusterId) => {
  if (newClusterId) {
    searchForm.namespace = ''
    pagination.page = 1
    fetchEndpointSliceList()
  } else {
    endpointSliceList.value = []
    pagination.total = 0
  }
}, { immediate: true })

// 页面加载时获取数据
onMounted(() => {
  if (clusterId.value) {
    fetchEndpointSliceList()
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
  viewYamlForm.clusterName = clusterName.value
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
  justify-content: space-between;
  align-items: center;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #303133;
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
}

.yaml-info {
  margin-bottom: 16px;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.yaml-editor-wrapper {
  flex: 1;
  min-height: 0;
}

.dialog-footer {
  text-align: right;
}
</style>
