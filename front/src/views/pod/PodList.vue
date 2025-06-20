<template>
  <div class="pod-list">
    <div class="page-header">
      <h1>Pod管理</h1>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="集群">
          <el-select 
            v-model="searchForm.clusterId" 
            placeholder="请选择集群"
            style="width: 200px"
            clearable
            @change="handleClusterChange"
          >
            <el-option
              v-for="cluster in clusterList"
              :key="cluster.id"
              :label="cluster.name"
              :value="cluster.id"
            />
          </el-select>
        </el-form-item>
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
                    <el-dropdown-item command="logs">查看日志</el-dropdown-item>
                    <el-dropdown-item command="exec">进入容器</el-dropdown-item>
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
    <el-dialog 
      v-model="yamlDialogVisible" 
      title="查看YAML" 
      width="90%"
      :before-close="() => yamlDialogVisible = false"
      destroy-on-close
    >
      <div class="yaml-dialog-content">
        <div class="yaml-info">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="集群">{{ currentClusterName }}</el-descriptions-item>
            <el-descriptions-item label="命名空间">{{ selectedPod?.namespace }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ selectedPod?.name }}</el-descriptions-item>
            <el-descriptions-item label="类型">Pod</el-descriptions-item>
          </el-descriptions>
        </div>
        
        <div class="yaml-editor-wrapper">
          <YamlEditor 
            :model-value="yamlContent"
            :loading="yamlLoading"
            :readonly="true"
            height="500px"
          />
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="yamlDialogVisible = false">关闭</el-button>
          <span style="color: #909399; font-size: 12px; margin-left: 10px;">注意：Pod不支持修改操作</span>
        </div>
      </template>
    </el-dialog>

    <!-- 日志查看对话框 -->
     <el-dialog
       v-model="logDialogVisible"
       :title="`查看日志 - ${selectedPod?.name || ''}`"
       width="90%"
       :close-on-click-modal="false"
       class="log-dialog"
       @close="handleCloseLogDialog"
     >
       <div class="log-header">
         <div class="log-controls">
           <el-form inline>
             <!-- Pod列表页面显示所有容器日志，无需选择容器 -->
             <el-form-item label="开始时间:">
               <el-date-picker
                 v-model="sinceTime"
                 type="datetime"
                 placeholder="选择开始时间"
                 style="width: 200px"
                 format="YYYY年MM月DD日 HH:mm:ss"
                 value-format="YYYY-MM-DD HH:mm:ss"
                 clearable
                 @change="handleTimeChange"
                 :teleported="false"
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
             <el-form-item label="实时日志:">
               <el-button 
                 :type="followLogs ? 'danger' : 'success'"
                 @click="handleRealTimeToggle"
                 :loading="logLoading"
               >
                 <el-icon>
                   <VideoPlay v-if="!followLogs" />
                   <VideoPause v-else />
                 </el-icon>
                 {{ followLogs ? '停止实时' : '开始实时' }}
               </el-button>
             </el-form-item>
             <el-form-item label="字体颜色:">
               <el-color-picker v-model="logTextColor" size="small" />
             </el-form-item>
             <el-form-item label="背景颜色:">
               <el-color-picker v-model="logBackgroundColor" size="small" />
             </el-form-item>
             <el-form-item>
               <el-button type="primary" @click="handleRefreshLogs" :loading="logLoading">
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
       </div>
       
       <div v-loading="logLoading" class="log-container" :style="{ backgroundColor: logBackgroundColor }">
         <div v-if="followLogs" class="real-time-indicator">
           <el-icon class="blinking"><VideoPause /></el-icon>
           实时日志连接中
         </div>
         <pre class="log-content" :style="{ color: logTextColor, backgroundColor: logBackgroundColor }">{{ logContent }}</pre>
       </div>
       
       <template #footer>
         <div class="dialog-footer">
           <el-button @click="handleCloseLogDialog">关闭</el-button>
         </div>
       </template>
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
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Search, 
  Refresh, 
  RefreshLeft,
  ArrowDown,
  Download,
  VideoPlay,
  VideoPause,
  Document
} from '@element-plus/icons-vue'
import { 
  podApi, 
  type PodVO, 
  type PodQueryDTO, 
  type PodDTO,
  type PodLogDTO
} from '@/api/pod'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import { clusterApi, type ClusterVO } from '@/api/cluster'
import { namespaceApi, type NamespaceVO } from '@/api/namespace'

const route = useRoute()
const router = useRouter()
const loading = ref(false)

// 搜索表单
const searchForm = reactive({
  clusterId: '',
  namespace: '',
  status: ''
})

// Pod列表
const podList = ref<PodVO[]>([])
const selectedPods = ref<PodVO[]>([])

