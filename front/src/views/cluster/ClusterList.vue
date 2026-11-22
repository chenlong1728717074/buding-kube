<template>
  <div class="cluster-list-page">
    <!-- 共享 SVG 渐变定义 -->
    <svg aria-hidden="true" class="svg-defs" focusable="false">
      <defs>
        <linearGradient id="heroK8sBg" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#53a8ff" />
          <stop offset="55%" stop-color="#2b75f6" />
          <stop offset="100%" stop-color="#1b56e0" />
        </linearGradient>
        <linearGradient id="clusterCloudGrad" x1="12%" y1="8%" x2="88%" y2="95%">
          <stop offset="0%" stop-color="#5db4ff" />
          <stop offset="52%" stop-color="#2e7df6" />
          <stop offset="100%" stop-color="#1d5be0" />
        </linearGradient>
        <linearGradient id="clusterCloudMutedGrad" x1="12%" y1="8%" x2="88%" y2="95%">
          <stop offset="0%" stop-color="#87c5ff" />
          <stop offset="100%" stop-color="#4e91f7" />
        </linearGradient>
        <linearGradient id="k8sWatermarkGrad" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.22" />
          <stop offset="100%" stop-color="#60a5fa" stop-opacity="0.06" />
        </linearGradient>
      </defs>
    </svg>

    <!-- 顶部玻璃态横幅 -->
    <section class="cluster-hero">
      <!-- 装饰流光波纹 -->
      <svg class="hero-waves" viewBox="0 0 1440 200" preserveAspectRatio="none" aria-hidden="true">
        <path
          d="M640,200 C800,120 960,30 1180,0 L1440,0 L1440,200 Z"
          fill="rgba(255, 255, 255, 0.28)"
        />
        <path
          d="M440,200 C650,140 830,65 1060,0 L1185,0 C975,72 805,146 635,200 Z"
          fill="rgba(255, 255, 255, 0.36)"
        />
        <path
          d="M0,165 C260,120 520,185 820,105 C1060,42 1240,58 1440,15 L1440,200 L0,200 Z"
          fill="rgba(191, 219, 254, 0.22)"
        />
      </svg>

      <!-- 右侧 K8s 水印 -->
      <svg class="hero-watermark" viewBox="0 0 100 100" aria-hidden="true">
        <g fill="none" stroke="rgba(147, 197, 253, 0.42)" stroke-width="5.5" stroke-linecap="round">
          <circle cx="50" cy="50" r="28" />
          <circle cx="50" cy="50" r="9" />
          <line x1="50" y1="9" x2="50" y2="41" />
          <line x1="50" y1="59" x2="50" y2="91" />
          <line x1="14.5" y1="29.5" x2="42.2" y2="45.5" />
          <line x1="57.8" y1="54.5" x2="85.5" y2="70.5" />
          <line x1="85.5" y1="29.5" x2="57.8" y2="45.5" />
          <line x1="42.2" y1="54.5" x2="14.5" y2="70.5" />
        </g>
      </svg>

      <div class="hero-left">
        <!-- 3D 蓝宝石 K8s 六边形徽章 -->
        <div class="hero-badge" aria-hidden="true">
          <svg viewBox="0 0 80 80" class="hero-badge-svg">
            <path
              d="M35.5 6.8a9 9 0 0 1 9 0l22.7 13.1a9 9 0 0 1 4.5 7.8v26.2a9 9 0 0 1-4.5 7.8L44.5 74.8a9 9 0 0 1-9 0L12.8 61.7a9 9 0 0 1-4.5-7.8V27.7a9 9 0 0 1 4.5-7.8L35.5 6.8z"
              fill="url(#heroK8sBg)"
            />
            <path
              d="M36.3 8.2a7.4 7.4 0 0 1 7.4 0l22 12.7a7.4 7.4 0 0 1 3.7 6.4v25.4a7.4 7.4 0 0 1-3.7 6.4l-22 12.7a7.4 7.4 0 0 1-7.4 0l-22-12.7a7.4 7.4 0 0 1-3.7-6.4V27.3a7.4 7.4 0 0 1 3.7-6.4l22-12.7z"
              fill="none"
              stroke="rgba(255,255,255,0.38)"
              stroke-width="1.5"
            />
            <g fill="none" stroke="#ffffff" stroke-width="3.2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="40" cy="40" r="14.2" />
              <circle cx="40" cy="40" r="4.6" fill="#ffffff" />
              <line x1="40" y1="18.5" x2="40" y2="35.4" />
              <line x1="56.8" y1="26.6" x2="43.6" y2="37.1" />
              <line x1="60.9" y1="44.8" x2="44.5" y2="41" />
              <line x1="49.3" y1="59.4" x2="42" y2="44.2" />
              <line x1="30.7" y1="59.4" x2="38" y2="44.2" />
              <line x1="19.1" y1="44.8" x2="35.5" y2="41" />
              <line x1="23.2" y1="26.6" x2="36.4" y2="37.1" />
            </g>
          </svg>
        </div>

        <div class="hero-copy">
          <h1 class="hero-title">集群管理</h1>
          <p class="hero-desc">管理和监控您的 Kubernetes 集群，简化运维操作</p>
        </div>
      </div>

      <div class="hero-actions">
        <button type="button" class="add-cluster-btn" @click="showAddDialog = true">
          <el-icon><Plus /></el-icon>
          <span>添加集群</span>
        </button>
      </div>
    </section>

    <!-- 搜索和状态筛选栏 -->
    <section class="filter-bar">
      <div class="filter-left">
        <el-input
          v-model="searchParams.keyword"
          placeholder="搜索集群名称"
          clearable
          class="keyword-input"
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>

        <div class="status-filter">
          <span class="filter-label">活跃状态:</span>
          <el-select
            v-model="searchParams.status"
            placeholder="全部状态"
            clearable
            class="status-select"
            @change="handleSearch"
          >
            <el-option label="全部状态" value="" />
            <el-option label="运行中" value="Active" />
            <el-option label="异常" value="Error" />
            <el-option label="未知" value="Unknown" />
          </el-select>
        </div>

        <button type="button" class="pill-btn pill-btn-primary" @click="handleSearch">
          <el-icon><Search /></el-icon>
          <span>搜索</span>
        </button>

        <button type="button" class="pill-btn pill-btn-ghost" @click="handleReset">
          <el-icon><Refresh /></el-icon>
          <span>重置</span>
        </button>
      </div>

      <div class="filter-summary">
        <span class="total-badge">
          <span class="total-dot" />
          <span>共 <strong>{{ total }}</strong> 个集群</span>
        </span>
      </div>
    </section>

    <!-- 全屏手势滑动 3D 焦点集群展示区 -->
    <section
      v-loading="loading"
      element-loading-background="rgba(241, 246, 255, 0.65)"
      class="cluster-stage"
    >
      <div v-if="clusters.length === 0 && !loading" class="empty-card">
        <el-empty description="暂无集群数据" :image-size="110">
          <button type="button" class="add-cluster-btn" @click="showAddDialog = true">
            <el-icon><Plus /></el-icon>
            <span>添加集群</span>
          </button>
        </el-empty>
      </div>

      <template v-else-if="clusters.length > 0">
        <div
          ref="viewportRef"
          class="carousel-viewport"
          :class="{ 'is-dragging': isDragging }"
          @pointerdown="handlePointerDown"
          @pointermove="handlePointerMove"
          @pointerup="handlePointerUp"
          @pointercancel="handlePointerUp"
          @wheel="handleWheel"
        >
          <!-- 滑动卡片轨道 -->
          <div class="carousel-track">
            <article
              v-for="(cluster, idx) in clusters"
              :key="cluster.id"
              class="cluster-card"
              :class="{
                'is-active': idx === visualActiveIndex,
                'is-side': idx !== visualActiveIndex,
                'is-hovered': hoveredIndex === idx && idx !== visualActiveIndex && !isDragging
              }"
              :style="getCardStyle(idx)"
              @mouseenter="hoveredIndex = idx"
              @mouseleave="handleCardLeave(idx)"
              @click="handleCardClick(idx, cluster)"
            >
              <!-- 卡片内部柔和 K8s 水印 -->
              <svg class="card-watermark" viewBox="0 0 120 120" aria-hidden="true">
                <g
                  fill="none"
                  stroke="url(#k8sWatermarkGrad)"
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

              <!-- 卡片头部: 蓝云立方图标 + 名称 + 状态胶囊 + 更多菜单 -->
              <div class="card-header">
                <div class="card-identity">
                  <div class="cloud-emblem" aria-hidden="true">
                    <svg viewBox="0 0 64 52" class="cloud-svg">
                      <path
                        d="M17.5 44h29.5c7.2 0 13-5.4 13-12.2 0-6.2-4.8-11.3-11-12.1C47.5 11.4 40.3 5 31.4 5c-8.3 0-15.3 5.7-17.2 13.4C7.9 19.3 3 24.7 3 31.2 3 38.3 9.5 44 17.5 44z"
                        :fill="idx === visualActiveIndex ? 'url(#clusterCloudGrad)' : 'url(#clusterCloudMutedGrad)'"
                      />
                      <!-- 白色 3D 立方体 -->
                      <g
                        fill="none"
                        stroke="#ffffff"
                        stroke-width="2.1"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      >
                        <polygon points="32,16.5 41.5,22 41.5,33 32,38.5 22.5,33 22.5,22" />
                        <polyline points="22.5,22 32,27.5 41.5,22" />
                        <line x1="32" y1="27.5" x2="32" y2="38.5" />
                      </g>
                    </svg>
                  </div>

                  <div class="title-wrap">
                    <div class="title-row">
                      <h3
                        class="cluster-name"
                        :title="cluster.alias ? `${cluster.alias} (${cluster.name})` : cluster.name"
                      >
                        {{ cluster.alias || cluster.name }}
                      </h3>
                      <span class="status-pill" :class="getStatusClass(cluster.status)">
                        <span class="status-dot" />
                        <span>{{ getStatusText(cluster.status) }}</span>
                      </span>
                    </div>
                    <p v-if="cluster.describe" class="cluster-describe" :title="cluster.describe">
                      {{ cluster.describe }}
                    </p>
                  </div>
                </div>

                <el-dropdown
                  trigger="click"
                  @command="(cmd: string) => handleCommand(cmd, cluster)"
                  @click.stop
                >
                  <button
                    type="button"
                    class="more-trigger"
                    aria-label="更多操作"
                    @pointerdown.stop
                    @click.stop
                  >
                    <el-icon><MoreFilled /></el-icon>
                  </button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="edit">
                        <el-icon><Edit /></el-icon>
                        编辑
                      </el-dropdown-item>
                      <el-dropdown-item command="delete" divided>
                        <el-icon><Delete /></el-icon>
                        删除
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>

              <!-- 卡片中部: API Server & 版本 -->
              <div class="card-meta">
                <div class="meta-row">
                  <span class="meta-label">API Server</span>
                  <span class="meta-value" :title="cluster.apiServer || '-'">
                    {{ cluster.apiServer || '-' }}
                  </span>
                </div>
                <div class="meta-row">
                  <span class="meta-label">版本</span>
                  <span class="meta-value" :title="cluster.version || '-'">
                    {{ cluster.version || '-' }}
                  </span>
                </div>
              </div>

              <!-- 卡片底部: 统计指标 + 进入集群按钮 -->
              <div class="card-footer">
                <div class="card-stats">
                  <div class="stat-item">
                    <div class="stat-top">
                      <svg class="stat-icon" viewBox="0 0 20 20" aria-hidden="true">
                        <rect x="2.5" y="3" width="15" height="4" rx="1.2" fill="currentColor" />
                        <rect x="2.5" y="8.2" width="15" height="4" rx="1.2" fill="currentColor" />
                        <rect x="2.5" y="13.4" width="15" height="4" rx="1.2" fill="currentColor" />
                        <circle cx="5.2" cy="5" r="0.85" fill="#ffffff" />
                        <circle cx="5.2" cy="10.2" r="0.85" fill="#ffffff" />
                        <circle cx="5.2" cy="15.4" r="0.85" fill="#ffffff" />
                      </svg>
                      <span class="stat-num">{{ cluster.nodeCount || 0 }}</span>
                    </div>
                    <span class="stat-label">节点</span>
                  </div>

                  <div class="stat-item">
                    <div class="stat-top">
                      <svg class="stat-icon" viewBox="0 0 20 20" aria-hidden="true">
                        <path
                          d="M10 2.2L17 6.2V13.8L10 17.8L3 13.8V6.2L10 2.2Z"
                          fill="currentColor"
                        />
                        <path
                          d="M3.6 6.6L10 10.2L16.4 6.6M10 10.2V17.2"
                          fill="none"
                          stroke="#ffffff"
                          stroke-width="1.4"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                        />
                      </svg>
                      <span class="stat-num">{{ cluster.podCount || 0 }}</span>
                    </div>
                    <span class="stat-label">Pod</span>
                  </div>

                  <div class="stat-item">
                    <div class="stat-top">
                      <svg class="stat-icon" viewBox="0 0 20 20" aria-hidden="true">
                        <path
                          d="M10 2.3L16.5 4.8V9.6C16.5 13.7 13.7 16.9 10 18.2C6.3 16.9 3.5 13.7 3.5 9.6V4.8L10 2.3Z"
                          fill="currentColor"
                        />
                        <path
                          d="M7.2 10.2L9.1 12.1L13.1 8.1"
                          fill="none"
                          stroke="#ffffff"
                          stroke-width="1.8"
                          stroke-linecap="round"
                          stroke-linejoin="round"
                        />
                      </svg>
                      <span class="stat-num">{{ cluster.namespaceCount || 0 }}</span>
                    </div>
                    <span class="stat-label">命名空间</span>
                  </div>
                </div>

                <button
                  type="button"
                  class="enter-btn"
                  :class="{ 'is-primary': idx === visualActiveIndex }"
                  @pointerdown.stop
                  @click.stop="enterCluster(cluster)"
                >
                  <span>进入集群</span>
                  <el-icon><ArrowRight /></el-icon>
                </button>
              </div>
            </article>
          </div>
        </div>

        <!-- 底部指示条与动态加载状态（无分页组件） -->
        <div class="stage-status">
          <div v-if="clusters.length > 1" class="carousel-dots" role="tablist" aria-label="集群卡片切换">
            <button
              v-for="dotIdx in visibleDotIndices"
              :key="clusters[dotIdx]?.id ?? dotIdx"
              type="button"
              class="carousel-dot"
              :class="{ 'is-active': dotIdx === activeIndex }"
              :aria-label="`切换到 ${clusters[dotIdx]?.alias || clusters[dotIdx]?.name}`"
              :aria-selected="dotIdx === activeIndex"
              @click="goToIndex(dotIdx)"
            />
          </div>

          <div class="stage-meta">
            <span class="stage-counter">
              第 <strong>{{ activeIndex + 1 }}</strong> / {{ Math.max(total, clusters.length) }} 个集群
            </span>
            <span v-if="loadingMore" class="stage-loading">
              <el-icon class="is-loading"><Loading /></el-icon>
              <span>正在加载更多集群...</span>
            </span>
            <span v-else-if="clusters.length > 1" class="stage-tip">
              左右拖动或滑动切换{{ hasMore ? ' · 滑至末尾自动加载' : '' }}
            </span>
          </div>
        </div>
      </template>
    </section>

    <!-- 添加集群对话框 -->
    <el-dialog
      v-model="showAddDialog"
      title="添加集群"
      width="960px" top="6vh"
      :close-on-click-modal="false" class="form-dialog form-dialog--wide"
    >
      <el-form :model="clusterForm" :rules="addRules" ref="formRef" label-width="100px" label-position="top">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="集群名称" prop="name">
              <el-input v-model="clusterForm.name" placeholder="请输入集群名称（唯一标识）" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="集群别名">
              <el-input v-model="clusterForm.alias" placeholder="请输入集群别名（显示名称）" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="集群描述">
          <el-input v-model="clusterForm.describe" type="textarea" :rows="2" placeholder="请输入集群描述" />
        </el-form-item>
        <ClusterConnectionForm
          v-model:auth-type="clusterForm.authType"
          v-model:config="clusterForm.config"
          v-model:uri="clusterForm.uri"
          v-model:token="clusterForm.token"
        />
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAddCluster" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑集群对话框 -->
    <el-dialog
      v-model="showEditDialog"
      title="编辑集群"
      width="960px" top="6vh"
      :close-on-click-modal="false" class="form-dialog form-dialog--wide"
    >
      <el-form :model="editClusterForm" :rules="editRules" ref="editFormRef" label-width="100px" label-position="top">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="集群名称" prop="name">
              <el-input v-model="editClusterForm.name" disabled placeholder="集群名称不可修改" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="集群别名">
              <el-input v-model="editClusterForm.alias" placeholder="请输入集群别名（显示名称）" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="集群描述">
          <el-input v-model="editClusterForm.describe" type="textarea" :rows="2" placeholder="请输入集群描述" />
        </el-form-item>
        <ClusterConnectionForm
          v-model:auth-type="editClusterForm.authType"
          v-model:config="editClusterForm.config"
          v-model:uri="editClusterForm.uri"
          v-model:token="editClusterForm.token"
          :required-config="false"
          config-placeholder="如需更新配置，请粘贴完整的 kubeconfig 配置文件内容"
        />
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="handleEditCluster" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Edit,
  Delete,
  MoreFilled,
  ArrowRight,
  Search,
  Refresh,
  Loading
} from '@element-plus/icons-vue'
import { useClusterStore } from '@/stores/cluster'
import { clusterApi } from '@/api/cluster'
import ClusterConnectionForm from '@/components/ClusterConnectionForm.vue'

