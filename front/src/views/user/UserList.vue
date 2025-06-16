<template>
  <div class="user-list-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">用户管理</h1>
        <p class="page-description">管理系统用户账户和权限</p>
      </div>
      <div class="header-right">
        <el-button type="primary" size="large" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          新增用户
        </el-button>
      </div>
    </div>

    <!-- 搜索和筛选区域 -->
    <el-card class="filter-card" shadow="never">
      <div class="filter-content">
        <div class="filter-left">
          <el-form :model="searchForm" inline>
            <el-form-item label="角色">
              <el-select 
                v-model="searchForm.role" 
                placeholder="选择角色" 
                clearable
                style="width: 140px;"
              >
                <el-option label="管理员" :value="2" />
                <el-option label="普通用户" :value="3" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select 
                v-model="searchForm.status" 
                placeholder="选择状态" 
                clearable
                style="width: 120px;"
              >
                <el-option label="正常" :value="1" />
                <el-option label="禁用" :value="0" />
              </el-select>
            </el-form-item>
          </el-form>
        </div>
        <div class="filter-right">
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 用户列表卡片 -->
    <el-card class="table-card" shadow="never">
      <!-- 表格工具栏 -->
      <div class="table-toolbar">
        <div class="toolbar-left">
          <span class="table-title">用户列表</span>
          <el-tag class="count-tag">{{ pagination.total }}</el-tag>
        </div>
        <div class="toolbar-right">
          <el-button 
            :disabled="selectedUsers.length === 0" 
            @click="handleBatchDelete"
            type="danger"
            plain
          >
            <el-icon><Delete /></el-icon>
            批量删除 ({{ selectedUsers.length }})
          </el-button>
          <el-button @click="fetchUserList" circle>
            <el-icon><Refresh /></el-icon>
          </el-button>
        </div>
      </div>
      
      <!-- 用户表格 -->
      <el-table 
        v-loading="loading" 
        :data="userList" 
        stripe
        class="user-table"
        @selection-change="handleSelectionChange"
        empty-text="暂无用户数据"
        style="width: 100%"
      >
        <el-table-column type="selection" width="50" />
        
        <el-table-column prop="id" label="用户ID" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <el-text class="user-id" type="info">{{ row.id }}</el-text>
          </template>
        </el-table-column>
        
        <el-table-column prop="username" label="用户名" min-width="150">
          <template #default="{ row }">
            <div class="user-info">
              <div class="user-avatar">
                <el-avatar :size="32" :src="getUserAvatar(row.username)">
                  {{ row.username.charAt(0).toUpperCase() }}
                </el-avatar>
              </div>
              <div class="user-details">
                <div class="user-name">{{ row.username }}</div>
              </div>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="email" label="邮箱" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.email">{{ row.email }}</span>
            <el-text v-else type="info">未设置</el-text>
          </template>
        </el-table-column>

        <el-table-column prop="role" label="角色" min-width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getRoleType(row.role)" effect="light">
              {{ getRoleText(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="status" label="状态" min-width="80" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" effect="light">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="department" label="部门" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.department">{{ row.department }}</span>
            <el-text v-else type="info">未设置</el-text>
          </template>
        </el-table-column>
        
        <el-table-column prop="phone" label="手机号" min-width="140">
          <template #default="{ row }">
            <span v-if="row.phone">{{ row.phone }}</span>
            <el-text v-else type="info">未设置</el-text>
          </template>
        </el-table-column>
        
        <el-table-column prop="createdAt" label="创建时间" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.createdAt">{{ formatDate(row.createdAt) }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" plain @click="handleEdit(row)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button 
                size="small" 
                :type="row.status === 1 ? 'warning' : 'success'"
                plain
                @click="handleToggleStatus(row)"
              >
                <el-icon><Switch /></el-icon>
                {{ row.status === 1 ? '禁用' : '启用' }}
              </el-button>
              <el-dropdown @command="(command) => handleMoreActions(command, row)">
                <el-button size="small" plain>
                  更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="resetPassword">
                      <el-icon><Key /></el-icon>
                      重置密码
                    </el-dropdown-item>
                    <el-dropdown-item command="permissions">
                      <el-icon><Lock /></el-icon>
                      权限管理
                    </el-dropdown-item>
                    <el-dropdown-item command="delete" divided>
                      <el-icon><Delete /></el-icon>
                      删除用户
                    </el-dropdown-item>
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
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 新增/编辑用户对话框 -->
    <el-dialog 
      v-model="dialogVisible" 
      :title="dialogTitle" 
      width="600px"
      :before-close="handleDialogClose"
      destroy-on-close
    >
      <el-form 
        ref="formRef" 
        :model="userForm" 
        :rules="formRules" 
        label-width="100px"
        class="user-form"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="用户名" prop="username">
              <el-input 
                v-model="userForm.username" 
                placeholder="请输入用户名" 
                :disabled="isEdit"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="userForm.email" placeholder="请输入邮箱" />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20" v-if="!isEdit">
          <el-col :span="24">
            <el-form-item label="密码" prop="password">
              <el-input 
                v-model="userForm.password" 
                type="password" 
                placeholder="请输入密码" 
                show-password
              />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="角色" prop="role">
              <el-select v-model="userForm.role" placeholder="请选择角色" style="width: 100%;">
                <el-option label="管理员" :value="2" />
                <el-option label="普通用户" :value="3" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态" prop="status">
              <el-radio-group v-model="userForm.status">
                <el-radio :label="1">正常</el-radio>
                <el-radio :label="0">禁用</el-radio>
              </el-radio-group>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="部门" prop="department">
              <el-input v-model="userForm.department" placeholder="请输入部门" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机号" prop="phone">
              <el-input v-model="userForm.phone" placeholder="请输入手机号" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="handleDialogClose">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitLoading">
            确定
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog 
      v-model="resetPasswordDialogVisible" 
      title="重置密码" 
      width="400px"
      destroy-on-close
    >
      <el-form 
        ref="resetPasswordFormRef" 
        :model="resetPasswordForm" 
        :rules="resetPasswordRules" 
        label-width="80px"
      >
        <el-form-item label="新密码" prop="newPassword">
          <el-input 
            v-model="resetPasswordForm.newPassword" 
            type="password" 
            placeholder="请输入新密码" 
            show-password
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="resetPasswordDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleResetPasswordSubmit" :loading="resetPasswordLoading">
            确定
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, 
  Search, 
  Refresh, 
  Delete, 
  ArrowDown,
  Edit,
  Switch,
  Key,
  Lock
} from '@element-plus/icons-vue'
import { 
  userApi, 
  type UserVO, 
  type UserQueryDTO, 
  type CreateUserDTO, 
  type UpdateUserDTO, 
  type ResetPasswordDTO 
} from '@/api/user'

const loading = ref(false)
const submitLoading = ref(false)
const resetPasswordLoading = ref(false)

// 搜索表单
const searchForm = reactive({
  role: undefined as number | undefined,
  status: undefined as number | undefined
})

// 用户列表
const userList = ref<UserVO[]>([])
const selectedUsers = ref<UserVO[]>([])

// 分页
const pagination = reactive({
  page: 1,
  size: 20,
  total: 0
})

// 对话框
const dialogVisible = ref(false)
const isEdit = ref(false)
const dialogTitle = ref('')
const formRef = ref()

// 用户表单
const userForm = reactive({
  id: '',
  username: '',
  email: '',
  password: '',
  role: 3,
  department: '',
  phone: '',
  status: 1
})

// 表单验证规则
const formRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ]
}

