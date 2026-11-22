<template>
  <div class="job-detail detail-page">
    <PageHeader
      :title="node?.server.hostname || hostname || '节点详情'"
      :subtitle="node ? `${node.roles.map(roleText).join('、')} · ${node.server.ip || '-'}` : '节点详情'"
      icon="node"
      back
      @back="goBack"
    >
      <template v-if="node" #meta>
        <el-tag :type="nodeStatusTag(node.server.status)" effect="dark">{{ nodeStatusText(node.server.status) }}</el-tag>
        <el-tag v-if="node.server.unSchedule" type="warning" effect="light">已停止调度</el-tag>
      </template>
      <template v-if="node">
        <el-button :loading="loading" @click="fetchDetail()">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button @click="toggleSchedule">
          <el-icon><component :is="node.server.unSchedule ? VideoPlay : VideoPause" /></el-icon>
          {{ node.server.unSchedule ? '启用调度' : '停止调度' }}
        </el-button>
        <el-button type="warning" plain @click="drainVisible = true">
          <el-icon><Promotion /></el-icon>
          排空节点
        </el-button>
        <el-button @click="labelsVisible = true">
          <el-icon><PriceTag /></el-icon>
          管理标签
        </el-button>
        <el-button @click="taintsVisible = true">
          <el-icon><WarningFilled /></el-icon>
          管理污点
        </el-button>
        <el-button @click="openYaml">
          <el-icon><Document /></el-icon>
          编辑YAML
        </el-button>
      </template>
    </PageHeader>

    <div v-if="loading && !node" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="node" class="detail-content">
      <el-card class="info-card jd-attr-card">
        <template #header>
          <div class="card-header" @click="attrOpen = !attrOpen">
            <span>属性</span>
            <el-icon :class="['jd-collapse', { 'is-collapsed': !attrOpen }]"><ArrowDown /></el-icon>
          </div>
        </template>
        <el-collapse-transition>
          <div v-show="attrOpen" class="jd-attr-grid">
            <div class="jd-attr"><label>集群</label><span>{{ clusterName || clusterId }}</span></div>
            <div class="jd-attr"><label>状态</label><span>{{ nodeStatusText(node.server.status) }}</span></div>
            <div class="jd-attr"><label>角色</label><span>{{ node.roles.map(roleText).join('、') }}</span></div>
            <div class="jd-attr"><label>内部 IP</label><span class="mono">{{ node.server.ip || '-' }}</span></div>
            <div v-for="a in otherAddresses" :key="a.type + a.address" class="jd-attr">
              <label>{{ ADDRESS_TEXT[a.type] || a.type }}</label><span class="mono">{{ a.address }}</span>
            </div>
            <div class="jd-attr"><label>调度</label><span :class="{ 'nd-warn': node.server.unSchedule }">{{ node.server.unSchedule ? '已停止调度' : '可调度' }}</span></div>
            <div class="jd-attr"><label>污点</label><span>{{ node.taints.length }}</span></div>
            <div class="jd-attr"><label>容器组</label><span>{{ node.resources.podCount }} / {{ node.resources.podAllocatable }}</span></div>
            <div class="jd-attr"><label>操作系统</label><span>{{ node.server.osImage || '-' }}</span></div>
            <div class="jd-attr"><label>架构</label><span>{{ node.server.osType }}/{{ node.server.arch }}</span></div>
            <div class="jd-attr"><label>内核版本</label><span class="mono">{{ node.server.kernelVersion || '-' }}</span></div>
            <div class="jd-attr"><label>容器运行时</label><span class="mono">{{ node.server.containerRuntime || '-' }}</span></div>
            <div class="jd-attr"><label>Kubelet 版本</label><span class="mono">{{ node.server.kubeletVersion || '-' }}</span></div>
            <div class="jd-attr"><label>KubeProxy 版本</label><span class="mono">{{ node.server.kubeProxyVersion || '-' }}</span></div>
            <div v-if="node.podCIDR" class="jd-attr"><label>容器组网段</label><span class="mono">{{ node.podCIDR }}</span></div>
            <div v-if="gpuText" class="jd-attr"><label>GPU</label><span>{{ gpuText }}</span></div>
            <div class="jd-attr"><label>临时存储（可分配）</label><span class="mono">{{ node.runtime.ephemeralStorage || '-' }}</span></div>
            <div class="jd-attr"><label>创建时间</label><span>{{ formatDateTime(node.server.createTime) }}</span></div>
          </div>
        </el-collapse-transition>
      </el-card>

      <!-- 资源 -->
      <el-card class="info-card">
        <template #header>
          <div class="card-header">
            <span>资源使用</span>
            <span class="card-header__hint">用量来自 metrics-server；已分配为节点上容器组预留（requests）的合计，决定还能调度多少容器组</span>
          </div>
        </template>
        <div class="nd-meters">
          <div v-for="m in meters" :key="m.title" class="nd-meter">
            <div class="nd-meter__head">
              <span class="nd-meter__title">{{ m.title }}</span>
              <span class="nd-meter__total">可分配 {{ m.allocatable }}<em v-if="m.capacity">（总量 {{ m.capacity }}）</em></span>
            </div>
            <div v-if="m.usage !== undefined" class="nd-meter__row">
              <label>实时用量</label>
              <div class="nd-meter__bar"><i :class="percentLevel(m.usagePercent)" :style="{ width: `${Math.min(m.usagePercent ?? 0, 100)}%` }" /></div>
              <span>{{ m.usage }}<em v-if="m.usagePercent !== null">{{ m.usagePercent }}%</em></span>
            </div>
            <div class="nd-meter__row">
              <label>已分配</label>
              <div class="nd-meter__bar"><i :class="percentLevel(m.requestPercent)" :style="{ width: `${Math.min(m.requestPercent ?? 0, 100)}%` }" /></div>
              <span>{{ m.requests }}<em v-if="m.requestPercent !== null">{{ m.requestPercent }}%</em></span>
            </div>
            <div v-if="m.limits !== undefined" class="nd-meter__row">
              <label>上限合计</label>
              <div class="nd-meter__bar"><i class="is-limit" :style="{ width: `${Math.min(m.limitPercent ?? 0, 100)}%` }" /></div>
              <span>{{ m.limits }}<em v-if="m.limitPercent !== null" :class="{ 'nd-warn': (m.limitPercent ?? 0) > 100 }">{{ m.limitPercent }}%</em></span>
            </div>
          </div>
        </div>
        <p v-if="node.resources.cpuUsage === null || node.resources.cpuUsage === undefined" class="nd-hint">没有获取到实时用量，请确认集群已安装 metrics-server。</p>
      </el-card>

      <div class="jd-tabbar">
        <button v-for="t in tabs" :key="t.key" type="button" :class="['jd-tab', { 'is-active': activeTab === t.key }]" @click="activeTab = t.key">
          {{ t.label }}
          <span v-if="t.count !== undefined" class="jd-tab__count">{{ t.count }}</span>
        </button>
      </div>

      <el-card v-if="activeTab === 'pods'" class="info-card">
        <template #header>
          <div class="card-header">
            <span>容器组</span>
            <el-tag size="small" type="info">{{ node.pods.length }}</el-tag>
            <span class="card-header__hint">运行在该节点上的全部容器组，包括已结束的</span>
          </div>
        </template>
        <JobPodsTable :pods="node.pods" :cluster-id="clusterId" show-namespace hide-node />
      </el-card>

      <el-card v-else-if="activeTab === 'status'" class="info-card">
        <template #header>
          <div class="card-header">
            <span>运行状态</span>
            <el-tag size="small" type="info">{{ node.conditions.length }}</el-tag>
          </div>
        </template>
        <el-table :data="node.conditions" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
          <el-table-column label="类型" min-width="200">
            <template #default="{ row }">{{ NODE_CONDITION_TEXT[row.type] || row.type }}</template>
          </el-table-column>
          <el-table-column label="状态" width="110">
            <template #default="{ row }">
              <el-tag :type="nodeConditionHealthy(row.type, row.status) ? 'success' : 'danger'" size="small">{{ nodeConditionHealthy(row.type, row.status) ? '正常' : '异常' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="原因" min-width="180" prop="reason" />
          <el-table-column label="消息" min-width="320" prop="message" />
          <el-table-column label="最近心跳" width="170">
            <template #default="{ row }">{{ formatDateTime(row.lastTime) }}</template>
          </el-table-column>
        </el-table>
      </el-card>

      <template v-else-if="activeTab === 'metadata'">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>污点</span>
              <el-tag size="small" type="info">{{ node.taints.length }}</el-tag>
              <el-button class="card-header__action" size="small" @click="taintsVisible = true">管理污点</el-button>
            </div>
          </template>
          <el-table :data="node.taints" style="width: 100%" :header-cell-class-name="tableHeaderClass">
            <el-table-column label="键" min-width="260">
              <template #default="{ row }">
                <span class="mono">{{ row.key }}</span>
                <el-tag v-if="row.system" size="small" type="info" effect="plain" class="nd-sys">系统</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="值" min-width="140">
              <template #default="{ row }"><span class="mono">{{ row.value || '-' }}</span></template>
            </el-table-column>
            <el-table-column label="效果" min-width="220">
              <template #default="{ row }">{{ taintEffectText(row.effect) }}</template>
            </el-table-column>
            <el-table-column label="添加时间" width="170">
              <template #default="{ row }">{{ formatDateTime(row.timeAdded) }}</template>
            </el-table-column>
            <template #empty>
              <span class="jd-muted">没有污点，所有容器组都可以调度到该节点</span>
            </template>
          </el-table>
        </el-card>
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>元数据</span>
              <el-button class="card-header__action" size="small" @click="labelsVisible = true">管理标签</el-button>
            </div>
          </template>
          <MetadataPanel :labels="node.labels" :annotations="node.annotations" />
        </el-card>
      </template>

      <el-card v-else class="info-card">
        <template #header>
          <div class="card-header">
            <span>事件</span>
            <el-tag size="small" type="info">{{ node.events.length }}</el-tag>
          </div>
        </template>
        <EventsTable :events="node.events" />
      </el-card>
    </div>

    <el-empty v-else description="节点不存在或已被删除" />

    <NodeLabelsDialog v-model="labelsVisible" :cluster-id="clusterId" :hostname="hostname" @success="fetchDetail(true)" />
    <NodeTaintsDialog v-model="taintsVisible" :cluster-id="clusterId" :hostname="hostname" @success="fetchDetail(true)" />
    <NodeDrainDialog v-model="drainVisible" :cluster-id="clusterId" :hostname="hostname" @success="fetchDetail(true)" />
    <YamlDialog
      v-model="yamlVisible"
      v-model:content="yamlContent"
      title="编辑YAML"
      subtitle="Node 配置"
      :info="[{ label: '集群', value: clusterName }, { label: '名称', value: hostname }, { label: '类型', value: 'Node' }]"
      :editor-title="`${hostname} - Node YAML`"
      :filename="`${hostname}-node.yaml`"
      tip="标签、污点和调度状态建议通过对应的按钮修改；status 由 kubelet 维护，不会显示和保存"
      :apply-target="{ clusterId, kind: 'node', name: hostname }"
      @applied="fetchDetail(true)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Document, PriceTag, Promotion, Refresh, VideoPause, VideoPlay, WarningFilled } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import JobPodsTable from '@/components/job/JobPodsTable.vue'
import MetadataPanel from '@/components/job/MetadataPanel.vue'
import EventsTable from '@/components/job/EventsTable.vue'
import NodeLabelsDialog from '@/components/node/NodeLabelsDialog.vue'
import NodeTaintsDialog from '@/components/node/NodeTaintsDialog.vue'
import NodeDrainDialog from '@/components/node/NodeDrainDialog.vue'
import { nodeApi, type NodeDetailVO } from '@/api/node'
import { fetchResourceYaml } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useTabBack } from '@/composables/useTabBack'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { formatDateTime } from '@/utils/job'
import { formatCpu, formatMemory } from '@/utils/resource'
import {
  NODE_CONDITION_TEXT,
  nodeConditionHealthy,
  nodeStatusTag,
  nodeStatusText,
  percentLevel,
  percentOf,
  roleText,
  taintEffectText
} from '@/utils/node'
import '@/assets/styles/job-detail.css'

type TabKey = 'pods' | 'status' | 'metadata' | 'events'

const ADDRESS_TEXT: Record<string, string> = {
  Hostname: '主机名',
  ExternalIP: '外部 IP',
  InternalDNS: '内部 DNS',
  ExternalDNS: '外部 DNS'
}

const route = useRoute()
// 返回当前标签内的上一页
const goBack = useTabBack()
const clusterStore = useClusterStore()

// 页面被 keep-alive 按完整路径缓存, 参数在创建时取一次
const clusterId = ref((route.params.clusterId as string) || (route.query.clusterId as string) || clusterStore.currentClusterId || '')
const hostname = ref((route.query.hostname as string) || (route.query.nodeName as string) || (route.query.node as string) || '')
const clusterName = computed(() => clusterStore.currentClusterName)

const loading = ref(false)
const node = ref<NodeDetailVO | null>(null)
const attrOpen = ref(true)
const activeTab = ref<TabKey>('pods')
const labelsVisible = ref(false)
const taintsVisible = ref(false)
const drainVisible = ref(false)
const yamlVisible = ref(false)
const yamlContent = ref('')

const tabs = computed(() => [
  { key: 'pods' as TabKey, label: '容器组', count: node.value?.pods.length },
  { key: 'status' as TabKey, label: '运行状态' },
  { key: 'metadata' as TabKey, label: '污点与元数据', count: node.value?.taints.length },
  { key: 'events' as TabKey, label: '事件', count: node.value?.events.length }
])

const otherAddresses = computed(() => (node.value?.addresses || []).filter(a => a.type !== 'InternalIP'))

const gpuText = computed(() => {
  const r = node.value?.runtime
  if (!r) return ''
  return [r.nvidiaGpu && `NVIDIA × ${r.nvidiaGpu}`, r.amdGpu && `AMD × ${r.amdGpu}`, r.intelGpu && `Intel × ${r.intelGpu}`].filter(Boolean).join('，')
})

const meters = computed(() => {
  const r = node.value?.resources
  if (!r) return []
  const hasMetrics = r.cpuUsage !== null && r.cpuUsage !== undefined
  return [
    {
      title: 'CPU',
      allocatable: formatCpu(r.cpuAllocatable),
      capacity: r.cpuCapacity !== r.cpuAllocatable ? formatCpu(r.cpuCapacity) : '',
      usage: hasMetrics ? formatCpu(r.cpuUsage) : undefined,
      usagePercent: percentOf(r.cpuUsage, r.cpuAllocatable),
      requests: formatCpu(r.cpuRequests),
      requestPercent: percentOf(r.cpuRequests, r.cpuAllocatable),
      limits: formatCpu(r.cpuLimits),
      limitPercent: percentOf(r.cpuLimits, r.cpuAllocatable)
    },
    {
      title: '内存',
      allocatable: formatMemory(r.memoryAllocatable),
      capacity: r.memoryCapacity !== r.memoryAllocatable ? formatMemory(r.memoryCapacity) : '',
      usage: hasMetrics ? formatMemory(r.memoryUsage) : undefined,
      usagePercent: percentOf(r.memoryUsage, r.memoryAllocatable),
      requests: formatMemory(r.memoryRequests),
      requestPercent: percentOf(r.memoryRequests, r.memoryAllocatable),
      limits: formatMemory(r.memoryLimits),
      limitPercent: percentOf(r.memoryLimits, r.memoryAllocatable)
    },
    {
      title: '容器组',
      allocatable: String(r.podAllocatable),
      capacity: '',
      usage: undefined,
      usagePercent: null,
      requests: `${r.podCount} 个`,
      requestPercent: percentOf(r.podCount, r.podAllocatable),
      limits: undefined,
      limitPercent: null
    }
  ]
})

// 每 30 秒刷新用量; 页面切走后停止
const active = ref(true)
let timer: ReturnType<typeof setTimeout> | undefined
const schedulePoll = () => {
  clearTimeout(timer)
  if (active.value && node.value) timer = setTimeout(() => fetchDetail(true), 30000)
}

const fetchDetail = async (silent = false) => {
  if (!clusterId.value || !hostname.value) return
  if (!silent) loading.value = true
  try {
    const resp = await nodeApi.getDetail({ clusterId: clusterId.value, hostname: hostname.value })
    node.value = resp.code === 200 ? resp.data : null
  } catch {
    if (!silent) node.value = null
  } finally {
    loading.value = false
    schedulePoll()
  }
}

const toggleSchedule = async () => {
  if (!node.value) return
  const unSchedule = !node.value.server.unSchedule
  try {
    await ElMessageBox.confirm(
      unSchedule
        ? '停止调度后，新的容器组不会再调度到该节点，已运行的容器组不受影响。如需迁移已有容器组，请使用「排空节点」。'
        : '启用后，新的容器组可以调度到该节点。',
      unSchedule ? '停止调度' : '启用调度',
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await nodeApi.setSchedule({ clusterId: clusterId.value, hostname: hostname.value, unSchedule })
    ElMessage.success(unSchedule ? '已停止调度' : '已启用调度')
    fetchDetail(true)
  } catch {
    // 请求拦截器已提示错误
  }
}

const openYaml = async () => {
  try {
    yamlContent.value = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'node', name: hostname.value })
    yamlVisible.value = true
  } catch {
    // 请求拦截器已提示错误
  }
}

