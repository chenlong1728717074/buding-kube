<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
<el-aside :width="isCollapsed ? '72px' : '250px'" class="sidebar">
      <div class="logo">
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
        <el-menu-item index="/dashboard">
          <el-icon><House /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        
        <el-sub-menu index="/cluster">
          <template #title>
            <el-icon><Monitor /></el-icon>
            <span>集群管理</span>
          </template>
          <el-menu-item index="/cluster/list">
            <el-icon><List /></el-icon>
            <span>集群列表</span>
          </el-menu-item>
          <el-menu-item index="/cluster/nodes">
            <el-icon><Monitor /></el-icon>
            <span>节点列表</span>
          </el-menu-item>
          <el-menu-item index="/namespace">
            <el-icon><List /></el-icon>
            <span>命名空间</span>
          </el-menu-item>
          <el-menu-item index="/pod">
            <el-icon><Box /></el-icon>
            <span>Pod管理</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="/workload">
          <template #title>
            <el-icon><Operation /></el-icon>
            <span>工作负载</span>
          </template>
          <el-menu-item index="/workload/deployment">
            <el-icon><Grid /></el-icon>
            <span>Deployment</span>
          </el-menu-item>
          <el-menu-item index="/workload/daemonset">
            <el-icon><Cpu /></el-icon>
            <span>DaemonSet</span>
          </el-menu-item>
          <el-menu-item index="/workload/statefulset">
            <el-icon><DataBoard /></el-icon>
            <span>StatefulSet</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="/service">
          <template #title>
            <el-icon><Connection /></el-icon>
            <span>服务发现</span>
          </template>
          <el-menu-item index="/service/service">
            <el-icon><Link /></el-icon>
            <span>Service</span>
          </el-menu-item>
          <el-menu-item index="/service/ingress">
            <el-icon><Share /></el-icon>
            <span>Ingress</span>
          </el-menu-item>
          <el-menu-item index="/service/endpoint">
            <el-icon><Position /></el-icon>
            <span>Endpoint</span>
          </el-menu-item>
          <el-menu-item index="/service/endpointslice">
            <el-icon><Position /></el-icon>
            <span>EndpointSlice</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="/config">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>配置管理</span>
          </template>
          <el-menu-item index="/config/configmap" @click="showComingSoon">
            <el-icon><Files /></el-icon>
            <span>ConfigMap</span>
          </el-menu-item>
          <el-menu-item index="/config/secret" @click="showComingSoon">
            <el-icon><Key /></el-icon>
            <span>Secret</span>
          </el-menu-item>
          <el-menu-item index="/config/serviceaccount" @click="showComingSoon">
            <el-icon><UserFilled /></el-icon>
            <span>ServiceAccount</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="/job">
          <template #title>
            <el-icon><Timer /></el-icon>
            <span>任务调度</span>
          </template>
          <el-menu-item index="/job/job" @click="showComingSoon">
            <el-icon><Clock /></el-icon>
            <span>Job</span>
          </el-menu-item>
          <el-menu-item index="/job/cronjob" @click="showComingSoon">
            <el-icon><AlarmClock /></el-icon>
            <span>CronJob</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu index="/storage">
          <template #title>
            <el-icon><FolderOpened /></el-icon>
            <span>存储管理</span>
          </template>
          <el-menu-item index="/storage/pv" @click="showComingSoon">
            <el-icon><Folder /></el-icon>
            <span>PersistentVolume</span>
          </el-menu-item>
          <el-menu-item index="/storage/pvc" @click="showComingSoon">
            <el-icon><FolderAdd /></el-icon>
            <span>PersistentVolumeClaim</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-sub-menu v-if="userStore.isAdmin" index="/user">
          <template #title>
            <el-icon><User /></el-icon>
            <span>用户管理</span>
          </template>
          <el-menu-item index="/user/list">
            <el-icon><UserFilled /></el-icon>
            <span>用户列表</span>
          </el-menu-item>
          <el-menu-item index="/user/profile">
            <el-icon><Avatar /></el-icon>
            <span>个人信息</span>
          </el-menu-item>
        </el-sub-menu>
        
        <el-menu-item v-else index="/user/profile">
          <el-icon><Avatar /></el-icon>
          <span>个人信息</span>
        </el-menu-item>
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
          <span class="page-title">{{ pageTitle }}</span>
        </div>
        <div class="header-right">
          <div class="header-extras">
            <el-tag size="small" type="info" effect="plain" class="capsule">{{ nowText }}</el-tag>
            <el-link :underline="false" href="https://github.com/chenlong1728717074/buding-kube" target="_blank" type="primary" class="gh-link">
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
import {
  House,
  Monitor,
  User,
  Avatar,
  ArrowDown,
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
  DataBoard
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import TagsView from '@/components/TagsView.vue'

import { Fold, Expand } from '@element-plus/icons-vue'
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

// 当前激活的菜单
const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/dashboard')) return '/dashboard'
  if (path.startsWith('/cluster')) return path
  if (path.startsWith('/namespace')) return '/namespace'
  if (path.startsWith('/pod')) return '/pod'
  if (path.startsWith('/workload')) return path
  if (path.startsWith('/job')) return path
  if (path.startsWith('/service')) return path
  if (path.startsWith('/config')) return path
  if (path.startsWith('/storage')) return path
  if (path.startsWith('/user')) return path
  return '/dashboard'
})

