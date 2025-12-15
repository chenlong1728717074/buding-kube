<template>
  <div class="namespace-list">
    <div class="page-header">
      <h1>命名空间列表</h1>
      <div class="header-actions">
        <el-button type="primary" @click="handleAddNamespace">
          <el-icon><Plus /></el-icon>
          添加命名空间
        </el-button>
        <el-button type="success" @click="handleAddNamespaceByYaml">
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
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
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
        <el-table-column label="操作" width="280" fixed="right" header-align="center" align="center">
          <template #default="{ row }">
            <div style="display: flex; gap: 4px; align-items: center; flex-wrap: nowrap; justify-content: center;">
              <el-button size="small" @click="handleViewDetail(row)">
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
                    <el-dropdown-item command="pods">Pod管理</el-dropdown-item>
                    <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
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

    <!-- 新增/编辑命名空间对话框 -->
    <UnifiedDialog 
      v-model="dialogVisible" 
      :title="dialogTitle" 
      subtitle="命名空间基础信息"
      width="600px"
    >
      <el-form 
        ref="formRef" 
        :model="namespaceForm" 
        :rules="formRules" 
        label-width="100px"
        class="namespace-form"
      >
        <el-form-item label="集群" prop="clusterId">
          <el-select 
            v-model="namespaceForm.clusterId" 
            placeholder="请选择集群" 
            style="width: 100%;"
            :disabled="isEdit"
          >
            <el-option 
              v-for="cluster in clusterList" 
              :key="cluster.id" 
              :label="cluster.name" 
              :value="cluster.id" 
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="命名空间" prop="namespace">
          <el-input 
            v-model="namespaceForm.namespace" 
            placeholder="请输入命名空间名称" 
            :disabled="isEdit"
          />
        </el-form-item>
        
        <el-form-item label="别名" prop="alias">
          <el-input 
            v-model="namespaceForm.alias" 
            placeholder="请输入别名" 
          />
        </el-form-item>
        
        <el-form-item label="描述" prop="describe">
          <el-input 
            v-model="namespaceForm.describe" 
            type="textarea" 
            :rows="3" 
            placeholder="请输入描述" 
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="handleDialogClose">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitLoading">确定</el-button>
        </div>
      </template>
    </UnifiedDialog>

    <!-- YAML添加命名空间对话框 -->
    <UnifiedDialog 
      v-model="yamlDialogVisible" 
      title="YAML添加命名空间" 
      subtitle="通过 YAML 快速创建"
      width="80%"
    >
      <el-form 
        :model="yamlForm" 
        label-width="100px"
      >
        <el-form-item label="集群">
          <el-select 
            v-model="yamlForm.clusterId" 
            placeholder="请选择集群" 
            style="width: 300px"
          >
            <el-option 
              v-for="cluster in clusterList" 
              :key="cluster.id" 
              :label="cluster.name" 
              :value="cluster.id" 
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="YAML配置" class="yaml-form-item">
           <YamlEditor
             v-model="yamlForm.yaml"
             title="命名空间 YAML"
             :readonly="false"
             height="500px"
             filename="namespace.yaml"
           />
         </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="yamlDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleApplyYaml" :loading="yamlSubmitLoading">应用</el-button>
        </div>
      </template>
    </UnifiedDialog>

    <!-- 查看/编辑YAML对话框 -->
    <UnifiedDialog 
      v-model="viewYamlDialogVisible" 
      title="查看/编辑YAML" 
      subtitle="命名空间配置"
      width="90%"
    >
      <div class="yaml-dialog-content">
        <div class="yaml-info">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="集群">{{ viewYamlForm.clusterName }}</el-descriptions-item>
            <el-descriptions-item label="命名空间">{{ viewYamlForm.namespace }}</el-descriptions-item>
            <el-descriptions-item label="名称">{{ currentNamespace?.name }}</el-descriptions-item>
            <el-descriptions-item label="类型">Namespace</el-descriptions-item>
          </el-descriptions>
        </div>
        
        <div class="yaml-editor-wrapper">
          <YamlEditor
            v-model="viewYamlForm.yaml"
            :title="`${currentNamespace?.name} - Namespace YAML`"
            :readonly="false"
            height="500px"
            :filename="`${currentNamespace?.name}-namespace.yaml`"
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
    </UnifiedDialog>

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
import { ref, reactive, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, 
  Search, 
  Refresh, 
  ArrowDown,
  Document,
  Edit
} from '@element-plus/icons-vue'
import { 
  namespaceApi, 
  type NamespaceVO, 
  type NamespacePageQueryDTO, 
  type NamespaceCreateDTO, 
  type NamespaceApplyDTO,
  type NamespaceBaseDTO
} from '@/api/namespace'
import { clusterApi, type ClusterVO } from '@/api/cluster'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import UnifiedDialog from '@/components/UnifiedDialog.vue'
import YamlEditor from '@/components/YamlEditor.vue'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const submitLoading = ref(false)
const applyLoading = ref(false)
const yamlSubmitLoading = ref(false)

// 搜索表单
const searchForm = reactive({
  clusterId: '',
  keyword: ''
})

// 命名空间列表
const namespaceList = ref<NamespaceVO[]>([])
const selectedNamespaces = ref<NamespaceVO[]>([])

// 集群列表
const clusterList = ref<ClusterVO[]>([])

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 对话框
const dialogVisible = ref(false)
const yamlDialogVisible = ref(false)
const isEdit = ref(false)
const dialogTitle = ref('')
const formRef = ref()

