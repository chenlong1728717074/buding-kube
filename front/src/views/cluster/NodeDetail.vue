<template>
  <div class="node-detail">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="handleBack" circle>
          <el-icon><ArrowLeft /></el-icon>
        </el-button>
        <div class="header-info">
          <h1>{{ nodeInfo?.server?.hostname || '节点详情' }}</h1>
          <p class="header-description">查看节点的详细信息</p>
        </div>
      </div>
      <div class="header-actions">
        <el-dropdown @command="handleMoreActions">
          <el-button>
            更多操作
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item :command="nodeInfo?.server?.unSchedule ? 'uncordon' : 'cordon'">
                {{ nodeInfo?.server?.unSchedule ? '启用调度' : '停止调度' }}
              </el-dropdown-item>
              <el-dropdown-item command="drain">驱逐Pod</el-dropdown-item>
              <el-dropdown-item command="label">管理标签</el-dropdown-item>
              <el-dropdown-item command="taint">管理污点</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <div class="detail-content" v-loading="loading">
      <el-row :gutter="20">
        <!-- 基本信息 -->
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>基本信息</span>
              </div>
            </template>
            <div class="info-grid">
              <div class="info-item">
                <label>节点名称:</label>
                <span>{{ nodeInfo?.server?.hostname || '-' }}</span>
              </div>
              <div class="info-item">
                <label>状态:</label>
                <el-tag :type="getStatusType(nodeInfo?.server?.status)">
                  {{ getStatusText(nodeInfo?.server?.status) }}
                </el-tag>
              </div>
              <div class="info-item">
                <label>角色:</label>
                <el-tag :type="nodeInfo?.server?.role === 'master' ? 'warning' : 'info'">
                  {{ nodeInfo?.server?.role === 'master' ? 'Master' : 'Worker' }}
                </el-tag>
              </div>
              <div class="info-item">
                <label>Kubelet版本:</label>
                <span>{{ nodeInfo?.server?.kubeletVersion || '-' }}</span>
              </div>
              <div class="info-item">
                <label>内部IP:</label>
                <span>{{ nodeInfo?.server?.ip || '-' }}</span>
              </div>
              <div class="info-item">
                <label>操作系统:</label>
                <span>{{ nodeInfo?.server?.osImage || '-' }}</span>
              </div>
              <div class="info-item">
                <label>系统类型:</label>
                <span>{{ nodeInfo?.server?.osType || '-' }}</span>
              </div>
              <div class="info-item">
                <label>架构:</label>
                <span>{{ nodeInfo?.server?.arch || '-' }}</span>
              </div>
              <div class="info-item">
                <label>内核版本:</label>
                <span>{{ nodeInfo?.server?.kernelVersion || '-' }}</span>
              </div>
              <div class="info-item">
                <label>容器运行时:</label>
                <span>{{ nodeInfo?.server?.containerRuntime || '-' }}</span>
              </div>
              <div class="info-item">
                <label>KubeProxy版本:</label>
                <span>{{ nodeInfo?.server?.kubeProxyVersion || '-' }}</span>
              </div>
              <div class="info-item">
                <label>可调度:</label>
                <el-tag :type="!nodeInfo?.server?.unSchedule ? 'success' : 'danger'">
                  {{ !nodeInfo?.server?.unSchedule ? '是' : '否' }}
                </el-tag>
              </div>
              <div class="info-item">
                <label>创建时间:</label>
                <span>{{ formatDate(nodeInfo?.server?.createTime) }}</span>
              </div>
            </div>
          </el-card>
        </el-col>

        <!-- 资源信息 -->
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>资源信息</span>
              </div>
            </template>
            <div class="info-list">
              <div class="info-item">
                <label>CPU：</label>
                <span>{{ nodeInfo?.runtime?.cpu || '-' }}</span>
              </div>
              <div class="info-item">
                <label>内存：</label>
                <span>{{ nodeInfo?.runtime?.memory || '-' }}</span>
              </div>
              <div class="info-item">
                <label>临时存储：</label>
                <span>{{ nodeInfo?.runtime?.ephemeralStorage || '-' }}</span>
              </div>
              <div class="info-item" v-if="nodeInfo?.runtime?.nvidiaGpu">
                <label>NVIDIA GPU：</label>
                <span>{{ nodeInfo.runtime.nvidiaGpu }}</span>
              </div>
              <div class="info-item" v-if="nodeInfo?.runtime?.amdGpu">
                <label>AMD GPU：</label>
                <span>{{ nodeInfo.runtime.amdGpu }}</span>
              </div>
              <div class="info-item" v-if="nodeInfo?.runtime?.intelGpu">
                <label>Intel GPU：</label>
                <span>{{ nodeInfo.runtime.intelGpu }}</span>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" style="margin-top: 20px;">
        <!-- 标签 -->
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>标签</span>
              </div>
            </template>
            <div class="labels-container">
              <el-tag 
                v-for="(value, key) in nodeInfo?.labels" 
                :key="key" 
                class="label-tag"
                size="small"
              >
                {{ key }}: {{ value }}
              </el-tag>
              <span v-if="!nodeInfo?.labels || Object.keys(nodeInfo.labels).length === 0">无标签</span>
            </div>
          </el-card>
        </el-col>

        <!-- 注解 -->
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>注解</span>
              </div>
            </template>
            <div class="annotations-container">
              <el-tag 
                v-for="(value, key) in nodeInfo?.annotations" 
                :key="key" 
                type="info"
                size="small"
                style="margin-right: 8px; margin-bottom: 8px;"
              >
                {{ key }}: {{ value }}
              </el-tag>
              <span v-if="!nodeInfo?.annotations || Object.keys(nodeInfo.annotations).length === 0">无注解</span>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- Pod列表 -->
      <el-row :gutter="20" style="margin-top: 20px;" v-if="nodeInfo?.pods && nodeInfo.pods.length > 0">
        <el-col :span="24">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>Pod列表 ({{ nodeInfo.pods.length }})</span>
              </div>
            </template>
            <el-table :data="nodeInfo.pods" style="width: 100%">
              <el-table-column prop="name" label="名称" min-width="200">
                <template #default="{ row }">
                  <el-link type="primary" @click="handleViewPodDetail(row)">
                    {{ row.name }}
                  </el-link>
                </template>
              </el-table-column>
              <el-table-column prop="namespace" label="命名空间" width="150" />
              <el-table-column prop="image" label="镜像" min-width="200" show-overflow-tooltip />
              <el-table-column prop="status" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="getPodStatusType(row.status)" size="small">
                    {{ row.status }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="ready" label="就绪" width="80" />
              <el-table-column prop="restarts" label="重启" width="80" />
              <el-table-column prop="age" label="运行时长" width="120" />
              <el-table-column prop="ip" label="Pod IP" width="140" />
              <el-table-column prop="ports" label="端口" width="120">
                <template #default="{ row }">
                  <span v-if="row.ports && row.ports.length > 0">
                    {{ row.ports.join(', ') }}
                  </span>
                  <span v-else>-</span>
                </template>
              </el-table-column>
            </el-table>
          </el-card>
        </el-col>
      </el-row>

      <!-- 事件列表 -->
      <el-row :gutter="20" style="margin-top: 20px;" v-if="nodeInfo?.events && nodeInfo.events.length > 0">
        <el-col :span="24">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>事件列表</span>
              </div>
            </template>
            <el-table :data="nodeInfo.events" style="width: 100%">
              <el-table-column prop="type" label="类型" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.type === 'Warning' ? 'warning' : 'info'" size="small">
                    {{ row.type }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="reason" label="原因" width="150" />
              <el-table-column prop="message" label="消息" min-width="300" show-overflow-tooltip />
              <el-table-column prop="count" label="次数" width="80" />
              <el-table-column prop="firstTime" label="首次时间" width="180" />
              <el-table-column prop="lastTime" label="最后时间" width="180" />
            </el-table>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  ArrowLeft,
  ArrowDown
} from '@element-plus/icons-vue'
import { clusterApi } from '@/api/cluster'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const nodeInfo = ref<any>(null)

// 获取节点详情
const fetchNodeDetail = async () => {
  try {
    loading.value = true
    const clusterId = (route.params.clusterId as string) || (route.query.clusterId as string)
    const hostname = (route.query.hostname as string) || (route.query.nodeName as string) || (route.query.node as string)
    
    if (!clusterId || !hostname) {
      ElMessage.error('缺少必要参数')
      handleBack()
      return
    }
    
    const response = await clusterApi.getNodeDetail(clusterId, hostname)
    nodeInfo.value = response.data
  } catch (error: any) {
    ElMessage.error('获取节点详情失败')
  } finally {
    loading.value = false
  }
}

// 返回列表
const handleBack = () => {
  router.back()
}

// 查看Pod详情
const handleViewPodDetail = (pod: any) => {
  const clusterId = (route.params.clusterId as string) || (route.query.clusterId as string)
  router.push({
    path: `/cluster/${clusterId}/pod/detail`,
    query: {
      namespace: pod.namespace,
      name: pod.name
    }
  })
}

// 更多操作
const handleMoreActions = async (command: string) => {
  const clusterId = (route.params.clusterId as string) || (route.query.clusterId as string)
  const hostname = nodeInfo.value?.server?.hostname
  
  switch (command) {
    case 'cordon':
    case 'uncordon':
      await handleToggleSchedule(clusterId, hostname, command === 'cordon')
      break
    case 'drain':
      await handleDrainNode(clusterId, hostname)
      break
    case 'label':
      ElMessage.info('管理标签功能暂未开放')
      break
    case 'taint':
      ElMessage.info('管理污点功能暂未开放')
      break
  }
}

// 切换调度状态
const handleToggleSchedule = async (clusterId: string, hostname: string, unSchedule: boolean) => {
  const action = unSchedule ? '停止' : '启用'
  try {
    await ElMessageBox.confirm(
      `确定要${action}节点 ${hostname} 的调度吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await clusterApi.toggleNodeSchedule(clusterId, hostname, unSchedule)
    ElMessage.success(`${action}调度成功`)
    fetchNodeDetail()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}调度失败`)
    }
  }
}

