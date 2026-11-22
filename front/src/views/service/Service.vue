<template>
  <div class="service-list list-page">
    <PageHeader title="Service管理" subtitle="管理服务发现与负载均衡" icon="svc">
      <el-button type="primary" @click="openCreate">
        <el-icon><Plus /></el-icon>
        创建服务
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
          <el-input v-model="searchForm.keyword" placeholder="输入Service名称" clearable style="width: 200px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="searchForm.type" placeholder="全部" clearable style="width: 200px" @change="handleSearch">
            <el-option v-for="t in SERVICE_TYPE_OPTIONS" :key="t.value" :label="t.label" :value="t.value" />
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
      <el-table v-loading="loading" :data="serviceList" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
        <el-table-column prop="name" label="名称" min-width="180" header-align="center" align="center">
          <template #default="{ row }">
            <div class="name-cell">
              <el-link type="primary" :title="row.describe || ''" @click="openDetail(row)">{{ row.name }}</el-link>
              <span v-if="row.alias" class="name-cell__alias">{{ row.alias }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="服务类型" min-width="120" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="serviceTypeTag(row)" size="small">{{ serviceTypeLabel(row) }}</el-tag>
            <span v-if="row.status !== 'Ready'" class="sub-text">{{ row.status }}</span>
          </template>
        </el-table-column>
        <el-table-column label="集群 IP" min-width="130" header-align="center" align="center">
          <template #default="{ row }"><span class="mono">{{ row.internalIP }}</span></template>
        </el-table-column>
        <el-table-column label="端口" min-width="170" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.ports?.length" class="mono">{{ row.ports.map(servicePortText).join('，') }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column label="外部访问" min-width="150" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.externalAccess && row.externalAccess !== '-'" class="mono">{{ row.externalAccess }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" min-width="120" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="openNamespace(row)">{{ row.namespace }}</el-link>
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
                    <el-dropdown-item command="edit">编辑</el-dropdown-item>
                    <el-dropdown-item command="yaml">编辑YAML</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-button size="small" type="danger" plain class="btn-danger" @click="handleDelete(row)">删除</el-button>
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
        @current-change="fetchList"
      />
    </el-card>

    <ServiceFormDialog v-model="formVisible" :cluster-id="clusterId" :default-namespace="searchForm.namespace" :edit-target="editTarget" @success="handleSaved" />
    <YamlDialog
      v-model="yamlAddVisible"
      title="YAML添加Service"
      subtitle="已内置示例，可直接修改后使用"
      v-model:content="yamlAddContent"
      :readonly="false"
      filename="service.yaml"
      :submitting="yamlAddLoading"
      @confirm="confirmYamlAdd"
    />
    <YamlDialog
      v-model="yamlVisible"
      title="编辑YAML"
      subtitle="Service 配置"
      :info="[{ label: '集群', value: clusterName }, { label: '命名空间', value: current?.namespace }, { label: '名称', value: current?.name }, { label: '类型', value: 'Service' }]"
      v-model:content="yamlContent"
      :editor-title="`${current?.name} - Service YAML`"
      :filename="`${current?.name}-service.yaml`"
      :apply-target="current ? { clusterId, kind: 'service', namespace: current.namespace, name: current.name } : null"
      @applied="fetchList"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Document, MoreFilled, Plus, Refresh, Search } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import ServiceFormDialog from '@/components/service/ServiceFormDialog.vue'
import { serviceApi, type ServiceVO } from '@/api/service'
import { fetchResourceYaml, resourceApi } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useLatestRequest } from '@/composables/useLatestRequest'
import { useNamespaceFetcher, namespaceSelectConfig } from '@/composables/useInfiniteSelect'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { formatDateTime } from '@/utils/job'
import { SERVICE_TYPE_OPTIONS, servicePortText, serviceTypeLabel, serviceTypeTag, serviceYamlExample } from '@/utils/network'
import '@/assets/styles/config-editor.css'

