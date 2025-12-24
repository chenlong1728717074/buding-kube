<template>
  <div class="cluster-list-page">
    <!-- 页面头部 -->
    <div class="page-header">
      <div>
        <h1 class="page-title">集群管理</h1>
        <p class="page-desc">管理和监控您的 Kubernetes 集群，简化运维操作</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" @click="showAddDialog = true">
          <el-icon><Plus /></el-icon>
          添加集群
        </el-button>
      </div>
    </div>

    <!-- 搜索和筛选 -->
    <div class="search-bar">
      <el-input
        v-model="searchParams.keyword"
        placeholder="搜索集群名称"
        clearable
        @clear="handleSearch"
        @keyup.enter="handleSearch"
        style="width: 300px"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-select
        v-model="searchParams.status"
        placeholder="状态筛选"
        clearable
        @change="handleSearch"
        style="width: 150px; margin-left: 12px"
      >
        <el-option label="全部状态" value="" />
        <el-option label="活跃" value="Active" />
        <el-option label="运行中" value="Running" />
        <el-option label="异常" value="Error" />
        <el-option label="失败" value="Failed" />
      </el-select>
      <el-button type="primary" @click="handleSearch" style="margin-left: 12px">
        <el-icon><Search /></el-icon>
        搜索
      </el-button>
      <el-button @click="handleReset">
        <el-icon><Refresh /></el-icon>
        重置
      </el-button>
    </div>

    <!-- 集群卡片网格 -->
    <div v-loading="loading" class="cluster-grid">
      <el-empty v-if="clusters.length === 0 && !loading" description="暂无集群数据" />
      <div 
        v-for="cluster in clusters" 
        :key="cluster.id" 
        class="cluster-card"
        @click="enterCluster(cluster)"
      >
        <div class="cluster-header">
          <div class="cluster-icon">
            <el-icon :size="24"><Monitor /></el-icon>
          </div>
          <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, cluster)" @click.stop>
            <el-button text circle class="more-btn" @click.stop>
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="edit">
                  <el-icon><Edit /></el-icon>
                  编辑
                </el-dropdown-item>
                <el-dropdown-item command="delete" divided>
                  <el-icon><Delete /></el-icon>
                  删除
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>

        <div class="cluster-body">
          <h3 class="cluster-name">
            {{ cluster.alias || cluster.name }}
            <span v-if="cluster.alias" class="cluster-name-sub">{{ cluster.name }}</span>
          </h3>
          <div class="cluster-status">
            <el-tag :type="getStatusType(cluster.status)" size="small">
              {{ getStatusText(cluster.status) }}
            </el-tag>
          </div>
          <div v-if="cluster.describe" class="cluster-describe">
            {{ cluster.describe }}
          </div>
          <div class="cluster-info">
            <div class="info-item" v-if="cluster.apiServer && cluster.apiServer !== '-'">
              <span class="info-label">API Server</span>
              <span class="info-value">{{ cluster.apiServer }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">版本</span>
              <span class="info-value">{{ cluster.version || '-' }}</span>
            </div>
          </div>
          <div class="cluster-stats">
            <div class="stat">
              <span class="stat-value">{{ cluster.nodeCount || 0 }}</span>
              <span class="stat-label">节点</span>
            </div>
            <div class="stat">
              <span class="stat-value">{{ cluster.podCount || 0 }}</span>
              <span class="stat-label">Pod</span>
            </div>
            <div class="stat">
              <span class="stat-value">{{ cluster.namespaceCount || 0 }}</span>
              <span class="stat-label">命名空间</span>
            </div>
          </div>
        </div>

        <div class="cluster-footer">
          <el-button type="primary" size="small" @click.stop="enterCluster(cluster)">
            进入集群
            <el-icon><ArrowRight /></el-icon>
          </el-button>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination-container" v-if="total > 0">
      <el-pagination
        v-model:current-page="searchParams.page"
        v-model:page-size="searchParams.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </div>

    <!-- 添加集群对话框 -->
    <el-dialog
      v-model="showAddDialog"
      title="添加集群"
      width="800px"
      :close-on-click-modal="false"
    >
      <el-form :model="clusterForm" :rules="rules" ref="formRef" label-width="100px" label-position="top">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="集群名称" prop="name">
              <el-input v-model="clusterForm.name" placeholder="请输入集群名称（唯一标识）" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="集群别名">
              <el-input v-model="clusterForm.alias" placeholder="请输入集群别名（显示名称）" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="集群描述">
          <el-input v-model="clusterForm.describe" type="textarea" :rows="2" placeholder="请输入集群描述" />
        </el-form-item>
        <el-form-item label="Kubeconfig 配置" prop="config">
          <el-input 
            v-model="clusterForm.config" 
            type="textarea" 
            :rows="12" 
            placeholder="请粘贴完整的 kubeconfig 配置文件内容"
            style="font-family: 'SF Mono', Monaco, Consolas, monospace; font-size: 13px;"
          />
          <div style="margin-top: 8px; font-size: 12px; color: #6b7280;">
            提示：请确保 kubeconfig 配置文件格式正确，包含完整的集群、用户和上下文信息
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false" size="large">取消</el-button>
        <el-button type="primary" @click="handleAddCluster" :loading="submitting" size="large">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑集群对话框 -->
    <el-dialog
      v-model="showEditDialog"
      title="编辑集群"
      width="800px"
      :close-on-click-modal="false"
    >
      <el-form :model="editClusterForm" :rules="rules" ref="editFormRef" label-width="100px" label-position="top">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="集群名称" prop="name">
              <el-input v-model="editClusterForm.name" disabled placeholder="集群名称不可修改" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="集群别名">
              <el-input v-model="editClusterForm.alias" placeholder="请输入集群别名（显示名称）" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="集群描述">
          <el-input v-model="editClusterForm.describe" type="textarea" :rows="2" placeholder="请输入集群描述" />
        </el-form-item>
        <el-form-item label="Kubeconfig 配置" prop="config">
          <el-input 
            v-model="editClusterForm.config" 
            type="textarea" 
            :rows="12" 
            placeholder="如需更新配置，请粘贴完整的 kubeconfig 配置文件内容"
            style="font-family: 'SF Mono', Monaco, Consolas, monospace; font-size: 13px;"
          />
          <div style="margin-top: 8px; font-size: 12px; color: #6b7280;">
            提示：留空则不更新配置，仅更新别名和描述
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false" size="large">取消</el-button>
        <el-button type="primary" @click="handleEditCluster" :loading="submitting" size="large">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Monitor, Plus, Edit, Delete, MoreFilled, ArrowRight, Search, Refresh } from '@element-plus/icons-vue'
