import request, { type ApiResponse } from '@/utils/request'
import { jsonToYaml } from '@/utils/yaml'

// 与后端 internal/service/resource_yaml.go 中的 resourceKinds 保持一致; custom 表示 CRD 的实例
export type YamlResourceKind =
  | 'deployment'
  | 'statefulset'
  | 'daemonset'
  | 'replicaset'
  | 'cronjob'
  | 'job'
  | 'pod'
  | 'service'
  | 'ingress'
  | 'endpoints'
  | 'endpointslice'
  | 'configmap'
  | 'secret'
  | 'serviceaccount'
  | 'namespace'
  | 'node'
  | 'persistentvolume'
  | 'persistentvolumeclaim'
  | 'storageclass'
  | 'customresourcedefinition'
  | 'horizontalpodautoscaler'
  | 'networkpolicy'
  | 'resourcequota'
  | 'limitrange'
  | 'poddisruptionbudget'
  | 'priorityclass'
  | 'ingressclass'
  | 'custom'

// 资源类型; kind 为 custom 时用 group / version / resource / kindName / namespaced 指定自定义资源
export interface ResourceRef {
  kind: YamlResourceKind
  group?: string
  version?: string
  resource?: string
  kindName?: string
  namespaced?: boolean
}

// 一个具体的资源; 集群级资源不传 namespace
export interface ResourceYamlTarget extends ResourceRef {
  clusterId: string
  namespace?: string
  name: string
}

// 列定义, 与 kubectl get 的列相同; priority 大于 0 的列对应 kubectl get -o wide
export interface ResourceTableColumn {
  name: string
  type: string
  format: string
  description: string
  priority: number
}

export interface ResourceTableRow {
  name: string
  namespace: string
  uid: string
  createTime: string
  labels?: Record<string, string> | null
  ownerKind?: string
  ownerName?: string
  // 与 columns 一一对应
  cells: unknown[]
}

export interface ResourceTable {
  columns: ResourceTableColumn[]
  items: ResourceTableRow[]
  total: number
  page: number
  pageSize: number
}

export interface ResourceTableQuery extends ResourceRef {
  clusterId: string
  namespace?: string
  keyword?: string
  labelSelector?: string
  page: number
  pageSize: number
}

// Kubernetes 对象(JSON); 各页面按需读取字段
export type KubeObject = Record<string, any>

export const resourceApi = {
  // 查询单个资源的 YAML; 列表接口不返回 yaml, 查看时按需获取
  getYaml(params: ResourceYamlTarget): Promise<ApiResponse<{ yaml: string }>> {
    return request.get('/resource/yaml', { params })
  },

  // 保存编辑后的 YAML(资源必须已存在, 不允许修改名称和命名空间)
  updateYaml(data: ResourceYamlTarget & { yaml: string }): Promise<ApiResponse<void>> {
    return request.put('/resource/yaml', data)
  },

  // 通过 YAML 创建资源, 已存在则更新; createOnly 为 true 时同名资源已存在会报错
  apply(data: ResourceRef & { clusterId: string; namespace?: string; yaml: string; createOnly?: boolean }): Promise<ApiResponse<void>> {
    return request.post('/resource/apply', data)
  },

  // 通用列表(Table 格式, 分页在后端完成)
  table(params: ResourceTableQuery): Promise<ApiResponse<ResourceTable>> {
    return request.get('/resource/table', { params })
  },

  // 单个资源的完整对象(包含 status)
  object<T = KubeObject>(params: ResourceYamlTarget): Promise<ApiResponse<T>> {
    return request.get('/resource/object', { params })
  },

  // 列出完整对象; namespace 为空时查询所有命名空间
  objects<T = KubeObject>(params: ResourceRef & { clusterId: string; namespace?: string; labelSelector?: string }): Promise<ApiResponse<T[]>> {
    return request.get('/resource/objects', { params })
  },

  delete(params: ResourceYamlTarget): Promise<ApiResponse<void>> {
    return request.delete('/resource', { params })
  }
}

// 获取 YAML 文本; 失败时抛出异常(请求拦截器已统一弹出错误提示)
export const fetchResourceYaml = async (params: ResourceYamlTarget): Promise<string> => {
  const resp = await resourceApi.getYaml(params)
  if (resp.code !== 200) {
    throw new Error(resp.msg || '获取YAML失败')
  }
  return resp.data?.yaml || ''
}

// 表单生成的对象提交: create 为新建(同名已存在时报错), 否则更新已有资源
export const saveResourceObject = (clusterId: string, ref: ResourceRef, obj: KubeObject, create: boolean) => {
  const yaml = jsonToYaml(JSON.parse(JSON.stringify(obj)))
  const namespace = obj.metadata?.namespace as string | undefined
  if (create) return resourceApi.apply({ ...ref, clusterId, namespace, yaml, createOnly: true })
  return resourceApi.updateYaml({ ...ref, clusterId, namespace, name: obj.metadata?.name as string, yaml })
}

// 编辑时以服务端对象为基础, 去掉由系统维护的字段
export const editableCopy = (obj: KubeObject): KubeObject => {
  const copy = JSON.parse(JSON.stringify(obj))
  delete copy.status
  if (copy.metadata) {
    delete copy.metadata.managedFields
    delete copy.metadata.resourceVersion
    delete copy.metadata.creationTimestamp
    delete copy.metadata.generation
    delete copy.metadata.uid
  }
  return copy
}

export default resourceApi
