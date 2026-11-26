import request, { type ApiResponse } from '@/utils/request'

export type TopologyStatusType = 'success' | 'warning' | 'danger' | 'info'

export interface TopologyNodeVO {
  id: string
  kind: string
  name: string
  namespace: string
  status: string
  statusType: TopologyStatusType
  layer: number
  summary: string
  extra?: Record<string, string>
}

export interface TopologyEdgeVO {
  id: string
  source: string
  target: string
  relation: string
  label?: string
  details?: string[]
}

export interface TopologyVO {
  namespace: string
  nodes: TopologyNodeVO[]
  edges: TopologyEdgeVO[]
  summary: Record<string, number>
}

export const topologyApi = {
  getNamespaceTopology(params: { clusterId: string; namespace: string }): Promise<ApiResponse<TopologyVO>> {
    return request.get('/topology', { params })
  }
}

export default topologyApi