// 重置密码对话框
const resetPasswordDialogVisible = ref(false)
const resetPasswordFormRef = ref()
const currentUserId = ref('')

// 重置密码表单
const resetPasswordForm = reactive<ResetPasswordDTO>({
  userId: '',
  newPassword: ''
})

// 重置密码验证规则
const resetPasswordRules = {
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ]
}

// 获取用户头像
const getUserAvatar = (username: string) => {
  // 这里可以根据实际需求返回头像URL
  return `https://api.dicebear.com/7.x/initials/svg?seed=${username}`
}

// 获取角色类型
const getRoleType = (role: number) => {
  switch (role) {
    case 1:
      return 'danger'
    case 2:
      return 'warning'
    case 3:
      return 'success'
    default:
      return 'info'
  }
}

// 获取角色文本
const getRoleText = (role: number) => {
  switch (role) {
    case 1:
      return '超级管理员'
    case 2:
      return '管理员'
    case 3:
      return '普通用户'
    default:
      return '未知'
  }
}

// 获取状态类型
const getStatusType = (status: number) => {
  return status === 1 ? 'success' : 'danger'
}

// 获取状态文本
const getStatusText = (status: number) => {
  return status === 1 ? '正常' : '禁用'
}

// 获取用户列表
const fetchUserList = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.size,
      ...searchForm
    }
    const res = await userApi.getUsers(params)
    
    if (res.code === 200 && res.data) {
      userList.value = res.data.items || res.data.list || []
      pagination.total = res.data.total || 0
    } else {
      userList.value = []
      pagination.total = 0
      ElMessage.warning(res.msg || '暂无用户数据')
    }
  } catch (error: any) {
    console.error('获取用户列表失败:', error)
    userList.value = []
    pagination.total = 0
    
    if (error.response?.status === 403) {
      ElMessage.error('没有权限访问用户列表，请联系管理员')
    } else if (error.response?.status === 401) {
      ElMessage.error('登录已过期，请重新登录')
    } else {
      ElMessage.error('获取用户列表失败')
    }
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.page = 1
  fetchUserList()
}

