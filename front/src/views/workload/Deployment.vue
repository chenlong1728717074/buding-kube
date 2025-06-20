<template>
  <div class="deployment-list">
    <div class="page-header">
      <h1>Deployment管理</h1>
      <div class="header-actions">
        <el-button type="primary" @click="handleAddDeployment">
          <el-icon><Plus /></el-icon>
          添加Deployment
        </el-button>
        <el-button type="success" @click="handleAddDeploymentByYaml">
          <el-icon><Document /></el-icon>
          YAML添加
        </el-button>
      </div>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="集群">
          <InfiniteSelect
            v-model="searchForm.clusterId"
            :fetch-data="clusterFetcher"
            v-bind="clusterSelectConfig"
            style="width: 200px"
            @change="handleClusterChange"
          />
        </el-form-item>
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
          <el-input 
            v-model="searchForm.name" 
            placeholder="输入Deployment名称" 
            clearable
            style="width: 200px"
          />
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
        :data="deploymentList" 
        stripe
        style="width: 100%"
      >
        <el-table-column prop="name" label="名称" min-width="120" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="handleViewDetail(row)" :title="row.describe || ''">
              {{ row.name }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" min-width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="handleNamespaceDetail(row)">
              {{ row.namespace }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="alias" label="别名" min-width="80" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.alias">{{ row.alias }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" min-width="80" align="center" header-align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="副本数" min-width="80" header-align="center" align="center">
          <template #default="{ row }">
            <span>{{ row.readyReplicas }}/{{ row.replicas }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" min-width="140" header-align="center" align="center">
          <template #default="{ row }">
            {{ formatTime(row.createTime) }}
          </template>
        </el-table-column>
        <el-table-column prop="updateTime" label="更新时间" min-width="140" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.updateTime">{{ formatTime(row.updateTime) }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right" header-align="center">
          <template #default="{ row }">
            <div style="display: flex; gap: 6px; align-items: center; justify-content: center; flex-wrap: nowrap;">
              <el-button size="small" @click="handleDetail(row)">
                详情
              </el-button>
              <el-button size="small" @click="handleEdit(row)">
                编辑
              </el-button>
              <el-dropdown @command="(command) => handleMoreAction(command, row)">
                <el-button size="small">
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="yaml">查看YAML</el-dropdown-item>
                    <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                    <el-dropdown-item command="restart">重建</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页器 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 编辑Deployment对话框 -->
    <el-dialog 
      v-model="editDialogVisible" 
      title="编辑Deployment" 
      width="600px"
      :before-close="() => editDialogVisible = false"
      destroy-on-close
    >
      <el-form 
        :model="editForm" 
        label-width="100px"
        class="deployment-form"
      >
        <el-form-item label="集群">
          <el-input 
            v-model="editForm.clusterName" 
            placeholder="集群名称" 
            disabled
            style="width: 100%;"
          />
        </el-form-item>
        
        <el-form-item label="命名空间">
          <el-input 
            v-model="editForm.namespace" 
            placeholder="命名空间" 
            disabled
            style="width: 100%;"
          />
        </el-form-item>
        
        <el-form-item label="名称">
          <el-input 
            v-model="editForm.name" 
            placeholder="Deployment名称" 
            disabled
            style="width: 100%;"
          />
        </el-form-item>
        
        <el-form-item label="别名">
          <el-input 
            v-model="editForm.alias" 
            placeholder="请输入别名" 
            style="width: 100%;"
          />
        </el-form-item>
        
        <el-form-item label="描述">
          <el-input 
            v-model="editForm.describe" 
            type="textarea" 
            :rows="3" 
            placeholder="请输入描述" 
            style="width: 100%;"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveEdit">保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- YAML添加对话框 -->
    <el-dialog 
      v-model="yamlDialogVisible" 
      title="YAML添加Deployment" 
      width="80%"
      :before-close="() => yamlDialogVisible = false"
      destroy-on-close
    >
      <el-form 
        :model="yamlForm" 
        label-width="100px"
      >
        <el-form-item label="集群">
          <InfiniteSelect
            v-model="yamlForm.clusterId"
            :fetch-data="clusterFetcher"
            v-bind="clusterSelectConfig"
            style="width: 300px"
          />
        </el-form-item>
        
        <el-form-item label="YAML配置">
          <el-input 
            v-model="yamlForm.yaml" 
            type="textarea" 
            :rows="20" 
            placeholder="请输入Deployment的YAML配置" 
            style="font-family: 'Courier New', monospace;"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="yamlDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleApplyYaml">应用</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 查看/编辑YAML对话框 -->
    <el-dialog 
      v-model="viewYamlDialogVisible" 
      title="查看/编辑YAML" 
      width="90%"
      :before-close="() => viewYamlDialogVisible = false"
      destroy-on-close
    >
      <div class="yaml-dialog-content">
        <div class="yaml-info">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="集群">{{ viewYamlForm.clusterName }}</el-descriptions-item>
            <el-descriptions-item label="命名空间">{{ viewYamlForm.namespace }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ currentDeployment?.name }}</el-descriptions-item>
            <el-descriptions-item label="类型">Deployment</el-descriptions-item>
          </el-descriptions>
        </div>
        
        <div class="yaml-editor-wrapper">
          <YamlEditor
            v-model="viewYamlForm.yaml"
            :title="`${currentDeployment?.name} - Deployment YAML`"
            :readonly="false"
            height="500px"
            :filename="`${currentDeployment?.name}-deployment.yaml`"
            @change="handleYamlChange"
          />
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="viewYamlDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleApplyEditYaml" :loading="applyLoading">应用修改</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, ArrowDown, Plus, Document } from '@element-plus/icons-vue'
import { deploymentApi, type DeploymentVO, type DeploymentQueryDTO } from '@/api/deployment'
import { clusterApi, type ClusterVO } from '@/api/cluster'
import { namespaceApi, type NamespaceVO } from '@/api/namespace'
import InfiniteSelect from '@/components/InfiniteSelect.vue'
import YamlEditor from '@/components/YamlEditor.vue'
import { useClusterFetcher, useNamespaceFetcher, clusterSelectConfig, namespaceSelectConfig } from '@/composables/useInfiniteSelect'

// 路由
const router = useRouter()

// 响应式数据
const loading = ref(false)
const deploymentList = ref<DeploymentVO[]>([])

// 渐进式加载的数据获取函数
const clusterFetcher = useClusterFetcher()
const namespaceFetcher = computed(() => useNamespaceFetcher(searchForm.clusterId))

// 对话框状态
const editDialogVisible = ref(false)
const yamlDialogVisible = ref(false)
const viewYamlDialogVisible = ref(false)
const currentDeployment = ref<DeploymentVO | null>(null)
const applyLoading = ref(false)

// 编辑表单
const editForm = reactive({
  clusterName: '',
  namespace: '',
  name: '',
  alias: '',
  describe: ''
})

// YAML表单
const yamlForm = reactive({
  clusterId: '',
  yaml: ''
})

// 查看YAML表单
const viewYamlForm = reactive({
  clusterName: '',
  namespace: '',
  yaml: ''
})

// 搜索表单
const searchForm = reactive<DeploymentQueryDTO>({
  clusterId: '',
  namespace: '',
  name: '',
  page: 1,
  pageSize: 10
})

// 分页信息
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})



