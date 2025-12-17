<template>
  <div class="serviceaccount-detail">
    <div class="page-header">
      <div class="header-left">
        <el-button @click="handleBack" style="margin-right: 16px;">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
        <h1>ServiceAccount详情</h1>
      </div>
    </div>

    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="sa" class="detail-content">
      <el-card class="info-card">
        <template #header>
          <span>基本信息</span>
        </template>
        <div class="info-grid">
          <div class="info-item"><label>集群ID:</label><span>{{ clusterId }}</span></div>
          <div class="info-item"><label>命名空间:</label><span>{{ sa.namespace }}</span></div>
          <div class="info-item"><label>名称:</label><span>{{ sa.name }}</span></div>
          <div class="info-item"><label>自动挂载Token:</label><span>{{ sa.automountServiceAccountToken === false ? '否' : '是' }}</span></div>
          <div class="info-item"><label>创建时间:</label><span>{{ formatDate(sa.creationTimestamp) }}</span></div>
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
          <span>关联</span>
        </template>
        <div class="related-grid">
          <div class="related-block">
            <div class="related-title">imagePullSecrets</div>
            <div class="related-tags">
              <el-tag v-for="n in imagePullSecretNames" :key="n" size="small" style="margin-right:8px;margin-bottom:8px;">{{ n }}</el-tag>
              <span v-if="imagePullSecretNames.length === 0">-</span>
            </div>
          </div>
          <div class="related-block">
            <div class="related-title">secrets</div>
            <div class="related-tags">
              <el-tag v-for="n in secretNames" :key="n" size="small" style="margin-right:8px;margin-bottom:8px;">{{ n }}</el-tag>
              <span v-if="secretNames.length === 0">-</span>
            </div>
          </div>
        </div>
      </el-card>
    </div>

    <div v-else class="empty-container">
      <el-empty description="未找到ServiceAccount" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { serviceAccountApi, type ServiceAccountVO } from '@/api/serviceaccount'

const route = useRoute()
const router = useRouter()

const clusterId = ref('')
const namespace = ref('')
const name = ref('')

const loading = ref(false)
const sa = ref<ServiceAccountVO | null>(null)

const labelRows = ref<{ key: string; value: string }[]>([])
const annotationRows = ref<{ key: string; value: string }[]>([])

const imagePullSecretNames = computed(() => (sa.value?.imagePullSecrets || []).map(x => x.name).filter(Boolean))
const secretNames = computed(() => (sa.value?.secrets || []).map(x => x.name).filter(Boolean))

const handleBack = () => router.back()

const fetchDetailFromList = async () => {
  if (!clusterId.value || !namespace.value || !name.value) {
    sa.value = null
    return
  }
  loading.value = true
  try {
    const resp = await serviceAccountApi.getList({ clusterId: clusterId.value, namespace: namespace.value, page: 1, pageSize: 10000 })
    if (resp.code === 200) {
      const items = resp.data.items || []
      const found = items.find((it: ServiceAccountVO) => it.name === name.value) || null
      sa.value = found
      labelRows.value = Object.entries(found?.labels || {}).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
      annotationRows.value = Object.entries(found?.annotations || {}).map(([k, v]) => ({ key: k, value: String(v ?? '') }))
      return
    }
    sa.value = null
  } catch {
    ElMessage.error('获取详情失败')
    sa.value = null
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
  fetchDetailFromList()
})
</script>

<style scoped>
.serviceaccount-detail { padding: 20px; }
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
.related-grid { display:grid; grid-template-columns: repeat(2, 1fr); gap: 16px; }
.related-title { font-weight: 600; margin-bottom: 10px; }
</style>

