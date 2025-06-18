<template>
  <div class="daemonset-list">
    <div class="page-header">
      <h1>DaemonSet管理</h1>
      <div class="header-actions">
        <el-button type="primary" @click="handleAddDaemonSet">
          <el-icon><Plus /></el-icon>
          添加DaemonSet
        </el-button>
        <el-button type="success" @click="handleAddDaemonSetByYaml">
          <el-icon><Document /></el-icon>
          YAML添加
        </el-button>
      </div>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="集群">
          <el-select 
            v-model="searchForm.clusterId" 
            placeholder="请选择集群"
            style="width: 200px"
            clearable
            @change="handleClusterChange"
          >
            <el-option
              v-for="cluster in clusterList"
              :key="cluster.id"
              :label="cluster.name"
              :value="cluster.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="命名空间">
          <el-select 
            v-model="searchForm.namespace" 
            placeholder="请选择命名空间"
            style="width: 200px"
            clearable
            @change="handleNamespaceChange"
          >
            <el-option
              v-for="namespace in namespaceList"
              :key="namespace.name"
              :label="namespace.name"
              :value="namespace.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="名称">
          <el-input 
            v-model="searchForm.name" 
            placeholder="输入DaemonSet名称" 
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
        :data="daemonSetList" 
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
        <el-table-column prop="namespace" label="命名空间" min-width="100" header-align="center" align="center" />
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

    <!-- 编辑DaemonSet对话框 -->
    <el-dialog 
      v-model="editDialogVisible" 
      title="编辑DaemonSet" 
      width="600px"
      :before-close="() => editDialogVisible = false"
      destroy-on-close
    >
      <el-form 
        :model="editForm" 
        label-width="100px"
        class="daemonset-form"
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
            placeholder="DaemonSet名称" 
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
      title="YAML添加DaemonSet" 
      width="80%"
      :before-close="() => yamlDialogVisible = false"
      destroy-on-close
    >
      <el-form 
        :model="yamlForm" 
        label-width="100px"
      >
        <el-form-item label="集群">
          <el-select 
            v-model="yamlForm.clusterId" 
            placeholder="请选择集群" 
            style="width: 300px;"
          >
            <el-option 
              v-for="cluster in clusterList" 
              :key="cluster.id" 
              :label="cluster.name" 
              :value="cluster.id" 
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="YAML配置">
          <el-input 
            v-model="yamlForm.yaml" 
            type="textarea" 
            :rows="20" 
            placeholder="请输入DaemonSet的YAML配置" 
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
      width="80%"
      :before-close="() => viewYamlDialogVisible = false"
      destroy-on-close
    >
      <el-form 
        :model="viewYamlForm" 
        label-width="100px"
      >
        <el-form-item label="集群">
          <el-input 
            v-model="viewYamlForm.clusterName" 
            placeholder="集群名称" 
            disabled
            style="width: 300px;"
          />
        </el-form-item>
        
        <el-form-item label="命名空间">
          <el-input 
            v-model="viewYamlForm.namespace" 
            placeholder="命名空间" 
            disabled
            style="width: 300px;"
          />
        </el-form-item>
        
        <el-form-item label="YAML配置">
          <el-input 
            v-model="viewYamlForm.yaml" 
            type="textarea" 
            :rows="20" 
            style="font-family: 'Courier New', monospace;"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="viewYamlDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleApplyEditYaml">应用修改</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, ArrowDown, Plus, Document } from '@element-plus/icons-vue'
import { daemonSetApi, type DaemonSetVO, type DaemonSetQueryDTO } from '@/api/daemonset'
import { clusterApi, type ClusterVO } from '@/api/cluster'
import { namespaceApi, type NamespaceVO } from '@/api/namespace'

// 响应式数据
const loading = ref(false)
const daemonSetList = ref<DaemonSetVO[]>([])
const clusterList = ref<ClusterVO[]>([])
const namespaceList = ref<NamespaceVO[]>([])