interface ClusterItem {
  id: string
  name: string
  alias?: string
  describe?: string
  apiServer: string
  status?: string
  version?: string
  nodeCount: number
  podCount: number
  namespaceCount: number
}

const PAGE_SIZE = 10
const PREFETCH_THRESHOLD = 3

const router = useRouter()
const clusterStore = useClusterStore()

const clusters = ref<ClusterItem[]>([])
const activeIndex = ref(0)
const renderPos = ref(0)
const hoveredIndex = ref<number | null>(null)
const visualActiveIndex = computed(() => {
  if (clusters.value.length === 0) return 0
  return Math.min(clusters.value.length - 1, Math.max(0, Math.round(renderPos.value)))
})
const loading = ref(false)
const loadingMore = ref(false)
const noMoreFromServer = ref(false)
const currentPage = ref(1)
const total = ref(0)
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const submitting = ref(false)
const formRef = ref()
const editFormRef = ref()
const viewportRef = ref<HTMLElement | null>(null)
const stageWidth = ref(typeof window !== 'undefined' ? window.innerWidth : 1440)

// 防止并发/过期请求覆盖新查询结果
let requestSeq = 0

// 搜索参数
const searchParams = ref({
  keyword: '',
  status: ''
})

const hasMore = computed(() => !noMoreFromServer.value && clusters.value.length < total.value)

