<template>
  <div class="pod-detail">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="handleBack" style="margin-right: 16px;">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1>Pod详情</h1>
      </div>
      <div class="header-actions" v-if="podInfo">
        <el-button size="small" @click="handleViewYaml" v-if="podInfo.yaml">
          <el-icon><Document /></el-icon>
          查看YAML
        </el-button>
        <el-button size="small" type="danger" @click="handleDelete">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="podInfo" class="detail-content">
      <!-- 基本信息 -->
      <el-card class="info-card">
        <template #header>
          <span>基本信息</span>
        </template>
        
        <div class="info-grid">
          <div class="info-item">
            <label>集群ID:</label>
            <span>{{ clusterId || '-' }}</span>
          </div>
          <div class="info-item">
            <label>集群名称:</label>
            <span>{{ clusterName || '-' }}</span>
          </div>
          <div class="info-item">
            <label>Pod名称:</label>
            <span>{{ podInfo.name }}</span>
          </div>
          <div class="info-item">
            <label>命名空间:</label>
            <span>{{ podInfo.namespace }}</span>
          </div>
          <div class="info-item">
            <label>状态:</label>
            <el-tag :type="getStatusType(podInfo.status)">
              {{ getStatusText(podInfo.status) }}
            </el-tag>
          </div>
          <div class="info-item">
            <label>节点:</label>
            <span>{{ podInfo.nodeName || '-' }}</span>
          </div>
          <div class="info-item">
            <label>Pod IP:</label>
            <span>{{ podInfo.podIP || '-' }}</span>
          </div>
          <div class="info-item">
            <label>主机IP:</label>
            <span>{{ podInfo.hostIP || '-' }}</span>
          </div>
          <div class="info-item">
            <label>创建时间:</label>
            <span>{{ formatDate(podInfo.creationTimestamp) }}</span>
          </div>
          <div class="info-item">
            <label>重启次数:</label>
            <span>{{ podInfo.restartCount || 0 }}</span>
          </div>
        </div>
      </el-card>

      <!-- 容器信息 -->
      <el-card class="info-card" v-if="podInfo.containers && podInfo.containers.length > 0">
        <template #header>
          <span>容器信息</span>
        </template>
        
        <el-table :data="podInfo.containers" stripe style="width: 100%">
          <el-table-column prop="name" label="容器名称" width="180" show-overflow-tooltip />
          <el-table-column prop="image" label="镜像" min-width="250" show-overflow-tooltip />
          <el-table-column prop="ready" label="就绪状态" width="120" align="center">
            <template #default="{ row }">
              <el-tag :type="row.ready ? 'success' : 'danger'" size="small">
                {{ row.ready ? '就绪' : '未就绪' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="restartCount" label="重启次数" width="120" align="center" />
          <el-table-column label="端口" width="200">
            <template #default="{ row }">
              <div v-if="row.ports && row.ports.length > 0" class="ports-container">
                <el-tag 
                  v-for="port in row.ports" 
                  :key="port.containerPort" 
                  size="small" 
                  style="margin-right: 4px; margin-bottom: 4px;"
                >
                  {{ port.containerPort }}{{ port.protocol ? '/' + port.protocol : '' }}
                </el-tag>
              </div>
              <span v-else class="text-muted">-</span>
            </template>
          </el-table-column>
          <el-table-column label="状态" min-width="150">
            <template #default="{ row }">
              <div class="container-status">
                <div v-if="row.state">
                  <el-tag 
                    :type="getContainerStateType(row.state)" 
                    size="small"
                  >
                    {{ getContainerStateText(row.state) }}
                  </el-tag>
                </div>
                <span v-else class="text-muted">-</span>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 环境变量 -->
      <el-card class="info-card" v-if="hasEnvironmentVariables">
        <template #header>
          <span>环境变量</span>
        </template>
        
        <div v-for="(container, index) in podInfo.containers" :key="index">
          <div v-if="container.envs && container.envs.length > 0">
            <h4>{{ container.name }}</h4>
            <el-table :data="container.envs" stripe size="small">
              <el-table-column prop="name" label="变量名" min-width="200" />
              <el-table-column prop="value" label="值" min-width="300">
                <template #default="{ row }">
                  <span v-if="row.value">{{ row.value }}</span>
                  <span v-else-if="row.valueFrom" class="text-info">引用值</span>
                  <span v-else>-</span>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </el-card>

      <!-- 标签 -->
      <el-card class="info-card" v-if="podInfo.labels && Object.keys(podInfo.labels).length > 0">
        <template #header>
          <span>标签</span>
        </template>
        
        <div class="labels-container">
          <el-tag 
            v-for="(value, key) in podInfo.labels" 
            :key="key"
            class="label-tag"
          >
            {{ key }}: {{ value }}
          </el-tag>
        </div>
      </el-card>

      <!-- 注解 -->
      <el-card class="info-card" v-if="podInfo.annotations && Object.keys(podInfo.annotations).length > 0">
        <template #header>
          <div class="card-header">
            <span>注解</span>
            <el-tag size="small" type="info">{{ Object.keys(podInfo.annotations).length }}</el-tag>
          </div>
        </template>
        
        <div class="annotations-grid">
          <div 
            v-for="(value, key) in podInfo.annotations" 
            :key="key"
            class="annotation-item"
          >
            <div class="annotation-key">{{ key }}</div>
            <div class="annotation-value">{{ value }}</div>
          </div>
        </div>
      </el-card>

      <!-- 事件 -->
      <el-card class="info-card" v-if="podInfo.events && podInfo.events.length > 0">
        <template #header>
          <span>事件</span>
        </template>
        
        <el-table :data="podInfo.events" stripe>
          <el-table-column prop="type" label="类型" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="row.type === 'Warning' ? 'warning' : 'info'" size="small">
                {{ row.type }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="reason" label="原因" min-width="150" />
          <el-table-column prop="message" label="消息" min-width="300" />
          <el-table-column prop="firstTimestamp" label="首次时间" min-width="180">
            <template #default="{ row }">
              {{ formatDate(row.firstTimestamp) }}
            </template>
          </el-table-column>
          <el-table-column prop="lastTimestamp" label="最后时间" min-width="180">
            <template #default="{ row }">
              {{ formatDate(row.lastTimestamp) }}
            </template>
          </el-table-column>
          <el-table-column prop="count" label="次数" width="80" align="center" />
        </el-table>
      </el-card>


    </div>

    <div v-else class="empty-container">
      <el-empty description="未找到Pod信息" />
    </div>

    <!-- 删除确认对话框 -->
    <DeleteConfirmDialog
      v-model="deleteDialogVisible"
      :item-name="podInfo?.name || ''"
      message="确定要删除Pod吗？"
      :loading="deleteLoading"
      @confirm="confirmDeletePod"
      @cancel="cancelDeletePod"
    />

    <!-- YAML查看对话框 -->
    <el-dialog
      v-model="yamlDialogVisible"
      title="查看YAML"
      width="80%"
      :close-on-click-modal="false"
    >
      <div v-loading="yamlLoading" class="yaml-container">
        <pre class="yaml-content">{{ yamlContent }}</pre>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="yamlDialogVisible = false">关闭</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  ArrowLeft,
  Delete,
  Document
} from '@element-plus/icons-vue'
import { 
  podApi, 
  type PodInfoVO, 
  type PodDTO
} from '@/api/pod'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const podInfo = ref<PodInfoVO | null>(null)

// 集群信息
const clusterId = ref('')
const clusterName = ref('')

// YAML对话框相关
const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlLoading = ref(false)

// 计算属性：是否有环境变量
const hasEnvironmentVariables = computed(() => {
  if (!podInfo.value?.containers) return false
  return podInfo.value.containers.some(container => 
    container.envs && container.envs.length > 0
  )
})

// 获取Pod详情
const fetchPodDetail = async () => {
  const clusterIdParam = route.query.clusterId as string
  const clusterNameParam = route.query.clusterName as string
  const namespace = route.query.namespace as string
  const name = route.query.name as string
  
  if (!clusterIdParam || !namespace || !name) {
    ElMessage.error('缺少必要参数')
    handleBack()
    return
  }
  
  // 设置集群信息
  clusterId.value = clusterIdParam
  clusterName.value = clusterNameParam || ''
  
  try {
    loading.value = true
    const params: PodDTO = {
      clusterId: clusterIdParam,
      namespace,
      name
    }
    
    const response = await podApi.getInfo(params)
    console.log('Pod详情API响应:', response)
    
    if (response.code === 200 && response.data) {
      podInfo.value = response.data
      console.log('解析后的Pod详情:', podInfo.value)
    } else {
      ElMessage.error('获取Pod详情失败')
    }
  } catch (error: any) {
    console.error('获取Pod详情失败:', error)
    ElMessage.error('获取Pod详情失败')
  } finally {
    loading.value = false
  }
}

// 查看YAML
const handleViewYaml = () => {
  if (podInfo.value?.yaml) {
    yamlContent.value = podInfo.value.yaml
    yamlDialogVisible.value = true
  }
}

// 返回
const handleBack = () => {
  router.back()
}

// 删除相关状态
const deleteDialogVisible = ref(false)
const deleteLoading = ref(false)

// 删除Pod
const handleDelete = () => {
  if (!podInfo.value) return
  deleteDialogVisible.value = true
}

// 确认删除Pod
const confirmDeletePod = async () => {
  if (!podInfo.value) return
  
  deleteLoading.value = true
  try {
    const params: PodDTO = {
      clusterId: route.query.clusterId as string,
      namespace: podInfo.value.namespace,
      name: podInfo.value.name
    }
    await podApi.delete(params)
    ElMessage.success('删除成功')
    deleteDialogVisible.value = false
    router.back()
  } catch (error) {
    ElMessage.error('删除失败')
  } finally {
    deleteLoading.value = false
  }
}

// 取消删除
const cancelDeletePod = () => {
  deleteDialogVisible.value = false
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

// 获取容器状态类型
const getContainerStateType = (state: any) => {
  if (state.running) return 'success'
  if (state.waiting) return 'warning'
  if (state.terminated) {
    return state.terminated.exitCode === 0 ? 'success' : 'danger'
  }
  return 'info'
}

// 获取容器状态文本
const getContainerStateText = (state: any) => {
  if (state.running) return '运行中'
  if (state.waiting) return `等待中: ${state.waiting.reason || ''}`
  if (state.terminated) {
    const reason = state.terminated.reason || '已终止'
    const exitCode = state.terminated.exitCode
    return `${reason} (${exitCode})`
  }
  return '未知'
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
      minute: '2-digit',
      second: '2-digit'
    })
  } catch (error) {
    return '-'
  }
}

