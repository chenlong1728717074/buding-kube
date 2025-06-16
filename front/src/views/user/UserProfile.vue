<template>
  <div class="user-profile">
    <div class="page-header">
      <h1>个人资料</h1>
    </div>

    <div class="profile-content">
      <!-- 个人信息卡片 -->
      <el-card class="profile-card">
        <template #header>
          <div class="card-header">
            <span>个人信息</span>
            <el-button v-if="!editMode" @click="handleEdit">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <div v-else class="edit-actions">
              <el-button @click="handleCancel">取消</el-button>
              <el-button type="primary" @click="handleSave" :loading="saveLoading">
                保存
              </el-button>
            </div>
          </div>
        </template>

        <div class="profile-info">
          <!-- 头像部分 -->
          <div class="avatar-section">
            <el-upload
              v-if="editMode"
              class="avatar-uploader"
              :action="uploadUrl"
              :headers="uploadHeaders"
              :show-file-list="false"
              :on-success="handleAvatarSuccess"
              :before-upload="beforeAvatarUpload"
            >
              <el-avatar :size="120" :src="userInfo.avatar" :icon="UserFilled" />
              <div class="avatar-overlay">
                <el-icon><Camera /></el-icon>
                <span>更换头像</span>
              </div>
            </el-upload>
            <el-avatar v-else :size="120" :src="userInfo.avatar" :icon="UserFilled" />
          </div>

          <!-- 基本信息 -->
          <div class="info-section">
            <el-form 
              ref="formRef" 
              :model="userInfo" 
              :rules="formRules" 
              label-width="100px"
              :disabled="!editMode"
              class="centered-form"
            >
              <el-row :gutter="24">
                <el-col :span="12">
                  <el-form-item label="用户名">
                    <el-input v-model="userInfo.username" disabled />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="邮箱" prop="email">
                    <el-input v-model="userInfo.email" />
                  </el-form-item>
                </el-col>
              </el-row>
              <el-row :gutter="24">
                <el-col :span="12">
                  <el-form-item label="真实姓名" prop="realName">
                    <el-input v-model="userInfo.realName" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="手机号" prop="phone">
                    <el-input v-model="userInfo.phone" />
                  </el-form-item>
                </el-col>
              </el-row>
              <el-row :gutter="24">
                <el-col :span="12">
                  <el-form-item label="部门">
                    <el-input v-model="userInfo.department" />
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="职位">
                    <el-input v-model="userInfo.position" />
                  </el-form-item>
                </el-col>
              </el-row>
              <el-row :gutter="24">
                <el-col :span="12">
                  <el-form-item label="角色">
                    <el-tag :type="getRoleType(userInfo.role)">
                      {{ getRoleText(userInfo.role) }}
                    </el-tag>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="状态">
                    <el-tag :type="getStatusType(userInfo.status)">
                      {{ getStatusText(userInfo.status) }}
                    </el-tag>
                  </el-form-item>
                </el-col>
              </el-row>
              <el-form-item label="个人简介">
                <el-input 
                  v-model="userInfo.bio" 
                  type="textarea" 
                  :rows="3" 
                  placeholder="请输入个人简介"
                />
              </el-form-item>
            </el-form>
          </div>
        </div>
      </el-card>

      <!-- 账户安全 -->
      <el-card class="security-card">
        <template #header>
          <span>账户安全</span>
        </template>
        
        <div class="security-items">
          <div class="security-item">
            <div class="security-info">
              <div class="security-title">
                <el-icon><Lock /></el-icon>
                登录密码
              </div>
              <div class="security-desc">定期更换密码可以提高账户安全性</div>
            </div>
            <el-button @click="handleChangePassword">修改密码</el-button>
          </div>
          
          <div class="security-item">
            <div class="security-info">
              <div class="security-title">
                <el-icon><Message /></el-icon>
                邮箱验证
              </div>
              <div class="security-desc">
                当前邮箱：{{ userInfo.email }}
                <el-tag v-if="userInfo.emailVerified" type="success" size="small">已验证</el-tag>
                <el-tag v-else type="warning" size="small">未验证</el-tag>
              </div>
            </div>
            <el-button v-if="!userInfo.emailVerified" @click="handleVerifyEmail">
              验证邮箱
            </el-button>
          </div>
          
          <div class="security-item">
            <div class="security-info">
              <div class="security-title">
                <el-icon><Phone /></el-icon>
                手机绑定
              </div>
              <div class="security-desc">
                当前手机：{{ userInfo.phone || '未绑定' }}
                <el-tag v-if="userInfo.phoneVerified" type="success" size="small">已验证</el-tag>
                <el-tag v-else-if="userInfo.phone" type="warning" size="small">未验证</el-tag>
              </div>
            </div>
            <el-button @click="handleBindPhone">
              {{ userInfo.phone ? '更换手机' : '绑定手机' }}
            </el-button>
          </div>
        </div>
      </el-card>

      <!-- 登录历史 -->
      <el-card class="history-card">
        <div class="empty-state">
          <el-empty description="登录历史功能暂未开放" />
        </div>
      </el-card>
    </div>

    <!-- 修改密码对话框 -->
    <el-dialog 
      v-model="passwordDialogVisible" 
      title="修改密码" 
      width="400px"
    >
      <el-form 
        ref="passwordFormRef" 
        :model="passwordForm" 
        :rules="passwordRules" 
        label-width="100px"
      >
        <el-form-item label="当前密码" prop="currentPassword">
          <el-input 
            v-model="passwordForm.currentPassword" 
            type="password" 
            placeholder="请输入当前密码" 
            show-password
          />
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input 
            v-model="passwordForm.newPassword" 
            type="password" 
            placeholder="请输入新密码" 
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input 
            v-model="passwordForm.confirmPassword" 
            type="password" 
            placeholder="请再次输入新密码" 
            show-password
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="passwordDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handlePasswordSubmit" :loading="passwordLoading">
            确定
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  Edit, 
  UserFilled, 
  Camera, 
  Lock, 
  Message, 
  Phone, 
  Refresh
} from '@element-plus/icons-vue'
import { userApi } from '@/api/user'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const editMode = ref(false)
const saveLoading = ref(false)

