<template>
  <div class="job-detail detail-page">
    <PageHeader
      :title="wl?.name || name || `${meta.label}详情`"
      :subtitle="wl ? wl.alias || wl.describe || `${meta.cn} · 命名空间：${wl.namespace}` : `${meta.label}详情`"
      :icon="meta.icon"
      back
      @back="goBack"
    >
      <template v-if="wl" #meta>
        <el-tag :type="workloadStatusTag(wl.status)" effect="dark">{{ wl.status }}</el-tag>
        <el-tag type="info" effect="light">{{ meta.label }}</el-tag>
      </template>
      <template v-if="wl">
        <el-button :loading="loading" @click="fetchDetail()">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button @click="infoVisible = true">
          <el-icon><Edit /></el-icon>
          编辑信息
        </el-button>
        <el-button @click="openImageDialog">
          <el-icon><Picture /></el-icon>
          更新镜像
        </el-button>
        <el-button @click="handleRestart">
          <el-icon><RefreshRight /></el-icon>
          重新创建
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

    <div v-if="loading && !wl" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="wl" class="detail-content">
      <el-card class="info-card jd-attr-card">
        <template #header>
          <div class="card-header" @click="attrOpen = !attrOpen">
            <span>属性</span>
            <el-icon :class="['jd-collapse', { 'is-collapsed': !attrOpen }]"><ArrowDown /></el-icon>
          </div>
        </template>
        <el-collapse-transition>
          <div v-show="attrOpen" class="wd-attr">
            <!-- 副本数(与 KubeSphere 一样可以直接加减) -->
            <div class="wd-replicas">
              <div class="wd-replicas__ring" :style="{ '--p': ringPercent }">
                <div class="wd-replicas__inner">
                  <strong>{{ wl.readyReplicas }}</strong>
                  <span>/ {{ wl.replicas }}</span>
                </div>
              </div>
              <div class="wd-replicas__text">{{ kind === 'daemonset' ? '就绪 / 期望节点数' : '就绪 / 期望副本数' }}</div>
              <div v-if="kind !== 'daemonset'" class="wd-replicas__ctrl">
                <el-button circle size="small" :disabled="pendingReplicas <= 0" @click="pendingReplicas--"><el-icon><Minus /></el-icon></el-button>
                <span :class="['wd-replicas__value', { 'is-changed': pendingReplicas !== wl.replicas }]">{{ pendingReplicas }}</span>
                <el-button circle size="small" :disabled="pendingReplicas >= 1000" @click="pendingReplicas++"><el-icon><Plus /></el-icon></el-button>
              </div>
              <div v-if="kind !== 'daemonset' && pendingReplicas !== wl.replicas" class="wd-replicas__save">
                <el-button size="small" type="primary" :loading="scaling" @click="applyScale">应用</el-button>
                <el-button size="small" @click="pendingReplicas = wl.replicas">取消</el-button>
              </div>
            </div>
            <div class="jd-attr-grid">
              <div class="jd-attr"><label>集群</label><span>{{ clusterName || clusterId }}</span></div>
              <div class="jd-attr">
                <label>命名空间</label>
                <span><el-link type="primary" @click="goNamespace">{{ wl.namespace }}</el-link></span>
              </div>
              <div class="jd-attr"><label>状态</label><span>{{ wl.status }}</span></div>
              <div class="jd-attr"><label>已更新 / 可用</label><span>{{ wl.updatedReplicas }} / {{ wl.availableReplicas }}</span></div>
              <div class="jd-attr"><label>更新策略</label><span>{{ strategyText(kind, wl.strategy) }}</span></div>
              <div class="jd-attr"><label>当前版本</label><span>{{ wl.currentRevision ? `#${wl.currentRevision}` : '-' }}</span></div>
              <div v-if="kind === 'deployment'" class="jd-attr">
                <label>最大超出 / 不可用</label>
                <span>{{ wl.maxSurge || '25%' }} / {{ wl.maxUnavailable || '25%' }}</span>
              </div>
              <div v-if="kind === 'daemonset'" class="jd-attr"><label>最大不可用数</label><span>{{ wl.maxUnavailable || '1' }}</span></div>
              <template v-if="kind === 'statefulset'">
                <div class="jd-attr">
                  <label>Headless 服务</label>
                  <span><el-link v-if="wl.serviceName" type="primary" @click="goService(wl.serviceName)">{{ wl.serviceName }}</el-link><template v-else>-</template></span>
                </div>
                <div class="jd-attr"><label>容器组管理策略</label><span>{{ wl.podManagementPolicy === 'Parallel' ? '并行（Parallel）' : '按顺序（OrderedReady）' }}</span></div>
                <div class="jd-attr"><label>分批更新序号</label><span>{{ wl.partition ?? 0 }}</span></div>
              </template>
              <div class="jd-attr"><label>最短就绪时间</label><span>{{ wl.minReadySeconds ? `${wl.minReadySeconds} 秒` : '-' }}</span></div>
              <div class="jd-attr"><label>保留历史版本</label><span>{{ wl.revisionHistoryLimit ?? 10 }}</span></div>
              <div class="jd-attr"><label>服务账户</label><span>{{ wl.serviceAccountName || 'default' }}</span></div>
              <div class="jd-attr">
                <label>选择器</label>
                <span class="mono">{{ formatMap(wl.selector) }}</span>
              </div>
              <div class="jd-attr"><label>节点选择器</label><span class="mono">{{ formatMap(wl.nodeSelector) || '不限制' }}</span></div>
              <div class="jd-attr"><label>创建时间</label><span>{{ formatDateTime(wl.createTime) }}</span></div>
              <div class="jd-attr"><label>更新时间</label><span>{{ formatDateTime(wl.updateTime) }}</span></div>
              <div class="jd-attr"><label>创建者</label><span>{{ wl.creator || '-' }}</span></div>
            </div>
          </div>
        </el-collapse-transition>
      </el-card>

      <div class="jd-tabbar">
        <button v-for="t in tabs" :key="t.key" type="button" :class="['jd-tab', { 'is-active': activeTab === t.key }]" @click="activeTab = t.key">
          {{ t.label }}
          <span v-if="t.count !== undefined" class="jd-tab__count">{{ t.count }}</span>
        </button>
      </div>

      <!-- 资源状态 -->
      <template v-if="activeTab === 'resources'">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>容器组</span>
              <el-tag size="small" type="info">{{ wl.pods.length }}</el-tag>
              <span class="card-header__hint">就绪 {{ wl.readyReplicas }}，已更新 {{ wl.updatedReplicas }}</span>
            </div>
          </template>
          <JobPodsTable :pods="wl.pods" :cluster-id="clusterId" />
        </el-card>
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>容器</span>
              <el-tag size="small" type="info">{{ wl.containers.length }}</el-tag>
            </div>
          </template>
          <ContainerTemplates :containers="wl.containers" />
        </el-card>
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>服务</span>
              <el-tag size="small" type="info">{{ wl.services.length }}</el-tag>
              <span class="card-header__hint">选择器选中这些容器组的服务</span>
            </div>
          </template>
          <el-table :data="wl.services" style="width: 100%" :header-cell-class-name="tableHeaderClass">
            <el-table-column label="服务名称" min-width="200">
              <template #default="{ row }"><el-link type="primary" @click="goService(row.name)">{{ row.name }}</el-link></template>
            </el-table-column>
            <el-table-column label="类型" width="130" prop="type" />
            <el-table-column label="集群 IP" width="160">
              <template #default="{ row }"><span class="mono">{{ row.clusterIP === 'None' ? '无（Headless）' : row.clusterIP || '-' }}</span></template>
            </el-table-column>
            <el-table-column label="端口" min-width="260">
              <template #default="{ row }"><span class="mono">{{ row.ports || '-' }}</span></template>
            </el-table-column>
            <template #empty>
              <el-empty description="还没有服务，集群内外都无法通过服务访问这些容器组" :image-size="72">
                <el-button size="small" type="primary" @click="goServiceList">去创建服务</el-button>
              </el-empty>
            </template>
          </el-table>
        </el-card>
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>存储</span>
              <el-tag size="small" type="info">{{ wl.volumes.length + wl.volumeClaimTemplates.length }}</el-tag>
            </div>
          </template>
          <el-table v-if="wl.volumeClaimTemplates.length" :data="wl.volumeClaimTemplates" style="width: 100%; margin-bottom: 16px" :header-cell-class-name="tableHeaderClass">
            <el-table-column label="存储卷模板" min-width="160" prop="name" />
            <el-table-column label="存储类" min-width="160">
              <template #default="{ row }">{{ row.storageClassName || '默认存储类' }}</template>
            </el-table-column>
            <el-table-column label="访问模式" width="160">
              <template #default="{ row }">{{ formatAccessModes(row.accessModes) }}</template>
            </el-table-column>
            <el-table-column label="容量" width="120" prop="capacity" />
          </el-table>
          <el-table :data="wl.volumes" style="width: 100%" :header-cell-class-name="tableHeaderClass">
            <el-table-column label="存储卷" min-width="160" prop="name" />
            <el-table-column label="类型" width="140">
              <template #default="{ row }">{{ volumeTypeText(row.type) }}</template>
            </el-table-column>
            <el-table-column label="来源" min-width="260">
              <template #default="{ row }"><span class="mono">{{ row.source || '-' }}</span></template>
            </el-table-column>
            <template #empty>
              <span class="jd-muted">没有挂载存储卷</span>
            </template>
          </el-table>
        </el-card>
      </template>

      <!-- 版本记录 -->
      <el-card v-else-if="activeTab === 'revisions'" class="info-card">
        <template #header>
          <div class="card-header">
            <span>版本记录</span>
            <el-tag size="small" type="info">{{ wl.revisions.length }}</el-tag>
            <span class="card-header__hint">修改容器组模板（镜像、环境变量等）会生成新版本，最多保留 {{ wl.revisionHistoryLimit ?? 10 }} 个历史版本</span>
          </div>
        </template>
        <el-table :data="revisionRows" style="width: 100%" :header-cell-class-name="tableHeaderClass">
          <el-table-column label="版本" width="130">
            <template #default="{ row }">
              #{{ row.revision }}
              <el-tag v-if="row.current" size="small" type="success" effect="plain">当前</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="镜像" min-width="300">
            <template #default="{ row }">
              <div v-for="img in row.images" :key="img" class="mono">{{ img }}</div>
            </template>
          </el-table-column>
          <el-table-column v-if="kind === 'deployment'" label="副本数" width="100" align="center" header-align="center">
            <template #default="{ row }">{{ row.readyReplicas }} / {{ row.replicas }}</template>
          </el-table-column>
          <el-table-column label="变更原因" min-width="200">
            <template #default="{ row }">{{ row.changeCause || '-' }}</template>
          </el-table-column>
          <el-table-column label="创建时间" width="170">
            <template #default="{ row }">{{ formatDateTime(row.createTime) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="110" align="center" header-align="center">
            <template #default="{ row }">
              <el-button v-if="!row.current" link type="primary" size="small" @click="handleRollback(row.revision)">回退到此版本</el-button>
              <span v-else class="jd-muted">-</span>
            </template>
          </el-table-column>
          <template #empty>
            <el-empty description="暂无版本记录" :image-size="72" />
          </template>
        </el-table>
        <ListPagination v-model:current-page="revPage" v-model:page-size="revPageSize" class="jd-pager" :total="revTotal" :page-sizes="[10, 20, 50]" />
      </el-card>

      <!-- 元数据 -->
      <template v-else-if="activeTab === 'metadata'">
        <el-card class="info-card">
          <template #header>
            <div class="card-header"><span>元数据</span></div>
          </template>
          <MetadataPanel :labels="wl.labels" :annotations="wl.annotations" />
        </el-card>
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>状态条件</span>
              <el-tag size="small" type="info">{{ wl.conditions.length }}</el-tag>
            </div>
          </template>
          <el-table :data="wl.conditions" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
            <el-table-column label="类型" min-width="140" prop="type" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.status === 'True' ? 'success' : 'warning'" size="small">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="原因" min-width="180" prop="reason" />
            <el-table-column label="消息" min-width="300" prop="message" />
            <el-table-column label="更新时间" width="170">
              <template #default="{ row }">{{ formatDateTime(row.lastTime) }}</template>
            </el-table-column>
            <template #empty>
              <span class="jd-muted">没有状态条件</span>
            </template>
          </el-table>
        </el-card>
      </template>

      <el-card v-else-if="activeTab === 'hpa'" class="info-card">
        <template #header>
          <div class="card-header">
            <span>弹性伸缩</span>
            <span class="card-header__hint">根据 CPU / 内存使用率自动调整副本数（HorizontalPodAutoscaler）</span>
          </div>
        </template>
        <HpaPanel :cluster-id="clusterId" :namespace="namespace" :kind="kind as 'deployment' | 'statefulset'" :name="name" @change="(n: number) => (hpaCount = n)" />
      </el-card>

      <el-card v-else-if="activeTab === 'env'" class="info-card">
        <template #header>
          <div class="card-header"><span>环境变量</span></div>
        </template>
        <ContainerEnvList :containers="wl.containers" />
      </el-card>

      <el-card v-else class="info-card">
        <template #header>
          <div class="card-header">
            <span>事件</span>
            <el-tag size="small" type="info">{{ wl.events.length }}</el-tag>
          </div>
        </template>
        <EventsTable :events="wl.events" />
      </el-card>
    </div>

    <el-empty v-else :description="`${meta.label} 不存在或已被删除`" />

    <ResourceInfoDialog
      v-model="infoVisible"
      :name="wl?.name || ''"
      :alias="wl?.alias"
      :describe="wl?.describe"
      :submit="saveInfo"
      @success="fetchDetail()"
    />

    <el-dialog v-model="imageVisible" title="更新镜像" width="720px" :close-on-click-modal="false" class="config-dialog form-dialog--compact">
      <template #header>
        <div class="dialog-header"><h3 class="dialog-title">更新镜像</h3></div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <p class="wd-tip">修改后会{{ wl?.strategy === 'OnDelete' ? '在手动删除容器组时' : '按更新策略逐步' }}替换容器组，并生成新的版本记录，可在「版本记录」中回退。</p>
          <el-form label-width="140px">
            <el-form-item v-for="c in wl?.containers || []" :key="(c.init ? 'init-' : '') + c.name" :label="c.name">
              <el-input v-model="imageForm[c.name]" class="mono-input">
                <template v-if="c.init" #prepend>初始化</template>
              </el-input>
            </el-form-item>
          </el-form>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="imageVisible = false">取消</el-button>
          <el-button type="primary" :loading="imageSaving" @click="saveImages">更新</el-button>
        </div>
      </template>
    </el-dialog>

    <YamlDialog
      v-model="yamlVisible"
      v-model:content="yamlContent"
      title="编辑YAML"
      :subtitle="`${meta.label} 配置`"
      :info="[{ label: '集群', value: clusterName }, { label: '命名空间', value: wl?.namespace }, { label: '名称', value: wl?.name }, { label: '类型', value: meta.label }]"
      :editor-title="`${wl?.name} - ${meta.label} YAML`"
      :filename="`${wl?.name}-${kind}.yaml`"
      :apply-target="wl ? { clusterId, kind, namespace: wl.namespace, name: wl.name } : null"
      @applied="fetchDetail()"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Delete, Document, Edit, Minus, Picture, Plus, Refresh, RefreshRight } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import ResourceInfoDialog from '@/components/ResourceInfoDialog.vue'
