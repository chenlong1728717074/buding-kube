<template>
  <div class="user-list">
    <div class="page-header">
      <h1>用户管理</h1>
      <div class="header-actions">
        <el-button type="primary" @click="handleAdd">
          <el-icon><Plus /></el-icon>
          新增用户
        </el-button>
      </div>
    </div>

    <!-- 搜索表单 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="用户名">
          <el-input 
            v-model="searchForm.username" 
            placeholder="请输入用户名" 
            clearable
            style="width: 200px;"
          />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input 
            v-model="searchForm.email" 
            placeholder="请输入邮箱" 
            clearable
            style="width: 200px;"
          />
        </el-form-item>
        <el-form-item label="角色">
          <el-select 
            v-model="searchForm.role" 
            placeholder="请选择角色" 
            clearable
            style="width: 150px;"
          >
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="user" />
            <el-option label="只读用户" value="readonly" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select 
            v-model="searchForm.status" 
            placeholder="请选择状态" 
            clearable
            style="width: 120px;"
          >
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="inactive" />
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

    <!-- 用户列表 -->
    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>用户列表 ({{ pagination.total }})</span>
          <div class="header-actions">
            <el-button 
              :disabled="selectedUsers.length === 0" 
              @click="handleBatchDelete"
            >
              <el-icon><Delete /></el-icon>
              批量删除
            </el-button>
          </div>
        </div>
      </template>
      
      <el-table 
        v-loading="loading" 
        :data="userList" 
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column label="用户信息" min-width="250">
          <template #default="{ row }">
            <div class="user-info">
              <el-avatar 
                :size="40" 
                :src="row.avatar" 
                :icon="UserFilled"
                class="user-avatar"
              />
              <div class="user-details">
                <div class="user-name">{{ row.username }}</div>
                <div class="user-email">{{ row.email }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="realName" label="真实姓名" width="120" />
        <el-table-column prop="role" label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="getRoleType(row.role)">
              {{ getRoleText(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="department" label="部门" width="150" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column prop="lastLoginTime" label="最后登录" width="180">
          <template #default="{ row }">
            <span v-if="row.lastLoginTime">{{ row.lastLoginTime }}</span>
            <span v-else class="never-login">从未登录</span>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button 
              size="small" 
              :type="row.status === 'active' ? 'warning' : 'success'"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 'active' ? '禁用' : '启用' }}
            </el-button>
            <el-dropdown @command="(command) => handleMoreActions(command, row)">
              <el-button size="small">
                更多<el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="resetPassword">
                    重置密码
                  </el-dropdown-item>
                  <el-dropdown-item command="permissions">
                    权限管理
                  </el-dropdown-item>
                  <el-dropdown-item command="loginHistory">
                    登录历史
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" divided>
                    删除用户
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

    <!-- 新增/编辑用户对话框 -->
    <el-dialog 
      v-model="dialogVisible" 
      :title="dialogTitle" 
      width="600px"
      :before-close="handleDialogClose"
    >
      <el-form 
        ref="formRef" 
        :model="userForm" 
        :rules="formRules" 
        label-width="100px"
      >
        <el-form-item label="用户名" prop="username">
          <el-input 
            v-model="userForm.username" 
            placeholder="请输入用户名" 
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="userForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item v-if="!isEdit" label="密码" prop="password">
          <el-input 
            v-model="userForm.password" 
            type="password" 
            placeholder="请输入密码" 
            show-password
          />
        </el-form-item>
        <el-form-item label="真实姓名" prop="realName">
          <el-input v-model="userForm.realName" placeholder="请输入真实姓名" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="userForm.role" placeholder="请选择角色" style="width: 100%;">
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="user" />
            <el-option label="只读用户" value="readonly" />
          </el-select>
        </el-form-item>
        <el-form-item label="部门" prop="department">
          <el-input v-model="userForm.department" placeholder="请输入部门" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="userForm.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="userForm.status">
            <el-radio label="active">启用</el-radio>
            <el-radio label="inactive">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注">
          <el-input 
            v-model="userForm.remark" 
            type="textarea" 
            :rows="3" 
            placeholder="请输入备注"
          />
        </el-form-item>
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
  UserFilled, 
  ArrowDown
} from '@element-plus/icons-vue'
import { userApi, type UserVO, type UserQueryDTO, type CreateUserDTO, type ResetPasswordDTO } from '@/api/user'

const loading = ref(false)
const submitLoading = ref(false)
const resetPasswordLoading = ref(false)

// 搜索表单
const searchForm = reactive<UserQueryDTO>({
  username: '',
  email: '',
  role: '',
  status: ''
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
const userForm = reactive<CreateUserDTO>({
  id: '',
  username: '',
  email: '',
  password: '',
  realName: '',
  role: 'user',
  department: '',
  phone: '',
  status: 'active',
  remark: ''
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
  realName: [
    { required: true, message: '请输入真实姓名', trigger: 'blur' }
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
      // 处理API返回的数据结构
      userList.value = res.data.items || res.data.list || []
      pagination.total = res.data.total || 0
      
      console.log('用户列表数据:', userList.value)
    } else {
      userList.value = []
      pagination.total = 0
      ElMessage.warning(res.msg || '暂无用户数据')
    }
  } catch (error) {
    console.error('获取用户列表失败:', error)
    userList.value = []
    pagination.total = 0
    
    // 检查是否是权限问题
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
    username: '',
    email: '',
    role: '',
    status: ''
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
const handleSelectionChange = (selection: any[]) => {
  selectedUsers.value = selection
}

// 新增用户
const handleAdd = () => {
  isEdit.value = false
  dialogTitle.value = '新增用户'
  resetForm()
  dialogVisible.value = true
}

// 编辑用户
const handleEdit = (user: UserVO) => {
  isEdit.value = true
  dialogTitle.value = '编辑用户'
  Object.assign(userForm, {
    ...user,
    password: '' // 编辑时不显示密码
  })
  dialogVisible.value = true
}

// 切换状态
const handleToggleStatus = async (user: UserVO) => {
  const action = user.status === 'active' ? '禁用' : '启用'
  try {
    await ElMessageBox.confirm(
      `确定要${action}用户 ${user.username} 吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const newStatus = user.status === 'active' ? 'inactive' : 'active'
    const res = await userApi.toggleUserStatus(user.id!, newStatus)
    if (res.code === 200) {
      ElMessage.success(`${action}成功`)
      fetchUserList()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error(`${action}失败:`, error)
      ElMessage.error(`${action}失败`)
    }
  }
}

// 更多操作
const handleMoreActions = async (command: string, user: UserVO) => {
  switch (command) {
    case 'resetPassword':
      handleResetPassword(user)
      break
    case 'permissions':
      handlePermissions(user)
      break
    case 'loginHistory':
      handleLoginHistory(user)
      break
    case 'delete':
      await handleDelete(user)
      break
  }
}

// 重置密码
const handleResetPassword = (user: UserVO) => {
  resetPasswordForm.userId = user.id!
  resetPasswordForm.newPassword = ''
  resetPasswordDialogVisible.value = true
}

// 权限管理
const handlePermissions = (user: UserVO) => {
  ElMessage.info('权限管理功能开发中')
}

// 登录历史
const handleLoginHistory = (user: UserVO) => {
  ElMessage.info('登录历史功能开发中')
}

// 删除用户
const handleDelete = async (user: UserVO) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 ${user.username} 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const res = await userApi.deleteUser(user.id!)
    if (res.code === 200) {
      ElMessage.success('删除成功')
      fetchUserList()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
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
      `确定要删除选中的 ${selectedUsers.value.length} 个用户吗？此操作不可恢复。`,
      '确认批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const ids = selectedUsers.value.map((user: UserVO) => user.id!)
    const res = await userApi.batchDeleteUsers(ids)
    if (res.code === 200) {
      ElMessage.success('批量删除成功')
      fetchUserList()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
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
    
    const res = isEdit.value 
      ? await userApi.updateUser(userForm.id!, userForm)
      : await userApi.createUser(userForm)
    
    if (res.code === 200) {
      ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
      dialogVisible.value = false
      fetchUserList()
    }
  } catch (error) {
    console.error('表单验证失败:', error)
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
    
    const res = await userApi.resetPassword(resetPasswordForm)
    if (res.code === 200) {
      ElMessage.success('重置密码成功')
      resetPasswordDialogVisible.value = false
    }
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    resetPasswordLoading.value = false
  }
}

// 重置表单
const resetForm = () => {
  Object.assign(userForm, {
    id: '',
    username: '',
    email: '',
    password: '',
    realName: '',
    role: 'user',
    department: '',
    phone: '',
    status: 'active',
    remark: ''
  })
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

// 关闭对话框
const handleDialogClose = () => {
  dialogVisible.value = false
  resetForm()
}

// 获取角色类型
const getRoleType = (role: string) => {
  const roleMap: Record<string, string> = {
    admin: 'danger',
    user: 'primary',
    readonly: 'info'
  }
  return roleMap[role] || 'info'
}

// 获取角色文本
const getRoleText = (role: string) => {
  const roleMap: Record<string, string> = {
    admin: '管理员',
    user: '普通用户',
    readonly: '只读用户'
  }
  return roleMap[role] || role
}

// 获取状态类型
const getStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    active: 'success',
    inactive: 'danger'
  }
  return statusMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    active: '启用',
    inactive: '禁用'
  }
  return statusMap[status] || status
}

// 初始化
onMounted(() => {
  fetchUserList()
})
</script>

<style scoped>
.user-list {
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-details {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.user-name {
  font-weight: 500;
  color: #2c3e50;
}

.user-email {
  font-size: 12px;
  color: #7f8c8d;
}

.never-login {
  color: #bdc3c7;
  font-style: italic;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.dialog-footer {
  text-align: right;
}

@media (max-width: 768px) {
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