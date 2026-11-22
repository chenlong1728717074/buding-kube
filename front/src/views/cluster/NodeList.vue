<template>
  <div class="node-list list-page">
    <PageHeader title="节点管理" subtitle="查看节点的运行状态与资源使用，管理调度、标签和污点" icon="node" />

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="名称 / IP">
          <el-input v-model="searchForm.keyword" placeholder="输入节点名称或 IP" clearable style="width: 220px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部" clearable style="width: 130px" @change="handleSearch">
            <el-option label="就绪" value="Ready" />
            <el-option label="未就绪" value="NotReady" />
            <el-option label="未知" value="Unknown" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select v-model="searchForm.role" placeholder="全部" clearable style="width: 130px" @change="handleSearch">
            <el-option label="控制节点" value="control-plane" />
            <el-option label="工作节点" value="worker" />
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
        没有获取到节点的实时用量，请确认集群已安装 metrics-server（kubectl top node 可以正常使用）；「分配」为容器组预留资源的合计，不依赖 metrics-server。
      </el-alert>
      <el-table v-loading="loading" :data="nodeList" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
        <el-table-column label="名称" min-width="200" header-align="center" align="center">
          <template #default="{ row }">
            <div class="name-cell">
              <el-link type="primary" @click="openDetail(row)">{{ row.hostname }}</el-link>
              <span class="name-cell__alias mono">{{ row.ip || '-' }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="110" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag :type="nodeStatusTag(row.status)" size="small">{{ nodeStatusText(row.status) }}</el-tag>
            <span v-if="row.unSchedule" class="sub-text warn-text">已停止调度</span>
          </template>
        </el-table-column>
        <el-table-column label="角色" min-width="110" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag v-for="r in row.roles" :key="r" size="small" :type="r === 'worker' ? 'info' : 'warning'" effect="plain" class="role-tag">{{ roleText(r) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="CPU" min-width="190" header-align="center" align="center">
          <template #default="{ row }">
            <NodeResourceCell kind="cpu" :used="row.cpuUsage" :requests="row.cpuRequests" :allocatable="row.cpuAllocatable" />
            <span class="sub-text">可分配 {{ formatCpu(row.cpuAllocatable) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="内存" min-width="190" header-align="center" align="center">
          <template #default="{ row }">
            <NodeResourceCell kind="memory" :used="row.memoryUsage" :requests="row.memoryRequests" :allocatable="row.memoryAllocatable" />
            <span class="sub-text">可分配 {{ formatMemory(row.memoryAllocatable) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="容器组数量" min-width="110" header-align="center" align="center">
          <template #default="{ row }">
            <span :class="{ 'warn-text': row.podAllocatable && row.podCount / row.podAllocatable >= 0.9 }">{{ row.podCount }} / {{ row.podAllocatable }}</span>
          </template>
        </el-table-column>
        <el-table-column label="污点" min-width="70" header-align="center" align="center" prop="taintCount" />
        <el-table-column label="Kubelet 版本" min-width="120" header-align="center" align="center">
          <template #default="{ row }"><span class="mono">{{ row.kubeletVersion || '-' }}</span></template>
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
                    <el-dropdown-item command="schedule">{{ row.unSchedule ? '启用调度' : '停止调度' }}</el-dropdown-item>
                    <el-dropdown-item command="drain">排空节点</el-dropdown-item>
                    <el-dropdown-item command="labels">管理标签</el-dropdown-item>
                    <el-dropdown-item command="taints">管理污点</el-dropdown-item>
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
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        @size-change="handleSearch"
        @current-change="fetchList()"
      />
    </el-card>

    <NodeLabelsDialog v-model="labelsVisible" :cluster-id="clusterId" :hostname="current?.hostname || ''" @success="fetchList(true)" />
    <NodeTaintsDialog v-model="taintsVisible" :cluster-id="clusterId" :hostname="current?.hostname || ''" @success="fetchList(true)" />
    <NodeDrainDialog v-model="drainVisible" :cluster-id="clusterId" :hostname="current?.hostname || ''" @success="fetchList(true)" />
    <YamlDialog
      v-model="yamlVisible"
      title="编辑YAML"
      subtitle="Node 配置"
      :info="[{ label: '集群', value: clusterName }, { label: '名称', value: current?.hostname }, { label: '类型', value: 'Node' }]"
      v-model:content="yamlContent"
      :editor-title="`${current?.hostname} - Node YAML`"
      :filename="`${current?.hostname}-node.yaml`"
      tip="标签、污点和调度状态建议通过对应的菜单修改；status 由 kubelet 维护，不会显示和保存"
      :apply-target="current ? { clusterId, kind: 'node', name: current.hostname } : null"
      @applied="fetchList(true)"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onActivated, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MoreFilled, Refresh, Search } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import NodeResourceCell from '@/components/node/NodeResourceCell.vue'
import NodeLabelsDialog from '@/components/node/NodeLabelsDialog.vue'
import NodeTaintsDialog from '@/components/node/NodeTaintsDialog.vue'
import NodeDrainDialog from '@/components/node/NodeDrainDialog.vue'
import { nodeApi, type NodeVO } from '@/api/node'
import { fetchResourceYaml } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { formatDateTime } from '@/utils/job'
import { formatCpu, formatMemory } from '@/utils/resource'
import { nodeStatusTag, nodeStatusText, roleText } from '@/utils/node'
import '@/assets/styles/config-editor.css'

const router = useRouter()
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)
const clusterName = computed(() => clusterStore.currentClusterName)

const loading = ref(false)
const nodeList = ref<NodeVO[]>([])
const searchForm = reactive({ keyword: '', status: '', role: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const labelsVisible = ref(false)
const taintsVisible = ref(false)
const drainVisible = ref(false)
const yamlVisible = ref(false)
const yamlContent = ref('')
const current = ref<NodeVO | null>(null)

const metricsMissing = computed(() => nodeList.value.length > 0 && nodeList.value.every(n => n.cpuUsage === null || n.cpuUsage === undefined))

const fetchList = async (silent = false) => {
  if (!clusterId.value) {
    nodeList.value = []
    pagination.total = 0
    return
  }
  if (!silent) loading.value = true
  try {
    const resp = await nodeApi.getList({
      clusterId: clusterId.value,
      keyword: searchForm.keyword.trim() || undefined,
      status: searchForm.status || undefined,
      role: searchForm.role || undefined,
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    nodeList.value = resp.data?.items || []
    pagination.total = resp.data?.total || 0
  } catch {
    if (!silent) nodeList.value = []
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  Object.assign(searchForm, { keyword: '', status: '', role: '' })
  handleSearch()
}

const openDetail = (row: NodeVO) => {
  router.push({ path: `/cluster/${clusterId.value}/node/detail`, query: { hostname: row.hostname } })
}

const toggleSchedule = async (row: NodeVO) => {
  const unSchedule = !row.unSchedule
  try {
    await ElMessageBox.confirm(
      unSchedule
        ? `停止调度后，新的容器组不会再调度到 ${row.hostname}，已运行的容器组不受影响。如需迁移已有容器组，请使用「排空节点」。`
        : `启用后，新的容器组可以调度到 ${row.hostname}。`,
      unSchedule ? '停止调度' : '启用调度',
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await nodeApi.setSchedule({ clusterId: clusterId.value, hostname: row.hostname, unSchedule })
    ElMessage.success(unSchedule ? '已停止调度' : '已启用调度')
    fetchList(true)
  } catch {
    // 请求拦截器已提示错误
  }
}

const handleMore = async (command: string, row: NodeVO) => {
  current.value = row
  if (command === 'schedule') toggleSchedule(row)
  else if (command === 'drain') drainVisible.value = true
  else if (command === 'labels') labelsVisible.value = true
  else if (command === 'taints') taintsVisible.value = true
  else if (command === 'yaml') {
    try {
      yamlContent.value = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'node', name: row.hostname })
      yamlVisible.value = true
    } catch {
      // 请求拦截器已提示错误
    }
  }
}

const handleDelete = async (row: NodeVO) => {
  try {
    await ElMessageBox.confirm(
      `删除的只是集群中的 Node 对象，节点上的容器不会停止，kubelet 重启后节点会重新注册。下线节点前建议先「排空节点」。确定删除 "${row.hostname}" 吗？`,
      '确认删除',
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await nodeApi.delete({ clusterId: clusterId.value, name: row.hostname })
    ElMessage.success('删除成功')
    fetchList()
  } catch {
    // 请求拦截器已提示错误
  }
}

watch(
  clusterId,
  id => {
    pagination.page = 1
    if (id) fetchList()
    else {
      nodeList.value = []
      pagination.total = 0
    }
  },
  { immediate: true }
)

// 从详情页返回时静默刷新
let activated = false
onActivated(() => {
  if (activated) fetchList(true)
  activated = true
})
</script>

<style scoped>
.node-list {
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

.role-tag + .role-tag {
  margin-left: 4px;
}
</style>
