<template>
  <div class="cluster-overview-page">
    <!-- 顶部平台横幅 -->
    <section class="overview-hero">
      <!-- 装饰流光波纹 -->
      <svg class="hero-waves" viewBox="0 0 1200 180" preserveAspectRatio="none" aria-hidden="true">
        <path
          d="M0,135 C220,90 430,165 720,95 C940,42 1080,68 1200,20 L1200,180 L0,180 Z"
          fill="rgba(255, 255, 255, 0.34)"
        />
        <path
          d="M460,180 C660,115 840,38 1060,0 L1200,0 L1200,180 Z"
          fill="rgba(191, 219, 254, 0.3)"
        />
      </svg>

      <div class="hero-left">
        <div class="hero-k8s-logo">
          <K8sIcon name="k8s" :size="58" />
        </div>
        <div class="hero-copy">
          <h1 class="hero-title">Kubernetes 集群管理平台</h1>
          <p class="hero-desc">统一管理 · 高效运维 · 稳定可靠</p>
        </div>
      </div>

      <!-- 右侧等距集群装饰插画 -->
      <svg class="hero-illustration" viewBox="0 0 360 150" aria-hidden="true">
        <defs>
          <linearGradient id="isoTop" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#ffffff" stop-opacity="0.95" />
            <stop offset="100%" stop-color="#dbeafe" stop-opacity="0.92" />
          </linearGradient>
          <linearGradient id="isoLeft" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stop-color="#93c5fd" stop-opacity="0.88" />
            <stop offset="100%" stop-color="#3b82f6" stop-opacity="0.85" />
          </linearGradient>
          <linearGradient id="isoRight" x1="0%" y1="0%" x2="0%" y2="100%">
            <stop offset="0%" stop-color="#60a5fa" stop-opacity="0.92" />
            <stop offset="100%" stop-color="#2563eb" stop-opacity="0.92" />
          </linearGradient>
          <linearGradient id="isoPlatform" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stop-color="#bfdbfe" stop-opacity="0.6" />
            <stop offset="100%" stop-color="#93c5fd" stop-opacity="0.2" />
          </linearGradient>
        </defs>

        <!-- 底座平台网格 -->
        <polygon points="180,142 325,94 180,48 35,94" fill="url(#isoPlatform)" />
        <polyline points="82,110 180,76 278,110" fill="none" stroke="rgba(59,130,246,0.28)" stroke-width="1.4" />
        <polyline points="118,124 180,102 242,124" fill="none" stroke="rgba(59,130,246,0.22)" stroke-width="1.2" />

        <!-- 左侧节点塔 -->
        <g transform="translate(86, 56)">
          <polygon points="28,0 56,14 28,28 0,14" fill="url(#isoTop)" />
          <polygon points="0,14 28,28 28,60 0,46" fill="url(#isoLeft)" />
          <polygon points="28,28 56,14 56,46 28,60" fill="url(#isoRight)" />
          <line x1="6" y1="26" x2="22" y2="34" stroke="#ffffff" stroke-width="2" stroke-linecap="round" opacity="0.8" />
          <line x1="6" y1="35" x2="22" y2="43" stroke="#ffffff" stroke-width="2" stroke-linecap="round" opacity="0.6" />
        </g>

        <!-- 右侧节点塔 -->
        <g transform="translate(218, 56)">
          <polygon points="28,0 56,14 28,28 0,14" fill="url(#isoTop)" />
          <polygon points="0,14 28,28 28,60 0,46" fill="url(#isoLeft)" />
          <polygon points="28,28 56,14 56,46 28,60" fill="url(#isoRight)" />
          <line x1="6" y1="26" x2="22" y2="34" stroke="#ffffff" stroke-width="2" stroke-linecap="round" opacity="0.8" />
          <line x1="6" y1="35" x2="22" y2="43" stroke="#ffffff" stroke-width="2" stroke-linecap="round" opacity="0.6" />
        </g>

        <!-- 中央核心集群立方 -->
        <g transform="translate(144, 62)">
          <polygon points="36,0 72,18 36,36 0,18" fill="url(#isoTop)" />
          <polygon points="0,18 36,36 36,74 0,56" fill="url(#isoLeft)" />
          <polygon points="36,36 72,18 72,56 36,74" fill="url(#isoRight)" />
          <line x1="9" y1="34" x2="27" y2="43" stroke="#ffffff" stroke-width="2.2" stroke-linecap="round" opacity="0.9" />
          <line x1="9" y1="44" x2="27" y2="53" stroke="#ffffff" stroke-width="2.2" stroke-linecap="round" opacity="0.7" />
        </g>

        <!-- 漂浮的蓝色 K8s 六边形核心徽标 -->
        <g transform="translate(154, 8)">
          <path
            d="M23.2 2.4a5.6 5.6 0 0 1 5.6 0l14.6 8.4a5.6 5.6 0 0 1 2.8 4.8v16.8a5.6 5.6 0 0 1-2.8 4.8l-14.6 8.4a5.6 5.6 0 0 1-5.6 0L8.6 37.2a5.6 5.6 0 0 1-2.8-4.8V15.6a5.6 5.6 0 0 1 2.8-4.8l14.6-8.4z"
            fill="#2563eb"
            opacity="0.92"
          />
          <g fill="none" stroke="#ffffff" stroke-width="2.1" stroke-linecap="round">
            <circle cx="26" cy="24" r="8.5" />
            <circle cx="26" cy="24" r="2.6" fill="#ffffff" />
            <line x1="26" y1="11.5" x2="26" y2="21.4" />
            <line x1="36" y1="16.2" x2="28.1" y2="22.4" />
            <line x1="38.2" y1="26.8" x2="28.6" y2="24.6" />
            <line x1="31.5" y1="35.2" x2="27.2" y2="26.5" />
            <line x1="20.5" y1="35.2" x2="24.8" y2="26.5" />
            <line x1="13.8" y1="26.8" x2="23.4" y2="24.6" />
            <line x1="16" y1="16.2" x2="23.9" y2="22.4" />
          </g>
        </g>
      </svg>
    </section>

    <!-- 集群信息卡片（无查看详情按钮） -->
    <section class="panel-card info-panel">
      <div class="panel-header">
        <div class="panel-title">
          <span class="title-accent" />
          <span>集群信息</span>
        </div>
      </div>

      <div class="info-strip">
        <div class="info-cell">
          <div class="info-icon-circle is-blue">
            <K8sIcon name="k8s" :size="22" />
          </div>
          <div class="info-text">
            <span class="info-label">集群名称</span>
            <span class="info-value" :title="displayClusterName">{{ displayClusterName }}</span>
          </div>
        </div>

        <div class="info-cell">
          <div class="info-icon-circle is-sky">
            <svg viewBox="0 0 24 24" class="mini-svg" aria-hidden="true">
              <path
                d="M7 18h10.5a4 4 0 0 0 .8-7.92A5.5 5.5 0 0 0 7.6 8.3 4.2 4.2 0 0 0 7 18z"
                fill="none"
                stroke="currentColor"
                stroke-width="1.9"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </div>
          <div class="info-text">
            <span class="info-label">API Server</span>
            <span class="info-value is-mono" :title="apiServerText">{{ apiServerText }}</span>
          </div>
        </div>

        <div class="info-cell">
          <div class="info-icon-circle" :class="statusCircleClass">
            <svg viewBox="0 0 24 24" class="mini-svg" aria-hidden="true">
              <path
                d="M12 3l7 3v5.5c0 4.5-3 8.1-7 9.5-4-1.4-7-5-7-9.5V6l7-3z"
                fill="currentColor"
                fill-opacity="0.16"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linejoin="round"
              />
              <path
                d="M9 12.2l2 2 4-4.2"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </div>
          <div class="info-text">
            <span class="info-label">状态</span>
            <span class="info-status" :class="statusTextClass">
              <span class="status-dot" />
              <span>{{ clusterStatusText(currentStatus) }}</span>
            </span>
          </div>
        </div>

        <div class="info-cell">
          <div class="info-icon-circle is-indigo">
            <svg viewBox="0 0 24 24" class="mini-svg" aria-hidden="true">
              <path
                d="M20.6 13.4l-7.2 7.2a2 2 0 0 1-2.8 0L3 13V4a1 1 0 0 1 1-1h9l7.6 7.6a2 2 0 0 1 0 2.8z"
                fill="none"
                stroke="currentColor"
                stroke-width="1.8"
                stroke-linejoin="round"
              />
              <circle cx="8" cy="8" r="1.4" fill="currentColor" />
            </svg>
          </div>
          <div class="info-text">
            <span class="info-label">版本</span>
            <span class="info-value is-mono">{{ detail?.version || '-' }}</span>
          </div>
        </div>
      </div>
    </section>

    <!-- 4 个核心统计卡片 -->
    <section class="stats-grid">
      <!-- 节点 -->
      <article class="stat-card" @click="goTo('nodes')">
        <div class="stat-top-row">
          <div class="stat-badge is-node">
            <K8sIcon name="node" :size="36" />
          </div>
          <div class="stat-numbers">
            <span class="stat-value">{{ stats.nodeCount }}</span>
            <span class="stat-label">节点</span>
          </div>
        </div>
        <div class="stat-foot">
          <span class="foot-dot" :class="nodeHealthClass" />
          <span>{{ nodeHealthText }}</span>
        </div>
        <svg class="stat-bg-art is-blue-art" viewBox="0 0 96 76" aria-hidden="true">
          <rect x="26" y="14" width="26" height="54" rx="4" fill="currentColor" opacity="0.22" />
          <rect x="56" y="24" width="26" height="44" rx="4" fill="currentColor" opacity="0.34" />
          <rect x="6" y="34" width="22" height="34" rx="4" fill="currentColor" opacity="0.18" />
        </svg>
      </article>

      <!-- Pod -->
      <article class="stat-card" @click="goTo('pod')">
        <div class="stat-top-row">
          <div class="stat-badge is-pod">
            <K8sIcon name="pod" :size="36" />
          </div>
          <div class="stat-numbers">
            <span class="stat-value">{{ stats.podCount }}</span>
            <span class="stat-label">Pod</span>
          </div>
        </div>
        <div class="stat-foot">
          <span class="foot-dot is-ok" />
          <span>{{ podHealthText }}</span>
        </div>
        <svg class="stat-bg-art is-green-art" viewBox="0 0 96 76" aria-hidden="true">
          <polygon points="48,10 72,22 72,46 48,58 24,46 24,22" fill="currentColor" opacity="0.24" />
          <polygon points="66,34 86,44 86,64 66,74 46,64 46,44" fill="currentColor" opacity="0.34" />
        </svg>
      </article>

      <!-- 命名空间 -->
      <article class="stat-card" @click="goTo('namespace')">
        <div class="stat-top-row">
          <div class="stat-badge is-ns">
            <K8sIcon name="ns" :size="36" />
          </div>
          <div class="stat-numbers">
            <span class="stat-value">{{ stats.namespaceCount }}</span>
            <span class="stat-label">命名空间</span>
          </div>
        </div>
        <div class="stat-foot is-muted">
          <span>逻辑隔离空间</span>
        </div>
        <svg class="stat-bg-art is-amber-art" viewBox="0 0 96 76" aria-hidden="true">
          <path d="M16 24a6 6 0 0 1 6-6h20l8 8h30a6 6 0 0 1 6 6v30a6 6 0 0 1-6 6H22a6 6 0 0 1-6-6V24z" fill="currentColor" opacity="0.26" />
        </svg>
      </article>

      <!-- Deployment -->
      <article class="stat-card" @click="goTo('workload/deployment')">
        <div class="stat-top-row">
          <div class="stat-badge is-deploy">
            <K8sIcon name="deploy" :size="36" />
          </div>
          <div class="stat-numbers">
            <span class="stat-value">{{ stats.deploymentCount }}</span>
            <span class="stat-label">Deployment</span>
          </div>
        </div>
        <div class="stat-foot is-muted">
          <span>无状态工作负载</span>
        </div>
        <svg class="stat-bg-art is-purple-art" viewBox="0 0 96 76" aria-hidden="true">
          <path
            d="M72 12c-14 2-26 10-34 22l-12 2-6 10 11 4 8 8 4 11 10-6 2-12c12-8 20-20 22-34l-5-5z"
            fill="currentColor"
            opacity="0.26"
          />
          <circle cx="56" cy="30" r="5" fill="#ffffff" opacity="0.7" />
        </svg>
      </article>
    </section>

    <!-- 底部三栏：资源使用情况 / 快捷操作 / 集群事件 -->
    <section class="bottom-grid">
      <!-- 1. 资源使用情况（对接后端 /node/usage 实时数据） -->
      <div class="panel-card usage-panel">
        <div class="panel-header">
          <div class="panel-title">
            <span class="title-accent" />
            <span>资源使用情况</span>
          </div>
          <div class="usage-header-right">
            <span class="live-tag">
              <span class="live-dot" />
              <span>{{ usageSourceLabel }}</span>
            </span>
            <button
              type="button"
              class="refresh-icon-btn"
              :disabled="usageLoading"
              title="刷新资源使用情况"
              @click="loadUsage"
            >
              <el-icon :class="{ 'is-spinning': usageLoading }"><Refresh /></el-icon>
            </button>
          </div>
        </div>

        <div v-loading="usageLoading" element-loading-background="rgba(255, 255, 255, 0.7)" class="rings-grid">
          <!-- CPU 使用率 -->
          <div class="ring-item">
            <div class="ring-box">
              <svg viewBox="0 0 120 120" class="ring-svg">
                <defs>
                  <linearGradient id="ringCpuGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stop-color="#60a5fa" />
                    <stop offset="100%" stop-color="#2563eb" />
                  </linearGradient>
                </defs>
                <circle cx="60" cy="60" :r="RING_RADIUS" class="ring-track" />
                <circle
                  cx="60"
                  cy="60"
                  :r="RING_RADIUS"
                  class="ring-progress"
                  stroke="url(#ringCpuGrad)"
                  :stroke-dasharray="RING_CIRCUMFERENCE"
                  :stroke-dashoffset="getRingOffset(usage.cpu.percent)"
                />
              </svg>
              <span class="ring-percent">{{ formatPercent(usage.cpu.percent) }}%</span>
            </div>
            <div class="ring-title">CPU 使用率</div>
            <div class="ring-sub">{{ usage.cpu.usedText }} / {{ usage.cpu.totalText }}</div>
          </div>

          <!-- 内存使用率 -->
          <div class="ring-item">
            <div class="ring-box">
              <svg viewBox="0 0 120 120" class="ring-svg">
                <defs>
                  <linearGradient id="ringMemGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stop-color="#4ade80" />
                    <stop offset="100%" stop-color="#10b981" />
                  </linearGradient>
                </defs>
                <circle cx="60" cy="60" :r="RING_RADIUS" class="ring-track" />
                <circle
                  cx="60"
                  cy="60"
                  :r="RING_RADIUS"
                  class="ring-progress"
                  stroke="url(#ringMemGrad)"
                  :stroke-dasharray="RING_CIRCUMFERENCE"
                  :stroke-dashoffset="getRingOffset(usage.memory.percent)"
                />
              </svg>
              <span class="ring-percent">{{ formatPercent(usage.memory.percent) }}%</span>
            </div>
            <div class="ring-title">内存使用率</div>
            <div class="ring-sub">{{ usage.memory.usedText }} / {{ usage.memory.totalText }}</div>
          </div>

          <!-- 存储使用率 -->
          <div class="ring-item">
            <div class="ring-box">
              <svg viewBox="0 0 120 120" class="ring-svg">
                <defs>
                  <linearGradient id="ringStorageGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stop-color="#a78bfa" />
                    <stop offset="100%" stop-color="#7c3aed" />
                  </linearGradient>
                </defs>
                <circle cx="60" cy="60" :r="RING_RADIUS" class="ring-track" />
                <circle
                  cx="60"
                  cy="60"
                  :r="RING_RADIUS"
                  class="ring-progress"
                  stroke="url(#ringStorageGrad)"
                  :stroke-dasharray="RING_CIRCUMFERENCE"
                  :stroke-dashoffset="getRingOffset(usage.storage.percent)"
                />
              </svg>
              <span class="ring-percent">{{ formatPercent(usage.storage.percent) }}%</span>
            </div>
            <div class="ring-title">存储使用率</div>
            <div class="ring-sub">{{ usage.storage.usedText }} / {{ usage.storage.totalText }}</div>
          </div>
        </div>
      </div>

      <!-- 2. 快捷操作（6 格，全部使用官方 Kubernetes 图标） -->
      <div class="panel-card actions-panel">
        <div class="panel-header">
          <div class="panel-title">
            <span class="title-accent" />
            <span>快捷操作</span>
          </div>
        </div>

        <div class="quick-grid">
          <button type="button" class="quick-tile" @click="goTo('namespace')">
            <K8sIcon name="ns" :size="30" />
            <span>命名空间</span>
          </button>
          <button type="button" class="quick-tile" @click="goTo('workload/deployment')">
            <K8sIcon name="deploy" :size="30" />
            <span>工作负载</span>
          </button>
          <button type="button" class="quick-tile" @click="goTo('service/service')">
            <K8sIcon name="svc" :size="30" />
            <span>服务</span>
          </button>
          <button type="button" class="quick-tile" @click="goTo('config/configmap')">
            <K8sIcon name="cm" :size="30" />
            <span>配置</span>
          </button>
          <button type="button" class="quick-tile" @click="goTo('storage/pvc')">
            <K8sIcon name="pvc" :size="30" />
            <span>存储卷</span>
          </button>
          <button type="button" class="quick-tile" @click="goTo('resource/tools')">
            <K8sIcon name="etcd" :size="30" />
            <span>工具</span>
          </button>
        </div>
      </div>

      <!-- 3. 集群事件: 最近的 Warning 事件 -->
      <div class="panel-card events-panel">
        <div class="panel-header">
          <div class="panel-title">
            <span class="title-accent" />
            <span>最近告警事件</span>
          </div>
          <el-link type="primary" :underline="false" @click="goTo('resource/events')">查看全部</el-link>
        </div>

        <div v-loading="eventsLoading" class="events-list">
          <div v-for="(ev, idx) in recentEvents" :key="idx" class="event-row" :title="`${ev.object}：${ev.message}`">
            <span class="event-badge">
              <span class="event-dot is-amber" />
              <span class="event-type">{{ ev.reason }}</span>
            </span>
            <span class="event-time">{{ timeAgo(ev.lastTime) }}</span>
            <span class="event-msg">{{ ev.object }}：{{ ev.message }}</span>
          </div>
          <el-empty v-if="!eventsLoading && !recentEvents.length" description="最近没有 Warning 事件" :image-size="60" />
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { clusterStatusText } from '@/utils/status'
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Refresh } from '@element-plus/icons-vue'
import { useClusterStore } from '@/stores/cluster'
import {
  clusterApi,
  type ClusterVO,
  type ClusterResourceUsageVO
} from '@/api/cluster'
import K8sIcon from '@/components/K8sIcon.vue'
import { eventApi } from '@/api/event'
import type { EventVO } from '@/api/job'
import { timeAgo } from '@/utils/navigation'

