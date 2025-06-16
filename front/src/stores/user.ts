import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { loginApi } from '@/api/login'
import { userApi, type UserVO, UserRole } from '@/api/user'
import type { LoginDTO } from '@/api/login'
import { ElMessage } from 'element-plus'
import router from '@/router'

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref<string>('')
  const userInfo = ref<UserVO | null>(null)
  const loading = ref<boolean>(false)

  // 计算属性
  const isLoggedIn = computed(() => {
    return !!token.value && !!userInfo.value
  })

  const isSuperAdmin = computed(() => {
    return userInfo.value?.role === UserRole.SUPER_ADMIN
  })

  const isAdmin = computed(() => {
    return userInfo.value?.role === UserRole.ADMIN || isSuperAdmin.value
  })

  const canManageUser = computed(() => {
    return isAdmin.value
  })

  const canDeleteUser = computed(() => {
    return isSuperAdmin.value
  })

  // 用户登录
  const login = async (loginData: LoginDTO): Promise<boolean> => {
    try {
      loading.value = true
      const response = await loginApi.login(loginData)
      
      if (response.code === 200 && response.data) {
        const userData = response.data
        
        // 保存token和用户信息
        token.value = userData.token || ''
        userInfo.value = userData
        
        // 持久化到localStorage
        localStorage.setItem('token', token.value)
        localStorage.setItem('userInfo', JSON.stringify(userData))
        
        ElMessage.success('登录成功')
        return true
      } else {
        ElMessage.error(response.msg || '登录失败')
        return false
      }
    } catch (error: any) {
      console.error('登录失败:', error)
      ElMessage.error(error.message || '登录失败，请检查网络连接')
      return false
    } finally {
      loading.value = false
    }
  }

  // 用户登出
  const logout = async (): Promise<void> => {
    try {
      // 调用后端登出接口
      await loginApi.logout()
    } catch (error) {
      console.error('登出接口调用失败:', error)
    } finally {
      // 无论接口是否成功，都清除本地数据
      clearUserData()
      router.push('/login')
      ElMessage.success('已退出登录')
    }
  }

  // 清除用户数据
  const clearUserData = (): void => {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
  }

  // 初始化用户信息（从localStorage恢复）
  const initUserInfo = (): void => {
    try {
      const savedToken = localStorage.getItem('token')
      const savedUserInfo = localStorage.getItem('userInfo')
      
      if (savedToken && savedUserInfo) {
        token.value = savedToken
        userInfo.value = JSON.parse(savedUserInfo)
      }
    } catch (error) {
      console.error('初始化用户信息失败:', error)
      clearUserData()
    }
  }

  // 获取当前用户信息（从服务器刷新）
  const getCurrentUser = async (): Promise<void> => {
    try {
      if (!token.value) return
      
      const response = await loginApi.getCurrentUser()
      if (response.code === 200 && response.data) {
        userInfo.value = response.data
        localStorage.setItem('userInfo', JSON.stringify(response.data))
      }
    } catch (error) {
      console.error('获取用户信息失败:', error)
      // 如果获取用户信息失败，可能token已过期，清除本地数据
      clearUserData()
    }
  }

  // 刷新token
  const refreshToken = async (): Promise<boolean> => {
    try {
      const response = await loginApi.refreshToken()
      if (response.code === 200 && response.data?.token) {
        token.value = response.data.token
        localStorage.setItem('token', response.data.token)
        return true
      }
      return false
    } catch (error) {
      console.error('刷新token失败:', error)
      clearUserData()
      return false
    }
  }

  return {
    // 状态
    token,
    userInfo,
    loading,
    
    // 计算属性
    isLoggedIn,
    isSuperAdmin,
    isAdmin,
    canManageUser,
    canDeleteUser,
    
    // 方法
    login,
    logout,
    clearUserData,
    initUserInfo,
    getCurrentUser,
    refreshToken
  }
})
