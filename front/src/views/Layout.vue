<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapsed ? '72px' : '250px'" class="sidebar">
      <div class="logo" @click="goHome" style="cursor: pointer;">
        <h2>{{ isCollapsed ? 'K' : 'K8s管理平台' }}</h2>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        class="sidebar-menu"
        :collapse="isCollapsed"
        :collapse-transition="false"
        unique-opened
      >
        <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/overview`">
          <el-icon><House /></el-icon>
          <span>集群概览</span>
        </el-menu-item>
        
        <el-sub-menu index="resource">
          <template #title>
            <el-icon><FolderOpened /></el-icon>
            <span>资源管理</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/namespace`">
            <el-icon><List /></el-icon>
            <span>命名空间</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/nodes`">
            <el-icon><Monitor /></el-icon>
            <span>节点管理</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/pod`">
            <el-icon><Box /></el-icon>
            <span>Pod管理</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="workload">
          <template #title>
            <el-icon><Operation /></el-icon>
            <span>工作负载</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/workload/deployment`">
            <el-icon><Grid /></el-icon>
            <span>Deployment</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/workload/daemonset`">
            <el-icon><Cpu /></el-icon>
            <span>DaemonSet</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/workload/statefulset`">
            <el-icon><DataBoard /></el-icon>
            <span>StatefulSet</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="service">
          <template #title>
            <el-icon><Connection /></el-icon>
            <span>服务发现</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/service/service`">
            <el-icon><Link /></el-icon>
            <span>Service</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/service/ingress`">
            <el-icon><Share /></el-icon>
            <span>Ingress</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/service/endpoint`">
            <el-icon><Position /></el-icon>
            <span>Endpoint</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/service/endpointslice`">
            <el-icon><Position /></el-icon>
            <span>EndpointSlice</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="config">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>配置管理</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/config/configmap`">
            <el-icon><Files /></el-icon>
            <span>ConfigMap</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/config/secret`">
            <el-icon><Key /></el-icon>
            <span>Secret</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/config/serviceaccount`">
            <el-icon><UserFilled /></el-icon>
            <span>ServiceAccount</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="job">
          <template #title>
            <el-icon><Timer /></el-icon>
            <span>任务调度</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/job/job`">
            <el-icon><Clock /></el-icon>
            <span>Job</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/job/cronjob`">
            <el-icon><AlarmClock /></el-icon>
            <span>CronJob</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="storage">
          <template #title>
            <el-icon><Folder /></el-icon>
            <span>存储管理</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/storage/pv`">
            <el-icon><Folder /></el-icon>
            <span>PersistentVolume</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/storage/pvc`">
            <el-icon><FolderAdd /></el-icon>
            <span>PersistentVolumeClaim</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/storage/storageclass`">
            <el-icon><Document /></el-icon>
            <span>StorageClass</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="advanced">
          <template #title>
            <el-icon><MoreFilled /></el-icon>
            <span>高级功能</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/resource/crd`">
            <el-icon><Document /></el-icon>
            <span>CRD</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/resource/tools`">
            <el-icon><MoreFilled /></el-icon>
            <span>工具</span>
          </el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航 -->
      <el-header class="header">
        <div class="header-left">
          <el-button text class="collapse-btn" @click="toggleCollapse">
            <el-icon><Fold v-if="!isCollapsed" /><Expand v-else /></el-icon>
          </el-button>
          <div class="title-stack">
            <!-- 集群切换下拉选择器 -->
            <el-dropdown trigger="click" @command="switchCluster" class="cluster-dropdown">
              <div class="cluster-selector">
                <el-icon><Monitor /></el-icon>
                <span>{{ clusterStore.currentClusterName }}</span>
                <el-icon class="arrow"><ArrowDown /></el-icon>
              </div>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item 
                    v-for="cluster in clusterList" 
                    :key="cluster.id"
                    :command="cluster"
                    :disabled="cluster.id === clusterStore.currentClusterId"
                  >
                    <span class="cluster-name">{{ cluster.name }}</span>
                    <el-tag v-if="cluster.id === clusterStore.currentClusterId" size="small" type="success" style="margin-left: 8px;">当前</el-tag>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          <el-button text class="manage-cluster-btn" @click="goToClusterManage">
            <el-icon><Setting /></el-icon>
            <span>集群管理</span>
          </el-button>
        </div>
        <div class="header-right">
          <div class="header-extras">
            <div class="clock">
              <el-icon class="clock-icon"><Timer /></el-icon>
              <span class="clock-time">{{ nowTimeText }}</span>
              <span class="clock-split">·</span>
              <span class="clock-date">{{ nowDateText }}</span>
            </div>
            <span class="header-divider" />
            <el-dropdown trigger="click" @command="handleToolsCommand" class="tools-dropdown">
              <el-button text circle class="icon-btn" title="工具">
                <el-icon><MoreFilled /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="refresh">
                    <el-icon><Refresh /></el-icon>
                    刷新
                  </el-dropdown-item>
                  <el-dropdown-item command="fullscreen">
                    <el-icon><FullScreen /></el-icon>
                    {{ fullscreenText }}
                  </el-dropdown-item>
                  <el-dropdown-item command="notify">
                    <el-icon><Bell /></el-icon>
                    通知
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <el-link
              :underline="false"
              href="https://github.com/chenlong1728717074/buding-kube"
              target="_blank"
              rel="noopener noreferrer"
              type="primary"
              class="gh-link"
            >
              GitHub
            </el-link>
          </div>
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-icon><Avatar /></el-icon>
              {{ userStore.userInfo?.username }}
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人信息</el-dropdown-item>
                <el-dropdown-item command="userManagement" v-if="isAdmin">用户管理</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 顶部标签导航 -->
      <TagsView />

      <!-- 主要内容 -->
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade-transform" mode="out-in">
            <keep-alive>
              <component :is="Component" :key="route.fullPath" />
            </keep-alive>
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useClusterStore } from '@/stores/cluster'
import {
  House,
  Monitor,
  User,
  Avatar,
  Bell,
  ArrowDown,
  ArrowRight,
  List,
  UserFilled,
  Box,
  Operation,
  Document,
  Timer,
  Clock,
  AlarmClock,
  Connection,
  Link,
  Share,
  Position,
  Setting,
  Files,
  Key,
  FolderOpened,
  Folder,
  FolderAdd,
  Grid,
  Cpu,
  DataBoard,
  Refresh,
  FullScreen,
  MoreFilled
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import TagsView from '@/components/TagsView.vue'
import { clusterApi } from '@/api/cluster'

import { Fold, Expand } from '@element-plus/icons-vue'
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const clusterStore = useClusterStore()
const nowDateText = ref('')
const nowTimeText = ref('')
const isFullscreen = ref(false)
const fullscreenText = computed(() => (isFullscreen.value ? '退出全屏' : '进入全屏'))
const clusterList = ref<any[]>([])

// 加载集群列表
const loadClusterList = async () => {
  try {
    const response = await clusterApi.getClusters({ page: 1, pageSize: 100 })
    // 后端返回的是 items 而不是 list
    const { items, list } = response.data
    const clusterData = items || list || []
    clusterList.value = clusterData.map((cluster: any) => ({
      id: cluster.id || cluster.name,
      name: cluster.alias || cluster.name,
      status: cluster.status
    }))
  } catch (error) {
    console.error('Failed to load cluster list:', error)
  }
}

// 切换集群
const switchCluster = (command: any) => {
  const cluster = command
  clusterStore.setCurrentCluster({
    id: cluster.id,
    name: cluster.name,
    apiServer: '',
    status: cluster.status
  })
  
  // 刷新当前页面或跳转到集群概览
  const currentPath = route.path
  if (currentPath.includes('/cluster/')) {
    const newPath = currentPath.replace(/\/cluster\/[^/]+/, `/cluster/${cluster.id}`)
    router.push(newPath)
  } else {
    router.push(`/cluster/${cluster.id}/overview`)
  }
}

// 返回主页
const goHome = () => {
  router.push('/home')
}

// 前往集群管理页面
const goToClusterManage = () => {
  router.push('/cluster')
}

// 当前激活的菜单
const activeMenu = computed(() => {
  return route.path
})

// 页面标题
const pageTitle = computed(() => {
  return route.meta.title || '集群概览'
})

const breadcrumbs = computed(() => {
  const result: Array<{ title: string; to?: string }> = []
  let acc = ''
  for (const r of route.matched) {
    const title = r.meta?.title
    if (!title) continue
    const rawPath = r.path || ''
    if (rawPath.startsWith('/')) {
      acc = rawPath
    } else if (rawPath) {
      acc = `${acc.replace(/\/$/, '')}/${rawPath}`.replace(/\/+/g, '/')
    }
    result.push({ title: String(title), to: acc.startsWith('/') ? acc : undefined })
  }
  return result
})

// 检查是否是管理员 - 使用store中的isAdmin
const isAdmin = computed(() => userStore.isAdmin)

// 处理下拉菜单命令
const handleCommand = (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/user/profile')
      break
    case 'userManagement':
      router.push('/user/management')
      break
    case 'logout':
      ElMessageBox.confirm('确定要退出登录吗？', '提示', { type: 'warning', confirmButtonText: '确认', cancelButtonText: '取消' })
        .then(async () => {
          await userStore.logout()
          const hour = new Date().getHours()
          const greet = hour < 12 ? '期待你下次回来' : hour < 18 ? '期待你下次回来' : '期待你下次回来'
          ElNotification({ message: `${greet}`, type: 'success', duration: 2500 })
        })
        .catch(() => {})
      break
  }
}

const isCollapsed = ref(false)
const toggleCollapse = () => {
  isCollapsed.value = !isCollapsed.value
  try {
    localStorage.setItem('sidebarCollapsed', isCollapsed.value ? '1' : '0')
  } catch {}
}

try {
  const saved = localStorage.getItem('sidebarCollapsed')
  isCollapsed.value = saved === '1'
} catch {}

let clockTimer: number | undefined
const updateClock = () => {
  const d = new Date()
  nowDateText.value = d.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', weekday: 'short' })
  nowTimeText.value = d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

const refreshPage = () => {
  window.location.reload()
}

const showNotifications = () => {
  ElMessage.info('暂无通知')
}

const handleToolsCommand = async (cmd: string) => {
  if (cmd === 'refresh') {
    refreshPage()
    return
  }
  if (cmd === 'fullscreen') {
    await toggleFullscreen()
    return
  }
  if (cmd === 'notify') {
    showNotifications()
  }
}

const toggleFullscreen = async () => {
  try {
    if (!document.fullscreenElement) {
      await document.documentElement.requestFullscreen()
      return
    }
    await document.exitFullscreen()
  } catch {
    ElMessage.error('全屏切换失败')
  }
}

const onFullscreenChange = () => {
  isFullscreen.value = !!document.fullscreenElement
}

onMounted(() => {
  updateClock()
  clockTimer = window.setInterval(updateClock, 1000)
  onFullscreenChange()
  document.addEventListener('fullscreenchange', onFullscreenChange)
  loadClusterList()
})
onUnmounted(() => {
  if (clockTimer) window.clearInterval(clockTimer)
  document.removeEventListener('fullscreenchange', onFullscreenChange)
})
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background: #ffffff;
  border-right: 1px solid rgba(59,130,246,0.16);
  overflow: hidden;
  border-radius: 18px;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ffffff;
  border-bottom: 1px solid rgba(59,130,246,0.16);
}

.logo h2 {
  color: var(--brand-500);
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.sidebar-menu {
  border: none;
  height: calc(100vh - 60px);
  overflow-y: auto;
}

.sidebar-menu .el-menu-item {
  height: 48px;
  line-height: 48px;
  border-radius: 16px;
  margin: 4px 10px;
}

.sidebar-menu .el-sub-menu__title {
  height: 48px;
  line-height: 48px;
  border-radius: 16px;
  margin: 4px 10px;
}

.sidebar-menu .el-menu-item:hover {
  background-color: #f5f8ff !important;
}

.sidebar-menu .el-menu-item.is-active {
  background-color: #eef2ff !important;
  color: var(--brand-600) !important;
}

.sidebar-menu .el-menu-item.is-active span { color: var(--brand-600) !important; }
.sidebar-menu .el-menu-item.is-active .el-icon { color: var(--brand-600) !important; }

.sidebar-menu .el-sub-menu.is-active .el-sub-menu__title {
  color: #93c5fd !important;
}

.sidebar-menu .el-sub-menu.is-active .el-sub-menu__title span {
  color: #409EFF !important;
}

.sidebar-menu .el-sub-menu.is-active .el-sub-menu__title .el-icon {
  color: #409EFF !important;
}

.header {
  background: #ffffff;
  border-bottom: 1px solid rgba(59,130,246,0.16);
  box-shadow: var(--shadow-card);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 60px !important;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.title-stack {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.cluster-dropdown {
  cursor: pointer;
}

.cluster-selector {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: rgba(59, 130, 246, 0.08);
  border-radius: 8px;
  font-size: 14px;
  font-weight: 600;
  color: #3b82f6;
  transition: all 0.2s ease;
}

.cluster-selector:hover {
  background: rgba(59, 130, 246, 0.15);
}

.cluster-selector .arrow {
  font-size: 12px;
}

.cluster-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-width: 180px;
}

.cluster-name {
  flex: 1;
}

.manage-cluster-btn {
  margin-left: 12px;
  color: #64748b;
  font-size: 14px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.manage-cluster-btn:hover {
  background: rgba(100, 116, 139, 0.08);
  color: #475569;
}

.header-left .page-title {
  font-size: 20px;
  font-weight: 600;
  color: #1e3a8a;
}

.breadcrumb {
  font-size: 12px;
  color: #64748b;
}

.collapse-btn {
  margin-right: 8px;
  color: #64748b;
}

.header-right .user-info {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  color: #1e3a8a;
  font-size: 14px;
  height: 34px;
  padding: 0 12px;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.04);
  border: 1px solid rgba(15, 23, 42, 0.08);
}
.header-extras {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-right: 12px;
}

.clock {
  display: inline-flex;
  align-items: center;
  height: 34px;
  padding: 0 10px;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.04);
  border: 1px solid rgba(15, 23, 42, 0.08);
  color: #0f172a;
  gap: 8px;
}

.clock-icon {
  color: rgba(15, 23, 42, 0.65);
}

.clock-time {
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 0.2px;
  font-variant-numeric: tabular-nums;
}

.clock-split {
  color: rgba(15, 23, 42, 0.35);
  font-size: 12px;
  line-height: 1;
}

.clock-date {
  font-size: 12px;
  color: rgba(15, 23, 42, 0.65);
  line-height: 1;
}

.icon-btn {
  color: rgba(15, 23, 42, 0.75);
  width: 34px;
  height: 34px;
  border-radius: 12px;
}

.icon-btn:hover {
  background: rgba(15, 23, 42, 0.06);
}
.gh-link {
  color: rgba(15, 23, 42, 0.85);
  display: inline-flex;
  align-items: center;
  height: 34px;
  padding: 0 12px;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.04);
  border: 1px solid rgba(15, 23, 42, 0.08);
}
.gh-link:hover {
  background: rgba(15, 23, 42, 0.06);
  color: #0f172a;
}

.header-divider {
  width: 1px;
  height: 18px;
  background: rgba(15, 23, 42, 0.10);
  display: inline-block;
}
.header-right .user-info:hover {
  background: rgba(15, 23, 42, 0.06);
  color: #0f172a;
}

.header-right .user-info .el-icon {
  margin: 0 4px;
}

.main-content {
  background: linear-gradient(180deg, #f5faff 0%, #fbfdff 100%);
  padding: 24px;
  overflow-y: auto;
}

@media (max-width: 768px) {
  .breadcrumb {
    display: none;
  }
  .clock-date {
    display: none;
  }
  .clock-split {
    display: none;
  }
  .gh-link {
    display: none;
  }
}

/* 页面切换动画 */
.fade-transform-enter-active,
.fade-transform-leave-active {
  transition: all 0.3s;
}

.fade-transform-enter-from {
  opacity: 0;
  transform: translateX(30px);
}

.fade-transform-leave-to {
  opacity: 0;
  transform: translateX(-30px);
}
</style>