// 页面加载时获取数据
onMounted(() => {
  fetchPodDetail()
})
</script>

<style scoped>
.pod-detail {
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

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.loading-container {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
}

.info-item {
  display: flex;
  align-items: center;
}

.info-item label {
  font-weight: 500;
  color: #606266;
  min-width: 120px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.labels-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.label-tag {
  margin: 0;
}

.annotations-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
}

.annotation-item {
  padding: 16px;
  background-color: #f8f9fa;
  border-radius: 8px;
  border-left: 4px solid #409eff;
  transition: all 0.2s ease;
}

.annotation-item:hover {
  background-color: #f0f2f5;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.annotation-key {
  font-weight: 600;
  color: #409eff;
  font-size: 13px;
  margin-bottom: 8px;
  word-break: break-all;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.annotation-value {
  color: #606266;
  font-size: 13px;
  word-break: break-all;
  line-height: 1.5;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  background-color: #ffffff;
  padding: 8px;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
}

.ports-container {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.container-status {
  display: flex;
  align-items: center;
}

.text-muted {
  color: #909399;
  font-style: italic;
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

.dialog-footer {
  text-align: right;
}

.empty-container {
  background: white;
  border-radius: 8px;
  padding: 40px;
  text-align: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.text-info {
  color: #909399;
  font-style: italic;
}

@media (max-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>