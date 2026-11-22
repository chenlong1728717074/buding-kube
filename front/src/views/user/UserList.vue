<template>
  <div class="user-list-container list-page">
    <!-- 头部横幅 -->
    <el-card class="hero-card" shadow="never">
      <div class="hero-content">
        <div class="hero-icon-block">
          <IconUsers :size="30" />
        </div>
        <div class="hero-text">
          <h1 class="hero-title">用户管理</h1>
          <p class="hero-desc">管理系统用户账号、角色权限和状态</p>
        </div>
        <div class="hero-decoration">
          <div class="hero-shield">
            <IconShieldLock :size="34" />
          </div>
        </div>
      </div>
    </el-card>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <el-card class="stat-card stat-blue" shadow="never">
        <div class="stat-icon-block icon-blue"><IconUsers :size="28" /></div>
        <div class="stat-info">
          <div class="stat-label">总用户数</div>
          <div class="stat-value">{{ pagination.total }}</div>
        </div>
      </el-card>
      <el-card class="stat-card stat-green" shadow="never">
        <div class="stat-icon-block icon-green"><IconUserCheck :size="28" /></div>
        <div class="stat-info">
          <div class="stat-label">正常用户</div>
          <div class="stat-value">{{ activeCount }}</div>
        </div>
      </el-card>
      <el-card class="stat-card stat-red" shadow="never">
        <div class="stat-icon-block icon-red"><IconUserX :size="28" /></div>
        <div class="stat-info">
          <div class="stat-label">禁用用户</div>
          <div class="stat-value">{{ inactiveCount }}</div>
        </div>
      </el-card>
      <el-card class="stat-card stat-purple" shadow="never">
        <div class="stat-icon-block icon-purple"><IconShieldCheck :size="28" /></div>
        <div class="stat-info">
          <div class="stat-label">管理员</div>
          <div class="stat-value">{{ adminCount }}</div>
        </div>
      </el-card>
    </div>

    <!-- 搜索与批量操作工具栏 -->
    <el-card class="filter-card" shadow="never">
      <div class="filter-content">
        <div class="filter-left">
          <el-input
            v-model="searchForm.keyword"
            placeholder="请输入用户名或邮箱"
            :prefix-icon="IconSearch"
            clearable
            style="width: 220px"
            @keyup.enter="handleSearch"
            @clear="handleSearch"
          />
          <el-select
            v-model="searchForm.role"
            placeholder="全部角色"
            clearable
            style="width: 130px"
            @change="handleSearch"
            @clear="handleSearch"
          >
            <el-option label="超级管理员" value="super" />
            <el-option label="管理员" value="admin" />
            <el-option label="普通用户" value="normal" />
          </el-select>
          <el-select
            v-model="searchForm.status"
            placeholder="全部状态"
            clearable
            style="width: 130px"
            @change="handleSearch"
            @clear="handleSearch"
          >
            <el-option label="正常" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </div>
        <div class="filter-right">
          <el-button type="primary" @click="handleAdd">
            <IconPlus :size="16" />
            新增用户
          </el-button>
          <el-button :disabled="selectedUsers.length !== 1" @click="handleToolbarEdit">
            <IconEdit :size="15" />
            编辑
          </el-button>
          <el-button :disabled="selectedUsers.length === 0" @click="handleBatchDisable">
            <IconBan :size="15" />
            禁用
          </el-button>
          <el-button :disabled="selectedUsers.length === 0" @click="handleBatchDelete">
            <IconTrash :size="15" />
            删除
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 用户列表卡片 -->
    <el-card class="table-card" shadow="never">
      
      <!-- 用户表格 -->
      <el-table
        v-loading="loading"
        :data="displayList"
        class="user-table"
        :row-class-name="tableRowClassName"
        @selection-change="handleSelectionChange"
        style="width: 100%"
        :header-cell-class-name="tableHeaderClass"
        :cell-class-name="tableCellClass"
      >
        <el-table-column type="selection" width="50" />

        <template #empty>
          <el-empty description="暂无用户数据">
            <el-button type="primary" @click="handleAdd">去创建用户</el-button>
          </el-empty>
        </template>

        <el-table-column prop="username" label="用户名" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="user-name">{{ row.username }}</span>
          </template>
        </el-table-column>

        <el-table-column prop="realName" label="姓名" min-width="120" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.realName">{{ row.realName }}</span>
            <el-text v-else type="info">未设置</el-text>
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
            <el-tag :type="userStatusType(row.status)" effect="light" class="status-tag">
              <el-icon class="status-icon"><component :is="getStatusIcon(row.status)" /></el-icon>
              <span>{{ userStatusText(row.status) }}</span>
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
            <span v-if="row.createdAt">{{ formatDateTime(row.createdAt) }}</span>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="220" fixed="right" align="center" header-align="center">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" class="btn-edit" @click="handleEdit(row)">
                <IconEdit :size="14" />
                编辑
              </el-button>
              <el-button
                size="small"
                class="btn-toggle"
                :type="isUserEnabled(row) ? 'warning' : 'success'"
                plain
                @click="handleToggleStatus(row)"
              >
                {{ isUserEnabled(row) ? '禁用' : '启用' }}
              </el-button>
              <el-dropdown trigger="click" @command="(command: string) => handleMoreActions(command, row)">
                <el-button size="small" class="more-text-btn">
                  更多<IconArrowDown :size="14" />
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="resetPassword">
                      <IconKey :size="15" />
                      重置密码
                    </el-dropdown-item>
                    <el-dropdown-item command="permissions">
                      <IconLock :size="15" />
                      权限管理
                    </el-dropdown-item>
                    <el-dropdown-item command="delete" divided>
                      <IconTrash :size="15" />
                      <span style="color: var(--el-color-danger);">删除用户</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页器 -->
      <ListPagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.size"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>

    <!-- 新增/编辑用户对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="880px" :close-on-click-modal="false" class="config-dialog form-dialog--compact">
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
            <el-form-item label="姓名" prop="realName">
              <el-input v-model="userForm.realName" placeholder="请输入姓名" maxlength="50" show-word-limit />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="userForm.email" placeholder="请输入邮箱" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="手机号" prop="phone">
              <el-input v-model="userForm.phone" placeholder="请输入手机号" />
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
          <el-col :span="24">
            <el-form-item label="部门" prop="department">
              <el-input v-model="userForm.department" placeholder="请输入部门" />
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
import { userStatusText, userStatusType } from '@/utils/status'
import { formatDateTime } from '@/utils/format'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import ListPagination from '@/components/ListPagination.vue'
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  CircleCloseFilled,
  SuccessFilled,
  WarningFilled,
  InfoFilled
} from '@element-plus/icons-vue'
import {
  IconUsers,
  IconUserCheck,
  IconUserX,
  IconShieldCheck,
  IconShieldLock,
  IconSearch,
  IconPlus,
  IconEdit,
  IconBan,
  IconTrash,
  IconKey,
  IconLock,
  IconArrowDown
} from '@tabler/icons-vue'
import {
  userApi,
  type UserVO,
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
  keyword: '',
  role: undefined as string | undefined,
  status: undefined as string | undefined
})