const router = useRouter()
const clusterStore = useClusterStore()

const RING_RADIUS = 48
const RING_CIRCUMFERENCE = 2 * Math.PI * RING_RADIUS

const detail = ref<ClusterVO | null>(null)
const usageLoading = ref(false)

const stats = ref({
  nodeCount: 0,
  podCount: 0,
  namespaceCount: 0,
  deploymentCount: 0
})

const createEmptyUsage = (): ClusterResourceUsageVO => ({
  cpu: { used: 0, total: 0, percent: 0, usedText: '0', totalText: '0 核' },
  memory: { used: 0, total: 0, percent: 0, usedText: '0', totalText: '0 Gi' },
  storage: { used: 0, total: 0, percent: 0, usedText: '0', totalText: '0 Gi' },
  nodeTotal: 0,
  nodeReady: 0,
  podTotal: 0,
  podRunning: 0,
  source: 'metrics-server'
})

const usage = ref<ClusterResourceUsageVO>(createEmptyUsage())

// 最近的 Warning 事件
const recentEvents = ref<EventVO[]>([])
const eventsLoading = ref(false)
const loadEvents = async () => {
  if (!clusterStore.currentClusterId) return
  eventsLoading.value = true
  try {
    const resp = await eventApi.list({ clusterId: clusterStore.currentClusterId, type: 'Warning', page: 1, pageSize: 6 })
    recentEvents.value = resp.data?.items || []
  } catch {
    recentEvents.value = []
  } finally {
    eventsLoading.value = false
  }
}

