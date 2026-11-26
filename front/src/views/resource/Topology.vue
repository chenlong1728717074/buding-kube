<template>
  <div class="topology-page">
    <PageHeader
      title="资源拓扑图"
      subtitle="按命名空间可视化查看 Kubernetes 资源的流量转发、所属管理与配置存储依赖关系"
      icon="netpol"
    >
      <el-button :disabled="layout.nodes.length === 0" @click="toggleExpand(true)">
        <el-icon><FullScreen /></el-icon>
        放大查看
      </el-button>
      <el-button :loading="loading" type="primary" @click="fetchTopology">
        <el-icon><Refresh /></el-icon>
        刷新拓扑
      </el-button>
    </PageHeader>

    <!-- 控制栏 -->
    <el-card class="toolbar-card" shadow="never">
      <div class="toolbar-row">
        <div class="toolbar-left">
          <div class="control-item">
            <span class="control-label">命名空间</span>
            <el-select
              v-model="selectedNamespace"
              filterable
              placeholder="请选择命名空间"
              style="width: 200px"
              :loading="nsLoading"
              @change="fetchTopology"
            >
              <el-option v-for="ns in namespaces" :key="ns" :label="ns" :value="ns" />
            </el-select>
          </div>

          <el-input
            v-model="keyword"
            placeholder="搜索定位资源名称"
            clearable
            style="width: 190px"
            @keyup.enter="focusFirstMatch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>

          <el-radio-group v-model="layoutDirection" size="small" class="direction-switch">
            <el-radio-button label="TB">上下布局</el-radio-button>
            <el-radio-button label="LR">左右布局</el-radio-button>
          </el-radio-group>

          <el-checkbox v-model="showEdgeLabels">显示绑定信息</el-checkbox>
          <el-checkbox v-model="onlyConnected">隐藏孤立资源</el-checkbox>
          <el-checkbox v-model="showReplicaSet">显示 ReplicaSet</el-checkbox>
        </div>

        <div class="toolbar-right">
          <div class="zoom-controls">
            <el-tooltip content="缩小">
              <el-button size="small" @click="zoomBy(-0.15)">
                <el-icon><ZoomOut /></el-icon>
              </el-button>
            </el-tooltip>
            <span class="zoom-label">{{ Math.round(scale * 100) }}%</span>
            <el-tooltip content="放大">
              <el-button size="small" @click="zoomBy(0.15)">
                <el-icon><ZoomIn /></el-icon>
              </el-button>
            </el-tooltip>
            <el-tooltip content="自适应居中">
              <el-button size="small" @click="fitView">
                <el-icon><Aim /></el-icon>
                自适应
              </el-button>
            </el-tooltip>
            <el-tooltip content="弹窗放大查看">
              <el-button size="small" type="primary" plain @click="toggleExpand(true)">
                <el-icon><FullScreen /></el-icon>
                放大弹窗
              </el-button>
            </el-tooltip>
          </div>
        </div>
      </div>

      <!-- 资源类型筛选胶囊 -->
      <div v-if="kindStats.length" class="kind-filters">
        <button
          type="button"
          class="kind-chip"
          :class="{ 'is-active': activeKinds.size === 0 }"
          @click="clearKindFilter"
        >
          <span>全部资源</span>
          <span class="chip-count">{{ visibleGraph.nodes.length }}</span>
        </button>
        <button
          v-for="item in kindStats"
          :key="item.kind"
          type="button"
          class="kind-chip"
          :class="{ 'is-active': activeKinds.has(item.kind) }"
          :style="{ '--chip-color': item.color }"
          @click="toggleKind(item.kind)"
        >
          <K8sIcon :name="item.icon" :size="15" />
          <span>{{ item.kind }}</span>
          <span class="chip-count">{{ item.count }}</span>
        </button>
      </div>
    </el-card>

    <!-- 拓扑画布主体（支持一键弹出大窗口） -->
    <Teleport to="body" :disabled="!isExpanded">
      <div
        class="topo-modal-shell"
        :class="{ 'is-expanded': isExpanded }"
        @click.self="isExpanded && toggleExpand(false)"
      >
        <div class="topo-modal-dialog" :class="{ 'is-expanded': isExpanded }">
          <!-- 弹窗放大模式的顶部工具栏 -->
          <div v-if="isExpanded" class="modal-topbar">
            <div class="modal-title-group">
              <K8sIcon name="netpol" :size="20" />
              <span class="modal-title">资源拓扑图</span>
              <el-tag v-if="selectedNamespace" size="small" effect="plain">
                {{ selectedNamespace }}
              </el-tag>
              <span class="modal-stat">
                {{ layout.nodes.length }} 个资源 · {{ layout.edges.length }} 条关联
              </span>
            </div>

            <div class="modal-actions">
              <el-input
                v-model="keyword"
                size="small"
                placeholder="搜索定位资源"
                clearable
                style="width: 170px"
                @keyup.enter="focusFirstMatch"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>

              <el-radio-group v-model="layoutDirection" size="small">
                <el-radio-button label="TB">上下</el-radio-button>
                <el-radio-button label="LR">左右</el-radio-button>
              </el-radio-group>

              <el-checkbox v-model="showEdgeLabels" size="small">绑定信息</el-checkbox>
              <el-checkbox v-model="onlyConnected" size="small">隐藏孤立</el-checkbox>
              <el-checkbox v-model="showReplicaSet" size="small">ReplicaSet</el-checkbox>

              <div class="zoom-controls">
                <el-button size="small" @click="zoomBy(-0.15)">
                  <el-icon><ZoomOut /></el-icon>
                </el-button>
                <span class="zoom-label">{{ Math.round(scale * 100) }}%</span>
                <el-button size="small" @click="zoomBy(0.15)">
                  <el-icon><ZoomIn /></el-icon>
                </el-button>
                <el-button size="small" @click="fitView">
                  <el-icon><Aim /></el-icon>
                  自适应
                </el-button>
              </div>

              <el-button size="small" @click="toggleExpand(false)">
                <el-icon><Close /></el-icon>
                退出放大
              </el-button>
            </div>
          </div>

          <div class="workspace-card" :class="{ 'is-expanded': isExpanded }" v-loading="loading">
            <div
              ref="canvasRef"
              class="topo-canvas"
              :class="{ 'is-panning': isPanning }"
              @pointerdown="onCanvasPointerDown"
              @pointermove="onCanvasPointerMove"
              @pointerup="onCanvasPointerUp"
              @pointercancel="onCanvasPointerUp"
              @wheel.prevent="onCanvasWheel"
              @dblclick="fitView"
              @click="onCanvasClick"
            >
              <!-- 画布右上角快捷放大/退出按钮 -->
              <div v-if="layout.nodes.length > 0 && !isExpanded" class="canvas-float-actions" @pointerdown.stop @click.stop>
                <el-tooltip content="点击放大弹窗查看完整拓扑" placement="left">
                  <button type="button" class="float-expand-btn" @click="toggleExpand(true)">
                    <el-icon><FullScreen /></el-icon>
                    <span>放大查看</span>
                  </button>
                </el-tooltip>
              </div>

              <!-- 悬停连线时的绑定信息浮层 -->
              <div v-if="hoveredEdgeInfo" class="edge-hover-banner" @pointerdown.stop @click.stop>
                <div class="edge-hover-head">
                  <el-tag size="small" type="primary" effect="dark">{{ hoveredEdgeInfo.edge.relation }}</el-tag>
                  <span class="edge-hover-endpoints">
                    <strong>{{ hoveredEdgeInfo.source?.kind }}/{{ hoveredEdgeInfo.source?.name }}</strong>
                    <span class="arrow-sep">→</span>
                    <strong>{{ hoveredEdgeInfo.target?.kind }}/{{ hoveredEdgeInfo.target?.name }}</strong>
                  </span>
                </div>
                <div v-if="hoveredEdgeInfo.details.length" class="edge-hover-details">
                  <span v-for="(line, idx) in hoveredEdgeInfo.details" :key="idx" class="edge-hover-pill">
                    {{ line }}
                  </span>
                </div>
              </div>

              <div v-if="!loading && !selectedNamespace" class="canvas-empty">
                <el-empty description="请选择一个命名空间查看资源关联拓扑" />
              </div>
              <div v-else-if="!loading && layout.nodes.length === 0" class="canvas-empty">
                <el-empty
                  :description="
                    onlyConnected
                      ? '该命名空间下暂无有关联关系的资源，可取消勾选“隐藏孤立资源”查看全部'
                      : '该命名空间下暂无可展示的资源'
                  "
                />
              </div>

              <div
                v-else
                class="topo-stage"
                :style="{
                  width: `${layout.width}px`,
                  height: `${layout.height}px`,
                  transform: `translate3d(${panX}px, ${panY}px, 0) scale(${scale})`
                }"
              >
                <!-- 分层泳道背景与标题（兼容上下布局 / 左右布局） -->
                <div
                  v-for="lane in layout.lanes"
                  :key="lane.key"
                  class="lane-band"
                  :class="layoutDirection === 'TB' ? 'is-horizontal' : 'is-vertical'"
                  :style="{
                    left: `${lane.x}px`,
                    top: `${lane.y}px`,
                    width: `${lane.width}px`,
                    height: `${lane.height}px`
                  }"
                >
                  <div class="lane-header" :class="layoutDirection === 'TB' ? 'is-horizontal' : 'is-vertical'">
                    <span class="lane-title">{{ lane.label }}</span>
                    <span class="lane-count">{{ lane.count }}</span>
                  </div>
                </div>

                <!-- 关系连线 SVG 层 -->
                <svg
                  class="edges-svg"
                  :width="layout.width"
                  :height="layout.height"
                  :viewBox="`0 0 ${layout.width} ${layout.height}`"
                >
                  <defs>
                    <marker
                      id="topo-arrow"
                      viewBox="0 0 10 10"
                      refX="9"
                      refY="5"
                      markerWidth="6.5"
                      markerHeight="6.5"
                      orient="auto"
                    >
                      <path d="M 0 1.2 L 9 5 L 0 8.8 z" fill="#94a3b8" />
                    </marker>
                    <marker
                      id="topo-arrow-active"
                      viewBox="0 0 10 10"
                      refX="9"
                      refY="5"
                      markerWidth="7"
                      markerHeight="7"
                      orient="auto"
                    >
                      <path d="M 0 1.2 L 9 5 L 0 8.8 z" fill="#2563eb" />
                    </marker>
                  </defs>

                  <g v-for="edge in layout.edges" :key="edge.id">
                    <!-- 宽点击热区，便于鼠标悬停和点击查看绑定信息 -->
                    <path
                      :d="edge.path"
                      class="topo-edge-hit"
                      @mouseenter="hoveredEdgeId = edge.id"
                      @mouseleave="hoveredEdgeId = null"
                      @pointerdown.stop
                      @click.stop="selectEdge(edge.id)"
                    />
                    <path
                      :d="edge.path"
                      class="topo-edge"
                      :class="{
                        'is-highlighted': isEdgeHighlighted(edge),
                        'is-selected': selectedEdgeId === edge.id,
                        'is-dimmed': hasFocus && !isEdgeHighlighted(edge)
                      }"
                      :marker-end="isEdgeHighlighted(edge) ? 'url(#topo-arrow-active)' : 'url(#topo-arrow)'"
                    />
                    <!-- 连线绑定信息标签 -->
                    <g
                      v-if="shouldShowEdgeLabel(edge)"
                      :transform="`translate(${edge.labelX}, ${edge.labelY})`"
                      class="edge-label-group"
                      :class="{
                        'is-highlighted': isEdgeHighlighted(edge),
                        'is-selected': selectedEdgeId === edge.id,
                        'is-dimmed': hasFocus && !isEdgeHighlighted(edge)
                      }"
                      @mouseenter="hoveredEdgeId = edge.id"
                      @mouseleave="hoveredEdgeId = null"
                      @pointerdown.stop
                      @click.stop="selectEdge(edge.id)"
                    >
                      <rect
                        :x="-edge.pillWidth / 2"
                        y="-11"
                        :width="edge.pillWidth"
                        height="22"
                        rx="11"
                        class="edge-label-bg"
                      />
                      <text text-anchor="middle" dominant-baseline="central" class="edge-label-text">
                        {{ edge.displayText }}
                      </text>
                    </g>
                  </g>
                </svg>

                <!-- 资源节点卡片层 -->
                <div
                  v-for="node in layout.nodes"
                  :key="node.id"
                  class="topo-node"
                  :class="{
                    'is-selected': selectedNodeId === node.id,
                    'is-highlighted': isNodeHighlighted(node.id),
                    'is-dimmed': hasFocus && !isNodeHighlighted(node.id),
                    'is-search-hit': isSearchHit(node),
                    [`status-${node.statusType}`]: true
                  }"
                  :style="{
                    left: `${node.x}px`,
                    top: `${node.y}px`,
                    width: `${NODE_W}px`,
                    height: `${NODE_H}px`,
                    '--kind-accent': kindMeta(node.kind).color
                  }"
                  @mouseenter="hoveredNodeId = node.id"
                  @mouseleave="hoveredNodeId = null"
                  @pointerdown.stop
                  @dblclick.stop
                  @click.stop="selectNode(node.id)"
                >
                  <div class="node-accent-bar" />
                  <div class="node-body">
                    <div class="node-head">
                      <K8sIcon :name="kindMeta(node.kind).icon" :size="24" />
                      <div class="node-title-col">
                        <div class="node-name" :title="node.name">{{ node.name }}</div>
                        <div class="node-kind-row">
                          <span class="node-kind">{{ node.kind }}</span>
                          <span class="node-status-dot" :class="`dot-${node.statusType}`" />
                          <span class="node-status-text">{{ node.status }}</span>
                        </div>
                      </div>
                    </div>
                    <div class="node-summary" :title="node.summary">{{ node.summary || '-' }}</div>
                  </div>
                </div>
              </div>

              <!-- 左下角图例与操作提示 -->
              <div v-if="layout.nodes.length > 0" class="canvas-legend">
                <span class="legend-item"><i class="dot dot-success" />正常</span>
                <span class="legend-item"><i class="dot dot-warning" />更新 / 等待</span>
                <span class="legend-item"><i class="dot dot-danger" />异常 / 缺失</span>
                <span class="legend-sep">|</span>
                <span class="legend-tip">滚轮缩放 · 拖拽画布 · 点击节点或连线查看绑定详情</span>
              </div>
            </div>

            <!-- 右侧详情侧栏：选中节点 -->
            <aside v-if="selectedNode" class="inspector-panel">
              <div class="inspector-header">
                <div class="inspector-title">
                  <K8sIcon :name="kindMeta(selectedNode.kind).icon" :size="26" />
                  <div class="inspector-name-wrap">
                    <div class="inspector-name" :title="selectedNode.name">{{ selectedNode.name }}</div>
                    <div class="inspector-sub">
                      <el-tag size="small" effect="plain">{{ selectedNode.kind }}</el-tag>
                      <el-tag size="small" :type="tagTypeOf(selectedNode.statusType)">{{ selectedNode.status }}</el-tag>
                    </div>
                  </div>
                </div>
                <el-button text circle size="small" @click="selectedNodeId = null">
                  <el-icon><Close /></el-icon>
                </el-button>
              </div>

              <div class="inspector-body">
                <div class="inspector-section">
                  <div class="section-title">基本信息</div>
                  <div class="kv-list">
                    <div class="kv-row">
                      <span class="kv-key">命名空间</span>
                      <span class="kv-val">{{ selectedNode.namespace }}</span>
                    </div>
                    <div class="kv-row">
                      <span class="kv-key">摘要</span>
                      <span class="kv-val">{{ selectedNode.summary || '-' }}</span>
                    </div>
                    <div v-for="(val, key) in selectedNode.extra || {}" :key="key" class="kv-row">
                      <span class="kv-key">{{ key }}</span>
                      <span class="kv-val mono">{{ val || '-' }}</span>
                    </div>
                  </div>
                </div>

                <div class="inspector-section">
                  <div class="section-title">
                    <span>上游关联</span>
                    <span class="section-count">{{ selectedRelations.incoming.length }}</span>
                  </div>
                  <div v-if="selectedRelations.incoming.length" class="rel-list">
                    <button
                      v-for="item in selectedRelations.incoming"
                      :key="item.edge.id"
                      type="button"
                      class="rel-item"
                      @click="focusNodeById(item.node.id)"
                    >
                      <K8sIcon :name="kindMeta(item.node.kind).icon" :size="18" />
                      <div class="rel-info">
                        <span class="rel-name">{{ item.node.name }}</span>
                        <span class="rel-meta">{{ item.node.kind }} · {{ item.edge.relation }}</span>
                        <span v-if="edgeBindingPreview(item.edge)" class="rel-binding">
                          {{ edgeBindingPreview(item.edge) }}
                        </span>
                      </div>
                      <el-icon class="rel-arrow"><ArrowRight /></el-icon>
                    </button>
                  </div>
                  <div v-else class="rel-empty">无上游入口依赖</div>
                </div>

                <div class="inspector-section">
                  <div class="section-title">
                    <span>下游关联</span>
                    <span class="section-count">{{ selectedRelations.outgoing.length }}</span>
                  </div>
                  <div v-if="selectedRelations.outgoing.length" class="rel-list">
                    <button
                      v-for="item in selectedRelations.outgoing"
                      :key="item.edge.id"
                      type="button"
                      class="rel-item"
                      @click="focusNodeById(item.node.id)"
                    >
                      <K8sIcon :name="kindMeta(item.node.kind).icon" :size="18" />
                      <div class="rel-info">
                        <span class="rel-name">{{ item.node.name }}</span>
                        <span class="rel-meta">{{ item.edge.relation }} → {{ item.node.kind }}</span>
                        <span v-if="edgeBindingPreview(item.edge)" class="rel-binding">
                          {{ edgeBindingPreview(item.edge) }}
                        </span>
                      </div>
                      <el-icon class="rel-arrow"><ArrowRight /></el-icon>
                    </button>
                  </div>
                  <div v-else class="rel-empty">无下游依赖资源</div>
                </div>
              </div>

              <div v-if="canOpenDetail(selectedNode)" class="inspector-footer">
                <el-button type="primary" style="width: 100%" @click="openResourceDetail(selectedNode)">
                  查看资源详情
                </el-button>
              </div>
            </aside>

            <!-- 右侧详情侧栏：选中链路连线 -->
            <aside v-else-if="selectedEdgeInfo" class="inspector-panel">
              <div class="inspector-header">
                <div class="inspector-title">
                  <div class="edge-inspector-icon">
                    <el-icon :size="18"><Connection /></el-icon>
                  </div>
                  <div class="inspector-name-wrap">
                    <div class="inspector-name">链路绑定详情</div>
                    <div class="inspector-sub">
                      <el-tag size="small" type="primary">{{ selectedEdgeInfo.edge.relation }}</el-tag>
                      <span v-if="selectedEdgeInfo.edge.label" class="edge-sub-label">
                        {{ selectedEdgeInfo.edge.label }}
                      </span>
                    </div>
                  </div>
                </div>
                <el-button text circle size="small" @click="selectedEdgeId = null">
                  <el-icon><Close /></el-icon>
                </el-button>
              </div>

              <div class="inspector-body">
                <div class="inspector-section">
                  <div class="section-title">关联双方</div>
                  <div class="edge-endpoint-cards">
                    <button
                      v-if="selectedEdgeInfo.source"
                      type="button"
                      class="rel-item"
                      @click="focusNodeById(selectedEdgeInfo.source.id)"
                    >
                      <K8sIcon :name="kindMeta(selectedEdgeInfo.source.kind).icon" :size="18" />
                      <div class="rel-info">
                        <span class="rel-name">{{ selectedEdgeInfo.source.name }}</span>
                        <span class="rel-meta">上游 · {{ selectedEdgeInfo.source.kind }}</span>
                      </div>
                      <el-icon class="rel-arrow"><ArrowRight /></el-icon>
                    </button>

                    <div class="endpoint-flow-arrow">
                      <span>{{ selectedEdgeInfo.edge.relation }}</span>
                    </div>

                    <button
                      v-if="selectedEdgeInfo.target"
                      type="button"
                      class="rel-item"
                      @click="focusNodeById(selectedEdgeInfo.target.id)"
                    >
                      <K8sIcon :name="kindMeta(selectedEdgeInfo.target.kind).icon" :size="18" />
                      <div class="rel-info">
                        <span class="rel-name">{{ selectedEdgeInfo.target.name }}</span>
                        <span class="rel-meta">下游 · {{ selectedEdgeInfo.target.kind }}</span>
                      </div>
                      <el-icon class="rel-arrow"><ArrowRight /></el-icon>
                    </button>
                  </div>
                </div>

                <div class="inspector-section">
                  <div class="section-title">
                    <span>绑定明细</span>
                    <span class="section-count">{{ selectedEdgeInfo.details.length }}</span>
                  </div>
                  <div v-if="selectedEdgeInfo.details.length" class="binding-detail-list">
                    <div
                      v-for="(item, idx) in selectedEdgeInfo.details"
                      :key="idx"
                      class="binding-detail-item"
                    >
                      {{ item }}
                    </div>
                  </div>
                  <div v-else class="rel-empty">暂无更多绑定明细</div>
                </div>
              </div>
            </aside>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  Aim,
  ArrowRight,
  Close,
  Connection,
  FullScreen,
  Refresh,
  Search,
  ZoomIn,
  ZoomOut
} from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import K8sIcon, { type K8sIconName } from '@/components/K8sIcon.vue'
import { useClusterStore } from '@/stores/cluster'
import { namespaceApi } from '@/api/namespace'
import { topologyApi, type TopologyEdgeVO, type TopologyNodeVO, type TopologyStatusType } from '@/api/topology'

