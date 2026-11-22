<template>
  <div class="job-detail detail-page">
    <PageHeader
      :title="cj?.name || name || '定时任务详情'"
      :subtitle="cj ? cj.alias || cj.describe || `命名空间：${cj.namespace}` : '定时任务详情'"
      icon="cronjob"
      back
      @back="handleBack"
    >
      <template v-if="cj" #meta>
        <el-tag :type="cj.suspend ? 'info' : 'success'" effect="dark">{{ cj.suspend ? '已暂停' : '运行中' }}</el-tag>
        <el-tag type="info" effect="light" class="mono">{{ cj.schedule }}</el-tag>
      </template>
      <template v-if="cj">
        <el-button :loading="loading" @click="fetchDetail()">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" :loading="triggering" @click="handleTrigger">
          <el-icon><VideoPlay /></el-icon>
          立即执行
        </el-button>
        <el-button @click="toggleSuspend">
          <el-icon><component :is="cj.suspend ? CaretRight : VideoPause" /></el-icon>
          {{ cj.suspend ? '启用' : '暂停' }}
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

    <div v-if="loading && !cj" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="cj" class="detail-content">
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
              <span><el-link type="primary" @click="goNamespace">{{ cj.namespace }}</el-link></span>
            </div>
            <div class="jd-attr"><label>状态</label><span>{{ cj.suspend ? '已暂停' : '运行中' }}</span></div>
            <div class="jd-attr">
              <label>调度规则</label>
              <span><span class="mono">{{ cj.schedule }}</span><span v-if="scheduleDesc" class="jd-sub">{{ scheduleDesc }}</span></span>
            </div>
            <div class="jd-attr"><label>时区</label><span>{{ cj.timeZone || '集群默认' }}</span></div>
            <div class="jd-attr"><label>并发策略</label><span>{{ concurrencyText(cj.concurrencyPolicy) }}</span></div>
            <div class="jd-attr"><label>运行中的任务</label><span>{{ cj.active }}</span></div>
            <div class="jd-attr"><label>上次调度时间</label><span>{{ formatDateTime(cj.lastRunTime) }}</span></div>
            <div class="jd-attr"><label>上次成功时间</label><span>{{ formatDateTime(cj.lastSuccessfulTime) }}</span></div>
            <div class="jd-attr"><label>最大重试次数</label><span>{{ cj.backoffLimit }}</span></div>
            <div class="jd-attr"><label>容器组完成数量</label><span>{{ cj.completions }}</span></div>
            <div class="jd-attr"><label>并行容器组数量</label><span>{{ cj.parallelism }}</span></div>
            <div class="jd-attr">
              <label>最大运行时间（s）</label>
              <span>{{ cj.activeDeadlineSeconds ? formatSeconds(cj.activeDeadlineSeconds) : '-' }}</span>
            </div>
            <div class="jd-attr">
              <label>启动截止时间（s）</label>
              <span>{{ cj.startingDeadlineSeconds ? formatSeconds(cj.startingDeadlineSeconds) : '-' }}</span>
            </div>
            <div class="jd-attr">
              <label>保留任务数</label>
              <span>成功 {{ cj.successfulJobsHistoryLimit }} / 失败 {{ cj.failedJobsHistoryLimit }}</span>
            </div>
            <div class="jd-attr"><label>创建时间</label><span>{{ formatDateTime(cj.createTime) }}</span></div>
            <div class="jd-attr"><label>创建者</label><span>{{ cj.creator || '-' }}</span></div>
            <div class="jd-attr"><label>重启策略</label><span>{{ cj.restartPolicy || '-' }}</span></div>
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

      <!-- 运行记录: 定时任务创建的任务 -->
      <el-card v-if="activeTab === 'records'" class="info-card">
        <template #header>
          <div class="card-header">
            <span>运行记录</span>
            <el-tag size="small" type="info">{{ cj.jobs.length }}</el-tag>
            <span class="card-header__hint">
              按保留设置，最多保留 {{ cj.successfulJobsHistoryLimit }} 个成功、{{ cj.failedJobsHistoryLimit }} 个失败的任务
            </span>
          </div>
        </template>
        <el-table :data="jobRows" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
          <el-table-column label="任务名称" min-width="240" show-overflow-tooltip>
            <template #default="{ row }">
              <el-link type="primary" @click="goJob(row.name)">{{ row.name }}</el-link>
            </template>
          </el-table-column>
          <el-table-column label="状态" min-width="170">
            <template #default="{ row }">
              <el-tag :type="jobStatusTag(row.status)" size="small">{{ jobStatusText(row.status) }}</el-tag>
              <span v-if="row.message" class="jd-msg">{{ row.message }}</span>
            </template>
          </el-table-column>
          <el-table-column label="触发方式" width="100" align="center" header-align="center">
            <template #default="{ row }">{{ row.manual ? '手动' : '定时' }}</template>
          </el-table-column>
          <el-table-column label="完成数量" width="100" align="center" header-align="center">
            <template #default="{ row }">{{ row.succeeded }} / {{ row.completions }}</template>
          </el-table-column>
          <el-table-column label="开始时间" width="170">
            <template #default="{ row }">{{ formatDateTime(row.startTime) }}</template>
          </el-table-column>
          <el-table-column label="结束时间" width="170">
            <template #default="{ row }">{{ formatDateTime(row.completionTime) }}</template>
          </el-table-column>
          <el-table-column label="时长" width="110">
            <template #default="{ row }">{{ formatDuration(row.startTime, row.completionTime) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="90" align="center" header-align="center">
            <template #default="{ row }">
              <el-button link type="danger" size="small" @click="deleteJob(row.name)">删除</el-button>
            </template>
          </el-table-column>
          <template #empty>
            <el-empty description="还没有运行记录，可以点击「立即执行」运行一次" :image-size="72" />
          </template>
        </el-table>
        <ListPagination v-model:current-page="jobPage" v-model:page-size="jobPageSize" class="jd-pager" :total="jobTotal" :page-sizes="[10, 20, 50]" />
      </el-card>

      <!-- 资源状态 -->
      <template v-else-if="activeTab === 'resources'">
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>容器组</span>
              <el-tag size="small" type="info">{{ cj.pods.length }}</el-tag>
              <span class="card-header__hint">保留的任务创建的容器组</span>
            </div>
          </template>
          <JobPodsTable :pods="cj.pods" :cluster-id="clusterId" show-job />
        </el-card>
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>容器</span>
              <el-tag size="small" type="info">{{ cj.containers.length }}</el-tag>
              <span class="card-header__hint">任务模板中的容器</span>
            </div>
          </template>
          <ContainerTemplates :containers="cj.containers" />
        </el-card>
      </template>

      <el-card v-else-if="activeTab === 'metadata'" class="info-card">
        <template #header>
          <div class="card-header"><span>元数据</span></div>
        </template>
        <MetadataPanel :labels="cj.labels" :annotations="cj.annotations" />
      </el-card>

      <el-card v-else-if="activeTab === 'env'" class="info-card">
        <template #header>
          <div class="card-header"><span>环境变量</span></div>
        </template>
        <ContainerEnvList :containers="cj.containers" />
      </el-card>

      <el-card v-else class="info-card">
        <template #header>
          <div class="card-header">
            <span>事件</span>
            <el-tag size="small" type="info">{{ cj.events.length }}</el-tag>
          </div>
        </template>
        <EventsTable :events="cj.events" />
      </el-card>
    </div>

    <el-empty v-else description="定时任务不存在或已被删除" />

    <JobEditDialog v-model="editVisible" kind="cronjob" :cluster-id="clusterId" :target="cj" @success="fetchDetail()" />
    <YamlDialog
      v-model="yamlVisible"
      v-model:content="yamlContent"
      title="编辑YAML"
      subtitle="CronJob 配置"
      :info="[{ label: '集群', value: clusterName }, { label: '命名空间', value: cj?.namespace }, { label: '名称', value: cj?.name }, { label: '类型', value: 'CronJob' }]"
      :editor-title="`${cj?.name} - CronJob YAML`"
      :filename="`${cj?.name}-cronjob.yaml`"
      :apply-target="cj ? { clusterId, kind: 'cronjob', namespace: cj.namespace, name: cj.name } : null"
      @applied="fetchDetail()"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, CaretRight, Delete, Document, Edit, Refresh, VideoPause, VideoPlay } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import JobEditDialog from '@/components/job/JobEditDialog.vue'