import JobPodsTable from '@/components/job/JobPodsTable.vue'
import ContainerTemplates from '@/components/job/ContainerTemplates.vue'
import ContainerEnvList from '@/components/job/ContainerEnvList.vue'
import MetadataPanel from '@/components/job/MetadataPanel.vue'
import EventsTable from '@/components/job/EventsTable.vue'
import HpaPanel from '@/components/workload/HpaPanel.vue'
import { workloadApi, type WorkloadDetailVO, type WorkloadKind } from '@/api/workload'
import { fetchResourceYaml, resourceApi } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useClientPage } from '@/composables/useClientPage'
import { useTabBack } from '@/composables/useTabBack'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { formatDateTime } from '@/utils/job'
import { formatAccessModes } from '@/utils/storageclass'
import { WORKLOAD_KINDS, strategyText, volumeTypeText, workloadStatusTag } from '@/utils/workload'
import '@/assets/styles/job-detail.css'

type TabKey = 'resources' | 'revisions' | 'hpa' | 'metadata' | 'env' | 'events'

const route = useRoute()
const router = useRouter()
// 返回当前标签内的上一页
const goBack = useTabBack()
const clusterStore = useClusterStore()

// 页面被 keep-alive 按完整路径缓存, 参数在创建时取一次
const kind = ref<WorkloadKind>((route.meta.kind as WorkloadKind) || 'deployment')
const clusterId = ref((route.params.clusterId as string) || clusterStore.currentClusterId || '')
const namespace = ref((route.query.namespace as string) || '')
const name = ref((route.query.name as string) || '')
const clusterName = computed(() => clusterStore.currentClusterName)
const meta = computed(() => WORKLOAD_KINDS[kind.value])

