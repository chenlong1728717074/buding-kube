import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useClusterStore } from '@/stores/cluster'
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
  // 简单布局（无侧边栏）- 用于首页和集群列表
  {
    path: '/',
    name: 'SimpleLayout',
    component: () => import('@/views/SimpleLayout.vue'),
    redirect: '/home',
    meta: {
      requiresAuth: true
    },
    children: [
      {
        path: '/home',
        name: 'Home',
        component: () => import('@/views/Home.vue'),
        meta: {
          title: '首页',
          requiresAuth: true
        }
      },
      {
        path: '/cluster',
        name: 'ClusterList',
        component: () => import('@/views/cluster/ClusterList.vue'),
        meta: {
          title: '集群管理',
          requiresAuth: true
        }
      },
      {
        path: '/user/profile',
        name: 'UserProfile',
        component: () => import('@/views/user/UserProfile.vue'),
        meta: {
          title: '个人信息',
          requiresAuth: true
        }
      },
      {
        path: '/user/management',
        name: 'UserManagement',
        component: () => import('@/views/user/UserList.vue'),
        meta: {
          title: '用户管理',
          requiresAuth: true,
          requiresAdmin: true
        }
      }
    ]
  },
  // 完整布局（有侧边栏）- 用于集群资源管理
  {
    path: '/cluster/:clusterId',
    name: 'ClusterLayout',
    component: () => import('@/views/Layout.vue'),
    redirect: (to) => `/cluster/${to.params.clusterId}/overview`,
    meta: {
      requiresAuth: true,
      requiresCluster: true
    },
    children: [
      {
        path: 'overview',
        name: 'ClusterOverview',
        component: () => import('@/views/cluster/ClusterOverview.vue'),
        meta: {
          title: '集群概览',
          requiresAuth: true
        }
      },
      {
        path: 'overview',
        name: 'ClusterOverview',
        component: () => import('@/views/cluster/ClusterOverview.vue'),
        meta: {
          title: '集群概览',
          requiresAuth: true
        }
      },
      {
        path: 'namespace',
        name: 'NamespaceList',
        component: () => import('@/views/namespace/NamespaceList.vue'),
        meta: { title: '命名空间管理', requiresAuth: true }
      },
      {
        path: 'namespace/detail',
        name: 'NamespaceDetail',
        component: () => import('@/views/namespace/NamespaceDetail.vue'),
        meta: { title: '命名空间详情', requiresAuth: true }
      },
      {
        path: 'namespace/edit',
        name: 'NamespaceEdit',
        component: () => import('@/views/namespace/NamespaceEdit.vue'),
        meta: { title: '编辑命名空间', requiresAuth: true }
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
        path: 'node/detail',
        name: 'NodeDetail',
        component: () => import('@/views/cluster/NodeDetail.vue'),
        meta: { title: '节点详情', requiresAuth: true }
      },
      {
        path: 'pod',
        name: 'PodList',
        component: () => import('@/views/pod/PodList.vue'),
        meta: {
          title: 'Pod管理',
          requiresAuth: true
        }
      },
      {
        path: 'pod/detail',
        name: 'PodDetail',
        component: () => import('@/views/pod/PodDetail.vue'),
        meta: {
          title: 'Pod详情',
          requiresAuth: true
        }
      },
      // 工作负载
      {
        path: 'workload/deployment',
        name: 'Deployment',
        component: () => import('@/views/workload/Deployment.vue'),
        meta: {
          title: 'Deployment',
          requiresAuth: true
        }
      },
      {
        path: 'workload/daemonset',
        name: 'DaemonSet',
        component: () => import('@/views/workload/DaemonSet.vue'),
        meta: {
          title: 'DaemonSet',
          requiresAuth: true
        }
      },
      {
        path: 'workload/statefulset',
        name: 'StatefulSet',
        component: () => import('@/views/workload/StatefulSet.vue'),
        meta: {
          title: 'StatefulSet',
          requiresAuth: true
        }
      },
      // 任务调度
      {
        path: 'job/job',
        name: 'Job',
        component: () => import('@/views/job/Job.vue'),
        meta: {
          title: 'Job',
          requiresAuth: true
        }
      },
      {
        path: 'job/cronjob',
        name: 'CronJob',
        component: () => import('@/views/job/CronJob.vue'),
        meta: {
          title: 'CronJob',
          requiresAuth: true
        }
      },
      // 服务发现
      {
        path: 'service/service',
        name: 'Service',
        component: () => import('@/views/service/Service.vue'),
        meta: {
          title: 'Service',
          requiresAuth: true
        }
      },
      {
        path: 'service/ingress',
        name: 'Ingress',
        component: () => import('@/views/service/Ingress.vue'),
        meta: {
          title: 'Ingress',
          requiresAuth: true
        }
      },
      {
        path: 'service/endpoint',
        name: 'Endpoint',
        component: () => import('@/views/service/Endpoint.vue'),
        meta: {
          title: 'Endpoint',
          requiresAuth: true
        }
      },
      {
        path: 'service/endpointslice',
        name: 'EndpointSlice',
        component: () => import('@/views/service/EndpointSlice.vue'),
        meta: {
          title: 'EndpointSlice',
          requiresAuth: true
        }
      },
      // 配置管理
      {
        path: 'config/configmap',
        name: 'ConfigMap',
        component: () => import('@/views/config/ConfigMap.vue'),
        meta: {
          title: 'ConfigMap',
          requiresAuth: true
        }
      },
      {
        path: 'config/configmap/detail',
        name: 'ConfigMapDetail',
        component: () => import('@/views/config/ConfigMapDetail.vue'),
        meta: {
          title: 'ConfigMap详情',
          requiresAuth: true
        }
      },
      {
        path: 'config/secret',
        name: 'Secret',
        component: () => import('@/views/config/Secret.vue'),
        meta: {
          title: 'Secret',
          requiresAuth: true
        }
      },
      {
        path: 'config/secret/detail',
        name: 'SecretDetail',
        component: () => import('@/views/config/SecretDetail.vue'),
        meta: {
          title: 'Secret详情',
          requiresAuth: true
        }
      },
      {
        path: 'config/serviceaccount',
        name: 'ServiceAccount',
        component: () => import('@/views/config/ServiceAccount.vue'),
        meta: {
          title: 'ServiceAccount',
          requiresAuth: true
        }
      },
      {
        path: 'config/serviceaccount/detail',
        name: 'ServiceAccountDetail',
        component: () => import('@/views/config/ServiceAccountDetail.vue'),
        meta: {
          title: 'ServiceAccount详情',
          requiresAuth: true
        }
      },
      // 存储管理
      {
        path: 'storage/pv',
        name: 'PersistentVolume',
        component: () => import('@/views/storage/PersistentVolume.vue'),
        meta: {
          title: 'PersistentVolume',
          requiresAuth: true
        }
      },
      {
        path: 'storage/pvc',
        name: 'PersistentVolumeClaim',
        component: () => import('@/views/storage/PersistentVolumeClaim.vue'),
        meta: {
          title: 'PersistentVolumeClaim',
          requiresAuth: true
        }
      },
      {
        path: 'storage/storageclass',
        name: 'StorageClass',
        component: () => import('@/views/storage/StorageClass.vue'),
        meta: {
          title: 'StorageClass',
          requiresAuth: true
        }
      },
      // 资源管理
      {
        path: 'resource/crd',
        name: 'CRD',
        component: () => import('@/views/resource/CRD.vue'),
        meta: {
          title: 'CRD',
          requiresAuth: true
        }
      },
      {
        path: 'resource/tools',
        name: 'Tools',
        component: () => import('@/views/resource/Tools.vue'),
        meta: {
          title: '工具',
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
  const clusterStore = useClusterStore()

  // 初始化用户信息和集群上下文
  userStore.initUserInfo()
  clusterStore.initClusterContext()

  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - Buding k8s多集群管理平台`
  }

  // 如果是公开页面，直接放行
  if (to.meta.public) {
    // 如果已登录且访问登录页，重定向到首页
    if (to.path === '/login' && userStore.isLoggedIn) {
      next('/home')
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
    next('/home')
    return
  }

  // 检查是否需要集群上下文
  if (to.meta.requiresCluster && !clusterStore.hasCluster) {
    ElMessage.warning('请先选择集群')
    next('/cluster')
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
