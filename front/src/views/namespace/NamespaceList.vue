<template>
  <div class="namespace-list list-page">
    <PageHeader title="命名空间列表" subtitle="管理集群中的命名空间" icon="ns">
      <el-button type="primary" @click="handleAddNamespace">
        <el-icon><Plus /></el-icon>
        添加命名空间
      </el-button>
      <el-button type="success" @click="handleAddNamespaceByYaml">
        <el-icon><Document /></el-icon>
        YAML添加
      </el-button>
    </PageHeader>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="命名空间">
          <el-input 
            v-model="searchForm.keyword" 
            placeholder="请输入命名空间名称" 
            clearable
            @keyup.enter="handleSearch"
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
        :data="namespaceList"
        style="width: 100%"
        @selection-change="handleSelectionChange"
        :header-cell-class-name="tableHeaderClass"
        :cell-class-name="tableCellClass"
      >
        <el-table-column type="selection" width="55" header-align="center" align="center" />
        <el-table-column prop="name" label="命名空间名称" min-width="180" header-align="center" align="center">
          <template #default="{ row }">
            <el-link type="primary" @click="handleViewDetail(row)">
              {{ row.name }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="alias" label="别名" min-width="120" :show-overflow-tooltip="true" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.alias">{{ row.alias }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="describe" label="描述" min-width="200" :show-overflow-tooltip="true" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.describe">{{ row.describe }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="active" label="状态" min-width="100" header-align="center" align="center">
          <template #default="{ row }">
            <el-tag 
              :type="getStatusType(row.active)"
              size="small"
            >
              {{ getStatusText(row.active) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="creationTimestamp" label="创建时间" min-width="160" :show-overflow-tooltip="true" header-align="center" align="center">
          <template #default="{ row }">
            <span v-if="row.creationTimestamp">{{ formatDate(row.creationTimestamp) }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right" header-align="center" align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" class="btn-main" @click="handleViewDetail(row)">
                详情
              </el-button>
              <el-button size="small" class="btn-edit" plain @click="handleEdit(row)">
                编辑
              </el-button>
              <el-dropdown trigger="click" @command="(command: string) => handleMoreAction(command, row)">
                <el-button size="small" text circle class="more-btn">
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="yaml">查看YAML</el-dropdown-item>
                    <el-dropdown-item command="pods">Pod管理</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-button size="small" type="danger" plain class="btn-danger" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页器 -->
      <ListPagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>

    <!-- 新增/编辑命名空间对话框 -->
    <NamespaceFormDialog
      v-model="dialogVisible"
      :cluster-id="clusterId"
      :namespace="editingNamespace"
      @success="fetchNamespaceList"
    />

    <!-- YAML添加命名空间对话框 -->
    <YamlDialog
      v-model="yamlDialogVisible"
      title="YAML添加命名空间"
      subtitle="通过 YAML 快速创建"
      v-model:content="yamlForm.yaml"
      :readonly="false"
      editor-title="命名空间 YAML"
      filename="namespace.yaml"
      :submitting="yamlSubmitLoading"
      @confirm="handleApplyYaml"
    />

    <!-- 查看/编辑YAML对话框 -->
    <YamlDialog
      v-model="viewYamlDialogVisible"
      title="查看/编辑YAML"
      subtitle="命名空间配置"
      :info="[{ label: '集群', value: viewYamlForm.clusterName }, { label: '命名空间', value: viewYamlForm.namespace }, { label: '名称', value: currentNamespace?.name }, { label: '类型', value: 'Namespace' }]"
      v-model:content="viewYamlForm.yaml"
      :readonly="false"
      :editor-title="`${currentNamespace?.name} - Namespace YAML`"
      :filename="`${currentNamespace?.name}-namespace.yaml`"
      @change="handleYamlChange"
      :submitting="applyLoading"
      confirm-text="应用修改"
      @confirm="handleApplyEditYaml"
    />

    <!-- 删除确认对话框 -->
    <DeleteConfirmDialog
      v-model="deleteDialogVisible"
      :item-name="currentDeleteNamespace?.name || ''"
      message="确定要删除命名空间吗？"
      :loading="deleteLoading"
      @confirm="confirmDeleteNamespace"
      @cancel="cancelDeleteNamespace"
    />
  </div>
</template>

<script setup lang="ts">
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Plus,
  Search,
  Refresh,
  MoreFilled,
  Document
} from '@element-plus/icons-vue'
import { 
  namespaceApi, 
  type NamespaceVO, 
  type NamespacePageQueryDTO, 
  type NamespaceApplyDTO,
  type NamespaceBaseDTO
} from '@/api/namespace'
import { clusterApi, type ClusterVO } from '@/api/cluster'
import { useClusterStore } from '@/stores/cluster'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import NamespaceFormDialog from '@/components/NamespaceFormDialog.vue'
import '@/assets/styles/config-editor.css'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const applyLoading = ref(false)
const yamlSubmitLoading = ref(false)

// 集群上下文
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)

// 搜索表单
const searchForm = reactive({
  keyword: ''
})

// 命名空间列表
const namespaceList = ref<NamespaceVO[]>([])
const selectedNamespaces = ref<NamespaceVO[]>([])

// 集群列表（用于对话框）
const clusterList = ref<ClusterVO[]>([])

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 对话框
const dialogVisible = ref(false)
const yamlDialogVisible = ref(false)
// 为空时是添加, 有值时是编辑
const editingNamespace = ref<{ name: string; alias?: string; describe?: string } | null>(null)

// YAML表单
const yamlForm = reactive({
  yaml: ''
})

// 查看YAML对话框
const viewYamlDialogVisible = ref(false)
const currentNamespace = ref<NamespaceVO | null>(null)
const viewYamlForm = reactive({
  clusterName: '',
  namespace: '',
  yaml: ''
})

// 获取命名空间列表
const fetchNamespaceList = async () => {
  if (!clusterId.value) {
    ElMessage.warning('请先选择集群')
    namespaceList.value = []
    pagination.total = 0
    return
  }
  
  try {
    loading.value = true
    const params: NamespacePageQueryDTO = {
      clusterId: clusterId.value,
      keyword: searchForm.keyword,
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    
    const response = await namespaceApi.getList(params)
    
    if (response.code === 200 && response.data) {
      namespaceList.value = response.data.items || []
      pagination.total = response.data.total || 0
    }
  } catch (error: any) {
    console.error('获取命名空间列表失败:', error)
    ElMessage.error('获取命名空间列表失败')
  } finally {
    loading.value = false
  }
}

// 获取集群列表（用于对话框）
const fetchClusterList = async () => {
  try {
    const response = await clusterApi.getClusters({ page: 1, pageSize: 10000 })
    
    if (response.code === 200 && response.data) {
      clusterList.value = response.data.items || []
    }
  } catch (error: any) {
    console.error('获取集群列表失败:', error)
    ElMessage.error('获取集群列表失败')
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchNamespaceList()
}

// 重置
const handleReset = () => {
  searchForm.keyword = ''
  pagination.page = 1
  fetchNamespaceList()
}

// 分页变化
const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchNamespaceList()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchNamespaceList()
}

// 选择变化
const handleSelectionChange = (selection: NamespaceVO[]) => {
  selectedNamespaces.value = selection
}

// 添加命名空间
const handleAddNamespace = () => {
  editingNamespace.value = null
  dialogVisible.value = true
}

// YAML方式添加命名空间
const handleAddNamespaceByYaml = () => {
  Object.assign(yamlForm, {
    yaml: `apiVersion: v1
kind: Namespace
metadata:
  name: my-namespace
  labels:
    name: my-namespace`
  })
  yamlDialogVisible.value = true
}

// 应用YAML
const handleApplyYaml = async () => {
  if (!yamlForm.yaml.trim()) {
    ElMessage.warning('YAML内容不能为空')
    return
  }
  
  try {
    yamlSubmitLoading.value = true
    
    const data: NamespaceApplyDTO = {
      clusterId: clusterId.value,
      yaml: yamlForm.yaml
    }
    
    await namespaceApi.apply(data)
    ElMessage.success('应用成功')
    yamlDialogVisible.value = false
    fetchNamespaceList()
  } catch (error: any) {
    ElMessage.error('应用失败')
  } finally {
    yamlSubmitLoading.value = false
  }
}

// 编辑命名空间
const handleEdit = (row: NamespaceVO) => {
  editingNamespace.value = { name: row.name, alias: row.alias, describe: row.describe }
  dialogVisible.value = true
}

// 查看详情
const handleViewDetail = (row: NamespaceVO) => {
  router.push({
    path: `/cluster/${clusterId.value}/namespace/detail`,
    query: {
      clusterId: clusterId.value,
      clusterName: clusterStore.currentClusterName,
      namespace: row.name
    }
  })
}

// 查看YAML
const handleViewYaml = async (row: NamespaceVO) => {
  try {
    currentNamespace.value = row
    
    // 获取当前集群名称
    viewYamlForm.clusterName = clusterStore.currentClusterName
    viewYamlForm.namespace = row.name
    
    const params: NamespaceBaseDTO = {
      clusterId: clusterId.value,
      namespace: row.name
    }
    const response = await namespaceApi.getInfo(params)
    viewYamlForm.yaml = response.data.yaml || '# YAML内容加载中...'
    viewYamlDialogVisible.value = true
  } catch (error: any) {
    ElMessage.error('获取YAML配置失败')
  }
}

// 更多操作
const handleMoreAction = (command: string, row: NamespaceVO) => {
  switch (command) {
    case 'yaml':
      handleViewYaml(row)
      break
    case 'pods':
      handleViewPods(row)
      break
    case 'delete':
      handleDelete(row)
      break
  }
}

// 查看Pod
const handleViewPods = (row: NamespaceVO) => {
  router.push({
    path: `/cluster/${clusterId.value}/pod`,
    query: {
      namespace: row.name
    }
  })
}

// 删除命名空间相关状态
const deleteDialogVisible = ref(false)
const currentDeleteNamespace = ref<NamespaceVO | null>(null)
const deleteLoading = ref(false)

// 删除命名空间
const handleDelete = (row: NamespaceVO) => {
  currentDeleteNamespace.value = row
  deleteDialogVisible.value = true
}

// 确认删除命名空间
const confirmDeleteNamespace = async () => {
  if (!currentDeleteNamespace.value) return
  
  deleteLoading.value = true
  try {
    const params: NamespaceBaseDTO = {
      clusterId: clusterId.value,
      namespace: currentDeleteNamespace.value.name
    }
    await namespaceApi.delete(params)
    ElMessage.success('删除成功')
    deleteDialogVisible.value = false
    fetchNamespaceList()
  } catch (error: any) {
    ElMessage.error('删除失败')
  } finally {
    deleteLoading.value = false
  }
}

// 取消删除
const cancelDeleteNamespace = () => {
  deleteDialogVisible.value = false
  currentDeleteNamespace.value = null
}

// 应用YAML修改
const handleApplyEditYaml = async () => {
  if (!viewYamlForm.yaml.trim()) {
    ElMessage.warning('YAML内容不能为空')
    return
  }
  
  try {
    applyLoading.value = true
    
    const data: NamespaceApplyDTO = {
      clusterId: clusterId.value,
      yaml: viewYamlForm.yaml
    }
    
    await namespaceApi.apply(data)
    ElMessage.success('应用成功')
    viewYamlDialogVisible.value = false
    fetchNamespaceList()
  } catch (error: any) {
    ElMessage.error('应用失败')
  } finally {
    applyLoading.value = false
  }
}

// 处理YAML内容变化
const handleYamlChange = (newYaml: string) => {
  viewYamlForm.yaml = newYaml
}

// 状态相关
const getStatusType = (status: string) => {
  switch (status) {
    case 'Active':
      return 'success'
    case 'Terminating':
      return 'warning'
    default:
      return 'danger'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'Active':
      return '活跃'
    case 'Terminating':
      return '终止中'
    default:
      return status
  }
}

// 格式化日期
const formatDate = (dateString?: string) => {
  if (!dateString) return '-'
  try {
    const date = new Date(dateString)
    if (isNaN(date.getTime())) return '-'
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch (error) {
    return '-'
  }
}

// 页面加载时获取数据
onMounted(() => {
  // 获取集群列表用于对话框（保留用于查看YAML时显示集群名称）
  fetchClusterList()
  
  // 如果有集群上下文，加载命名空间列表
  if (clusterId.value) {
    fetchNamespaceList()
  }
})
</script>

<style scoped>
.namespace-list {
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

.namespace-form {
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
  margin-left: 8px;
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

/* YAML添加对话框样式 */
.yaml-form-item {
  margin-bottom: 0;
}

.yaml-form-item :deep(.el-form-item__content) {
  width: 100%;
  max-width: none;
}

.yaml-form-item :deep(.yaml-editor) {
  width: 100%;
}

.yaml-form-item :deep(.yaml-editor .editor-container) {
  width: 100%;
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

</style>
