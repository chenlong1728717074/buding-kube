<template>
  <div class="storageclass-page">
    <div class="page-header">
      <h1>StorageClass管理</h1>
    </div>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="名称关键字" clearable style="width: 200px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="存储类型">
          <el-select v-model="searchForm.storageType" placeholder="全部类型" clearable style="width: 140px" @change="handleSearch">
            <el-option label="块存储" value="块存储" />
            <el-option label="对象存储" value="对象存储" />
            <el-option label="文件存储" value="文件存储" />
            <el-option label="未分类" value="未分类" />
          </el-select>
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
        <el-table-column label="存储类型" min-width="110" align="center" header-align="center">
          <template #default="{ row }">
            <el-tag :type="storageTypeTag(row.storageType)" size="small">{{ row.storageType || '未分类' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="pvcCount" label="持久卷声明数量" min-width="130" align="center" header-align="center">
          <template #default="{ row }">{{ row.pvcCount ?? 0 }}</template>
        </el-table-column>
        <el-table-column label="允许卷克隆" min-width="110" align="center" header-align="center">
          <template #default="{ row }">
            <el-tag :type="row.allowVolumeClone ? 'success' : 'info'" size="small">{{ row.allowVolumeClone ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="允许卷扩展" min-width="110" align="center" header-align="center">
          <template #default="{ row }">
            <el-tag :type="row.allowVolumeExpansion ? 'success' : 'info'" size="small">{{ row.allowVolumeExpansion ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="provisioner" label="供应者" min-width="200" show-overflow-tooltip />
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, MoreFilled } from '@element-plus/icons-vue'
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
  keyword: '',
  storageType: ''
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
      keyword: searchForm.keyword || '',
      storageType: searchForm.storageType || undefined
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

// 删除StorageClass
const handleDelete = (row: StorageClassVO) => {
  ElMessageBox.confirm(
    `确定要删除 StorageClass "${row.name}" 吗？此操作不可恢复。`,
    '确认删除',
    {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      confirmButtonClass: 'el-button--danger'
    }
  ).then(async () => {
    try {
      await storageClassApi.deleteStorageClass({ clusterId: clusterId.value, name: row.name })
      ElMessage.success('删除成功')
      fetchList()
    } catch (error) {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }).catch(() => {})
}

const handleReset = async () => {
  searchForm.keyword = ''
  searchForm.storageType = ''
  pagination.page = 1
  await fetchList()
}

// 存储类型标签颜色
const storageTypeTag = (storageType?: string) => {
  const map: Record<string, string> = {
    块存储: 'primary',
    对象存储: 'warning',
    文件存储: 'success'
  }
  return (map[storageType || ''] as any) || 'info'
}

// 详情(暂未实现, 与全站统一)
const handleViewDetail = (row: StorageClassVO) => {
  ElMessage.info('详情功能暂未实现')
}

// 行操作分发
const handleRowAction = (cmd: string, row: StorageClassVO) => {
  if (cmd === 'yaml') {
    openYaml(row)
  }
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
