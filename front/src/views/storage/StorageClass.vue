<template>
  <div class="storageclass-page">
    <div class="page-header">
      <h1>StorageClass管理</h1>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="名称关键字" clearable style="width: 220px" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <el-table v-loading="loading" :data="items" stripe style="width: 100%">
        <el-table-column prop="name" label="名称" min-width="160" />
        <el-table-column prop="provisioner" label="Provisioner" min-width="220" />
        <el-table-column prop="reclaimPolicy" label="ReclaimPolicy" min-width="140" />
        <el-table-column prop="volumeBindingMode" label="BindingMode" min-width="140" />
        <el-table-column prop="allowVolumeExpansion" label="扩容" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.allowVolumeExpansion ? 'success' : 'info'">{{ row.allowVolumeExpansion ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" min-width="180">
          <template #default="{ row }">{{ formatDate(row.createTime || row.creationTimestamp) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="140" align="center" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="openYaml(row)">YAML</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <el-dialog v-model="yamlDialogVisible" title="查看YAML" width="90%" :close-on-click-modal="false" class="config-dialog yaml-dialog">
      <template #header>
        <div class="dialog-header">
          <div>
            <h3 class="dialog-title">查看YAML</h3>
            <div style="margin-top:4px;color:#6b7280;font-size:12px;">StorageClass 配置</div>
          </div>
        </div>
      </template>
      <div class="config-editor">
        <div class="config-content">
          <div class="yaml-view">
            <YamlEditor :model-value="yamlContent" :readonly="true" height="100%" />
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="yamlDialogVisible = false">关闭</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import { clusterApi, type ClusterVO } from '@/api/cluster'
import { storageClassApi, type StorageClassVO, type StorageClassPageQueryDTO } from '@/api/storageclass'
import { useClusterStore } from '@/stores/cluster'
import YamlEditor from '@/components/YamlEditor.vue'
import '@/assets/styles/config-editor.css'

// 集群上下文
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)

const loading = ref(false)
const items = ref<StorageClassVO[]>([])
const clusterList = ref<ClusterVO[]>([])

const searchForm = reactive<StorageClassPageQueryDTO>({
  page: 1,
  pageSize: 20,
  keyword: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const yamlDialogVisible = ref(false)
const yamlContent = ref('')

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

const fetchList = async () => {
  if (!clusterId.value) {
    items.value = []
    pagination.total = 0
    return
  }
  loading.value = true
  try {
    const resp = await storageClassApi.getList({
      clusterId: clusterId.value,
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword || ''
    })
    if (resp.code === 200) {
      const list = (resp.data as any).items || (resp.data as any).list || []
      const total = (resp.data as any).total || (resp.data as any).totalCount || 0
      items.value = list
      pagination.total = total
    }
  } catch {
    ElMessage.error('获取StorageClass列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = async () => {
  searchForm.keyword = ''
  pagination.page = 1
  await fetchList()
}

const handleSizeChange = (size: number) => {
  pagination.pageSize = size
  pagination.page = 1
  fetchList()
}

const handleCurrentChange = (page: number) => {
  pagination.page = page
  fetchList()
}

const openYaml = (row: StorageClassVO) => {
  yamlContent.value = row.yaml || ''
  yamlDialogVisible.value = true
}

// 监听集群变化
watch(clusterId, (newClusterId) => {
  if (newClusterId) {
    pagination.page = 1
    fetchList()
  } else {
    items.value = []
    pagination.total = 0
  }
}, { immediate: true })

onMounted(async () => {
  if (clusterId.value) {
    await fetchList()
  }
})
</script>

<style scoped>
.storageclass-page { padding: 20px; }
</style>