// 重置
const handleReset = () => {
  Object.assign(searchForm, {
    role: undefined,
    status: undefined
  })
  pagination.page = 1
  fetchUserList()
}

// 分页大小改变
const handleSizeChange = (size: number) => {
  pagination.size = size
  pagination.page = 1
  fetchUserList()
}

// 当前页改变
const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchUserList()
}

// 选择改变
const handleSelectionChange = (selection: UserVO[]) => {
  selectedUsers.value = selection
}

// 新增用户
const handleAdd = () => {
  isEdit.value = false
  dialogTitle.value = '新增用户'
  Object.assign(userForm, {
    id: '',
    username: '',
    email: '',
    password: '',
    role: 3,
    department: '',
    phone: '',
    status: 1
  })
  dialogVisible.value = true
}

// 编辑用户
const handleEdit = (row: UserVO) => {
  isEdit.value = true
  dialogTitle.value = '编辑用户'
  Object.assign(userForm, {
    id: row.id,
    username: row.username,
    email: row.email || '',
    password: '',
    role: row.role,
    department: row.department || '',
    phone: row.phone || '',
    status: row.status
  })
  dialogVisible.value = true
}

// 切换用户状态
const handleToggleStatus = async (row: UserVO) => {
  const newStatus = row.status === 1 ? 0 : 1
  const action = newStatus === 1 ? '启用' : '禁用'
  
  try {
    await ElMessageBox.confirm(
      `确定要${action}用户 "${row.username}" 吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await userApi.toggleUserStatus(row.id, newStatus)
    ElMessage.success(`${action}成功`)
    fetchUserList()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}失败`)
    }
  }
}

// 更多操作
const handleMoreActions = (command: string, row: UserVO) => {
  switch (command) {
    case 'resetPassword':
      handleResetPassword(row)
      break
    case 'permissions':
      ElMessage.info('权限管理功能开发中')
      break
    case 'delete':
      handleDelete(row)
      break
  }
}

