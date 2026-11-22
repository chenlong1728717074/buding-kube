<template>
  <div>
    <el-table :data="rows" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
      <el-table-column label="容器组名称" min-width="240" show-overflow-tooltip>
        <template #default="{ row }">
          <el-link type="primary" @click="openPod(row)">{{ row.name }}</el-link>
        </template>
      </el-table-column>
      <el-table-column v-if="showNamespace" label="命名空间" min-width="140" show-overflow-tooltip prop="namespace" />
      <el-table-column v-if="showJob" label="所属任务" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">{{ row.jobName || '-' }}</template>
      </el-table-column>
      <el-table-column label="状态" min-width="150">
        <template #default="{ row }">
          <el-tag :type="podStatusTag(row.status)" size="small">{{ podStatusText(row.status) }}</el-tag>
          <span v-if="row.message" class="jd-msg" :title="row.message">{{ row.message }}</span>
        </template>
      </el-table-column>
      <el-table-column label="就绪" width="80" align="center" header-align="center" prop="ready" />
      <el-table-column label="CPU" width="130" align="center" header-align="center">
        <template #default="{ row }">
          <ResourceUsageBar kind="cpu" compact :used="row.cpuUsage" :request="row.cpuRequest" :limit="row.cpuLimit" />
        </template>
      </el-table-column>
      <el-table-column label="内存" width="140" align="center" header-align="center">
        <template #default="{ row }">
          <ResourceUsageBar kind="memory" compact :used="row.memoryUsage" :request="row.memoryRequest" :limit="row.memoryLimit" />
        </template>
      </el-table-column>
      <el-table-column label="重启次数" width="96" align="center" header-align="center" prop="restarts" />
      <el-table-column v-if="!hideNode" label="节点" min-width="150" show-overflow-tooltip>
        <template #default="{ row }">{{ row.nodeName || '-' }}</template>
      </el-table-column>
      <el-table-column label="容器组 IP" width="140">
        <template #default="{ row }"><span class="mono">{{ row.podIP || '-' }}</span></template>
      </el-table-column>
      <el-table-column label="创建时间" width="170">
        <template #default="{ row }">{{ formatDateTime(row.createTime) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="90" align="center" header-align="center">
        <template #default="{ row }">
          <el-button link type="primary" size="small" :disabled="row.status !== 'Running'" @click="openTerminal(row)">终端</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="暂无容器组" :image-size="72" />
      </template>
    </el-table>
    <ListPagination
      v-if="total > 10"
      v-model:current-page="page"
      v-model:page-size="pageSize"
      class="jd-pager"
      :total="total"
      :page-sizes="[10, 20, 50]"
    />
    <PodTerminal v-model="terminalVisible" :cluster-id="clusterId" :namespace="terminalPod?.namespace || ''" :pod-name="terminalPod?.name || ''" />
  </div>
</template>

<script setup lang="ts">
import { ref, toRef } from 'vue'
import { useRouter } from 'vue-router'
import ListPagination from '@/components/ListPagination.vue'
import PodTerminal from '@/components/pod/PodTerminal.vue'
import ResourceUsageBar from '@/components/ResourceUsageBar.vue'
import type { JobPodVO } from '@/api/job'
import { useClientPage } from '@/composables/useClientPage'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { formatDateTime, podStatusTag, podStatusText } from '@/utils/job'

const props = withDefaults(
  defineProps<{ pods: JobPodVO[]; clusterId: string; showJob?: boolean; showNamespace?: boolean; hideNode?: boolean }>(),
  { showJob: false, showNamespace: false, hideNode: false }
)

const router = useRouter()
const { page, pageSize, total, rows } = useClientPage(toRef(props, 'pods'))

const openPod = (row: JobPodVO) => {
  router.push({ path: `/cluster/${props.clusterId}/pod/detail`, query: { namespace: row.namespace, name: row.name } })
}

const terminalVisible = ref(false)
const terminalPod = ref<JobPodVO | null>(null)
const openTerminal = (row: JobPodVO) => {
  terminalPod.value = row
  terminalVisible.value = true
}
</script>
