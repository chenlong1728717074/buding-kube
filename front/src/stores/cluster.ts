import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface ClusterInfo {
  id: string
  name: string
  apiServer: string
  status?: string
}

export const useClusterStore = defineStore('cluster', () => {
  // 当前选中的集群
  const currentCluster = ref<ClusterInfo | null>(null)
  
  // 从 localStorage 初始化
  const initClusterContext = () => {
    try {
      const saved = localStorage.getItem('currentCluster')
      if (saved) {
        currentCluster.value = JSON.parse(saved)
      }
    } catch (error) {
      console.error('Failed to load cluster context:', error)
      localStorage.removeItem('currentCluster')
    }
  }
  
  // 设置当前集群
  const setCurrentCluster = (cluster: ClusterInfo | null) => {
    currentCluster.value = cluster
    if (cluster) {
      localStorage.setItem('currentCluster', JSON.stringify(cluster))
    } else {
      localStorage.removeItem('currentCluster')
    }
  }
  
  // 清除当前集群
  const clearCurrentCluster = () => {
    currentCluster.value = null
    localStorage.removeItem('currentCluster')
  }
  
  // 是否已选择集群
  const hasCluster = computed(() => currentCluster.value !== null)
  
  // 当前集群 ID
  const currentClusterId = computed(() => currentCluster.value?.id || '')
  
  // 当前集群名称
  const currentClusterName = computed(() => currentCluster.value?.name || '')
  
  return {
    currentCluster,
    hasCluster,
    currentClusterId,
    currentClusterName,
    initClusterContext,
    setCurrentCluster,
    clearCurrentCluster
  }
})