const NODE_W = 224
const NODE_H = 74

// 左右布局常量
const LR_COL_GAP = 140
const LR_ROW_GAP = 22
const LR_PAD_X = 44
const LR_LANE_TOP = 60
const LR_PAD_BOTTOM = 44

// 上下布局常量
const TB_COL_GAP = 28
const TB_ROW_GAP = 104
const TB_PAD_X = 32
const TB_PAD_Y = 30
const TB_HEADER_W = 148
const TB_LANE_PAD_X = 22
const TB_LANE_PAD_Y = 16

type LayoutDirection = 'TB' | 'LR'

interface KindMeta {
  icon: K8sIconName
  color: string
  order: number
}

const KIND_META: Record<string, KindMeta> = {
  Ingress: { icon: 'ing', color: '#0ea5e9', order: 1 },
  HPA: { icon: 'hpa', color: '#06b6d4', order: 2 },
  Service: { icon: 'svc', color: '#2563eb', order: 3 },
  Deployment: { icon: 'deploy', color: '#4f46e5', order: 4 },
  StatefulSet: { icon: 'sts', color: '#7c3aed', order: 5 },
  DaemonSet: { icon: 'ds', color: '#9333ea', order: 6 },
  CronJob: { icon: 'cronjob', color: '#db2777', order: 7 },
  Job: { icon: 'job', color: '#e11d48', order: 8 },
  ReplicaSet: { icon: 'rs', color: '#6366f1', order: 9 },
  Pod: { icon: 'pod', color: '#10b981', order: 10 },
  ConfigMap: { icon: 'cm', color: '#f59e0b', order: 11 },
  Secret: { icon: 'secret', color: '#f97316', order: 12 },
  PVC: { icon: 'pvc', color: '#14b8a6', order: 13 },
  ServiceAccount: { icon: 'sa', color: '#64748b', order: 14 }
}

