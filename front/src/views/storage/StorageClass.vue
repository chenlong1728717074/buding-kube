<template>
  <div class="storageclass-page list-page">
    <PageHeader title="StorageClass管理" subtitle="管理存储类型与动态供给策略" icon="sc">
      <el-button type="primary" @click="createVisible = true">
        <el-icon><Plus /></el-icon>
        添加存储类
      </el-button>
      <el-button type="success" @click="openYamlAdd">
        <el-icon><Document /></el-icon>
        YAML添加
      </el-button>
    </PageHeader>

    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="关键词">
          <el-input v-model="searchForm.keyword" placeholder="名称关键字" clearable style="width: 200px" @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="存储类型">
          <el-select v-model="searchForm.storageType" placeholder="全部类型" clearable style="width: 140px" @change="handleSearch">
            <el-option v-for="opt in STORAGE_TYPE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
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
      <el-table v-loading="loading" :data="items" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
        <el-table-column prop="name" label="名称" min-width="200">
          <template #default="{ row }">
            <div class="sc-name">
              <div class="sc-name__main">
                <el-link type="primary" @click="openDetail(row)">{{ row.name }}</el-link>
                <span v-if="row.isDefault" class="sc-default-tag">默认</span>
              </div>
              <div v-if="row.alias" class="sc-name__alias">{{ row.alias }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="存储类型" min-width="110" align="center" header-align="center">
          <template #default="{ row }">
            <el-tooltip :disabled="!row.storageTypeAuto" content="根据供应者自动识别，可在编辑中手动设置" placement="top">
              <el-tag :type="storageTypeTag(row.storageTypeKey)" size="small">{{ row.storageType || '未分类' }}</el-tag>
            </el-tooltip>
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
          <template #default="{ row }">{{ formatDateTime(row.createTime || row.creationTimestamp) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" header-align="center" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button size="small" type="primary" class="btn-main" @click="openDetail(row)">详情</el-button>
              <el-dropdown trigger="click" @command="(cmd: string) => handleRowAction(cmd, row)">
                <el-button size="small" text circle class="more-btn">
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="edit">编辑</el-dropdown-item>
                    <el-dropdown-item command="yaml">编辑YAML</el-dropdown-item>
                    <el-dropdown-item command="default">{{ row.isDefault ? '取消默认存储类' : '设为默认存储类' }}</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
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

    <StorageClassCreateDialog v-model="createVisible" :cluster-id="clusterId" @success="fetchList" />
    <StorageClassEditDialog v-model="editVisible" :cluster-id="clusterId" :storage-class="currentRow" @success="fetchList" />

    <YamlDialog
      v-model="yamlDialogVisible"
      title="编辑YAML"
      :subtitle="`StorageClass: ${yamlName}`"
      :content="yamlContent"
      :apply-target="yamlName ? { clusterId, kind: 'storageclass', name: yamlName } : null"
      :filename="`${yamlName}.yaml`"
      tip="供应者、参数、回收策略、绑定模式创建后不能修改"
      @applied="fetchList"
    />

    <YamlDialog
      v-model="yamlAddVisible"
      v-model:content="yamlAddContent"
      title="YAML添加存储类"
      subtitle="已内置 NFS 示例，修改后点击应用"
      :readonly="false"
      editor-title="StorageClass YAML"
      filename="storageclass.yaml"
      :submitting="yamlAddLoading"
      @confirm="confirmYamlAdd"
    />
  </div>
</template>

<script setup lang="ts">
import { formatDateTime } from '@/utils/format'
import { fetchResourceYaml } from '@/api/resource'
import PageHeader from '@/components/PageHeader.vue'
import ListPagination from '@/components/ListPagination.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import StorageClassCreateDialog from '@/components/storage/StorageClassCreateDialog.vue'
import StorageClassEditDialog from '@/components/storage/StorageClassEditDialog.vue'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { STORAGE_TYPE_OPTIONS, DEFAULT_STORAGECLASS_YAML } from '@/utils/storageclass'
import { ref, reactive, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, MoreFilled, Plus, Document } from '@element-plus/icons-vue'
import { storageClassApi, type StorageClassVO } from '@/api/storageclass'
import { useClusterStore } from '@/stores/cluster'
import '@/assets/styles/config-editor.css'

// 集群上下文
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)

const loading = ref(false)
const items = ref<StorageClassVO[]>([])

