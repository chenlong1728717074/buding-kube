<template>
  <div class="pod-list list-page">
    <PageHeader title="Pod管理" subtitle="查看和管理集群中的 Pod" icon="pod">
      <el-button type="primary" @click="createVisible = true">
        <el-icon><Plus /></el-icon>
        创建Pod
      </el-button>
      <el-button type="success" @click="openYamlAdd">
        <el-icon><Document /></el-icon>
        YAML添加
      </el-button>
    </PageHeader>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="命名空间">
          <InfiniteSelect
            v-model="searchForm.namespace"
            :fetch-data="namespaceFetcher"
            v-bind="namespaceSelectConfig"
            style="width: 200px"
            @change="handleSearch"
          />
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="searchForm.keyword" placeholder="输入Pod名称" clearable style="width: 200px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 130px" @change="handleSearch">
            <el-option v-for="o in POD_PHASE_OPTIONS" :key="o.value" :label="o.label" :value="o.value" />
          </el-select>
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
      <el-alert v-if="metricsMissing" type="info" show-icon :closable="false" class="metrics-alert">
        没有获取到容器组的实时用量，请确认集群已安装 metrics-server（kubectl top pod 可以正常使用）；新启动的容器组需要等待约 1 分钟才有数据。
      </el-alert>
      <el-table
        v-loading="loading"
        :data="podList"
        style="width: 100%"
        :header-cell-class-name="tableHeaderClass"
        :cell-class-name="tableCellClass"
        @sort-change="handleSortChange"
      >
        <el-table-column prop="name" label="Pod名称" min-width="220" header-align="center" align="center">
          <template #default="{ row }">
            <div class="name-cell">
              <el-link type="primary" :title="row.describe || ''" @click="openDetail(row)">{{ row.name }}</el-link>
              <span v-if="row.alias" class="name-cell__alias">{{ row.alias }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="140" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="podStatusTag(row.displayStatus || row.status)" size="small">{{ podStatusText(row.displayStatus || row.status) }}</el-tag>
            <span v-if="row.message" class="status-msg" :title="row.message">{{ row.message }}</span>
          </template>
        </el-table-column>
        <el-table-column label="就绪" min-width="70" header-align="center" align="center" prop="ready" />
        <el-table-column label="CPU" prop="cpu" min-width="130" header-align="center" align="center" sortable="custom">
          <template #default="{ row }">
            <ResourceUsageBar kind="cpu" :used="row.cpuUsage" :request="row.cpuRequest" :limit="row.cpuLimit" />
          </template>
        </el-table-column>
        <el-table-column label="内存" prop="memory" min-width="140" header-align="center" align="center" sortable="custom">
          <template #default="{ row }">
            <ResourceUsageBar kind="memory" :used="row.memoryUsage" :request="row.memoryRequest" :limit="row.memoryLimit" />
          </template>
        </el-table-column>
        <el-table-column label="重启次数" prop="restarts" min-width="100" header-align="center" align="center" sortable="custom">
          <template #default="{ row }">
            <span :class="{ 'warn-text': row.restarts > 0 }">{{ row.restarts }}</span>
          </template>
        </el-table-column>
        <el-table-column label="所属工作负载" min-width="170" header-align="center" align="center">
          <template #default="{ row }">
            <template v-if="row.ownerName">
              <el-link v-if="ownerDetailPath(row.ownerKind)" type="primary" @click="openOwner(row)">{{ row.ownerName }}</el-link>
              <span v-else>{{ row.ownerName }}</span>
              <span class="sub-text">{{ row.ownerKind }}</span>
            </template>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column label="节点" min-width="140" header-align="center" align="center">
          <template #default="{ row }">
            <el-link v-if="row.nodeName" type="primary" @click="openNode(row)">{{ row.nodeName }}</el-link>
            <el-text v-else type="info">未调度</el-text>
          </template>
        </el-table-column>
        <el-table-column label="容器组 IP" min-width="130" header-align="center" align="center">
          <template #default="{ row }"><span class="mono">{{ row.podIP || '-' }}</span></template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" min-width="120" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="openNamespace(row)">{{ row.namespace }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="160" header-align="center" align="center">
          <template #default="{ row }">{{ formatDateTime(row.creationTimestamp) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" header-align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" class="btn-main" @click="openDetail(row)">详情</el-button>
              <el-dropdown trigger="click" @command="(command: string) => handleMore(command, row)">
                <el-button size="small" text circle class="more-btn">
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="terminal" :disabled="row.status !== 'Running' || row.displayStatus === 'Terminating'">进入终端</el-dropdown-item>
                    <el-dropdown-item command="logs" :disabled="row.status === 'Pending'">查看日志</el-dropdown-item>
                    <el-dropdown-item command="yaml">编辑YAML</el-dropdown-item>
                    <el-dropdown-item command="expel" :disabled="row.displayStatus === 'Terminating'">驱逐</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-button size="small" type="danger" plain class="btn-danger" @click="handleDelete(row)">
                {{ row.displayStatus === 'Terminating' ? '强制删除' : '删除' }}
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <ListPagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        @size-change="handleSizeChange"
        @current-change="fetchList()"
      />
    </el-card>

    <PodCreateDialog v-model="createVisible" :cluster-id="clusterId" :default-namespace="searchForm.namespace" @success="handleCreated" />
    <YamlDialog
      v-model="yamlAddVisible"
      title="YAML添加Pod"
      subtitle="已内置示例，可直接修改后使用"
      v-model:content="yamlAddContent"
      :readonly="false"
      filename="pod.yaml"
      :submitting="yamlAddLoading"
      @confirm="confirmYamlAdd"
    />
    <YamlDialog
      v-model="yamlVisible"
      title="编辑YAML"
      subtitle="Pod 配置"
      :info="[{ label: '集群', value: clusterName }, { label: '命名空间', value: current?.namespace }, { label: '名称', value: current?.name }, { label: '类型', value: 'Pod' }]"
      v-model:content="yamlContent"
      :editor-title="`${current?.name} - Pod YAML`"
      :filename="`${current?.name}-pod.yaml`"
      tip="Pod 创建后只能修改镜像、标签、注解等少量字段，其余修改会被拒绝"
      :apply-target="current ? { clusterId, kind: 'pod', namespace: current.namespace, name: current.name } : null"
      @applied="fetchList()"
    />
    <PodTerminal v-model="terminalVisible" :cluster-id="clusterId" :namespace="current?.namespace || ''" :pod-name="current?.name || ''" />
    <PodLogViewer
      v-model="logVisible"
      :title="`查看日志 - ${current?.name || ''}`"
      :cluster-id="clusterId"
      :namespace="current?.namespace || ''"
      :pod-name="current?.name || ''"
      :container-selectable="false"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, MoreFilled, Plus, Refresh, Search } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import PodLogViewer from '@/components/PodLogViewer.vue'
