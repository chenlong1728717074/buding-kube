<template>
  <div class="pvc-page list-page">
    <PageHeader title="PersistentVolumeClaim管理" subtitle="管理命名空间内的存储申请" icon="pvc">
      <el-button type="primary" @click="createVisible = true">
        <el-icon><Plus /></el-icon>
        添加持久卷声明
      </el-button>
      <el-button type="success" @click="openYamlAdd">
        <el-icon><Document /></el-icon>
        YAML添加
      </el-button>
    </PageHeader>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="命名空间">
          <el-select v-model="searchForm.namespace" placeholder="全部命名空间" clearable filterable style="width: 200px" @change="handleSearch">
            <el-option v-for="ns in namespaceList" :key="ns" :label="ns" :value="ns" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="名称关键字" clearable style="width: 200px" @keyup.enter="handleSearch" />
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
              <span v-if="row.annotations?.alias" class="res-name__alias">{{ row.annotations.alias }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" min-width="130" align="center" header-align="center" />
        <el-table-column label="容量" min-width="100" align="center" header-align="center">
          <template #default="{ row }">{{ row.capacity?.storage || requestedCapacity(row) }}</template>
        </el-table-column>
        <el-table-column label="访问模式" min-width="120" align="center" header-align="center">
          <template #default="{ row }">{{ formatAccessModes(row.accessModes) }}</template>
        </el-table-column>
        <el-table-column prop="storageClassName" label="StorageClass" min-width="140" align="center" header-align="center">
          <template #default="{ row }">
            <el-link v-if="row.storageClassName" type="primary" @click="goStorageClass(row.storageClassName)">{{ row.storageClassName }}</el-link>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="volumeName" label="绑定Volume" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ row.volumeName || '-' }}</template>
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

    <PersistentVolumeClaimCreateDialog v-model="createVisible" :cluster-id="clusterId" :default-namespace="searchForm.namespace" @success="fetchList" />

    <YamlDialog
      v-model="yamlDialogVisible"
      title="编辑YAML"
      :subtitle="`PersistentVolumeClaim: ${currentNamespace}/${currentName}`"
      :content="yamlContent"
      :apply-target="currentName ? { clusterId, kind: 'persistentvolumeclaim', namespace: currentNamespace, name: currentName } : null"
      :filename="`${currentName}.yaml`"
      tip="已绑定的 PVC 只能扩大容量（需要存储类允许卷扩展），其余规格不能修改"
      @applied="fetchList"
    />

    <YamlDialog
      v-model="yamlAddVisible"
      v-model:content="yamlAddContent"
      title="YAML添加持久卷声明"
      subtitle="已内置示例，修改后点击应用"
      :readonly="false"
      editor-title="PersistentVolumeClaim YAML"
      filename="pvc.yaml"
      :submitting="yamlAddLoading"
      @confirm="confirmYamlAdd"
    />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, Document } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import PersistentVolumeClaimCreateDialog from '@/components/storage/PersistentVolumeClaimCreateDialog.vue'
import { fetchResourceYaml, resourceApi } from '@/api/resource'
import { namespaceApi } from '@/api/namespace'
import { pvcApi, type PVCVO } from '@/api/pvc'
import { useClusterStore } from '@/stores/cluster'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { defaultPvcYaml, formatAccessModes } from '@/utils/storageclass'
import '@/assets/styles/config-editor.css'

const router = useRouter()
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)

const loading = ref(false)
const items = ref<PVCVO[]>([])
const namespaceList = ref<string[]>([])
const searchForm = reactive({ namespace: '', keyword: '' })
const pagination = reactive({ page: 1, pageSize: 10, total: 0 })

const createVisible = ref(false)

const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const currentName = ref('')
const currentNamespace = ref('')

const yamlAddVisible = ref(false)
const yamlAddContent = ref('')
const yamlAddLoading = ref(false)

// 未绑定时 status.capacity 为空, 显示申请的容量
const requestedCapacity = (row: PVCVO & { resources?: { requests?: Record<string, string> } }) =>
  row.resources?.requests?.storage ? `${row.resources.requests.storage}（申请）` : '-'

const getStatusType = (phase?: string) => {
  const map: Record<string, 'success' | 'warning' | 'danger'> = { Bound: 'success', Pending: 'warning', Lost: 'danger' }
  return (phase && map[phase]) || 'info'
}

const formatDate = (dateString?: string) => {
  if (!dateString) return '-'
  const d = new Date(dateString)
  if (isNaN(d.getTime())) return '-'
  return d.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

const fetchNamespaces = async () => {
  if (!clusterId.value) {
    namespaceList.value = []
    return
  }
  try {
    const resp = await namespaceApi.getList({ clusterId: clusterId.value, page: 1, pageSize: 10000 })
    namespaceList.value = (resp.data?.items || []).map(ns => ns.name)
  } catch {
    namespaceList.value = []
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
    const resp = await pvcApi.getPVCList({
      clusterId: clusterId.value,
      namespace: searchForm.namespace || undefined,
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
  searchForm.namespace = ''
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

const openYaml = async (row: PVCVO) => {
  let yamlText = ''
  try {
    yamlText = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'persistentvolumeclaim', namespace: row.namespace, name: row.name })
  } catch {
    return
  }
  currentName.value = row.name
  currentNamespace.value = row.namespace
  yamlContent.value = yamlText
  yamlDialogVisible.value = true
}

const openYamlAdd = () => {
  yamlAddContent.value = defaultPvcYaml(searchForm.namespace || 'default')
  yamlAddVisible.value = true
}

const confirmYamlAdd = async () => {
  if (!yamlAddContent.value.trim()) {
    ElMessage.warning('YAML内容不能为空')
    return
  }
  yamlAddLoading.value = true
  try {
    await resourceApi.apply({ clusterId: clusterId.value, kind: 'persistentvolumeclaim', namespace: searchForm.namespace || undefined, yaml: yamlAddContent.value })
    ElMessage.success('应用成功')
    yamlAddVisible.value = false
    fetchList()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    yamlAddLoading.value = false
  }
}

const handleDelete = (row: PVCVO) => {
  ElMessageBox.confirm(
    `确定要删除 PersistentVolumeClaim "${row.namespace}/${row.name}" 吗？正在使用它的 Pod 不受影响，但 Pod 删除后数据可能随持久卷一起删除（取决于回收策略）。`,
    '确认删除',
    { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' }
  )
    .then(async () => {
      try {
        await pvcApi.deletePVC({ clusterId: clusterId.value, namespace: row.namespace, name: row.name })
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
      searchForm.namespace = ''
      pagination.page = 1
      fetchNamespaces()
      fetchList()
    } else {
      namespaceList.value = []
      items.value = []
      pagination.total = 0
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.pvc-page {
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