onMounted(() => fetchDetail())
onActivated(() => {
  if (!active.value) {
    active.value = true
    fetchDetail(true)
  }
})
onDeactivated(() => {
  active.value = false
  clearTimeout(timer)
})
onBeforeUnmount(() => clearTimeout(timer))
</script>

<style scoped>
.nd-warn {
  color: #f59e0b;
  font-weight: 600;
}

.nd-sys {
  margin-left: 6px;
}

.nd-meters {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.nd-meter {
  padding: 14px 16px;
  border: 1px solid #e6ecf5;
  border-radius: 10px;
  background: #fbfcfe;
}

.nd-meter__head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 12px;
}

.nd-meter__title {
  color: #1e293b;
  font-size: 15px;
  font-weight: 600;
}

.nd-meter__total {
  color: #64748b;
  font-size: 12px;
}

.nd-meter__total em {
  color: #94a3b8;
  font-style: normal;
}

.nd-meter__row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 5px 0;
  font-size: 13px;
}

.nd-meter__row > label {
  flex: 0 0 60px;
  color: #64748b;
}

.nd-meter__row > span {
  flex: 0 0 120px;
  color: #1e293b;
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12.5px;
  text-align: right;
}

.nd-meter__row > span em {
  margin-left: 6px;
  color: #64748b;
  font-style: normal;
}

.nd-meter__bar {
  flex: 1;
  height: 8px;
  border-radius: 4px;
  background: #e8eef8;
  overflow: hidden;
}

.nd-meter__bar i {
  display: block;
  height: 100%;
  border-radius: 4px;
  background: #3b74f2;
}

.nd-meter__bar i.is-warning {
  background: #f59e0b;
}

.nd-meter__bar i.is-danger {
  background: #ef4444;
}

.nd-meter__bar i.is-limit {
  background: #a5b4fc;
}

.nd-hint {
  margin: 12px 0 0;
  color: #94a3b8;
  font-size: 12px;
}

@media (max-width: 1200px) {
  .nd-meters {
    grid-template-columns: 1fr;
  }
}
</style>