import PodCreateDialog from '@/components/pod/PodCreateDialog.vue'
import PodTerminal from '@/components/pod/PodTerminal.vue'
import ResourceUsageBar from '@/components/ResourceUsageBar.vue'
import { podApi, type PodVO } from '@/api/pod'
import { fetchResourceYaml, resourceApi } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useLatestRequest } from '@/composables/useLatestRequest'
import { useNamespaceFetcher, namespaceSelectConfig } from '@/composables/useInfiniteSelect'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { formatDateTime, podStatusTag, podStatusText } from '@/utils/job'
import { ownerDetailPath } from '@/utils/workload'
import { POD_PHASE_OPTIONS, podYamlExample } from '@/utils/pod'
import '@/assets/styles/config-editor.css'

const route = useRoute()
const router = useRouter()
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)
const clusterName = computed(() => clusterStore.currentClusterName)
const namespaceFetcher = computed(() => useNamespaceFetcher(clusterId.value))

const loading = ref(false)
const podList = ref<PodVO[]>([])
// 从其他页面跳转过来时可以带上命名空间
const searchForm = reactive({ namespace: (route.query.namespace as string) || '', keyword: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })
// 排序在后端完成(分页前), 这样按 CPU / 内存排序能看到整个集群的前几名
const sort = reactive({ by: '', order: 'desc' as 'asc' | 'desc' })