// 用户列表
const userList = ref<UserVO[]>([])
const selectedUsers = ref<UserVO[]>([])

const activeCount = computed(() => displayList.value.filter(u => u.status === 'active').length)
const inactiveCount = computed(() => displayList.value.filter(u => u.status === 'inactive').length)
const adminCount = computed(() => displayList.value.filter(u => ['super', 'admin', '1', '2'].includes(String(u.role))).length)

// 分页
const pagination = reactive({
  page: 1,
  size: 10,
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
// 获取状态文本
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
    userList.value = res.data?.items || []
    pagination.total = res.data?.total || 0
  } catch {
    // 请求拦截器已提示错误(例如权限不足、登录过期)
    userList.value = []
    pagination.total = 0
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
    keyword: '',
    role: undefined,
    status: undefined
  })
  pagination.page = 1
  fetchUserList()
}

// 关键词过滤(用户名/邮箱, 当前页内模糊匹配)
const displayList = computed(() => {
  const kw = (searchForm.keyword || '').trim().toLowerCase()
  if (!kw) return userList.value
  return userList.value.filter(u =>
    String(u.username || '').toLowerCase().includes(kw) ||
    String(u.email || '').toLowerCase().includes(kw)
  )
})

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
    realName: row.realName || '',
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
  } catch {
    // 请求拦截器已提示错误
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

// 重置密码（弹窗输入新密码, 留空则重置为 123456）
const handleResetPassword = async (row: UserVO) => {
  try {
    const { value } = await ElMessageBox.prompt(
      `请输入用户 "${row.username}" 的新密码，留空则重置为 123456`,
      '重置密码',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputPlaceholder: '留空则为 123456',
        inputType: 'password',
        type: 'warning'
      }
    )

    await userApi.resetPassword({ username: row.username, password: value || '' })
    if (value) {
      ElMessage.success('密码已修改')
    } else {
      ElMessage.success('密码已重置为 123456')
    }
  } catch {
    // 请求拦截器已提示错误
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
  } catch {
    // 请求拦截器已提示错误
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

// 工具栏: 编辑选中用户(仅支持单选)
const handleToolbarEdit = () => {
  if (selectedUsers.value.length !== 1) {
    ElMessage.warning('请选择一个要编辑的用户')
    return
  }
  handleEdit(selectedUsers.value[0])
}

// 工具栏: 批量禁用选中用户
const handleBatchDisable = async () => {
  const targets = selectedUsers.value.filter(u => isUserEnabled(u))
  if (targets.length === 0) {
    ElMessage.warning('选中的用户中没有可禁用的账号')
    return
  }
  try {
    await ElMessageBox.confirm(
      `确定要禁用选中的 ${targets.length} 个用户吗？`,
      '批量禁用',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    for (const u of targets) {
      await userApi.toggleUserEnable({ username: u.username, enable: false })
    }
    ElMessage.success('禁用成功')
    fetchUserList()
  } catch {
    // 请求拦截器已提示错误
  }
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
  } catch {
    // 请求拦截器已提示错误
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

  // 表单校验失败只提示字段错误, 不弹"创建/更新失败"
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  submitLoading.value = true

  try {
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
  } catch {
    // 请求拦截器已提示错误
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

/* 头部横幅: 浅蓝渐变 + 图标块 + 标题 + 右侧盾形装饰 */
.hero-card {
  border: none;
  border-radius: 12px;
  background: linear-gradient(115deg, #d3e6fb 0%, #c7dcf8 55%, #b7d0f5 100%);
  overflow: hidden;
}

.hero-content {
  display: flex;
  align-items: center;
  gap: 20px;
  position: relative;
}

.hero-icon-block {
  width: 68px;
  height: 68px;
  border-radius: 16px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 18px rgba(37, 99, 235, 0.28);
  flex-shrink: 0;
}

.hero-text {
  min-width: 0;
}

.hero-title {
  margin: 0;
  font-size: 26px;
  font-weight: 700;
  color: #0e2a52;
}

.hero-desc {
  margin: 8px 0 0;
  font-size: 14px;
  color: #4f6a94;
}

.hero-decoration {
  margin-left: auto;
  padding-right: 24px;
  display: flex;
}

.hero-shield {
  width: 84px;
  height: 84px;
  border-radius: 22px 22px 40px 40px;
  background: rgba(255, 255, 255, 0.45);
  color: #7ea6e0;
  display: flex;
  align-items: center;
  justify-content: center;
  transform: rotate(8deg);
}

/* 统计卡片: 图标 + 标签 + 数值, 右下角波浪装饰 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.stat-card {
  border: none;
  border-radius: 12px;
  overflow: hidden;
  position: relative;
}

.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 22px;
  position: relative;
}

.stat-card::after {
  content: '';
  position: absolute;
  right: -30px;
  bottom: -34px;
  width: 130px;
  height: 80px;
  border-radius: 50%;
  pointer-events: none;
}

.stat-blue::after { background: rgba(59, 130, 246, 0.08); }
.stat-green::after { background: rgba(16, 185, 129, 0.08); }
.stat-red::after { background: rgba(239, 68, 68, 0.08); }
.stat-purple::after { background: rgba(139, 92, 246, 0.08); }

.stat-icon-block {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.icon-blue { background: rgba(59, 130, 246, 0.12); color: #3b82f6; }
.icon-green { background: rgba(16, 185, 129, 0.12); color: #10b981; }
.icon-red { background: rgba(239, 68, 68, 0.12); color: #ef4444; }
.icon-purple { background: rgba(139, 92, 246, 0.12); color: #8b5cf6; }

.stat-info {
  position: relative;
  z-index: 1;
}

.stat-label {
  font-size: 13px;
  color: #475569;
}

.stat-value {
  margin-top: 6px;
  font-size: 26px;
  line-height: 1;
  font-weight: 700;
  color: #1e293b;
}

.text-success { color: var(--page-success); }
.text-danger { color: var(--page-danger); }
.text-warning { color: var(--page-warning); }

.filter-card {
  margin-bottom: 16px;
  border: none;
  border-radius: 12px;
}

.filter-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.filter-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
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

/* 分页: 总数居左, 页码/每页/跳转居右(与原型一致) */
.pagination-wrapper {
  margin-top: 20px;
}

.pagination-wrapper :deep(.el-pagination) {
  display: flex;
  width: 100%;
  flex-wrap: wrap;
}

.pagination-wrapper :deep(.el-pagination__total) {
  margin-right: auto;
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
