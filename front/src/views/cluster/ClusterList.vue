<template>
  <div class="cluster-list">
    <div class="page-header">
      <h1>集群列表</h1>
      <el-button type="primary" @click="handleAddCluster">
        <el-icon><Plus /></el-icon>
        添加集群
      </el-button>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="集群名称">
          <el-input 
            v-model="searchForm.name" 
            placeholder="请输入集群名称" 
            clearable
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select 
            v-model="searchForm.status" 
            placeholder="请选择状态" 
            clearable
            style="width: 150px;"
          >
            <el-option label="运行中" value="Active" />
            <el-option label="异常" value="Error" />
            <el-option label="离线" value="Offline" />
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
      <el-table 
        v-loading="loading" 
        :data="clusterList" 
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column v-if="hasColumnData('id')" prop="id" label="集群ID" width="280" show-overflow-tooltip />
        <el-table-column prop="name" label="集群名称" width="80">
          <template #default="{ row }">
            <el-link type="primary" @click="handleViewDetail(row)">
              {{ row.name }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column v-if="hasColumnData('alias')" prop="alias" label="集群别名" width="160" show-overflow-tooltip />
        <el-table-column v-if="hasColumnData('describe') || hasColumnData('description')" prop="description" label="描述" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.describe || row.description }}
          </template>
        </el-table-column>
        <el-table-column v-if="hasColumnData('version')" prop="version" label="版本" width="120" />
        <el-table-column v-if="hasColumnData('status')" prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag 
              :type="getStatusType(row.status)"
              size="small"
            >
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <div style="display: flex; gap: 4px; align-items: center; flex-wrap: nowrap;">
              <el-button size="small" @click="handleViewDetail(row)">
                详情
              </el-button>
              <el-button size="small" @click="handleViewNodes(row)">
                节点
              </el-button>
              <el-button size="small" @click="handleViewPods(row)">
                Pod
              </el-button>
              <el-dropdown @command="(command) => handleMoreAction(command, row)">
                <el-button size="small">
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="edit">编辑</el-dropdown-item>
                    <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

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

    <!-- 添加/编辑集群对话框 -->
    <el-dialog 
      v-model="dialogVisible" 
      :title="dialogTitle" 
      width="800px"
      @close="handleDialogClose"
    >
      <el-form ref="formRef" :model="clusterForm" :rules="formRules" label-width="100px" class="centered-form">
        <el-form-item label="集群名称" prop="name">
          <el-input v-model="clusterForm.name" placeholder="请输入集群名称" />
        </el-form-item>
        <el-form-item v-if="showField('alias')" label="集群别名" prop="alias">
          <el-input v-model="clusterForm.alias" placeholder="请输入集群别名" />
        </el-form-item>
        <el-form-item v-if="showField('describe')" label="描述" prop="describe">
          <el-input 
            v-model="clusterForm.describe" 
            type="textarea" 
            :rows="3"
            placeholder="请输入集群描述"
          />
        </el-form-item>
        <el-form-item v-if="showField('config')" label="集群配置" prop="config">
          <div class="yaml-editor-container" style="width: 100%; max-width: 100%; overflow: hidden;">
            <codemirror
              v-model="clusterForm.config"
              placeholder="请输入集群配置（kubeconfig内容）"
              :style="{ height: '300px', width: '100%', maxWidth: '100%', border: '1px solid #dcdfe6', borderRadius: '4px', overflow: 'hidden' }"
              :autofocus="true"
              :indent-with-tab="true"
              :tab-size="2"
              :extensions="extensions"
            />
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitLoading" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, 
  Search, 
  Refresh, 
  ArrowDown 
} from '@element-plus/icons-vue'
import { clusterApi, type ClusterVO, type ClusterQueryDTO, type CreateClusterDTO } from '@/api/cluster'
import { Codemirror } from 'vue-codemirror'
import { yaml } from '@codemirror/lang-yaml'
import { oneDark } from '@codemirror/theme-one-dark'

const router = useRouter()

// 搜索表单
const searchForm = reactive<ClusterQueryDTO>({
  keyword: '',
  status: '',
  version: ''
})

// 集群列表
const clusterList = ref<ClusterVO[]>([])
const loading = ref(false)
const selectedClusters = ref([])

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 对话框
const dialogVisible = ref(false)
const dialogTitle = ref('添加集群')
const submitLoading = ref(false)
const formRef = ref()

// 集群表单
const clusterForm = reactive<CreateClusterDTO>({
  name: '',
  alias: '',
  describe: '',
  config: ''
})

// CodeMirror 扩展配置
const extensions = [yaml()]

const currentClusterId = ref('')

// 可用字段（根据后端返回的数据动态设置）
const availableFields = ref<Set<string>>(new Set(['name'])) // 默认总是显示name字段

// 判断是否显示字段
const showField = (fieldName: string) => {
  // 新增时显示所有字段
  if (!currentClusterId.value) {
    return true
  }
  // 编辑时只显示有数据的字段
  return availableFields.value.has(fieldName)
}

// 判断列表中是否有某个字段的数据
const hasColumnData = (fieldName: string) => {
  if (!clusterList.value || clusterList.value.length === 0) {
    return false
  }
  // 检查列表中是否有任何一条记录包含该字段且有值
  return clusterList.value.some(item => {
    const value = item[fieldName as keyof ClusterVO]
    return value !== undefined && value !== null && value !== ''
  })
}

// 表单验证规则
const formRules = {
  name: [
    { required: true, message: '请输入集群名称', trigger: 'blur' }
  ],
  config: [
    { required: true, message: '请输入集群配置', trigger: 'blur' }
  ]
}

