<template>
  <div class="login-container">
    <div class="login-background">
      <div class="bg-animation"></div>
    </div>
    <div class="login-content">
      <div class="login-center">
        <div class="brand-header">
          <div class="brand-icon">
            <el-icon size="48"><Monitor /></el-icon>
          </div>
          <h1 class="brand-title">K8s多集群管理平台</h1>
          <p class="brand-desc">统一管理多个Kubernetes集群，简化运维操作</p>
        </div>
        <el-card class="login-card" shadow="always">
          <div class="login-header">
            <h2 class="login-title">欢迎登录</h2>
            <p class="login-subtitle">请输入您的账号信息</p>
          </div>
          <el-form :model="loginForm" :rules="rules" ref="formRef" class="login-form" @keyup.enter="onSubmit">
            <el-form-item prop="username">
              <el-input 
                v-model="loginForm.username" 
                placeholder="请输入用户名"
                prefix-icon="User"
                size="large"
                autocomplete="username" 
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input 
                v-model="loginForm.password" 
                type="password" 
                placeholder="请输入密码"
                prefix-icon="Lock"
                size="large"
                show-password
                autocomplete="current-password" 
              />
            </el-form-item>
            <el-form-item>
              <el-button 
                type="primary" 
                :loading="loading" 
                @click="onSubmit" 
                size="large"
                class="login-btn"
              >
                <span v-if="!loading">立即登录</span>
                <span v-else>登录中...</span>
              </el-button>
            </el-form-item>
          </el-form>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Monitor, Check } from '@element-plus/icons-vue'
import { loginApi } from '@/api/auth'

const router = useRouter()
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
      const res = await loginApi(loginForm.value)
      if (res.code === 200 && res.data?.token) {
        localStorage.setItem('token', res.data.token)
        localStorage.setItem('userInfo', JSON.stringify(res.data))
        ElMessage.success('登录成功')
        router.push('/')
      } else {
        ElMessage.error(res.msg || '登录失败')
      }
    } catch (e) {
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
  overflow: hidden;
}

.login-background {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  z-index: 1;
}

.bg-animation {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><defs><pattern id="grain" width="100" height="100" patternUnits="userSpaceOnUse"><circle cx="50" cy="50" r="1" fill="%23ffffff" opacity="0.1"/></pattern></defs><rect width="100" height="100" fill="url(%23grain)"/></svg>') repeat;
  animation: float 20s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0px); }
  50% { transform: translateY(-20px); }
}

.login-content {
  position: relative;
  z-index: 2;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
}

.login-center {
  display: flex;
  flex-direction: column;
  align-items: center;
  max-width: 480px;
  width: 100%;
}

.brand-header {
  text-align: center;
  color: white;
  margin-bottom: 40px;
}

.brand-icon {
  margin-bottom: 20px;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.05); }
}

.brand-title {
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 12px;
  background: linear-gradient(45deg, #ffffff, #e3f2fd);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.brand-desc {
  font-size: 16px;
  margin: 0;
  opacity: 0.9;
  line-height: 1.6;
}

.login-card {
  width: 100%;
  max-width: 400px;
  padding: 40px;
  border-radius: 16px;
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.2);
  animation: slideInRight 0.8s ease-out;
}

@keyframes slideInRight {
  from {
    opacity: 0;
    transform: translateX(50px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-title {
  font-size: 28px;
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 8px;
}

.login-subtitle {
  color: #7f8c8d;
  font-size: 14px;
  margin: 0;
}

.login-form {
  margin-top: 24px;
}

.login-form .el-form-item {
  margin-bottom: 24px;
}

.login-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  transition: all 0.3s ease;
}

.login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.3);
}

.login-btn:active {
  transform: translateY(0);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .login-content {
    padding: 20px 15px;
  }
  
  .login-center {
    max-width: 100%;
  }
  
  .brand-header {
    margin-bottom: 30px;
  }
  
  .brand-title {
    font-size: 28px;
  }
  
  .brand-desc {
    font-size: 14px;
  }
  
  .login-card {
    padding: 24px;
    margin: 0 10px;
  }
}

@media (max-width: 480px) {
  .brand-title {
    font-size: 24px;
  }
  
  .login-card {
    padding: 20px;
  }
}
</style>