const searchForm = reactive({
  keyword: '',
  storageType: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const createVisible = ref(false)
const editVisible = ref(false)
const currentRow = ref<StorageClassVO | null>(null)

const yamlDialogVisible = ref(false)
const yamlContent = ref('')
const yamlName = ref('')
const router = useRouter()

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
    const resp = await storageClassApi.getList({
      clusterId: clusterId.value,
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchForm.keyword || '',
      storageType: searchForm.storageType || undefined
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
  searchForm.storageType = ''
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

// 存储类型标签颜色
const storageTypeTag = (key?: string) => {
  const map: Record<string, 'primary' | 'warning' | 'success'> = {
    block: 'primary',
    object: 'warning',
    file: 'success'
  }
  return map[key || ''] || 'info'
}

const openDetail = (row: StorageClassVO) => {
  router.push({ path: `/cluster/${clusterId.value}/storage/storageclass/detail`, query: { clusterId: clusterId.value, name: row.name } })
}

const openEdit = (row: StorageClassVO) => {
  currentRow.value = row
  editVisible.value = true
}

// 行操作分发
const handleRowAction = (cmd: string, row: StorageClassVO) => {
  if (cmd === 'yaml') {
    openYaml(row)
  } else if (cmd === 'edit') {
    openEdit(row)
  } else if (cmd === 'default') {
    toggleDefault(row)
  }
}

const toggleDefault = async (row: StorageClassVO) => {
  const makeDefault = !row.isDefault
  try {
    await ElMessageBox.confirm(
      makeDefault
        ? `将 "${row.name}" 设为默认存储类后，未指定存储类的 PVC 会使用它，同时会取消其他存储类的默认标记。`
        : `取消后，集群将没有默认存储类，未指定存储类的 PVC 会一直处于等待状态。`,
      makeDefault ? '设为默认存储类' : '取消默认存储类',
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await storageClassApi.update({
      clusterId: clusterId.value,
      name: row.name,
      alias: row.alias,
      describe: row.describe,
      storageType: row.storageTypeAuto ? '' : row.storageTypeKey,
      isDefault: makeDefault
    })
    ElMessage.success('设置成功')
    fetchList()
  } catch {
    // 请求拦截器已提示错误
  }
}

const openYaml = async (row: StorageClassVO) => {
  let yamlText = ''
  try {
    yamlText = await fetchResourceYaml({ clusterId: clusterId.value, kind: 'storageclass', name: row.name })
  } catch {
    return
  }
  yamlName.value = row.name
  yamlContent.value = yamlText
  yamlDialogVisible.value = true
}

const openYamlAdd = () => {
  yamlAddContent.value = DEFAULT_STORAGECLASS_YAML
  yamlAddVisible.value = true
}

const confirmYamlAdd = async () => {
  if (!yamlAddContent.value.trim()) {
    ElMessage.warning('YAML内容不能为空')
    return
  }
  yamlAddLoading.value = true
  try {
    await storageClassApi.apply({ clusterId: clusterId.value, yaml: yamlAddContent.value })
    ElMessage.success('应用成功')
    yamlAddVisible.value = false
    fetchList()
  } catch {
    // 请求拦截器已提示错误
  } finally {
    yamlAddLoading.value = false
  }
}

// 删除StorageClass
const handleDelete = (row: StorageClassVO) => {
  ElMessageBox.confirm(
    `确定要删除 StorageClass "${row.name}" 吗？此操作不可恢复。`,
    '确认删除',
    {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    }
  ).then(async () => {
    try {
      await storageClassApi.deleteStorageClass({ clusterId: clusterId.value, name: row.name })
      ElMessage.success('删除成功')
      fetchList()
    } catch {
      // 请求拦截器已提示错误
    }
  }).catch(() => {})
}

// 集群变化时重新加载(immediate 也负责首次加载, 不需要再在 onMounted 里请求一次)
watch(clusterId, (newClusterId) => {
  if (newClusterId) {
    pagination.page = 1
    fetchList()
  } else {
    items.value = []
    pagination.total = 0
  }
}, { immediate: true })
</script>

<style scoped>
.storageclass-page { padding: 20px; }

.sc-name {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.sc-name__main {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.sc-default-tag {
  flex: none;
  padding: 0 8px;
  border-radius: 999px;
  background: linear-gradient(135deg, #5b9cff 0%, #3b74f2 100%);
  color: #ffffff;
  font-size: 12px;
  line-height: 20px;
}

.sc-name__alias {
  overflow: hidden;
  font-size: 12px;
  color: #94a3b8;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
