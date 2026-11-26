<template>
  <el-container class="layout-container">
    <!-- 深海蓝侧边栏 -->
    <el-aside :width="isCollapsed ? '72px' : '228px'" class="sidebar">
      <!-- 底部柔和 K8s 舵轮水印 -->
      <svg v-if="!isCollapsed" class="sidebar-watermark" viewBox="0 0 120 120" aria-hidden="true">
        <g
          fill="none"
          stroke="rgba(96, 165, 250, 0.16)"
          stroke-width="7.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <circle cx="60" cy="60" r="34" />
          <circle cx="60" cy="60" r="11" />
          <line x1="60" y1="10" x2="60" y2="49" />
          <line x1="99" y1="29" x2="68.5" y2="53" />
          <line x1="109" y1="71" x2="70.5" y2="62.5" />
          <line x1="82" y1="105" x2="64.8" y2="70" />
          <line x1="38" y1="105" x2="55.2" y2="70" />
          <line x1="11" y1="71" x2="49.5" y2="62.5" />
          <line x1="21" y1="29" x2="51.5" y2="53" />
        </g>
      </svg>

      <div class="logo" :class="{ 'is-collapsed': isCollapsed }" @click="goHome" style="cursor: pointer;">
        <div class="logo-icon">
          <K8sIcon name="k8s" :size="30" />
        </div>
        <h2 v-if="!isCollapsed">Kubernetes</h2>
      </div>
      <el-menu
        :default-active="activeMenu"
        @select="handleMenuSelect"
        class="sidebar-menu"
        :collapse="isCollapsed"
        :collapse-transition="false"
        unique-opened
        popper-class="k8s-dark-menu-popper"
      >
        <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/overview`">
          <el-icon class="k8s-menu-icon"><K8sIcon name="k8s" :size="20" /></el-icon>
          <span>集群概览</span>
        </el-menu-item>

        <el-sub-menu index="resource">
          <template #title>
            <el-icon class="k8s-menu-icon"><K8sIcon name="group" :size="20" /></el-icon>
            <span>资源管理</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/namespace`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="ns" :size="19" /></el-icon>
            <span>命名空间</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/nodes`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="node" :size="19" /></el-icon>
            <span>节点管理</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/pod`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="pod" :size="19" /></el-icon>
            <span>Pod管理</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="workload">
          <template #title>
            <el-icon class="k8s-menu-icon"><K8sIcon name="rs" :size="20" /></el-icon>
            <span>工作负载</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/workload/deployment`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="deploy" :size="19" /></el-icon>
            <span>Deployment</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/workload/daemonset`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="ds" :size="19" /></el-icon>
            <span>DaemonSet</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/workload/statefulset`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="sts" :size="19" /></el-icon>
            <span>StatefulSet</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/workload/replicaset`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="rs" :size="19" /></el-icon>
            <span>ReplicaSet</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/workload/hpa`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="hpa" :size="19" /></el-icon>
            <span>弹性伸缩</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="service">
          <template #title>
            <el-icon class="k8s-menu-icon"><K8sIcon name="netpol" :size="20" /></el-icon>
            <span>服务发现</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/service/service`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="svc" :size="19" /></el-icon>
            <span>Service</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/service/ingress`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="ing" :size="19" /></el-icon>
            <span>Ingress</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/service/endpoint`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="ep" :size="19" /></el-icon>
            <span>Endpoint</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/service/endpointslice`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="ep" :size="19" /></el-icon>
            <span>EndpointSlice</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/service/ingressclass`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="ing" :size="19" /></el-icon>
            <span>IngressClass</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="config">
          <template #title>
            <el-icon class="k8s-menu-icon"><K8sIcon name="cm" :size="20" /></el-icon>
            <span>配置管理</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/config/configmap`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="cm" :size="19" /></el-icon>
            <span>ConfigMap</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/config/secret`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="secret" :size="19" /></el-icon>
            <span>Secret</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/config/serviceaccount`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="sa" :size="19" /></el-icon>
            <span>ServiceAccount</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="job">
          <template #title>
            <el-icon class="k8s-menu-icon"><K8sIcon name="cronjob" :size="20" /></el-icon>
            <span>任务调度</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/job/job`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="job" :size="19" /></el-icon>
            <span>Job</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/job/cronjob`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="cronjob" :size="19" /></el-icon>
            <span>CronJob</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="storage">
          <template #title>
            <el-icon class="k8s-menu-icon"><K8sIcon name="vol" :size="20" /></el-icon>
            <span>存储管理</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/storage/pv`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="pv" :size="19" /></el-icon>
            <span>PersistentVolume</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/storage/pvc`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="pvc" :size="19" /></el-icon>
            <span>PersistentVolumeClaim</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/storage/storageclass`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="sc" :size="19" /></el-icon>
            <span>StorageClass</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="policy">
          <template #title>
            <el-icon class="k8s-menu-icon"><K8sIcon name="quota" :size="20" /></el-icon>
            <span>策略管理</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/policy/resourcequota`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="quota" :size="19" /></el-icon>
            <span>资源配额</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/policy/limitrange`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="limits" :size="19" /></el-icon>
            <span>默认资源限制</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/policy/networkpolicy`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="netpol" :size="19" /></el-icon>
            <span>网络策略</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/policy/pdb`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="quota" :size="19" /></el-icon>
            <span>中断预算</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/policy/priorityclass`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="etcd" :size="19" /></el-icon>
            <span>优先级</span>
          </el-menu-item>
        </el-sub-menu>

        <el-sub-menu index="advanced">
          <template #title>
            <el-icon class="k8s-menu-icon"><K8sIcon name="etcd" :size="20" /></el-icon>
            <span>运维工具</span>
          </template>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/resource/crd`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="crd" :size="19" /></el-icon>
            <span>CRD</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/resource/events`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="etcd" :size="19" /></el-icon>
            <span>事件中心</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/resource/kubectl`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="k8s" :size="19" /></el-icon>
            <span>Web kubectl</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/resource/topology`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="netpol" :size="19" /></el-icon>
            <span>资源拓扑图</span>
          </el-menu-item>
          <el-menu-item :index="`/cluster/${clusterStore.currentClusterId}/resource/tools`">
            <el-icon class="k8s-menu-icon"><K8sIcon name="etcd" :size="19" /></el-icon>
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
                <K8sIcon name="k8s" :size="18" />
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
              <span class="clock-live-dot" />
              <span class="clock-time">{{ nowTimeText }}</span>
              <span class="clock-split">|</span>
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
              <span class="user-avatar-circle">{{ userInitial }}</span>
              <span class="user-name">{{ userStore.userInfo?.username }}</span>
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
            <!-- 列表页缓存, 从详情返回时保留筛选和分页; 详情 / 编辑页每次进入都重新加载 -->
            <keep-alive :exclude="NO_CACHE" :max="20">
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
import { useTagsStore } from '@/stores/tags'
import {
  Bell,
  ArrowDown,
  Setting,
  Refresh,
  FullScreen,
  MoreFilled,
  Fold,
  Expand
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import TagsView from '@/components/TagsView.vue'
import K8sIcon from '@/components/K8sIcon.vue'
import { clusterApi } from '@/api/cluster'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const clusterStore = useClusterStore()
const tagsStore = useTagsStore()
const nowDateText = ref('')
const nowTimeText = ref('')
const isFullscreen = ref(false)
const fullscreenText = computed(() => (isFullscreen.value ? '退出全屏' : '进入全屏'))
const userInitial = computed(() => {
  const name = userStore.userInfo?.username || 'A'
  return name.slice(0, 1).toUpperCase()
})
const clusterList = ref<any[]>([])

// 加载集群列表
const loadClusterList = async () => {
  try {
    const response = await clusterApi.getClusters({ page: 1, pageSize: 100 })
    // 后端返回的是 items 而不是 list
    const { items } = response.data
    const clusterData = items || []
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
  // 详情页的资源在新集群里不一定存在, 切换后打开当前标签所属的菜单页
  const currentPath = tagsStore.activeKey || route.path
  if (currentPath.includes('/cluster/')) {
    const newPath = currentPath.replace(/\/cluster\/[^/]+/, `/cluster/${cluster.id}`)
    tagsStore.openMenu(router, newPath)
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

// 当前激活的菜单: 当前标签所属的菜单(详情页高亮进入它的菜单)
const activeMenu = computed(() => tagsStore.activeKey || route.path)

// 菜单是唯一会打开新标签的入口
const handleMenuSelect = (index: string) => {
  if (index.startsWith('/')) tagsStore.openMenu(router, index)
}

// 不缓存的页面(按组件文件名匹配)
const NO_CACHE = /(Detail|Edit)$/

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
          ElNotification({ message: '期待你下次回来', type: 'success', duration: 2500 })
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
  position: relative;
  background: linear-gradient(180deg, #0b1f44 0%, #102a5c 52%, #0c2354 100%);
  border-right: 1px solid rgba(59, 130, 246, 0.14);
  box-shadow: 4px 0 24px rgba(10, 25, 60, 0.14);
  overflow: hidden;
  transition: width 0.24s ease;
}

.sidebar-watermark {
  position: absolute;
  left: -32px;
  bottom: -26px;
  width: 176px;
  height: 176px;
  pointer-events: none;
  z-index: 0;
}

.logo {
  position: relative;
  z-index: 2;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  padding: 0 18px;
  gap: 11px;
  background: transparent;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.logo.is-collapsed {
  justify-content: center;
  padding: 0;
}

.logo-icon {
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  filter: drop-shadow(0 4px 12px rgba(59, 130, 246, 0.45));
}

.logo h2 {
  color: #ffffff;
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 0.3px;
  white-space: nowrap;
}

.sidebar-menu {
  position: relative;
  z-index: 2;
  border: none;
  height: calc(100vh - 60px);
  overflow-y: auto;
  overflow-x: hidden;
  padding: 12px 0 24px;
  background: transparent !important;
  --el-menu-bg-color: transparent;
  --el-menu-text-color: rgba(219, 234, 254, 0.82);
  --el-menu-hover-bg-color: rgba(255, 255, 255, 0.08);
  --el-menu-active-color: #ffffff;
}

.sidebar-menu::-webkit-scrollbar {
  width: 4px;
}

.sidebar-menu::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.25);
  border-radius: 999px;
}

.sidebar-menu :deep(.el-menu) {
  background: rgba(6, 17, 42, 0.28) !important;
  border-radius: 12px;
  margin: 2px 10px 6px;
  padding: 4px 0;
}

.sidebar-menu :deep(.el-menu-item),
.sidebar-menu :deep(.el-sub-menu__title) {
  height: 44px;
  line-height: 44px;
  border-radius: 11px;
  margin: 3px 10px;
  color: rgba(219, 234, 254, 0.85);
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.sidebar-menu :deep(.el-sub-menu .el-menu-item) {
  height: 40px;
  line-height: 40px;
  margin: 2px 6px;
  padding-left: 38px !important;
  font-size: 13px;
  color: rgba(203, 213, 225, 0.86);
}

.sidebar-menu.el-menu--collapse :deep(.el-menu-item),
.sidebar-menu.el-menu--collapse :deep(.el-sub-menu__title) {
  margin: 4px 10px;
  padding: 0 !important;
  justify-content: center;
}

.sidebar-menu .k8s-menu-icon {
  width: 22px;
  height: 22px;
  margin-right: 10px;
  flex-shrink: 0;
}

.sidebar-menu.el-menu--collapse .k8s-menu-icon {
  margin-right: 0;
}

.sidebar-menu :deep(.el-sub-menu__icon-arrow) {
  color: rgba(148, 163, 184, 0.75);
}

.sidebar-menu :deep(.el-menu-item:hover),
.sidebar-menu :deep(.el-sub-menu__title:hover) {
  background-color: rgba(255, 255, 255, 0.09) !important;
  color: #ffffff !important;
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  background: linear-gradient(90deg, #3b82f6 0%, #2563eb 78%, rgba(37, 99, 235, 0.85) 100%) !important;
  color: #ffffff !important;
  font-weight: 600;
  box-shadow: 0 6px 18px rgba(37, 99, 235, 0.36);
}

.sidebar-menu :deep(.el-menu-item.is-active span) {
  color: #ffffff !important;
}

.sidebar-menu :deep(.el-sub-menu.is-active > .el-sub-menu__title) {
  color: #ffffff !important;
}

.sidebar-menu :deep(.el-sub-menu.is-active > .el-sub-menu__title span) {
  color: #ffffff !important;
}

.header {
  background: rgba(255, 255, 255, 0.94);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(226, 232, 240, 0.9);
  box-shadow: 0 2px 12px rgba(15, 23, 42, 0.03);
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
  gap: 8px;
  cursor: pointer;
  color: #1e293b;
  font-size: 14px;
  font-weight: 500;
  height: 36px;
  padding: 0 12px 0 6px;
  border-radius: 999px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  transition: all 0.2s ease;
}

.user-avatar-circle {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: #ffffff;
  font-size: 13px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.user-name {
  color: #1e293b;
}

.header-extras {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-right: 6px;
}

.clock {
  display: inline-flex;
  align-items: center;
  height: 34px;
  padding: 0 14px;
  border-radius: 999px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  color: #0f172a;
  gap: 8px;
}

.clock-live-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #22c55e;
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.55);
  flex-shrink: 0;
}

.clock-time {
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.2px;
  font-variant-numeric: tabular-nums;
}

.clock-split {
  color: #cbd5e1;
  font-size: 12px;
  line-height: 1;
}

.clock-date {
  font-size: 12px;
  color: #64748b;
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
  background: linear-gradient(155deg, #f2f6fe 0%, #eaf1fc 52%, #f4f7fe 100%);
  padding: 20px 24px;
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

<style>
.k8s-dark-menu-popper .el-menu {
  background: #0e2552 !important;
  border: 1px solid rgba(96, 165, 250, 0.2) !important;
  border-radius: 12px !important;
  padding: 6px !important;
}

.k8s-dark-menu-popper .el-menu-item {
  color: rgba(219, 234, 254, 0.88) !important;
  border-radius: 8px !important;
  height: 40px !important;
  line-height: 40px !important;
}

.k8s-dark-menu-popper .el-menu-item:hover {
  background: rgba(255, 255, 255, 0.1) !important;
  color: #ffffff !important;
}

.k8s-dark-menu-popper .el-menu-item.is-active {
  background: linear-gradient(90deg, #3b82f6 0%, #2563eb 100%) !important;
  color: #ffffff !important;
}
</style>
