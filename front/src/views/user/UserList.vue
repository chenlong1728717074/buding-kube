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

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <el-card class="stat-card" shadow="never">
        <div class="stat-label">总用户</div>
        <div class="stat-value">{{ pagination.total }}</div>
      </el-card>
      <el-card class="stat-card" shadow="never">
        <div class="stat-label">当前页启用</div>
        <div class="stat-value text-success">{{ activeCount }}</div>
      </el-card>
      <el-card class="stat-card" shadow="never">
        <div class="stat-label">当前页禁用</div>
        <div class="stat-value text-danger">{{ inactiveCount }}</div>
      </el-card>
      <el-card class="stat-card" shadow="never">
        <div class="stat-label">当前页管理员</div>
        <div class="stat-value text-warning">{{ adminCount }}</div>
      </el-card>
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
                @change="handleSearch"
                @clear="handleSearch"
              >
                <el-option label="超级管理员" value="super" />
                <el-option label="管理员" value="admin" />
                <el-option label="普通用户" value="normal" />
              </el-select>
            </el-form-item>
            <el-form-item label="状态">
              <el-select 
                v-model="searchForm.status" 
                placeholder="选择状态" 
                clearable
                style="width: 120px;"
                @change="handleSearch"
                @clear="handleSearch"
              >
                <el-option label="正常" value="active" />
                <el-option label="禁用" value="inactive" />
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
      <transition name="filter-fade">
        <div v-if="searchForm.role || searchForm.status" class="active-filters">
          <span class="active-filters-label">已筛选：</span>
          <el-tag v-if="searchForm.role" closable @close="removeFilter('role')">
            角色: {{ getRoleText(searchForm.role) }}
          </el-tag>
          <el-tag v-if="searchForm.status" closable @close="removeFilter('status')">
            状态: {{ getStatusText(searchForm.status) }}
          </el-tag>
          <el-button link type="primary" class="clear-filter-btn" @click="handleReset">清空筛选</el-button>
        </div>
      </transition>
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
        :row-class-name="tableRowClassName"
        @selection-change="handleSelectionChange"
        style="width: 100%"
      >
        <el-table-column type="selection" width="50" />

        <template #empty>
          <el-empty description="暂无用户数据">
            <el-button type="primary" @click="handleAdd">去创建用户</el-button>
          </el-empty>
        </template>

        <el-table-column prop="username" label="用户名" min-width="150">
          <template #default="{ row }">
            <div class="user-info">
              <div class="user-name">{{ row.username }}</div>
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
        
        <el-table-column prop="status" label="状态" min-width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" effect="light" class="status-tag">
              <el-icon class="status-icon"><component :is="getStatusIcon(row.status)" /></el-icon>
              <span>{{ getStatusText(row.status) }}</span>
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
        
        <el-table-column label="操作" width="280" fixed="right" align="center" header-align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" class="btn-edit" @click="handleEdit(row)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button
                size="small"
                class="btn-toggle"
                :type="isUserEnabled(row) ? 'warning' : 'success'"
                plain
                @click="handleToggleStatus(row)"
              >
                <el-icon><Switch /></el-icon>
                {{ isUserEnabled(row) ? '禁用' : '启用' }}
              </el-button>
              <el-dropdown @command="(command) => handleMoreActions(command, row)">
                <el-button size="small" class="btn-more" plain>
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
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="80%" :close-on-click-modal="false" class="config-dialog">
      <template #header>
        <div class="dialog-header">
          <h3 class="dialog-title">{{ dialogTitle }}</h3>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <el-form ref="formRef" :model="userForm" :rules="formRules" label-width="100px" class="user-form">
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
                <el-option
                  v-if="isEdit && (userForm.username === 'admin' || userForm.role === 'super')"
                  label="超级管理员"
                  value="super"
                />
                <el-option label="管理员" value="admin" />
                <el-option label="普通用户" value="normal" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态" prop="status">
              <el-radio-group v-model="userForm.status">
                <el-radio label="active">正常</el-radio>
                <el-radio label="inactive">禁用</el-radio>
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
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="handleDialogClose">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitLoading">
            确定
          </el-button>
        </div>
      </template>
    </el-dialog>


    <!-- 删除确认对话框（单个用户） -->
    <DeleteConfirmDialog
      v-model="deleteDialogVisible"
      :item-name="deleteItemName"
      message="确定要删除用户吗？此操作不可恢复！"
      :loading="deleteLoading"
      @confirm="confirmDeleteUser"
      @cancel="cancelDeleteUser"
    />

    <!-- 删除确认对话框（批量删除） -->
    <DeleteConfirmDialog
      v-model="batchDeleteDialogVisible"
      :item-name="`${selectedUsers.length} 个用户`"
      message="确定要批量删除选中的用户吗？此操作不可恢复！"
      :loading="batchDeleteLoading"
      @confirm="confirmBatchDeleteUsers"
      @cancel="cancelBatchDeleteUsers"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
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
  Lock,
  SuccessFilled,
  CircleCloseFilled,
  WarningFilled,
  InfoFilled
} from '@element-plus/icons-vue'
import {
  userApi,
  type UserVO,
  type UserQueryDTO,
  type CreateUserDTO,
  type UpdateUserDTO,
  UserStatus
} from '@/api/user'
import DeleteConfirmDialog from '@/components/DeleteConfirmDialog.vue'
import '@/assets/styles/config-editor.css'

