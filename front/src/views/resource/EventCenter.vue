<template>
  <div class="event-center list-page">
    <PageHeader title="事件中心" subtitle="查看集群中所有资源的事件，事件默认只保留 1 小时（由 apiserver 的 --event-ttl 决定）" icon="etcd">
      <el-switch v-model="autoRefresh" active-text="自动刷新" />
      <el-button @click="fetchList()">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </PageHeader>

    <el-card class="search-card">
      <el-form :model="searchForm" inline @submit.prevent>
        <el-form-item label="命名空间">
          <InfiniteSelect
            v-model="searchForm.namespace"
            :fetch-data="namespaceFetcher"
            v-bind="namespaceSelectConfig"
            style="width: 200px"
            @change="handleSearch"
          />
        </el-form-item>
        <el-form-item label="类型">
          <el-radio-group v-model="searchForm.type" @change="handleSearch">
            <el-radio-button value="">全部</el-radio-button>
            <el-radio-button value="Warning">Warning</el-radio-button>
            <el-radio-button value="Normal">Normal</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="资源类型">
          <el-select v-model="searchForm.kind" clearable filterable allow-create placeholder="全部" style="width: 160px" @change="handleSearch">
            <el-option v-for="k in KINDS" :key="k" :label="k" :value="k" />
          </el-select>
        </el-form-item>
        <el-form-item label="资源名称">
          <el-input v-model="searchForm.name" placeholder="精确匹配" clearable style="width: 180px" @keyup.enter="handleSearch" @clear="handleSearch" />
        </el-form-item>
        <el-form-item label="关键字">
          <el-input v-model="searchForm.keyword" placeholder="原因 / 消息" clearable style="width: 180px" @keyup.enter="handleSearch" @clear="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table v-loading="loading" :data="events" style="width: 100%" :header-cell-class-name="tableHeaderClass" :row-class-name="rowClass">
        <el-table-column label="类型" width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="row.type === 'Warning' ? 'warning' : 'success'" size="small" effect="light">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="原因" prop="reason" min-width="150" show-overflow-tooltip />
        <el-table-column label="资源" min-width="230">
          <template #default="{ row }">
            <div class="ec-object">
              <span class="ec-kind">{{ row.objectKind }}</span>
              <el-link v-if="detailRoute(row)" type="primary" @click="goObject(row)">{{ row.objectName }}</el-link>
              <span v-else>{{ row.objectName }}</span>
            </div>
            <span v-if="row.namespace" class="ec-ns">{{ row.namespace }}</span>
          </template>
        </el-table-column>
        <el-table-column label="消息" prop="message" min-width="360">
          <template #default="{ row }"><span class="ec-message">{{ row.message }}</span></template>
        </el-table-column>
        <el-table-column label="来源" prop="source" min-width="140" show-overflow-tooltip />
        <el-table-column label="次数" prop="count" width="80" header-align="center" align="center" />
        <el-table-column label="最后发生" min-width="150" header-align="center" align="center">
          <template #default="{ row }">
            <el-tooltip :content="`首次 ${formatDateTime(row.firstTime)}，最后 ${formatDateTime(row.lastTime)}`" placement="top">
              <span>{{ timeAgo(row.lastTime) }}</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="没有符合条件的事件" :image-size="80" />
        </template>
      </el-table>

      <ListPagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[20, 50, 100, 200]"
        @size-change="handleSearch"
        @current-change="fetchList()"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Refresh, Search } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import { eventApi } from '@/api/event'
import type { EventVO } from '@/api/job'
import { useClusterStore } from '@/stores/cluster'
import { useNamespaceFetcher, namespaceSelectConfig } from '@/composables/useInfiniteSelect'
import { tableHeaderClass } from '@/utils/table'
import { formatDateTime } from '@/utils/job'
import { objectDetailRoute, timeAgo } from '@/utils/navigation'

const KINDS = ['Pod', 'Deployment', 'ReplicaSet', 'StatefulSet', 'DaemonSet', 'Job', 'CronJob', 'Service', 'Ingress', 'PersistentVolumeClaim', 'PersistentVolume', 'Node', 'HorizontalPodAutoscaler', 'Namespace']

const route = useRoute()
const router = useRouter()
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)
const namespaceFetcher = computed(() => useNamespaceFetcher(clusterId.value))

const loading = ref(false)
const events = ref<EventVO[]>([])
// 默认只看 Warning, 可以通过地址参数指定
const searchForm = reactive({
  namespace: (route.query.namespace as string) || '',
  type: route.query.type !== undefined ? (route.query.type as string) : 'Warning',
  kind: (route.query.kind as string) || '',
  name: (route.query.name as string) || '',
  keyword: ''
})
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const autoRefresh = ref(false)

const fetchList = async (silent = false) => {
  if (!clusterId.value) return
  if (!silent) loading.value = true
  try {
    const resp = await eventApi.list({
      clusterId: clusterId.value,
      namespace: searchForm.namespace || undefined,
      type: searchForm.type || undefined,
      kind: searchForm.kind || undefined,
      name: searchForm.name.trim() || undefined,
      keyword: searchForm.keyword.trim() || undefined,
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    events.value = resp.data?.items || []
    pagination.total = resp.data?.total || 0
  } catch {
    events.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  Object.assign(searchForm, { namespace: '', type: 'Warning', kind: '', name: '', keyword: '' })
  handleSearch()
}

const rowClass = ({ row }: { row: EventVO }) => (row.type === 'Warning' ? 'ec-row--warning' : '')

const detailRoute = (row: EventVO) => objectDetailRoute(clusterId.value, row.objectKind, row.objectNamespace || row.namespace, row.objectName)
const goObject = (row: EventVO) => {
  const target = detailRoute(row)
  if (target) router.push(target)
}

// 自动刷新: 每 10 秒, 页面切走后暂停
let timer: ReturnType<typeof setInterval> | undefined
const stopTimer = () => clearInterval(timer)
const startTimer = () => {
  stopTimer()
  if (autoRefresh.value) timer = setInterval(() => fetchList(true), 10000)
}
watch(autoRefresh, startTimer)
onActivated(startTimer)
onDeactivated(stopTimer)
onBeforeUnmount(stopTimer)

watch(
  clusterId,
  (id, old) => {
    if (old && id !== old) searchForm.namespace = ''
    pagination.page = 1
    fetchList()
  },
  { immediate: true }
)
</script>

<style scoped>
.ec-object {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.ec-kind {
  flex-shrink: 0;
  padding: 0 6px;
  border-radius: 4px;
  background: #eef2f7;
  color: #475569;
  font-size: 12px;
}

.ec-ns {
  display: block;
  margin-top: 2px;
  color: #94a3b8;
  font-size: 12px;
}

.ec-message {
  color: #334155;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

:deep(.ec-row--warning td.el-table__cell:first-child) {
  box-shadow: inset 3px 0 0 #f59e0b;
}
</style>