// 集群列表
const clusterList = ref<ClusterVO[]>([])
const currentClusterName = ref('')

// 命名空间列表
const namespaceList = ref<NamespaceVO[]>([])



// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 获取集群列表
const fetchClusterList = async () => {
  try {
    const response = await clusterApi.getClusters({ page: 1, pageSize: 10000 })
    console.log('集群列表API响应:', response)
    
    if (response.code === 200 && response.data && response.data.items) {
      clusterList.value = response.data.items
      
      // 如果没有选中集群且有集群数据，自动选择第一个
      if (!searchForm.clusterId && clusterList.value.length > 0) {
        searchForm.clusterId = clusterList.value[0].id
        fetchNamespaceList()
      }
    }
  } catch (error: any) {
    console.error('获取集群列表失败:', error)
    ElMessage.error('获取集群列表失败')
  }
}

// 获取命名空间列表
const fetchNamespaceList = async () => {
  if (!searchForm.clusterId) {
    namespaceList.value = []
    return
  }
  
  try {
    const params = {
      clusterId: searchForm.clusterId,
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
  if (!searchForm.clusterId) {
    podList.value = []
    pagination.total = 0
    return
  }
  
  try {
    loading.value = true
    const params: PodQueryDTO = {
      clusterId: searchForm.clusterId,
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

// 集群选择变化
const handleClusterChange = () => {
  searchForm.namespace = ''
  pagination.page = 1
  fetchNamespaceList()
  podList.value = []
  pagination.total = 0
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
  searchForm.clusterId = ''
  searchForm.namespace = ''
  searchForm.status = ''
  pagination.page = 1
  podList.value = []
  namespaceList.value = []
  pagination.total = 0
  
  // 重新获取集群列表并自动选择第一个
  await fetchClusterList()
  if (clusterList.value.length > 0) {
    searchForm.clusterId = clusterList.value[0].id
    // 获取命名空间列表并自动选择第一个
    await fetchNamespaceList()
  }
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
  // 获取当前选中的集群名称
  const currentCluster = clusterList.value.find(cluster => cluster.id === searchForm.clusterId)
  
  router.push({
    path: '/pod/detail',
    query: {
      clusterId: searchForm.clusterId,
      clusterName: currentCluster?.name || '',
      namespace: row.namespace,
      name: row.name
    }
  })
}

// 跳转到命名空间详情
const handleNamespaceDetail = (row: PodVO) => {
  router.push({
    path: '/namespace/detail',
    query: {
      clusterId: searchForm.clusterId,
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
const logContent = ref('')
const logLoading = ref(false)
const selectedPod = ref<PodVO | null>(null)
// Pod列表页面不需要选择容器，显示所有容器日志
const followLogs = ref(false)
const tailLines = ref(1000)
const sinceTime = ref('')

// 日志显示配置
const logTextColor = ref('#ffffff')
const logBackgroundColor = ref('#000000')

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
    
    // 获取当前集群名称
    const cluster = clusterList.value.find(c => c.id === searchForm.clusterId)
    currentClusterName.value = cluster?.name || ''
    
    const params: PodDTO = {
      clusterId: searchForm.clusterId,
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
  try {
    selectedPod.value = row
    logDialogVisible.value = true
    logContent.value = ''
    
    // Pod列表页面直接获取所有容器日志
    await fetchPodLogs()
    // 获取日志后自动滚动到底部
    nextTick(() => {
      const logContainer = document.querySelector('.log-content')
      if (logContainer) {
        logContainer.scrollTop = logContainer.scrollHeight
      }
    })
  } catch (error: any) {
    console.error('获取Pod信息失败:', error)
    ElMessage.error('获取Pod信息失败')
  }
}

// 当前日志流控制器
let currentLogController: AbortController | null = null

// 获取Pod日志 - 流式处理
const fetchPodLogs = async () => {
  if (!selectedPod.value) {
    ElMessage.warning('请先选择Pod')
    return
  }
  
  // 只在非实时模式下显示loading
  if (!followLogs.value) {
    logLoading.value = true
  }
  
  try {
    // 停止之前的日志流
    if (currentLogController) {
      currentLogController.abort()
      currentLogController = null
    }
    
    // 清空之前的日志内容
    logContent.value = ''
    
    const params: PodLogDTO = {
      clusterId: searchForm.clusterId,
      namespace: selectedPod.value.namespace,
      name: selectedPod.value.name,
      // Pod列表页面不传递container参数，获取所有容器日志
      follow: followLogs.value,
      tailLines: tailLines.value
    }
    
    // 只有当sinceTime有值时才添加到参数中
    if (sinceTime.value) {
      params.sinceTime = sinceTime.value
    }
    
    // 创建新的控制器
    currentLogController = new AbortController()
    
    await podApi.getLogs(
      params,
      // onData 回调 - 处理流式数据
      (data: string) => {
        console.log('收到日志数据:', data) // 调试日志
        
        // 过滤掉控制信息，但保留实际日志内容
        const trimmedData = data.trim()
        if (trimmedData && 
            !trimmedData.startsWith('[START]') && 
            !trimmedData.startsWith('[END]') &&
            !trimmedData.startsWith('[ERROR]') &&
            !trimmedData.startsWith('[STOP]') &&
            trimmedData !== 'ping' &&
            trimmedData !== 'Connected successfully') {
          
          // 实时追加日志内容
          logContent.value += data
          
          // 自动滚动到底部
          nextTick(() => {
            const logContainer = document.querySelector('.log-content')
            if (logContainer) {
              logContainer.scrollTop = logContainer.scrollHeight
            }
          })
        }
      },
      // onError 回调
      (error: Error) => {
        console.error('日志流错误:', error)
        if (error.name !== 'AbortError') {
          ElMessage.error('日志流连接中断')
          followLogs.value = false // 出错时停止实时模式
        }
      },
      // 传递AbortSignal
      currentLogController?.signal
    )
    
  } catch (error: any) {
    console.error('获取日志失败:', error)
    if (!error.name || error.name !== 'AbortError') {
      logContent.value = '获取日志失败: ' + (error.message || '网络错误')
      ElMessage.error('获取日志失败')
      followLogs.value = false // 出错时停止实时模式
    }
  } finally {
    // 只在非实时模式下关闭loading
    if (!followLogs.value) {
      logLoading.value = false
    }
    // 注意：不要在这里设置currentLogController = null，因为实时模式需要保持连接
  }
}

// 刷新日志
const handleRefreshLogs = () => {
  fetchPodLogs()
}

// 实时日志切换
const handleRealTimeToggle = () => {
  if (followLogs.value) {
    // 停止实时日志
    stopRealTimeLogs()
  } else {
    // 开始实时日志
    startRealTimeLogs()
  }
}

// 开始实时日志
const startRealTimeLogs = () => {
  followLogs.value = true
  fetchPodLogs() // 重新获取日志，这次会传递follow=true
}

// 停止实时日志
const stopRealTimeLogs = () => {
  followLogs.value = false
  
  // 停止日志流
  if (currentLogController) {
    currentLogController.abort()
    currentLogController = null
  }
  
  // 停止loading状态
  logLoading.value = false
  
  ElMessage.success('已停止实时日志')
}

// 停止日志刷新
const stopLogs = () => {
  followLogs.value = false
  
  // 停止日志流
  if (currentLogController) {
    currentLogController.abort()
    currentLogController = null
  }
  
  ElMessage.success('已停止日志刷新')
}

// 下载日志
const handleDownloadLogs = () => {
  if (!logContent.value || !selectedPod.value) {
    ElMessage.warning('暂无日志可下载')
    return
  }
  
  const blob = new Blob([logContent.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${selectedPod.value.name}-all-containers.log`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

// 时间选择变化
const handleTimeChange = () => {
  if (selectedPod.value) {
    fetchPodLogs()
  }
}

// 重置日志配置
const handleResetLogs = () => {
  sinceTime.value = ''
  tailLines.value = 1000
  logTextColor.value = '#ffffff'
  logBackgroundColor.value = '#000000'
  if (followLogs.value) {
    stopRealTimeLogs()
  }
  if (selectedPod.value && selectedContainer.value) {
    fetchPodLogs()
  }
  ElMessage.success('日志配置已重置')
}

// 关闭日志弹窗时清理
const handleCloseLogDialog = () => {
  logDialogVisible.value = false
  stopRealTimeLogs()
  selectedPod.value = null
  logContent.value = ''
  sinceTime.value = ''
  tailLines.value = 1000
}

// 进入容器
const handleExec = (row: PodVO) => {
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
      clusterId: searchForm.clusterId,
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
  const clusterId = route.query.clusterId as string
  const namespace = route.query.namespace as string
  
  if (clusterId) {
    searchForm.clusterId = clusterId
  }
  if (namespace) {
    searchForm.namespace = namespace
  }
  
  fetchClusterList().then(() => {
    if (searchForm.clusterId) {
      fetchNamespaceList().then(() => {
        // 如果有命名空间参数，则加载Pod列表；否则用第一个集群查询所有命名空间的Pod
        if (searchForm.namespace) {
          fetchPodList()
        } else if (searchForm.clusterId) {
          fetchPodList()
        }
      })
    }
  })
})

// 组件卸载时清理日志流
onUnmounted(() => {
  if (currentLogController) {
    currentLogController.abort()
    currentLogController = null
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