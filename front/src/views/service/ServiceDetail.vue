<template>
  <div class="job-detail detail-page">
    <PageHeader
      :title="svc?.name || name || '服务详情'"
      :subtitle="svc ? svc.alias || svc.describe || `命名空间：${svc.namespace}` : '服务详情'"
      icon="svc"
      back
      @back="goBack"
    >
      <template v-if="svc" #meta>
        <el-tag type="primary" effect="dark">{{ typeLabel }}</el-tag>
        <el-tag v-if="svc.status !== 'Ready'" type="warning" effect="light">{{ svc.status }}</el-tag>
      </template>
      <template v-if="svc">
        <el-button :loading="loading" @click="fetchDetail">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button @click="editVisible = true">
          <el-icon><Edit /></el-icon>
          编辑
        </el-button>
        <el-button @click="openYaml">
          <el-icon><Document /></el-icon>
          编辑YAML
        </el-button>
        <el-button type="danger" @click="handleDelete">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </template>
    </PageHeader>

    <div v-if="loading && !svc" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="svc" class="detail-content">
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
            <div class="jd-attr">
              <label>命名空间</label>
              <span><el-link type="primary" @click="goNamespace">{{ svc.namespace }}</el-link></span>
            </div>
            <div class="jd-attr"><label>类型</label><span>{{ typeLabel }}</span></div>
            <div class="jd-attr">
              <label>集群 IP</label>
              <span class="mono">{{ svc.headless ? '无（Headless）' : (svc.clusterIPs?.length ? svc.clusterIPs.join(', ') : svc.clusterIP) || '-' }}</span>
            </div>
            <div class="jd-attr">
              <label>内部域名</label>
              <span class="mono">{{ svc.internalAccess }}<el-button link size="small" @click="copy(svc.internalAccess)"><el-icon><CopyDocument /></el-icon></el-button></span>
            </div>
            <div class="jd-attr"><label>外部访问</label><span class="mono">{{ externalText }}</span></div>
            <div v-if="svc.externalName" class="jd-attr"><label>外部服务域名</label><span class="mono">{{ svc.externalName }}</span></div>
            <div class="jd-attr">
              <label>会话保持</label>
              <span>{{ svc.sessionAffinity === 'ClientIP' ? `按客户端 IP（${svc.sessionAffinityTimeout || 10800} 秒）` : '无' }}</span>
            </div>
            <div v-if="svc.externalTrafficPolicy" class="jd-attr"><label>外部流量策略</label><span>{{ svc.externalTrafficPolicy }}</span></div>
            <div class="jd-attr"><label>选择器</label><span class="mono">{{ formatMap(svc.selector) || '无（需手动维护 Endpoints）' }}</span></div>
            <div v-if="svc.ipFamilies.length" class="jd-attr"><label>IP 协议族</label><span>{{ svc.ipFamilies.join(', ') }}</span></div>
            <div class="jd-attr"><label>创建时间</label><span>{{ formatDateTime(svc.createTime) }}</span></div>
            <div class="jd-attr"><label>创建者</label><span>{{ svc.creator || '-' }}</span></div>
          </div>
        </el-collapse-transition>
      </el-card>

      <div class="jd-tabbar">
        <button v-for="t in tabs" :key="t.key" type="button" :class="['jd-tab', { 'is-active': activeTab === t.key }]" @click="activeTab = t.key">
          {{ t.label }}
          <span v-if="t.count !== undefined" class="jd-tab__count">{{ t.count }}</span>
        </button>
      </div>

      <template v-if="activeTab === 'resources'">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>端口</span>
              <el-tag size="small" type="info">{{ svc.ports.length }}</el-tag>
            </div>
          </template>
          <el-table :data="svc.ports" style="width: 100%" :header-cell-class-name="tableHeaderClass">
            <el-table-column label="名称" min-width="140">
              <template #default="{ row }">{{ row.name || '-' }}</template>
            </el-table-column>
            <el-table-column label="协议" width="100" prop="protocol" />
            <el-table-column label="服务端口" width="120" prop="port" />
            <el-table-column label="容器端口" width="140" prop="targetPort" />
            <el-table-column label="节点端口" width="120">
              <template #default="{ row }">{{ row.nodePort || '-' }}</template>
            </el-table-column>
            <el-table-column label="集群内访问地址" min-width="280">
              <template #default="{ row }">
                <span class="mono">{{ svc.headless ? `${svc.internalAccess}:${row.port}` : `${svc.clusterIP}:${row.port}` }}</span>
              </template>
            </el-table-column>
            <template #empty>
              <span class="jd-muted">没有端口</span>
            </template>
          </el-table>
        </el-card>

        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>工作负载</span>
              <el-tag size="small" type="info">{{ svc.workloads.length }}</el-tag>
              <span class="card-header__hint">容器组标签与选择器匹配的工作负载</span>
            </div>
          </template>
          <el-table :data="svc.workloads" style="width: 100%" :header-cell-class-name="tableHeaderClass">
            <el-table-column label="名称" min-width="220">
              <template #default="{ row }"><el-link type="primary" @click="goWorkload(row.kind, row.name)">{{ row.name }}</el-link></template>
            </el-table-column>
            <el-table-column label="类型" width="160">
              <template #default="{ row }">{{ WORKLOAD_KINDS[row.kind as WorkloadKind]?.label || row.kind }}</template>
            </el-table-column>
            <el-table-column label="就绪 / 副本" width="140">
              <template #default="{ row }">{{ row.readyReplicas }} / {{ row.replicas }}</template>
            </el-table-column>
            <template #empty>
              <span class="jd-muted">没有匹配的工作负载</span>
            </template>
          </el-table>
        </el-card>

        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>后端地址</span>
              <el-tag size="small" type="info">{{ svc.endpoints.length }}</el-tag>
              <span class="card-header__hint">就绪 {{ readyEndpoints }} 个，只有就绪的地址会接收流量</span>
            </div>
          </template>
          <el-table :data="svc.endpoints" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
            <el-table-column label="IP" width="170">
              <template #default="{ row }"><span class="mono">{{ row.ip }}</span></template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.ready ? 'success' : 'warning'" size="small">{{ row.ready ? '就绪' : '未就绪' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="容器组" min-width="220">
              <template #default="{ row }">
                <el-link v-if="row.podName" type="primary" @click="goPod(row.podName)">{{ row.podName }}</el-link>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="节点" min-width="160">
              <template #default="{ row }">{{ row.nodeName || '-' }}</template>
            </el-table-column>
            <el-table-column label="端口" min-width="200">
              <template #default="{ row }"><span class="mono">{{ row.ports || '-' }}</span></template>
            </el-table-column>
            <template #empty>
              <span class="jd-muted">{{ svc.serviceType === 'ExternalName' ? '外部服务没有后端地址' : '没有后端地址，请检查选择器和容器组的就绪状态' }}</span>
            </template>
          </el-table>
        </el-card>

        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>容器组</span>
              <el-tag size="small" type="info">{{ svc.pods.length }}</el-tag>
            </div>
          </template>
          <JobPodsTable :pods="svc.pods" :cluster-id="clusterId" />
        </el-card>

        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>应用路由</span>
              <el-tag size="small" type="info">{{ svc.ingresses.length }}</el-tag>
              <span class="card-header__hint">把该服务作为后端的 Ingress</span>
              <el-button class="card-header__action" size="small" type="primary" @click="goIngressList">
                <el-icon><Plus /></el-icon>
                创建应用路由
              </el-button>
            </div>
          </template>
          <div v-if="svc.ingresses.length" class="sd-tags">
            <el-tag v-for="ing in svc.ingresses" :key="ing" class="sd-link" @click="goIngress(ing)">{{ ing }}</el-tag>
          </div>
          <span v-else class="jd-muted">还没有应用路由，集群外无法通过域名访问该服务</span>
        </el-card>
      </template>

      <el-card v-else-if="activeTab === 'metadata'" class="info-card">
        <template #header>
          <div class="card-header"><span>元数据</span></div>
        </template>
        <MetadataPanel :labels="svc.labels" :annotations="svc.annotations" />
      </el-card>

      <el-card v-else class="info-card">
        <template #header>
          <div class="card-header">
            <span>事件</span>
            <el-tag size="small" type="info">{{ svc.events.length }}</el-tag>
          </div>
        </template>
        <EventsTable :events="svc.events" />
      </el-card>
    </div>

    <el-empty v-else description="服务不存在或已被删除" />

    <ServiceFormDialog v-model="editVisible" :cluster-id="clusterId" :edit-target="svc ? { namespace: svc.namespace, name: svc.name } : null" @success="fetchDetail" />
    <YamlDialog
      v-model="yamlVisible"
      v-model:content="yamlContent"
      title="编辑YAML"
      subtitle="Service 配置"
      :info="[{ label: '集群', value: clusterName }, { label: '命名空间', value: svc?.namespace }, { label: '名称', value: svc?.name }, { label: '类型', value: 'Service' }]"
      :editor-title="`${svc?.name} - Service YAML`"
      :filename="`${svc?.name}-service.yaml`"
      :apply-target="svc ? { clusterId, kind: 'service', namespace: svc.namespace, name: svc.name } : null"
      @applied="fetchDetail"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, CopyDocument, Delete, Document, Edit, Plus, Refresh } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import JobPodsTable from '@/components/job/JobPodsTable.vue'
