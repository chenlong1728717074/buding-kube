<template>
  <div class="home-page">
    <!-- 欢迎区域 -->
    <div class="welcome-section">
      <h1>欢迎，{{ userStore.userInfo?.username }}</h1>
      <p class="welcome-desc">Buding K8s 多集群管理平台</p>
    </div>

    <!-- 模块卡片 -->
    <div class="module-grid">
      <!-- 集群管理 -->
      <div class="module-card" @click="goToClusterManagement">
        <div class="module-icon cluster">
          <el-icon :size="40"><Monitor /></el-icon>
        </div>
        <div class="module-info">
          <h2>集群管理</h2>
          <p>管理和监控您的 Kubernetes 集群</p>
          <div class="module-stats">
            <span class="stat-item">
              <span class="stat-value">{{ clusterCount }}</span>
              <span class="stat-label">个集群</span>
            </span>
          </div>
        </div>
      </div>

      <!-- 用户管理 (仅管理员可见) -->
      <div v-if="userStore.isAdmin" class="module-card" @click="goToUserManagement">
        <div class="module-icon user">
          <el-icon :size="40"><User /></el-icon>
        </div>
        <div class="module-info">
          <h2>用户管理</h2>
          <p>管理系统用户和权限</p>
          <div class="module-stats">
            <span class="stat-item">
              <span class="stat-value">{{ userCount }}</span>
              <span class="stat-label">个用户</span>
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- 快速访问 -->
    <div class="quick-access-section">
      <h3>快速访问</h3>
      <div class="quick-links">
        <el-link :underline="false" @click="openDocs">
          <el-icon><Document /></el-icon>
          查看文档
        </el-link>
        <el-link :underline="false" href="https://github.com/chenlong1728717074/buding-kube" target="_blank">
          <el-icon><Link /></el-icon>
          GitHub
        </el-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { Monitor, User, Document, Link } from '@element-plus/icons-vue'
import { clusterApi } from '@/api/cluster'

const router = useRouter()
const userStore = useUserStore()

const clusterCount = ref(0)
const userCount = ref(0)

const loadStats = async () => {
  try {
    const response = await clusterApi.getClusters({ page: 1, pageSize: 100 })
    clusterCount.value = response.data.total
    userCount.value = 8 // TODO: 从用户 API 获取
  } catch (error) {
    console.error('Failed to load stats:', error)
  }
}

const goToClusterManagement = () => {
  router.push('/cluster')
}

const goToUserManagement = () => {
  router.push('/user/management')
}

const openDocs = () => {
  window.open('https://kubernetes.io/zh-cn/docs', '_blank', 'noopener,noreferrer')
}

onMounted(() => {
  loadStats()
})
</script>

<style scoped>
.home-page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 60px 20px;
}

/* 欢迎区域 */
.welcome-section {
  text-align: center;
  margin-bottom: 60px;
}

.welcome-section h1 {
  font-size: 36px;
  font-weight: 600;
  color: #1e3a8a;
  margin: 0 0 12px 0;
}

.welcome-desc {
  font-size: 16px;
  color: #64748b;
  margin: 0;
}

/* 模块卡片网格 */
.module-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 32px;
  margin-bottom: 60px;
}

.module-card {
  background: #fff;
  border-radius: 18px;
  padding: 40px 32px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.06);
  border: 1px solid rgba(59, 130, 246, 0.12);
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  gap: 24px;
}

.module-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 32px rgba(59, 130, 246, 0.15);
}

.module-icon {
  width: 80px;
  height: 80px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.module-icon.cluster {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
}

.module-icon.user {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
}

.module-info {
  flex: 1;
}

.module-info h2 {
  font-size: 24px;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 8px 0;
}

.module-info p {
  font-size: 14px;
  color: #6b7280;
  margin: 0 0 20px 0;
}

.module-stats {
  display: flex;
  gap: 24px;
}

.stat-item {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #3b82f6;
}

.stat-label {
  font-size: 14px;
  color: #9ca3af;
}

/* 快速访问 */
.quick-access-section {
  text-align: center;
}

.quick-access-section h3 {
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 20px 0;
}

.quick-links {
  display: flex;
  justify-content: center;
  gap: 24px;
}

.quick-links .el-link {
  font-size: 14px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 8px;
  background: rgba(59, 130, 246, 0.06);
  transition: all 0.2s ease;
}

.quick-links .el-link:hover {
  background: rgba(59, 130, 246, 0.12);
}

/* 响应式 */
@media (max-width: 768px) {
  .home-page {
    padding: 40px 16px;
  }

  .welcome-section {
    margin-bottom: 40px;
  }

  .welcome-section h1 {
    font-size: 28px;
  }

  .module-grid {
    grid-template-columns: 1fr;
    gap: 20px;
    margin-bottom: 40px;
  }

  .module-card {
    flex-direction: column;
    padding: 28px 24px;
  }

  .module-icon {
    width: 64px;
    height: 64px;
  }

  .quick-links {
    flex-direction: column;
    gap: 12px;
  }
}
</style>