const kindMeta = (kind: string): KindMeta => KIND_META[kind] || { icon: 'k8s', color: '#3b82f6', order: 99 }

const router = useRouter()
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)

const nsLoading = ref(false)
const namespaces = ref<string[]>([])
const selectedNamespace = ref('')
const loading = ref(false)
const rawNodes = ref<TopologyNodeVO[]>([])
const rawEdges = ref<TopologyEdgeVO[]>([])

const keyword = ref('')
const layoutDirection = ref<LayoutDirection>('TB')
const showEdgeLabels = ref(true)
const onlyConnected = ref(false)
const showReplicaSet = ref(false)
const activeKinds = ref<Set<string>>(new Set())
const isExpanded = ref(false)

const selectedNodeId = ref<string | null>(null)
const hoveredNodeId = ref<string | null>(null)
const selectedEdgeId = ref<string | null>(null)
const hoveredEdgeId = ref<string | null>(null)

// 画布平移与缩放
const canvasRef = ref<HTMLElement | null>(null)
const scale = ref(1)
const panX = ref(0)
const panY = ref(0)
const isPanning = ref(false)
let panStartX = 0
let panStartY = 0
let panBaseX = 0
let panBaseY = 0
let movedDuringPan = false

const toggleKind = (kind: string) => {
  const next = new Set(activeKinds.value)
  if (next.has(kind)) next.delete(kind)
  else next.add(kind)
  activeKinds.value = next
}

const clearKindFilter = () => {
  activeKinds.value = new Set()
}

// 计算文本视觉宽度与截断，保证 SVG 胶囊宽度贴合中英文内容
const visualLength = (text: string): number => {
  let len = 0
  for (const ch of text) {
    len += ch.charCodeAt(0) > 255 ? 1.85 : 1
  }
  return len
}