const route = useRoute()
const router = useRouter()
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)
const clusterName = computed(() => clusterStore.currentClusterName)
const namespaceFetcher = computed(() => useNamespaceFetcher(clusterId.value))

const loading = ref(false)
const serviceList = ref<ServiceVO[]>([])
// 从其他页面跳转过来时可以带上命名空间
const searchForm = reactive({ namespace: (route.query.namespace as string) || '', keyword: '', type: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const formVisible = ref(false)
const editTarget = ref<{ namespace: string; name: string } | null>(null)
const yamlVisible = ref(false)
const yamlContent = ref('')
const yamlAddVisible = ref(false)
const yamlAddContent = ref('')
const yamlAddLoading = ref(false)
const current = ref<ServiceVO | null>(null)

// 快速切换条件时只采用最后一次查询的结果
const nextRequest = useLatestRequest()
const fetchList = async () => {
  if (!clusterId.value) {
    serviceList.value = []
    pagination.total = 0
    return
  }
  loading.value = true
  const isLatest = nextRequest()
  try {
    const resp = await serviceApi.getServices({
      clusterId: clusterId.value,
      namespace: searchForm.namespace,
      keyword: searchForm.keyword.trim(),
      type: searchForm.type,
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    if (!isLatest()) return
    serviceList.value = resp.data?.items || []
    pagination.total = resp.data?.total || 0
  } catch {
    if (!isLatest()) return
    serviceList.value = []
  } finally {
    if (isLatest()) loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  Object.assign(searchForm, { namespace: '', keyword: '', type: '' })
  handleSearch()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  handleSearch()
}

const openDetail = (row: { namespace: string; name: string }) => {
  router.push({ path: `/cluster/${clusterId.value}/service/service/detail`, query: { namespace: row.namespace, name: row.name } })
}

const openNamespace = (row: ServiceVO) => {
  router.push({ path: `/cluster/${clusterId.value}/namespace/detail`, query: { namespace: row.namespace } })
}

const openCreate = () => {
  editTarget.value = null
  formVisible.value = true
}

const handleSaved = (namespace: string, name: string) => {
  const created = !editTarget.value
  fetchList()
  if (!created) return
  ElMessageBox.confirm('服务已创建，是否查看详情？', '创建成功', { type: 'success', confirmButtonText: '查看详情', cancelButtonText: '留在列表' })
    .then(() => openDetail({ namespace, name }))
    .catch(() => {})
}

const handleMore = async (command: string, row: ServiceVO) => {
  current.value = row
  if (command === 'edit') {
    editTarget.value = { namespace: row.namespace, name: row.name }
    formVisible.value = true
  } else if (command === 'yaml') {
    try {
      yamlContent.value = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'service', namespace: row.namespace, name: row.name })
      yamlVisible.value = true
    } catch {
      // 请求拦截器已提示错误
    }
  }
}

const openYamlAdd = () => {
  yamlAddContent.value = serviceYamlExample(searchForm.namespace || 'default')
  yamlAddVisible.value = true
}

const confirmYamlAdd = async () => {
  if (!yamlAddContent.value.trim()) {
    ElMessage.warning('请输入YAML内容')
    return
  }
  yamlAddLoading.value = true
  try {
    await resourceApi.apply({ clusterId: clusterId.value, kind: 'service', namespace: searchForm.namespace || undefined, yaml: yamlAddContent.value })
    ElMessage.success('创建成功')
    yamlAddVisible.value = false
    fetchList()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    yamlAddLoading.value = false
  }
}

const handleDelete = async (row: ServiceVO) => {
  try {
    await ElMessageBox.confirm(`确定要删除服务 "${row.name}" 吗？此操作不可恢复。`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await serviceApi.deleteService({ clusterId: clusterId.value, namespace: row.namespace, name: row.name })
    ElMessage.success('删除成功')
    if (serviceList.value.length === 1 && pagination.page > 1) pagination.page--
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
      serviceList.value = []
      pagination.total = 0
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.service-list {
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
</style>
