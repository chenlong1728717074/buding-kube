<template>
  <div class="namespace-detail detail-page job-detail">
    <PageHeader :title="ns?.name || namespace" subtitle="命名空间详情" icon="ns" back @back="goBack">
      <template v-if="ns" #meta>
        <el-tag :type="ns.active === 'Active' ? 'success' : 'warning'" effect="light">{{ ns.active === 'Active' ? '活跃' : ns.active === 'Terminating' ? '删除中' : ns.active }}</el-tag>
      </template>
      <el-button @click="refresh">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
      <el-button :disabled="!ns" @click="editVisible = true">
        <el-icon><Edit /></el-icon>
        编辑信息
      </el-button>
      <el-button :disabled="!ns" @click="openYaml">
        <el-icon><Document /></el-icon>
        编辑YAML
      </el-button>
      <el-button type="danger" :disabled="!ns" @click="deleteVisible = true">
        <el-icon><Delete /></el-icon>
        删除
      </el-button>
    </PageHeader>

    <div v-if="loading && !ns" class="loading-container">
      <el-skeleton :rows="8" animated />
    </div>

    <div v-else-if="ns" class="detail-content">
      <el-card class="info-card jd-attr-card">
        <template #header>
          <div class="card-header" @click="attrOpen = !attrOpen">
            <span>属性</span>
            <el-icon :class="['jd-collapse', { 'is-collapsed': !attrOpen }]"><ArrowDown /></el-icon>
          </div>
        </template>
        <el-collapse-transition>
          <div v-show="attrOpen" class="jd-attr-grid">
            <div class="jd-attr"><label>名称</label><span>{{ ns.name }}</span></div>
            <div class="jd-attr"><label>别名</label><span>{{ ns.alias || '-' }}</span></div>
            <div class="jd-attr"><label>描述</label><span>{{ ns.describe || '-' }}</span></div>
            <div class="jd-attr"><label>集群</label><span>{{ clusterName || '-' }}</span></div>
            <div class="jd-attr"><label>创建时间</label><span>{{ formatDateTime(ns.creationTimestamp) }}</span></div>
            <div class="jd-attr"><label>Finalizers</label><span>{{ ns.spec?.finalizers?.join(', ') || '-' }}</span></div>
            <div class="jd-attr"><label>UID</label><span class="mono">{{ ns.uid }}</span></div>
            <div class="jd-attr">
              <label>事件</label>
              <span><el-link type="primary" @click="goEventCenter">在事件中心查看全部事件</el-link></span>
            </div>
          </div>
        </el-collapse-transition>
      </el-card>

      <div class="jd-tabbar">
        <button v-for="t in tabs" :key="t.key" type="button" :class="['jd-tab', { 'is-active': activeTab === t.key }]" @click="activeTab = t.key">
          {{ t.label }}
          <span v-if="t.count !== undefined && t.count !== null" class="jd-tab__count">{{ t.count }}</span>
        </button>
      </div>

      <!-- 切换页签时保留各面板状态, 避免重复请求 -->
      <el-card v-show="activeTab === 'quota'" class="info-card">
        <template #header>
          <div class="card-header">
            <span>资源配额</span>
            <span class="card-header__hint">限制命名空间可以使用的资源总量（ResourceQuota）</span>
          </div>
        </template>
        <QuotaPanel :key="`q${panelKey}`" :cluster-id="clusterId" :namespace="namespace" @change="(n: number) => (counts.quota = n)" />
      </el-card>
      <el-card v-show="activeTab === 'limits'" class="info-card">
        <template #header>
          <div class="card-header">
            <span>默认资源限制</span>
            <span class="card-header__hint">容器的默认资源预留 / 上限及允许范围（LimitRange）</span>
          </div>
        </template>
        <LimitRangePanel :key="`l${panelKey}`" :cluster-id="clusterId" :namespace="namespace" @change="(n: number) => (counts.limits = n)" />
      </el-card>
      <el-card v-show="activeTab === 'network'" class="info-card">
        <template #header>
          <div class="card-header">
            <span>网络策略</span>
            <span class="card-header__hint">控制容器组的入站 / 出站流量（NetworkPolicy），需要网络插件支持</span>
          </div>
        </template>
        <NetworkPolicyPanel :key="`n${panelKey}`" :cluster-id="clusterId" :namespace="namespace" @change="(n: number) => (counts.network = n)" />
      </el-card>
      <el-card v-show="activeTab === 'metadata'" class="info-card">
        <template #header>
          <div class="card-header"><span>元数据</span></div>
        </template>
        <MetadataPanel :labels="ns.labels" :annotations="ns.annotations" />
      </el-card>
      <el-card v-show="activeTab === 'events'" class="info-card">
        <template #header>
          <div class="card-header">
            <span>事件</span>
            <el-tag size="small" type="info">{{ events.length }}</el-tag>
            <span class="card-header__hint">命名空间对象本身的事件，命名空间内其他资源的事件请到事件中心查看</span>
          </div>
        </template>
        <EventsTable :events="events" />
      </el-card>
    </div>

    <el-empty v-else description="命名空间不存在或已被删除" />

    <YamlDialog
      v-model="yamlVisible"
      title="编辑YAML"
      subtitle="Namespace 配置"
      :info="[{ label: '集群', value: clusterName }, { label: '命名空间', value: namespace }, { label: '资源类型', value: 'Namespace' }]"
      :content="yamlContent"
      :filename="`${namespace}.yaml`"
      :apply-target="{ clusterId, kind: 'namespace', name: namespace }"
      @applied="fetchDetail(true)"
    />

    <NamespaceFormDialog v-model="editVisible" :cluster-id="clusterId" :namespace="editingNamespace" @success="fetchDetail(true)" />

    <DeleteConfirmDialog
      v-model="deleteVisible"
      :item-name="namespace"
      message="删除命名空间会同时删除其中的所有资源，确定要删除吗？"
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="deleteVisible = false"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowDown, Delete, Document, Edit, Refresh } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import NamespaceFormDialog from '@/components/NamespaceFormDialog.vue'
import MetadataPanel from '@/components/job/MetadataPanel.vue'
import EventsTable from '@/components/job/EventsTable.vue'
import QuotaPanel from '@/components/policy/QuotaPanel.vue'
import LimitRangePanel from '@/components/policy/LimitRangePanel.vue'
import NetworkPolicyPanel from '@/components/policy/NetworkPolicyPanel.vue'
import { namespaceApi, type NamespaceVO } from '@/api/namespace'
import { fetchResourceYaml } from '@/api/resource'
import { eventApi } from '@/api/event'
import type { EventVO } from '@/api/job'
import { useClusterStore } from '@/stores/cluster'
import { useTabBack } from '@/composables/useTabBack'
import { formatDateTime } from '@/utils/job'
import '@/assets/styles/job-detail.css'

