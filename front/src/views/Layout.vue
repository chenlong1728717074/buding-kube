<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside width="250px" class="sidebar">
      <div class="logo">
        <h2>K8s管理平台</h2>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        class="sidebar-menu"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
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
          <el-menu-item index="/workload/deployment" @click="showComingSoon">
            <el-icon><Grid /></el-icon>
            <span>Deployment</span>
          </el-menu-item>
          <el-menu-item index="/workload/daemonset" @click="showComingSoon">
            <el-icon><Cpu /></el-icon>
            <span>DaemonSet</span>
          </el-menu-item>
          <el-menu-item index="/workload/statefulset" @click="showComingSoon">
            <el-icon><DataBoard /></el-icon>
            <span>StatefulSet</span>
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
        
        <el-sub-menu index="/service">
          <template #title>
            <el-icon><Connection /></el-icon>
            <span>服务发现</span>
          </template>
          <el-menu-item index="/service/service" @click="showComingSoon">
            <el-icon><Link /></el-icon>
            <span>Service</span>
          </el-menu-item>
          <el-menu-item index="/service/ingress" @click="showComingSoon">
            <el-icon><Share /></el-icon>
            <span>Ingress</span>
          </el-menu-item>
          <el-menu-item index="/service/endpoint" @click="showComingSoon">
            <el-icon><Position /></el-icon>
            <span>Endpoint</span>
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
          <span class="page-title">{{ pageTitle }}</span>
        </div>
        <div class="header-right">
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

      <!-- 主要内容 -->
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade-transform" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
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
import { ElMessage } from 'element-plus'

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
      userStore.logout()
      break
  }
}

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
  background-color: #304156;
  overflow: hidden;
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #2b3a4b;
  border-bottom: 1px solid #1f2d3d;
}

.logo h2 {
  color: #fff;
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
  height: 50px;
  line-height: 50px;
}

.sidebar-menu .el-menu-item:hover {
  background-color: #263445 !important;
}

.sidebar-menu .el-menu-item.is-active {
  background-color: #409EFF !important;
  color: #fff !important;
}

.sidebar-menu .el-menu-item.is-active span {
  color: #fff !important;
}

.sidebar-menu .el-menu-item.is-active .el-icon {
  color: #fff !important;
}

.sidebar-menu .el-sub-menu.is-active .el-sub-menu__title {
  color: #409EFF !important;
}

.sidebar-menu .el-sub-menu.is-active .el-sub-menu__title span {
  color: #409EFF !important;
}

.sidebar-menu .el-sub-menu.is-active .el-sub-menu__title .el-icon {
  color: #409EFF !important;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e6e6e6;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 60px !important;
}

.header-left .page-title {
  font-size: 18px;
  font-weight: 500;
  color: #303133;
}

.header-right .user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  color: #606266;
  font-size: 14px;
}

.header-right .user-info:hover {
  color: #409EFF;
}

.header-right .user-info .el-icon {
  margin: 0 4px;
}

.main-content {
  background-color: #f5f5f5;
  padding: 20px;
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