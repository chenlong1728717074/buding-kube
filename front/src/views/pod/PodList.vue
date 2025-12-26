<template>
  <div class="pod-list">
    <div class="page-header">
      <h1>Pod管理</h1>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="命名空间">
          <el-select 
            v-model="searchForm.namespace" 
            placeholder="请选择命名空间"
            style="width: 200px"
            clearable
            @change="handleNamespaceChange"
          >
            <el-option
              v-for="namespace in namespaceList"
              :key="namespace.name"
              :label="namespace.name"
              :value="namespace.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select 
            v-model="searchForm.status" 
            placeholder="请选择状态"
            style="width: 150px"
            clearable
            @change="handleStatusChange"
          >
            <el-option label="运行中" value="Running" />
            <el-option label="等待中" value="Pending" />
            <el-option label="成功" value="Succeeded" />
            <el-option label="失败" value="Failed" />
            <el-option label="未知" value="Unknown" />
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
      <el-table 
        v-loading="loading" 
        :data="podList" 
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" align="center" />
        <el-table-column prop="name" label="Pod名称" min-width="160" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="handleViewDetail(row)">
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
        <el-table-column prop="status" label="状态" width="90" align="center" header-align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="nodeName" label="节点" min-width="120" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.nodeName">{{ row.nodeName }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="podIP" label="Pod IP" width="120" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.podIP">{{ row.podIP }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="hostIP" label="主机IP" width="120" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.hostIP">{{ row.hostIP }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="creationTimestamp" label="创建时间" width="150" header-align="center" align="center">
          <template #default="{ row }">
            {{ formatDate(row.creationTimestamp) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right" header-align="center">
          <template #default="{ row }">
            <div style="display: flex; gap: 6px; align-items: center; justify-content: center; flex-wrap: nowrap;">
              <el-button size="small" @click="handleViewDetail(row)">
                详情
              </el-button>
              <el-dropdown @command="(command) => handleMoreAction(command, row)">
                <el-button size="small">
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="yaml">查看YAML</el-dropdown-item>
                    <el-dropdown-item command="logs" :disabled="row.status !== 'Running'">查看日志</el-dropdown-item>
                    <el-dropdown-item command="exec" :disabled="row.status !== 'Running'">进入容器</el-dropdown-item>
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
    <el-dialog v-model="yamlDialogVisible" title="查看YAML" width="90%" :close-on-click-modal="false" class="config-dialog yaml-dialog">
      <template #header>
        <div class="dialog-header">
          <div>
            <h3 class="dialog-title">查看YAML</h3>
            <div style="margin-top:4px;color:#6b7280;font-size:12px;">Pod 配置</div>
          </div>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <div class="yaml-dialog-content">
            <div class="yaml-info">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="集群">{{ currentClusterName }}</el-descriptions-item>
                <el-descriptions-item label="命名空间">{{ selectedPod?.namespace }}</el-descriptions-item>
                <el-descriptions-item label="名称">{{ selectedPod?.name }}</el-descriptions-item>
                <el-descriptions-item label="类型">Pod</el-descriptions-item>
              </el-descriptions>
            </div>

            <div class="yaml-view">
              <YamlEditor :model-value="yamlContent" :loading="yamlLoading" :readonly="true" height="100%" />
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="yamlDialogVisible = false">关闭</el-button>
          <span style="color: #909399; font-size: 12px; margin-left: 10px;">注意：Pod不支持修改操作</span>
        </div>
      </template>
    </el-dialog>

    <PodLogViewer
      v-model="logDialogVisible"
      :title="`查看日志 - ${selectedPod?.name || ''}`"
      :cluster-id="clusterId"
      :namespace="selectedPod?.namespace || ''"
      :pod-name="selectedPod?.name || ''"
      :container-selectable="false"
    />

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
import { ref, reactive, onMounted, onActivated, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Search, 
  Refresh, 
  ArrowDown,
  Document
} from '@element-plus/icons-vue'
import { 
  podApi, 
  type PodVO, 
  type PodQueryDTO, 
  type PodDTO
} from '@/api/pod'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import PodLogViewer from '@/components/PodLogViewer.vue'
import { namespaceApi, type NamespaceVO } from '@/api/namespace'
import { useClusterStore } from '@/stores/cluster'
import '@/assets/styles/config-editor.css'

const route = useRoute()
const router = useRouter()
const clusterStore = useClusterStore()
const loading = ref(false)

// 使用顶部选中的集群ID
const clusterId = computed(() => clusterStore.currentClusterId)
const currentClusterName = computed(() => clusterStore.currentClusterName)

// 搜索表单
const searchForm = reactive({
  namespace: '',
  status: ''
})

// Pod列表
const podList = ref<PodVO[]>([])
const selectedPods = ref<PodVO[]>([])

// 命名空间列表
const namespaceList = ref<NamespaceVO[]>([])

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 获取命名空间列表
const fetchNamespaceList = async () => {
  if (!clusterId.value) {
    namespaceList.value = []
    return
  }
  
  try {
    const params = {
      clusterId: clusterId.value,
      keyword: '',
      page: 1,
      pageSize: 10000
    }
    
    const response = await namespaceApi.getList(params)
    console.log('命名空间列表API响应:', response)
    
    if (response.code === 200 && response.data) {
      namespaceList.value = response.data.items || []
    }
  } catch (error: any) {
    console.error('获取命名空间列表失败:', error)
    ElMessage.error('获取命名空间列表失败')
  }
}

// 获取Pod列表
const fetchPodList = async () => {
  if (!clusterId.value) {
    ElMessage.warning('请先选择集群')
    return
  }
  
  try {
    loading.value = true
    const params: PodQueryDTO = {
      clusterId: clusterId.value,
      ...(searchForm.namespace && { namespace: searchForm.namespace }),
      status: searchForm.status,
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    
    const response = await podApi.getList(params)
    console.log('Pod列表API响应:', response)
    
    if (response.code === 200 && response.data) {
      podList.value = response.data.items || []
      pagination.total = response.data.total || 0
      console.log('解析后的Pod列表:', podList.value)
    }
  } catch (error: any) {
    console.error('获取Pod列表失败:', error)
    ElMessage.error('获取Pod列表失败')
  } finally {
    loading.value = false
  }
}

// 命名空间选择变化
const handleNamespaceChange = () => {
  pagination.page = 1
  fetchPodList()
}

// 状态选择变化
const handleStatusChange = () => {
  pagination.page = 1
  fetchPodList()
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchPodList()
}

// 重置
const handleReset = async () => {
  searchForm.namespace = ''
  searchForm.status = ''
  pagination.page = 1
  podList.value = []
  pagination.total = 0
  
  // 重新获取命名空间列表
  await fetchNamespaceList()
}

// 分页变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchPodList()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchPodList()
}

// 选择变化
const handleSelectionChange = (selection: PodVO[]) => {
  selectedPods.value = selection
}

// 查看详情
const handleViewDetail = (row: PodVO) => {
  router.push({
    path: `/cluster/${clusterId.value}/pod/detail`,
    query: {
      namespace: row.namespace,
      name: row.name
    }
  })
}

// 跳转到命名空间详情
const handleNamespaceDetail = (row: PodVO) => {
  router.push({
    path: `/cluster/${clusterId.value}/namespace/detail`,
    query: {
      namespace: row.namespace
    }
  })
}

// YAML对话框相关
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

// 日志对话框相关
const logDialogVisible = ref(false)
const selectedPod = ref<PodVO | null>(null)

// 更多操作
const handleMoreAction = (command: string, row: PodVO) => {
  switch (command) {
    case 'yaml':
      handleViewYaml(row)
      break
    case 'logs':
      handleViewLogs(row)
      break
    case 'exec':
      handleExec(row)
      break
    case 'delete':
      handleDelete(row)
      break
  }
}

// 查看YAML
const handleViewYaml = async (row: PodVO) => {
  try {
    yamlLoading.value = true
    yamlDialogVisible.value = true
    yamlContent.value = ''
    selectedPod.value = row
    
    const params: PodDTO = {
      clusterId: clusterId.value,
      namespace: row.namespace,
      name: row.name
    }
    
    const response = await podApi.getInfo(params)
    if (response.code === 200 && response.data) {
      yamlContent.value = response.data.yaml || '暂无YAML数据'
    } else {
      yamlContent.value = '获取YAML失败'
    }
  } catch (error: any) {
    console.error('获取YAML失败:', error)
    yamlContent.value = '获取YAML失败'
    ElMessage.error('获取YAML失败')
  } finally {
    yamlLoading.value = false
  }
}

// 查看日志
const handleViewLogs = async (row: PodVO) => {
  if (row.status !== 'Running') {
    ElMessage.warning('Pod未运行，无法查看日志')
    return
  }
  selectedPod.value = row
  logDialogVisible.value = true
}

// 进入容器
const handleExec = (row: PodVO) => {
  if (row.status !== 'Running') {
    ElMessage.warning('Pod未运行，无法进入容器')
    return
  }
  ElMessage.info('进入容器功能开发中')
}

// 删除Pod相关状态
const deleteDialogVisible = ref(false)
const currentDeletePod = ref<PodVO | null>(null)
const deleteLoading = ref(false)

// 删除Pod
const handleDelete = (row: PodVO) => {
  currentDeletePod.value = row
  deleteDialogVisible.value = true
}

// 确认删除Pod
const confirmDeletePod = async () => {
  if (!currentDeletePod.value) return
  
  deleteLoading.value = true
  try {
    const params: PodDTO = {
      clusterId: clusterId.value,
      namespace: currentDeletePod.value.namespace,
      name: currentDeletePod.value.name
    }
    await podApi.delete(params)
    ElMessage.success('删除成功')
    deleteDialogVisible.value = false
    fetchPodList()
  } catch (error: any) {
    ElMessage.error('删除失败')
  } finally {
    deleteLoading.value = false
  }
}

// 取消删除
const cancelDeletePod = () => {
  deleteDialogVisible.value = false
  currentDeletePod.value = null
}

// 获取状态类型
const getStatusType = (status: string) => {
  switch (status) {
    case 'Running':
      return 'success'
    case 'Pending':
      return 'warning'
    case 'Succeeded':
      return 'success'
    case 'Failed':
      return 'danger'
    default:
      return 'info'
  }
}

// 获取状态文本
const getStatusText = (status: string) => {
  switch (status) {
    case 'Running':
      return '运行中'
    case 'Pending':
      return '等待中'
    case 'Succeeded':
      return '成功'
    case 'Failed':
      return '失败'
    case 'Unknown':
      return '未知'
    default:
      return status
  }
}

// 格式化日期
const formatDate = (dateString?: string) => {
  if (!dateString) return '-'
  try {
    const date = new Date(dateString)
    if (isNaN(date.getTime())) return '-'
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch (error) {
    return '-'
  }
}

// 页面加载时获取数据
onMounted(() => {
  // 检查是否从其他页面跳转过来带有参数
  const namespace = route.query.namespace as string
  
  if (namespace) {
    searchForm.namespace = namespace
  }
  
  if (clusterId.value) {
    fetchNamespaceList().then(() => {
      fetchPodList()
    })
  }
})

// keep-alive恢复时根据刷新参数与现有筛选条件刷新列表
onActivated(() => {
  const refresh = route.query.refresh as string
  if (refresh === '1' && clusterId.value) {
    fetchNamespaceList().then(() => fetchPodList())
  }
})

</script>

<style scoped>
.pod-list {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
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

.search-card {
  margin-bottom: 20px;
}

.table-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
  padding: 0 20px 20px;
}

.dialog-footer {
  text-align: right;
}

.yaml-container {
  max-height: 60vh;
  overflow-y: auto;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  background-color: #f5f7fa;
}

.yaml-content {
  margin: 0;
  padding: 16px;
  font-family: 'Courier New', Consolas, monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #2c3e50;
  background-color: #f5f7fa;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* 日志弹窗样式 */
.log-dialog .el-dialog__body {
  padding: 10px 20px;
}

.log-header {
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #ebeef5;
}

.log-controls .el-form {
  margin: 0;
}

.log-controls .el-form-item {
  margin-bottom: 0;
  margin-right: 15px;
}

.log-container {
  position: relative;
  height: 600px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}

.log-content {
  height: 100%;
  overflow-y: auto;
  padding: 12px;
  margin: 0;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* 实时日志指示器 */
.real-time-indicator {
  position: absolute;
  top: 10px;
  right: 10px;
  background: rgba(103, 194, 58, 0.1);
  color: #67c23a;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
  z-index: 10;
}

.blinking {
  animation: blink 1.5s ease-in-out infinite;
}

@keyframes blink {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.3;
  }
}

.log-content::-webkit-scrollbar {
  width: 8px;
}

.log-content::-webkit-scrollbar-track {
  background: #2d2d2d;
}

.log-content::-webkit-scrollbar-thumb {
  background: #555;
  border-radius: 4px;
}

.log-content::-webkit-scrollbar-thumb:hover {
  background: #777;
}

.yaml-dialog-content {
  display: flex;
  flex-direction: column;
}

.yaml-info {
  margin-bottom: 16px;
}

.yaml-info .info-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.yaml-info .label {
  font-weight: 500;
  color: #606266;
}

.yaml-info .value {
  color: #303133;
  font-weight: 600;
}

.yaml-editor-wrapper {
  flex: 1;
  min-height: 0;
}

</style>
