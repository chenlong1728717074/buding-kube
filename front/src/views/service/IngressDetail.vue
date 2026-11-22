<template>
  <div class="job-detail detail-page">
    <PageHeader
      :title="ing?.name || name || '应用路由详情'"
      :subtitle="ing ? ing.alias || ing.describe || `命名空间：${ing.namespace}` : '应用路由详情'"
      icon="ing"
      back
      @back="goBack"
    >
      <template v-if="ing" #meta>
        <el-tag v-if="ing.tls" type="success" effect="dark">HTTPS</el-tag>
        <el-tag v-else type="primary" effect="dark">HTTP</el-tag>
        <el-tag v-if="!ing.addresses.length" type="warning" effect="light">等待分配地址</el-tag>
      </template>
      <template v-if="ing">
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

    <div v-if="loading && !ing" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="ing" class="detail-content">
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
              <span><el-link type="primary" @click="goNamespace">{{ ing.namespace }}</el-link></span>
            </div>
            <div class="jd-attr"><label>Ingress 类</label><span>{{ ing.ingressClassName || '集群默认' }}</span></div>
            <div class="jd-attr"><label>访问地址</label><span class="mono">{{ ing.addresses.join(', ') || '等待控制器分配' }}</span></div>
            <div class="jd-attr"><label>域名</label><span class="mono">{{ ing.hosts.join(', ') || '-' }}</span></div>
            <div class="jd-attr"><label>转发路径</label><span>{{ ing.ruleCount }}</span></div>
            <div class="jd-attr"><label>HTTPS</label><span>{{ ing.tls ? `已启用（${ing.tlsList.length} 个证书）` : '未启用' }}</span></div>
            <div class="jd-attr"><label>默认后端</label><span class="mono">{{ ing.defaultBackend || '-' }}</span></div>
            <div class="jd-attr"><label>创建时间</label><span>{{ formatDateTime(ing.createTime) }}</span></div>
            <div class="jd-attr"><label>创建者</label><span>{{ ing.creator || '-' }}</span></div>
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
              <span>路由规则</span>
              <el-tag size="small" type="info">{{ ing.rules.length }}</el-tag>
              <span class="card-header__hint">域名需要解析到访问地址，或在本地 hosts 中配置</span>
            </div>
          </template>
          <el-table :data="ing.rules" style="width: 100%" :header-cell-class-name="tableHeaderClass">
            <el-table-column label="域名" min-width="200">
              <template #default="{ row }"><span class="mono">{{ row.host || '*（所有域名）' }}</span></template>
            </el-table-column>
            <el-table-column label="路径" min-width="160">
              <template #default="{ row }"><span class="mono">{{ row.path || '/' }}</span></template>
            </el-table-column>
            <el-table-column label="匹配方式" width="150">
              <template #default="{ row }">{{ pathTypeText(row.pathType) }}</template>
            </el-table-column>
            <el-table-column label="服务" min-width="200">
              <template #default="{ row }">
                <el-link v-if="row.serviceExists" type="primary" @click="goService(row.serviceName)">{{ row.serviceName }}</el-link>
                <span v-else class="ig-missing" title="服务不存在，请求会返回 503">{{ row.serviceName }}（不存在）</span>
              </template>
            </el-table-column>
            <el-table-column label="端口" width="100" prop="servicePort" />
            <el-table-column label="访问链接" min-width="280">
              <template #default="{ row }">
                <a v-if="row.host && !row.host.startsWith('*')" class="ig-url mono" :href="urlOf(row)" target="_blank" rel="noopener">{{ urlOf(row) }}</a>
                <span v-else class="jd-muted">-</span>
              </template>
            </el-table-column>
            <template #empty>
              <span class="jd-muted">没有路由规则</span>
            </template>
          </el-table>
        </el-card>
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>HTTPS 证书</span>
              <el-tag size="small" type="info">{{ ing.tlsList.length }}</el-tag>
            </div>
          </template>
          <el-table :data="ing.tlsList" style="width: 100%" :header-cell-class-name="tableHeaderClass">
            <el-table-column label="证书（Secret）" min-width="220">
              <template #default="{ row }"><el-link type="primary" @click="goSecret(row.secretName)">{{ row.secretName }}</el-link></template>
            </el-table-column>
            <el-table-column label="域名" min-width="320">
              <template #default="{ row }"><span class="mono">{{ row.hosts.join(', ') || '所有域名' }}</span></template>
            </el-table-column>
            <template #empty>
              <span class="jd-muted">没有配置 HTTPS</span>
            </template>
          </el-table>
        </el-card>
      </template>

      <el-card v-else-if="activeTab === 'metadata'" class="info-card">
        <template #header>
          <div class="card-header"><span>元数据</span></div>
        </template>
        <MetadataPanel :labels="ing.labels" :annotations="ing.annotations" />
      </el-card>

      <el-card v-else class="info-card">
        <template #header>
          <div class="card-header">
            <span>事件</span>
            <el-tag size="small" type="info">{{ ing.events.length }}</el-tag>
          </div>
        </template>
        <EventsTable :events="ing.events" />
      </el-card>
    </div>

    <el-empty v-else description="应用路由不存在或已被删除" />

    <IngressFormDialog v-model="editVisible" :cluster-id="clusterId" :edit-target="ing ? { namespace: ing.namespace, name: ing.name } : null" @success="fetchDetail" />
    <YamlDialog
      v-model="yamlVisible"
      v-model:content="yamlContent"
      title="编辑YAML"
      subtitle="Ingress 配置"
      :info="[{ label: '集群', value: clusterName }, { label: '命名空间', value: ing?.namespace }, { label: '名称', value: ing?.name }, { label: '类型', value: 'Ingress' }]"
      :editor-title="`${ing?.name} - Ingress YAML`"
      :filename="`${ing?.name}-ingress.yaml`"
      :apply-target="ing ? { clusterId, kind: 'ingress', namespace: ing.namespace, name: ing.name } : null"
      @applied="fetchDetail"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Delete, Document, Edit, Refresh } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import MetadataPanel from '@/components/job/MetadataPanel.vue'