// 当加载集群数量较多时，底部指示条保持滑动窗口（最多展示 11 个点）
const visibleDotIndices = computed<number[]>(() => {
  const len = clusters.value.length
  const maxDots = 11
  if (len <= maxDots) {
    return Array.from({ length: len }, (_, i) => i)
  }
  const half = Math.floor(maxDots / 2)
  let start = Math.max(0, activeIndex.value - half)
  let end = start + maxDots - 1
  if (end >= len) {
    end = len - 1
    start = Math.max(0, end - maxDots + 1)
  }
  return Array.from({ length: end - start + 1 }, (_, i) => start + i)
})

// 动态扇形叠卡几何计算：按可视宽度自适应分配每张后方卡片的露出面积，保证 5 张卡可同时露出并精准点选
const deckLayout = computed(() => {
  const w = Math.max(360, stageWidth.value)
  const cardW = Math.min(500, Math.max(290, w - 48))
  const halfCard = cardW / 2
  const halfStage = w / 2
  const availableSide = Math.max(110, halfStage - halfCard - 20)

  const scales = [1, 0.88, 0.78, 0.69, 0.61]
  const opacities = [1, 0.92, 0.8, 0.66, 0.5]
  const rotations = [0, 7, 11.5, 14.5, 16.5]

  let maxVisible: number
  let exposures: number[]

  if (w >= 1480) {
    maxVisible = 4
    exposures = [
      Math.min(340, availableSide * 0.54),
      Math.min(115, availableSide * 0.19),
      Math.min(92, availableSide * 0.15),
      Math.min(76, availableSide * 0.12)
    ]
  } else if (w >= 1080) {
    maxVisible = 4
    exposures = [
      Math.min(250, availableSide * 0.48),
      Math.min(96, availableSide * 0.21),
      Math.min(78, availableSide * 0.17),
      Math.min(62, availableSide * 0.14)
    ]
  } else if (w >= 760) {
    maxVisible = 3
    exposures = [
      Math.min(185, availableSide * 0.54),
      Math.min(82, availableSide * 0.26),
      Math.min(64, availableSide * 0.2),
      44
    ]
  } else {
    maxVisible = 2
    exposures = [Math.max(54, availableSide * 0.64), Math.max(38, availableSide * 0.36), 26, 20]
  }

  // 由目标露出右边缘反推每张卡中心坐标，确保所有可见卡片都在视口内且互不完全遮挡
  const txStops = [0]
  let prevEdge = halfCard
  for (let k = 1; k <= 4; k++) {
    const edge = prevEdge + exposures[k - 1]
    const center = edge - halfCard * scales[k]
    txStops.push(Math.max(txStops[k - 1] + 44, center))
    prevEdge = edge
  }

  const dragStep = Math.max(210, Math.min(330, Math.round(cardW * 0.62)))
  return { cardW, txStops, scales, opacities, rotations, maxVisible, dragStep }
})