const loading = ref(false)
const submitLoading = ref(false)

// 搜索表单
const searchForm = reactive({
  role: undefined as string | undefined,
  status: undefined as string | undefined
})

// 用户列表
const userList = ref<UserVO[]>([])
const selectedUsers = ref<UserVO[]>([])

const activeCount = computed(() => userList.value.filter(u => u.status === 'active').length)
const inactiveCount = computed(() => userList.value.filter(u => u.status === 'inactive').length)
const adminCount = computed(() => userList.value.filter(u => ['super', 'admin', '1', '2'].includes(String(u.role))).length)

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
  username: '',
  realName: '',
  email: '',
  password: '',
  role: 'normal',
  department: '',
  phone: '',
  status: 'active'
})

// 表单验证规则
const formRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  email: [
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


const normalizeRole = (role: string) => {
  switch (String(role)) {
    case '1':
    case 'super':
      return 'super'
    case '2':
    case 'admin':
      return 'admin'
    case '3':
    case 'normal':
      return 'normal'
    default:
      return String(role || '')
  }
}

// 获取角色类型
const getRoleType = (role: string) => {
  switch (normalizeRole(role)) {
    case 'super':
      return 'danger'
    case 'admin':
      return 'warning'
    case 'normal':
      return 'success'
    default:
      return 'info'
  }
}

// 获取角色文本
const getRoleText = (role: string) => {
  switch (normalizeRole(role)) {
    case 'super':
      return '超级管理员'
    case 'admin':
      return '管理员'
    case 'normal':
      return '普通用户'
    default:
      return role || '未知'
  }
}

// 获取状态类型
const getStatusType = (status: string) => {
  switch (status) {
    case 'active':
      return 'success'
    case 'inactive':
      return 'danger'
    case 'suspended':
      return 'warning'
    case 'expired':
      return 'info'
    default:
      return 'info'
  }
}

// 获取状态文本
const getStatusText = (status: string) => {
  switch (status) {
    case 'active':
      return '正常'
    case 'inactive':
      return '禁用'
    case 'suspended':
      return '暂停'
    case 'expired':
      return '过期'
    default:
      return status || '未知'
  }
}

// 获取状态图标
const getStatusIcon = (status: string) => {
  switch (status) {
    case 'active':
      return SuccessFilled
    case 'inactive':
      return CircleCloseFilled
    case 'suspended':
      return WarningFilled
    default:
      return InfoFilled
  }
}

const isUserEnabled = (user: UserVO) => {
  if (typeof user.enabled === 'boolean') {
    return user.enabled
  }
  return user.status === UserStatus.ACTIVE
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

const removeFilter = (key: 'role' | 'status') => {
  searchForm[key] = undefined
  handleSearch()
}

const tableRowClassName = ({ row }: { row: UserVO }) => {
  return selectedUsers.value.some(item => item.username === row.username) ? 'row-selected' : ''
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
    username: '',
    realName: '',
    email: '',
    password: '',
    role: 'normal',
    department: '',
    phone: '',
    status: 'active'
  })
  dialogVisible.value = true
}

// 编辑用户
const handleEdit = (row: UserVO) => {
  isEdit.value = true
  dialogTitle.value = '编辑用户'
  Object.assign(userForm, {
    username: row.username,
    realName: row.username,
    email: row.email || '',
    password: '',
    role: normalizeRole(String(row.role)),
    department: row.department || '',
    phone: row.phone || '',
    status: row.status
  })
  dialogVisible.value = true
}

