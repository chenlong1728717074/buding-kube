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
    </div>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="podInfo" class="detail-content">
      <!-- 基本信息 -->
      <el-card class="info-card">
        <template #header>
          <div class="card-header">
            <span>基本信息</span>
            <div class="header-actions">
              <el-button size="small" @click="handleViewYaml" v-if="podInfo.yaml">
                <el-icon><Document /></el-icon>
                查看YAML
              </el-button>
              <el-button size="small" @click="handleDelete">
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </div>
        </template>
        
        <el-descriptions :column="2" border>
          <el-descriptions-item label="集群ID">
            {{ clusterId || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="集群名称">
            {{ clusterName || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="Pod名称">
            {{ podInfo.name }}
          </el-descriptions-item>
          <el-descriptions-item label="命名空间">
            {{ podInfo.namespace }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(podInfo.status)">
              {{ getStatusText(podInfo.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="节点">
            {{ podInfo.nodeName || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="Pod IP">
            {{ podInfo.podIP || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="主机IP">
            {{ podInfo.hostIP || '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatDate(podInfo.creationTimestamp) }}
          </el-descriptions-item>
          <el-descriptions-item label="重启次数">
            {{ podInfo.restartCount || 0 }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 容器信息 -->
      <el-card class="info-card" v-if="podInfo.containers && podInfo.containers.length > 0">
        <template #header>
          <span>容器信息</span>
        </template>
        
        <el-table :data="podInfo.containers" stripe>
          <el-table-column prop="name" label="容器名称" min-width="150" />
          <el-table-column prop="image" label="镜像" min-width="200" />
          <el-table-column prop="ready" label="就绪状态" width="100" align="center">
            <template #default="{ row }">
              <el-tag :type="row.ready ? 'success' : 'danger'">
                {{ row.ready ? '就绪' : '未就绪' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="restartCount" label="重启次数" width="100" align="center" />
          <el-table-column label="端口" min-width="150">
            <template #default="{ row }">
              <div v-if="row.ports && row.ports.length > 0">
                <el-tag 
                  v-for="port in row.ports" 
                  :key="port.containerPort" 
                  size="small" 
                  style="margin-right: 4px; margin-bottom: 4px;"
                >
                  {{ port.containerPort }}{{ port.protocol ? '/' + port.protocol : '' }}
                </el-tag>
              </div>
              <span v-else>-</span>
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
          <span>注解</span>
        </template>
        
        <div class="annotations-container">
          <div 
            v-for="(value, key) in podInfo.annotations" 
            :key="key"
            class="annotation-item"
          >
            <strong>{{ key }}:</strong>
            <span class="annotation-value">{{ value }}</span>
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

// 删除Pod
const handleDelete = () => {
  if (!podInfo.value) return
  
  ElMessageBox.confirm(
    `确定要删除Pod "${podInfo.value.name}" 吗？此操作不可恢复！`,
    '确认删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      const params: PodDTO = {
        clusterId: route.query.clusterId as string,
        namespace: podInfo.value!.namespace,
        name: podInfo.value!.name
      }
      await podApi.delete(params)
      ElMessage.success('删除成功')
      handleBack()
    } catch (error: any) {
      ElMessage.error('删除失败')
    }
  })
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
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.labels-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.label-tag {
  margin: 0;
}

.annotations-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.annotation-item {
  padding: 8px;
  background-color: #f8f9fa;
  border-radius: 4px;
  border-left: 3px solid #409eff;
}

.annotation-value {
  margin-left: 8px;
  word-break: break-all;
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
</style>