const cardStep = computed(() => deckLayout.value.dragStep)

const lerpStops = (stops: number[], absD: number): number => {
  const maxIdx = stops.length - 1
  if (absD <= 0) return stops[0]
  if (absD >= maxIdx) {
    const lastSlope = stops[maxIdx] - stops[maxIdx - 1]
    return stops[maxIdx] + (absD - maxIdx) * lastSlope
  }
  const lower = Math.floor(absD)
  const frac = absD - lower
  return stops[lower] + (stops[lower + 1] - stops[lower]) * frac
}

// ==================== 平滑帧动画 + 移动端级顺滑拖动与多卡连滑 ====================
const isDragging = ref(false)

let rafId: number | null = null
let activePointerId: number | null = null
let pointerCaptured = false
let startX = 0
let startY = 0
let lastX = 0
let lastTime = 0
let velocityX = 0
let maxMovedPx = 0
let dragAxisLocked: 'x' | 'y' | null = null
let dragBasePos = 0
let suppressNextClick = false

const stopAnimation = () => {
  if (rafId !== null) {
    window.cancelAnimationFrame(rafId)
    rafId = null
  }
}

const animateToIndex = (targetIdx: number, immediate = false) => {
  const maxIdx = Math.max(0, clusters.value.length - 1)
  const clamped = Math.min(maxIdx, Math.max(0, targetIdx))
  stopAnimation()
  hoveredIndex.value = null
  activeIndex.value = clamped

  const startPos = renderPos.value
  const distance = clamped - startPos
  if (immediate || Math.abs(distance) < 0.002) {
    renderPos.value = clamped
    return
  }

  // 跨越多张卡片时自动延长动画时长，让经过的卡片依次流畅滑过并把目标卡送至正中
  const duration = Math.min(680, 320 + Math.abs(distance) * 85)
  const startTime = performance.now()

  const tick = (now: number) => {
    const t = Math.min(1, (now - startTime) / duration)
    // Cubic ease-out，起速自然、落位柔和
    const eased = 1 - Math.pow(1 - t, 3)
    renderPos.value = startPos + distance * eased
    if (t < 1) {
      rafId = window.requestAnimationFrame(tick)
    } else {
      renderPos.value = clamped
      rafId = null
    }
  }

  rafId = window.requestAnimationFrame(tick)
}

const isInteractiveTarget = (target: EventTarget | null) => {
  if (!(target instanceof HTMLElement)) return false
  return Boolean(target.closest('button, a, input, textarea, .el-dropdown, .el-dropdown-menu'))
}

const applyPosResistance = (rawPos: number) => {
  const maxIdx = Math.max(0, clusters.value.length - 1)
  if (rawPos < 0) {
    return -0.35 * Math.tanh(Math.abs(rawPos) / 0.9)
  }
  if (rawPos > maxIdx) {
    return maxIdx + 0.35 * Math.tanh((rawPos - maxIdx) / 0.9)
  }
  return rawPos
}

const handlePointerDown = (e: PointerEvent) => {
  if (clusters.value.length <= 1) return
  if (e.pointerType === 'mouse' && e.button !== 0) return
  if (isInteractiveTarget(e.target)) return

  activePointerId = e.pointerId
  pointerCaptured = false
  startX = e.clientX
  startY = e.clientY
  lastX = e.clientX
  lastTime = performance.now()
  velocityX = 0
  maxMovedPx = 0
  dragAxisLocked = null
  dragBasePos = renderPos.value
}

const handlePointerMove = (e: PointerEvent) => {
  if (activePointerId === null || e.pointerId !== activePointerId) return

  const dx = e.clientX - startX
  const dy = e.clientY - startY

  // 区分点击、横向滑动与纵向滚动；仅在确认横向拖动后才捕获指针，避免吞掉卡片点击事件
  if (!dragAxisLocked) {
    if (Math.abs(dx) < 6 && Math.abs(dy) < 6) return
    if (Math.abs(dy) > Math.abs(dx) * 1.15) {
      activePointerId = null
      isDragging.value = false
      return
    }
    dragAxisLocked = 'x'
    isDragging.value = true
    stopAnimation()
    dragBasePos = renderPos.value
    try {
      viewportRef.value?.setPointerCapture?.(e.pointerId)
      pointerCaptured = true
    } catch {
      pointerCaptured = false
    }
  }

  const now = performance.now()
  const dt = Math.max(1, now - lastTime)
  const instVel = (e.clientX - lastX) / dt
  velocityX = velocityX * 0.6 + instVel * 0.4
  lastX = e.clientX
  lastTime = now
  maxMovedPx = Math.max(maxMovedPx, Math.abs(dx))

  const rawPos = dragBasePos - dx / cardStep.value
  renderPos.value = applyPosResistance(rawPos)

  // 拉到末尾时提前触发动态加载
  if (renderPos.value >= clusters.value.length - 1.1 && dx < -30 && hasMore.value) {
    loadMoreClusters()
  }
}

const handlePointerUp = (e: PointerEvent) => {
  if (activePointerId === null || e.pointerId !== activePointerId) return

  if (pointerCaptured) {
    try {
      viewportRef.value?.releasePointerCapture?.(e.pointerId)
    } catch {
      // ignore
    }
    pointerCaptured = false
  }
  activePointerId = null

  if (!isDragging.value) {
    return
  }

  const maxIdx = Math.max(0, clusters.value.length - 1)
  let targetIdx = Math.round(renderPos.value)

  // 快速轻拂（flick）按惯性前进/后退
  if (Math.abs(velocityX) > 0.35 && Math.abs(renderPos.value - dragBasePos) < 0.65) {
    targetIdx = Math.round(dragBasePos) + (velocityX < 0 ? 1 : -1)
  } else if (Math.abs(velocityX) > 0.5) {
    const momentum = -velocityX * 0.25
    targetIdx = Math.round(renderPos.value + momentum)
  }

  targetIdx = Math.min(maxIdx, Math.max(0, targetIdx))

  if (maxMovedPx > 6) {
    suppressNextClick = true
    window.setTimeout(() => {
      suppressNextClick = false
    }, 80)
  }

  isDragging.value = false
  animateToIndex(targetIdx)
}

