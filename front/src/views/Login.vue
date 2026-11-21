<template>
  <div class="login-page">
    <!-- 背景装饰波浪 -->
    <svg class="wave wave-left" viewBox="0 0 900 600" preserveAspectRatio="none">
      <path d="M0,480 C220,380 420,520 640,440 C760,396 850,420 900,380 L900,600 L0,600 Z" fill="#d8e8fb" opacity="0.5" />
      <path d="M0,540 C240,460 460,580 700,500 C800,468 870,480 900,450 L900,600 L0,600 Z" fill="#cfe0f8" opacity="0.45" />
    </svg>
    <svg class="wave wave-right" viewBox="0 0 900 600" preserveAspectRatio="none">
      <path d="M0,300 C180,240 380,340 560,280 C720,226 830,260 900,220 L900,0 L0,0 Z" fill="#e3eefc" opacity="0.55" />
      <path d="M0,180 C200,120 400,220 620,160 C760,120 850,150 900,110 L900,0 L0,0 Z" fill="#d5e5fa" opacity="0.4" />
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
import heroImg from '@/assets/images/login3d.jpg'

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
/* 版式按原型 1668x943 实测比例复现, 关键尺寸使用 vw/vh 保证同比例缩放 */
.login-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: linear-gradient(118deg, #f0f6fe 0%, #e9f2fd 40%, #dcebfc 100%);
}

/* 背景波浪: z-index:-1 沉到内容之下(与渐变背景融合, 且不遮卡片) */
.wave {
  position: absolute;
  z-index: -1;
  width: 54%;
  height: 58%;
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

/* 顶部品牌: 原型 x≈55 y≈50, logo 44px */
.brand-header {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 3.4vh 3.3vw 0;
}

.brand-logo {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: #ffffff;
  box-shadow: 0 4px 14px rgba(59, 130, 246, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
}

.brand-logo img {
  width: 34px;
  height: 34px;
}

.brand-name {
  font-size: 20px;
  font-weight: 700;
  color: #1e3a5f;
  letter-spacing: 0.5px;
}

/* 主体: 不创建堆叠上下文, 保证 3D 图 multiply 混合能与页面渐变背景融合 */
.login-main {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 3.5vw;
  padding: 0 9vw 1vh 7vw;
}

/* 左侧介绍: 原型标题 x≈210(12.6%), 大图 x≈115(6.9%) 宽≈755(45vw) */
.intro {
  flex: 1 1 auto;
  min-width: 0;
}

.intro-title {
  margin: 0 0 0 5.2vw;
  font-size: clamp(30px, 2.75vw, 46px);
  line-height: 1.42;
  font-weight: 700;
  color: #16325c;
  letter-spacing: 1px;
}

.intro-desc {
  margin: 1.8vh 0 0 5.2vw;
  font-size: clamp(13px, 1vw, 16px);
  color: #5b7396;
}

.intro-bar {
  width: 44px;
  height: 4px;
  border-radius: 2px;
  background: linear-gradient(90deg, #3b82f6, #60a5fa);
  margin: 2.2vh 0 0 5.2vw;
}

.intro-hero {
  display: block;
  width: 44vw;
  max-width: 780px;
  margin: -1vh 0 0 -1.8vw;
  mix-blend-mode: multiply;
  animation: hero-float 6s ease-in-out infinite;
}

@keyframes hero-float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

/* 登录卡片: 原型宽≈505px(30.3vw), 距右≈14vw */
.login-card {
  width: clamp(400px, 30.5vw, 520px);
  flex-shrink: 0;
  background: #ffffff;
  border-radius: 18px;
  box-shadow: 0 20px 60px rgba(59, 130, 246, 0.16);
  padding: 4.8vh 2.9vw 3.4vh;
}

.card-title {
  margin: 0 0 3.6vh;
  text-align: center;
  font-size: clamp(24px, 1.7vw, 28px);
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
  height: 44px;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 2.6vh;
}

.login-btn {
  width: 100%;
  height: 52px;
  margin-top: 1.4vh;
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
  margin: 3.4vh 0 2.2vh;
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
  margin-top: 1vh;
  font-size: clamp(15px, 1.1vw, 18px);
  font-weight: 700;
  color: #3b82f6;
  cursor: pointer;
  user-select: none;
}

.demo-value:hover {
  text-decoration: underline;
}

.demo-notice {
  margin-top: 1.4vh;
  font-size: 12px;
  color: #b6c2d4;
}

/* 底部 */
.page-footer {
  position: relative;
  z-index: 2;
  text-align: center;
  padding: 2vh 0 2.4vh;
  font-size: 13px;
  color: #8ba3c7;
}

/* 响应式: 窄屏隐藏大图, 卡片自适应 */
@media (max-width: 1080px) {
  .login-main {
    flex-direction: column;
    gap: 32px;
    padding: 16px 24px 32px;
  }

  .intro {
    text-align: center;
  }

  .intro-title,
  .intro-desc,
  .intro-bar {
    margin-left: 0;
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
    max-width: 480px;
  }

  .brand-header {
    padding: 20px 24px 0;
  }
}
</style>