// 驱逐Pod
const handleDrainNode = async (clusterId: string, hostname: string) => {
  try {
    await ElMessageBox.confirm(
      `确定要驱逐节点 ${hostname} 上的所有Pod吗？此操作会将Pod重新调度到其他节点。`,
      '确认驱逐',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    ElMessage.info('驱逐Pod功能暂未开放')
  } catch (error: any) {
    if (error !== 'cancel') {
      // 用户取消操作
    }
  }
}

// 格式化日期
const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  try {
    return new Date(dateStr).toLocaleString('zh-CN')
  } catch {
    return dateStr
  }
}

// 获取状态类型
const getStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    Ready: 'success',
    NotReady: 'danger',
    Unknown: 'warning'
  }
  return statusMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    Ready: '就绪',
    NotReady: '未就绪',
    Unknown: '未知'
  }
  return statusMap[status] || status
}

// 获取Pod状态类型
const getPodStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    Running: 'success',
    Pending: 'warning',
    Failed: 'danger',
    Succeeded: 'info'
  }
  return statusMap[status] || 'info'
}

// 初始化
onMounted(() => {
  fetchNodeDetail()
})
</script>

<style scoped>
.node-detail {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-info h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
}

.header-description {
  margin: 4px 0 0 0;
  color: #666;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-card {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
}

.info-item {
  display: flex;
  align-items: center;
}

.info-item label {
  font-weight: 500;
  color: #606266;
  min-width: 120px;
}

.info-item span {
  color: #303133;
}

@media (max-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
  
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }
}

.labels-container,
.annotations-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.label-tag {
  margin: 0;
}
</style>
