<template>
  <div class="job-detail detail-page">
    <PageHeader
      :title="job?.name || name || '任务详情'"
      :subtitle="job ? job.alias || job.describe || `命名空间：${job.namespace}` : '任务详情'"
      icon="job"
      back
      @back="handleBack"
    >
      <template v-if="job" #meta>
        <el-tag :type="jobStatusTag(job.status)" effect="dark">{{ jobStatusText(job.status) }}</el-tag>
        <el-tag v-if="job.ownerCronJob" type="info" effect="light">{{ job.manual ? '手动触发' : '定时创建' }}</el-tag>
      </template>
      <template v-if="job">
        <el-button :loading="loading" @click="fetchDetail()">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button @click="editVisible = true">
          <el-icon><Edit /></el-icon>
          编辑
        </el-button>
        <el-button type="primary" :disabled="job.active > 0" :loading="rerunning" @click="handleRerun">
          <el-icon><VideoPlay /></el-icon>
          重新执行
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

    <div v-if="loading && !job" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="job" class="detail-content">
      <!-- 属性 -->
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
              <span><el-link type="primary" @click="goNamespace">{{ job.namespace }}</el-link></span>
            </div>
            <div class="jd-attr">
              <label>状态</label>
              <span>
                {{ jobStatusText(job.status) }}
                <span v-if="job.message" class="jd-msg">{{ job.message }}</span>
              </span>
            </div>
            <div class="jd-attr"><label>最大重试次数</label><span>{{ job.backoffLimit }}</span></div>
            <div class="jd-attr">
              <label>容器组完成数量</label>
              <span>{{ job.succeeded }} / {{ job.completions }}<span v-if="job.failed" class="jd-sub">失败 {{ job.failed }}</span></span>
            </div>
            <div class="jd-attr"><label>并行容器组数量</label><span>{{ job.parallelism }}</span></div>
            <div class="jd-attr">
              <label>最大运行时间（s）</label>
              <span>{{ job.activeDeadlineSeconds ? formatSeconds(job.activeDeadlineSeconds) : '-' }}</span>
            </div>
            <div class="jd-attr"><label>创建时间</label><span>{{ formatDateTime(job.createTime) }}</span></div>
            <div class="jd-attr"><label>创建者</label><span>{{ job.creator || '-' }}</span></div>
            <div class="jd-attr"><label>开始时间</label><span>{{ formatDateTime(job.startTime) }}</span></div>
            <div class="jd-attr"><label>结束时间</label><span>{{ formatDateTime(job.completionTime) }}</span></div>
            <div class="jd-attr"><label>耗时</label><span>{{ formatDuration(job.startTime, job.completionTime) }}</span></div>
            <div class="jd-attr"><label>重启策略</label><span>{{ restartText(job.restartPolicy) }}</span></div>
            <div class="jd-attr">
              <label>结束后自动删除</label>
              <span>{{ job.ttlSecondsAfterFinished !== null && job.ttlSecondsAfterFinished !== undefined ? formatSeconds(job.ttlSecondsAfterFinished) + '后' : '否' }}</span>
            </div>
            <div class="jd-attr">
              <label>所属定时任务</label>
              <span>
                <el-link v-if="job.ownerCronJob" type="primary" @click="goCronJob">{{ job.ownerCronJob }}</el-link>
                <template v-else>-</template>
              </span>
            </div>
          </div>
        </el-collapse-transition>
      </el-card>

      <!-- 页签 -->
      <div class="jd-tabbar">
        <button v-for="t in tabs" :key="t.key" type="button" :class="['jd-tab', { 'is-active': activeTab === t.key }]" @click="setTab(t.key)">
          {{ t.label }}
          <span v-if="t.count !== undefined" class="jd-tab__count">{{ t.count }}</span>
        </button>
      </div>

      <!-- 运行记录 -->
      <el-card v-if="activeTab === 'records'" class="info-card">
        <template #header>
          <div class="card-header">
            <span>运行记录</span>
            <el-tag size="small" type="info">{{ job.records.length }}</el-tag>
            <span class="card-header__hint">每次「重新执行」都会生成一条新记录</span>
          </div>
        </template>
        <el-table :data="recordRows" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
          <el-table-column label="序号" width="110">
            <template #default="{ row }">
              第 {{ row.run }} 次
              <el-tag v-if="row.current" size="small" type="primary" effect="plain">当前</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="状态" min-width="180">
            <template #default="{ row }">
              <el-tag :type="jobStatusTag(row.status)" size="small">{{ jobStatusText(row.status) }}</el-tag>
              <span v-if="row.message" class="jd-msg">{{ row.message }}</span>
            </template>
          </el-table-column>
          <el-table-column label="完成数量" width="110" align="center" header-align="center">
            <template #default="{ row }">{{ row.succeeded }} / {{ row.completions }}</template>
          </el-table-column>
          <el-table-column label="失败次数" width="100" align="center" header-align="center" prop="failed" />
          <el-table-column label="开始时间" width="180">
            <template #default="{ row }">{{ formatDateTime(row.startTime) }}</template>
          </el-table-column>
          <el-table-column label="结束时间" width="180">
            <template #default="{ row }">{{ formatDateTime(row.completionTime) }}</template>
          </el-table-column>
          <el-table-column label="时长" width="120">
            <template #default="{ row }">{{ formatDuration(row.startTime, row.completionTime) }}</template>
          </el-table-column>
          <template #empty>
            <el-empty description="暂无运行记录" :image-size="72" />
          </template>
        </el-table>
        <ListPagination v-model:current-page="recordPage" v-model:page-size="recordPageSize" class="jd-pager" :total="recordTotal" :page-sizes="[10, 20, 50]" />
      </el-card>

      <!-- 资源状态 -->
      <template v-else-if="activeTab === 'resources'">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>容器组</span>
              <el-tag size="small" type="info">{{ job.pods.length }}</el-tag>
              <span class="card-header__hint">运行中 {{ job.active }}，成功 {{ job.succeeded }}，失败 {{ job.failed }}</span>
            </div>
          </template>
          <JobPodsTable :pods="job.pods" :cluster-id="clusterId" />
        </el-card>
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>容器</span>
              <el-tag size="small" type="info">{{ job.containers.length }}</el-tag>
            </div>
          </template>
          <ContainerTemplates :containers="job.containers" />
        </el-card>
      </template>

      <!-- 元数据 -->
      <el-card v-else-if="activeTab === 'metadata'" class="info-card">
        <template #header>
          <div class="card-header"><span>元数据</span></div>
        </template>
        <MetadataPanel :labels="job.labels" :annotations="job.annotations" />
      </el-card>

      <!-- 环境变量 -->
      <el-card v-else-if="activeTab === 'env'" class="info-card">
        <template #header>
          <div class="card-header"><span>环境变量</span></div>
        </template>
        <ContainerEnvList :containers="job.containers" />
      </el-card>

      <!-- 事件 -->
      <el-card v-else class="info-card">
        <template #header>
          <div class="card-header">
            <span>事件</span>
            <el-tag size="small" type="info">{{ job.events.length }}</el-tag>
          </div>
        </template>
        <EventsTable :events="job.events" />
      </el-card>
    </div>

    <el-empty v-else description="任务不存在或已被删除" />

    <JobEditDialog v-model="editVisible" kind="job" :cluster-id="clusterId" :target="job" @success="fetchDetail()" />
    <YamlDialog
      v-model="yamlVisible"
      v-model:content="yamlContent"
      title="编辑YAML"
      subtitle="Job 配置"
      :info="[{ label: '集群', value: clusterName }, { label: '命名空间', value: job?.namespace }, { label: '名称', value: job?.name }, { label: '类型', value: 'Job' }]"
      :editor-title="`${job?.name} - Job YAML`"
      :filename="`${job?.name}-job.yaml`"
      tip="任务创建后容器等设置不能修改，保存失败时可以复制 YAML 后删除任务再重新创建"
      :apply-target="job ? { clusterId, kind: 'job', namespace: job.namespace, name: job.name } : null"
      @applied="fetchDetail()"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Delete, Document, Edit, Refresh, VideoPlay } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import JobEditDialog from '@/components/job/JobEditDialog.vue'
