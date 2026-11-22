<template>
  <div class="resource-table-page list-page">
    <PageHeader :title="title" :subtitle="subtitle" :icon="icon" :back="back" @back="emit('back')">
      <el-button v-if="createText" type="primary" @click="emit('create', searchForm.namespace)">
        <el-icon><Plus /></el-icon>
        {{ createText }}
      </el-button>
      <el-button v-if="yamlTemplate" type="success" @click="openYamlAdd">
        <el-icon><Document /></el-icon>
        YAML添加
      </el-button>
      <slot name="header" />
    </PageHeader>

    <!-- 页头下方的提示 -->
    <slot name="notice" />

    <el-card class="search-card">
      <el-form :model="searchForm" inline @submit.prevent>
        <el-form-item v-if="namespaced" label="命名空间">
          <InfiniteSelect
            v-model="searchForm.namespace"
            :fetch-data="namespaceFetcher"
            v-bind="namespaceSelectConfig"
            style="width: 200px"
            @change="handleSearch"
          />
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="searchForm.keyword" placeholder="输入名称" clearable style="width: 200px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="searchForm.labelSelector" placeholder="例如 app=nginx" clearable style="width: 200px" @keyup.enter="handleSearch" />
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
        <el-form-item v-if="hasWideColumns" class="rt-wide">
          <el-switch v-model="wide" active-text="显示更多列" />
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-alert v-if="loadError" type="warning" show-icon :closable="false" class="rt-alert">{{ loadError }}</el-alert>
      <el-table v-loading="loading" :data="rows" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
        <el-table-column label="名称" min-width="200" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="openDetail(row)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column v-if="namespaced" label="命名空间" min-width="130" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="openNamespace(row.namespace)">{{ row.namespace }}</el-link>
          </template>
        </el-table-column>
        <el-table-column
          v-for="col in columns"
          :key="col.name"
          :label="columnText(col)"
          min-width="130"
          header-align="center"
          align="center"
          show-overflow-tooltip
        >
          <template #header>
            <el-tooltip :content="col.description" :disabled="!col.description" placement="top" :show-after="300">
              <span>{{ columnText(col) }}</span>
            </el-tooltip>
          </template>
          <template #default="{ row }">
            <span :class="{ mono: col.type !== 'string' || col.format === 'name' }">{{ formatCell(row.cells[col.index]) }}</span>
          </template>
        </el-table-column>
        <el-table-column v-if="showOwner" label="所属" min-width="170" header-align="center" align="center">
          <template #default="{ row }">
            <template v-if="row.ownerName">
              <el-link v-if="ownerRoute(row)" type="primary" @click="goOwner(row)">{{ row.ownerName }}</el-link>
              <span v-else>{{ row.ownerName }}</span>
              <span class="sub-text">{{ row.ownerKind }}</span>
            </template>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="160" header-align="center" align="center">
          <template #default="{ row }">{{ formatDateTime(row.createTime) }}</template>
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
                    <el-dropdown-item v-for="a in actions" :key="a.command" :command="a.command">{{ a.label }}</el-dropdown-item>
                    <el-dropdown-item command="__yaml">编辑YAML</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-button size="small" type="danger" plain class="btn-danger" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty :description="emptyText" :image-size="80" />
        </template>
      </el-table>

      <ListPagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        @size-change="handleSearch"
        @current-change="fetchList()"
      />
    </el-card>

    <ResourceDetailDrawer
      v-model="detailVisible"
      :cluster-id="clusterId"
      :resource="resource"
      :kind-label="kindLabel"
      :namespace="current?.namespace"
      :name="current?.name || ''"
      @edit="current && openYaml(current)"
    >
      <template #summary="{ obj }">
        <slot name="summary" :obj="obj" />
      </template>
    </ResourceDetailDrawer>

    <YamlDialog
      v-model="yamlAddVisible"
      :title="`YAML添加${kindLabel}`"
      subtitle="已内置示例，可直接修改后使用"
      v-model:content="yamlAddContent"
      :readonly="false"
      :filename="`${kindLabel.toLowerCase()}.yaml`"
      :submitting="yamlAddLoading"
      @confirm="confirmYamlAdd"
    />
    <YamlDialog
      v-model="yamlVisible"
      title="编辑YAML"
      :subtitle="`${kindLabel} 配置`"
      :info="[{ label: '集群', value: clusterName }, ...(current?.namespace ? [{ label: '命名空间', value: current.namespace }] : []), { label: '名称', value: current?.name }, { label: '类型', value: kindLabel }]"
      v-model:content="yamlContent"
      :editor-title="`${current?.name} - ${kindLabel} YAML`"
      :filename="`${current?.name}.yaml`"
      :apply-target="current ? { ...resource, clusterId, namespace: current.namespace || undefined, name: current.name } : null"
      @applied="fetchList(true)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, MoreFilled, Plus, Refresh, Search } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import type { K8sIconName } from '@/components/K8sIcon.vue'