// 触控板双指左右滑 / 鼠标滚轮切换
let wheelAccum = 0
let wheelCooldownUntil = 0
let wheelResetTimer: number | undefined

const handleWheel = (e: WheelEvent) => {
  if (clusters.value.length <= 1 || showAddDialog.value || showEditDialog.value) return
  const primaryDelta = Math.abs(e.deltaX) > Math.abs(e.deltaY) ? e.deltaX : e.deltaY
  if (Math.abs(primaryDelta) < 2) return

  e.preventDefault()

  const now = performance.now()
  if (now < wheelCooldownUntil) return

  wheelAccum += primaryDelta
  if (wheelResetTimer) window.clearTimeout(wheelResetTimer)
  wheelResetTimer = window.setTimeout(() => {
    wheelAccum = 0
  }, 160)

  if (Math.abs(wheelAccum) >= 52) {
    const dir = wheelAccum > 0 ? 1 : -1
    const nextIdx = Math.min(clusters.value.length - 1, Math.max(0, activeIndex.value + dir))
    if (nextIdx !== activeIndex.value) {
      animateToIndex(nextIdx)
    } else if (dir > 0 && hasMore.value) {
      loadMoreClusters()
    }
    wheelAccum = 0
    wheelCooldownUntil = now + 260
  }
}

const goToIndex = (idx: number) => {
  animateToIndex(idx)
}

const handleCardLeave = (idx: number) => {
  if (hoveredIndex.value === idx) {
    hoveredIndex.value = null
  }
}

const handleCardClick = (idx: number, cluster: ClusterItem) => {
  if (suppressNextClick) return
  // 点击任意非居中卡片（包括后方第 2、3 张），平滑将其滑到正中间
  if (idx !== activeIndex.value || Math.abs(renderPos.value - idx) > 0.15) {
    animateToIndex(idx)
    return
  }
  enterCluster(cluster)
}

// 连续实时插值计算每张卡片的扇形层级与透明度（严格按 z-index 叠放与命中，杜绝 3D 透视吞点击）
const getCardStyle = (idx: number) => {
  const { cardW, txStops, scales, opacities, rotations, maxVisible } = deckLayout.value
  const d = idx - renderPos.value
  const absD = Math.abs(d)
  const hideLimit = maxVisible + 0.65

  // 远离可视范围的卡片完全隐藏，保证动态加载大量集群后依然流畅
  if (absD > hideLimit) {
    const farX = Math.sign(d) * (txStops[maxVisible] + 80)
    return {
      width: `${cardW}px`,
      transform: `translate3d(${farX.toFixed(1)}px, 0, 0) scale(0.54)`,
      opacity: 0,
      zIndex: 1,
      pointerEvents: 'none' as const,
      visibility: 'hidden' as const
    }
  }

  const sign = Math.sign(d)
  let tx = sign * lerpStops(txStops, absD)
  let scale = Math.max(0.55, lerpStops(scales, absD))
  let ty = 0
  const rotateY = -sign * Math.min(18, lerpStops(rotations, absD))

  // 悬停非居中卡片时轻微上浮并略微探出，给出明确的可点击反馈
  if (hoveredIndex.value === idx && idx !== visualActiveIndex.value && !isDragging.value) {
    ty = -7
    tx += sign * 10
    scale = Math.min(0.95, scale + 0.018)
  }

  let opacity: number
  if (absD <= maxVisible) {
    opacity = lerpStops(opacities, absD)
  } else {
    const edgeOpacity = opacities[Math.min(maxVisible, opacities.length - 1)]
    opacity = Math.max(0, edgeOpacity * (1 - (absD - maxVisible) / 0.65))
  }

  const zIndex = Math.max(2, Math.round(100 - absD * 14))
  const clickable = absD <= maxVisible + 0.2 && opacity > 0.15

  return {
    width: `${cardW}px`,
    transform: `translate3d(${tx.toFixed(1)}px, ${ty}px, 0) perspective(1200px) rotateY(${rotateY.toFixed(2)}deg) scale(${scale.toFixed(3)})`,
    opacity: Number(opacity.toFixed(3)),
    zIndex,
    pointerEvents: clickable ? ('auto' as const) : ('none' as const),
    visibility: 'visible' as const
  }
}

// 键盘左右方向键辅助切换
const handleKeydown = (e: KeyboardEvent) => {
  if (showAddDialog.value || showEditDialog.value || clusters.value.length <= 1) return
  const target = e.target as HTMLElement | null
  if (
    target &&
    (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable)
  ) {
    return
  }
  if (e.key === 'ArrowLeft') {
    goToIndex(activeIndex.value - 1)
  } else if (e.key === 'ArrowRight') {
    goToIndex(activeIndex.value + 1)
  }
}

const clusterForm = ref({
  name: '',
  alias: '',
  describe: '',
  authType: 'config',
  config: '',
  uri: '',
  token: ''
})

const editClusterForm = ref({
  name: '',
  alias: '',
  describe: '',
  authType: 'config',
  config: '',
  uri: '',
  token: ''
})

const addRules = {
  name: [
    { required: true, message: '请输入集群名称', trigger: 'blur' },
    {
      pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/,
      message: "仅支持小写字母、数字、'-' 和 '.'，以字母或数字开头结尾",
      trigger: 'blur'
    }
  ],
  config: [{ required: true, message: '请输入集群配置', trigger: 'blur' }],
  uri: [{ required: true, message: '请输入 API Server URI', trigger: 'blur' }],
  token: [{ required: true, message: '请输入 Token', trigger: 'blur' }]
}

const editRules = {
  name: [{ required: true, message: '集群名称为空', trigger: 'blur' }]
}

const mapClusterItem = (cluster: any): ClusterItem => ({
  id: cluster.id || cluster.name,
  name: cluster.name,
  alias: cluster.alias,
  describe: cluster.describe,
  apiServer: cluster.apiServer || '-',
  status: cluster.status,
  version: cluster.version,
  nodeCount: cluster.nodeCount || 0,
  podCount: cluster.podCount || 0,
  namespaceCount: cluster.namespaceCount || 0
})

