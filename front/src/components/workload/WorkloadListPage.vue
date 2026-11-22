<template>
  <div class="workload-list list-page">
    <PageHeader :title="`${meta.label}管理`" :subtitle="meta.desc" :icon="meta.icon">
      <el-button type="primary" @click="createVisible = true">
        <el-icon><Plus /></el-icon>
        创建{{ meta.label }}
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
          <el-input v-model="searchForm.name" :placeholder="`输入${meta.label}名称`" clearable style="width: 200px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 130px">
            <el-option v-for="s in WORKLOAD_STATUS_OPTIONS" :key="s" :label="s" :value="s" />
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
      <el-table v-loading="loading" :data="pageRows" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
        <el-table-column prop="name" label="名称" min-width="180" header-align="center" align="center">
          <template #default="{ row }">
            <div class="name-cell">
              <el-link type="primary" :title="row.describe || ''" @click="openDetail(row)">{{ row.name }}</el-link>
              <span v-if="row.alias" class="name-cell__alias">{{ row.alias }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="workloadStatusTag(row.status)" size="small">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="kind === 'daemonset' ? '就绪 / 节点数' : '就绪 / 副本数'" min-width="110" header-align="center" align="center">
          <template #default="{ row }">
            <span :class="{ 'warn-text': row.readyReplicas < row.replicas }">{{ row.readyReplicas }} / {{ row.replicas }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" min-width="120" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="openNamespace(row)">{{ row.namespace }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="容器数量" min-width="90" header-align="center" align="center" prop="containers" />
        <el-table-column label="更新时间" min-width="160" header-align="center" align="center">
          <template #default="{ row }">{{ formatDateTime(row.updateTime) }}</template>
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
                    <el-dropdown-item command="info">编辑信息</el-dropdown-item>
                    <el-dropdown-item v-if="kind !== 'daemonset'" command="scale">调整副本</el-dropdown-item>
                    <el-dropdown-item command="restart">重新创建</el-dropdown-item>
                    <el-dropdown-item command="yaml">编辑YAML</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-button size="small" type="danger" plain class="btn-danger" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <ListPagination v-model:current-page="page" v-model:page-size="pageSize" :total="total" :page-sizes="[10, 20, 50, 100]" />
    </el-card>

    <WorkloadCreateDialog v-model="createVisible" :kind="kind" :cluster-id="clusterId" :default-namespace="searchForm.namespace" @success="handleCreated" />
    <ResourceInfoDialog
      v-model="infoVisible"
      :name="current?.name || ''"
      :alias="current?.alias"
      :describe="current?.describe"
      :submit="saveInfo"
      @success="fetchList"
    />
    <YamlDialog
      v-model="yamlAddVisible"
      :title="`YAML添加${meta.label}`"
      subtitle="已内置示例，可直接修改后使用"
      v-model:content="yamlAddContent"
      :readonly="false"
      :filename="`${kind}.yaml`"
      :submitting="yamlAddLoading"
      @confirm="confirmYamlAdd"
    />
    <YamlDialog
      v-model="yamlVisible"
      title="编辑YAML"
      :subtitle="`${meta.label} 配置`"
      :info="[{ label: '集群', value: clusterName }, { label: '命名空间', value: current?.namespace }, { label: '名称', value: current?.name }, { label: '类型', value: meta.label }]"
      v-model:content="yamlContent"
      :editor-title="`${current?.name} - ${meta.label} YAML`"
      :filename="`${current?.name}-${kind}.yaml`"
      :apply-target="current ? { clusterId, kind, namespace: current.namespace, name: current.name } : null"
      @applied="fetchList"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, MoreFilled, Plus, Refresh, Search } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import ResourceInfoDialog from '@/components/ResourceInfoDialog.vue'
import WorkloadCreateDialog from '@/components/workload/WorkloadCreateDialog.vue'
import { workloadApi, type WorkloadKind, type WorkloadVO } from '@/api/workload'
import { fetchResourceYaml, resourceApi } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useLatestRequest } from '@/composables/useLatestRequest'
import { useNamespaceFetcher, namespaceSelectConfig } from '@/composables/useInfiniteSelect'
import { useClientPage } from '@/composables/useClientPage'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { formatDateTime } from '@/utils/job'
import { WORKLOAD_KINDS, WORKLOAD_STATUS_OPTIONS, workloadDetailPath, workloadStatusTag, workloadYamlExample } from '@/utils/workload'
import '@/assets/styles/config-editor.css'

// 三种工作负载的列表结构相同, 字段一致
type WorkloadRow = WorkloadVO

const props = defineProps<{ kind: WorkloadKind }>()

const router = useRouter()
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)
const clusterName = computed(() => clusterStore.currentClusterName)
const meta = computed(() => WORKLOAD_KINDS[props.kind])
const namespaceFetcher = computed(() => useNamespaceFetcher(clusterId.value))

const loading = ref(false)
const list = ref<WorkloadRow[]>([])
const searchForm = reactive({ namespace: '', name: '', status: '' })

// 后端返回全部数据, 状态筛选和分页在前端完成
const filtered = computed(() => list.value.filter(r => !searchForm.status || r.status === searchForm.status))
const { page, pageSize, total, rows: pageRows } = useClientPage(filtered)