// 获取集群列表
const fetchClusterList = async () => {
  loading.value = true
  try {
    const response = await clusterApi.getClusters({
      ...searchForm,
      page: pagination.page,
      pageSize: pagination.size
    })
    
    if (response.code === 200 && response.data) {
      // 处理API返回的数据结构
      clusterList.value = response.data.items || []
      pagination.total = response.data.total || 0
      

    } else {
      clusterList.value = []
      pagination.total = 0
      ElMessage.warning(response.msg || '暂无集群数据')
    }
  } catch (error) {
    console.error('加载集群列表失败:', error)
    clusterList.value = []
    pagination.total = 0
    ElMessage.error('加载集群列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchClusterList()
}

// 重置
const handleReset = () => {
  searchForm.name = ''
  searchForm.status = ''
  pagination.page = 1
  fetchClusterList()
}

// 分页处理
const handleSizeChange = (size: number) => {
  pagination.size = size
  fetchClusterList()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchClusterList()
}

// 选择处理
const handleSelectionChange = (selection: any[]) => {
  selectedClusters.value = selection
}

// 添加集群
const handleAddCluster = () => {
  dialogTitle.value = '添加集群'
  // 新增时显示所有字段
  availableFields.value = new Set(['name', 'alias', 'describe', 'config'])
  resetForm()
  dialogVisible.value = true
}

// 查看详情
const handleViewDetail = (row: ClusterVO) => {
  router.push(`/cluster/detail/${row.name}`)
}

// 查看节点
const handleViewNodes = (row: ClusterVO) => {
  router.push({
    path: '/cluster/nodes',
    query: { clusterId: row.id || row.name }
  })
}

// 查看Pod
const handleViewPods = (row: ClusterVO) => {
  router.push(`/cluster/pods/${row.name}`)
}

// 更多操作
const handleMoreAction = async (command: string, row: ClusterVO) => {
  switch (command) {
    case 'edit':
      handleEditCluster(row)
      break
    case 'delete':
      await handleDeleteCluster(row)
      break
  }
}

// 编辑集群
const handleEditCluster = async (row: ClusterVO) => {
  try {
    dialogTitle.value = '编辑集群'
    
    // 调用详情接口获取完整信息包括配置
    const response = await clusterApi.getCluster(row.id || row.name)
    
    if (response.code === 200 && response.data) {
      const clusterDetail = response.data
      
      // 重置可用字段
      availableFields.value = new Set(['name']) // name字段总是显示
      
      // 根据返回的数据设置表单值和可用字段
      const formData: any = { name: clusterDetail.name || '' }
      
      if (clusterDetail.alias !== undefined && clusterDetail.alias !== null && clusterDetail.alias !== '') {
        formData.alias = clusterDetail.alias
        availableFields.value.add('alias')
      }
      
      if ((clusterDetail.describe !== undefined && clusterDetail.describe !== null && clusterDetail.describe !== '') ||
          (clusterDetail.description !== undefined && clusterDetail.description !== null && clusterDetail.description !== '')) {
        formData.describe = clusterDetail.describe || clusterDetail.description || ''
        availableFields.value.add('describe')
      }
      
      if (clusterDetail.config !== undefined && clusterDetail.config !== null && clusterDetail.config !== '') {
        formData.config = clusterDetail.config
        availableFields.value.add('config')
      }
      
      Object.assign(clusterForm, formData)
    } else {
      // 如果详情接口失败，使用列表数据
      availableFields.value = new Set(['name'])
      Object.assign(clusterForm, {
        name: row.name || '',
        alias: '',
        describe: '',
        config: ''
      })
    }
    
    currentClusterId.value = row.id || row.name
    dialogVisible.value = true
  } catch (error) {
    console.error('获取集群详情失败:', error)
    ElMessage.error('获取集群详情失败')
  }
}

// 删除集群
const handleDeleteCluster = async (row: ClusterVO) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除集群 "${row.name}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await clusterApi.deleteCluster(row.name!)
    
    ElMessage.success('集群删除成功')
    fetchClusterList()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除集群失败:', error)
      ElMessage.error('删除集群失败')
    }
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    
    submitLoading.value = true
    try {
      if (currentClusterId.value) {
        await clusterApi.updateCluster(currentClusterId.value, clusterForm)
        ElMessage.success('集群更新成功')
      } else {
        await clusterApi.createCluster(clusterForm)
        ElMessage.success('集群创建成功')
      }
      
      dialogVisible.value = false
      fetchClusterList()
    } catch (error) {
      console.error('操作失败:', error)
      ElMessage.error('操作失败')
    } finally {
      submitLoading.value = false
    }
  })
}

// 重置表单
const resetForm = () => {
  Object.assign(clusterForm, {
    name: '',
    alias: '',
    describe: '',
    config: ''
  })
  currentClusterId.value = ''
  // 重置可用字段为默认状态
  availableFields.value = new Set(['name'])
  formRef.value?.clearValidate()
}

// 关闭对话框
const handleDialogClose = () => {
  resetForm()
}

// 获取状态类型
const getStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    Active: 'success',
    Error: 'danger',
    Offline: 'info',
    Unknown: 'warning'
  }
  return statusMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    Active: '运行中',
    Error: '异常', 
    Offline: '离线',
    Unknown: '未知'
  }
  return statusMap[status] || status
}

// 初始化
onMounted(() => {
  fetchClusterList()
})
</script>

<style scoped>
.cluster-list {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.search-card {
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.el-form--inline .el-form-item {
  margin-right: 20px;
}

.centered-form {
  .el-form-item__label {
    text-align: center;
    justify-content: center;
  }
  
  .el-input,
  .el-select,
  .el-textarea {
    text-align: center;
  }
  
  .el-input__inner,
  .el-textarea__inner {
    text-align: center;
  }
}
</style>