// 首次加载 / 搜索 / 重置
const loadClusters = async (preserveIndex = false) => {
  const seq = ++requestSeq
  loading.value = true
  loadingMore.value = false
  noMoreFromServer.value = false
  currentPage.value = 1

  try {
    const response = await clusterApi.getClusters({
      page: 1,
      pageSize: PAGE_SIZE,
      keyword: searchParams.value.keyword || undefined,
      status: searchParams.value.status || undefined
    })

    if (seq !== requestSeq) return

    if (response && response.data) {
      const { items, total: totalCount } = response.data
      const mapped = (items || []).map(mapClusterItem)
      clusters.value = mapped
      total.value = totalCount ?? mapped.length
      if (mapped.length === 0 || mapped.length >= total.value) {
        noMoreFromServer.value = true
      }
      if (preserveIndex) {
        const nextIdx = Math.min(activeIndex.value, Math.max(0, mapped.length - 1))
        animateToIndex(nextIdx, true)
      } else {
        animateToIndex(0, true)
      }
    }
  } catch (error) {
    if (seq !== requestSeq) return
    console.error('Failed to load clusters:', error)
    ElMessage.error('加载集群列表失败')
    clusters.value = []
    total.value = 0
    animateToIndex(0, true)
  } finally {
    if (seq === requestSeq) {
      loading.value = false
    }
  }
}

// 滑动接近末尾时动态加载下一页
const loadMoreClusters = async () => {
  if (loading.value || loadingMore.value || !hasMore.value) return
  const seq = requestSeq
  const nextPage = currentPage.value + 1
  loadingMore.value = true

  try {
    const response = await clusterApi.getClusters({
      page: nextPage,
      pageSize: PAGE_SIZE,
      keyword: searchParams.value.keyword || undefined,
      status: searchParams.value.status || undefined
    })

    if (seq !== requestSeq) return

    if (response && response.data) {
      const { items, total: totalCount } = response.data
      const incoming = (items || []).map(mapClusterItem)
      if (typeof totalCount === 'number') {
        total.value = totalCount
      }
      if (incoming.length === 0) {
        noMoreFromServer.value = true
      } else {
        const existingIds = new Set(clusters.value.map((c) => c.id))
        const deduped = incoming.filter((item) => !existingIds.has(item.id))
        clusters.value = [...clusters.value, ...deduped]
        currentPage.value = nextPage
        if (deduped.length === 0 || clusters.value.length >= total.value) {
          noMoreFromServer.value = true
        }
      }
    }
  } catch (error) {
    if (seq !== requestSeq) return
    console.error('Failed to load more clusters:', error)
    ElMessage.error('加载更多集群失败')
  } finally {
    if (seq === requestSeq) {
      loadingMore.value = false
    }
  }
}

// 当滑动接近已加载列表尾部时自动预取下一页
watch(activeIndex, (idx) => {
  if (hasMore.value && !loading.value && !loadingMore.value) {
    if (idx >= clusters.value.length - PREFETCH_THRESHOLD) {
      loadMoreClusters()
    }
  }
})

// 搜索
const handleSearch = () => {
  loadClusters(false)
}

// 重置
const handleReset = () => {
  searchParams.value = {
    keyword: '',
    status: ''
  }
  loadClusters(false)
}

// 获取状态样式类
const getStatusClass = (status?: string) => {
  switch (status) {
    case 'Active':
    case 'Running':
      return 'is-running'
    case 'Error':
    case 'Failed':
      return 'is-error'
    case 'Pending':
      return 'is-warning'
    default:
      return 'is-unknown'
  }
}

// 获取状态文本
const getStatusText = (status?: string) => {
  const statusMap: Record<string, string> = {
    Active: '运行中',
    Running: '运行中',
    Error: '异常',
    Failed: '失败',
    Pending: '等待中',
    Unknown: '未知'
  }
  return (status && statusMap[status]) || status || '未知'
}

// 进入集群
const enterCluster = (cluster: ClusterItem) => {
  clusterStore.setCurrentCluster({
    id: cluster.name,
    name: cluster.name,
    apiServer: cluster.apiServer,
    status: cluster.status
  })
  router.push(`/cluster/${cluster.name}/overview`)
}

// 处理命令
const handleCommand = (command: string, cluster: ClusterItem) => {
  if (command === 'edit') {
    editClusterForm.value = {
      name: cluster.name,
      alias: cluster.alias || '',
      describe: cluster.describe || '',
      authType: 'config',
      config: '',
      uri: '',
      token: ''
    }
    showEditDialog.value = true
  } else if (command === 'delete') {
    ElMessageBox.confirm(`确定要删除集群 "${cluster.alias || cluster.name}" 吗？`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
      .then(async () => {
        try {
          await clusterApi.deleteCluster(cluster.name)
          ElMessage.success('删除成功')
          loadClusters(true)
        } catch (error) {
          console.error('Failed to delete cluster:', error)
          ElMessage.error('删除失败')
        }
      })
      .catch(() => {})
  }
}

// 添加集群
const handleAddCluster = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    submitting.value = true
    try {
      await clusterApi.createCluster({
        name: clusterForm.value.name,
        alias: clusterForm.value.alias,
        describe: clusterForm.value.describe,
        config: clusterForm.value.authType === 'config' ? clusterForm.value.config : undefined,
        uri: clusterForm.value.authType === 'token' ? clusterForm.value.uri : undefined,
        token: clusterForm.value.authType === 'token' ? clusterForm.value.token : undefined
      })
      ElMessage.success('添加成功')
      showAddDialog.value = false
      clusterForm.value = {
        name: '',
        alias: '',
        describe: '',
        authType: 'config',
        config: '',
        uri: '',
        token: ''
      }
      loadClusters(false)
    } catch (error) {
      console.error('Failed to add cluster:', error)
      ElMessage.error('添加失败')
    } finally {
      submitting.value = false
    }
  })
}

// 编辑集群
const handleEditCluster = async () => {
  if (!editFormRef.value) return
  await editFormRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    submitting.value = true
    try {
      await clusterApi.updateCluster(editClusterForm.value.name, {
        name: editClusterForm.value.name,
        alias: editClusterForm.value.alias,
        describe: editClusterForm.value.describe,
        config: editClusterForm.value.authType === 'config' ? editClusterForm.value.config : undefined,
        uri: editClusterForm.value.authType === 'token' ? editClusterForm.value.uri : undefined,
        token: editClusterForm.value.authType === 'token' ? editClusterForm.value.token : undefined
      })
      ElMessage.success('更新成功')
      showEditDialog.value = false
      loadClusters(true)
    } catch (error) {
      console.error('Failed to update cluster:', error)
      ElMessage.error('更新失败')
    } finally {
      submitting.value = false
    }
  })
}

let resizeObserver: ResizeObserver | null = null
const updateStageWidth = () => {
  stageWidth.value = viewportRef.value?.clientWidth || window.innerWidth
}

watch(viewportRef, (el) => {
  resizeObserver?.disconnect()
  if (el && typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => {
      updateStageWidth()
    })
    resizeObserver.observe(el)
    updateStageWidth()
  }
})