import ListPagination from '@/components/ListPagination.vue'
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import ResourceDetailDrawer from '@/components/resource/ResourceDetailDrawer.vue'
import { fetchResourceYaml, resourceApi, type ResourceRef, type ResourceTableColumn, type ResourceTableRow } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useLatestRequest } from '@/composables/useLatestRequest'
import { useNamespaceFetcher, namespaceSelectConfig } from '@/composables/useInfiniteSelect'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { formatDateTime } from '@/utils/job'
import { columnText, formatCell, visibleColumns, type ResourceAction } from '@/utils/resourceTable'
import { objectDetailRoute } from '@/utils/navigation'
import '@/assets/styles/config-editor.css'

// 通用资源列表: 列与 kubectl get 一致(由 apiserver 的 Table 格式提供), 支持详情、YAML 添加 / 编辑和删除.
// 有专门表单的资源通过 createText / actions 交给页面处理

const props = withDefaults(
  defineProps<{
    resource: ResourceRef
    // 资源类型名称, 如 ReplicaSet
    kindLabel: string
    title: string
    subtitle?: string
    icon?: K8sIconName
    namespaced?: boolean
    yamlTemplate?: (namespace: string) => string
    createText?: string
    actions?: ResourceAction[]
    emptyText?: string
    back?: boolean
  }>(),
  { subtitle: '', icon: 'k8s', namespaced: true, yamlTemplate: undefined, createText: '', actions: () => [], emptyText: '暂无数据', back: false }
)

const emit = defineEmits<{
  create: [namespace: string]
  action: [command: string, row: ResourceTableRow]
  back: []
}>()

const route = useRoute()
const router = useRouter()
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)
const clusterName = computed(() => clusterStore.currentClusterName)
const namespaceFetcher = computed(() => useNamespaceFetcher(clusterId.value))

