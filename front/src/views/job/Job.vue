<template>
  <div class="job-list list-page">
    <PageHeader title="Job管理" subtitle="管理一次性执行的批处理任务" icon="job">
      <el-button type="primary" @click="createVisible = true">
        <el-icon><Plus /></el-icon>
        创建任务
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
          <el-input v-model="searchForm.keyword" placeholder="输入Job名称" clearable style="width: 200px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 140px" @change="handleSearch">
            <el-option v-for="opt in JOB_STATUS_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
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
        :data="jobList"
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
        <el-table-column label="状态" min-width="110" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="jobStatusTag(row.status)" size="small">{{ jobStatusText(row.status) }}</el-tag>
            <span v-if="row.status === 'Failed' && row.message" class="status-msg" :title="row.message">{{ row.message }}</span>
          </template>
        </el-table-column>
        <el-table-column label="完成数量" min-width="100" header-align="center" align="center">
          <template #default="{ row }">
            {{ row.succeeded }} / {{ row.completions }}
            <span v-if="row.failed" class="sub-text">失败 {{ row.failed }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" min-width="120" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="handleNamespaceDetail(row)">{{ row.namespace }}</el-link>
          </template>
        </el-table-column>
        <el-table-column label="所属定时任务" min-width="140" header-align="center" align="center">
          <template #default="{ row }">
            <el-link v-if="row.ownerCronJob" type="primary" @click="handleCronJobDetail(row)">{{ row.ownerCronJob }}</el-link>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column label="开始时间" min-width="160" header-align="center" align="center">
          <template #default="{ row }">{{ formatDateTime(row.startTime) }}</template>
        </el-table-column>
        <el-table-column label="运行时长" min-width="100" header-align="center" align="center">
          <template #default="{ row }">{{ formatDuration(row.startTime, row.completionTime) }}</template>
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
                    <el-dropdown-item command="edit">编辑</el-dropdown-item>
                    <el-dropdown-item command="rerun" :disabled="row.active > 0">重新执行</el-dropdown-item>
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

    <JobCreateDialog v-model="createVisible" kind="job" :cluster-id="clusterId" :default-namespace="searchForm.namespace" @success="handleCreated" />
    <JobEditDialog v-model="editVisible" kind="job" :cluster-id="clusterId" :target="currentJob" @success="fetchJobList" />

    <!-- YAML添加对话框 -->
    <YamlDialog
      v-model="yamlAddDialogVisible"
      title="YAML添加Job"
      subtitle="已内置示例，可直接修改后使用"
      v-model:content="yamlAddContent"
      :readonly="false"
      filename="job.yaml"
      :submitting="yamlAddLoading"
      @confirm="confirmYamlAdd"
    />

    <!-- 编辑YAML对话框 -->
    <YamlDialog
      v-model="viewYamlDialogVisible"
      title="编辑YAML"
      subtitle="Job 配置"
      :info="[{ label: '集群', value: clusterName }, { label: '命名空间', value: currentJob?.namespace }, { label: '名称', value: currentJob?.name }, { label: '类型', value: 'Job' }]"
      v-model:content="viewYaml"
      :editor-title="`${currentJob?.name} - Job YAML`"
      :filename="`${currentJob?.name}-job.yaml`"
      :apply-target="currentJob ? { clusterId, kind: 'job', namespace: currentJob.namespace, name: currentJob.name } : null"
      @applied="fetchJobList"
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
import { jobApi, type JobVO } from '@/api/job'
import { fetchResourceYaml, resourceApi } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useNamespaceFetcher, namespaceSelectConfig } from '@/composables/useInfiniteSelect'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { JOB_STATUS_OPTIONS, formatDateTime, formatDuration, jobStatusTag, jobStatusText, jobYamlExample } from '@/utils/job'
import '@/assets/styles/config-editor.css'

const router = useRouter()
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)
const clusterName = computed(() => clusterStore.currentClusterName)

const loading = ref(false)
const jobList = ref<JobVO[]>([])
const namespaceFetcher = computed(() => useNamespaceFetcher(clusterId.value))

const createVisible = ref(false)
const editVisible = ref(false)
const viewYamlDialogVisible = ref(false)
const viewYaml = ref('')
const currentJob = ref<JobVO | null>(null)
const yamlAddDialogVisible = ref(false)
const yamlAddLoading = ref(false)
const yamlAddContent = ref('')

