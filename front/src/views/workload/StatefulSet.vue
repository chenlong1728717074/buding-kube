<template>
  <div class="statefulset-list">
    <div class="page-header">
      <h1>StatefulSet管理</h1>
      <div class="header-actions">
        <el-button type="primary" @click="handleAddStatefulSet">
          <el-icon><Plus /></el-icon>
          添加StatefulSet
        </el-button>
        <el-button type="success" @click="handleAddStatefulSetByYaml">
          <el-icon><Document /></el-icon>
          YAML添加
        </el-button>
      </div>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
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
            placeholder="输入StatefulSet名称"
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
        :data="statefulSetList"
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

    <!-- 编辑StatefulSet对话框 -->
    <el-dialog v-model="editDialogVisible" title="编辑StatefulSet" width="80%" :close-on-click-modal="false" class="config-dialog">
      <template #header>
        <div class="dialog-header">
          <div>
            <h3 class="dialog-title">编辑StatefulSet</h3>
            <div style="margin-top:4px;color:#6b7280;font-size:12px;">修改别名与描述</div>
          </div>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <el-form :model="editForm" label-width="100px" class="statefulset-form">
            <el-form-item label="集群">
              <el-input v-model="editForm.clusterName" placeholder="集群名称" disabled style="width: 100%;" />
            </el-form-item>

            <el-form-item label="命名空间">
              <el-input v-model="editForm.namespace" placeholder="命名空间" disabled style="width: 100%;" />
            </el-form-item>

            <el-form-item label="名称">
              <el-input v-model="editForm.name" placeholder="StatefulSet名称" disabled style="width: 100%;" />
            </el-form-item>

            <el-form-item label="别名">
              <el-input v-model="editForm.alias" placeholder="请输入别名" style="width: 100%;" />
            </el-form-item>

            <el-form-item label="描述">
              <el-input v-model="editForm.describe" type="textarea" :rows="3" placeholder="请输入描述" style="width: 100%;" />
            </el-form-item>
          </el-form>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveEdit">保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- YAML添加对话框 -->
    <el-dialog v-model="yamlDialogVisible" title="YAML添加StatefulSet" width="80%" :close-on-click-modal="false" class="config-dialog yaml-dialog">
      <template #header>
        <div class="dialog-header">
          <div>
            <h3 class="dialog-title">YAML添加StatefulSet</h3>
            <div style="margin-top:4px;color:#6b7280;font-size:12px;">通过 YAML 快速创建</div>
          </div>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <div class="yaml-editor-wrapper">
            <YamlEditor
              v-model="yamlForm.yaml"
              title="StatefulSet YAML"
              :readonly="false"
              height="100%"
              filename="statefulset.yaml"
            />
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="yamlDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleApplyYaml">应用</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 查看/编辑YAML对话框 -->
    <el-dialog v-model="viewYamlDialogVisible" title="查看/编辑YAML" width="90%" :close-on-click-modal="false" class="config-dialog yaml-dialog">
      <template #header>
        <div class="dialog-header">
          <div>
            <h3 class="dialog-title">查看/编辑YAML</h3>
            <div style="margin-top:4px;color:#6b7280;font-size:12px;">StatefulSet 配置</div>
          </div>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <div class="yaml-dialog-content">
            <div class="yaml-info">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="集群">{{ viewYamlForm.clusterName }}</el-descriptions-item>
                <el-descriptions-item label="命名空间">{{ viewYamlForm.namespace }}</el-descriptions-item>
                <el-descriptions-item label="名称">{{ currentStatefulSet?.name }}</el-descriptions-item>
                <el-descriptions-item label="类型">StatefulSet</el-descriptions-item>
              </el-descriptions>
            </div>

            <div class="yaml-view">
              <YamlEditor
                v-model="viewYamlForm.yaml"
                :title="`${currentStatefulSet?.name} - StatefulSet YAML`"
                :readonly="false"
                height="100%"
                :filename="`${currentStatefulSet?.name}-statefulset.yaml`"
              />
            </div>
          </div>
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
import { statefulSetApi, type StatefulSetVO, type StatefulSetQueryDTO } from '@/api/statefulset'
import { clusterApi, type ClusterVO } from '@/api/cluster'
import { namespaceApi, type NamespaceVO } from '@/api/namespace'
import { useClusterStore } from '@/stores/cluster'
import YamlEditor from '@/components/YamlEditor.vue'
import '@/assets/styles/config-editor.css'

