<template>
  <div class="cluster-nodes">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="handleBack">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1>节点管理</h1>
        <el-tag v-if="clusterName" type="info">{{ clusterName }}</el-tag>
      </div>
      <div class="header-right">
        <el-button @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 搜索表单 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="节点名称">
          <el-input 
            v-model="searchForm.name" 
            placeholder="请输入节点名称" 
            clearable
            style="width: 200px;"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select 
            v-model="searchForm.status" 
            placeholder="请选择状态" 
            clearable
            style="width: 150px;"
          >
            <el-option label="就绪" value="Ready" />
            <el-option label="未就绪" value="NotReady" />
            <el-option label="未知" value="Unknown" />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select 
            v-model="searchForm.role" 
            placeholder="请选择角色" 
            clearable
            style="width: 150px;"
          >
            <el-option label="Master" value="master" />
            <el-option label="Worker" value="worker" />
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

    <!-- 节点列表 -->
    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>节点列表 ({{ pagination.total }})</span>
        </div>
      </template>
      
      <el-table 
        v-loading="loading" 
        :data="nodeList" 
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="name" label="节点名称" min-width="200">
          <template #default="{ row }">
            <div class="node-name">
              <el-icon class="node-icon"><Monitor /></el-icon>
              <span>{{ row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="role" label="角色" width="100">
          <template #default="{ row }">
            <el-tag 
              :type="row.role === 'master' ? 'warning' : 'info'" 
              size="small"
            >
              {{ row.role === 'master' ? 'Master' : 'Worker' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="120" />
        <el-table-column prop="internalIP" label="内部IP" width="140" />
        <el-table-column prop="externalIP" label="外部IP" width="140">
          <template #default="{ row }">
            <span>{{ row.externalIP || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="资源使用" width="200">
          <template #default="{ row }">
            <div class="resource-usage">
              <div class="usage-item">
                <span class="usage-label">CPU:</span>
                <el-progress 
                  :percentage="row.cpuUsage || 0" 
                  :stroke-width="6"
                  :show-text="false"
                  :color="getProgressColor(row.cpuUsage || 0)"
                />
                <span class="usage-text">{{ row.cpuUsage || 0 }}%</span>
              </div>
              <div class="usage-item">
                <span class="usage-label">内存:</span>
                <el-progress 
                  :percentage="row.memoryUsage || 0" 
                  :stroke-width="6"
                  :show-text="false"
                  :color="getProgressColor(row.memoryUsage || 0)"
                />
                <span class="usage-text">{{ row.memoryUsage || 0 }}%</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="podCount" label="Pod数量" width="100">
          <template #default="{ row }">
            <el-button text @click="handleViewPods(row)">
              {{ row.podCount || 0 }}
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleViewDetail(row)">
              详情
            </el-button>
            <el-button size="small" @click="handleViewPods(row)">
              Pod
            </el-button>
            <el-dropdown @command="(command) => handleMoreActions(command, row)">
              <el-button size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="cordon">
                    {{ row.schedulable ? '禁止调度' : '允许调度' }}
                  </el-dropdown-item>
                  <el-dropdown-item command="drain">
                    驱逐Pod
                  </el-dropdown-item>
                  <el-dropdown-item command="label">
                    管理标签
                  </el-dropdown-item>
                  <el-dropdown-item command="taint">
                    管理污点
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 节点详情对话框 -->
    <UnifiedDialog 
      v-model="detailDialogVisible" 
      title="节点详情" 
      subtitle="基础信息与资源"
      width="80%"
    >
      <div v-if="selectedNode" class="node-detail">
        <!-- 基本信息 -->
        <el-descriptions title="基本信息" :column="2" border>
          <el-descriptions-item label="节点名称">{{ selectedNode.name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(selectedNode.status)">
              {{ getStatusText(selectedNode.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="角色">
            <el-tag :type="selectedNode.role === 'master' ? 'warning' : 'info'">
              {{ selectedNode.role === 'master' ? 'Master' : 'Worker' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="版本">{{ selectedNode.version }}</el-descriptions-item>
          <el-descriptions-item label="内部IP">{{ selectedNode.internalIP }}</el-descriptions-item>
          <el-descriptions-item label="外部IP">{{ selectedNode.externalIP || '-' }}</el-descriptions-item>
          <el-descriptions-item label="操作系统">{{ selectedNode.osImage || '-' }}</el-descriptions-item>
          <el-descriptions-item label="内核版本">{{ selectedNode.kernelVersion || '-' }}</el-descriptions-item>
          <el-descriptions-item label="容器运行时">{{ selectedNode.containerRuntime || '-' }}</el-descriptions-item>
          <el-descriptions-item label="可调度">{{ selectedNode.schedulable ? '是' : '否' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ selectedNode.createTime }}</el-descriptions-item>
        </el-descriptions>

        <!-- 资源信息 -->
        <el-descriptions title="资源信息" :column="2" border style="margin-top: 20px;">
          <el-descriptions-item label="CPU容量">{{ selectedNode.cpuCapacity || '-' }}</el-descriptions-item>
          <el-descriptions-item label="CPU可分配">{{ selectedNode.cpuAllocatable || '-' }}</el-descriptions-item>
          <el-descriptions-item label="内存容量">{{ selectedNode.memoryCapacity || '-' }}</el-descriptions-item>
          <el-descriptions-item label="内存可分配">{{ selectedNode.memoryAllocatable || '-' }}</el-descriptions-item>
          <el-descriptions-item label="存储容量">{{ selectedNode.storageCapacity || '-' }}</el-descriptions-item>
          <el-descriptions-item label="存储可分配">{{ selectedNode.storageAllocatable || '-' }}</el-descriptions-item>
          <el-descriptions-item label="Pod容量">{{ selectedNode.podCapacity || '-' }}</el-descriptions-item>
          <el-descriptions-item label="Pod数量">{{ selectedNode.podCount || 0 }}</el-descriptions-item>
        </el-descriptions>

        <!-- 标签 -->
        <div style="margin-top: 20px;">
          <h4>标签</h4>
          <div class="labels-container">
            <el-tag 
              v-for="(value, key) in selectedNode.labels" 
              :key="key"
              style="margin: 4px;"
            >
              {{ key }}: {{ value }}
            </el-tag>
          </div>
        </div>

        <!-- 污点 -->
        <div style="margin-top: 20px;">
          <h4>污点</h4>
          <div class="taints-container">
            <el-tag 
              v-for="taint in selectedNode.taints" 
              :key="taint.key"
              type="warning"
              style="margin: 4px;"
            >
              {{ taint.key }}={{ taint.value }}:{{ taint.effect }}
            </el-tag>
            <span v-if="!selectedNode.taints || selectedNode.taints.length === 0">无污点</span>
          </div>
        </div>
      </div>
    </UnifiedDialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import UnifiedDialog from '@/components/UnifiedDialog.vue'
import { 
  ArrowLeft, 
  Refresh, 
  Search, 
  Monitor, 
  ArrowDown
} from '@element-plus/icons-vue'
import { clusterApi } from '@/api/cluster'

const route = useRoute()
const router = useRouter()

const clusterName = route.params.id as string
const loading = ref(false)
const clusterDisplayName = ref('')

// 搜索表单
const searchForm = reactive({
  name: '',
  status: '',
  role: ''
})

// 节点列表
const nodeList = ref([])
const selectedNodes = ref([])

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 详情对话框
const detailDialogVisible = ref(false)
const selectedNode = ref(null)

// 获取节点列表
const fetchNodeList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.size,
      ...searchForm
    }
    const response = await clusterApi.getNodes(clusterName, params)
    nodeList.value = response.data.list || []
    pagination.total = response.data.total || 0
    
    ElMessage.success('节点列表加载成功')
  } catch (error) {
    console.error('获取节点列表失败:', error)
    ElMessage.error('获取节点列表失败')
  } finally {
    loading.value = false
  }
}

// 获取集群信息
const fetchClusterInfo = async () => {
  try {
    const response = await clusterApi.getCluster(clusterName)
    clusterDisplayName.value = response.data.name
  } catch (error) {
    console.error('获取集群信息失败:', error)
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchNodeList()
}

// 重置
const handleReset = () => {
  Object.assign(searchForm, {
    name: '',
    status: '',
    role: ''
  })
  pagination.page = 1
  fetchNodeList()
}

// 分页大小改变
const handleSizeChange = (size: number) => {
  pagination.size = size
  pagination.page = 1
  fetchNodeList()
}

// 当前页改变
const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchNodeList()
}

// 选择改变
const handleSelectionChange = (selection: any[]) => {
  selectedNodes.value = selection
}

// 返回
const handleBack = () => {
  router.back()
}

// 刷新
const handleRefresh = () => {
  fetchNodeList()
}

// 查看详情
const handleViewDetail = (node: any) => {
  router.push({
    path: '/node/detail',
    query: {
      clusterName: clusterName,
      nodeName: node.name
    }
  })
}

// 查看Pod
const handleViewPods = (node: any) => {
  router.push(`/cluster/pods/${clusterName}?node=${node.name}`)
}

// 更多操作
const handleMoreActions = async (command: string, node: any) => {
  switch (command) {
    case 'cordon':
      await handleCordonNode(node)
      break
    case 'drain':
      await handleDrainNode(node)
      break
    case 'label':
      handleManageLabels(node)
      break
    case 'taint':
      handleManageTaints(node)
      break
  }
}

// 禁止/允许调度
const handleCordonNode = async (node: any) => {
  const action = node.schedulable ? '禁止' : '允许'
  try {
    await ElMessageBox.confirm(
      `确定要${action}节点 ${node.name} 的调度吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await clusterApi.cordonNode(clusterName, node.name, !node.schedulable)
    ElMessage.success(`${action}调度成功`)
    fetchNodeList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}调度失败`)
    }
  }
}

// 驱逐Pod
const handleDrainNode = async (node: any) => {
  try {
    await ElMessageBox.confirm(
      `确定要驱逐节点 ${node.name} 上的所有Pod吗？此操作会将Pod重新调度到其他节点。`,
      '确认驱逐',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await clusterApi.drainNode(clusterName, node.name)
    ElMessage.success('驱逐Pod成功')
    fetchNodeList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('驱逐Pod失败')
    }
  }
}

// 管理标签
const handleManageLabels = (node: any) => {
  ElMessage.info('标签管理功能开发中')
}

// 管理污点
const handleManageTaints = (node: any) => {
  ElMessage.info('污点管理功能开发中')
}

// 关闭详情对话框
const handleDetailDialogClose = () => {
  detailDialogVisible.value = false
  selectedNode.value = null
}

// 获取状态类型
const getStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    Ready: 'success',
    NotReady: 'danger',
    Unknown: 'warning'
  }
  return statusMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    Ready: '就绪',
    NotReady: '未就绪',
    Unknown: '未知'
  }
  return statusMap[status] || status
}

// 获取进度条颜色
const getProgressColor = (percentage: number) => {
  if (percentage < 60) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}

// 初始化
onMounted(() => {
  fetchClusterInfo()
  fetchNodeList()
})
</script>

<style scoped>
.cluster-nodes {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-left h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.search-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.node-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.node-icon {
  color: #409eff;
}

.resource-usage {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.usage-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.usage-label {
  width: 30px;
  color: #606266;
}

.usage-text {
  width: 35px;
  text-align: right;
  color: #606266;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.node-detail {
  max-height: 600px;
  overflow-y: auto;
}

.labels-container,
.taints-container {
  min-height: 40px;
  padding: 8px;
  background: #f5f7fa;
  border-radius: 4px;
}

@media (max-width: 768px) {
  .resource-usage {
    display: none;
  }
}
</style>