const truncateVisual = (text: string, maxUnits: number): string => {
  let len = 0
  let out = ''
  for (const ch of text) {
    const w = ch.charCodeAt(0) > 255 ? 1.85 : 1
    if (len + w > maxUnits) {
      return `${out}…`
    }
    out += ch
    len += w
  }
  return out
}

const formatEdgePill = (edge: TopologyEdgeVO) => {
  const raw = edge.label ? `${edge.relation} · ${edge.label}` : edge.relation
  const displayText = truncateVisual(raw, 26)
  const pillWidth = Math.max(46, Math.round(visualLength(displayText) * 6.2 + 18))
  return { displayText, pillWidth }
}

const edgeBindingPreview = (edge: TopologyEdgeVO): string => {
  if (edge.label) return edge.label
  if (edge.details && edge.details.length > 0) return edge.details[0]
  return ''
}

// 根据 showReplicaSet / onlyConnected / activeKinds 构造展示图
const visibleGraph = computed(() => {
  let nodes = [...rawNodes.value]
  let edges = [...rawEdges.value]

  // 默认折叠 ReplicaSet：将 Deployment -> RS -> Pod 自动桥接为 Deployment -> Pod，并合并绑定信息
  if (!showReplicaSet.value) {
    const rsMap = new Map(nodes.filter(n => n.kind === 'ReplicaSet').map(n => [n.id, n]))
    const rsIds = new Set(rsMap.keys())
    const incomingToRs = new Map<string, TopologyEdgeVO[]>()
    const outgoingFromRs = new Map<string, TopologyEdgeVO[]>()
    const keptEdges: TopologyEdgeVO[] = []

    for (const e of edges) {
      if (rsIds.has(e.target)) {
        const list = incomingToRs.get(e.target) || []
        list.push(e)
        incomingToRs.set(e.target, list)
      } else if (rsIds.has(e.source)) {
        const list = outgoingFromRs.get(e.source) || []
        list.push(e)
        outgoingFromRs.set(e.source, list)
      } else {
        keptEdges.push(e)
      }
    }

    const bridgedMap = new Map<string, TopologyEdgeVO>()
    keptEdges.forEach(e => bridgedMap.set(e.id, e))

    rsIds.forEach(rsId => {
      const rsNode = rsMap.get(rsId)
      const ins = incomingToRs.get(rsId) || []
      const outs = outgoingFromRs.get(rsId) || []
      for (const ie of ins) {
        for (const oe of outs) {
          if (ie.source === oe.target) continue
          const id = `${ie.source}->${oe.target}:管理`
          if (!bridgedMap.has(id)) {
            const details: string[] = []
            if (rsNode) details.push(`经由 ReplicaSet: ${rsNode.name}`)
            ;(ie.details || []).forEach(d => {
              if (!details.includes(d)) details.push(d)
            })
            ;(oe.details || []).forEach(d => {
              if (!details.includes(d)) details.push(d)
            })
            bridgedMap.set(id, {
              id,
              source: ie.source,
              target: oe.target,
              relation: '管理',
              label: oe.label || ie.label,
              details
            })
          }
        }
      }
    })

    nodes = nodes.filter(n => !rsIds.has(n.id))
    edges = Array.from(bridgedMap.values())
  }

  // 资源类型筛选
  if (activeKinds.value.size > 0) {
    const allowed = new Set(nodes.filter(n => activeKinds.value.has(n.kind)).map(n => n.id))
    nodes = nodes.filter(n => allowed.has(n.id))
    edges = edges.filter(e => allowed.has(e.source) && allowed.has(e.target))
  }

  // 隐藏孤立资源
  if (onlyConnected.value) {
    const connected = new Set<string>()
    edges.forEach(e => {
      connected.add(e.source)
      connected.add(e.target)
    })
    nodes = nodes.filter(n => connected.has(n.id))
  }

  const validIds = new Set(nodes.map(n => n.id))
  edges = edges.filter(e => validIds.has(e.source) && validIds.has(e.target))

  return { nodes, edges }
})

const kindStats = computed(() => {
  const counts = new Map<string, number>()
  const baseNodes = showReplicaSet.value ? rawNodes.value : rawNodes.value.filter(n => n.kind !== 'ReplicaSet')
  baseNodes.forEach(n => counts.set(n.kind, (counts.get(n.kind) || 0) + 1))
  return Array.from(counts.entries())
    .map(([kind, count]) => ({ kind, count, ...kindMeta(kind) }))
    .sort((a, b) => a.order - b.order)
})

interface PositionedNode extends TopologyNodeVO {
  x: number
  y: number
  laneKey: string
}

interface PositionedEdge extends TopologyEdgeVO {
  path: string
  labelX: number
  labelY: number
  displayText: string
  pillWidth: number
}

interface LaneBand {
  key: string
  label: string
  x: number
  y: number
  width: number
  height: number
  count: number
}

// 计算三次贝塞尔曲线上 t 处的坐标，用于精准放置连线标签
const cubicBezierPoint = (
  t: number,
  x0: number,
  y0: number,
  x1: number,
  y1: number,
  x2: number,
  y2: number,
  x3: number,
  y3: number
) => {
  const u = 1 - t
  const tt = t * t
  const uu = u * u
  const uuu = uu * u
  const ttt = tt * t
  return {
    x: uuu * x0 + 3 * uu * t * x1 + 3 * u * tt * x2 + ttt * x3,
    y: uuu * y0 + 3 * uu * t * y1 + 3 * u * tt * y2 + ttt * y3
  }
}