const loading = ref(false)
const loadError = ref('')
const rows = ref<ResourceTableRow[]>([])
const allColumns = ref<ResourceTableColumn[]>([])
const wide = ref(false)
const searchForm = reactive({ namespace: (route.query.namespace as string) || '', keyword: '', labelSelector: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const columns = computed(() => visibleColumns(allColumns.value, wide.value))
const hasWideColumns = computed(() => allColumns.value.some(c => c.priority > 0))
const showOwner = computed(() => rows.value.some(r => r.ownerName))

const current = ref<ResourceTableRow | null>(null)
const detailVisible = ref(false)
const yamlVisible = ref(false)
const yamlContent = ref('')
const yamlAddVisible = ref(false)
const yamlAddContent = ref('')
const yamlAddLoading = ref(false)

// 快速切换条件时只采用最后一次查询的结果
const nextRequest = useLatestRequest()
const fetchList = async (silent = false) => {
  if (!clusterId.value) {
    rows.value = []
    pagination.total = 0
    return
  }
  if (!silent) loading.value = true
  const isLatest = nextRequest()
  try {
    const resp = await resourceApi.table({
      ...props.resource,
      clusterId: clusterId.value,
      namespace: props.namespaced ? searchForm.namespace || undefined : undefined,
      keyword: searchForm.keyword.trim() || undefined,
      labelSelector: searchForm.labelSelector.trim() || undefined,
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    if (!isLatest()) return
    allColumns.value = resp.data?.columns || []
    rows.value = resp.data?.items || []
    pagination.total = resp.data?.total || 0
    loadError.value = ''
  } catch (e) {
    if (!isLatest()) return
    rows.value = []
    pagination.total = 0
    loadError.value = (e as Error)?.message || ''
  } finally {
    if (isLatest()) loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  Object.assign(searchForm, { namespace: '', keyword: '', labelSelector: '' })
  handleSearch()
}

const openDetail = (row: ResourceTableRow) => {
  current.value = row
  detailVisible.value = true
}

const openNamespace = (namespace: string) => router.push({ path: `/cluster/${clusterId.value}/namespace/detail`, query: { namespace } })

const ownerRoute = (row: ResourceTableRow) => objectDetailRoute(clusterId.value, row.ownerKind, row.namespace, row.ownerName)
const goOwner = (row: ResourceTableRow) => {
  const target = ownerRoute(row)
  if (target) router.push(target)
}

const openYaml = async (row: ResourceTableRow) => {
  try {
    yamlContent.value = await fetchResourceYaml({ ...props.resource, clusterId: clusterId.value, namespace: row.namespace || undefined, name: row.name })
    current.value = row
    yamlVisible.value = true
  } catch {
    // 请求拦截器已提示错误
  }
}

const handleMore = (command: string, row: ResourceTableRow) => {
  current.value = row
  if (command === '__yaml') openYaml(row)
  else emit('action', command, row)
}

const openYamlAdd = () => {
  if (!props.yamlTemplate) return
  yamlAddContent.value = props.yamlTemplate(searchForm.namespace || 'default')
  yamlAddVisible.value = true
}

const confirmYamlAdd = async () => {
  if (!yamlAddContent.value.trim()) {
    ElMessage.warning('请输入YAML内容')
    return
  }
  yamlAddLoading.value = true
  try {
    await resourceApi.apply({
      ...props.resource,
      clusterId: clusterId.value,
      namespace: props.namespaced ? searchForm.namespace || undefined : undefined,
      yaml: yamlAddContent.value
    })
    ElMessage.success('创建成功')
    yamlAddVisible.value = false
    fetchList(true)
  } catch {
    // 请求拦截器已提示错误
  } finally {
    yamlAddLoading.value = false
  }
}

const handleDelete = async (row: ResourceTableRow) => {
  try {
    await ElMessageBox.confirm(`确定要删除 ${props.kindLabel} "${row.name}" 吗？此操作不可恢复。`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await resourceApi.delete({ ...props.resource, clusterId: clusterId.value, namespace: row.namespace || undefined, name: row.name })
    ElMessage.success('删除成功')
    if (rows.value.length === 1 && pagination.page > 1) pagination.page--
    fetchList(true)
  } catch {
    // 请求拦截器已提示错误
  }
}

watch(
  [clusterId, () => JSON.stringify(props.resource)],
  ([id], old) => {
    if (old && old[0] !== id) searchForm.namespace = ''
    pagination.page = 1
    if (id) fetchList()
    else rows.value = []
  },
  { immediate: true }
)

// 从详情页返回时静默刷新; 首次进入时 watch 已经加载过
let activated = false
onActivated(() => {
  if (activated) fetchList(true)
  activated = true
})

defineExpose({ refresh: () => fetchList(true) })
</script>

<style scoped>
.rt-alert {
  margin-bottom: 12px;
}

.rt-wide {
  margin-left: auto;
}

.sub-text {
  display: block;
  color: #94a3b8;
  font-size: 12px;
}

.mono {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12.5px;
}

.action-buttons {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  flex-wrap: nowrap;
}

.btn-main {
  font-weight: 500;
}

:deep(.action-buttons .el-button) {
  height: 28px;
  padding: 0 10px;
  border-radius: 8px;
}
</style>
