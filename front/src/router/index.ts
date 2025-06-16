import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'

// 路由配置
const routes: Array<RouteRecordRaw> = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: {
      title: '登录',
      public: true
    }
  },
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/views/Layout.vue'),
    redirect: '/dashboard',
    meta: {
      requiresAuth: true
    },
    children: [
      {
        path: '/dashboard',
        name: 'Dashboard',
        component: () => import('@/views/Home.vue'),
        meta: {
          title: '仪表盘',
          requiresAuth: true
        }
      },
      {
        path: '/cluster',
        name: 'ClusterManagement',
        redirect: '/cluster/list',
        meta: {
          title: '集群管理',
          requiresAuth: true
        },
        children: [
          {
            path: 'list',
            name: 'ClusterList',
            component: () => import('@/views/cluster/ClusterList.vue'),
            meta: {
              title: '集群列表',
              requiresAuth: true
            }
          },
          {
            path: 'nodes',
            name: 'NodeList',
            component: () => import('@/views/cluster/NodeList.vue'),
            meta: {
              title: '节点列表',
              requiresAuth: true
            }
          },
          {
            path: 'detail/:id',
            name: 'ClusterDetail',
            component: () => import('@/views/cluster/ClusterDetail.vue'),
            meta: {
              title: '集群详情',
              requiresAuth: true
            }
          },
          {
            path: 'nodes/:clusterId',
            name: 'ClusterNodes',
            component: () => import('@/views/cluster/ClusterNodes.vue'),
            meta: {
              title: '节点管理',
              requiresAuth: true
            }
          },
          {
            path: 'pods/:clusterId',
            name: 'ClusterPods',
            component: () => import('@/views/cluster/ClusterPods.vue'),
            meta: {
              title: 'Pod管理',
              requiresAuth: true
            }
          },
          {
            path: '/namespace',
            name: 'NamespaceList',
            component: () => import('@/views/namespace/NamespaceList.vue'),
            meta: { title: '命名空间管理', requiresAuth: true }
          },
          {
            path: '/namespace/detail',
            name: 'NamespaceDetail',
            component: () => import('@/views/namespace/NamespaceDetail.vue'),
            meta: { title: '命名空间详情', requiresAuth: true }
          },
          {
            path: '/namespace/edit',
            name: 'NamespaceEdit',
            component: () => import('@/views/namespace/NamespaceEdit.vue'),
            meta: { title: '编辑命名空间', requiresAuth: true }
          }
        ]
      },
      {
        path: '/user',
        name: 'UserManagement',
        redirect: '/user/list',
        meta: {
          title: '用户管理',
          requiresAuth: true,
          requiresAdmin: true
        },
        children: [
          {
            path: 'list',
            name: 'UserList',
            component: () => import('@/views/user/UserList.vue'),
            meta: {
              title: '用户列表',
              requiresAuth: true,
              requiresAdmin: true
            }
          }
        ]
      },
      {
        path: '/user/profile',
        name: 'UserProfile',
        component: () => import('@/views/user/UserProfile.vue'),
        meta: {
          title: '个人信息',
          requiresAuth: true
        }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue'),
    meta: {
      title: '页面不存在'
    }
  }
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  }
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore()
  
  // 初始化用户信息
  userStore.initUserInfo()
  
  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - K8s多集群管理平台`
  }
  
  // 如果是公开页面，直接放行
  if (to.meta.public) {
    // 如果已登录且访问登录页，重定向到首页
    if (to.path === '/login' && userStore.isLoggedIn) {
      next('/dashboard')
      return
    }
    next()
    return
  }
  
  // 检查是否需要登录
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    ElMessage.warning('请先登录')
    next('/login')
    return
  }
  
  // 检查是否需要管理员权限
  if (to.meta.requiresAdmin && !userStore.isAdmin) {
    ElMessage.error('没有权限访问该页面')
    next('/dashboard')
    return
  }
  
  next()
})

// 路由错误处理
router.onError((error) => {
  console.error('路由错误:', error)
  ElMessage.error('页面加载失败')
})

export default router

// 导出路由类型
export type { RouteRecordRaw }