const createVisible = ref(false)
const infoVisible = ref(false)
const yamlVisible = ref(false)
const yamlContent = ref('')
const yamlAddVisible = ref(false)
const yamlAddContent = ref('')
const yamlAddLoading = ref(false)
const current = ref<WorkloadRow | null>(null)

// 快速切换条件时只采用最后一次查询的结果
const nextRequest = useLatestRequest()
const fetchList = async () => {
  if (!clusterId.value) {
    list.value = []
    return
  }
  loading.value = true
  const params = { clusterId: clusterId.value, namespace: searchForm.namespace, name: searchForm.name.trim(), page: 1, pageSize: 100000 }
  const isLatest = nextRequest()
  try {
    // 后端已按创建时间倒序
    const resp = await workloadApi.list(props.kind, params)
    if (!isLatest()) return
    list.value = resp.data?.items || []
  } catch {
    if (!isLatest()) return
    list.value = []
  } finally {
    if (isLatest()) loading.value = false
  }
}

const handleSearch = () => {
  page.value = 1
  fetchList()
}

const handleReset = () => {
  Object.assign(searchForm, { namespace: '', name: '', status: '' })
  handleSearch()
}

const target = (row: { namespace: string; name: string }) => ({ clusterId: clusterId.value, kind: props.kind, namespace: row.namespace, name: row.name })

const openDetail = (row: { namespace: string; name: string }) => {
  router.push({ path: workloadDetailPath(clusterId.value, props.kind), query: { namespace: row.namespace, name: row.name } })
}

const openNamespace = (row: WorkloadRow) => {
  router.push({ path: `/cluster/${clusterId.value}/namespace/detail`, query: { namespace: row.namespace } })
}

const handleCreated = (namespace: string, name: string) => {
  fetchList()
  ElMessageBox.confirm(`${meta.value.label} 已创建，是否查看详情？`, '创建成功', { type: 'success', confirmButtonText: '查看详情', cancelButtonText: '留在列表' })
    .then(() => openDetail({ namespace, name }))
    .catch(() => {})
}

const saveInfo = (alias: string, describe: string) => workloadApi.updateInfo({ ...target(current.value as WorkloadRow), alias, describe })

const handleMore = async (command: string, row: WorkloadRow) => {
  current.value = row
  if (command === 'info') infoVisible.value = true
  else if (command === 'yaml') {
    try {
      yamlContent.value = await fetchResourceYaml(target(row))
      yamlVisible.value = true
    } catch {
      // 请求拦截器已提示错误
    }
  } else if (command === 'restart') {
    try {
      await ElMessageBox.confirm(`将按当前配置逐个重新创建 "${row.name}" 的容器组，与 kubectl rollout restart 相同。`, '重新创建', {
        type: 'warning',
        confirmButtonText: '重新创建',
        cancelButtonText: '取消'
      })
      await workloadApi.restart(target(row))
      ElMessage.success('已开始重新创建容器组')
      fetchList()
    } catch {
      // 取消或请求失败
    }
  } else if (command === 'scale') {
    try {
      const { value } = await ElMessageBox.prompt(`当前副本数 ${row.replicas}，就绪 ${row.readyReplicas}`, `调整 "${row.name}" 的副本数`, {
        inputValue: String(row.replicas),
        inputPattern: /^(\d|[1-9]\d{1,2}|1000)$/,
        inputErrorMessage: '请输入 0 到 1000 之间的整数',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      })
      await workloadApi.scale({ ...target(row), replicas: Number(value) })
      ElMessage.success('副本数已调整')
      fetchList()
    } catch {
      // 取消或请求失败
    }
  }
}

const openYamlAdd = () => {
  yamlAddContent.value = workloadYamlExample(props.kind, searchForm.namespace || 'default')
  yamlAddVisible.value = true
}

const confirmYamlAdd = async () => {
  if (!yamlAddContent.value.trim()) {
    ElMessage.warning('请输入YAML内容')
    return
  }
  yamlAddLoading.value = true
  try {
    await resourceApi.apply({ clusterId: clusterId.value, kind: props.kind, namespace: searchForm.namespace || undefined, yaml: yamlAddContent.value })
    ElMessage.success('创建成功')
    yamlAddVisible.value = false
    fetchList()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    yamlAddLoading.value = false
  }
}

const handleDelete = async (row: WorkloadRow) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除 ${meta.value.label} "${row.name}" 吗？它的容器组会一起删除${props.kind === 'statefulset' ? '，持久卷声明会保留' : ''}，此操作不可恢复。`,
      '确认删除',
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await workloadApi.delete(target(row))
    ElMessage.success('删除成功')
    fetchList()
  } catch {
    // 请求拦截器已提示错误
  }
}

watch(
  clusterId,
  id => {
    searchForm.namespace = ''
    page.value = 1
    if (id) fetchList()
    else list.value = []
  },
  { immediate: true }
)
</script>

<style scoped>
.workload-list {
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

.warn-text {
  color: #f59e0b;
  font-weight: 600;
}
</style>
