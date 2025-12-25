<template>
  <div class="cluster-overview-page">
    <!-- 集群信息卡片 -->
    <el-card class="info-card">
      <template #header>
        <div class="card-header">
          <span>集群信息</span>
        </div>
      </template>
      <div class="info-grid">
        <div class="info-item">
          <span class="label">集群名称</span>
          <span class="value">{{ clusterStore.currentClusterName }}</span>
        </div>
        <div class="info-item">
          <span class="label">API Server</span>
          <span class="value">{{ clusterStore.currentCluster?.apiServer }}</span>
        </div>
        <div class="info-item">
          <span class="label">状态</span>
          <el-tag :type="getStatusType(currentStatus)" size="small">
            {{ getStatusText(currentStatus) }}
          </el-tag>
        </div>
        <div class="info-item">
          <span class="label">版本</span>
          <span class="value">v1.28.0</span>
        </div>
      </div>
    </el-card>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <el-card class="stat-card" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon node">
            <el-icon :size="28"><Monitor /></el-icon>
          </div>
          <div class="stat-info">
            <h3>{{ stats.nodeCount }}</h3>
            <p>节点</p>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon pod">
            <el-icon :size="28"><Box /></el-icon>
          </div>
          <div class="stat-info">
            <h3>{{ stats.podCount }}</h3>
            <p>Pod</p>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon namespace">
            <el-icon :size="28"><FolderOpened /></el-icon>
          </div>
          <div class="stat-info">
            <h3>{{ stats.namespaceCount }}</h3>
            <p>命名空间</p>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card" shadow="hover">
        <div class="stat-content">
          <div class="stat-icon deployment">
            <el-icon :size="28"><Grid /></el-icon>
          </div>
          <div class="stat-info">
            <h3>{{ stats.deploymentCount }}</h3>
            <p>Deployment</p>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 快速操作 -->
    <el-card class="action-card">
      <template #header>
        <div class="card-header">
          <span>快速操作</span>
        </div>
      </template>
      <div class="action-grid">
        <div class="action-item" @click="$router.push(`/cluster/${clusterStore.currentClusterId}/namespace`)">
          <el-icon :size="24"><FolderOpened /></el-icon>
          <span>命名空间</span>
        </div>
        <div class="action-item" @click="$router.push(`/cluster/${clusterStore.currentClusterId}/workload/deployment`)">
          <el-icon :size="24"><Grid /></el-icon>
          <span>工作负载</span>
        </div>
        <div class="action-item" @click="$router.push(`/cluster/${clusterStore.currentClusterId}/service/service`)">
          <el-icon :size="24"><Connection /></el-icon>
          <span>服务</span>
        </div>
        <div class="action-item" @click="$router.push(`/cluster/${clusterStore.currentClusterId}/config/configmap`)">
          <el-icon :size="24"><Setting /></el-icon>
          <span>配置</span>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useClusterStore } from '@/stores/cluster'
import { Monitor, Box, FolderOpened, Grid, Connection, Setting } from '@element-plus/icons-vue'

const clusterStore = useClusterStore()

const currentStatus = computed(() => clusterStore.currentCluster?.status || 'Active')

const getStatusType = (status: string) => {
  const statusMap: Record<string, any> = {
    Active: 'success',
    Running: 'success',
    Pending: 'warning',
    Unknown: 'info',
    Error: 'danger',
    Failed: 'danger'
  }
  return statusMap[status] || 'info'
}

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    Active: '活跃',
    Running: '运行中',
    Pending: '等待中',
    Unknown: '未知',
    Error: '异常',
    Failed: '失败'
  }
  return statusMap[status] || status
}

const stats = ref({
  nodeCount: 0,
  podCount: 0,
  namespaceCount: 0,
  deploymentCount: 0
})

const loadStats = async () => {
  try {
    // TODO: 调用 API 获取集群统计数据
    stats.value = {
      nodeCount: 5,
      podCount: 32,
      namespaceCount: 8,
      deploymentCount: 12
    }
  } catch (error) {
    console.error('Failed to load stats:', error)
  }
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.cluster-overview-page {
  padding: 24px;
}

.info-card {
  margin-bottom: 24px;
}

.card-header {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-item .label {
  font-size: 13px;
  color: #9ca3af;
}

.info-item .value {
  font-size: 14px;
  color: #1f2937;
  font-weight: 500;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  cursor: default;
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 10px 0;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.stat-icon.node {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
}

.stat-icon.pod {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}

.stat-icon.namespace {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
}

.stat-icon.deployment {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
}

.stat-info h3 {
  font-size: 28px;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 4px 0;
}

.stat-info p {
  font-size: 14px;
  color: #6b7280;
  margin: 0;
}

.action-card {
  margin-bottom: 24px;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 16px;
}

.action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 24px 16px;
  background: rgba(59, 130, 246, 0.04);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-item:hover {
  background: rgba(59, 130, 246, 0.1);
  transform: translateY(-2px);
}

.action-item .el-icon {
  color: #3b82f6;
}

.action-item span {
  font-size: 14px;
  color: #1f2937;
  font-weight: 500;
}

/* 响应式 */
@media (max-width: 768px) {
  .cluster-overview-page {
    padding: 16px;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .action-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