const searchForm = reactive({ namespace: '', keyword: '', status: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const fetchJobList = async () => {
  if (!clusterId.value) {
    jobList.value = []
    pagination.total = 0
    return
  }
  loading.value = true
  try {
    const response = await jobApi.getJobs({
      clusterId: clusterId.value,
      namespace: searchForm.namespace,
      keyword: searchForm.keyword.trim(),
      status: searchForm.status,
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    if (response.code === 200) {
      jobList.value = response.data.items || []
      pagination.total = response.data.total || 0
    }
  } catch (error) {
    console.error('获取Job列表失败:', error)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchJobList()
}

const handleReset = () => {
  Object.assign(searchForm, { namespace: '', keyword: '', status: '' })
  handleSearch()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchJobList()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchJobList()
}

const handleNamespaceChange = (namespace: string) => {
  searchForm.namespace = namespace
  handleSearch()
}

const handleDetail = (row: { namespace: string; name: string }) => {
  router.push({ path: `/cluster/${clusterId.value}/job/job/detail`, query: { namespace: row.namespace, name: row.name } })
}

const handleCronJobDetail = (row: JobVO) => {
  router.push({ path: `/cluster/${clusterId.value}/job/cronjob/detail`, query: { namespace: row.namespace, name: row.ownerCronJob } })
}

const handleNamespaceDetail = (row: JobVO) => {
  router.push({ path: `/cluster/${clusterId.value}/namespace/detail`, query: { namespace: row.namespace } })
}

const handleCreated = (namespace: string, name: string) => {
  fetchJobList()
  ElMessageBox.confirm('任务已创建，是否查看详情？', '创建成功', { type: 'success', confirmButtonText: '查看详情', cancelButtonText: '留在列表' })
    .then(() => handleDetail({ namespace, name }))
    .catch(() => {})
}

const handleMoreAction = (command: string, row: JobVO) => {
  currentJob.value = row
  if (command === 'edit') editVisible.value = true
  else if (command === 'rerun') handleRerun(row)
  else if (command === 'yaml') handleViewYaml(row)
}

const handleRerun = async (row: JobVO) => {
  try {
    await ElMessageBox.confirm(
      `将删除任务 "${row.name}" 当前的容器组，并按相同配置重新创建任务。本次运行结果会保存在详情的运行记录中。`,
      '重新执行',
      { type: 'warning', confirmButtonText: '重新执行', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await jobApi.rerun({ clusterId: clusterId.value, namespace: row.namespace, name: row.name })
    ElMessage.success('已重新执行')
    fetchJobList()
  } catch {
    // 请求拦截器已提示错误
  }
}

const openYamlAdd = () => {
  yamlAddContent.value = jobYamlExample(searchForm.namespace || 'default')
  yamlAddDialogVisible.value = true
}

const confirmYamlAdd = async () => {
  if (!yamlAddContent.value.trim()) {
    ElMessage.warning('请输入YAML内容')
    return
  }
  yamlAddLoading.value = true
  try {
    await resourceApi.apply({ clusterId: clusterId.value, kind: 'job', namespace: searchForm.namespace || undefined, yaml: yamlAddContent.value })
    ElMessage.success('创建成功')
    yamlAddDialogVisible.value = false
    fetchJobList()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    yamlAddLoading.value = false
  }
}

const handleViewYaml = async (row: JobVO) => {
  try {
    viewYaml.value = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'job', namespace: row.namespace, name: row.name })
  } catch {
    return
  }
  currentJob.value = row
  viewYamlDialogVisible.value = true
}

const handleDelete = async (row: JobVO) => {
  try {
    await ElMessageBox.confirm(`确定要删除任务 "${row.name}" 吗？它创建的容器组会一起删除，此操作不可恢复。`, '确认删除', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    await jobApi.delete({ clusterId: clusterId.value, namespace: row.namespace, name: row.name })
    ElMessage.success('删除成功')
    if (jobList.value.length === 1 && pagination.page > 1) pagination.page--
    fetchJobList()
  } catch {
    // 请求拦截器已提示错误
  }
}

watch(
  clusterId,
  id => {
    searchForm.namespace = ''
    pagination.page = 1
    if (id) fetchJobList()
    else {
      jobList.value = []
      pagination.total = 0
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.job-list {
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