const loading = ref(false)
const wl = ref<WorkloadDetailVO | null>(null)
const attrOpen = ref(true)
const activeTab = ref<TabKey>('resources')
// 指向该工作负载的弹性伸缩数量; 未加载时为 null
const hpaCount = ref<number | null>(null)
const tabs = computed(() => [
  { key: 'resources' as TabKey, label: '资源状态', count: wl.value?.pods.length },
  { key: 'revisions' as TabKey, label: '版本记录', count: wl.value?.revisions.length },
  ...(kind.value === 'daemonset' ? [] : [{ key: 'hpa' as TabKey, label: '弹性伸缩', count: hpaCount.value ?? undefined }]),
  { key: 'metadata' as TabKey, label: '元数据' },
  { key: 'env' as TabKey, label: '环境变量' },
  { key: 'events' as TabKey, label: '事件', count: wl.value?.events.length }
])

const { page: revPage, pageSize: revPageSize, total: revTotal, rows: revisionRows } = useClientPage(() => wl.value?.revisions || [])

const ringPercent = computed(() => {
  if (!wl.value || !wl.value.replicas) return 0
  return Math.min(100, Math.round((wl.value.readyReplicas / wl.value.replicas) * 100))
})

const target = () => ({ clusterId: clusterId.value, kind: kind.value, namespace: namespace.value, name: name.value })