// 路由
const router = useRouter()

// 集群上下文
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)

// 响应式数据
const loading = ref(false)
const statefulSetList = ref<StatefulSetVO[]>([])
const namespaceList = ref<NamespaceVO[]>([])

// 对话框状态
const editDialogVisible = ref(false)
const yamlDialogVisible = ref(false)
const viewYamlDialogVisible = ref(false)
const currentStatefulSet = ref<StatefulSetVO | null>(null)
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
  yaml: ''
})

// 查看YAML表单
const viewYamlForm = reactive({
  clusterName: '',
  namespace: '',
  yaml: ''
})

// 搜索表单
const searchForm = reactive<StatefulSetQueryDTO>({
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

// 获取命名空间列表
const fetchNamespaceList = async () => {
  if (!clusterId.value) {
    namespaceList.value = []
    return
  }

  try {
    const params = {
      clusterId: clusterId.value,
      keyword: '',
      page: 1,
      pageSize: 10000
    }

    const response = await namespaceApi.getList(params)
    console.log('命名空间列表API响应:', response)

    if (response.code === 200 && response.data) {
      namespaceList.value = response.data.items || []
    }
  } catch (error: any) {
    console.error('获取命名空间列表失败:', error)
    ElMessage.error('获取命名空间列表失败')
  }
}

// 获取StatefulSet列表
const fetchStatefulSetList = async () => {
  if (!clusterId.value) {
    ElMessage.warning('请先选择集群')
    return
  }

  loading.value = true
  try {
    const params = {
      clusterId: clusterId.value,
      ...(searchForm.namespace && { namespace: searchForm.namespace }),
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    const response = await statefulSetApi.getStatefulSets(params)
    if (response.code === 200) {
      statefulSetList.value = response.data.items || []
      pagination.total = response.data.total || 0
    }
  } catch (error) {
    console.error('获取StatefulSet列表失败:', error)
    ElMessage.error('获取StatefulSet列表失败')
  } finally {
    loading.value = false
  }
}

// 命名空间变化处理
const handleNamespaceChange = () => {
  fetchStatefulSetList()
}

// 搜索处理
const handleSearch = () => {
  pagination.page = 1
  fetchStatefulSetList()
}

// 重置处理
const handleReset = () => {
  searchForm.namespace = ''
  searchForm.name = ''
  pagination.page = 1
  fetchStatefulSetList()
}

// 分页大小变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchStatefulSetList()
}

// 当前页变化
const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchStatefulSetList()
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
const handleViewDetail = (row: StatefulSetVO) => {
  ElMessage.info('查看详情功能开发中')
}

// 跳转到命名空间详情
const handleNamespaceDetail = (row: StatefulSetVO) => {
  router.push({
    path: `/cluster/${clusterId.value}/namespace/detail`,
    query: {
      namespace: row.namespace
    }
  })
}

// 格式化时间
const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 查看详情
const handleDetail = (row: StatefulSetVO) => {
  ElMessage.info('详情功能暂未开放')
}

// 编辑StatefulSet
const handleEdit = async (row: StatefulSetVO) => {
  currentStatefulSet.value = row
  editDialogVisible.value = true
  
  // 获取集群名称
  try {
    const response = await clusterApi.getCluster(clusterId.value)
    editForm.clusterName = response.data?.name || ''
  } catch (error) {
    editForm.clusterName = clusterId.value
  }
  
  editForm.namespace = row.namespace
  editForm.name = row.name
  editForm.alias = row.alias || ''
  editForm.describe = row.describe || ''
}

// 保存编辑
const handleSaveEdit = async () => {
  if (!currentStatefulSet.value) return

  try {
    const params = {
      clusterId: clusterId.value,
      namespace: currentStatefulSet.value.namespace,
      name: currentStatefulSet.value.name,
      alias: editForm.alias,
      describe: editForm.describe
    }

    const response = await statefulSetApi.updateStatefulSet(params)
    if (response.code === 200) {
      ElMessage.success('修改成功')
      editDialogVisible.value = false
      fetchStatefulSetList()
    } else {
      ElMessage.error(response.message || '修改失败')
    }
  } catch (error: any) {
    console.error('修改StatefulSet失败:', error)
    ElMessage.error('修改失败')
  }
}

// 添加StatefulSet
const handleAddStatefulSet = () => {
  ElMessage.info('普通添加功能暂未实现，请使用YAML添加')
}

// YAML添加StatefulSet
const handleAddStatefulSetByYaml = () => {
  yamlForm.yaml = ''
  yamlDialogVisible.value = true
}

// 应用YAML
const handleApplyYaml = async () => {
  if (!clusterId.value) {
    ElMessage.warning('请先选择集群')
    return
  }

  if (!yamlForm.yaml.trim()) {
    ElMessage.warning('请输入YAML配置')
    return
  }

  try {
    const params = {
      clusterId: clusterId.value,
      namespace: searchForm.namespace || 'default',
      yaml: yamlForm.yaml
    }

    const response = await statefulSetApi.applyStatefulSet(params)
    if (response.code === 200) {
      ElMessage.success('应用成功')
      yamlDialogVisible.value = false
      fetchStatefulSetList()
    } else {
      ElMessage.error(response.message || '应用失败')
    }
  } catch (error: any) {
    console.error('应用YAML失败:', error)
    ElMessage.error('应用失败')
  }
}

// 更多操作处理
const handleMoreAction = (command: string, row: StatefulSetVO) => {
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
const handleViewYaml = async (row: StatefulSetVO) => {
  if (!row.yaml) {
    ElMessage.error('YAML数据不可用')
    return
  }

  try {
    const response = await clusterApi.getCluster(clusterId.value)
    viewYamlForm.clusterName = response.data?.name || ''
  } catch (error) {
    viewYamlForm.clusterName = clusterId.value
  }
  
  viewYamlForm.namespace = row.namespace
  viewYamlForm.yaml = row.yaml
  currentStatefulSet.value = row
  viewYamlDialogVisible.value = true
}

// 处理YAML内容变化
const handleYamlChange = (newYaml: string) => {
  viewYamlForm.yaml = newYaml
}

// 应用编辑的YAML
const handleApplyEditYaml = async () => {
  if (!currentStatefulSet.value) return

  try {
    applyLoading.value = true
    const params = {
      clusterId: clusterId.value,
      namespace: currentStatefulSet.value.namespace,
      yaml: viewYamlForm.yaml
    }

    const response = await statefulSetApi.applyStatefulSet(params)
    if (response.code === 200) {
      ElMessage.success('应用成功')
      viewYamlDialogVisible.value = false
      fetchStatefulSetList()
    } else {
      ElMessage.error(response.message || '应用失败')
    }
  } catch (error: any) {
    console.error('应用YAML失败:', error)
    ElMessage.error('应用失败')
  } finally {
    applyLoading.value = false
  }
}

// 删除StatefulSet
const handleDelete = async (row: StatefulSetVO) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除StatefulSet "${row.name}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    const params = {
      clusterId: clusterId.value,
      namespace: row.namespace,
      name: row.name
    }

    const response = await statefulSetApi.deleteStatefulSet(params)
    if (response.code === 200) {
      ElMessage.success('删除成功')
      fetchStatefulSetList()
    } else {
      ElMessage.error(response.message || '删除失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('删除StatefulSet失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 重建StatefulSet
const handleRestart = async (row: StatefulSetVO) => {
  try {
    await ElMessageBox.confirm(
      `确定要重建StatefulSet "${row.name}" 吗？`,
      '确认重建',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    const params = {
      clusterId: clusterId.value,
      namespace: row.namespace,
      name: row.name
    }

    const response = await statefulSetApi.rolloutStatefulSet(params)
    if (response.code === 200) {
      ElMessage.success('重建成功')
      fetchStatefulSetList()
    } else {
      ElMessage.error(response.message || '重建失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('重建StatefulSet失败:', error)
      ElMessage.error('重建失败')
    }
  }
}

// 页面加载时获取数据
onMounted(() => {
  if (clusterId.value) {
    fetchNamespaceList()
    fetchStatefulSetList()
  }
})
</script>

<style scoped>
.statefulset-list {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.yaml-editor-wrapper {
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
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

.statefulset-form {
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