// 命名空间表单
const namespaceForm = reactive({
  clusterId: '',
  namespace: '',
  alias: '',
  describe: ''
})

// YAML表单
const yamlForm = reactive({
  clusterId: '',
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

// 表单验证规则
const formRules = {
  clusterId: [
    { required: true, message: '请选择集群', trigger: 'change' }
  ],
  namespace: [
    { required: true, message: '请输入命名空间名称', trigger: 'blur' },
    { pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/, message: '命名空间名称格式不正确', trigger: 'blur' }
  ]
}

// 获取命名空间列表
const fetchNamespaceList = async () => {
  if (!searchForm.clusterId) {
    namespaceList.value = []
    pagination.total = 0
    return
  }
  
  try {
    loading.value = true
    const params: NamespacePageQueryDTO = {
      clusterId: searchForm.clusterId,
      keyword: searchForm.keyword,
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    
    const response = await namespaceApi.getList(params)
    console.log('命名空间列表API响应:', response)
    
    if (response.code === 200 && response.data) {
      namespaceList.value = response.data.items || []
      pagination.total = response.data.total || 0
      console.log('解析后的命名空间列表:', namespaceList.value)
    }
  } catch (error: any) {
    console.error('获取命名空间列表失败:', error)
    ElMessage.error('获取命名空间列表失败')
  } finally {
    loading.value = false
  }
}

// 获取集群列表
const fetchClusterList = async () => {
  try {
    const response = await clusterApi.getClusters({ page: 1, pageSize: 10000 })
    console.log('集群列表API响应:', response)
    
    if (response.code === 200 && response.data) {
      clusterList.value = response.data.items || []
      console.log('解析后的集群列表:', clusterList.value)
      
      // 如果没有选中集群且有集群数据，选择第一个
      if (!searchForm.clusterId && clusterList.value.length > 0) {
        searchForm.clusterId = clusterList.value[0].id || clusterList.value[0].name
        console.log('自动选择集群:', searchForm.clusterId)
        // 自动查询第一个集群的命名空间列表
        fetchNamespaceList()
      }
    }
  } catch (error: any) {
    console.error('获取集群列表失败:', error)
    ElMessage.error('获取集群列表失败')
  }
}

// 集群选择变化
const handleClusterChange = () => {
  pagination.page = 1
  fetchNamespaceList()
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchNamespaceList()
}

// 重置
const handleReset = () => {
  searchForm.clusterId = ''
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
  isEdit.value = false
  dialogTitle.value = '添加命名空间'
  Object.assign(namespaceForm, {
    clusterId: searchForm.clusterId || '',
    namespace: '',
    alias: '',
    describe: ''
  })
  dialogVisible.value = true
}

// YAML方式添加命名空间
const handleAddNamespaceByYaml = () => {
  Object.assign(yamlForm, {
    clusterId: searchForm.clusterId || '',
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
  if (!yamlForm.clusterId) {
    ElMessage.warning('请选择集群')
    return
  }
  
  if (!yamlForm.yaml.trim()) {
    ElMessage.warning('YAML内容不能为空')
    return
  }
  
  try {
    yamlSubmitLoading.value = true
    
    const data: NamespaceApplyDTO = {
      clusterId: yamlForm.clusterId,
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
  isEdit.value = true
  dialogTitle.value = '编辑命名空间'
  Object.assign(namespaceForm, {
    clusterId: searchForm.clusterId,
    namespace: row.name,
    alias: row.alias || '',
    describe: row.describe || ''
  })
  dialogVisible.value = true
}

// 查看详情
const handleViewDetail = (row: NamespaceVO) => {
  router.push({
    path: '/namespace/detail',
    query: {
      clusterId: searchForm.clusterId,
      namespace: row.name
    }
  })
}

// 查看YAML
const handleViewYaml = async (row: NamespaceVO) => {
  try {
    currentNamespace.value = row
    
    // 获取当前集群名称
    const cluster = clusterList.value.find(c => c.id === searchForm.clusterId)
    viewYamlForm.clusterName = cluster?.name || ''
    viewYamlForm.namespace = row.name
    
    const params: NamespaceBaseDTO = {
      clusterId: searchForm.clusterId,
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
    path: '/pod',
    query: {
      clusterId: searchForm.clusterId,
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
      clusterId: searchForm.clusterId,
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

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    submitLoading.value = true
    
    const data: NamespaceCreateDTO = {
      clusterId: namespaceForm.clusterId,
      namespace: namespaceForm.namespace,
      alias: namespaceForm.alias,
      describe: namespaceForm.describe
    }
    
    if (isEdit.value) {
      await namespaceApi.update(data)
      ElMessage.success('更新成功')
    } else {
      await namespaceApi.create(data)
      ElMessage.success('创建成功')
    }
    
    dialogVisible.value = false
    fetchNamespaceList()
  } catch (error: any) {
    ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
  } finally {
    submitLoading.value = false
  }
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
      clusterId: searchForm.clusterId,
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

// 关闭对话框
const handleDialogClose = () => {
  dialogVisible.value = false
  if (formRef.value) {
    formRef.value.resetFields()
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
  // 检查是否从其他页面跳转过来带有集群ID
  const clusterId = route.query.clusterId as string
  if (clusterId) {
    searchForm.clusterId = clusterId
  }
  
  fetchClusterList()
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
</style>