// 更新中时定时刷新; 页面切走后停止
const active = ref(true)
let timer: ReturnType<typeof setTimeout> | undefined
const schedulePoll = () => {
  clearTimeout(timer)
  if (active.value && wl.value && wl.value.status === '更新中') timer = setTimeout(() => fetchDetail(true), 5000)
}

const pendingReplicas = ref(0)
const scaling = ref(false)

const fetchDetail = async (silent = false) => {
  if (!clusterId.value || !namespace.value || !name.value) return
  if (!silent) loading.value = true
  try {
    const resp = await workloadApi.getDetail(target())
    const changed = !wl.value || pendingReplicas.value === wl.value.replicas
    wl.value = resp.code === 200 ? resp.data : null
    // 正在调整副本数时不覆盖用户输入
    if (wl.value && changed) pendingReplicas.value = wl.value.replicas
  } catch {
    if (!silent) wl.value = null
  } finally {
    loading.value = false
    schedulePoll()
  }
}

const formatMap = (m?: Record<string, string> | null) =>
  Object.entries(m || {})
    .map(([k, v]) => `${k}=${v}`)
    .join(', ')

const goNamespace = () => router.push({ path: `/cluster/${clusterId.value}/namespace/detail`, query: { namespace: namespace.value } })
const goService = (svc: string) => router.push({ path: `/cluster/${clusterId.value}/service/service/detail`, query: { namespace: namespace.value, name: svc } })
const goServiceList = () => router.push({ path: `/cluster/${clusterId.value}/service/service`, query: { namespace: namespace.value } })