// 获取Deployment列表
const fetchDeploymentList = async () => {
  if (!searchForm.clusterId) {
    // 没有集群时清空列表
    deploymentList.value = []
    pagination.total = 0
    return
  }
  
  loading.value = true
  try {
    const params = {
      ...searchForm,
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    const response = await deploymentApi.getDeployments(params)
    if (response.code === 200) {
      deploymentList.value = response.data.items || []
      pagination.total = response.data.total || 0
    }
  } catch (error) {
    console.error('获取Deployment列表失败:', error)
    ElMessage.error('获取Deployment列表失败')
  } finally {
    loading.value = false
  }
}



// 搜索处理
const handleSearch = () => {
  pagination.page = 1
  fetchDeploymentList()
}

// 重置处理
const handleReset = () => {
  searchForm.namespace = ''
  searchForm.name = ''
  pagination.page = 1
  fetchDeploymentList()
}

// 分页大小变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchDeploymentList()
}

// 当前页变化
const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchDeploymentList()
}

// 获取状态类型
const getStatusType = (status: string) => {
  switch (status) {
    case 'Available':
    case 'Running':
    case '运行中':
      return 'success'  // 绿色 - 运行中
    case 'Progressing':
    case 'Pending':
    case '更新中':
      return 'warning'  // 橙色 - 更新中
    case 'ReplicaFailure':
    case 'Failed':
    case 'Paused':
    case '已停止':
      return 'danger'   // 红色 - 已停止
    case 'Unknown':
    case '未知':
    default:
      return 'info'     // 蓝色 - 未知
  }
}