import JobPodsTable from '@/components/job/JobPodsTable.vue'
import ContainerTemplates from '@/components/job/ContainerTemplates.vue'
import ContainerEnvList from '@/components/job/ContainerEnvList.vue'
import MetadataPanel from '@/components/job/MetadataPanel.vue'
import EventsTable from '@/components/job/EventsTable.vue'
import { cronJobApi, type CronJobDetailVO } from '@/api/cronjob'
import { jobApi } from '@/api/job'
import { fetchResourceYaml } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useClientPage } from '@/composables/useClientPage'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { concurrencyText, describeCron, formatDateTime, formatDuration, formatSeconds, jobStatusTag, jobStatusText } from '@/utils/job'
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
const cj = ref<CronJobDetailVO | null>(null)
const attrOpen = ref(true)
const editVisible = ref(false)
const triggering = ref(false)
const yamlVisible = ref(false)
const yamlContent = ref('')

const scheduleDesc = computed(() => describeCron(cj.value?.schedule))

const activeTab = ref<TabKey>('records')
const tabs = computed(() => [
  { key: 'records' as TabKey, label: '运行记录', count: cj.value?.jobs.length },
  { key: 'resources' as TabKey, label: '资源状态', count: cj.value?.pods.length },
  { key: 'metadata' as TabKey, label: '元数据' },
  { key: 'env' as TabKey, label: '环境变量' },
  { key: 'events' as TabKey, label: '事件', count: cj.value?.events.length }
])
// 页签只是页面内状态, 不写入地址: 地址变化会让标签栏和 keep-alive 当成新页面
const setTab = (key: TabKey) => {
  activeTab.value = key
}