// 分层有向无环图布局（支持 TB 上下布局 / LR 左右布局 + 重心排序减少交叉）
const layout = computed(() => {
  const { nodes, edges } = visibleGraph.value
  if (!nodes.length) {
    return {
      width: 800,
      height: 420,
      nodes: [] as PositionedNode[],
      edges: [] as PositionedEdge[],
      lanes: [] as LaneBand[]
    }
  }

  const laneDefs = [
    { key: 'entry', label: '入口 / 伸缩' },
    { key: 'service', label: '服务发现' },
    { key: 'workload', label: '工作负载' },
    { key: 'replicaset', label: 'ReplicaSet' },
    { key: 'pod', label: '容器组 (Pod)' },
    { key: 'deps', label: '配置 / 存储 / 账户' }
  ]

  const resolveLane = (n: TopologyNodeVO): string => {
    if (n.kind === 'Ingress' || n.kind === 'HPA') return 'entry'
    if (n.kind === 'Service') return 'service'
    if (n.kind === 'ReplicaSet') return 'replicaset'
    if (n.kind === 'Pod') return 'pod'
    if (n.layer === 4) return 'deps'
    return 'workload'
  }

  const grouped = new Map<string, TopologyNodeVO[]>()
  for (const n of nodes) {
    const lk = resolveLane(n)
    const list = grouped.get(lk) || []
    list.push(n)
    grouped.set(lk, list)
  }

  const activeLanes = laneDefs.filter(l => (grouped.get(l.key)?.length || 0) > 0)

  // 建立邻接表用于重心法排序
  const adjIn = new Map<string, string[]>()
  const adjOut = new Map<string, string[]>()
  edges.forEach(e => {
    const out = adjOut.get(e.source) || []
    out.push(e.target)
    adjOut.set(e.source, out)
    const ins = adjIn.get(e.target) || []
    ins.push(e.source)
    adjIn.set(e.target, ins)
  })

  // 初始按类型与名称排序
  activeLanes.forEach(l => {
    grouped.get(l.key)!.sort((a, b) => {
      const d = kindMeta(a.kind).order - kindMeta(b.kind).order
      return d !== 0 ? d : a.name.localeCompare(b.name)
    })
  })

  const orderIndex = new Map<string, number>()
  const refreshOrder = () => {
    activeLanes.forEach(l => {
      grouped.get(l.key)!.forEach((n, idx) => orderIndex.set(n.id, idx))
    })
  }
  refreshOrder()

  const sortLaneByNeighbors = (laneKey: string, neighborMap: Map<string, string[]>) => {
    const list = grouped.get(laneKey)!
    list.sort((a, b) => {
      const na = neighborMap.get(a.id) || []
      const nb = neighborMap.get(b.id) || []
      const avg = (ids: string[], fallback: number) => {
        const vals = ids.map(id => orderIndex.get(id)).filter((v): v is number => v !== undefined)
        if (!vals.length) return fallback + 1000
        return vals.reduce((s, v) => s + v, 0) / vals.length
      }
      const ba = avg(na, orderIndex.get(a.id) ?? 0)
      const bb = avg(nb, orderIndex.get(b.id) ?? 0)
      if (Math.abs(ba - bb) > 0.01) return ba - bb
      return a.name.localeCompare(b.name)
    })
    refreshOrder()
  }

  for (let pass = 0; pass < 2; pass++) {
    for (let i = 1; i < activeLanes.length; i++) {
      sortLaneByNeighbors(activeLanes[i].key, adjIn)
    }
    for (let i = activeLanes.length - 2; i >= 0; i--) {
      sortLaneByNeighbors(activeLanes[i].key, adjOut)
    }
  }

  const positionedNodes: PositionedNode[] = []
  const nodePosMap = new Map<string, PositionedNode>()
  const lanes: LaneBand[] = []
  let totalWidth = 800
  let totalHeight = 460

  if (layoutDirection.value === 'TB') {
    // ================= 从上到下（纵向分层） =================
    const maxCols = Math.max(...activeLanes.map(l => grouped.get(l.key)!.length), 1)
    const contentWidth = Math.max(520, maxCols * NODE_W + Math.max(0, maxCols - 1) * TB_COL_GAP)
    const laneBandWidth = TB_HEADER_W + TB_LANE_PAD_X * 2 + contentWidth
    const laneBandHeight = NODE_H + TB_LANE_PAD_Y * 2
    totalWidth = TB_PAD_X * 2 + laneBandWidth
    totalHeight = Math.max(
      460,
      TB_PAD_Y * 2 + activeLanes.length * laneBandHeight + Math.max(0, activeLanes.length - 1) * TB_ROW_GAP
    )

    activeLanes.forEach((lane, rowIdx) => {
      const list = grouped.get(lane.key)!
      const bandY = TB_PAD_Y + rowIdx * (laneBandHeight + TB_ROW_GAP)
      const rowWidth = list.length * NODE_W + Math.max(0, list.length - 1) * TB_COL_GAP
      const startX = TB_PAD_X + TB_HEADER_W + TB_LANE_PAD_X + Math.max(0, (contentWidth - rowWidth) / 2)
      const nodeY = bandY + TB_LANE_PAD_Y

      lanes.push({
        key: lane.key,
        label: lane.label,
        x: TB_PAD_X,
        y: bandY,
        width: laneBandWidth,
        height: laneBandHeight,
        count: list.length
      })

      list.forEach((n, colIdx) => {
        const pn: PositionedNode = {
          ...n,
          x: startX + colIdx * (NODE_W + TB_COL_GAP),
          y: nodeY,
          laneKey: lane.key
        }
        positionedNodes.push(pn)
        nodePosMap.set(pn.id, pn)
      })
    })
  } else {
    // ================= 从左到右（横向分层） =================
    const maxRows = Math.max(...activeLanes.map(l => grouped.get(l.key)!.length), 1)
    const contentHeight = maxRows * NODE_H + Math.max(0, maxRows - 1) * LR_ROW_GAP
    totalHeight = Math.max(460, LR_LANE_TOP + contentHeight + LR_PAD_BOTTOM)
    totalWidth = LR_PAD_X * 2 + activeLanes.length * NODE_W + Math.max(0, activeLanes.length - 1) * LR_COL_GAP

    activeLanes.forEach((lane, colIdx) => {
      const list = grouped.get(lane.key)!
      const colX = LR_PAD_X + colIdx * (NODE_W + LR_COL_GAP)
      const colHeight = list.length * NODE_H + Math.max(0, list.length - 1) * LR_ROW_GAP
      const startY = LR_LANE_TOP + Math.max(0, (contentHeight - colHeight) / 2)

      lanes.push({
        key: lane.key,
        label: lane.label,
        x: colX - 18,
        y: 14,
        width: NODE_W + 36,
        height: Math.max(totalHeight - 28, 220),
        count: list.length
      })

      list.forEach((n, rowIdx) => {
        const pn: PositionedNode = {
          ...n,
          x: colX,
          y: startY + rowIdx * (NODE_H + LR_ROW_GAP),
          laneKey: lane.key
        }
        positionedNodes.push(pn)
        nodePosMap.set(pn.id, pn)
      })
    })
  }

  // 统计同一源节点扇出计数，用于轻微错开连线标签 t 值，减少重叠
  const laneIdxMap = new Map<string, number>()
  activeLanes.forEach((l, idx) => laneIdxMap.set(l.key, idx))

  const outIndexMap = new Map<string, number>()
  const outCountMap = new Map<string, number>()
  edges.forEach(e => outCountMap.set(e.source, (outCountMap.get(e.source) || 0) + 1))

  const positionedEdges: PositionedEdge[] = []
  for (const e of edges) {
    const src = nodePosMap.get(e.source)
    const dst = nodePosMap.get(e.target)
    if (!src || !dst) continue

    const siblingIdx = outIndexMap.get(e.source) || 0
    outIndexMap.set(e.source, siblingIdx + 1)
    const siblingTotal = outCountMap.get(e.source) || 1
    const laneSpan = Math.abs((laneIdxMap.get(dst.laneKey) ?? 0) - (laneIdxMap.get(src.laneKey) ?? 0))
    // 跨多层连线（如 Ingress -> TLS Secret）将标签落在源泳道紧邻的空白间隙中，避免被中间层卡片遮挡
    const labelT =
      laneSpan >= 2
        ? 0.2 + ((siblingIdx % 3) * 0.04)
        : siblingTotal > 1
          ? 0.42 + ((siblingIdx % 3) * 0.08)
          : 0.5

    let path = ''
    let labelX = 0
    let labelY = 0

    if (layoutDirection.value === 'TB') {
      if (src.y === dst.y) {
        // 同层关系（如 CronJob -> Job）：走卡片上方柔和弧线
        const sx = src.x + NODE_W / 2
        const sy = src.y
        const tx = dst.x + NODE_W / 2
        const ty = dst.y
        const arch = 44
        path = `M ${sx} ${sy} C ${sx} ${sy - arch}, ${tx} ${ty - arch}, ${tx} ${ty}`
        const pt = cubicBezierPoint(0.5, sx, sy, sx, sy - arch, tx, ty - arch, tx, ty)
        labelX = pt.x
        labelY = pt.y
      } else {
        const downward = src.y < dst.y
        const sx = src.x + NODE_W / 2
        const sy = downward ? src.y + NODE_H : src.y
        const tx = dst.x + NODE_W / 2
        const ty = downward ? dst.y : dst.y + NODE_H
        const dy = Math.abs(ty - sy)
        const curve = Math.max(40, dy * 0.48)
        const c1y = downward ? sy + curve : sy - curve
        const c2y = downward ? ty - curve : ty + curve
        path = `M ${sx} ${sy} C ${sx} ${c1y}, ${tx} ${c2y}, ${tx} ${ty}`
        const pt = cubicBezierPoint(labelT, sx, sy, sx, c1y, tx, c2y, tx, ty)
        labelX = pt.x
        labelY = pt.y
      }
    } else {
      if (src.x === dst.x) {
        // 同列关系：走右侧柔和弧线
        const sx = src.x + NODE_W
        const sy = src.y + NODE_H / 2
        const tx = dst.x + NODE_W
        const ty = dst.y + NODE_H / 2
        const bulge = 54
        path = `M ${sx} ${sy} C ${sx + bulge} ${sy}, ${tx + bulge} ${ty}, ${tx} ${ty}`
        const pt = cubicBezierPoint(0.5, sx, sy, sx + bulge, sy, tx + bulge, ty, tx, ty)
        labelX = pt.x
        labelY = pt.y
      } else {
        const rightward = src.x < dst.x
        const sx = rightward ? src.x + NODE_W : src.x
        const sy = src.y + NODE_H / 2
        const tx = rightward ? dst.x : dst.x + NODE_W
        const ty = dst.y + NODE_H / 2
        const dx = Math.abs(tx - sx)
        const curve = Math.max(48, dx * 0.45)
        const c1x = rightward ? sx + curve : sx - curve
        const c2x = rightward ? tx - curve : tx + curve
        path = `M ${sx} ${sy} C ${c1x} ${sy}, ${c2x} ${ty}, ${tx} ${ty}`
        const pt = cubicBezierPoint(labelT, sx, sy, c1x, sy, c2x, ty, tx, ty)
        labelX = pt.x
        labelY = pt.y
      }
    }

    const { displayText, pillWidth } = formatEdgePill(e)
    positionedEdges.push({ ...e, path, labelX, labelY, displayText, pillWidth })
  }

  return {
    width: totalWidth,
    height: totalHeight,
    nodes: positionedNodes,
    edges: positionedEdges,
    lanes
  }
})

