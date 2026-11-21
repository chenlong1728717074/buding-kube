<template>
  <div class="login-page">
    <!-- 背景装饰波浪 -->
    <svg class="wave wave-left" viewBox="0 0 900 600" preserveAspectRatio="none">
      <path d="M0,480 C220,380 420,520 640,440 C760,396 850,420 900,380 L900,600 L0,600 Z" fill="#d8e8fb" opacity="0.55" />
      <path d="M0,540 C240,460 460,580 700,500 C800,468 870,480 900,450 L900,600 L0,600 Z" fill="#cfe0f8" opacity="0.5" />
    </svg>
    <svg class="wave wave-right" viewBox="0 0 900 600" preserveAspectRatio="none">
      <path d="M0,300 C180,240 380,340 560,280 C720,226 830,260 900,220 L900,0 L0,0 Z" fill="#e3eefc" opacity="0.6" />
      <path d="M0,180 C200,120 400,220 620,160 C760,120 850,150 900,110 L900,0 L0,0 Z" fill="#d5e5fa" opacity="0.45" />
    </svg>

    <!-- 顶部品牌 -->
    <header class="brand-header">
      <div class="brand-logo">
        <img :src="logoK8s" alt="Kubernetes" />
      </div>
      <span class="brand-name">BuDing Kube</span>
    </header>

    <!-- 主体 -->
    <main class="login-main">
      <!-- 左侧介绍 -->
      <section class="intro">
        <h1 class="intro-title">BuDing Kube<br />多集群管理平台</h1>
        <p class="intro-desc">统一管理多个 Kubernetes 集群，简化运维操作</p>
        <div class="intro-bar"></div>
        <img class="intro-hero" :src="heroImg" alt="多集群管理" />
      </section>

      <!-- 登录卡片 -->
      <section class="login-card">
        <h2 class="card-title">登录</h2>

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
              placeholder="请输入用户名"
              :prefix-icon="User"
              size="large"
              autocomplete="username"
            />
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="请输入密码"
              :prefix-icon="Lock"
              size="large"
              show-password
              autocomplete="current-password"
            />
          </el-form-item>

          <el-button
            type="primary"
            :loading="loading"
            size="large"
            class="login-btn"
            @click="onSubmit"
          >
            <el-icon v-if="!loading" class="login-btn-icon"><SwitchButton /></el-icon>
            <span>{{ loading ? '登录中...' : '登录' }}</span>
          </el-button>
        </el-form>

        <el-divider class="card-divider" />

        <div class="demo-section">
          <div class="demo-label">演示账号</div>
          <div class="demo-value" @click="fillDemo">admin / 123456</div>
          <div class="demo-notice">演示环境部分功能未开放</div>
        </div>
      </section>
    </main>

    <!-- 底部 -->
    <footer class="page-footer">© 2026 BuDing · 统一管理多个 Kubernetes 集群</footer>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElNotification } from 'element-plus'
import { User, Lock, SwitchButton } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import logoK8s from '@/assets/logo-k8s.svg'
import heroImg from '@/assets/images/modern-soft-3d-isometric.jpg'

const router = useRouter()
const userStore = useUserStore()
const loginForm = ref({ username: '', password: '' })
const loading = ref(false)
const formRef = ref()

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

// 一键填充演示账号
const fillDemo = () => {
  loginForm.value.username = 'admin'
  loginForm.value.password = '123456'
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
          message: `${greet} ${emoji} ${loginForm.value.username}，欢迎使用 BuDing Kube 管理平台`,
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
.login-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: linear-gradient(135deg, #f2f7fe 0%, #e8f1fd 45%, #dbeafc 100%);
}

/* 背景波浪 */
.wave {
  position: absolute;
  width: 55%;
  height: 60%;
  pointer-events: none;
}

.wave-left {
  left: -60px;
  bottom: -40px;
}

.wave-right {
  right: -80px;
  top: -60px;
  transform: rotate(180deg);
}

/* 顶部品牌 */
.brand-header {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 26px 56px 0;
}

.brand-logo {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: #ffffff;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.18);
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-logo img {
  width: 30px;
  height: 30px;
}

.brand-name {
  font-size: 20px;
  font-weight: 700;
  color: #1e3a5f;
  letter-spacing: 0.5px;
}

/* 主体 */
.login-main {
  position: relative;
  z-index: 2;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 72px;
  padding: 24px 56px;
}

/* 左侧介绍 */
.intro {
  max-width: 560px;
}

.intro-title {
  margin: 0;
  font-size: 42px;
  line-height: 1.3;
  font-weight: 700;
  color: #16325c;
  letter-spacing: 1px;
}

.intro-desc {
  margin: 14px 0 0;
  font-size: 16px;
  color: #5b7396;
}

.intro-bar {
  width: 44px;
  height: 4px;
  border-radius: 2px;
  background: linear-gradient(90deg, #3b82f6, #60a5fa);
  margin: 18px 0 8px;
}

.intro-hero {
  display: block;
  width: 100%;
  max-width: 560px;
  mix-blend-mode: multiply;
  filter: drop-shadow(0 18px 30px rgba(59, 130, 246, 0.16));
  animation: hero-float 6s ease-in-out infinite;
}

@keyframes hero-float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

/* 登录卡片 */
.login-card {
  width: 440px;
  flex-shrink: 0;
  background: #ffffff;
  border-radius: 20px;
  box-shadow: 0 18px 50px rgba(59, 130, 246, 0.14);
  padding: 44px 44px 32px;
}

.card-title {
  margin: 0 0 28px;
  text-align: center;
  font-size: 26px;
  font-weight: 700;
  color: #1f2937;
}

.login-form :deep(.el-input__wrapper) {
  border-radius: 10px;
  padding: 4px 14px;
  box-shadow: 0 0 0 1px #d9e4f2 inset;
}

.login-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #93b8ee inset;
}

.login-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1.5px #3b82f6 inset;
}

.login-form :deep(.el-input__inner) {
  height: 40px;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 22px;
}

.login-btn {
  width: 100%;
  height: 48px;
  margin-top: 6px;
  border-radius: 12px;
  font-size: 16px;
  letter-spacing: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.login-btn-icon {
  font-size: 17px;
}

.card-divider {
  margin: 26px 0 18px;
}

/* 演示信息 */
.demo-section {
  text-align: center;
}

.demo-label {
  font-size: 13px;
  color: #9ca3af;
}

.demo-value {
  margin-top: 6px;
  font-size: 15px;
  font-weight: 700;
  color: #3b82f6;
  cursor: pointer;
  user-select: none;
}

.demo-value:hover {
  text-decoration: underline;
}

.demo-notice {
  margin-top: 12px;
  font-size: 12px;
  color: #b6c2d4;
}

/* 底部 */
.page-footer {
  position: relative;
  z-index: 2;
  text-align: center;
  padding: 18px 0 22px;
  font-size: 13px;
  color: #8ba3c7;
}

/* 响应式: 窄屏隐藏大图, 卡片自适应 */
@media (max-width: 1080px) {
  .login-main {
    flex-direction: column;
    gap: 36px;
    padding: 24px 24px 40px;
  }

  .intro {
    text-align: center;
    max-width: 100%;
  }

  .intro-bar {
    margin-left: auto;
    margin-right: auto;
  }

  .intro-hero {
    display: none;
  }

  .login-card {
    width: 100%;
    max-width: 460px;
  }

  .brand-header {
    padding: 20px 24px 0;
  }
}
</style>