const { page: jobPage, pageSize: jobPageSize, total: jobTotal, rows: jobRows } = useClientPage(() => cj.value?.jobs || [])

// 有任务在运行时定时刷新
let timer: ReturnType<typeof setTimeout> | undefined
const schedulePoll = () => {
  clearTimeout(timer)
  if (active.value && cj.value && cj.value.jobs.some(j => j.status === 'Running' || j.status === 'Pending')) {
    timer = setTimeout(() => fetchDetail(true), 5000)
  }
}

const target = () => ({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })

const fetchDetail = async (silent = false) => {
  if (!clusterId.value || !namespace.value || !name.value) return
  if (!silent) loading.value = true
  try {
    const resp = await cronJobApi.getDetail(target())
    cj.value = resp.code === 200 ? resp.data : null
  } catch {
    if (!silent) cj.value = null
  } finally {
    loading.value = false
    schedulePoll()
  }
}

const handleBack = goBack

const goNamespace = () => router.push({ path: `/cluster/${clusterId.value}/namespace/detail`, query: { namespace: namespace.value } })
const goJob = (jobName: string) => router.push({ path: `/cluster/${clusterId.value}/job/job/detail`, query: { namespace: namespace.value, name: jobName } })

const openYaml = async () => {
  try {
    yamlContent.value = await fetchResourceYaml({ ...target(), kind: 'cronjob' })
  } catch {
    return
  }
  yamlVisible.value = true
}

const handleTrigger = async () => {
  if (!cj.value) return
  try {
    await ElMessageBox.confirm(`将按任务模板立即创建一个任务，不影响原有的调度计划。`, `立即执行 "${cj.value.name}"`, {
      type: 'info',
      confirmButtonText: '立即执行',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  triggering.value = true
  try {
    const resp = await cronJobApi.trigger(target())
    ElMessage.success(`已创建任务 ${resp.data?.jobName || ''}`)
    setTab('records')
    await fetchDetail()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    triggering.value = false
  }
}

const toggleSuspend = async () => {
  if (!cj.value) return
  const suspend = !cj.value.suspend
  try {
    await ElMessageBox.confirm(
      suspend ? '暂停后不会再按计划创建新任务，已经在运行的任务不受影响。' : '启用后会按调度规则继续创建任务。',
      suspend ? '暂停定时任务' : '启用定时任务',
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await cronJobApi.suspend({ ...target(), suspend })
    ElMessage.success(suspend ? '已暂停' : '已启用')
    fetchDetail()
  } catch {
    // 请求拦截器已提示错误
  }
}

const deleteJob = async (jobName: string) => {
  try {
    await ElMessageBox.confirm(`确定要删除任务 "${jobName}" 吗？它创建的容器组会一起删除。`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await jobApi.delete({ clusterId: clusterId.value, namespace: namespace.value, name: jobName })
    ElMessage.success('删除成功')
    fetchDetail()
  } catch {
    // 请求拦截器已提示错误
  }
}

const handleDelete = async () => {
  if (!cj.value) return
  try {
    await ElMessageBox.confirm(
      `确定要删除定时任务 "${cj.value.name}" 吗？它创建的 ${cj.value.jobs.length} 个任务及容器组会一起删除，此操作不可恢复。`,
      '确认删除',
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await cronJobApi.delete(target())
    ElMessage.success('删除成功')
    router.push(`/cluster/${clusterId.value}/job/cronjob`)
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
