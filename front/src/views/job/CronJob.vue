<template>
  <div class="cronjob-list list-page">
    <PageHeader title="CronJob管理" subtitle="管理按计划周期执行的定时任务" icon="cronjob">
      <el-button type="primary" @click="createVisible = true">
        <el-icon><Plus /></el-icon>
        创建定时任务
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
            @change="handleNamespaceChange"
          />
        </el-form-item>
        <el-form-item label="名称">
          <el-input v-model="searchForm.keyword" placeholder="输入CronJob名称" clearable style="width: 200px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 140px" @change="handleSearch">
            <el-option label="运行中" value="Active" />
            <el-option label="已暂停" value="Suspended" />
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
      <el-table
        v-loading="loading"
        :data="cronJobList"
        style="width: 100%"
        :header-cell-class-name="tableHeaderClass"
        :cell-class-name="tableCellClass"
      >
        <el-table-column prop="name" label="名称" min-width="180" header-align="center" align="center">
          <template #default="{ row }">
            <div class="name-cell">
              <el-link type="primary" :title="row.describe || ''" @click="handleDetail(row)">{{ row.name }}</el-link>
              <span v-if="row.alias" class="name-cell__alias">{{ row.alias }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="row.suspend ? 'info' : 'success'" size="small">{{ row.suspend ? '已暂停' : '运行中' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="schedule" label="调度规则" min-width="170" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag type="info" size="small" class="mono">{{ row.schedule }}</el-tag>
            <span v-if="describeCron(row.schedule)" class="sub-text">{{ describeCron(row.schedule) }}{{ row.timeZone ? `（${row.timeZone}）` : '' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" min-width="120" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="handleNamespaceDetail(row)">{{ row.namespace }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="运行中的任务" min-width="110" header-align="center" align="center">
          <template #default="{ row }">{{ row.active }}</template>
        </el-table-column>
        <el-table-column label="上次调度时间" min-width="160" header-align="center" align="center">
          <template #default="{ row }">{{ formatDateTime(row.lastRunTime) }}</template>
        </el-table-column>
        <el-table-column label="上次成功时间" min-width="160" header-align="center" align="center">
          <template #default="{ row }">{{ formatDateTime(row.lastSuccessfulTime) }}</template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" min-width="160" header-align="center" align="center">
          <template #default="{ row }">{{ formatDateTime(row.createTime) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" header-align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" class="btn-main" @click="handleDetail(row)">详情</el-button>
              <el-dropdown trigger="click" @command="(command: string) => handleMoreAction(command, row)">
                <el-button size="small" text circle class="more-btn">
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="trigger">立即执行</el-dropdown-item>
                    <el-dropdown-item command="suspend">{{ row.suspend ? '启用' : '暂停' }}</el-dropdown-item>
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
        @current-change="handleCurrentChange"
      />
    </el-card>

    <JobCreateDialog v-model="createVisible" kind="cronjob" :cluster-id="clusterId" :default-namespace="searchForm.namespace" @success="handleCreated" />
    <JobEditDialog v-model="editVisible" kind="cronjob" :cluster-id="clusterId" :target="currentCronJob" @success="fetchCronJobList" />

    <YamlDialog
      v-model="yamlAddDialogVisible"
      title="YAML添加CronJob"
      subtitle="已内置示例，可直接修改后使用"
      v-model:content="yamlAddContent"
      :readonly="false"
      filename="cronjob.yaml"
      :submitting="yamlAddLoading"
      @confirm="confirmYamlAdd"
    />

    <YamlDialog
      v-model="viewYamlDialogVisible"
      title="编辑YAML"
      subtitle="CronJob 配置"
      :info="[{ label: '集群', value: clusterName }, { label: '命名空间', value: currentCronJob?.namespace }, { label: '名称', value: currentCronJob?.name }, { label: '类型', value: 'CronJob' }]"
      v-model:content="viewYaml"
      :editor-title="`${currentCronJob?.name} - CronJob YAML`"
      :filename="`${currentCronJob?.name}-cronjob.yaml`"
      :apply-target="currentCronJob ? { clusterId, kind: 'cronjob', namespace: currentCronJob.namespace, name: currentCronJob.name } : null"
      @applied="fetchCronJobList"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, MoreFilled, Document, Plus } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import JobCreateDialog from '@/components/job/JobCreateDialog.vue'
import JobEditDialog from '@/components/job/JobEditDialog.vue'
import { cronJobApi, type CronJobVO } from '@/api/cronjob'
import { fetchResourceYaml, resourceApi } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useNamespaceFetcher, namespaceSelectConfig } from '@/composables/useInfiniteSelect'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { cronJobYamlExample, describeCron, formatDateTime } from '@/utils/job'
import '@/assets/styles/config-editor.css'

const router = useRouter()
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)
const clusterName = computed(() => clusterStore.currentClusterName)

const loading = ref(false)
const cronJobList = ref<CronJobVO[]>([])
const namespaceFetcher = computed(() => useNamespaceFetcher(clusterId.value))

const createVisible = ref(false)
const editVisible = ref(false)
const viewYamlDialogVisible = ref(false)
const viewYaml = ref('')
const currentCronJob = ref<CronJobVO | null>(null)
const yamlAddDialogVisible = ref(false)
const yamlAddLoading = ref(false)
const yamlAddContent = ref('')

const searchForm = reactive({ namespace: '', keyword: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const fetchCronJobList = async () => {
  if (!clusterId.value) {
    cronJobList.value = []
    pagination.total = 0
    return
  }
  loading.value = true
  try {
    const response = await cronJobApi.getCronJobs({
      clusterId: clusterId.value,
      namespace: searchForm.namespace,
      keyword: searchForm.keyword.trim(),
      status: searchForm.status,
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    if (response.code === 200) {
      cronJobList.value = response.data.items || []
      pagination.total = response.data.total || 0
    }
  } catch (error) {
    console.error('获取CronJob列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchCronJobList()
}

const handleReset = () => {
  Object.assign(searchForm, { namespace: '', keyword: '', status: '' })
  handleSearch()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchCronJobList()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchCronJobList()
}

const handleNamespaceChange = (namespace: string) => {
  searchForm.namespace = namespace
  handleSearch()
}

const handleDetail = (row: { namespace: string; name: string }) => {
  router.push({ path: `/cluster/${clusterId.value}/job/cronjob/detail`, query: { namespace: row.namespace, name: row.name } })
}

const handleNamespaceDetail = (row: CronJobVO) => {
  router.push({ path: `/cluster/${clusterId.value}/namespace/detail`, query: { namespace: row.namespace } })
}

const handleCreated = (namespace: string, name: string) => {
  fetchCronJobList()
  ElMessageBox.confirm('定时任务已创建，是否查看详情？', '创建成功', { type: 'success', confirmButtonText: '查看详情', cancelButtonText: '留在列表' })
    .then(() => handleDetail({ namespace, name }))
    .catch(() => {})
}

const handleMoreAction = (command: string, row: CronJobVO) => {
  currentCronJob.value = row
  if (command === 'trigger') handleTrigger(row)
  else if (command === 'suspend') toggleSuspend(row)
  else if (command === 'edit') editVisible.value = true
  else if (command === 'yaml') handleViewYaml(row)
}

const handleTrigger = async (row: CronJobVO) => {
  try {
    await ElMessageBox.confirm('将按任务模板立即创建一个任务，不影响原有的调度计划。', `立即执行 "${row.name}"`, {
      type: 'info',
      confirmButtonText: '立即执行',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    const resp = await cronJobApi.trigger({ clusterId: clusterId.value, namespace: row.namespace, name: row.name })
    ElMessage.success(`已创建任务 ${resp.data?.jobName || ''}`)
    fetchCronJobList()
  } catch {
    // 请求拦截器已提示错误
  }
}

const toggleSuspend = async (row: CronJobVO) => {
  const suspend = !row.suspend
  try {
    await ElMessageBox.confirm(
      suspend ? '暂停后不会再按计划创建新任务，已经在运行的任务不受影响。' : '启用后会按调度规则继续创建任务。',
      suspend ? `暂停 "${row.name}"` : `启用 "${row.name}"`,
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await cronJobApi.suspend({ clusterId: clusterId.value, namespace: row.namespace, name: row.name, suspend })
    ElMessage.success(suspend ? '已暂停' : '已启用')
    fetchCronJobList()
  } catch {
    // 请求拦截器已提示错误
  }
}

const openYamlAdd = () => {
  yamlAddContent.value = cronJobYamlExample(searchForm.namespace || 'default')
  yamlAddDialogVisible.value = true
}

const confirmYamlAdd = async () => {
  if (!yamlAddContent.value.trim()) {
    ElMessage.warning('请输入YAML内容')
    return
  }
  yamlAddLoading.value = true
  try {
    await resourceApi.apply({ clusterId: clusterId.value, kind: 'cronjob', namespace: searchForm.namespace || undefined, yaml: yamlAddContent.value })
    ElMessage.success('创建成功')
    yamlAddDialogVisible.value = false
    fetchCronJobList()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    yamlAddLoading.value = false
  }
}

const handleViewYaml = async (row: CronJobVO) => {
  try {
    viewYaml.value = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'cronjob', namespace: row.namespace, name: row.name })
  } catch {
    return
  }
  currentCronJob.value = row
  viewYamlDialogVisible.value = true
}

const handleDelete = async (row: CronJobVO) => {
  try {
    await ElMessageBox.confirm(`确定要删除定时任务 "${row.name}" 吗？它创建的任务和容器组会一起删除，此操作不可恢复。`, '确认删除', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    await cronJobApi.delete({ clusterId: clusterId.value, namespace: row.namespace, name: row.name })
    ElMessage.success('删除成功')
    if (cronJobList.value.length === 1 && pagination.page > 1) pagination.page--
    fetchCronJobList()
  } catch {
    // 请求拦截器已提示错误
  }
}

watch(
  clusterId,
  id => {
    searchForm.namespace = ''
    pagination.page = 1
    if (id) fetchCronJobList()
    else {
      cronJobList.value = []
      pagination.total = 0
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.cronjob-list {
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