const currentStatus = computed(
  () => detail.value?.status || clusterStore.currentCluster?.status || 'Active'
)

const displayClusterName = computed(
  () => detail.value?.alias || clusterStore.currentClusterName || '-'
)

const apiServerText = computed(
  () => detail.value?.apiServer || clusterStore.currentCluster?.apiServer || '-'
)

const statusCircleClass = computed(() => {
  switch (currentStatus.value) {
    case 'Active':
    case 'Running':
      return 'is-green'
    case 'Error':
    case 'Failed':
      return 'is-red'
    case 'Pending':
      return 'is-amber'
    default:
      return 'is-gray'
  }
})

const statusTextClass = computed(() => statusCircleClass.value)

const nodeHealthClass = computed(() => {
  const total = usage.value.nodeTotal || stats.value.nodeCount
  const ready = usage.value.nodeReady
  if (total === 0) return 'is-muted-dot'
  return ready >= total ? 'is-ok' : 'is-warn'
})

const nodeHealthText = computed(() => {
  const total = usage.value.nodeTotal || stats.value.nodeCount
  const ready = usage.value.nodeReady
  if (total === 0) return '暂无节点'
  if (usage.value.nodeTotal > 0 && ready === total) return '全部正常'
  if (usage.value.nodeTotal > 0) return `就绪 ${ready} / 总计 ${total}`
  return `共 ${total} 个节点`
})