import MetadataPanel from '@/components/job/MetadataPanel.vue'
import EventsTable from '@/components/job/EventsTable.vue'
import ServiceFormDialog from '@/components/service/ServiceFormDialog.vue'
import { serviceApi, type ServiceDetailVO } from '@/api/service'
import { fetchResourceYaml } from '@/api/resource'
import type { WorkloadKind } from '@/api/workload'
import { useClusterStore } from '@/stores/cluster'
import { useTabBack } from '@/composables/useTabBack'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { formatDateTime } from '@/utils/job'
import { WORKLOAD_KINDS, workloadDetailPath } from '@/utils/workload'
import { serviceTypeLabel } from '@/utils/network'
import '@/assets/styles/job-detail.css'

type TabKey = 'resources' | 'metadata' | 'events'

const route = useRoute()
const router = useRouter()
const goBack = useTabBack()
const clusterStore = useClusterStore()

// 页面被 keep-alive 按完整路径缓存, 参数在创建时取一次
const clusterId = ref((route.params.clusterId as string) || clusterStore.currentClusterId || '')
const namespace = ref((route.query.namespace as string) || '')
const name = ref((route.query.name as string) || '')
const clusterName = computed(() => clusterStore.currentClusterName)

const loading = ref(false)
const svc = ref<ServiceDetailVO | null>(null)
const attrOpen = ref(true)
const activeTab = ref<TabKey>('resources')
const editVisible = ref(false)
const yamlVisible = ref(false)
const yamlContent = ref('')