// 对话框状态
const editDialogVisible = ref(false)
const yamlDialogVisible = ref(false)
const viewYamlDialogVisible = ref(false)
const currentDaemonSet = ref<DaemonSetVO | null>(null)

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
const searchForm = reactive<DaemonSetQueryDTO>({
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

// 获取集群列表
const fetchClusterList = async () => {
  try {
    const response = await clusterApi.getClusters({ page: 1, pageSize: 100 })
    console.log('集群列表API响应:', response)
    
    if (response.code === 200 && response.data && response.data.items) {
      clusterList.value = response.data.items
      
      // 如果没有选中集群且有集群数据，自动选择第一个
      if (!searchForm.clusterId && clusterList.value.length > 0) {
        searchForm.clusterId = clusterList.value[0].id
        fetchNamespaceList()
      }
    }
  } catch (error: any) {
    console.error('获取集群列表失败:', error)
    ElMessage.error('获取集群列表失败')
  }
}

// 获取命名空间列表
const fetchNamespaceList = async () => {
  if (!searchForm.clusterId) {
    namespaceList.value = []
    return
  }
  
  try {
    const params = {
      clusterId: searchForm.clusterId,
      keyword: '',
      page: 1,
      pageSize: 1000
    }
    
    const response = await namespaceApi.getList(params)
    console.log('命名空间列表API响应:', response)
    
    if (response.code === 200 && response.data) {
      namespaceList.value = response.data.items || []
      
      // 如果没有选中命名空间且有命名空间数据，自动选择第一个
      if (!searchForm.namespace && namespaceList.value.length > 0) {
        searchForm.namespace = namespaceList.value[0].name
      }
    }
  } catch (error: any) {
    console.error('获取命名空间列表失败:', error)
    ElMessage.error('获取命名空间列表失败')
  }
}

// 获取DaemonSet列表
const fetchDaemonSetList = async () => {
  if (!searchForm.clusterId) {
    ElMessage.warning('请先选择集群')
    return
  }
  
  loading.value = true
  try {
    const params = {
      ...searchForm,
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    const response = await daemonSetApi.getDaemonSets(params)
    if (response.code === 200) {
      daemonSetList.value = response.data.items || []
      pagination.total = response.data.total || 0
    }
  } catch (error) {
    console.error('获取DaemonSet列表失败:', error)
    ElMessage.error('获取DaemonSet列表失败')
  } finally {
    loading.value = false
  }
}

// 集群变化处理
const handleClusterChange = async () => {
  searchForm.namespace = ''
  namespaceList.value = []
  if (searchForm.clusterId) {
    await fetchNamespaceList()
  }
}

// 命名空间变化处理
const handleNamespaceChange = () => {
  fetchDaemonSetList()
}

// 搜索处理
const handleSearch = () => {
  pagination.page = 1
  fetchDaemonSetList()
}

// 重置处理
const handleReset = () => {
  searchForm.namespace = ''
  searchForm.name = ''
  pagination.page = 1
  fetchDaemonSetList()
}

// 分页大小变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchDaemonSetList()
}

// 当前页变化
const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchDaemonSetList()
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
const handleViewDetail = (row: DaemonSetVO) => {
  ElMessage.info('查看详情功能开发中')
}

// 格式化时间
const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 查看详情
const handleDetail = (row: DaemonSetVO) => {
  ElMessage.info('详情功能暂未开放')
}

// 编辑DaemonSet
const handleEdit = (row: DaemonSetVO) => {
  currentDaemonSet.value = row
  editForm.clusterName = clusterList.value.find(c => c.id === searchForm.clusterId)?.name || ''
  editForm.namespace = row.namespace
  editForm.name = row.name
  editForm.alias = row.alias || ''
  editForm.describe = row.describe || ''
  editDialogVisible.value = true
}

// 保存编辑
const handleSaveEdit = async () => {
  if (!currentDaemonSet.value) return
  
  try {
    const params = {
      clusterId: searchForm.clusterId,
      namespace: currentDaemonSet.value.namespace,
      name: currentDaemonSet.value.name,
      alias: editForm.alias,
      describe: editForm.describe
    }
    
    const response = await daemonSetApi.updateDaemonSet(params)
    if (response.code === 200) {
      ElMessage.success('修改成功')
      editDialogVisible.value = false
      fetchDaemonSetList()
    } else {
      ElMessage.error(response.message || '修改失败')
    }
  } catch (error: any) {
    console.error('修改DaemonSet失败:', error)
    ElMessage.error('修改失败')
  }
}

// 添加DaemonSet
const handleAddDaemonSet = () => {
  ElMessage.info('普通添加功能暂未实现，请使用YAML添加')
}

// YAML添加DaemonSet
const handleAddDaemonSetByYaml = () => {
  yamlForm.clusterId = searchForm.clusterId
  yamlForm.yaml = ''
  yamlDialogVisible.value = true
}

// 应用YAML
const handleApplyYaml = async () => {
  if (!yamlForm.clusterId) {
    ElMessage.warning('请选择集群')
    return
  }
  
  if (!yamlForm.yaml.trim()) {
    ElMessage.warning('请输入YAML配置')
    return
  }
  
  try {
    const params = {
      clusterId: yamlForm.clusterId,
      namespace: searchForm.namespace || 'default',
      yaml: yamlForm.yaml
    }
    
    const response = await daemonSetApi.applyDaemonSet(params)
    if (response.code === 200) {
      ElMessage.success('应用成功')
      yamlDialogVisible.value = false
      fetchDaemonSetList()
    } else {
      ElMessage.error(response.message || '应用失败')
    }
  } catch (error: any) {
    console.error('应用YAML失败:', error)
    ElMessage.error('应用失败')
  }
}

// 更多操作处理
const handleMoreAction = (command: string, row: DaemonSetVO) => {
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
const handleViewYaml = (row: DaemonSetVO) => {
  if (!row.yaml) {
    ElMessage.error('YAML数据不可用')
    return
  }
  
  viewYamlForm.clusterName = clusterList.value.find(c => c.id === searchForm.clusterId)?.name || ''
  viewYamlForm.namespace = row.namespace
  viewYamlForm.yaml = row.yaml
  currentDaemonSet.value = row
  viewYamlDialogVisible.value = true
}

// 应用编辑的YAML
const handleApplyEditYaml = async () => {
  if (!currentDaemonSet.value) return
  
  try {
    const params = {
      clusterId: searchForm.clusterId,
      namespace: currentDaemonSet.value.namespace,
      yaml: viewYamlForm.yaml
    }
    
    const response = await daemonSetApi.applyDaemonSet(params)
    if (response.code === 200) {
      ElMessage.success('应用成功')
      viewYamlDialogVisible.value = false
      fetchDaemonSetList()
    } else {
      ElMessage.error(response.message || '应用失败')
    }
  } catch (error: any) {
    console.error('应用YAML失败:', error)
    ElMessage.error('应用失败')
  }
}

// 删除DaemonSet
const handleDelete = async (row: DaemonSetVO) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除DaemonSet "${row.name}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    const params = {
      clusterId: searchForm.clusterId,
      namespace: row.namespace,
      name: row.name
    }
    
    const response = await daemonSetApi.deleteDaemonSet(params)
    if (response.code === 200) {
      ElMessage.success('删除成功')
      fetchDaemonSetList()
    } else {
      ElMessage.error(response.message || '删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除DaemonSet失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 重建DaemonSet
const handleRestart = async (row: DaemonSetVO) => {
  try {
    await ElMessageBox.confirm(
      `确定要重建DaemonSet "${row.name}" 吗？`,
      '确认重建',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    const params = {
      clusterId: searchForm.clusterId,
      namespace: row.namespace,
      name: row.name
    }
    
    const response = await daemonSetApi.rolloutDaemonSet(params)
    if (response.code === 200) {
      ElMessage.success('重建成功')
      fetchDaemonSetList()
    } else {
      ElMessage.error(response.message || '重建失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('重建DaemonSet失败:', error)
      ElMessage.error('重建失败')
    }
  }
}

// 页面加载时获取数据
onMounted(() => {
  fetchClusterList().then(() => {
    if (searchForm.clusterId) {
      fetchNamespaceList().then(() => {
        // 如果有命名空间参数或者自动选择了第一个命名空间，则加载DaemonSet列表
        if (searchForm.namespace) {
          fetchDaemonSetList()
        }
      })
    }
  })
})
</script>

<style scoped>
.daemonset-list {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
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
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
  padding: 0 20px 20px;
}

.daemonset-form {
  padding: 0 20px;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-table) {
  border-radius: 8px;
}

:deep(.el-table__header) {
  background-color: #f8f9fa;
}

:deep(.el-table th) {
  background-color: #f8f9fa;
  color: #606266;
  font-weight: 600;
}

:deep(.el-table td) {
  border-bottom: 1px solid #ebeef5;
}

:deep(.el-table tr:hover > td) {
  background-color: #f5f7fa;
}

:deep(.el-card__body) {
  padding: 20px;
}

:deep(.el-form--inline .el-form-item) {
  margin-right: 20px;
}

:deep(.el-button + .el-button) {
  margin-left: 0;
}
</style>