onMounted(() => {
  loadClusters(false)
  window.addEventListener('resize', updateStageWidth)
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  stopAnimation()
  resizeObserver?.disconnect()
  if (wheelResetTimer) window.clearTimeout(wheelResetTimer)
  window.removeEventListener('resize', updateStageWidth)
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<style scoped>
.svg-defs {
  position: absolute;
  width: 0;
  height: 0;
  overflow: hidden;
  pointer-events: none;
}

.cluster-list-page {
  width: 100%;
  padding: 6px 32px 24px;
  min-height: calc(100vh - 64px);
  display: flex;
  flex-direction: column;
}

/* ==================== 顶部横幅 ==================== */
.cluster-hero {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 26px 36px;
  border-radius: 20px;
  overflow: hidden;
  background: linear-gradient(120deg, #eff5ff 0%, #ddeaff 52%, #e9f2ff 100%);
  border: 1px solid rgba(255, 255, 255, 0.9);
  box-shadow:
    0 12px 32px rgba(37, 99, 235, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.95);
}

.hero-waves {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.hero-watermark {
  position: absolute;
  right: 120px;
  top: 50%;
  transform: translateY(-48%) rotate(12deg);
  width: 170px;
  height: 170px;
  pointer-events: none;
  opacity: 0.6;
}

.hero-left {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 22px;
}

.hero-badge {
  width: 68px;
  height: 68px;
  flex-shrink: 0;
  filter: drop-shadow(0 10px 18px rgba(37, 99, 235, 0.3));
}

.hero-badge-svg {
  width: 100%;
  height: 100%;
  display: block;
}

.hero-copy {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.hero-title {
  margin: 0;
  font-size: 28px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.25;
  letter-spacing: 0.4px;
}

.hero-desc {
  margin: 0;
  font-size: 14px;
  color: #475569;
  line-height: 1.5;
}

.hero-actions {
  position: relative;
  z-index: 2;
  flex-shrink: 0;
}

.add-cluster-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 42px;
  padding: 0 24px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, #4b94ff 0%, #2563eb 100%);
  color: #ffffff;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 8px 20px rgba(37, 99, 235, 0.35);
  transition: transform 0.2s ease, box-shadow 0.2s ease, filter 0.2s ease;
}

.add-cluster-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 10px 24px rgba(37, 99, 235, 0.42);
  filter: brightness(1.04);
}

.add-cluster-btn:active {
  transform: translateY(0);
}

/* ==================== 搜索筛选栏 ==================== */
.filter-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 14px;
  margin-top: 18px;
  padding: 14px 22px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.78);
  backdrop-filter: blur(14px);
  -webkit-backdrop-filter: blur(14px);
  border: 1px solid rgba(255, 255, 255, 0.92);
  box-shadow: 0 6px 24px rgba(37, 99, 235, 0.05);
}

.filter-left {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 14px;
}

.filter-summary {
  display: flex;
  align-items: center;
  margin-left: auto;
}

.total-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 36px;
  padding: 0 16px;
  border-radius: 999px;
  background: rgba(239, 246, 255, 0.95);
  border: 1px solid rgba(191, 219, 254, 0.9);
  color: #334155;
  font-size: 14px;
  font-weight: 500;
  white-space: nowrap;
}

.total-badge strong {
  color: #1d4ed8;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.total-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: linear-gradient(135deg, #4b94ff 0%, #2563eb 100%);
  box-shadow: 0 0 6px rgba(37, 99, 235, 0.45);
}

.keyword-input {
  width: 260px;
}

.keyword-input :deep(.el-input__wrapper) {
  height: 38px;
  border-radius: 10px !important;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 0 0 1px #dbe4f2 inset;
  padding: 0 12px;
}

.keyword-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1.5px #3b82f6 inset;
}

.status-filter {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.filter-label {
  font-size: 14px;
  color: #334155;
  font-weight: 500;
  white-space: nowrap;
}

.status-select {
  width: 144px;
}

.status-select :deep(.el-select__wrapper),
.status-select :deep(.el-input__wrapper) {
  height: 38px;
  min-height: 38px;
  border-radius: 999px !important;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 0 0 1px #dbe4f2 inset;
  padding-left: 16px;
  padding-right: 12px;
}

.pill-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 38px;
  padding: 0 22px;
  border-radius: 999px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.pill-btn-primary {
  border: none;
  background: linear-gradient(135deg, #4b94ff 0%, #2563eb 100%);
  color: #ffffff;
  box-shadow: 0 6px 16px rgba(37, 99, 235, 0.28);
}

.pill-btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 20px rgba(37, 99, 235, 0.36);
  filter: brightness(1.04);
}

.pill-btn-ghost {
  border: 1px solid #d5e0f0;
  background: rgba(255, 255, 255, 0.92);
  color: #334155;
}

.pill-btn-ghost:hover {
  border-color: #93c5fd;
  color: #2563eb;
  background: #ffffff;
}

/* ==================== 3D 滑动焦点区 ==================== */
.cluster-stage {
  position: relative;
  flex: 1;
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 28px 0 18px;
  min-height: 440px;
  overflow: hidden;
}

.empty-card {
  width: 100%;
  max-width: 520px;
  padding: 40px 24px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.78);
  border: 1px solid rgba(255, 255, 255, 0.92);
  box-shadow: 0 12px 32px rgba(37, 99, 235, 0.08);
  backdrop-filter: blur(12px);
}

.carousel-viewport {
  position: relative;
  width: 100%;
  height: 356px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: grab;
  user-select: none;
  touch-action: pan-y;
}

.carousel-viewport.is-dragging {
  cursor: grabbing;
}

