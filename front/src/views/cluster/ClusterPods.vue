<template>
  <div class="cluster-pods">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="handleBack">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1>Pod管理</h1>
        <el-tag v-if="clusterDisplayName" type="info">{{ clusterDisplayName }}</el-tag>
        <el-tag v-if="nodeName" type="warning">节点: {{ nodeName }}</el-tag>
      </div>
      <div class="header-right">
        <el-button @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 搜索表单 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="Pod名称">
          <el-input 
            v-model="searchForm.name" 
            placeholder="请输入Pod名称" 
            clearable
            style="width: 200px;"
          />
        </el-form-item>
        <el-form-item label="命名空间">
          <el-select 
            v-model="searchForm.namespace" 
            placeholder="请选择命名空间" 
            clearable
            style="width: 180px;"
          >
            <el-option 
              v-for="ns in namespaces" 
              :key="ns" 
              :label="ns" 
              :value="ns" 
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select 
            v-model="searchForm.status" 
            placeholder="请选择状态" 
            clearable
            style="width: 150px;"
          >
            <el-option label="运行中" value="Running" />
            <el-option label="等待中" value="Pending" />
            <el-option label="成功" value="Succeeded" />
            <el-option label="失败" value="Failed" />
            <el-option label="未知" value="Unknown" />
          </el-select>
        </el-form-item>
        <el-form-item label="节点">
          <el-select 
            v-model="searchForm.node" 
            placeholder="请选择节点" 
            clearable
            style="width: 180px;"
          >
            <el-option 
              v-for="node in nodes" 
              :key="node" 
              :label="node" 
              :value="node" 
            />
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

    <!-- Pod列表 -->
    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>Pod列表 ({{ pagination.total }})</span>
        </div>
      </template>
      
      <el-table 
        v-loading="loading" 
        :data="podList" 
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="name" label="Pod名称" min-width="200">
          <template #default="{ row }">
            <div class="pod-name">
              <el-icon class="pod-icon"><Box /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" width="150" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ready" label="就绪状态" width="100">
          <template #default="{ row }">
            <span :class="{ 'ready': row.ready, 'not-ready': !row.ready }">
              {{ row.readyContainers }}/{{ row.totalContainers }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="restarts" label="重启次数" width="100">
          <template #default="{ row }">
            <span :class="{ 'high-restarts': row.restarts > 5 }">
              {{ row.restarts || 0 }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="node" label="节点" width="150">
          <template #default="{ row }">
            <el-button v-if="row.node" text @click="handleViewNode(row.node)">
              {{ row.node }}
            </el-button>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="资源使用" width="200">
          <template #default="{ row }">
            <div class="resource-usage">
              <div class="usage-item">
                <span class="usage-label">CPU:</span>
                <span class="usage-value">{{ row.cpuUsage || '-' }}</span>
              </div>
              <div class="usage-item">
                <span class="usage-label">内存:</span>
                <span class="usage-value">{{ row.memoryUsage || '-' }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="age" label="运行时间" width="120" />
        <el-table-column prop="createTime" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewDetail(row)">
              详情
            </el-button>
            <el-button size="small" @click="handleViewLogs(row)">
              日志
            </el-button>
            <el-dropdown @command="(command) => handleMoreActions(command, row)">
              <el-button size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="yaml">
                    查看YAML
                  </el-dropdown-item>
                  <el-dropdown-item command="events">
                    查看事件
                  </el-dropdown-item>
                  <el-dropdown-item command="exec">
                    进入容器
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" divided>
                    删除Pod
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- Pod详情对话框 -->
    <el-dialog 
      v-model="detailDialogVisible" 
      title="Pod详情" 
      width="80%"
      :before-close="handleDetailDialogClose"
    >
      <div v-if="selectedPod" class="pod-detail">
        <!-- 基本信息 -->
        <el-descriptions title="基本信息" :column="2" border>
          <el-descriptions-item label="Pod名称">{{ selectedPod.name }}</el-descriptions-item>
          <el-descriptions-item label="命名空间">{{ selectedPod.namespace }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(selectedPod.status)">
              {{ getStatusText(selectedPod.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="就绪状态">
            <span :class="{ 'ready': selectedPod.ready, 'not-ready': !selectedPod.ready }">
              {{ selectedPod.readyContainers }}/{{ selectedPod.totalContainers }}
            </span>
          </el-descriptions-item>
          <el-descriptions-item label="重启次数">{{ selectedPod.restarts || 0 }}</el-descriptions-item>
          <el-descriptions-item label="节点">{{ selectedPod.node || '-' }}</el-descriptions-item>
          <el-descriptions-item label="Pod IP">{{ selectedPod.podIP || '-' }}</el-descriptions-item>
          <el-descriptions-item label="主机IP">{{ selectedPod.hostIP || '-' }}</el-descriptions-item>
          <el-descriptions-item label="QoS类别">{{ selectedPod.qosClass || '-' }}</el-descriptions-item>
          <el-descriptions-item label="运行时间">{{ selectedPod.age }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ selectedPod.createTime }}</el-descriptions-item>
        </el-descriptions>

        <!-- 容器信息 -->
        <div style="margin-top: 20px;">
          <h4>容器信息</h4>
          <el-table :data="selectedPod.containers" border>
            <el-table-column prop="name" label="容器名称" />
            <el-table-column prop="image" label="镜像" show-overflow-tooltip />
            <el-table-column prop="state" label="状态">
              <template #default="{ row }">
                <el-tag :type="getContainerStateType(row.state)">
                  {{ row.state }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="ready" label="就绪">
              <template #default="{ row }">
                <el-tag :type="row.ready ? 'success' : 'danger'" size="small">
                  {{ row.ready ? '是' : '否' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="restartCount" label="重启次数" />
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button size="small" @click="handleContainerLogs(row)">
                  日志
                </el-button>
                <el-button size="small" @click="handleContainerExec(row)">
                  终端
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 标签 -->
        <div style="margin-top: 20px;">
          <h4>标签</h4>
          <div class="labels-container">
            <el-tag 
              v-for="(value, key) in selectedPod.labels" 
              :key="key"
              style="margin: 4px;"
            >
              {{ key }}: {{ value }}
            </el-tag>
            <span v-if="!selectedPod.labels || Object.keys(selectedPod.labels).length === 0">无标签</span>
          </div>
        </div>

        <!-- 注解 -->
        <div style="margin-top: 20px;">
          <h4>注解</h4>
          <div class="annotations-container">
            <el-tag 
              v-for="(value, key) in selectedPod.annotations" 
              :key="key"
              type="info"
              style="margin: 4px;"
            >
              {{ key }}: {{ value }}
            </el-tag>
            <span v-if="!selectedPod.annotations || Object.keys(selectedPod.annotations).length === 0">无注解</span>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 日志对话框 -->
    <el-dialog 
      v-model="logDialogVisible" 
      title="Pod日志" 
      width="80%"
      :before-close="handleLogDialogClose"
    >
      <div class="log-container">
        <div class="log-header">
          <el-form inline>
            <el-form-item label="容器:">
              <el-select v-model="selectedContainer" placeholder="选择容器" style="width: 200px;">
                <el-option 
                  v-for="container in currentPodContainers" 
                  :key="container.name" 
                  :label="container.name" 
                  :value="container.name" 
                />
              </el-select>
            </el-form-item>
            <el-form-item label="开始时间:">
              <el-date-picker
                v-model="sinceTime"
                type="datetime"
                placeholder="选择开始时间"
                style="width: 200px"
                format="YYYY-MM-DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss"
                clearable
                @change="handleTimeChange"
              />
            </el-form-item>
            <el-form-item label="行数:">
              <el-input-number 
                v-model="tailLines" 
                :min="100" 
                :max="10000" 
                :step="100"
                style="width: 140px"
              />
            </el-form-item>
            <el-form-item label="字体颜色:">
              <el-color-picker v-model="logTextColor" />
            </el-form-item>
            <el-form-item label="背景颜色:">
              <el-color-picker v-model="logBackgroundColor" />
            </el-form-item>
            <el-form-item>
              <el-button @click="handleRefreshLogs">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
              <el-button @click="handleResetLogs">
                <el-icon><RefreshLeft /></el-icon>
                重置
              </el-button>
              <el-button @click="handleDownloadLogs">
                <el-icon><Download /></el-icon>
                下载
              </el-button>
            </el-form-item>
          </el-form>
        </div>
        <div class="log-content" :style="{ backgroundColor: logBackgroundColor }">
          <pre v-loading="logLoading" :style="{ color: logTextColor, backgroundColor: logBackgroundColor }">{{ logContent }}</pre>
        </div>
      </div>
    </el-dialog>

    <!-- 删除确认对话框 -->
    <DeleteConfirmDialog
      v-model="deleteDialogVisible"
      :item-name="currentDeletePod?.name || ''"
      message="确定要删除Pod吗？"
      :loading="deleteLoading"
      @confirm="confirmDeletePod"
      @cancel="cancelDeletePod"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  ArrowLeft, 
  Refresh, 
  RefreshLeft,
  Search, 
  Box, 
  ArrowDown,
  Download
} from '@element-plus/icons-vue'
import { clusterApi } from '@/api/cluster'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'

const route = useRoute()
const router = useRouter()

const clusterName = route.params.id as string
const nodeName = ref(route.query.node as string || '')
const loading = ref(false)
const clusterDisplayName = ref('')

// 搜索表单
const searchForm = reactive({
  name: '',
  namespace: '',
  status: '',
  node: nodeName.value
})

// Pod列表
const podList = ref([])
const selectedPods = ref([])
const namespaces = ref([])
const nodes = ref([])

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 详情对话框
const detailDialogVisible = ref(false)
const selectedPod = ref(null)

// 日志对话框
const logDialogVisible = ref(false)
const logLoading = ref(false)
const logContent = ref('')
const selectedContainer = ref('')
const currentPodContainers = ref([])
const sinceTime = ref('')
const tailLines = ref(1000)

// 日志显示配置
const logTextColor = ref('#ffffff')
const logBackgroundColor = ref('#000000')

// 当前Pod的容器列表
const currentContainers = computed(() => {
  return selectedPod.value?.containers || []
})

// 获取Pod列表
const fetchPodList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      ...searchForm
    }
    const res = await clusterApi.getPods(clusterName, params)
    if (res.code === 200) {
      podList.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    ElMessage.error('获取Pod列表失败')
  } finally {
    loading.value = false
  }
}

// 获取集群信息
const fetchClusterInfo = async () => {
  try {
    const res = await clusterApi.getDetail(clusterName)
    if (res.code === 200) {
      clusterDisplayName.value = res.data.name
    }
  } catch (error) {
    console.error('获取集群信息失败:', error)
  }
}

// 获取命名空间列表
const fetchNamespaces = async () => {
  try {
    const response = await clusterApi.getNamespaces(clusterName)
    namespaces.value = response.data.map((ns: any) => ({
      name: ns.name,
      displayName: ns.name
    }))
  } catch (error) {
    console.error('获取命名空间失败:', error)
  }
}

// 获取节点列表
const fetchNodes = async () => {
  try {
    const response = await clusterApi.getNodes(clusterName)
    nodes.value = response.data.list?.map((node: any) => ({
      name: node.name,
      displayName: node.name
    })) || []
  } catch (error) {
    console.error('获取节点失败:', error)
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchPodList()
}

// 重置
const handleReset = () => {
  Object.assign(searchForm, {
    name: '',
    namespace: '',
    status: '',
    node: ''
  })
  pagination.page = 1
  fetchPodList()
}

// 分页大小改变
const handleSizeChange = (size: number) => {
  pagination.size = size
  pagination.page = 1
  fetchPodList()
}

// 当前页改变
const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchPodList()
}

// 选择改变
const handleSelectionChange = (selection: any[]) => {
  selectedPods.value = selection
}

// 返回
const handleBack = () => {
  router.back()
}

// 刷新
const handleRefresh = () => {
  fetchPodList()
}

// 查看详情
const handleViewDetail = (pod: any) => {
  selectedPod.value = pod
  detailDialogVisible.value = true
}

// 查看日志
const handleViewLogs = async (pod: any) => {
  selectedPod.value = pod
  currentPodContainers.value = pod.containers || []
  if (currentPodContainers.value.length > 0) {
    selectedContainer.value = currentPodContainers.value[0].name
  }
  logDialogVisible.value = true
  await fetchPodLogs()
  // 获取日志后自动滚动到底部
  nextTick(() => {
    const logContainer = document.querySelector('.log-content')
    if (logContainer) {
      logContainer.scrollTop = logContainer.scrollHeight
    }
  })
}

// 查看节点
const handleViewNode = (nodeName: string) => {
  router.push(`/cluster/nodes/${clusterId}?highlight=${nodeName}`)
}

// 更多操作
const handleMoreActions = async (command: string, pod: any) => {
  switch (command) {
    case 'yaml':
      await handleViewYaml(pod)
      break
    case 'events':
      await handleViewEvents(pod)
      break
    case 'exec':
      handleExecPod(pod)
      break
    case 'delete':
      await handleDeletePod(pod)
      break
  }
}

// 查看YAML
const handleViewYaml = async (pod: any) => {
  ElMessage.info('查看YAML功能开发中')
}

// 查看事件
const handleViewEvents = async (pod: any) => {
  ElMessage.info('查看事件功能开发中')
}

// 进入容器
const handleExecPod = (pod: any) => {
  ElMessage.info('进入容器功能开发中')
}

// 删除Pod相关状态
const deleteDialogVisible = ref(false)
const currentDeletePod = ref<any>(null)
const deleteLoading = ref(false)

// 删除Pod
const handleDeletePod = async (pod: any) => {
  currentDeletePod.value = pod
  deleteDialogVisible.value = true
}

// 确认删除Pod
const confirmDeletePod = async () => {
  if (!currentDeletePod.value) return
  
  deleteLoading.value = true
  try {
    const res = await clusterApi.deletePod(clusterName, currentDeletePod.value.namespace, currentDeletePod.value.name)
    if (res.code === 200) {
      ElMessage.success('删除Pod成功')
      deleteDialogVisible.value = false
      fetchPodList()
    }
  } catch (error: any) {
    ElMessage.error('删除Pod失败')
  } finally {
    deleteLoading.value = false
  }
}

// 取消删除
const cancelDeletePod = () => {
  deleteDialogVisible.value = false
  currentDeletePod.value = null
}

// 获取Pod日志
const fetchPodLogs = async () => {
  if (!selectedPod.value || !selectedContainer.value) return
  
  logLoading.value = true
  try {
    const params: any = {
      lines: tailLines.value,
      follow: false
    }
    
    if (sinceTime.value) {
      params.sinceTime = sinceTime.value
    }
    
    const response = await clusterApi.getPodLogs(
      clusterName, 
      selectedPod.value.namespace, 
      selectedPod.value.name, 
      selectedContainer.value,
      params
    )
    logContent.value = response.data || '暂无日志'
    // 获取日志后自动滚动到底部
    nextTick(() => {
      const logContainer = document.querySelector('.log-content')
      if (logContainer) {
        logContainer.scrollTop = logContainer.scrollHeight
      }
    })
  } catch (error) {
    ElMessage.error('获取日志失败')
    logContent.value = '获取日志失败'
  } finally {
    logLoading.value = false
  }
}

// 刷新日志
const handleRefreshLogs = () => {
  fetchPodLogs()
}

// 下载日志
const handleDownloadLogs = () => {
  if (!logContent.value) {
    ElMessage.warning('暂无日志可下载')
    return
  }
  
  const blob = new Blob([logContent.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${selectedPod.value.name}-${selectedContainer.value}.log`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

// 时间选择变化处理
const handleTimeChange = () => {
  fetchPodLogs()
}

// 重置日志配置
const handleResetLogs = () => {
  sinceTime.value = ''
  tailLines.value = 1000
  logTextColor.value = '#ffffff'
  logBackgroundColor.value = '#000000'
  fetchPodLogs()
}

// 容器日志
const handleContainerLogs = (container: any) => {
  selectedContainer.value = container.name
  fetchPodLogs()
}

// 容器终端
const handleContainerExec = (container: any) => {
  ElMessage.info('容器终端功能开发中')
}

// 关闭详情对话框
const handleDetailDialogClose = () => {
  detailDialogVisible.value = false
  selectedPod.value = null
}

// 关闭日志对话框
const handleLogDialogClose = () => {
  logDialogVisible.value = false
  logContent.value = ''
  selectedContainer.value = ''
  currentPodContainers.value = []
}

// 获取状态类型
const getStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    Running: 'success',
    Pending: 'warning',
    Succeeded: 'success',
    Failed: 'danger',
    Unknown: 'info'
  }
  return statusMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    Running: '运行中',
    Pending: '等待中',
    Succeeded: '成功',
    Failed: '失败',
    Unknown: '未知'
  }
  return statusMap[status] || status
}

// 获取容器状态类型
const getContainerStateType = (state: string) => {
  const stateMap: Record<string, string> = {
    running: 'success',
    waiting: 'warning',
    terminated: 'info'
  }
  return stateMap[state] || 'info'
}

// 初始化
onMounted(() => {
  fetchClusterInfo()
  fetchNamespaces()
  fetchNodes()
  fetchPodList()
})
</script>

<style scoped>
.cluster-pods {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-left h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.search-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pod-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pod-icon {
  color: #409eff;
}

.ready {
  color: #67c23a;
  font-weight: 500;
}

.not-ready {
  color: #f56c6c;
  font-weight: 500;
}

.high-restarts {
  color: #e6a23c;
  font-weight: 500;
}

.resource-usage {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
}

.usage-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.usage-label {
  width: 30px;
  color: #606266;
}

.usage-value {
  color: #2c3e50;
  font-weight: 500;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.pod-detail {
  max-height: 600px;
  overflow-y: auto;
}

.labels-container,
.annotations-container {
  min-height: 40px;
  padding: 8px;
  background: #f5f7fa;
  border-radius: 4px;
}

.log-container {
  height: 500px;
  display: flex;
  flex-direction: column;
}

.log-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.log-content {
  flex: 1;
  border-radius: 4px;
  overflow: auto;
}

.log-content pre {
  margin: 0;
  padding: 16px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-wrap: break-word;
}

@media (max-width: 768px) {
  .resource-usage {
    display: none;
  }
}
</style>