import { useClusterStore } from '@/stores/cluster'
import { clusterApi } from '@/api/cluster'

const router = useRouter()
const clusterStore = useClusterStore()

const clusters = ref<any[]>([])
const loading = ref(false)
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const submitting = ref(false)
const formRef = ref()
const editFormRef = ref()
const total = ref(0)

// 搜索参数
const searchParams = ref({
  page: 1,
  pageSize: 12,
  keyword: '',
  status: 'Active' // 默认筛选活跃状态
})

const clusterForm = ref({
  name: '',
  alias: '',
  describe: '',
  config: ''
})

const editClusterForm = ref({
  name: '',
  alias: '',
  describe: '',
  config: ''
})

const rules = {
  name: [{ required: true, message: '请输入集群名称', trigger: 'blur' }],
  config: [{ required: true, message: '请输入集群配置', trigger: 'blur' }]
}

// 加载集群列表
const loadClusters = async () => {
  loading.value = true
  try {
    const response = await clusterApi.getClusters({
      page: searchParams.value.page,
      pageSize: searchParams.value.pageSize,
      keyword: searchParams.value.keyword || undefined,
      status: searchParams.value.status || undefined
    })
    
    // 正确解析后端返回的数据
    if (response && response.data) {
      // 后端返回的是 items 而不是 list
      const { items, list, total: totalCount } = response.data
      const clusterList = items || list || []
      
      clusters.value = clusterList.map((cluster: any) => ({
        id: cluster.id || cluster.name,
        name: cluster.name,
        alias: cluster.alias,
        describe: cluster.describe,
        apiServer: cluster.endpoint || '-',
        status: cluster.status,
        version: cluster.version,
        nodeCount: cluster.nodeCount || 0,
        podCount: cluster.podCount || 0,
        namespaceCount: cluster.namespaceCount || 0
      }))
      total.value = totalCount || 0
    }
  } catch (error) {
    console.error('Failed to load clusters:', error)
    ElMessage.error('加载集群列表失败')
    clusters.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  searchParams.value.page = 1
  loadClusters()
}

// 重置
const handleReset = () => {
  searchParams.value = {
    page: 1,
    pageSize: 12,
    keyword: '',
    status: ''
  }
  loadClusters()
}

// 分页大小改变
const handleSizeChange = (size: number) => {
  searchParams.value.pageSize = size
  searchParams.value.page = 1
  loadClusters()
}

// 页码改变
const handlePageChange = (page: number) => {
  searchParams.value.page = page
  loadClusters()
}

// 获取状态类型
const getStatusType = (status: string) => {
  const statusMap: Record<string, any> = {
    'Active': 'success',
    'Running': 'success',
    'Error': 'danger',
    'Failed': 'danger',
    'Pending': 'warning',
    'Unknown': 'info'
  }
  return statusMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    'Active': '活跃',
    'Running': '运行中',
    'Error': '异常',
    'Failed': '失败',
    'Pending': '等待中',
    'Unknown': '未知'
  }
  return statusMap[status] || status
}