// 切换用户状态
const handleToggleStatus = async (row: UserVO) => {
  const enable = !isUserEnabled(row)
  const action = enable ? '启用' : '禁用'

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

    await userApi.toggleUserEnable({ username: row.username, enable })
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

// 重置密码（默认重置为 123456）
const handleResetPassword = async (row: UserVO) => {
  try {
    await ElMessageBox.confirm(
      `确定要将用户 "${row.username}" 的密码重置为默认值 123456 吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await userApi.resetPassword({ username: row.username })
    ElMessage.success('密码已重置为 123456')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('重置密码失败')
    }
  }
}

// 删除用户
const deleteDialogVisible = ref(false)
const deleteLoading = ref(false)
const deleteItemName = ref('')
const deleteUserId = ref('')

const handleDelete = (row: UserVO) => {
  deleteItemName.value = row.username
  deleteUserId.value = row.username
  deleteDialogVisible.value = true
}

const confirmDeleteUser = async () => {
  if (!deleteUserId.value) return
  deleteLoading.value = true
  try {
    await userApi.deleteUser(deleteUserId.value)
    ElMessage.success('删除成功')
    deleteDialogVisible.value = false
    fetchUserList()
  } catch (error: any) {
    ElMessage.error('删除失败')
  } finally {
    deleteLoading.value = false
    deleteUserId.value = ''
    deleteItemName.value = ''
  }
}

const cancelDeleteUser = () => {
  deleteDialogVisible.value = false
  deleteUserId.value = ''
  deleteItemName.value = ''
}

// 批量删除
const batchDeleteDialogVisible = ref(false)
const batchDeleteLoading = ref(false)

const handleBatchDelete = () => {
  if (selectedUsers.value.length === 0) {
    ElMessage.warning('请选择要删除的用户')
    return
  }
  batchDeleteDialogVisible.value = true
}

const confirmBatchDeleteUsers = async () => {
  if (selectedUsers.value.length === 0) return
  batchDeleteLoading.value = true
  try {
    const ids = selectedUsers.value.map(user => user.username)
    await userApi.batchDeleteUsers(ids)
    ElMessage.success('批量删除成功')
    selectedUsers.value = []
    batchDeleteDialogVisible.value = false
    fetchUserList()
  } catch (error: any) {
    ElMessage.error('批量删除失败')
  } finally {
    batchDeleteLoading.value = false
  }
}

const cancelBatchDeleteUsers = () => {
  batchDeleteDialogVisible.value = false
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    submitLoading.value = true
    
    if (isEdit.value) {
      const updateData: UpdateUserDTO = {
        username: userForm.username,
        realName: userForm.realName || userForm.username,
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
        realName: userForm.realName || userForm.username,
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
  --page-primary: var(--el-color-primary);
  --page-success: var(--el-color-success);
  --page-danger: var(--el-color-danger);
  --page-warning: var(--el-color-warning);
  --page-text-main: #1e293b;
  --page-text-sub: #475569;
  --page-border: #e2e8f0;
  --page-soft-bg: #f8fafc;
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

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 20px;
}

.stat-card {
  border: none;
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.06);
}

.stat-label {
  font-size: 12px;
  color: #64748b;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 24px;
  line-height: 1;
  font-weight: 700;
  color: #1e293b;
}

.text-success { color: var(--page-success); }
.text-danger { color: var(--page-danger); }
.text-warning { color: var(--page-warning); }

.filter-card {
  margin-bottom: 20px;
  border: none;
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.06);
}

.filter-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.active-filters {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  padding: 10px 12px;
  border-radius: 8px;
  background: var(--page-soft-bg);
  border: 1px solid var(--page-border);
}

.active-filters-label {
  font-size: 13px;
  color: var(--page-text-sub);
  font-weight: 500;
}

.clear-filter-btn {
  margin-left: auto;
  font-weight: 600;
  white-space: nowrap;
  color: var(--page-primary) !important;
  background: transparent !important;
  border: none !important;
  padding: 4px 6px !important;
}

.clear-filter-btn:hover,
.clear-filter-btn:focus {
  color: #2563eb !important;
  background: rgba(59, 130, 246, 0.1) !important;
  border-radius: 6px;
}

.filter-fade-enter-active,
.filter-fade-leave-active {
  transition: all 0.2s ease;
}

.filter-fade-enter-from,
.filter-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
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
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.06);
}

.table-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 16px;
  border-bottom: 1px solid #ebeef5;
}

:deep(.filter-card .el-card__body),
:deep(.table-card .el-card__body) {
  padding: 18px 20px;
}

:deep(.user-table .el-table__header th) {
  background: #f8fafc;
  color: #334155;
}

:deep(.user-table .el-table__row:hover > td) {
  background-color: #f8fbff;
}

:deep(.user-table .el-table__row.row-selected > td:first-child) {
  border-left: 3px solid var(--page-primary);
}

:deep(.user-table .action-buttons .el-button) {
  opacity: 0.85;
  transition: all 0.18s ease;
}

:deep(.user-table .el-table__row:hover .action-buttons .el-button) {
  opacity: 1;
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
  background-color: rgba(59, 130, 246, 0.1);
  color: var(--page-primary);
  border: 1px solid rgba(59, 130, 246, 0.25);
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-table {
  margin-bottom: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  min-width: 0;
}

.user-name {
  font-weight: 500;
  color: #2c3e50;
}

.status-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.status-icon {
  font-size: 12px;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: nowrap;
  justify-content: center;
}

.btn-edit {
  font-weight: 500;
}

.btn-toggle {
  border-color: rgba(59, 130, 246, 0.25);
}

.btn-more {
  color: var(--page-text-sub);
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

  .stats-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
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

  .active-filters {
    gap: 6px;
  }

  .clear-filter-btn {
    margin-left: 0;
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
    justify-content: flex-start;
  }
}
</style>