import EventsTable from '@/components/job/EventsTable.vue'
import IngressFormDialog from '@/components/service/IngressFormDialog.vue'
import { ingressApi, type IngressDetailVO, type IngressRuleVO } from '@/api/ingress'
import { fetchResourceYaml } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useTabBack } from '@/composables/useTabBack'
import { tableHeaderClass } from '@/utils/table'
import { formatDateTime } from '@/utils/job'
import { PATH_TYPE_OPTIONS } from '@/utils/network'
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
const ing = ref<IngressDetailVO | null>(null)
const attrOpen = ref(true)
const activeTab = ref<TabKey>('resources')
const editVisible = ref(false)
const yamlVisible = ref(false)
const yamlContent = ref('')

const tabs = computed(() => [
  { key: 'resources' as TabKey, label: '资源状态', count: ing.value?.rules.length },
  { key: 'metadata' as TabKey, label: '元数据' },
  { key: 'events' as TabKey, label: '事件', count: ing.value?.events.length }
])

const pathTypeText = (t?: string) => PATH_TYPE_OPTIONS.find(o => o.value === t)?.label.replace(/（.*）/, '') || t || '-'

const urlOf = (row: IngressRuleVO) => {
  const https = ing.value?.tlsList.some(t => t.hosts.includes(row.host) || !t.hosts.length)
  return `${https ? 'https' : 'http'}://${row.host}${row.path || '/'}`
}

const fetchDetail = async () => {
  if (!clusterId.value || !namespace.value || !name.value) return
  loading.value = true
  try {
    const resp = await ingressApi.getDetail({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })
    ing.value = resp.code === 200 ? resp.data : null
  } catch {
    ing.value = null
  } finally {
    loading.value = false
  }
}

const base = `/cluster/${clusterId.value}`
const goNamespace = () => router.push({ path: `${base}/namespace/detail`, query: { namespace: namespace.value } })
const goService = (svc: string) => router.push({ path: `${base}/service/service/detail`, query: { namespace: namespace.value, name: svc } })
const goSecret = (secret: string) => router.push({ path: `${base}/config/secret/detail`, query: { namespace: namespace.value, name: secret } })

const openYaml = async () => {
  try {
    yamlContent.value = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'ingress', namespace: namespace.value, name: name.value })
    yamlVisible.value = true
  } catch {
    // 请求拦截器已提示错误
  }
}

const handleDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除应用路由 "${name.value}" 吗？此操作不可恢复。`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await ingressApi.deleteIngress({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })
    ElMessage.success('删除成功')
    goBack()
  } catch {
    // 请求拦截器已提示错误
  }
}

onMounted(fetchDetail)
</script>

<style scoped>
.ig-missing {
  color: #ef4444;
}

.ig-url {
  color: #3b74f2;
  text-decoration: none;
}

.ig-url:hover {
  text-decoration: underline;
}
</style>
