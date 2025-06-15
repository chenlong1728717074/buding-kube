<template>
  <div class="cluster-detail">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="handleBack">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1>{{ clusterInfo?.name || '集群详情' }}</h1>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <div v-loading="loading" class="detail-content">
      <!-- 基本信息 -->
      <el-card class="info-card">
        <template #header>
          <div class="card-header">
            <span>基本信息</span>
            <el-button size="small" @click="handleEdit">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
          </div>
        </template>
        <div class="info-grid">
          <div class="info-item">
            <label>集群名称：</label>
            <span>{{ clusterInfo?.name }}</span>
          </div>
          <div class="info-item">
            <label>API地址：</label>
            <span>{{ clusterInfo?.endpoint }}</span>
          </div>
          <div class="info-item">
            <label>版本：</label>
            <span>{{ clusterInfo?.version }}</span>
          </div>
          <div class="info-item">
            <label>状态：</label>
            <el-tag :type="getStatusType(clusterInfo?.status)">
              {{ getStatusText(clusterInfo?.status) }}
            </el-tag>
          </div>
          <div class="info-item">
            <label>创建时间：</label>
            <span>{{ clusterInfo?.createTime }}</span>
          </div>
          <div class="info-item">
            <label>描述：</label>
            <span>{{ clusterInfo?.description || '暂无描述' }}</span>
          </div>
        </div>
      </el-card>

      <!-- 统计信息 -->
      <div class="stats-grid">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon nodes">
              <el-icon size="24"><Monitor /></el-icon>
            </div>
            <div class="stat-info">
              <h3>{{ clusterStats.nodeCount || 0 }}</h3>
              <p>节点总数</p>
            </div>
          </div>
          <div class="stat-footer">
            <el-button text @click="handleViewNodes">
              查看详情 <el-icon><ArrowRight /></el-icon>
            </el-button>
          </div>
        </el-card>

        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon pods">
              <el-icon size="24"><Box /></el-icon>
            </div>
            <div class="stat-info">
              <h3>{{ clusterStats.podCount || 0 }}</h3>
              <p>Pod总数</p>
            </div>
          </div>
          <div class="stat-footer">
            <el-button text @click="handleViewPods">
              查看详情 <el-icon><ArrowRight /></el-icon>
            </el-button>
          </div>
        </el-card>

        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon namespaces">
              <el-icon size="24"><Folder /></el-icon>
            </div>
            <div class="stat-info">
              <h3>{{ clusterStats.namespaceCount || 0 }}</h3>
              <p>命名空间</p>
            </div>
          </div>
        </el-card>

        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon services">
              <el-icon size="24"><Connection /></el-icon>
            </div>
            <div class="stat-info">
              <h3>{{ clusterStats.serviceCount || 0 }}</h3>
              <p>服务总数</p>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 资源使用情况 -->
      <el-card class="resource-card">
        <template #header>
          <span>资源使用情况</span>
        </template>
        <div class="resource-grid">
          <div class="resource-item">
            <div class="resource-header">
              <span>CPU使用率</span>
              <span class="resource-value">{{ clusterStats.cpuUsage || 0 }}%</span>
            </div>
            <el-progress 
              :percentage="clusterStats.cpuUsage || 0" 
              :color="getProgressColor(clusterStats.cpuUsage || 0)"
            />
          </div>
          <div class="resource-item">
            <div class="resource-header">
              <span>内存使用率</span>
              <span class="resource-value">{{ clusterStats.memoryUsage || 0 }}%</span>
            </div>
            <el-progress 
              :percentage="clusterStats.memoryUsage || 0" 
              :color="getProgressColor(clusterStats.memoryUsage || 0)"
            />
          </div>
          <div class="resource-item">
            <div class="resource-header">
              <span>存储使用率</span>
              <span class="resource-value">{{ clusterStats.storageUsage || 0 }}%</span>
            </div>
            <el-progress 
              :percentage="clusterStats.storageUsage || 0" 
              :color="getProgressColor(clusterStats.storageUsage || 0)"
            />
          </div>
        </div>
      </el-card>

      <!-- 最近事件 -->
      <el-card class="events-card">
        <template #header>
          <span>最近事件</span>
        </template>
        <el-table :data="recentEvents" stripe>
          <el-table-column prop="type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag 
                :type="getEventType(row.type)" 
                size="small"
              >
                {{ row.type }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="reason" label="原因" width="150" />
          <el-table-column prop="message" label="消息" show-overflow-tooltip />
          <el-table-column prop="time" label="时间" width="180" />
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  ArrowLeft, 
  Refresh, 
  Edit, 
  Monitor, 
  Box, 
  Folder, 
  Connection,
  ArrowRight
} from '@element-plus/icons-vue'
import { clusterApi, type ClusterVO } from '@/api/cluster'