// 聚焦对象（支持节点上下游链路高亮 或 单条连线端点高亮）
const activeEdgeFocusId = computed(() => hoveredEdgeId.value || (hoveredNodeId.value ? null : selectedEdgeId.value))
const focusNodeId = computed(() => (activeEdgeFocusId.value ? null : hoveredNodeId.value || selectedNodeId.value))
const hasFocus = computed(() => Boolean(activeEdgeFocusId.value || focusNodeId.value))

const focusChain = computed(() => {
  const nodeIds = new Set<string>()
  const edgeIds = new Set<string>()
  const edges = visibleGraph.value.edges

  if (activeEdgeFocusId.value) {
    const targetEdge = edges.find(e => e.id === activeEdgeFocusId.value)
    if (targetEdge) {
      edgeIds.add(targetEdge.id)
      nodeIds.add(targetEdge.source)
      nodeIds.add(targetEdge.target)
    }
    return { nodeIds, edgeIds }
  }

  const root = focusNodeId.value
  if (!root) return { nodeIds, edgeIds }

  nodeIds.add(root)

  // 向上游回溯
  const upQueue = [root]
  const upVisited = new Set([root])
  while (upQueue.length) {
    const cur = upQueue.shift()!
    for (const e of edges) {
      if (e.target === cur) {
        edgeIds.add(e.id)
        nodeIds.add(e.source)
        if (!upVisited.has(e.source)) {
          upVisited.add(e.source)
          upQueue.push(e.source)
        }
      }
    }
  }

  // 向下游扩散
  const downQueue = [root]
  const downVisited = new Set([root])
  while (downQueue.length) {
    const cur = downQueue.shift()!
    for (const e of edges) {
      if (e.source === cur) {
        edgeIds.add(e.id)
        nodeIds.add(e.target)
        if (!downVisited.has(e.target)) {
          downVisited.add(e.target)
          downQueue.push(e.target)
        }
      }
    }
  }

  return { nodeIds, edgeIds }
})

const isNodeHighlighted = (id: string) => hasFocus.value && focusChain.value.nodeIds.has(id)
const isEdgeHighlighted = (edge: TopologyEdgeVO) => hasFocus.value && focusChain.value.edgeIds.has(edge.id)

const shouldShowEdgeLabel = (edge: PositionedEdge) => {
  if (!edge.displayText) return false
  if (isEdgeHighlighted(edge) || selectedEdgeId.value === edge.id || hoveredEdgeId.value === edge.id) return true
  return showEdgeLabels.value
}

const isSearchHit = (node: TopologyNodeVO) => {
  const q = keyword.value.trim().toLowerCase()
  if (!q) return false
  return node.name.toLowerCase().includes(q) || node.kind.toLowerCase().includes(q)
}

const nodeLookupMap = computed(() => new Map(visibleGraph.value.nodes.map(n => [n.id, n])))

const selectedNode = computed(() =>
  selectedNodeId.value ? nodeLookupMap.value.get(selectedNodeId.value) || null : null
)

const buildEdgeInfo = (edgeId: string | null) => {
  if (!edgeId) return null
  const edge = visibleGraph.value.edges.find(e => e.id === edgeId)
  if (!edge) return null
  const details = edge.details && edge.details.length > 0 ? edge.details : edge.label ? [edge.label] : []
  return {
    edge,
    source: nodeLookupMap.value.get(edge.source) || null,
    target: nodeLookupMap.value.get(edge.target) || null,
    details
  }
}

const selectedEdgeInfo = computed(() => buildEdgeInfo(selectedEdgeId.value))
const hoveredEdgeInfo = computed(() => buildEdgeInfo(hoveredEdgeId.value))

const selectedRelations = computed(() => {
  const id = selectedNodeId.value
  if (!id) {
    return {
      incoming: [] as { edge: TopologyEdgeVO; node: TopologyNodeVO }[],
      outgoing: [] as { edge: TopologyEdgeVO; node: TopologyNodeVO }[]
    }
  }
  const map = nodeLookupMap.value
  const incoming: { edge: TopologyEdgeVO; node: TopologyNodeVO }[] = []
  const outgoing: { edge: TopologyEdgeVO; node: TopologyNodeVO }[] = []
  for (const e of visibleGraph.value.edges) {
    if (e.target === id && map.has(e.source)) incoming.push({ edge: e, node: map.get(e.source)! })
    if (e.source === id && map.has(e.target)) outgoing.push({ edge: e, node: map.get(e.target)! })
  }
  return { incoming, outgoing }
})

const tagTypeOf = (t: TopologyStatusType) => t

const selectNode = (id: string) => {
  selectedEdgeId.value = null
  selectedNodeId.value = selectedNodeId.value === id ? null : id
}

const selectEdge = (id: string) => {
  selectedNodeId.value = null
  selectedEdgeId.value = selectedEdgeId.value === id ? null : id
}

const focusNodeById = (id: string) => {
  selectedEdgeId.value = null
  selectedNodeId.value = id
  const target = layout.value.nodes.find(n => n.id === id)
  const el = canvasRef.value
  if (!target || !el) return
  const vw = el.clientWidth
  const vh = el.clientHeight
  panX.value = vw / 2 - (target.x + NODE_W / 2) * scale.value
  panY.value = vh / 2 - (target.y + NODE_H / 2) * scale.value
}

const focusFirstMatch = () => {
  const hit = layout.value.nodes.find(n => isSearchHit(n))
  if (hit) focusNodeById(hit.id)
}

const fitView = () => {
  const el = canvasRef.value
  if (!el || !layout.value.nodes.length) return
  const vw = el.clientWidth
  const vh = el.clientHeight
  if (vw <= 0 || vh <= 0) return
  const pad = 28
  const fitScale = Math.min(
    1,
    Math.max(0.38, Math.min((vw - pad * 2) / layout.value.width, (vh - pad * 2) / layout.value.height))
  )
  scale.value = Number(fitScale.toFixed(2))
  panX.value = Math.round((vw - layout.value.width * scale.value) / 2)
  panY.value = Math.round((vh - layout.value.height * scale.value) / 2)
}

const zoomBy = (delta: number) => {
  const el = canvasRef.value
  const oldScale = scale.value
  const nextScale = Math.min(1.8, Math.max(0.3, Number((oldScale + delta).toFixed(2))))
  if (!el) {
    scale.value = nextScale
    return
  }
  const cx = el.clientWidth / 2
  const cy = el.clientHeight / 2
  panX.value = cx - ((cx - panX.value) / oldScale) * nextScale
  panY.value = cy - ((cy - panY.value) / oldScale) * nextScale
  scale.value = nextScale
}

const onCanvasWheel = (e: WheelEvent) => {
  const el = canvasRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  const mx = e.clientX - rect.left
  const my = e.clientY - rect.top
  const oldScale = scale.value
  const factor = e.deltaY < 0 ? 1.08 : 0.92
  const nextScale = Math.min(1.8, Math.max(0.3, Number((oldScale * factor).toFixed(3))))
  panX.value = mx - ((mx - panX.value) / oldScale) * nextScale
  panY.value = my - ((my - panY.value) / oldScale) * nextScale
  scale.value = nextScale
}

const onCanvasPointerDown = (e: PointerEvent) => {
  if (e.button !== 0) return
  isPanning.value = true
  movedDuringPan = false
  panStartX = e.clientX
  panStartY = e.clientY
  panBaseX = panX.value
  panBaseY = panY.value
  canvasRef.value?.setPointerCapture?.(e.pointerId)
}

const onCanvasPointerMove = (e: PointerEvent) => {
  if (!isPanning.value) return
  const dx = e.clientX - panStartX
  const dy = e.clientY - panStartY
  if (Math.abs(dx) > 4 || Math.abs(dy) > 4) movedDuringPan = true
  panX.value = panBaseX + dx
  panY.value = panBaseY + dy
}

const onCanvasPointerUp = (e: PointerEvent) => {
  if (!isPanning.value) return
  isPanning.value = false
  canvasRef.value?.releasePointerCapture?.(e.pointerId)
}

const onCanvasClick = () => {
  if (!movedDuringPan) {
    selectedNodeId.value = null
    selectedEdgeId.value = null
  }
}

const toggleExpand = (open: boolean) => {
  if (isExpanded.value === open) return
  isExpanded.value = open
  document.body.style.overflow = open ? 'hidden' : ''
  nextTick(() => {
    fitView()
  })
}

const onKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && isExpanded.value) {
    toggleExpand(false)
  }
}

// 资源详情跳转映射
const detailRouteOf = (node: TopologyNodeVO) => {
  const cid = clusterId.value
  const q = { namespace: node.namespace, name: node.name }
  switch (node.kind) {
    case 'Deployment':
      return { path: `/cluster/${cid}/workload/deployment/detail`, query: q }
    case 'StatefulSet':
      return { path: `/cluster/${cid}/workload/statefulset/detail`, query: q }
    case 'DaemonSet':
      return { path: `/cluster/${cid}/workload/daemonset/detail`, query: q }
    case 'Pod':
      return { path: `/cluster/${cid}/pod/detail`, query: q }
    case 'Service':
      return { path: `/cluster/${cid}/service/service/detail`, query: q }
    case 'Ingress':
      return { path: `/cluster/${cid}/service/ingress/detail`, query: q }
    case 'Job':
      return { path: `/cluster/${cid}/job/job/detail`, query: q }
    case 'CronJob':
      return { path: `/cluster/${cid}/job/cronjob/detail`, query: q }
    case 'ConfigMap':
      return { path: `/cluster/${cid}/config/configmap/detail`, query: q }
    case 'Secret':
      return { path: `/cluster/${cid}/config/secret/detail`, query: q }
    case 'ServiceAccount':
      return { path: `/cluster/${cid}/config/serviceaccount/detail`, query: q }
    case 'PVC':
      return { path: `/cluster/${cid}/storage/pvc` }
    case 'HPA':
      return { path: `/cluster/${cid}/workload/hpa` }
    case 'ReplicaSet':
      return { path: `/cluster/${cid}/workload/replicaset` }
    default:
      return null
  }
}

const canOpenDetail = (node: TopologyNodeVO) => node.status !== '未找到' && Boolean(detailRouteOf(node))

const openResourceDetail = (node: TopologyNodeVO) => {
  const target = detailRouteOf(node)
  if (!target) return
  if (isExpanded.value) toggleExpand(false)
  router.push(target)
}

const fetchNamespaces = async () => {
  if (!clusterId.value) return
  nsLoading.value = true
  try {
    const resp = await namespaceApi.getList({ clusterId: clusterId.value, page: 1, pageSize: 1000 })
    const items = resp.data?.items || []
    namespaces.value = items.map(n => n.name)
    if (!selectedNamespace.value || !namespaces.value.includes(selectedNamespace.value)) {
      selectedNamespace.value = namespaces.value.includes('default') ? 'default' : namespaces.value[0] || ''
    }
    if (selectedNamespace.value) {
      await fetchTopology()
    } else {
      rawNodes.value = []
      rawEdges.value = []
    }
  } catch {
    namespaces.value = []
    selectedNamespace.value = ''
    rawNodes.value = []
    rawEdges.value = []
  } finally {
    nsLoading.value = false
  }
}

const fetchTopology = async () => {
  if (!clusterId.value || !selectedNamespace.value) return
  loading.value = true
  selectedNodeId.value = null
  selectedEdgeId.value = null
  try {
    const resp = await topologyApi.getNamespaceTopology({
      clusterId: clusterId.value,
      namespace: selectedNamespace.value
    })
    rawNodes.value = resp.data?.nodes || []
    rawEdges.value = resp.data?.edges || []
    activeKinds.value = new Set()
    await nextTick()
    fitView()
  } catch {
    rawNodes.value = []
    rawEdges.value = []
  } finally {
    loading.value = false
  }
}

watch([showReplicaSet, onlyConnected, activeKinds, layoutDirection], () => {
  if (selectedNodeId.value && !visibleGraph.value.nodes.some(n => n.id === selectedNodeId.value)) {
    selectedNodeId.value = null
  }
  if (selectedEdgeId.value && !visibleGraph.value.edges.some(e => e.id === selectedEdgeId.value)) {
    selectedEdgeId.value = null
  }
  nextTick(() => fitView())
})

watch(clusterId, id => {
  selectedNamespace.value = ''
  selectedNodeId.value = null
  selectedEdgeId.value = null
  if (id) fetchNamespaces()
})

onMounted(() => {
  window.addEventListener('keydown', onKeydown)
  if (clusterId.value) fetchNamespaces()
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKeydown)
  document.body.style.overflow = ''
})
</script>

<style scoped>
.topology-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: calc(100vh - 138px);
  min-height: 620px;
}

.toolbar-card {
  border-radius: 14px;
  border: 1px solid rgba(226, 232, 240, 0.9);
  flex-shrink: 0;
}

.toolbar-card :deep(.el-card__body) {
  padding: 14px 18px;
}

.toolbar-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.control-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.control-label {
  font-size: 13px;
  font-weight: 600;
  color: #334155;
  white-space: nowrap;
}

.zoom-controls {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.zoom-label {
  min-width: 44px;
  text-align: center;
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  font-variant-numeric: tabular-nums;
}

.kind-filters {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed #e2e8f0;
}

.kind-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  border: 1px solid #dbe4f2;
  background: #f8fafc;
  color: #334155;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.18s ease;
}

.kind-chip:hover {
  border-color: #93c5fd;
  background: #eff6ff;
}

.kind-chip.is-active {
  border-color: var(--chip-color, #2563eb);
  background: #eff6ff;
  color: #1d4ed8;
  font-weight: 600;
}

.chip-count {
  padding: 0 6px;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.18);
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}

.kind-chip.is-active .chip-count {
  background: rgba(37, 99, 235, 0.14);
  color: #1d4ed8;
}