const passwordLoading = ref(false)
const formRef = ref()
const passwordFormRef = ref()

// 用户信息
const userInfo = reactive({
  id: '',
  username: '',
  email: '',
  realName: '',
  phone: '',
  department: '',
  position: '',
  role: '',
  status: '',
  avatar: '',
  bio: '',
  emailVerified: false,
  phoneVerified: false,

})

// 原始用户信息（用于取消编辑时恢复）
const originalUserInfo = reactive({})

// 表单验证规则
const formRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  realName: [
    { required: true, message: '请输入真实姓名', trigger: 'blur' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号', trigger: 'blur' }
  ]
}



// 修改密码对话框
const passwordDialogVisible = ref(false)
const passwordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: ''
})

// 密码验证规则
const passwordRules = {
  currentPassword: [
    { required: true, message: '请输入当前密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, max: 20, message: '密码长度在 6 到 20 个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (rule: any, value: string, callback: Function) => {
        if (value !== passwordForm.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

// 头像上传
const uploadUrl = '/api/upload/avatar'
const uploadHeaders = {
  Authorization: `Bearer ${userStore.token}`
}

// 获取用户信息
const fetchUserInfo = async () => {
  ElMessage.info('个人信息功能暂未开放')
  // 使用当前登录用户的基本信息
  const currentUser = userStore.userInfo
  if (currentUser) {
    Object.assign(userInfo, {
      id: currentUser.id || '',
      username: currentUser.username || '',
      email: currentUser.email || '',
      role: currentUser.role || '',
      status: currentUser.status || '',
      department: currentUser.department || '',
      phone: currentUser.phone || ''
    })
    Object.assign(originalUserInfo, userInfo)
  }
}



// 编辑模式
const handleEdit = () => {
  editMode.value = true
}

// 取消编辑
const handleCancel = () => {
  editMode.value = false
  Object.assign(userInfo, originalUserInfo)
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

// 保存
const handleSave = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    saveLoading.value = true
    
    const res = await userApi.updateProfile(userInfo)
    if (res.code === 200) {
      ElMessage.success('保存成功')
      editMode.value = false
      Object.assign(originalUserInfo, userInfo)
      // 更新store中的用户信息
      userStore.updateUserInfo(userInfo)
    }
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    saveLoading.value = false
  }
}

// 头像上传成功
const handleAvatarSuccess = (response: any) => {
  if (response.code === 200) {
    userInfo.avatar = response.data.url
    ElMessage.success('头像上传成功')
  } else {
    ElMessage.error('头像上传失败')
  }
}

// 头像上传前验证
const beforeAvatarUpload = (file: File) => {
  const isJPG = file.type === 'image/jpeg' || file.type === 'image/png'
  const isLt2M = file.size / 1024 / 1024 < 2

  if (!isJPG) {
    ElMessage.error('头像只能是 JPG/PNG 格式!')
  }
  if (!isLt2M) {
    ElMessage.error('头像大小不能超过 2MB!')
  }
  return isJPG && isLt2M
}

// 修改密码
const handleChangePassword = () => {
  Object.assign(passwordForm, {
    currentPassword: '',
    newPassword: '',
    confirmPassword: ''
  })
  passwordDialogVisible.value = true
}

// 密码提交
const handlePasswordSubmit = async () => {
  if (!passwordFormRef.value) return
  
  try {
    await passwordFormRef.value.validate()
    passwordLoading.value = true
    
    const res = await userApi.changePassword({
      currentPassword: passwordForm.currentPassword,
      newPassword: passwordForm.newPassword
    })
    
    if (res.code === 200) {
      ElMessage.success('密码修改成功')
      passwordDialogVisible.value = false
    }
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    passwordLoading.value = false
  }
}

// 验证邮箱
const handleVerifyEmail = async () => {
  try {
    const res = await userApi.sendEmailVerification()
    if (res.code === 200) {
      ElMessage.success('验证邮件已发送，请查收')
    }
  } catch (error) {
    ElMessage.error('发送验证邮件失败')
  }
}

// 绑定手机
const handleBindPhone = () => {
  ElMessage.info('手机绑定功能开发中')
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
    active: '正常',
    inactive: '禁用'
  }
  return statusMap[status] || status
}

// 初始化
onMounted(() => {
  fetchUserInfo()

})
</script>

<style scoped>
.user-profile {
  padding: 20px;
}

.page-header h1 {
  margin: 0 0 20px 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.profile-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.edit-actions {
  display: flex;
  gap: 8px;
}

.profile-info {
  display: flex;
  gap: 40px;
}

.avatar-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 140px;
}

.avatar-uploader {
  position: relative;
  cursor: pointer;
}

.avatar-uploader:hover .avatar-overlay {
  opacity: 1;
}

.avatar-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: white;
  opacity: 0;
  transition: opacity 0.3s;
  font-size: 12px;
}

.info-section {
  flex: 1;
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
  
  .el-tag {
    margin: 0 auto;
    display: block;
    width: fit-content;
  }
}

.security-items {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.security-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
}

.security-info {
  flex: 1;
}

.security-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: #2c3e50;
  margin-bottom: 4px;
}

.security-desc {
  color: #7f8c8d;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
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
  .profile-info {
    flex-direction: column;
    gap: 20px;
  }
  
  .security-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
}
</style>