const podHealthText = computed(() => {
  const total = usage.value.podTotal || stats.value.podCount
  if (usage.value.podTotal > 0) {
    return `运行中 ${usage.value.podRunning} / 总计 ${total}`
  }
  return `总计 ${total}`
})

const usageSourceLabel = computed(() =>
  usage.value.source === 'requests' ? '配额估算' : '实时数据'
)

const formatPercent = (pct: number) => {
  const clamped = Math.min(100, Math.max(0, Number(pct) || 0))
  return Math.round(clamped)
}

const getRingOffset = (pct: number) => {
  const clamped = Math.min(100, Math.max(0, Number(pct) || 0))
  return RING_CIRCUMFERENCE * (1 - clamped / 100)
}

const goTo = (subPath: string) => {
  if (!clusterStore.currentClusterId) return
  router.push(`/cluster/${clusterStore.currentClusterId}/${subPath}`)
}

const loadStats = async () => {
  const clusterName = clusterStore.currentClusterId
  if (!clusterName) return
  try {
    const response = await clusterApi.getCluster(clusterName)
    detail.value = response.data
    stats.value = {
      nodeCount: response.data?.nodeCount || 0,
      podCount: response.data?.podCount || 0,
      namespaceCount: response.data?.namespaceCount || 0,
      deploymentCount: response.data?.deploymentCount || 0
    }
  } catch (error) {
    console.error('Failed to load cluster stats:', error)
  }
}