// 统计指向该工作负载的弹性伸缩(用于页签数量和手动调整副本时的提示)
const loadHpaCount = async () => {
  if (kind.value === 'daemonset') return
  try {
    const { data } = await resourceApi.objects({ kind: 'horizontalpodautoscaler', clusterId: clusterId.value, namespace: namespace.value })
    const targetKind = kind.value === 'statefulset' ? 'StatefulSet' : 'Deployment'
    hpaCount.value = (data || []).filter(h => h.spec?.scaleTargetRef?.kind === targetKind && h.spec?.scaleTargetRef?.name === name.value).length
  } catch {
    hpaCount.value = null
  }
}

const applyScale = async () => {
  if (!wl.value) return
  if (hpaCount.value) {
    try {
      await ElMessageBox.confirm('该工作负载已设置弹性伸缩，手动调整的副本数会被弹性伸缩重新调整。仍要继续吗？', '提示', {
        type: 'warning',
        confirmButtonText: '继续调整',
        cancelButtonText: '取消'
      })
    } catch {
      return
    }
  }
  scaling.value = true
  try {
    await workloadApi.scale({ ...target(), replicas: pendingReplicas.value })
    ElMessage.success('副本数已调整')
    await fetchDetail(true)
    if (wl.value) pendingReplicas.value = wl.value.replicas
  } catch {
    // 请求拦截器已提示错误
  } finally {
    scaling.value = false
  }
}