type TabKey = 'quota' | 'limits' | 'network' | 'metadata' | 'events'

const route = useRoute()
const router = useRouter()
const goBack = useTabBack()
const clusterStore = useClusterStore()

// 详情页按完整路径缓存, 参数在创建时取一次
const clusterId = ref((route.query.clusterId as string) || (route.params.clusterId as string) || clusterStore.currentClusterId || '')
const namespace = ref((route.query.namespace as string) || '')
const clusterName = computed(() => clusterStore.currentClusterName)

const loading = ref(false)
const ns = ref<NamespaceVO | null>(null)
const events = ref<EventVO[]>([])
const attrOpen = ref(true)
const activeTab = ref<TabKey>('quota')
const counts = reactive<{ quota: number | null; limits: number | null; network: number | null }>({ quota: null, limits: null, network: null })
const tabs = computed(() => [
  { key: 'quota' as TabKey, label: '资源配额', count: counts.quota },
  { key: 'limits' as TabKey, label: '默认资源限制', count: counts.limits },
  { key: 'network' as TabKey, label: '网络策略', count: counts.network },
  { key: 'metadata' as TabKey, label: '元数据' },
  { key: 'events' as TabKey, label: '事件', count: events.value.length }
])

const editVisible = ref(false)
const editingNamespace = computed(() => (ns.value ? { name: ns.value.name, alias: ns.value.alias, describe: ns.value.describe } : null))

const fetchDetail = async (silent = false) => {
  if (!clusterId.value || !namespace.value) return
  if (!silent) loading.value = true
  try {
    const resp = await namespaceApi.getInfo({ clusterId: clusterId.value, namespace: namespace.value })
    ns.value = resp.data
    const ev = await eventApi.list({ clusterId: clusterId.value, kind: 'Namespace', name: namespace.value, page: 1, pageSize: 200 })
    events.value = ev.data?.items || []
  } catch {
    if (!silent) ns.value = null
  } finally {
    loading.value = false
  }
}

// 各面板在挂载时加载, 刷新时通过改变 key 重新挂载
const panelKey = ref(0)
const refresh = () => {
  fetchDetail(true)
  panelKey.value++
}

const yamlVisible = ref(false)
const yamlContent = ref('')
const openYaml = async () => {
  try {
    yamlContent.value = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'namespace', name: namespace.value })
    yamlVisible.value = true
  } catch {
    // 请求拦截器已提示错误
  }
}

const goEventCenter = () => router.push({ path: `/cluster/${clusterId.value}/resource/events`, query: { namespace: namespace.value } })

const deleteVisible = ref(false)
const deleting = ref(false)
const confirmDelete = async () => {
  deleting.value = true
  try {
    await namespaceApi.delete({ clusterId: clusterId.value, namespace: namespace.value })
    ElMessage.success('已提交删除，命名空间中的资源会被逐步清理')
    deleteVisible.value = false
    goBack()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    deleting.value = false
  }
}

onMounted(() => {
  if (!clusterId.value || !namespace.value) {
    ElMessage.error('缺少必要参数')
    goBack()
    return
  }
  fetchDetail()
})
</script>

<style scoped>
.mono {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12px;
}
</style>