// 有运行中的容器组却都没有用量数据, 说明没有安装 metrics-server
const metricsMissing = computed(
  () => podList.value.some(p => p.status === 'Running') && podList.value.every(p => p.cpuUsage === null || p.cpuUsage === undefined)
)

const handleSortChange = ({ prop, order }: { prop: string; order: 'ascending' | 'descending' | null }) => {
  sort.by = order ? prop : ''
  sort.order = order === 'ascending' ? 'asc' : 'desc'
  pagination.page = 1
  fetchList()
}

const createVisible = ref(false)
const yamlVisible = ref(false)
const yamlContent = ref('')
const yamlAddVisible = ref(false)
const yamlAddContent = ref('')
const yamlAddLoading = ref(false)
const logVisible = ref(false)
const terminalVisible = ref(false)
const current = ref<PodVO | null>(null)

// 快速切换条件时只采用最后一次查询的结果
const nextRequest = useLatestRequest()
const fetchList = async (silent = false) => {
  if (!clusterId.value) {
    podList.value = []
    pagination.total = 0
    return
  }
  if (!silent) loading.value = true
  const isLatest = nextRequest()
  try {
    const resp = await podApi.getList({
      clusterId: clusterId.value,
      namespace: searchForm.namespace || undefined,
      keyword: searchForm.keyword.trim() || undefined,
      status: searchForm.status || undefined,
      sortBy: sort.by || undefined,
      sortOrder: sort.by ? sort.order : undefined,
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    if (!isLatest()) return
    podList.value = resp.data?.items || []
    pagination.total = resp.data?.total || 0
  } catch {
    if (!isLatest()) return
    if (!silent) podList.value = []
  } finally {
    if (isLatest()) loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  Object.assign(searchForm, { namespace: '', keyword: '', status: '' })
  handleSearch()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  handleSearch()
}

const base = computed(() => `/cluster/${clusterId.value}`)
const openDetail = (row: { namespace: string; name: string }) =>
  router.push({ path: `${base.value}/pod/detail`, query: { namespace: row.namespace, name: row.name } })
const openNamespace = (row: PodVO) => router.push({ path: `${base.value}/namespace/detail`, query: { namespace: row.namespace } })
const openNode = (row: PodVO) => router.push({ path: `${base.value}/node/detail`, query: { hostname: row.nodeName } })
const openOwner = (row: PodVO) =>
  router.push({ path: `${base.value}/${ownerDetailPath(row.ownerKind)}`, query: { namespace: row.namespace, name: row.ownerName } })

const handleCreated = (namespace: string, name: string) => {
  fetchList()
  ElMessageBox.confirm('容器组已创建，是否查看详情？', '创建成功', { type: 'success', confirmButtonText: '查看详情', cancelButtonText: '留在列表' })
    .then(() => openDetail({ namespace, name }))
    .catch(() => {})
}

const handleMore = async (command: string, row: PodVO) => {
  current.value = row
  if (command === 'terminal') {
    terminalVisible.value = true
  } else if (command === 'logs') {
    logVisible.value = true
  } else if (command === 'yaml') {
    try {
      yamlContent.value = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'pod', namespace: row.namespace, name: row.name })
      yamlVisible.value = true
    } catch {
      // 请求拦截器已提示错误
    }
  } else if (command === 'expel') {
    try {
      await ElMessageBox.confirm(
        `驱逐会遵守 PodDisruptionBudget，由所属的工作负载重建容器组${row.ownerKind ? '' : '；该容器组没有所属工作负载，驱逐后不会重建'}。`,
        `驱逐 "${row.name}"`,
        { type: 'warning', confirmButtonText: '驱逐', cancelButtonText: '取消' }
      )
      await podApi.expel({ clusterId: clusterId.value, namespace: row.namespace, name: row.name })
      ElMessage.success('已发起驱逐')
      fetchList()
    } catch {
      // 取消或请求失败
    }
  }
}

const openYamlAdd = () => {
  yamlAddContent.value = podYamlExample(searchForm.namespace || 'default')
  yamlAddVisible.value = true
}

const confirmYamlAdd = async () => {
  if (!yamlAddContent.value.trim()) {
    ElMessage.warning('请输入YAML内容')
    return
  }
  yamlAddLoading.value = true
  try {
    await resourceApi.apply({ clusterId: clusterId.value, kind: 'pod', namespace: searchForm.namespace || undefined, yaml: yamlAddContent.value })
    ElMessage.success('创建成功')
    yamlAddVisible.value = false
    fetchList()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    yamlAddLoading.value = false
  }
}

const handleDelete = async (row: PodVO) => {
  const force = row.displayStatus === 'Terminating'
  try {
    await ElMessageBox.confirm(
      force
        ? `容器组 "${row.name}" 一直处于删除中，强制删除会立即移除记录，不再等待容器退出。确定强制删除吗？`
        : `确定要删除容器组 "${row.name}" 吗？${row.ownerKind ? '它由工作负载管理，删除后会自动重建。' : '删除后不会重建。'}`,
      force ? '强制删除' : '确认删除',
      { type: 'warning', confirmButtonText: force ? '强制删除' : '确定删除', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await podApi.delete({ clusterId: clusterId.value, namespace: row.namespace, name: row.name, force })
    ElMessage.success(force ? '已强制删除' : '已开始删除')
    if (podList.value.length === 1 && pagination.page > 1) pagination.page--
    fetchList()
  } catch {
    // 请求拦截器已提示错误
  }
}

watch(
  clusterId,
  (id, old) => {
    if (old !== undefined) searchForm.namespace = ''
    pagination.page = 1
    if (id) fetchList()
    else {
      podList.value = []
      pagination.total = 0
    }
  },
  { immediate: true }
)

// 从详情页返回时静默刷新(详情页里可能删除或驱逐了容器组); 首次进入时 watch 已经加载过
let activated = false
onActivated(() => {
  if (activated) fetchList(true)
  activated = true
})
</script>

<style scoped>
.pod-list {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.search-card {
  margin-bottom: 20px;
}

.table-card {
  min-height: 400px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.yaml-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.yaml-info {
  margin-bottom: 16px;
}

.yaml-editor-wrapper {
  flex: 1;
}

.dialog-footer {
  text-align: right;
}
.action-buttons {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  flex-wrap: nowrap;
}

.btn-main,
.btn-edit {
  font-weight: 500;
}

.btn-more {
  color: #475569;
  border-color: #d5deea;
  background: #fff;
}

.btn-more:hover,
.btn-more:focus {
  color: #334155;
  border-color: #b8c4d5;
  background: #f8fafc;
}

:deep(.action-buttons .el-button) {
  height: 28px;
  padding: 0 10px;
  border-radius: 8px;
}

.name-cell {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  line-height: 1.3;
}

.name-cell__alias {
  color: #94a3b8;
  font-size: 12px;
}

.status-msg {
  display: block;
  max-width: 220px;
  margin: 2px auto 0;
  overflow: hidden;
  color: #ef4444;
  font-size: 12px;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.sub-text {
  display: block;
  color: #94a3b8;
  font-size: 12px;
}

.mono {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}

.metrics-alert {
  margin-bottom: 12px;
}

.warn-text {
  color: #f59e0b;
  font-weight: 600;
}
</style>
