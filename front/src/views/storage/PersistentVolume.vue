<template>
  <div class="pv-page">
    <div class="page-header">
      <h1>PersistentVolume管理</h1>
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
        <el-table-column prop="name" label="名称" min-width="180" />
        <el-table-column label="容量" min-width="100" align="center" header-align="center">
          <template #default="{ row }">{{ formatCapacity(row.capacity) }}</template>
        </el-table-column>
        <el-table-column label="访问模式" min-width="140" align="center" header-align="center">
          <template #default="{ row }">{{ formatAccessModes(row.accessModes) }}</template>
        </el-table-column>
        <el-table-column prop="persistentVolumeReclaimPolicy" label="回收策略" min-width="110" align="center" header-align="center" />
        <el-table-column prop="storageClassName" label="StorageClass" min-width="140" align="center" header-align="center">
          <template #default="{ row }">{{ row.storageClassName || '-' }}</template>
        </el-table-column>
        <el-table-column prop="phase" label="状态" min-width="100" align="center" header-align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.phase)" size="small">{{ row.phase || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" min-width="170" align="center" header-align="center">
          <template #default="{ row }">{{ formatDate(row.createTime || row.creationTimestamp) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" header-align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" class="btn-main" @click="handleViewDetail(row)">详情</el-button>
              <el-dropdown trigger="click" @command="(cmd: string) => handleRowAction(cmd, row)">
                <el-button size="small" text circle class="more-btn">
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="yaml">查看YAML</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-button size="small" type="danger" plain class="btn-danger" @click="handleDelete(row)">删除</el-button>
            </div>
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

    <el-dialog v-model="yamlDialogVisible" :title="`YAML详情 - ${currentName}`" width="90%" :close-on-click-modal="false" class="config-dialog yaml-dialog">
      <template #header>
        <div class="dialog-header">
          <div>
            <h3 class="dialog-title">YAML详情</h3>
            <div style="margin-top:4px;color:#6b7280;font-size:12px;">PersistentVolume: {{ currentName }}</div>
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, MoreFilled } from '@element-plus/icons-vue'
import { pvApi, type PVVO, type PVQueryDTO } from '@/api/pv'
import { useClusterStore } from '@/stores/cluster'
import YamlEditor from '@/components/YamlEditor.vue'
import '@/assets/styles/config-editor.css'

// 集群上下文
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)

const loading = ref(false)
const items = ref<PVVO[]>([])

const searchForm = reactive<PVQueryDTO>({
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
const currentName = ref('')

const formatCapacity = (capacity?: Record<string, string>) => {
  if (!capacity) return '-'
  return capacity.storage || Object.values(capacity)[0] || '-'
}

const formatAccessModes = (modes?: string[]) => {
  if (!modes || modes.length === 0) return '-'
  return modes.join(', ')
}

const getStatusType = (phase?: string) => {
  const map: Record<string, string> = {
    Available: 'success',
    Bound: 'success',
    Released: 'warning',
    Failed: 'danger',
    Pending: 'warning'
  }
  return (map[phase] as any) || 'info'
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

const fetchList = async () => {
  if (!clusterId.value) {
    items.value = []
    pagination.total = 0
    return
  }
  loading.value = true
  try {
    const resp = await pvApi.getPVList({
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
    ElMessage.error('获取PersistentVolume列表失败')
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

const openYaml = (row: PVVO) => {
  currentName.value = row.name
  yamlContent.value = row.yaml || ''
  yamlDialogVisible.value = true
}

// 详情(暂未实现, 与全站统一)
const handleViewDetail = (row: PVVO) => {
  ElMessage.info('详情功能暂未实现')
}

// 行操作分发
const handleRowAction = (cmd: string, row: PVVO) => {
  if (cmd === 'yaml') {
    openYaml(row)
  }
}

// 删除PersistentVolume
const handleDelete = (row: PVVO) => {
  ElMessageBox.confirm(
    `确定要删除 PersistentVolume "${row.name}" 吗？此操作不可恢复。`,
    '确认删除',
    {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      confirmButtonClass: 'el-button--danger'
    }
  ).then(async () => {
    try {
      await pvApi.deletePV({ clusterId: clusterId.value, name: row.name })
      ElMessage.success('删除成功')
      fetchList()
    } catch (error) {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
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
.pv-page { padding: 20px; }
.search-card {
  margin-bottom: 16px;
}
.action-buttons {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  flex-wrap: nowrap;
}

.btn-main {
  font-weight: 500;
}

:deep(.action-buttons .el-button) {
  height: 28px;
  padding: 0 10px;
  border-radius: 8px;
}
</style>