/* ==================== 弹窗放大容器（兼容内嵌与 Teleport 全屏大框） ==================== */
.topo-modal-shell {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.topo-modal-shell.is-expanded {
  position: fixed;
  inset: 0;
  z-index: 2100;
  background: rgba(15, 23, 42, 0.56);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.topo-modal-dialog {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.topo-modal-dialog.is-expanded {
  flex: none;
  width: min(95vw, 1840px);
  height: 92vh;
  border-radius: 18px;
  overflow: hidden;
  background: #ffffff;
  border: 1px solid rgba(226, 232, 240, 0.95);
  box-shadow: 0 28px 72px rgba(15, 23, 42, 0.36);
}

.modal-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  padding: 12px 18px;
  background: #ffffff;
  border-bottom: 1px solid #e2e8f0;
  flex-shrink: 0;
}

.modal-title-group {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.modal-title {
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
}

.modal-stat {
  font-size: 12px;
  color: #64748b;
}

.modal-actions {
  display: inline-flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

/* ==================== 画布 + 详情面板 ==================== */
.workspace-card {
  flex: 1;
  min-height: 0;
  display: flex;
  border-radius: 16px;
  overflow: hidden;
  background: #ffffff;
  border: 1px solid rgba(226, 232, 240, 0.9);
  box-shadow: 0 6px 24px rgba(15, 23, 42, 0.05);
}

.workspace-card.is-expanded {
  border-radius: 0;
  border: none;
  box-shadow: none;
}

.topo-canvas {
  position: relative;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  cursor: grab;
  user-select: none;
  background-color: #f8fbff;
  background-image: radial-gradient(#cbd5e1 1px, transparent 1px);
  background-size: 20px 20px;
}

.topo-canvas.is-panning {
  cursor: grabbing;
}

.canvas-float-actions {
  position: absolute;
  top: 14px;
  right: 16px;
  z-index: 12;
}

.float-expand-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 32px;
  padding: 0 12px;
  border-radius: 999px;
  border: 1px solid #cbd5e1;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(8px);
  color: #1e293b;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(15, 23, 42, 0.08);
  transition: all 0.18s ease;
}

.float-expand-btn:hover {
  border-color: #2563eb;
  color: #2563eb;
  background: #eff6ff;
}

/* 悬停连线绑定详情浮层 */
.edge-hover-banner {
  position: absolute;
  top: 14px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 15;
  max-width: min(760px, calc(100% - 160px));
  padding: 10px 14px;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.92);
  backdrop-filter: blur(10px);
  color: #f8fafc;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.25);
  pointer-events: none;
}

.edge-hover-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  font-size: 12px;
}

.edge-hover-endpoints {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #e2e8f0;
  word-break: break-all;
}

.arrow-sep {
  color: #60a5fa;
  font-weight: 700;
}

.edge-hover-details {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.edge-hover-pill {
  padding: 2px 8px;
  border-radius: 6px;
  background: rgba(59, 130, 246, 0.22);
  border: 1px solid rgba(96, 165, 250, 0.35);
  color: #dbeafe;
  font-size: 11px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}

.canvas-empty {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.topo-stage {
  position: relative;
  transform-origin: 0 0;
  will-change: transform;
}

/* 泳道背景 */
.lane-band {
  position: absolute;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.58);
  border: 1px dashed rgba(148, 163, 184, 0.38);
  pointer-events: none;
}

.lane-header.is-vertical {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid rgba(226, 232, 240, 0.75);
}

.lane-header.is-horizontal {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 148px;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: flex-start;
  gap: 6px;
  border-right: 1px solid rgba(226, 232, 240, 0.8);
  background: rgba(248, 250, 252, 0.72);
  border-radius: 14px 0 0 14px;
}

.lane-title {
  font-size: 12px;
  font-weight: 700;
  color: #475569;
  letter-spacing: 0.3px;
}

.lane-count {
  padding: 1px 7px;
  border-radius: 999px;
  background: #e2e8f0;
  color: #334155;
  font-size: 11px;
  font-weight: 600;
}

.edges-svg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  overflow: visible;
}

.topo-edge-hit {
  fill: none;
  stroke: transparent;
  stroke-width: 16;
  pointer-events: stroke;
  cursor: pointer;
}

.topo-edge {
  fill: none;
  stroke: #94a3b8;
  stroke-width: 1.7;
  opacity: 0.75;
  pointer-events: none;
  transition: stroke 0.2s ease, opacity 0.2s ease, stroke-width 0.2s ease;
}

.topo-edge.is-highlighted {
  stroke: #2563eb;
  stroke-width: 2.5;
  opacity: 1;
  stroke-dasharray: 6 4;
  animation: edge-flow 1.1s linear infinite;
}

.topo-edge.is-selected {
  stroke: #1d4ed8;
  stroke-width: 3;
}

.topo-edge.is-dimmed {
  opacity: 0.14;
}

@keyframes edge-flow {
  from {
    stroke-dashoffset: 20;
  }
  to {
    stroke-dashoffset: 0;
  }
}

.edge-label-group {
  pointer-events: all;
  cursor: pointer;
  transition: opacity 0.2s ease;
}

.edge-label-group.is-dimmed {
  opacity: 0.12;
}

.edge-label-bg {
  fill: rgba(255, 255, 255, 0.96);
  stroke: #cbd5e1;
  stroke-width: 1;
  transition: fill 0.18s ease, stroke 0.18s ease;
}

.edge-label-group:hover .edge-label-bg,
.edge-label-group.is-highlighted .edge-label-bg {
  fill: #eff6ff;
  stroke: #3b82f6;
}

.edge-label-group.is-selected .edge-label-bg {
  fill: #2563eb;
  stroke: #1d4ed8;
}

.edge-label-text {
  font-size: 10.5px;
  font-weight: 600;
  fill: #334155;
  user-select: none;
}

.edge-label-group:hover .edge-label-text,
.edge-label-group.is-highlighted .edge-label-text {
  fill: #1d4ed8;
}

.edge-label-group.is-selected .edge-label-text {
  fill: #ffffff;
}

/* 节点卡片 */
.topo-node {
  position: absolute;
  display: flex;
  border-radius: 12px;
  background: #ffffff;
  border: 1.5px solid #e2e8f0;
  box-shadow: 0 4px 14px rgba(15, 23, 42, 0.06);
  overflow: hidden;
  cursor: pointer;
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease,
    border-color 0.18s ease,
    opacity 0.2s ease;
}

.topo-node:hover {
  transform: translateY(-2px);
  border-color: var(--kind-accent, #3b82f6);
  box-shadow: 0 10px 24px rgba(37, 99, 235, 0.15);
}

.topo-node.is-highlighted {
  border-color: var(--kind-accent, #2563eb);
  box-shadow: 0 8px 22px rgba(37, 99, 235, 0.18);
}

.topo-node.is-selected {
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.22), 0 10px 26px rgba(37, 99, 235, 0.22);
}

.topo-node.is-search-hit {
  border-color: #f59e0b;
  box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.25);
}

.topo-node.is-dimmed {
  opacity: 0.26;
}

.topo-node.status-danger {
  border-color: rgba(239, 68, 68, 0.55);
}

.node-accent-bar {
  width: 4px;
  flex-shrink: 0;
  background: var(--kind-accent, #3b82f6);
}

.node-body {
  flex: 1;
  min-width: 0;
  padding: 9px 11px 8px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.node-head {
  display: flex;
  align-items: center;
  gap: 9px;
  min-width: 0;
}

.node-title-col {
  flex: 1;
  min-width: 0;
}

.node-name {
  font-size: 13px;
  font-weight: 700;
  color: #0f172a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.3;
}

.node-kind-row {
  display: flex;
  align-items: center;
  gap: 5px;
  margin-top: 2px;
  font-size: 11px;
  color: #64748b;
}

.node-kind {
  font-weight: 600;
  color: #475569;
}

.node-status-dot,
.dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.dot-success {
  background: #22c55e;
}

.dot-warning {
  background: #f59e0b;
}

.dot-danger {
  background: #ef4444;
}

.dot-info {
  background: #94a3b8;
}

.node-status-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-summary {
  font-size: 11px;
  color: #64748b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding-top: 4px;
  border-top: 1px solid #f1f5f9;
}

/* 左下角图例 */
.canvas-legend {
  position: absolute;
  left: 16px;
  bottom: 14px;
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 7px 14px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(8px);
  border: 1px solid #e2e8f0;
  box-shadow: 0 4px 14px rgba(15, 23, 42, 0.06);
  font-size: 12px;
  color: #475569;
  pointer-events: none;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.legend-sep {
  color: #cbd5e1;
}

.legend-tip {
  color: #94a3b8;
}

/* 右侧详情侧栏 */
.inspector-panel {
  width: 336px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  border-left: 1px solid #e2e8f0;
  background: #ffffff;
}

.inspector-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
  padding: 16px 16px 14px;
  border-bottom: 1px solid #f1f5f9;
}

.inspector-title {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.edge-inspector-icon {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  background: #eff6ff;
  color: #2563eb;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.inspector-name-wrap {
  min-width: 0;
  flex: 1;
}

.inspector-name {
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.inspector-sub {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
  flex-wrap: wrap;
}

.edge-sub-label {
  font-size: 12px;
  color: #475569;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}

.inspector-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  font-weight: 700;
  color: #334155;
  margin-bottom: 10px;
}

.section-count {
  padding: 1px 7px;
  border-radius: 999px;
  background: #eff6ff;
  color: #2563eb;
  font-size: 11px;
}

.kv-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.kv-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 12px;
}

.kv-key {
  color: #64748b;
  flex-shrink: 0;
}

.kv-val {
  color: #0f172a;
  font-weight: 500;
  text-align: right;
  word-break: break-all;
}

.kv-val.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}

.rel-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.rel-item {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 8px 10px;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
  background: #f8fafc;
  text-align: left;
  cursor: pointer;
  transition: all 0.18s ease;
}

.rel-item:hover {
  border-color: #93c5fd;
  background: #eff6ff;
}

.rel-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.rel-name {
  font-size: 12px;
  font-weight: 600;
  color: #0f172a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rel-meta {
  font-size: 11px;
  color: #64748b;
}

.rel-binding {
  display: inline-block;
  margin-top: 2px;
  padding: 1px 6px;
  border-radius: 5px;
  background: #eff6ff;
  color: #1d4ed8;
  font-size: 11px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  word-break: break-all;
}

.rel-arrow {
  color: #94a3b8;
  flex-shrink: 0;
}

.rel-empty {
  font-size: 12px;
  color: #94a3b8;
  padding: 6px 0;
}

.edge-endpoint-cards {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.endpoint-flow-arrow {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 600;
  color: #2563eb;
  position: relative;
}

.endpoint-flow-arrow span {
  padding: 2px 10px;
  border-radius: 999px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
}

.binding-detail-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.binding-detail-item {
  padding: 8px 10px;
  border-radius: 8px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  color: #1e293b;
  font-size: 12px;
  line-height: 1.5;
  word-break: break-all;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}

.inspector-footer {
  padding: 12px 16px;
  border-top: 1px solid #f1f5f9;
}

@media (max-width: 960px) {
  .workspace-card {
    flex-direction: column;
  }

  .inspector-panel {
    width: 100%;
    max-height: 280px;
    border-left: none;
    border-top: 1px solid #e2e8f0;
  }

  .legend-tip,
  .legend-sep {
    display: none;
  }
}
</style>