import JobPodsTable from '@/components/job/JobPodsTable.vue'
import ContainerTemplates from '@/components/job/ContainerTemplates.vue'
import ContainerEnvList from '@/components/job/ContainerEnvList.vue'
import MetadataPanel from '@/components/job/MetadataPanel.vue'
import EventsTable from '@/components/job/EventsTable.vue'
import { jobApi, type JobDetailVO } from '@/api/job'
import { fetchResourceYaml } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useClientPage } from '@/composables/useClientPage'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { formatDateTime, formatDuration, formatSeconds, jobStatusTag, jobStatusText } from '@/utils/job'
import '@/assets/styles/job-detail.css'
import { useTabBack } from '@/composables/useTabBack'

type TabKey = 'records' | 'resources' | 'metadata' | 'env' | 'events'

const route = useRoute()
const router = useRouter()
// 返回当前标签内的上一页
const goBack = useTabBack()
const clusterStore = useClusterStore()

// 页面被 keep-alive 按完整路径缓存, 每个实例只对应一个资源.
// 参数在创建时取一次, 不能用 computed 跟随全局路由: 否则切到其他页面后, 缓存中的实例会用新页面的参数发请求
const clusterId = ref((route.params.clusterId as string) || (route.query.clusterId as string) || clusterStore.currentClusterId || '')
const clusterName = computed(() => clusterStore.currentClusterName)
const namespace = ref((route.query.namespace as string) || '')
const name = ref((route.query.name as string) || '')

