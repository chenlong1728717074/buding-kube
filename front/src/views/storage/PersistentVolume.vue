<template>
  <div class="pv-page list-page">
    <PageHeader title="PersistentVolume管理" subtitle="管理集群级别的持久化存储卷" icon="pv">
      <el-button type="primary" @click="createVisible = true">
        <el-icon><Plus /></el-icon>
        添加持久卷
      </el-button>
      <el-button type="success" @click="openYamlAdd">
        <el-icon><Document /></el-icon>
        YAML添加
      </el-button>
    </PageHeader>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="名称关键字" clearable style="width: 220px" @keyup.enter="handleSearch" />
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
      <el-table v-loading="loading" :data="items" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
        <el-table-column prop="name" label="名称" min-width="200">
          <template #default="{ row }">
            <div class="res-name">
              <span class="res-name__main">{{ row.name }}</span>
              <span v-if="row.alias" class="res-name__alias">{{ row.alias }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="容量" min-width="90" align="center" header-align="center">
          <template #default="{ row }">{{ row.capacity?.storage || '-' }}</template>
        </el-table-column>
        <el-table-column label="访问模式" min-width="130" align="center" header-align="center">
          <template #default="{ row }">{{ formatAccessModes(row.accessModes) }}</template>
        </el-table-column>
        <el-table-column prop="persistentVolumeReclaimPolicy" label="回收策略" min-width="100" align="center" header-align="center" />
        <el-table-column label="来源类型" min-width="100" align="center" header-align="center">
          <template #default="{ row }">{{ PV_SOURCE_LABELS[row.sourceType] || row.sourceType || '-' }}</template>
        </el-table-column>
        <el-table-column prop="storageClassName" label="StorageClass" min-width="140" align="center" header-align="center">
          <template #default="{ row }">
            <el-link v-if="row.storageClassName" type="primary" @click="goStorageClass(row.storageClassName)">{{ row.storageClassName }}</el-link>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="绑定的PVC" min-width="170" show-overflow-tooltip>
          <template #default="{ row }">{{ row.claimName ? `${row.claimNamespace}/${row.claimName}` : '-' }}</template>
        </el-table-column>
        <el-table-column prop="phase" label="状态" min-width="100" align="center" header-align="center">
          <template #default="{ row }">
            <el-tag :type="volumePhaseType(row.phase)" size="small">{{ row.phase || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" min-width="170" align="center" header-align="center">
          <template #default="{ row }">{{ formatDateTime(row.createTime || row.creationTimestamp) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" header-align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" class="btn-main" @click="openYaml(row)">编辑YAML</el-button>
              <el-button size="small" type="danger" plain class="btn-danger" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <ListPagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </el-card>

    <PersistentVolumeCreateDialog v-model="createVisible" :cluster-id="clusterId" @success="fetchList" />

    <YamlDialog
      v-model="yamlDialogVisible"
      title="编辑YAML"
      :subtitle="`PersistentVolume: ${currentName}`"
      :content="yamlContent"
      :apply-target="currentName ? { clusterId, kind: 'persistentvolume', name: currentName } : null"
      :filename="`${currentName}.yaml`"
      tip="容量只能扩大；存储来源、访问模式等字段创建后不能修改"
      @applied="fetchList"
    />

    <YamlDialog
      v-model="yamlAddVisible"
      v-model:content="yamlAddContent"
      title="YAML添加持久卷"
      subtitle="已内置 NFS 示例，修改后点击应用"
      :readonly="false"
      editor-title="PersistentVolume YAML"
      filename="pv.yaml"
      :submitting="yamlAddLoading"
      @confirm="confirmYamlAdd"
    />
  </div>
</template>

<script setup lang="ts">
import { volumePhaseType } from '@/utils/status'
import { formatDateTime } from '@/utils/format'
import { reactive, ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, Document } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import PersistentVolumeCreateDialog from '@/components/storage/PersistentVolumeCreateDialog.vue'
import { fetchResourceYaml, resourceApi } from '@/api/resource'
import { pvApi, type PVVO } from '@/api/pv'
import { useClusterStore } from '@/stores/cluster'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { DEFAULT_PV_YAML, PV_SOURCE_LABELS, formatAccessModes } from '@/utils/storageclass'
import '@/assets/styles/config-editor.css'

const router = useRouter()
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)

const loading = ref(false)
const items = ref<PVVO[]>([])
const searchForm = reactive({ keyword: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const createVisible = ref(false)

const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const currentName = ref('')

const yamlAddVisible = ref(false)
const yamlAddContent = ref('')
const yamlAddLoading = ref(false)

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
      items.value = resp.data.items || []
      pagination.total = resp.data.total || 0
    }
  } catch {
    // 请求拦截器已提示错误
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  fetchList()
}

const handleReset = () => {
  searchForm.keyword = ''
  pagination.page = 1
  fetchList()
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

const goStorageClass = (name: string) => {
  router.push({ path: `/cluster/${clusterId.value}/storage/storageclass/detail`, query: { clusterId: clusterId.value, name } })
}

const openYaml = async (row: PVVO) => {
  let yamlText = ''
  try {
    yamlText = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'persistentvolume', name: row.name })
  } catch {
    return
  }
  currentName.value = row.name
  yamlContent.value = yamlText
  yamlDialogVisible.value = true
}

const openYamlAdd = () => {
  yamlAddContent.value = DEFAULT_PV_YAML
  yamlAddVisible.value = true
}

const confirmYamlAdd = async () => {
  if (!yamlAddContent.value.trim()) {
    ElMessage.warning('YAML内容不能为空')
    return
  }
  yamlAddLoading.value = true
  try {
    await resourceApi.apply({ clusterId: clusterId.value, kind: 'persistentvolume', yaml: yamlAddContent.value })
    ElMessage.success('应用成功')
    yamlAddVisible.value = false
    fetchList()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    yamlAddLoading.value = false
  }
}

const handleDelete = (row: PVVO) => {
  ElMessageBox.confirm(
    row.claimName
      ? `持久卷 "${row.name}" 正被 ${row.claimNamespace}/${row.claimName} 使用，删除后会一直处于 Terminating，直到该 PVC 被删除。确定删除吗？`
      : `确定要删除 PersistentVolume "${row.name}" 吗？此操作不可恢复。`,
    '确认删除',
    { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' }
  )
    .then(async () => {
      try {
        await pvApi.deletePV({ clusterId: clusterId.value, name: row.name })
        ElMessage.success('删除成功')
        fetchList()
      } catch {
        // 请求拦截器已提示错误
      }
    })
    .catch(() => {})
}

watch(
  clusterId,
  value => {
    if (value) {
      pagination.page = 1
      fetchList()
    } else {
      items.value = []
      pagination.total = 0
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.pv-page {
  padding: 20px;
}

.res-name {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.res-name__main {
  overflow: hidden;
  color: #1e293b;
  font-weight: 500;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.res-name__alias {
  font-size: 12px;
  color: #94a3b8;
}
</style>
