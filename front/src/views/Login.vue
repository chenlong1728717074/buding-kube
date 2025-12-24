<template>
  <div class="login-container">
    <div class="login-background"></div>
    
    <div class="login-content">
      <!-- 顶部品牌 -->
      <div class="brand-header">
        <h1 class="brand-title">Buding k8s多集群管理平台</h1>
        <p class="brand-desc">统一管理多个 Kubernetes 集群，简化运维操作</p>
      </div>

      <!-- 登录卡片 -->
      <div class="login-card">
        <div class="card-title">登录</div>
        
        <el-form 
          :model="loginForm" 
          :rules="rules" 
          ref="formRef" 
          class="login-form"
          @keyup.enter="onSubmit"
        >
          <el-form-item prop="username">
            <el-input
              v-model="loginForm.username"
              placeholder="用户名"
              :prefix-icon="User"
              size="large"
              autocomplete="username"
            />
          </el-form-item>
          
          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="密码"
              :prefix-icon="Lock"
              size="large"
              show-password
              autocomplete="current-password"
            />
          </el-form-item>
          
          <el-form-item class="submit-item">
            <el-button
              type="primary"
              :loading="loading"
              @click="onSubmit"
              size="large"
              class="login-btn"
            >
              {{ loading ? '登录中...' : '登 录' }}
            </el-button>
          </el-form-item>
        </el-form>
        
        <!-- 演示信息 -->
        <div class="demo-section">
          <div class="demo-label">演示账号</div>
          <div class="demo-value">admin / 123456</div>
          <div class="demo-notice">演示环境部分功能未开放</div>
        </div>
      </div>
      
      <!-- 底部 -->
      <div class="footer">© 2024 Buding · 统一管理多个 Kubernetes 集群</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { Monitor, User, Lock } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const loginForm = ref({ username: '', password: '' })
const loading = ref(false)
const formRef = ref()

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const onSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    loading.value = true
    try {
      const success = await userStore.login(loginForm.value)
      if (success) {
        const hour = new Date().getHours()
        const greet = hour < 12 ? '早上好' : hour < 18 ? '下午好' : '晚上好'
        const emoji = hour < 12 ? '☀️' : hour < 18 ? '🌤️' : '🌛'
        ElNotification({
          message: `${greet} ${emoji} ${loginForm.value.username}，欢迎使用 K8s 管理平台`,
          type: 'success',
          duration: 3000
        })
        router.push('/')
      }
    } catch (error) {
      console.error('登录失败:', error)
      ElMessage.error('登录失败')
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.login-background {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(180deg, #f5faff 0%, #fbfdff 100%);
  z-index: 0;
}

.login-content {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 380px;
  padding: 20px;
}

/* 品牌区域 */
.brand-header {
  text-align: center;
  margin-bottom: 32px;
}

.brand-title {
  font-size: 24px;
  font-weight: 600;
  color: #1e3a8a;
  margin: 0 0 8px 0;
}

.brand-desc {
  font-size: 14px;
  color: #64748b;
  margin: 0;
}

/* 登录卡片 */
.login-card {
  background: #ffffff;
  border-radius: 18px;
  padding: 32px 28px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.06);
  border: 1px solid rgba(59, 130, 246, 0.12);
}

.card-title {
  font-size: 20px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 24px;
  text-align: center;
}

/* 表单 */
.login-form :deep(.el-form-item) {
  margin-bottom: 18px;
}

.login-form :deep(.el-input__wrapper) {
  border-radius: 10px;
  padding: 2px 14px;
  box-shadow: 0 0 0 1px #e5e7eb;
}

.login-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #d1d5db;
}

.login-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

.login-form :deep(.el-input__inner) {
  height: 42px;
  font-size: 14px;
}

.login-form :deep(.el-input__prefix) {
  color: #9ca3af;
}

/* 修复浏览器自动填充样式 */
.login-form :deep(.el-input__inner:-webkit-autofill),
.login-form :deep(.el-input__inner:-webkit-autofill:hover),
.login-form :deep(.el-input__inner:-webkit-autofill:focus) {
  -webkit-box-shadow: 0 0 0 1000px #fff inset !important;
  -webkit-text-fill-color: #1f2937 !important;
  transition: background-color 5000s ease-in-out 0s;
}

.submit-item {
  margin-bottom: 0 !important;
  margin-top: 8px;
}

/* 登录按钮 */
.login-btn {
  width: 100%;
  height: 44px;
  font-size: 15px;
  font-weight: 500;
  border-radius: 10px;
  background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);
  border: none;
}

.login-btn:hover,
.login-btn:focus {
  background: linear-gradient(135deg, #2563eb 0%, #3b82f6 100%);
}

/* 演示信息 */
.demo-section {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #f3f4f6;
  text-align: center;
}

.demo-label {
  font-size: 12px;
  color: #9ca3af;
  margin-bottom: 8px;
}

.demo-value {
  font-size: 14px;
  color: #3b82f6;
  font-family: 'SF Mono', Monaco, Consolas, monospace;
  margin-bottom: 8px;
}

.demo-notice {
  font-size: 12px;
  color: #9ca3af;
}

/* 底部 */
.footer {
  margin-top: 32px;
  text-align: center;
  font-size: 12px;
  color: #9ca3af;
}

/* 响应式 */
@media (max-width: 480px) {
  .login-content {
    padding: 16px;
  }

  .login-card {
    padding: 24px 20px;
    border-radius: 16px;
  }

  .brand-title {
    font-size: 20px;
  }

  .brand-desc {
    font-size: 13px;
  }
}
</style>
