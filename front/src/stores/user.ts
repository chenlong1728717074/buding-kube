import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { loginApi } from '@/api/login'
import { type UserVO, UserRole } from '@/api/user'
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
    const role = userInfo.value?.role as any
    return role === UserRole.SUPER_ADMIN || role === 1
  })

  const isAdmin = computed(() => {
    const role = userInfo.value?.role as any
    return role === UserRole.ADMIN || role === 2 || isSuperAdmin.value
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
        // 登录接口返回的是 LoginVO(数字角色), 与用户列表的 UserVO(字符串角色)兼容存取
        userInfo.value = userData as unknown as UserVO
        sessionStorage.setItem('token', token.value)
        sessionStorage.setItem('userInfo', JSON.stringify(userData))
        localStorage.removeItem('token')
        localStorage.removeItem('userInfo')
        
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
      clearUserData()
      router.push('/login')
    }
  }

  // 清除用户数据
  const clearUserData = (): void => {
    token.value = ''
    userInfo.value = null
    sessionStorage.removeItem('token')
    sessionStorage.removeItem('userInfo')
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
  }

  // 初始化用户信息（从sessionStorage恢复，兼容迁移localStorage）
  const initUserInfo = (): void => {
    try {
      const savedToken = sessionStorage.getItem('token') || localStorage.getItem('token')
      const savedUserInfo = sessionStorage.getItem('userInfo') || localStorage.getItem('userInfo')
      
      if (savedToken && savedUserInfo) {
        token.value = savedToken
        userInfo.value = JSON.parse(savedUserInfo)
        sessionStorage.setItem('token', savedToken)
        sessionStorage.setItem('userInfo', savedUserInfo)
        localStorage.removeItem('token')
        localStorage.removeItem('userInfo')
      }
    } catch (error) {
      console.error('初始化用户信息失败:', error)
      clearUserData()
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

    // 方法
    login,
    logout,
    clearUserData,
    initUserInfo
  }
})