// 查看详情
const handleViewDetail = (row: DeploymentVO) => {
  // TODO: 实现查看详情功能
  ElMessage.info('查看详情功能开发中')
}

// 跳转到命名空间详情
const handleNamespaceDetail = (row: DeploymentVO) => {
  router.push({
    path: '/namespace/detail',
    query: {
      clusterId: searchForm.clusterId,
      namespace: row.namespace
    }
  })
}

// 格式化时间
const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 处理集群变化
const handleClusterChange = (clusterId: string) => {
  searchForm.clusterId = clusterId
  searchForm.namespace = '' // 重置命名空间选择
  if (clusterId) {
    // 命名空间下拉框会自动重新加载数据
    // 如果有集群选择，立即加载Deployment列表
    fetchDeploymentList()
  } else {
    // 清空列表
    deploymentList.value = []
    pagination.total = 0
  }
}

// 处理命名空间变化
const handleNamespaceChange = (namespace: string) => {
  searchForm.namespace = namespace
  // 只要有集群选择，就重新查询（无论命名空间是否为空）
  if (searchForm.clusterId) {
    fetchDeploymentList()
  }
}

// 页面加载时获取数据
onMounted(async () => {
  // 自动选择第一个集群并查询数据
  try {
    // 获取第一个集群
    const clusterResponse = await clusterApi.getClusters({ page: 1, pageSize: 1 })
    if (clusterResponse.code === 200 && clusterResponse.data.items && clusterResponse.data.items.length > 0) {
      const firstCluster = clusterResponse.data.items[0]
      searchForm.clusterId = firstCluster.id
      
      // 自动加载Deployment列表
      fetchDeploymentList()
    }
  } catch (error) {
    console.error('自动选择集群失败:', error)
    // 如果自动选择失败，用户仍可以手动选择
  }
})

// 编辑Deployment
const handleEdit = async (row: DeploymentVO) => {
  currentDeployment.value = row
  // 获取集群名称
  try {
    const response = await clusterApi.getCluster(searchForm.clusterId)
    editForm.clusterName = response.data?.name || ''
  } catch (error) {
    editForm.clusterName = searchForm.clusterId
  }
  editForm.namespace = row.namespace
  editForm.name = row.name
  editForm.alias = row.alias || ''
  editForm.describe = row.describe || ''
  editDialogVisible.value = true
}

// 查看详情
const handleDetail = (row: DeploymentVO) => {
  ElMessage.info('详情功能暂未开放')
}

// 更多操作处理
const handleMoreAction = (command: string, row: DeploymentVO) => {
  switch (command) {
    case 'yaml':
      handleViewYaml(row)
      break
    case 'delete':
      handleDelete(row)
      break
    case 'restart':
      handleRestart(row)
      break
  }
}