const loading = ref(false)
const job = ref<JobDetailVO | null>(null)
const attrOpen = ref(true)
const editVisible = ref(false)
const rerunning = ref(false)
const yamlVisible = ref(false)
const yamlContent = ref('')

const activeTab = ref<TabKey>('records')
const tabs = computed(() => [
  { key: 'records' as TabKey, label: '运行记录', count: job.value?.records.length },
  { key: 'resources' as TabKey, label: '资源状态', count: job.value?.pods.length },
  { key: 'metadata' as TabKey, label: '元数据' },
  { key: 'env' as TabKey, label: '环境变量' },
  { key: 'events' as TabKey, label: '事件', count: job.value?.events.length }
])
// 页签只是页面内状态, 不写入地址: 地址变化会让标签栏和 keep-alive 当成新页面
const setTab = (key: TabKey) => {
  activeTab.value = key
}

const { page: recordPage, pageSize: recordPageSize, total: recordTotal, rows: recordRows } = useClientPage(() => job.value?.records || [])

// 运行中时定时刷新
let timer: ReturnType<typeof setTimeout> | undefined
const schedulePoll = () => {
  clearTimeout(timer)
  if (active.value && job.value && (job.value.status === 'Running' || job.value.status === 'Pending')) {
    timer = setTimeout(() => fetchDetail(true), 5000)
  }
}

const fetchDetail = async (silent = false) => {
  if (!clusterId.value || !namespace.value || !name.value) return
  if (!silent) loading.value = true
  try {
    const resp = await jobApi.getDetail({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })
    job.value = resp.code === 200 ? resp.data : null
  } catch {
    if (!silent) job.value = null
  } finally {
    loading.value = false
    schedulePoll()
  }
}

const handleBack = goBack

const goNamespace = () => router.push({ path: `/cluster/${clusterId.value}/namespace/detail`, query: { namespace: namespace.value } })
const goCronJob = () =>
  job.value?.ownerCronJob &&
  router.push({ path: `/cluster/${clusterId.value}/job/cronjob/detail`, query: { namespace: namespace.value, name: job.value.ownerCronJob } })

const restartText = (policy?: string) =>
  policy === 'OnFailure' ? '失败时重启（OnFailure）' : policy === 'Never' ? '不重启（Never）' : policy || '-'

const openYaml = async () => {
  try {
    yamlContent.value = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'job', namespace: namespace.value, name: name.value })
  } catch {
    return
  }
  yamlVisible.value = true
}

const handleRerun = async () => {
  if (!job.value) return
  try {
    await ElMessageBox.confirm(
      `将删除任务 "${job.value.name}" 当前的容器组，并按相同配置重新创建任务。本次运行结果会保存在运行记录中。`,
      '重新执行',
      { type: 'warning', confirmButtonText: '重新执行', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  rerunning.value = true
  try {
    await jobApi.rerun({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })
    ElMessage.success('已重新执行')
    setTab('records')
    await fetchDetail()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    rerunning.value = false
  }
}

const handleDelete = async () => {
  if (!job.value) return
  try {
    await ElMessageBox.confirm(`确定要删除任务 "${job.value.name}" 吗？它创建的容器组会一起删除，此操作不可恢复。`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await jobApi.delete({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })
    ElMessage.success('删除成功')
    router.push(`/cluster/${clusterId.value}/job/job`)
  } catch {
    // 请求拦截器已提示错误
  }
}

// 页面切走后停止轮询, 切回来时刷新
const active = ref(true)
onMounted(() => fetchDetail())
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
