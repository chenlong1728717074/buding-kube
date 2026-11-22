import { ref } from 'vue'
import { pvcApi } from '@/api/pvc'
import { configMapApi } from '@/api/configmap'
import { secretApi } from '@/api/secret'
import { serviceAccountApi } from '@/api/serviceaccount'

// 创建工作负载 / 容器组时, 命名空间内可选的资源: 持久卷声明、ConfigMap、Secret、服务账户
export function useNamespaceOptions(clusterId: () => string) {
  const loading = ref(false)
  const pvcOptions = ref<string[]>([])
  const configMapOptions = ref<string[]>([])
  const secretOptions = ref<string[]>([])
  const pullSecretOptions = ref<string[]>([])
  const serviceAccountOptions = ref<string[]>([])
  let latest = ''

  const names = (items?: { name: string }[]) => (items || []).map(i => i.name).sort()

  const load = async (namespace: string) => {
    latest = namespace
    if (!namespace) {
      pvcOptions.value = configMapOptions.value = secretOptions.value = pullSecretOptions.value = serviceAccountOptions.value = []
      return
    }
    loading.value = true
    const params = { clusterId: clusterId(), namespace, page: 1, pageSize: 10000 }
    const [pvcs, cms, secrets, sas] = await Promise.allSettled([
      pvcApi.getPVCList(params),
      configMapApi.getList(params),
      secretApi.getList(params),
      serviceAccountApi.getList(params)
    ])
    // 加载期间切换了命名空间, 丢弃旧结果
    if (namespace !== latest) return
    pvcOptions.value = pvcs.status === 'fulfilled' ? names(pvcs.value.data?.items) : []
    configMapOptions.value = cms.status === 'fulfilled' ? names(cms.value.data?.items) : []
    if (secrets.status === 'fulfilled') {
      const items = secrets.value.data?.items || []
      secretOptions.value = names(items)
      pullSecretOptions.value = names(items.filter(s => !s.type || s.type === 'kubernetes.io/dockerconfigjson' || s.type === 'kubernetes.io/dockercfg'))
    } else {
      secretOptions.value = []
      pullSecretOptions.value = []
    }
    serviceAccountOptions.value = sas.status === 'fulfilled' ? names(sas.value.data?.items) : []
    loading.value = false
  }

  return { loading, pvcOptions, configMapOptions, secretOptions, pullSecretOptions, serviceAccountOptions, load }
}
