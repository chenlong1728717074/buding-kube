<template>
  <div class="simple-layout">
    <!-- 顶部导航 -->
    <header class="header">
      <div class="header-left">
        <h1 class="logo" @click="goHome">Buding K8s</h1>
      </div>
      <div class="header-right">
        <div class="header-extras">
          <div class="clock">
            <el-icon class="clock-icon"><Timer /></el-icon>
            <span class="clock-time">{{ nowTimeText }}</span>
            <span class="clock-split">·</span>
            <span class="clock-date">{{ nowDateText }}</span>
          </div>
        </div>
        <el-dropdown @command="handleCommand">
          <span class="user-info">
            <el-icon><Avatar /></el-icon>
            {{ userStore.userInfo?.username }}
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人信息</el-dropdown-item>
              <el-dropdown-item command="userManagement" v-if="isAdmin">用户管理</el-dropdown-item>
              <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <!-- 主内容区 -->
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { Avatar, ArrowDown, Timer } from '@element-plus/icons-vue'
import { ElMessageBox, ElNotification } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()
const nowDateText = ref('')
const nowTimeText = ref('')

// 检查是否是管理员 - 使用store中的isAdmin
const isAdmin = computed(() => userStore.isAdmin)

// 返回首页
const goHome = () => {
  router.push('/home')
}

const handleCommand = (command: string) => {
  switch (command) {
    case 'profile':
      router.push('/user/profile')
      break
    case 'userManagement':
      router.push('/user/management')
      break
    case 'logout':
      ElMessageBox.confirm('确定要退出登录吗？', '提示', {
        type: 'warning',
        confirmButtonText: '确认',
        cancelButtonText: '取消'
      })
        .then(async () => {
          await userStore.logout()
          ElNotification({
            message: '期待你下次回来',
            type: 'success',
            duration: 2500
          })
        })
        .catch(() => {})
      break
  }
}

let clockTimer: number | undefined
const updateClock = () => {
  const d = new Date()
  nowDateText.value = d.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    weekday: 'short'
  })
  nowTimeText.value = d.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

onMounted(() => {
  updateClock()
  clockTimer = window.setInterval(updateClock, 1000)
})

onUnmounted(() => {
  if (clockTimer) window.clearInterval(clockTimer)
})
</script>

<style scoped>
.simple-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #f5faff 0%, #fbfdff 100%);
}

.header {
  background: #ffffff;
  border-bottom: 1px solid rgba(59, 130, 246, 0.16);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.04);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 32px;
  height: 64px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
}

.logo {
  font-size: 20px;
  font-weight: 600;
  color: #1e3a8a;
  margin: 0;
  cursor: pointer;
  transition: color 0.2s ease;
}

.logo:hover {
  color: #3b82f6;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-extras {
  display: flex;
  align-items: center;
  gap: 12px;
}

.clock {
  display: inline-flex;
  align-items: center;
  height: 34px;
  padding: 0 12px;
  border-radius: 10px;
  background: rgba(15, 23, 42, 0.04);
  border: 1px solid rgba(15, 23, 42, 0.08);
  color: #0f172a;
  gap: 8px;
}

.clock-icon {
  color: rgba(15, 23, 42, 0.65);
}

.clock-time {
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.2px;
  font-variant-numeric: tabular-nums;
}

.clock-split {
  color: rgba(15, 23, 42, 0.35);
  font-size: 12px;
}

.clock-date {
  font-size: 12px;
  color: rgba(15, 23, 42, 0.65);
}

.user-info {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
  color: #1e3a8a;
  font-size: 14px;
  height: 34px;
  padding: 0 12px;
  border-radius: 10px;
  background: rgba(15, 23, 42, 0.04);
  border: 1px solid rgba(15, 23, 42, 0.08);
  gap: 6px;
}

.user-info:hover {
  background: rgba(15, 23, 42, 0.06);
  color: #0f172a;
}

.main-content {
  flex: 1;
  overflow-y: auto;
}

@media (max-width: 768px) {
  .header {
    padding: 0 16px;
  }

  .clock-date,
  .clock-split {
    display: none;
  }
}
</style>
