<template>
  <div class="dashboard">
    <!-- 欢迎信息 -->
    <div class="welcome-section">
      <h1>欢迎回来，{{ userStore.userInfo?.username }}！</h1>
      <p class="welcome-text">这里是您的Kubernetes集群管理仪表盘</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-icon cluster">
            <el-icon size="24"><Monitor /></el-icon>
          </div>
          <div class="stat-info">
            <h3>{{ clusterCount }}</h3>
            <p>集群总数</p>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-icon node">
            <el-icon size="24"><Monitor /></el-icon>
          </div>
          <div class="stat-info">
            <h3>{{ nodeCount }}</h3>
            <p>节点总数</p>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card">
        <div class="stat-content">
          <div class="stat-icon pod">
            <el-icon size="24"><Box /></el-icon>
          </div>
          <div class="stat-info">
            <h3>{{ podCount }}</h3>
            <p>Pod总数</p>
          </div>
        </div>
      </el-card>

      <el-card class="stat-card" v-if="userStore.isAdmin">
        <div class="stat-content">
          <div class="stat-icon user">
            <el-icon size="24"><User /></el-icon>
          </div>
          <div class="stat-info">
            <h3>{{ userCount }}</h3>
            <p>用户总数</p>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 快速操作 -->
    <div class="quick-actions">
      <h2>快速操作</h2>
      <div class="action-grid">
        <el-card class="action-card" @click="$router.push('/clusters')">
          <div class="action-content">
            <el-icon size="32" class="action-icon"><Monitor /></el-icon>
            <h3>管理集群</h3>
            <p>查看和管理Kubernetes集群</p>
          </div>
        </el-card>

        <el-card class="action-card" v-if="userStore.isAdmin" @click="$router.push('/users')">
          <div class="action-content">
            <el-icon size="32" class="action-icon"><User /></el-icon>
            <h3>用户管理</h3>
            <p>管理系统用户和权限</p>
          </div>
        </el-card>

        <el-card class="action-card">
          <div class="action-content">
            <el-icon size="32" class="action-icon"><Document /></el-icon>
            <h3>查看文档</h3>
            <p>阅读使用指南和API文档</p>
          </div>
        </el-card>
      </div>
    </div>

    <!-- 系统状态 -->
    <div class="system-status">
      <h2>系统状态</h2>
      <el-row :gutter="20">
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>集群健康状态</span>
            </template>
            <div class="status-item">
              <span class="status-label">健康集群:</span>
              <el-tag type="success">{{ healthyClusterCount }}</el-tag>
            </div>
            <div class="status-item">
              <span class="status-label">异常集群:</span>
              <el-tag type="danger">{{ unhealthyClusterCount }}</el-tag>
            </div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card>
            <template #header>
              <span>最近活动</span>
            </template>
            <div class="activity-list">
              <div class="activity-item" v-for="activity in recentActivities" :key="activity.id">
                <span class="activity-time">{{ activity.time }}</span>
                <span class="activity-desc">{{ activity.description }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import {
  Monitor,
  Box,
  User,
  Document
} from '@element-plus/icons-vue'

const userStore = useUserStore()

// 统计数据
const clusterCount = ref(0)
const nodeCount = ref(0)
const podCount = ref(0)
const userCount = ref(0)
const healthyClusterCount = ref(0)
const unhealthyClusterCount = ref(0)

// 最近活动
const recentActivities = ref([
  {
    id: 1,
    time: '10:30',
    description: '集群 prod-cluster 状态正常'
  },
  {
    id: 2,
    time: '09:15',
    description: '用户 admin 登录系统'
  },
  {
    id: 3,
    time: '08:45',
    description: '新增节点 worker-node-3'
  }
])

// 加载仪表盘数据
const loadDashboardData = async () => {
  try {
    // TODO: 调用API获取真实数据
    // 这里使用模拟数据
    clusterCount.value = 3
    nodeCount.value = 12
    podCount.value = 45
    userCount.value = 8
    healthyClusterCount.value = 2
    unhealthyClusterCount.value = 1
  } catch (error) {
    console.error('加载仪表盘数据失败:', error)
  }
}

onMounted(() => {
  loadDashboardData()
})
</script>

<style scoped>
.dashboard {
  max-width: 1200px;
  margin: 0 auto;
}

.welcome-section {
  margin-bottom: 30px;
}

.welcome-section h1 {
  font-size: 28px;
  color: #303133;
  margin: 0 0 8px 0;
}

.welcome-text {
  font-size: 16px;
  color: #606266;
  margin: 0;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 40px;
}

.stat-card {
  cursor: default;
}

.stat-content {
  display: flex;
  align-items: center;
  padding: 10px 0;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  color: white;
}

.stat-icon.cluster {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.node {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.pod {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-icon.user {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-info h3 {
  font-size: 32px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 4px 0;
}

.stat-info p {
  font-size: 14px;
  color: #909399;
  margin: 0;
}

.quick-actions {
  margin-bottom: 40px;
}

.quick-actions h2 {
  font-size: 20px;
  color: #303133;
  margin: 0 0 20px 0;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 20px;
}

.action-card {
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.action-content {
  text-align: center;
  padding: 20px;
}

.action-icon {
  color: #409EFF;
  margin-bottom: 16px;
}

.action-content h3 {
  font-size: 18px;
  color: #303133;
  margin: 0 0 8px 0;
}

.action-content p {
  font-size: 14px;
  color: #606266;
  margin: 0;
}

.system-status h2 {
  font-size: 20px;
  color: #303133;
  margin: 0 0 20px 0;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.status-item:last-child {
  margin-bottom: 0;
}

.status-label {
  font-size: 14px;
  color: #606266;
}

.activity-list {
  max-height: 200px;
  overflow-y: auto;
}

.activity-item {
  display: flex;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
}

.activity-item:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.activity-time {
  font-size: 12px;
  color: #909399;
  margin-right: 12px;
  min-width: 40px;
}

.activity-desc {
  font-size: 14px;
  color: #606266;
  flex: 1;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .action-grid {
    grid-template-columns: 1fr;
  }
  
  .welcome-section h1 {
    font-size: 24px;
  }
}
</style>