// 重置密码
const handleResetPassword = (row: UserVO) => {
  currentUserId.value = row.id
  resetPasswordForm.userId = row.id
  resetPasswordForm.newPassword = ''
  resetPasswordDialogVisible.value = true
}

// 删除用户
const handleDelete = async (row: UserVO) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 "${row.username}" 吗？此操作不可恢复！`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await userApi.deleteUser(row.id)
    ElMessage.success('删除成功')
    fetchUserList()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 批量删除
const handleBatchDelete = async () => {
  if (selectedUsers.value.length === 0) {
    ElMessage.warning('请选择要删除的用户')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedUsers.value.length} 个用户吗？此操作不可恢复！`,
      '确认批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const ids = selectedUsers.value.map(user => user.id)
    await userApi.batchDeleteUsers(ids)
    ElMessage.success('批量删除成功')
    selectedUsers.value = []
    fetchUserList()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    submitLoading.value = true
    
    if (isEdit.value) {
      const updateData: UpdateUserDTO = {
        id: userForm.id,
        username: userForm.username,
        email: userForm.email,
        role: userForm.role,
        department: userForm.department,
        phone: userForm.phone,
        status: userForm.status
      }
      await userApi.updateUser(updateData)
      ElMessage.success('更新成功')
    } else {
      const createData: CreateUserDTO = {
        username: userForm.username,
        password: userForm.password,
        email: userForm.email,
        role: userForm.role,
        department: userForm.department,
        phone: userForm.phone,
        status: userForm.status
      }
      await userApi.createUser(createData)
      ElMessage.success('创建成功')
    }
    
    dialogVisible.value = false
    fetchUserList()
  } catch (error: any) {
    ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
  } finally {
    submitLoading.value = false
  }
}

// 重置密码提交
const handleResetPasswordSubmit = async () => {
  if (!resetPasswordFormRef.value) return
  
  try {
    await resetPasswordFormRef.value.validate()
    resetPasswordLoading.value = true
    
    await userApi.resetPassword(resetPasswordForm)
    ElMessage.success('密码重置成功')
    resetPasswordDialogVisible.value = false
  } catch (error: any) {
    ElMessage.error('密码重置失败')
  } finally {
    resetPasswordLoading.value = false
  }
}

// 关闭对话框
const handleDialogClose = () => {
  dialogVisible.value = false
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

// 格式化日期
const formatDate = (dateString: string) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 页面加载时获取用户列表
onMounted(() => {
  fetchUserList()
})
</script>

<style scoped>
.user-list-container {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header-left {
  flex: 1;
}

.page-title {
  margin: 0 0 8px 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.page-description {
  margin: 0;
  color: #7f8c8d;
  font-size: 14px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-card {
  margin-bottom: 20px;
  border: none;
}

.filter-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-left {
  flex: 1;
}

.filter-right {
  display: flex;
  gap: 12px;
}

.table-card {
  border: none;
}

.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #ebeef5;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.table-title {
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
}

.count-tag {
  background-color: #f0f9ff;
  color: #0369a1;
  border: 1px solid #bae6fd;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-table {
  margin-bottom: 20px;
}

.user-id {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 12px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-avatar {
  flex-shrink: 0;
}

.user-details {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-weight: 500;
  color: #2c3e50;
  margin-bottom: 4px;
}

.user-email {
  font-size: 12px;
  color: #7f8c8d;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.user-form {
  padding: 0 20px;
}

.dialog-footer {
  text-align: right;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .user-list-container {
    padding: 10px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .filter-content {
    flex-direction: column;
    gap: 16px;
    align-items: stretch;
  }
  
  .table-toolbar {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }
  
  .toolbar-left,
  .toolbar-right {
    justify-content: center;
  }
  
  .action-buttons {
    flex-direction: column;
  }
  
  .user-info {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .user-avatar {
    align-self: center;
  }
}
</style>