// 查看YAML
const handleViewYaml = async (row: DeploymentVO) => {
  currentDeployment.value = row
  // 获取集群名称
  try {
    const response = await clusterApi.getCluster(searchForm.clusterId)
    viewYamlForm.clusterName = response.data?.name || ''
  } catch (error) {
    viewYamlForm.clusterName = searchForm.clusterId
  }
  viewYamlForm.namespace = row.namespace
  // TODO: 调用API获取YAML内容
  viewYamlForm.yaml = row.yaml || '# YAML内容加载中...'
  viewYamlDialogVisible.value = true
}

// 删除Deployment
const handleDelete = (row: DeploymentVO) => {
  ElMessageBox.confirm(
    `确定要删除Deployment "${row.name}" 吗？此操作不可恢复。`,
    '确认删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await deploymentApi.deleteDeployment({
        clusterId: searchForm.clusterId,
        namespace: row.namespace,
        name: row.name
      })
      ElMessage.success('删除成功')
      fetchDeploymentList()
    } catch (error) {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }).catch(() => {
    ElMessage.info('已取消删除')
  })
}

// 重建Deployment
const handleRestart = (row: DeploymentVO) => {
  ElMessageBox.confirm(
    `确定要重建Deployment "${row.name}" 吗？`,
    '确认重建',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await deploymentApi.rolloutDeployment({
        clusterId: searchForm.clusterId,
        namespace: row.namespace,
        name: row.name
      })
      ElMessage.success('重建成功')
      fetchDeploymentList()
    } catch (error) {
      console.error('重建失败:', error)
      ElMessage.error('重建失败')
    }
  }).catch(() => {
    ElMessage.info('已取消重建')
  })
}

// 添加Deployment
const handleAddDeployment = () => {
  ElMessage.info('普通添加功能暂未实现，请使用YAML添加')
}

// YAML添加Deployment
const handleAddDeploymentByYaml = () => {
  yamlForm.clusterId = searchForm.clusterId || ''
  yamlForm.yaml = ''
  yamlDialogVisible.value = true
}

// 保存编辑
const handleSaveEdit = async () => {
  if (!currentDeployment.value) return
  
  try {
    await deploymentApi.updateDeployment({
      clusterId: searchForm.clusterId,
      namespace: currentDeployment.value.namespace,
      name: currentDeployment.value.name,
      alias: editForm.alias,
      describe: editForm.describe
    })
    ElMessage.success('修改成功')
    editDialogVisible.value = false
    fetchDeploymentList()
  } catch (error) {
    console.error('修改失败:', error)
    ElMessage.error('修改失败')
  }
}

// 应用YAML
const handleApplyYaml = async () => {
  if (!yamlForm.yaml.trim()) {
    ElMessage.warning('请输入YAML内容')
    return
  }
  
  if (!yamlForm.clusterId) {
    ElMessage.warning('请选择集群')
    return
  }
  
  try {
    await deploymentApi.applyDeployment({
      clusterId: yamlForm.clusterId,
      yaml: yamlForm.yaml
    })
    ElMessage.success('应用成功')
    yamlDialogVisible.value = false
    fetchDeploymentList()
  } catch (error) {
    console.error('应用失败:', error)
    ElMessage.error('应用失败')
  }
}

// YAML内容变化处理
const handleYamlChange = (value: string) => {
  viewYamlForm.value.yaml = value
}

// 应用编辑YAML
const handleApplyEditYaml = async () => {
  if (!viewYamlForm.yaml.trim()) {
    ElMessage.warning('请输入YAML内容')
    return
  }
  
  if (!currentDeployment.value) return
  
  applyLoading.value = true
  try {
    await deploymentApi.applyDeployment({
      clusterId: searchForm.clusterId,
      namespace: currentDeployment.value.namespace,
      yaml: viewYamlForm.yaml
    })
    ElMessage.success('应用成功')
    viewYamlDialogVisible.value = false
    fetchDeploymentList()
  } catch (error) {
    console.error('应用失败:', error)
    ElMessage.error('应用失败')
  } finally {
    applyLoading.value = false
  }
}
</script>

<style scoped>
.deployment-list {
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
  margin-bottom: 20px;
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
</style>