const tabs = computed(() => [
  { key: 'resources' as TabKey, label: '资源状态', count: svc.value?.endpoints.length },
  { key: 'metadata' as TabKey, label: '元数据' },
  { key: 'events' as TabKey, label: '事件', count: svc.value?.events.length }
])

const typeLabel = computed(() => (svc.value ? serviceTypeLabel(svc.value) : ''))
const readyEndpoints = computed(() => svc.value?.endpoints.filter(e => e.ready).length || 0)

const externalText = computed(() => {
  const s = svc.value
  if (!s) return '-'
  if (s.serviceType === 'NodePort') return s.ports.map(p => `<节点 IP>:${p.nodePort}`).join('，') || '-'
  if (s.serviceType === 'LoadBalancer') {
    if (!s.loadBalancerIngress.length) return '等待分配外部地址'
    return s.loadBalancerIngress.flatMap(ip => s.ports.map(p => `${ip}:${p.port}`)).join('，')
  }
  if (s.externalIPs?.length) return s.externalIPs.join(', ')
  return '不开放'
})

const formatMap = (m?: Record<string, string> | null) =>
  Object.entries(m || {})
    .map(([k, v]) => `${k}=${v}`)
    .join(', ')

const copy = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制')
  } catch {
    ElMessage.warning('复制失败，请手动选择复制')
  }
}

const fetchDetail = async () => {
  if (!clusterId.value || !namespace.value || !name.value) return
  loading.value = true
  try {
    const resp = await serviceApi.getDetail({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })
    svc.value = resp.code === 200 ? resp.data : null
  } catch {
    svc.value = null
  } finally {
    loading.value = false
  }
}

const base = `/cluster/${clusterId.value}`
const goNamespace = () => router.push({ path: `${base}/namespace/detail`, query: { namespace: namespace.value } })
const goPod = (pod: string) => router.push({ path: `${base}/pod/detail`, query: { namespace: namespace.value, name: pod } })
const goWorkload = (kind: WorkloadKind, wl: string) => router.push({ path: workloadDetailPath(clusterId.value, kind), query: { namespace: namespace.value, name: wl } })
const goIngress = (ing: string) => router.push({ path: `${base}/service/ingress/detail`, query: { namespace: namespace.value, name: ing } })
const goIngressList = () => router.push({ path: `${base}/service/ingress`, query: { namespace: namespace.value } })

const openYaml = async () => {
  try {
    yamlContent.value = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'service', namespace: namespace.value, name: name.value })
    yamlVisible.value = true
  } catch {
    // 请求拦截器已提示错误
  }
}

const handleDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除服务 "${name.value}" 吗？${svc.value?.ingresses.length ? `应用路由 ${svc.value.ingresses.join('、')} 将无法访问后端，` : ''}此操作不可恢复。`,
      '确认删除',
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await serviceApi.deleteService({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })
    ElMessage.success('删除成功')
    goBack()
  } catch {
    // 请求拦截器已提示错误
  }
}

onMounted(fetchDetail)
</script>

<style scoped>
.sd-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.sd-link {
  cursor: pointer;
}
</style>
