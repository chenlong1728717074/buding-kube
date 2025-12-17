<template>
  <div class="secret-detail">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="handleBack" style="margin-right: 16px;">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1>Secret详情</h1>
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="secret" class="detail-content">
      <el-card class="info-card">
        <template #header>
          <span>基本信息</span>
        </template>
        <div class="info-grid">
          <div class="info-item"><label>集群ID:</label><span>{{ clusterId }}</span></div>
          <div class="info-item"><label>命名空间:</label><span>{{ secret.namespace }}</span></div>
          <div class="info-item"><label>名称:</label><span>{{ secret.name }}</span></div>
          <div class="info-item"><label>类型:</label><span>{{ secret.type || '-' }}</span></div>
          <div class="info-item"><label>别名:</label><span>{{ secret.alias || '-' }}</span></div>
          <div class="info-item"><label>备注:</label><span>{{ secret.describe || '-' }}</span></div>
          <div class="info-item"><label>创建时间:</label><span>{{ formatDate(secret.creationTimestamp) }}</span></div>
        </div>
      </el-card>

      <el-card class="info-card">
        <template #header>
          <span>标签</span>
        </template>
        <el-table :data="labelRows" size="small" style="width:100%">
          <el-table-column prop="key" label="键" width="260" />
          <el-table-column prop="value" label="值" />
        </el-table>
      </el-card>

      <el-card class="info-card">
        <template #header>
          <span>注解</span>
        </template>
        <el-table :data="annotationRows" size="small" style="width:100%">
          <el-table-column prop="key" label="键" width="260" />
          <el-table-column prop="value" label="值" />
        </el-table>
      </el-card>

      <el-card class="info-card">
        <template #header>
          <span>数据</span>
        </template>
        <el-table :data="dataRows" size="small" style="width:100%">
          <el-table-column prop="key" label="键" width="220" />
          <el-table-column label="值">
            <template #default="{ row }">
              <el-input :model-value="row.value" type="textarea" :rows="2" readonly />
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <div v-else class="empty-container">
      <el-empty description="未找到Secret" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { secretApi, type SecretVO } from '@/api/secret'

const route = useRoute()
const router = useRouter()

const clusterId = ref('')
const namespace = ref('')
const name = ref('')

const loading = ref(false)
const secret = ref<SecretVO | null>(null)

const labelRows = ref<{ key: string; value: string }[]>([])
const annotationRows = ref<{ key: string; value: string }[]>([])
const dataRows = ref<{ key: string; value: string }[]>([])

const handleBack = () => router.back()

const decodeBase64ToUtf8 = (b64: string) => {
  try {
    const bin = atob(b64)
    const bytes = new Uint8Array(bin.length)
    for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i)
    return new TextDecoder().decode(bytes)
  } catch {
    return b64
  }
}

const fetchDetail = async () => {
  if (!clusterId.value || !namespace.value || !name.value) {
    secret.value = null
    return
  }
  loading.value = true
  try {
    const resp = await secretApi.getInfo({ clusterId: clusterId.value, namespace: namespace.value, name: name.value })
    if (resp.code === 200) {
      secret.value = resp.data || null
      labelRows.value = Object.entries(resp.data?.labels || {}).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
      annotationRows.value = Object.entries(resp.data?.annotations || {})
        .filter(([k]) => k !== 'alias' && k !== 'describe')
        .map(([k, v]) => ({ key: k, value: String(v ?? '') }))
      const displayData = resp.data?.stringData && Object.keys(resp.data.stringData).length > 0 ? resp.data.stringData : undefined
      if (displayData) {
        dataRows.value = Object.entries(displayData).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
      } else {
        const data = (resp.data?.data || {}) as Record<string, string>
        dataRows.value = Object.entries(data).map(([k, v]) => ({ key: k, value: decodeBase64ToUtf8(String(v ?? '')) }))
      }
      return
    }
    secret.value = null
  } catch {
    ElMessage.error('获取详情失败')
    secret.value = null
  } finally {
    loading.value = false
  }
}

const formatDate = (dateString?: string) => {
  if (!dateString) return '-'
  try {
    const d = new Date(dateString)
    if (isNaN(d.getTime())) return '-'
    return d.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  } catch {
    return '-'
  }
}

onMounted(() => {
  clusterId.value = (route.query.clusterId as string) || ''
  namespace.value = (route.query.namespace as string) || ''
  name.value = (route.query.name as string) || ''
  fetchDetail()
})
</script>

<style scoped>
.secret-detail { padding: 20px; }
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
.header-left { display:flex; align-items:center; gap:16px; }
.page-header h1 { margin:0; font-size:24px; font-weight:600; color:#2c3e50; }
.loading-container { padding: 20px; }
.detail-content { display:flex; flex-direction:column; gap:16px; }
.info-card { border-radius:8px; }
.info-grid { display:grid; grid-template-columns: repeat(2, 1fr); gap:12px; }
.info-item label { color:#64748b; margin-right:8px; }
</style>