// 进入集群
const enterCluster = (cluster: any) => {
  clusterStore.setCurrentCluster({
    id: cluster.id,
    name: cluster.name,
    apiServer: cluster.apiServer,
    status: cluster.status
  })
  router.push(`/cluster/${cluster.id}/overview`)
}

// 处理命令
const handleCommand = (command: string, cluster: any) => {
  if (command === 'edit') {
    // 加载集群详情并打开编辑对话框
    editClusterForm.value = {
      name: cluster.name,
      alias: cluster.alias || '',
      describe: cluster.describe || '',
      config: ''
    }
    showEditDialog.value = true
  } else if (command === 'delete') {
    ElMessageBox.confirm(`确定要删除集群 "${cluster.alias || cluster.name}" 吗？`, '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    }).then(async () => {
      try {
        await clusterApi.deleteCluster(cluster.name)
        ElMessage.success('删除成功')
        loadClusters()
      } catch (error) {
        console.error('Failed to delete cluster:', error)
        ElMessage.error('删除失败')
      }
    }).catch(() => {})
  }
}

// 添加集群
const handleAddCluster = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    submitting.value = true
    try {
      await clusterApi.createCluster({
        name: clusterForm.value.name,
        alias: clusterForm.value.alias,
        describe: clusterForm.value.describe,
        config: clusterForm.value.config
      })
      ElMessage.success('添加成功')
      showAddDialog.value = false
      clusterForm.value = { name: '', alias: '', describe: '', config: '' }
      loadClusters()
    } catch (error) {
      console.error('Failed to add cluster:', error)
      ElMessage.error('添加失败')
    } finally {
      submitting.value = false
    }
  })
}

// 编辑集群
const handleEditCluster = async () => {
  if (!editFormRef.value) return
  await editFormRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    submitting.value = true
    try {
      await clusterApi.updateCluster(editClusterForm.value.name, {
        name: editClusterForm.value.name,
        alias: editClusterForm.value.alias,
        describe: editClusterForm.value.describe,
        config: editClusterForm.value.config
      })
      ElMessage.success('更新成功')
      showEditDialog.value = false
      loadClusters()
    } catch (error) {
      console.error('Failed to update cluster:', error)
      ElMessage.error('更新失败')
    } finally {
      submitting.value = false
    }
  })
}

onMounted(() => {
  loadClusters()
})
</script>

<style scoped>
.cluster-list-page {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 8px 0;
}

.page-desc {
  font-size: 14px;
  color: #6b7280;
  margin: 0;
}

.search-bar {
  display: flex;
  align-items: center;
  margin-bottom: 24px;
  padding: 20px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.cluster-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
  min-height: 400px;
}

.cluster-card {
  background: #fff;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  border: 1px solid rgba(59, 130, 246, 0.12);
  cursor: pointer;
  transition: all 0.3s ease;
}

.cluster-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(59, 130, 246, 0.12);
}

.cluster-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.cluster-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
}

.more-btn {
  color: #9ca3af;
}

.cluster-body {
  margin-bottom: 16px;
}

.cluster-name {
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
  margin: 0 0 6px 0;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.cluster-name-sub {
  font-size: 12px;
  font-weight: 400;
  color: #9ca3af;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
}

.cluster-describe {
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 10px;
  line-height: 1.5;
}

.cluster-status {
  margin-bottom: 16px;
}

.cluster-info {
  margin-bottom: 16px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 13px;
}

.info-label {
  color: #9ca3af;
}

.info-value {
  color: #4b5563;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cluster-stats {
  display: flex;
  gap: 20px;
  padding-top: 16px;
  border-top: 1px solid #f3f4f6;
}

.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 18px;
  font-weight: 600;
  color: #3b82f6;
}

.stat-label {
  font-size: 11px;
  color: #9ca3af;
  margin-top: 2px;
}

.cluster-footer {
  display: flex;
  justify-content: flex-end;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 32px;
  padding: 20px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

/* 响应式 */
@media (max-width: 768px) {
  .cluster-list-page {
    padding: 16px;
  }

  .page-header {
    flex-direction: column;
    gap: 16px;
  }

  .search-bar {
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
  }

  .search-bar .el-input,
  .search-bar .el-select {
    width: 100% !important;
    margin-left: 0 !important;
  }

  .cluster-grid {
    grid-template-columns: 1fr;
  }
}
</style>
