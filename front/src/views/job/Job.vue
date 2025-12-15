<template>
  <div class="job-list">
    <div class="page-header">
      <h1>Job管理</h1>
    </div>

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
            placeholder="输入Job名称" 
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
        :data="jobList" 
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
        <el-table-column prop="namespace" label="命名空间" min-width="100" header-align="center" align="center">
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
        <el-table-column prop="lastRunTime" label="上次运行时间" min-width="140" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.lastRunTime">{{ formatTime(row.lastRunTime) }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" min-width="140" header-align="center" align="center">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right" header-align="center">
          <template #default="{ row }">
            <div style="display: flex; gap: 6px; align-items: center; justify-content: center; flex-wrap: nowrap;">
              <el-button size="small" @click="handleDetail(row)">
                详情
              </el-button>
              <el-dropdown @command="(command) => handleMoreAction(command, row)">
                <el-button size="small">
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="yaml">查看YAML</el-dropdown-item>
                    <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
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
      subtitle="Job 配置"
      width="90%"
    >
      <div class="yaml-dialog-content">
        <div class="yaml-info">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="集群">{{ viewYamlForm.clusterName }}</el-descriptions-item>
            <el-descriptions-item label="命名空间">{{ viewYamlForm.namespace }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ currentJob?.name }}</el-descriptions-item>
            <el-descriptions-item label="类型">Job</el-descriptions-item>
          </el-descriptions>
        </div>
        
        <div class="yaml-editor-wrapper">
          <YamlEditor
            v-model="viewYamlForm.yaml"
            :title="`${currentJob?.name} - Job YAML`"
            :readonly="true"
            height="500px"
            :filename="`${currentJob?.name}-job.yaml`"
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, ArrowDown } from '@element-plus/icons-vue'
import { jobApi, type JobVO, type JobQueryDTO } from '@/api/job'
import { clusterApi } from '@/api/cluster'
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import UnifiedDialog from '@/components/UnifiedDialog.vue'
import { useClusterFetcher, useNamespaceFetcher, clusterSelectConfig, namespaceSelectConfig } from '@/composables/useInfiniteSelect'

// 路由
const router = useRouter()

// 响应式数据
const loading = ref(false)
const jobList = ref<JobVO[]>([])

// 渐进式加载的数据获取函数
const clusterFetcher = useClusterFetcher()
const namespaceFetcher = computed(() => useNamespaceFetcher(searchForm.clusterId))

// 对话框状态
const viewYamlDialogVisible = ref(false)
const currentJob = ref<JobVO | null>(null)

// 查看YAML表单
const viewYamlForm = reactive({
  clusterName: '',
  namespace: '',
  yaml: ''
})

// 搜索表单
const searchForm = reactive<JobQueryDTO>({
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

// 获取Job列表
const fetchJobList = async () => {
  if (!searchForm.clusterId) {
    // 没有集群时清空列表
    jobList.value = []
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
    const response = await jobApi.getJobs(params)
    if (response.code === 200) {
      jobList.value = response.data.items || []
      pagination.total = response.data.total || 0
    }
  } catch (error) {
    console.error('获取Job列表失败:', error)
    ElMessage.error('获取Job列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索处理
const handleSearch = () => {
  pagination.page = 1
  fetchJobList()
}

// 重置处理
const handleReset = () => {
  searchForm.namespace = ''
  searchForm.name = ''
  pagination.page = 1
  fetchJobList()
}

// 分页大小变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchJobList()
}

// 当前页变化
const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchJobList()
}

// 查看详情
const handleViewDetail = (row: JobVO) => {
  ElMessage.info('查看详情功能开发中')
}

// 跳转到命名空间详情
const handleNamespaceDetail = (row: JobVO) => {
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
    fetchJobList()
  } else {
    jobList.value = []
    pagination.total = 0
  }
}

// 处理命名空间变化
const handleNamespaceChange = (namespace: string) => {
  searchForm.namespace = namespace
  if (searchForm.clusterId) {
    fetchJobList()
  }
}

// 详情处理
const handleDetail = (row: JobVO) => {
  ElMessage.info('详情功能开发中')
}

// 更多操作处理
const handleMoreAction = (command: string, row: JobVO) => {
  switch (command) {
    case 'yaml':
      handleViewYaml(row)
      break
    case 'delete':
      handleDelete(row)
      break
  }
}

// 查看YAML
const handleViewYaml = (row: JobVO) => {
  currentJob.value = row
  viewYamlForm.clusterName = searchForm.clusterId
  viewYamlForm.namespace = row.namespace
  viewYamlForm.yaml = row.yaml || ''
  viewYamlDialogVisible.value = true
}

// 删除Job
const handleDelete = (row: JobVO) => {
  ElMessageBox.confirm(
    `确定要删除Job "${row.name}" 吗？`,
    '确认删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(() => {
    ElMessage.info('删除功能开发中')
  }).catch(() => {
    // 用户取消删除
  })
}

// 页面挂载时的处理
onMounted(async () => {
  // 自动选择第一个集群并查询数据
  try {
    // 获取第一个集群
    const clusterResponse = await clusterApi.getClusters({ page: 1, pageSize: 1 })
    if (clusterResponse.code === 200 && clusterResponse.data.items && clusterResponse.data.items.length > 0) {
      const firstCluster = clusterResponse.data.items[0]
      searchForm.clusterId = firstCluster.id
      
      // 自动加载Job列表
      fetchJobList()
    }
  } catch (error) {
    console.error('自动选择集群失败:', error)
    // 如果自动选择失败，用户仍可以手动选择
  }
})
</script>

<style scoped>
.job-list {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
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
