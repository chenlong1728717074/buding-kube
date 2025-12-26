<template>
  <div class="service-list">
    <el-card class="header-card">
      <div class="page-header">
        <h1>Service管理</h1>
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
            placeholder="输入Service名称" 
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
        :data="serviceList" 
        stripe
        style="width: 100%"
      >
        <el-table-column prop="name" label="名称" min-width="120" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="handleViewDetail(row)" :title="row.describe || ''">
              {{ row.name }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="120" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="handleNamespaceDetail(row)">
              {{ row.namespace }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="alias" label="别名" min-width="80" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.alias">{{ row.alias }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" min-width="80" align="center" header-align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="serviceType" label="类型" min-width="100" header-align="center" align="center" />
        <el-table-column prop="internalIP" label="内网IP" min-width="120" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.internalIP">{{ row.internalIP }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="externalIP" label="外网IP" min-width="120" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.externalIP">{{ row.externalIP }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" min-width="140" header-align="center" align="center">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right" header-align="center">
          <template #default="{ row }">
            <div style="display: flex; gap: 6px; align-items: center; justify-content: center; flex-wrap: nowrap;">
              <el-dropdown @command="(command) => handleMoreAction(command, row)" trigger="click">
                <el-button size="small">
                  更多
                  <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="yaml">查看YAML</el-dropdown-item>
                    <el-dropdown-item command="detail">详情</el-dropdown-item>
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
    <el-dialog v-model="viewYamlDialogVisible" title="查看YAML" width="80%" :close-on-click-modal="false" class="config-dialog yaml-dialog">
      <template #header>
        <div class="dialog-header">
          <div>
            <h3 class="dialog-title">查看YAML</h3>
            <div style="margin-top:4px;color:#6b7280;font-size:12px;">Service 配置</div>
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
                <el-descriptions-item label="名称" v-if="currentService">{{ currentService.name }}</el-descriptions-item>
                <el-descriptions-item label="类型">Service</el-descriptions-item>
              </el-descriptions>
            </div>

            <div class="yaml-view">
              <YamlEditor :modelValue="viewYamlForm.yaml" @change="handleYamlChange" :readonly="true" height="100%" />
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
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh, ArrowDown } from '@element-plus/icons-vue'
import { serviceApi, type ServiceVO, type ServiceQueryDTO, type ServiceBaseDTO } from '@/api/service'
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
const serviceList = ref<ServiceVO[]>([])
const viewYamlDialogVisible = ref(false)
const currentService = ref<ServiceVO | null>(null)
const applyLoading = ref(false)

// 查看YAML表单
const viewYamlForm = reactive({
  clusterName: '',
  namespace: '',
  yaml: ''
})

// 渐进式加载的数据获取函数
const namespaceFetcher = computed(() => useNamespaceFetcher(clusterId.value))

// 搜索表单
const searchForm = reactive<ServiceQueryDTO>({
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

// 获取Service列表
const fetchServiceList = async () => {
  if (!clusterId.value) {
    // 没有集群时清空列表
    serviceList.value = []
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
    
    const response = await serviceApi.getServices(params)
    console.log('Service API响应:', response)
    if (response.code === 200) {
      serviceList.value = response.data.items || []
      pagination.total = response.data.total || 0
      // 调试：检查第一个项目的yaml字段
      if (response.data.items && response.data.items.length > 0) {
        console.log('第一个Service的yaml字段:', response.data.items[0].yaml ? '存在' : '不存在')
        console.log('第一个Service数据:', response.data.items[0])
      }
    }
  } catch (error) {
    console.error('获取Service列表失败:', error)
    ElMessage.error('获取Service列表失败')
  } finally {
    loading.value = false
  }
}

// 命名空间变化处理
const handleNamespaceChange = (namespace: string) => {
  searchForm.namespace = namespace
  if (clusterId.value) {
    // 重置分页
    pagination.page = 1
    // 重新获取数据（包括清空namespace的情况）
    fetchServiceList()
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.page = 1
  fetchServiceList()
}

// 重置处理
const handleReset = () => {
  searchForm.namespace = ''
  searchForm.name = ''
  pagination.page = 1
  fetchServiceList()
}

// 分页大小变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchServiceList()
}

// 当前页变化
const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchServiceList()
}

// 更多操作处理
const handleMoreAction = (command: string, row: ServiceVO) => {
  switch (command) {
    case 'yaml':
      handleViewYaml(row)
      break
    case 'detail':
      handleDetail(row)
      break
  }
}

// 查看YAML
const handleViewYaml = async (row: ServiceVO) => {
  if (!row.yaml) {
    ElMessage.error('YAML数据不可用')
    return
  }
  
  viewYamlForm.clusterName = clusterName.value
  viewYamlForm.namespace = row.namespace
  viewYamlForm.yaml = row.yaml
  currentService.value = row
  viewYamlDialogVisible.value = true
}

// 处理YAML内容变化
const handleYamlChange = (newYaml: string) => {
  viewYamlForm.yaml = newYaml
}

// 查看详情
const handleDetail = (row: ServiceVO) => {
  ElMessage.info('详情功能开发中')
}

// 跳转到命名空间详情
const handleNamespaceDetail = (row: ServiceVO) => {
  router.push({
    path: '/namespace/detail',
    query: {
      clusterId: clusterId.value,
      namespace: row.namespace
    }
  })
}

// 时间格式化
const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 获取状态类型
const getStatusType = (status: string) => {
  switch (status) {
    case 'Running':
    case 'Active':
    case 'Ready':
      return 'success'
    case 'Pending':
      return 'warning'
    case 'Failed':
    case 'Error':
      return 'danger'
    default:
      return 'info'
  }
}

// 监听集群变化
watch(clusterId, (newClusterId) => {
  if (newClusterId) {
    // 清空命名空间选择
    searchForm.namespace = ''
    // 重置分页
    pagination.page = 1
    // 重新获取数据
    fetchServiceList()
  } else {
    // 没有集群时清空列表
    serviceList.value = []
    pagination.total = 0
  }
}, { immediate: true })

// 组件挂载时初始化
onMounted(() => {
  // 如果已有集群ID，加载数据
  if (clusterId.value) {
    fetchServiceList()
  }
})
</script>

<style scoped>
.service-list {
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

.header-actions {
  display: flex;
  gap: 12px;
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

:deep(.el-table .el-table__cell) {
  padding: 8px 0;
}

:deep(.el-button + .el-button) {
  margin-left: 6px;
}

.yaml-dialog-content {
  display: flex;
  flex-direction: column;
}

.yaml-info {
  margin-bottom: 16px;
}

.yaml-editor-wrapper {
  flex: 1;
  min-height: 0;
}
</style>
