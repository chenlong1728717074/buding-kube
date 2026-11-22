import request, { type ApiResponse } from '@/utils/request'

export interface NetworkPolicySupportVO {
  // 识别到的网络插件, 未识别时为空
  plugin: string
  support: 'supported' | 'unsupported' | 'unknown'
  message: string
}

// 同一集群只检测一次(后端也有 5 分钟缓存)
const networkPolicyCache = new Map<string, Promise<NetworkPolicySupportVO | null>>()

export const capabilityApi = {
  networkPolicy(clusterId: string): Promise<NetworkPolicySupportVO | null> {
    let cached = networkPolicyCache.get(clusterId)
    if (!cached) {
      cached = request
        .get('/capability/networkPolicy', { params: { clusterId } })
        .then((resp: ApiResponse<NetworkPolicySupportVO>) => resp.data || null)
        .catch(() => {
          // 检测失败时下次重新检测
          networkPolicyCache.delete(clusterId)
          return null
        })
      networkPolicyCache.set(clusterId, cached)
    }
    return cached
  }
}