const infoVisible = ref(false)
const saveInfo = (alias: string, describe: string) => workloadApi.updateInfo({ ...target(), alias, describe })

const imageVisible = ref(false)
const imageSaving = ref(false)
const imageForm = reactive<Record<string, string>>({})
const openImageDialog = () => {
  Object.keys(imageForm).forEach(k => delete imageForm[k])
  wl.value?.containers.forEach(c => (imageForm[c.name] = c.image))
  imageVisible.value = true
}
const saveImages = async () => {
  const images: Record<string, string> = {}
  wl.value?.containers.forEach(c => {
    const v = (imageForm[c.name] || '').trim()
    if (v && v !== c.image) images[c.name] = v
  })
  if (!Object.keys(images).length) {
    ElMessage.info('镜像没有变化')
    return
  }
  imageSaving.value = true
  try {
    await workloadApi.updateImages({ ...target(), images })
    ElMessage.success('镜像已更新')
    imageVisible.value = false
    fetchDetail()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    imageSaving.value = false
  }
}

const handleRestart = async () => {
  try {
    await ElMessageBox.confirm(`将按当前配置逐个重新创建 "${name.value}" 的容器组，与 kubectl rollout restart 相同。`, '重新创建', {
      type: 'warning',
      confirmButtonText: '重新创建',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await workloadApi.restart(target())
    ElMessage.success('已开始重新创建容器组')
    fetchDetail()
  } catch {
    // 请求拦截器已提示错误
  }
}

const handleRollback = async (revision: number) => {
  try {
    await ElMessageBox.confirm(`将容器组模板回退到版本 #${revision}，并按更新策略替换容器组。`, '版本回退', {
      type: 'warning',
      confirmButtonText: '回退',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await workloadApi.rollback({ ...target(), revision })
    ElMessage.success('已回退')
    fetchDetail()
  } catch {
    // 请求拦截器已提示错误
  }
}

const yamlVisible = ref(false)
const yamlContent = ref('')
const openYaml = async () => {
  try {
    yamlContent.value = await fetchResourceYaml(target())
    yamlVisible.value = true
  } catch {
    // 请求拦截器已提示错误
  }
}

const handleDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除 ${meta.value.label} "${name.value}" 吗？它的容器组会一起删除${kind.value === 'statefulset' ? '，持久卷声明会保留' : ''}，此操作不可恢复。`,
      '确认删除',
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await workloadApi.delete(target())
    ElMessage.success('删除成功')
    goBack()
  } catch {
    // 请求拦截器已提示错误
  }
}

onMounted(() => {
  fetchDetail()
  loadHpaCount()
})
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
.wd-attr {
  display: flex;
  gap: 28px;
}

.wd-attr .jd-attr-grid {
  flex: 1;
  min-width: 0;
}

.wd-replicas {
  display: flex;
  flex: 0 0 170px;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 6px 0;
}

.wd-replicas__ring {
  --p: 0;
  display: grid;
  place-items: center;
  width: 112px;
  height: 112px;
  border-radius: 50%;
  background: conic-gradient(#3b74f2 calc(var(--p) * 1%), #e8eef8 0);
}

.wd-replicas__inner {
  display: flex;
  align-items: baseline;
  justify-content: center;
  width: 88px;
  height: 88px;
  padding-top: 28px;
  border-radius: 50%;
  background: #fff;
}

.wd-replicas__inner strong {
  color: #1e293b;
  font-size: 26px;
}

.wd-replicas__inner span {
  margin-left: 2px;
  color: #94a3b8;
  font-size: 14px;
}

.wd-replicas__text {
  color: #64748b;
  font-size: 12px;
}

.wd-replicas__ctrl {
  display: flex;
  align-items: center;
  gap: 12px;
}

.wd-replicas__value {
  min-width: 28px;
  color: #1e293b;
  font-size: 18px;
  font-weight: 600;
  text-align: center;
}

.wd-replicas__value.is-changed {
  color: #3b74f2;
}

.wd-tip {
  margin: 0 0 14px;
  color: #64748b;
  font-size: 13px;
  line-height: 1.6;
}

.mono-input :deep(input) {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}

@media (max-width: 1100px) {
  .wd-attr {
    flex-direction: column;
  }
}
</style>