// 页面标题
const pageTitle = computed(() => {
  return route.meta.title || '仪表盘'
})

// 处理下拉菜单命令
const handleCommand = (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/user/profile')
      break
    case 'logout':
      ElMessageBox.confirm('确定要退出登录吗？', '提示', { type: 'warning', confirmButtonText: '确认', cancelButtonText: '取消' })
        .then(async () => {
          await userStore.logout()
          const hour = new Date().getHours()
          const greet = hour < 12 ? '期待你下次回来 ☀️' : hour < 18 ? '期待你下次回来 🌤️' : '期待你下次回来 🌛'
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

// 显示暂未开放提示
const showComingSoon = (event: Event) => {
  event.preventDefault()
  event.stopPropagation()
  ElMessage({
    message: '该功能暂未开放，敬请期待！',
    type: 'info',
    duration: 2000
  })
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background: #ffffff;
  border-right: 1px solid rgba(59,130,246,0.16);
  overflow: hidden;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ffffff;
  border-bottom: 1px solid rgba(59,130,246,0.12);
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
  border-radius: 12px;
  margin: 4px 10px;
}

.sidebar-menu .el-sub-menu__title {
  height: 48px;
  line-height: 48px;
  border-radius: 12px;
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
  height: 64px !important;
}

.header-left .page-title {
  font-size: 20px;
  font-weight: 600;
  color: #1e3a8a;
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
  padding: 6px 10px;
  border-radius: 999px;
  background: rgba(59, 130, 246, 0.12);
  border: 1px solid rgba(59, 130, 246, 0.25);
}
.header-extras {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-right: 12px;
}
.gh-link {
  color: #1e3a8a;
}
.header-right .user-info:hover {
  background: rgba(59, 130, 246, 0.18);
  color: #0f1e5a;
}

.header-right .user-info .el-icon {
  margin: 0 4px;
}

.main-content {
  background: linear-gradient(180deg, #f5faff 0%, #fbfdff 100%);
  padding: 24px;
  overflow-y: auto;
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
const nowText = ref('')
let clockTimer: number | undefined
const updateClock = () => {
  const d = new Date()
  nowText.value = d.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
onMounted(() => {
  updateClock()
  clockTimer = window.setInterval(updateClock, 1000)
})
onUnmounted(() => {
  if (clockTimer) window.clearInterval(clockTimer)
})
