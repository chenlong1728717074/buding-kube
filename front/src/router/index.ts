import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { useClusterStore } from '@/stores/cluster'
import { useTagsStore } from '@/stores/tags'
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
        path: 'namespace',
        name: 'NamespaceList',
        component: () => import('@/views/namespace/NamespaceList.vue'),
        meta: { title: '命名空间管理', requiresAuth: true }
      },
      {
        path: 'namespace/detail',
        name: 'NamespaceDetail',
        component: () => import('@/views/namespace/NamespaceDetail.vue'),
        meta: {
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'namespace', title: '命名空间详情', requiresAuth: true }
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
        meta: {
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'nodes', title: '节点详情', requiresAuth: true }
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
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'pod',
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
        path: 'workload/deployment/detail',
        name: 'WorkloadDeploymentDetail',
        component: () => import('@/views/workload/WorkloadDetail.vue'),
        meta: {
          title: 'Deployment详情',
          requiresAuth: true,
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'workload/deployment',
          kind: 'deployment',
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
        path: 'workload/daemonset/detail',
        name: 'WorkloadDaemonSetDetail',
        component: () => import('@/views/workload/WorkloadDetail.vue'),
        meta: {
          title: 'DaemonSet详情',
          requiresAuth: true,
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'workload/daemonset',
          kind: 'daemonset',
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
      {
        path: 'workload/statefulset/detail',
        name: 'WorkloadStatefulSetDetail',
        component: () => import('@/views/workload/WorkloadDetail.vue'),
        meta: {
          title: 'StatefulSet详情',
          requiresAuth: true,
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'workload/statefulset',
          kind: 'statefulset',
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
      {
        path: 'job/job/detail',
        name: 'JobDetail',
        component: () => import('@/views/job/JobDetail.vue'),
        meta: {
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'job/job',
          title: 'Job详情',
          requiresAuth: true
        }
      },
      {
        path: 'job/cronjob/detail',
        name: 'CronJobDetail',
        component: () => import('@/views/job/CronJobDetail.vue'),
        meta: {
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'job/cronjob',
          title: 'CronJob详情',
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
        path: 'service/service/detail',
        name: 'ServiceDetail',
        component: () => import('@/views/service/ServiceDetail.vue'),
        meta: {
          title: 'Service详情',
          requiresAuth: true,
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'service/service',
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
        path: 'service/ingress/detail',
        name: 'IngressDetail',
        component: () => import('@/views/service/IngressDetail.vue'),
        meta: {
          title: 'Ingress详情',
          requiresAuth: true,
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'service/ingress',
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
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'config/configmap',
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
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'config/secret',
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
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'config/serviceaccount',
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
      {
        path: 'storage/storageclass/detail',
        name: 'StorageClassDetail',
        component: () => import('@/views/storage/StorageClassDetail.vue'),
        meta: {
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'storage/storageclass',
          title: 'StorageClass详情',
          requiresAuth: true
        }
      },
      // 工作负载扩展
      {
        path: 'workload/replicaset',
        name: 'ReplicaSet',
        component: () => import('@/views/workload/ReplicaSet.vue'),
        meta: {
          title: 'ReplicaSet',
          requiresAuth: true
        }
      },
      {
        path: 'workload/hpa',
        name: 'Hpa',
        component: () => import('@/views/workload/Hpa.vue'),
        meta: {
          title: '弹性伸缩',
          requiresAuth: true
        }
      },
      {
        path: 'service/ingressclass',
        name: 'IngressClass',
        component: () => import('@/views/service/IngressClass.vue'),
        meta: {
          title: 'IngressClass',
          requiresAuth: true
        }
      },
      // 策略管理
      {
        path: 'policy/resourcequota',
        name: 'ResourceQuota',
        component: () => import('@/views/policy/ResourceQuota.vue'),
        meta: {
          title: '资源配额',
          requiresAuth: true
        }
      },
      {
        path: 'policy/limitrange',
        name: 'LimitRange',
        component: () => import('@/views/policy/LimitRange.vue'),
        meta: {
          title: '默认资源限制',
          requiresAuth: true
        }
      },
      {
        path: 'policy/networkpolicy',
        name: 'NetworkPolicy',
        component: () => import('@/views/policy/NetworkPolicy.vue'),
        meta: {
          title: '网络策略',
          requiresAuth: true
        }
      },
      {
        path: 'policy/pdb',
        name: 'PodDisruptionBudget',
        component: () => import('@/views/policy/PodDisruptionBudget.vue'),
        meta: {
          title: '中断预算',
          requiresAuth: true
        }
      },
      {
        path: 'policy/priorityclass',
        name: 'PriorityClass',
        component: () => import('@/views/policy/PriorityClass.vue'),
        meta: {
          title: '优先级',
          requiresAuth: true
        }
      },
      // 运维工具
      {
        path: 'resource/events',
        name: 'EventCenter',
        component: () => import('@/views/resource/EventCenter.vue'),
        meta: {
          title: '事件中心',
          requiresAuth: true
        }
      },
      {
        path: 'resource/kubectl',
        name: 'Kubectl',
        component: () => import('@/views/resource/Kubectl.vue'),
        meta: {
          title: 'Web kubectl',
          requiresAuth: true
        }
      },
      {
        path: 'resource/crd/instances',
        name: 'CustomResourceList',
        component: () => import('@/views/resource/CustomResourceList.vue'),
        meta: {
          // 所属菜单: 从这里进入的页面显示在该菜单的标签里
          parent: 'resource/crd',
          title: '自定义资源',
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

// 顶部标签: 跳转完成后更新当前标签
router.afterEach((to, from, failure) => {
  const tagsStore = useTagsStore()
  if (failure) tagsStore.clearIntent()
  // 刷新浏览器、或从首页 / 集群列表进入集群时, 按地址找对应标签, 不覆盖之前停留的标签
  else tagsStore.onRoute(to, !from.matched.length || !from.path.startsWith('/cluster/'))
})

// 页面代码加载失败: 开发时 Vite 重新预构建依赖、或发布新版本后旧文件已不存在.
// 直接重新加载目标页面一次; 短时间内再次失败时才提示, 避免无限刷新
const CHUNK_RELOAD_KEY = 'chunkReloadAt'
const isChunkLoadError = (error: unknown) =>
  error instanceof Error &&
  /Failed to fetch dynamically imported module|Importing a module script failed|error loading dynamically imported module/i.test(error.message)

router.onError((error, to) => {
  console.error('路由错误:', error)
  if (isChunkLoadError(error)) {
    const last = Number(sessionStorage.getItem(CHUNK_RELOAD_KEY) || 0)
    if (Date.now() - last > 10000) {
      sessionStorage.setItem(CHUNK_RELOAD_KEY, String(Date.now()))
      window.location.assign(to.fullPath)
      return
    }
    ElMessage.error('页面加载失败，请检查前端服务是否正常运行，或查看终端中的编译错误')
    return
  }
  ElMessage.error('页面加载失败')
})

export default router

// 导出路由类型
export type { RouteRecordRaw }