const route = useRoute()
const router = useRouter()

const loading = ref(false)

// 集群信息
const clusterInfo = ref<ClusterVO | null>(null)

// 集群统计
const clusterStats = reactive({
  nodeCount: 0,
  podCount: 0,
  namespaceCount: 0,
  serviceCount: 0,
  cpuUsage: 0,
  memoryUsage: 0,
  storageUsage: 0
})

// 最近事件
const recentEvents = ref([])

// 获取集群详情
const fetchClusterDetail = async () => {
  loading.value = true
  try {
    const clusterName = route.params.id as string
    const response = await clusterApi.getCluster(clusterName)
    clusterInfo.value = response.data
    
    ElMessage.success('集群详情加载成功')
  } catch (error) {
    console.error('加载集群详情失败:', error)
    ElMessage.error('获取集群详情失败')
  } finally {
    loading.value = false
  }
}

// 获取集群统计
const fetchClusterStats = async () => {
  try {
    const clusterName = route.params.id as string
    const response = await clusterApi.getStats(clusterName)
    Object.assign(clusterStats, response.data)
  } catch (error) {
    console.error('获取集群统计失败:', error)
  }
}

// 获取最近事件
const fetchRecentEvents = async () => {
  try {
    const clusterName = route.params.id as string
    const response = await clusterApi.getEvents(clusterName, {
      limit: 10,
      sort: 'createTime',
      order: 'desc'
    })
    recentEvents.value = response.data.list || []
  } catch (error) {
    console.error('获取最近事件失败:', error)
  }
}

// 返回
const handleBack = () => {
  router.back()
}

// 刷新
const handleRefresh = () => {
  fetchClusterDetail()
  fetchClusterStats()
  fetchRecentEvents()
}

// 编辑
const handleEdit = () => {
  // TODO: 打开编辑对话框
  ElMessage.info('编辑功能开发中')
}

// 查看节点
const handleViewNodes = () => {
  if (clusterInfo.value) {
    router.push(`/cluster/nodes/${clusterInfo.value.name}`)
  }
}

// 查看Pod
const handleViewPods = () => {
  if (clusterInfo.value) {
    router.push(`/cluster/pods/${clusterInfo.value.name}`)
  }
}

// 获取状态类型
const getStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    running: 'success',
    error: 'danger',
    offline: 'info'
  }
  return statusMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    running: '运行中',
    error: '异常',
    offline: '离线'
  }
  return statusMap[status] || '未知'
}

// 获取进度条颜色
const getProgressColor = (percentage: number) => {
  if (percentage < 60) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}

// 获取事件类型
const getEventType = (type: string) => {
  const typeMap: Record<string, string> = {
    Normal: 'success',
    Warning: 'warning',
    Error: 'danger'
  }
  return typeMap[type] || 'info'
}

// 初始化
onMounted(() => {
  fetchClusterDetail()
  fetchClusterStats()
  fetchRecentEvents()
})
</script>

<style scoped>
.cluster-detail {
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

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
  min-width: 100px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
}

.stat-card {
  transition: transform 0.2s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 0;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-icon.nodes {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.pods {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.namespaces {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-icon.services {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-info h3 {
  margin: 0 0 4px 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.stat-info p {
  margin: 0;
  color: #7f8c8d;
  font-size: 14px;
}

.stat-footer {
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
}

.resource-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
}

.resource-item {
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
}

.resource-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.resource-value {
  font-weight: 600;
  color: #2c3e50;
}

.events-card {
  min-height: 300px;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .resource-grid {
    grid-template-columns: 1fr;
  }
  
  .info-grid {
    grid-template-columns: 1fr;
  }
}
</style>