const loadUsage = async () => {
  const clusterName = clusterStore.currentClusterId
  if (!clusterName) return
  usageLoading.value = true
  try {
    const response = await clusterApi.getClusterUsage(clusterName)
    if (response && response.data) {
      usage.value = response.data
    }
  } catch (error) {
    console.error('Failed to load cluster usage:', error)
    usage.value = createEmptyUsage()
  } finally {
    usageLoading.value = false
  }
}

const loadAll = () => {
  loadStats()
  loadUsage()
  loadEvents()
}

watch(
  () => clusterStore.currentClusterId,
  (newId, oldId) => {
    if (newId && newId !== oldId) {
      loadAll()
    }
  }
)

onMounted(() => {
  loadAll()
})
</script>

<style scoped>
.cluster-overview-page {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

/* ==================== 顶部平台横幅 ==================== */
.overview-hero {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 124px;
  padding: 22px 34px;
  border-radius: 18px;
  overflow: hidden;
  background: linear-gradient(120deg, #eff6ff 0%, #dbeafe 54%, #e8f1ff 100%);
  border: 1px solid rgba(255, 255, 255, 0.92);
  box-shadow: 0 10px 28px rgba(37, 99, 235, 0.08);
}

.hero-waves {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.hero-left {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 20px;
}

.hero-k8s-logo {
  flex-shrink: 0;
  filter: drop-shadow(0 8px 16px rgba(37, 99, 235, 0.28));
}

.hero-title {
  margin: 0 0 6px;
  font-size: 26px;
  font-weight: 700;
  color: #0f172a;
  letter-spacing: 0.3px;
}

.hero-desc {
  margin: 0;
  font-size: 14px;
  color: #475569;
  letter-spacing: 0.4px;
}

.hero-illustration {
  position: relative;
  z-index: 2;
  width: 310px;
  height: 126px;
  flex-shrink: 0;
  margin-right: 6px;
}

/* ==================== 通用白色圆角面板 ==================== */
.panel-card {
  background: #ffffff;
  border-radius: 16px;
  padding: 20px 24px;
  border: 1px solid rgba(226, 232, 240, 0.78);
  box-shadow: 0 6px 22px rgba(15, 23, 42, 0.04);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
}

.panel-title {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 700;
  color: #0f172a;
}

.title-accent {
  width: 4px;
  height: 16px;
  border-radius: 999px;
  background: linear-gradient(180deg, #3b82f6 0%, #2563eb 100%);
}

/* ==================== 集群信息条 ==================== */
.info-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.info-cell {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 4px 16px 4px 4px;
  position: relative;
  min-width: 0;
}

.info-cell:not(:last-child)::after {
  content: '';
  position: absolute;
  right: 4px;
  top: 50%;
  transform: translateY(-50%);
  width: 1px;
  height: 36px;
  background: #e9eff7;
}

.info-icon-circle {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.info-icon-circle.is-blue {
  background: #eff6ff;
  color: #2563eb;
}

.info-icon-circle.is-sky {
  background: #f0f9ff;
  color: #0284c7;
}

.info-icon-circle.is-green {
  background: #ecfdf5;
  color: #16a34a;
}

.info-icon-circle.is-red {
  background: #fef2f2;
  color: #dc2626;
}

.info-icon-circle.is-amber {
  background: #fffbeb;
  color: #d97706;
}

.info-icon-circle.is-gray {
  background: #f1f5f9;
  color: #64748b;
}

.info-icon-circle.is-indigo {
  background: #eef2ff;
  color: #4f46e5;
}

.mini-svg {
  width: 22px;
  height: 22px;
}

.info-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.info-label {
  font-size: 12px;
  color: #94a3b8;
}

.info-value {
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.info-value.is-mono {
  font-size: 14px;
  font-variant-numeric: tabular-nums;
}

.info-status {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-size: 15px;
  font-weight: 700;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.info-status.is-green {
  color: #16a34a;
}
.info-status.is-green .status-dot {
  background: #22c55e;
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.55);
}

.info-status.is-red {
  color: #dc2626;
}
.info-status.is-red .status-dot {
  background: #ef4444;
}

.info-status.is-amber {
  color: #d97706;
}
.info-status.is-amber .status-dot {
  background: #f59e0b;
}

.info-status.is-gray {
  color: #64748b;
}
.info-status.is-gray .status-dot {
  background: #94a3b8;
}

/* ==================== 4 个统计卡片 ==================== */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 18px;
}

.stat-card {
  position: relative;
  overflow: hidden;
  background: #ffffff;
  border-radius: 16px;
  padding: 20px 22px 18px;
  border: 1px solid rgba(226, 232, 240, 0.8);
  box-shadow: 0 6px 22px rgba(15, 23, 42, 0.04);
  cursor: pointer;
  transition: transform 0.22s ease, box-shadow 0.22s ease, border-color 0.22s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  border-color: rgba(147, 197, 253, 0.85);
  box-shadow: 0 12px 28px rgba(37, 99, 235, 0.1);
}

.stat-top-row {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 14px;
}

.stat-badge {
  width: 54px;
  height: 54px;
  border-radius: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 1px solid rgba(50, 108, 229, 0.12);
}

.stat-badge.is-node {
  background: linear-gradient(135deg, #eff6ff 0%, #dbeafe 100%);
}

.stat-badge.is-pod {
  background: linear-gradient(135deg, #ecfdf5 0%, #d1fae5 100%);
}

.stat-badge.is-ns {
  background: linear-gradient(135deg, #fff7ed 0%, #ffedd5 100%);
}

.stat-badge.is-deploy {
  background: linear-gradient(135deg, #f5f3ff 0%, #ede9fe 100%);
}

.stat-numbers {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.1;
  font-variant-numeric: tabular-nums;
}

.stat-label {
  font-size: 13px;
  color: #64748b;
  font-weight: 500;
}

.stat-foot {
  position: relative;
  z-index: 2;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #475569;
  font-weight: 500;
}

.stat-foot.is-muted {
  color: #94a3b8;
}

.foot-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.foot-dot.is-ok {
  background: #22c55e;
}

.foot-dot.is-warn {
  background: #f59e0b;
}

.foot-dot.is-muted-dot {
  background: #cbd5e1;
}

.stat-bg-art {
  position: absolute;
  right: 8px;
  bottom: 4px;
  width: 86px;
  height: 68px;
  pointer-events: none;
  z-index: 1;
}

.is-blue-art {
  color: #60a5fa;
}

.is-green-art {
  color: #34d399;
}

.is-amber-art {
  color: #fb923c;
}

.is-purple-art {
  color: #a78bfa;
}

/* ==================== 底部三栏布局 ==================== */
.bottom-grid {
  display: grid;
  grid-template-columns: 1.48fr 0.96fr 1.06fr;
  gap: 18px;
  align-items: stretch;
}

/* 资源使用情况 */
.usage-header-right {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.live-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #475569;
  font-weight: 500;
}

.live-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #22c55e;
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.5);
}

.refresh-icon-btn {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  color: #475569;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
}

.refresh-icon-btn:hover:not(:disabled) {
  border-color: #93c5fd;
  background: #eff6ff;
  color: #2563eb;
}

.refresh-icon-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.is-spinning {
  animation: spin 0.9s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.rings-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
  padding: 12px 4px 6px;
}

.ring-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.ring-box {
  position: relative;
  width: 114px;
  height: 114px;
  margin-bottom: 14px;
}

.ring-svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.ring-track {
  fill: none;
  stroke: #edf2f7;
  stroke-width: 9;
}

.ring-progress {
  fill: none;
  stroke-width: 9;
  stroke-linecap: round;
  transition: stroke-dashoffset 0.65s cubic-bezier(0.22, 1, 0.36, 1);
}

.ring-percent {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  font-weight: 700;
  color: #0f172a;
  font-variant-numeric: tabular-nums;
}

.ring-title {
  font-size: 14px;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 4px;
}

.ring-sub {
  font-size: 12px;
  color: #64748b;
  font-variant-numeric: tabular-nums;
}

/* 快捷操作 3x2 卡片 */
.quick-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.quick-tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 18px 10px;
  border-radius: 12px;
  border: 1px solid rgba(219, 234, 254, 0.75);
  background: #f5f9ff;
  color: #1e293b;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.quick-tile:hover {
  transform: translateY(-2px);
  background: #ebf3ff;
  border-color: #93c5fd;
  color: #1d4ed8;
  box-shadow: 0 6px 16px rgba(37, 99, 235, 0.1);
}

/* 集群事件列表 */
.events-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.event-row {
  display: grid;
  grid-template-columns: 128px 64px minmax(0, 1fr);
  align-items: start;
  gap: 10px;
  font-size: 13px;
  line-height: 1.45;
  padding-bottom: 12px;
  border-bottom: 1px dashed #eef2f7;
}

.event-row:last-child {
  padding-bottom: 0;
  border-bottom: none;
}

.event-badge {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-weight: 600;
  color: #334155;
}

.event-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.event-dot.is-green {
  background: #22c55e;
}

.event-dot.is-blue {
  background: #3b82f6;
}

.event-dot.is-amber {
  background: #f59e0b;
}

.event-time {
  color: #94a3b8;
  font-size: 12px;
  white-space: nowrap;
}

.event-msg {
  display: -webkit-box;
  overflow: hidden;
  color: #1e293b;
  word-break: break-word;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.event-type {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.event-badge {
  min-width: 0;
}

/* ==================== 响应式适配 ==================== */
@media (max-width: 1280px) {
  .bottom-grid {
    grid-template-columns: 1fr 1fr;
  }

  .usage-panel {
    grid-column: 1 / -1;
  }
}

@media (max-width: 960px) {
  .hero-illustration {
    display: none;
  }

  .info-strip {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .info-cell::after {
    display: none;
  }

  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .bottom-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 560px) {
  .overview-hero {
    padding: 18px 20px;
  }

  .hero-title {
    font-size: 20px;
  }

  .info-strip {
    grid-template-columns: 1fr;
  }

  .rings-grid {
    grid-template-columns: 1fr;
    gap: 20px;
  }

  .quick-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