.carousel-track {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cluster-card {
  position: absolute;
  max-width: calc(100% - 40px);
  min-height: 292px;
  padding: 28px 32px 26px;
  border-radius: 24px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  overflow: hidden;
  cursor: pointer;
  user-select: none;
  background: rgba(255, 255, 255, 0.84);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border: 1.5px solid rgba(219, 234, 254, 0.95);
  box-shadow: 0 14px 34px rgba(37, 99, 235, 0.11);
  transition:
    box-shadow 0.25s ease,
    background 0.25s ease,
    border-color 0.25s ease;
  will-change: transform, opacity;
}

.carousel-viewport.is-dragging .cluster-card {
  cursor: grabbing;
}

.cluster-card.is-active {
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  border: 1.5px solid rgba(255, 255, 255, 0.98);
  box-shadow:
    0 24px 56px rgba(37, 99, 235, 0.18),
    0 6px 18px rgba(15, 23, 42, 0.05);
}

/* 非居中卡片整张表面只负责“点击滑到正中”，避免误触侧卡边缘的进入集群或更多按钮 */
.cluster-card.is-side .enter-btn,
.cluster-card.is-side .more-trigger,
.cluster-card.is-side :deep(.el-dropdown) {
  pointer-events: none;
}

.cluster-card.is-side.is-hovered {
  background: rgba(255, 255, 255, 0.95);
  border-color: rgba(96, 165, 250, 0.9);
  box-shadow: 0 18px 42px rgba(37, 99, 235, 0.18);
}

.card-watermark {
  position: absolute;
  right: -14px;
  top: 52%;
  transform: translateY(-50%) rotate(10deg);
  width: 164px;
  height: 164px;
  pointer-events: none;
  opacity: 0.32;
  transition: opacity 0.35s ease;
}

.cluster-card.is-active .card-watermark {
  opacity: 0.95;
}

/* 卡片头部 */
.card-header {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 18px;
}

.card-identity {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
  flex: 1;
}

.cloud-emblem {
  width: 60px;
  height: 50px;
  flex-shrink: 0;
  filter: drop-shadow(0 8px 14px rgba(37, 99, 235, 0.28));
}

.cluster-card:not(.is-active) .cloud-emblem {
  filter: drop-shadow(0 6px 10px rgba(59, 130, 246, 0.18));
}

.cloud-svg {
  width: 100%;
  height: 100%;
  display: block;
}

.title-wrap {
  min-width: 0;
  flex: 1;
}

.title-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.cluster-name {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1.25;
  max-width: 250px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cluster-card:not(.is-active) .cluster-name {
  color: #334155;
}

.cluster-describe {
  margin: 4px 0 0;
  font-size: 12px;
  color: #64748b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 状态胶囊 */
.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.2;
  white-space: nowrap;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-pill.is-running {
  background: rgba(34, 197, 94, 0.14);
  color: #16a34a;
}
.status-pill.is-running .status-dot {
  background: #22c55e;
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.55);
}

.status-pill.is-error {
  background: rgba(239, 68, 68, 0.12);
  color: #dc2626;
}
.status-pill.is-error .status-dot {
  background: #ef4444;
  box-shadow: 0 0 6px rgba(239, 68, 68, 0.5);
}

.status-pill.is-warning {
  background: rgba(245, 158, 11, 0.14);
  color: #d97706;
}
.status-pill.is-warning .status-dot {
  background: #f59e0b;
}

.status-pill.is-unknown {
  background: rgba(148, 163, 184, 0.18);
  color: #64748b;
}
.status-pill.is-unknown .status-dot {
  background: #94a3b8;
}

.more-trigger {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: #64748b;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.2s ease, color 0.2s ease;
}

.more-trigger:hover {
  background: rgba(59, 130, 246, 0.1);
  color: #2563eb;
}

/* 卡片中部信息 */
.card-meta {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 18px 0;
  border-top: 1px solid rgba(148, 163, 184, 0.16);
  border-bottom: 1px solid rgba(148, 163, 184, 0.16);
}

.meta-row {
  display: grid;
  grid-template-columns: 96px 1fr;
  align-items: center;
  gap: 12px;
  font-size: 14px;
}

.meta-label {
  color: #64748b;
  font-weight: 500;
}

.meta-value {
  color: #1e293b;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-variant-numeric: tabular-nums;
}

.cluster-card:not(.is-active) .meta-value {
  color: #475569;
}

/* 卡片底部 */
.card-footer {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 18px;
}

.card-stats {
  display: flex;
  align-items: center;
  gap: 28px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.stat-top {
  display: flex;
  align-items: center;
  gap: 7px;
}

.stat-icon {
  width: 19px;
  height: 19px;
  color: #2563eb;
  flex-shrink: 0;
}

.cluster-card:not(.is-active) .stat-icon {
  color: #60a5fa;
}

.stat-num {
  font-size: 20px;
  font-weight: 700;
  color: #0f172a;
  line-height: 1;
  font-variant-numeric: tabular-nums;
}

.cluster-card:not(.is-active) .stat-num {
  color: #475569;
}

.stat-label {
  font-size: 12px;
  color: #64748b;
  line-height: 1.2;
}

.enter-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 40px;
  padding: 0 22px;
  border-radius: 999px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  border: 1px solid rgba(59, 130, 246, 0.35);
  background: rgba(255, 255, 255, 0.72);
  color: #3b82f6;
  transition: all 0.22s ease;
}

.enter-btn:hover {
  border-color: #2563eb;
  background: #ffffff;
  color: #1d4ed8;
}

.enter-btn.is-primary {
  border: none;
  background: linear-gradient(135deg, #4b94ff 0%, #2563eb 100%);
  color: #ffffff;
  box-shadow: 0 8px 22px rgba(37, 99, 235, 0.36);
}

.enter-btn.is-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 10px 26px rgba(37, 99, 235, 0.44);
  filter: brightness(1.04);
}

/* ==================== 底部指示条与状态 ==================== */
.stage-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  margin-top: 22px;
}

.carousel-dots {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
}

.carousel-dot {
  width: 14px;
  height: 6px;
  padding: 0;
  border: none;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.35);
  cursor: pointer;
  transition: all 0.28s ease;
}

.carousel-dot:hover {
  background: rgba(59, 130, 246, 0.45);
}

.carousel-dot.is-active {
  width: 26px;
  background: linear-gradient(90deg, #4b94ff 0%, #2563eb 100%);
  box-shadow: 0 2px 8px rgba(37, 99, 235, 0.38);
}

.stage-meta {
  display: inline-flex;
  align-items: center;
  flex-wrap: wrap;
  justify-content: center;
  gap: 12px;
  padding: 6px 16px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.65);
  border: 1px solid rgba(255, 255, 255, 0.9);
  box-shadow: 0 4px 14px rgba(37, 99, 235, 0.05);
  font-size: 13px;
  color: #64748b;
}

.stage-counter strong {
  color: #1d4ed8;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.stage-tip {
  color: #94a3b8;
  font-size: 12px;
}

.stage-loading {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #2563eb;
  font-weight: 500;
}

/* ==================== 响应式适配 ==================== */
@media (max-width: 900px) {
  .cluster-list-page {
    padding: 8px 16px 20px;
  }

  .cluster-hero {
    padding: 20px;
    flex-direction: column;
    align-items: flex-start;
  }

  .hero-badge {
    width: 56px;
    height: 56px;
  }

  .hero-title {
    font-size: 24px;
  }

  .hero-watermark {
    display: none;
  }

  .filter-bar {
    padding: 14px 16px;
  }

  .filter-left {
    width: 100%;
  }

  .filter-summary {
    margin-left: 0;
  }

  .keyword-input {
    width: 100%;
  }

  .status-filter {
    width: 100%;
  }

  .status-select {
    flex: 1;
  }

  .cluster-card {
    width: calc(100% - 40px);
    max-width: 460px;
    padding: 22px;
  }

  .card-stats {
    gap: 18px;
  }
}

@media (max-width: 520px) {
  .card-footer {
    flex-direction: column;
    align-items: stretch;
    gap: 14px;
  }

  .card-stats {
    justify-content: space-around;
  }

  .enter-btn {
    width: 100%;
  }

  .meta-row {
    grid-template-columns: 80px 1fr;
    font-size: 13px;
  }
}
</style>
