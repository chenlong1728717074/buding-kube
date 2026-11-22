<template>
  <div class="storageclass-detail detail-page">
    <PageHeader
      :title="sc?.name || name || '存储类详情'"
      :subtitle="sc?.alias || sc?.describe || '存储类详情'"
      icon="sc"
      back
      @back="handleBack"
    >
      <template v-if="sc" #meta>
        <el-tag v-if="sc.isDefault" type="primary" effect="dark">默认存储类</el-tag>
        <el-tag :type="typeTag(sc.storageTypeKey)" effect="light">{{ sc.storageType || '未分类' }}</el-tag>
      </template>
      <template v-if="sc">
        <el-button @click="editVisible = true">
          <el-icon><Edit /></el-icon>
          编辑
        </el-button>
        <el-button @click="openScYaml">
          <el-icon><Document /></el-icon>
          编辑YAML
        </el-button>
        <el-button @click="toggleDefault">
          <el-icon><Star /></el-icon>
          {{ sc.isDefault ? '取消默认' : '设为默认' }}
        </el-button>
        <el-button type="danger" @click="handleDelete">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </template>
    </PageHeader>

    <div v-if="loading && !sc" class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <div v-else-if="sc" v-loading="loading" class="detail-content">
      <!-- 基本信息 -->
      <el-card class="info-card">
        <template #header>
          <div class="card-header"><span>基本信息</span></div>
        </template>
        <div class="info-grid">
          <div class="info-item"><label>名称</label><span>{{ sc.name }}</span></div>
          <div class="info-item"><label>别名</label><span>{{ sc.alias || '-' }}</span></div>
          <div class="info-item">
            <label>存储类型</label>
            <span>{{ sc.storageType || '未分类' }}<em v-if="sc.storageTypeAuto" class="sc-auto">自动识别</em></span>
          </div>
          <div class="info-item"><label>供应者</label><span class="mono">{{ sc.provisioner || '-' }}</span></div>
          <div class="info-item"><label>回收策略</label><span>{{ reclaimText(sc.reclaimPolicy) }}</span></div>
          <div class="info-item"><label>卷绑定模式</label><span>{{ bindingText(sc.volumeBindingMode) }}</span></div>
          <div class="info-item"><label>允许卷扩展</label><span>{{ sc.allowVolumeExpansion ? '是' : '否' }}</span></div>
          <div class="info-item"><label>允许卷克隆</label><span>{{ sc.allowVolumeClone ? '是' : '否' }}</span></div>
          <div class="info-item"><label>支持的访问模式</label><span>{{ sc.accessModes?.length ? formatAccessModes(sc.accessModes) : '不限制' }}</span></div>
          <div class="info-item"><label>挂载选项</label><span class="mono">{{ sc.mountOptions?.length ? sc.mountOptions.join(', ') : '-' }}</span></div>
          <div class="info-item"><label>持久卷声明数量</label><span>{{ claims.length }}</span></div>
          <div class="info-item"><label>持久卷数量</label><span>{{ sc.pvCount ?? 0 }}</span></div>
          <div class="info-item"><label>创建时间</label><span>{{ formatDateTime(sc.creationTimestamp || sc.createTime) }}</span></div>
          <div class="info-item"><label>描述</label><span>{{ sc.describe || '-' }}</span></div>
        </div>
      </el-card>

      <el-row :gutter="18">
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>参数</span>
                <el-tag size="small" type="info">{{ paramRows.length }}</el-tag>
              </div>
            </template>
            <div class="labels-content">
              <el-empty v-if="!paramRows.length" description="没有参数" :image-size="60" />
              <div v-for="row in paramRows" v-else :key="row.key" class="label-item">
                <div class="label-key">{{ row.key }}</div>
                <div class="label-value">{{ row.value || '""' }}</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card class="info-card">
            <template #header>
              <div class="card-header">
                <span>注解与标签</span>
                <el-tag size="small" type="info">{{ metaRows.length }}</el-tag>
              </div>
            </template>
            <div class="labels-content">
              <el-empty v-if="!metaRows.length" description="没有注解和标签" :image-size="60" />
              <div v-for="row in metaRows" v-else :key="row.type + row.key" class="label-item">
                <div class="label-key">
                  <span class="meta-type">{{ row.type }}</span>
                  {{ row.key }}
                </div>
                <div class="label-value">{{ row.value || '""' }}</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 持久卷声明 -->
      <el-card class="info-card">
        <template #header>
          <div class="card-header">
            <span>持久卷声明</span>
            <el-tag size="small" type="info">{{ claims.length }}</el-tag>
            <span class="card-header__hint">使用该存储类的持久卷声明（PVC）</span>
            <el-button class="card-header__action" size="small" type="primary" @click="pvcCreateVisible = true">
              <el-icon><Plus /></el-icon>
              创建持久卷声明
            </el-button>
          </div>
        </template>
        <el-table :data="claimRows" style="width: 100%" :header-cell-class-name="tableHeaderClass" :cell-class-name="tableCellClass">
          <el-table-column prop="name" label="名称" min-width="220" show-overflow-tooltip>
            <template #default="{ row }">
              <div class="pv-name">
                <span class="pv-name__main">{{ row.name }}</span>
                <span v-if="row.alias" class="pv-name__alias">{{ row.alias }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="命名空间" min-width="140" show-overflow-tooltip>
            <template #default="{ row }">
              <el-link type="primary" @click="goNamespace(row.namespace)">{{ row.namespace }}</el-link>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="110" align="center" header-align="center">
            <template #default="{ row }">
              <el-tag :type="phaseTag(row.phase)" size="small">{{ phaseText(row.phase) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="容量" width="110" align="center" header-align="center">
            <template #default="{ row }">{{ row.capacity?.storage || row.resources?.requests?.storage || '-' }}</template>
          </el-table-column>
          <el-table-column label="访问模式" width="140" align="center" header-align="center">
            <template #default="{ row }">{{ formatAccessModes(row.accessModes) }}</template>
          </el-table-column>
          <el-table-column label="绑定的持久卷" min-width="220" show-overflow-tooltip>
            <template #default="{ row }">
              <span v-if="row.volumeName" class="mono">{{ row.volumeName }}</span>
              <el-text v-else type="info">{{ sc.volumeBindingMode === 'WaitForFirstConsumer' ? '等待首次使用' : '未绑定' }}</el-text>
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="170" align="center" header-align="center">
            <template #default="{ row }">{{ formatDateTime(row.createTime || row.creationTimestamp) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="120" align="center" header-align="center" fixed="right">
            <template #default="{ row }">
              <el-button size="small" link type="primary" @click="openPvcYaml(row)">编辑YAML</el-button>
            </template>
          </el-table-column>
          <template #empty>
            <el-empty description="还没有使用该存储类的持久卷声明" :image-size="72" />
          </template>
        </el-table>
        <ListPagination
          v-if="claims.length > 10"
          v-model:current-page="claimPage"
          v-model:page-size="claimPageSize"
          style="margin-top: 12px"
          :total="claimTotal"
          :page-sizes="[10, 20, 50, 100]"
        />
      </el-card>
    </div>

    <el-empty v-else description="存储类不存在或已被删除" />

    <StorageClassEditDialog v-model="editVisible" :cluster-id="clusterId" :storage-class="sc" @success="fetchDetail" />
    <PersistentVolumeClaimCreateDialog v-model="pvcCreateVisible" :cluster-id="clusterId" :default-storage-class="sc?.name" @success="fetchDetail" />
    <YamlDialog
      v-model="yamlVisible"
      :content="yamlContent"
      :title="yamlTitle"
      :subtitle="yamlTarget?.name"
      :apply-target="yamlTarget"
      :filename="`${yamlTarget?.name || 'resource'}.yaml`"
      @applied="fetchDetail"
    />
  </div>
</template>

<script setup lang="ts">
import { formatDateTime } from '@/utils/format'
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Document, Edit, Plus, Star } from '@element-plus/icons-vue'
import PageHeader from '@/components/PageHeader.vue'
import YamlDialog from '@/components/YamlDialog.vue'
import StorageClassEditDialog from '@/components/storage/StorageClassEditDialog.vue'
import PersistentVolumeClaimCreateDialog from '@/components/storage/PersistentVolumeClaimCreateDialog.vue'
import ListPagination from '@/components/ListPagination.vue'
import { storageClassApi, type StorageClassDetailVO } from '@/api/storageclass'
import { fetchResourceYaml, type ResourceYamlTarget } from '@/api/resource'
import type { PVCVO } from '@/api/pvc'
import { useClientPage } from '@/composables/useClientPage'
import { useClusterStore } from '@/stores/cluster'
import { tableHeaderClass, tableCellClass } from '@/utils/table'
import { SC_ANNOTATIONS, formatAccessModes, mapToRows } from '@/utils/storageclass'
import { useTabBack } from '@/composables/useTabBack'

const route = useRoute()
const router = useRouter()
// 返回当前标签内的上一页
const goBack = useTabBack()
const clusterStore = useClusterStore()

// 页面被 keep-alive 按完整路径缓存, 参数在创建时取一次, 避免切到其他页面后用新页面的参数发请求
const clusterId = ref((route.query.clusterId as string) || clusterStore.currentClusterId || '')
const name = ref((route.query.name as string) || '')

const loading = ref(false)
const sc = ref<StorageClassDetailVO | null>(null)
const claims = computed<PVCVO[]>(() => sc.value?.persistentVolumeClaims || [])
const { page: claimPage, pageSize: claimPageSize, total: claimTotal, rows: claimRows } = useClientPage(claims)

const editVisible = ref(false)
const pvcCreateVisible = ref(false)

const yamlVisible = ref(false)
const yamlContent = ref('')
const yamlTarget = ref<ResourceYamlTarget | null>(null)
const yamlTitle = computed(() => (yamlTarget.value?.kind === 'persistentvolumeclaim' ? '编辑持久卷声明 YAML' : '编辑存储类 YAML'))

// 别名 / 描述 / 存储类型等已在基本信息里展示, 这里不重复
const HIDDEN_ANNOTATIONS = new Set(['alias', 'describe', SC_ANNOTATIONS.storageType, SC_ANNOTATIONS.allowClone, SC_ANNOTATIONS.accessModes])

const paramRows = computed(() => mapToRows(sc.value?.parameters))
const metaRows = computed(() => [
  ...mapToRows(sc.value?.annotations)
    .filter(r => !HIDDEN_ANNOTATIONS.has(r.key))
    .map(r => ({ ...r, type: '注解' })),
  ...mapToRows(sc.value?.labels).map(r => ({ ...r, type: '标签' }))
])

const fetchDetail = async () => {
  if (!clusterId.value || !name.value) return
  loading.value = true
  try {
    const resp = await storageClassApi.getDetail({ clusterId: clusterId.value, name: name.value })
    sc.value = resp.code === 200 ? resp.data : null
  } catch {
    sc.value = null
  } finally {
    loading.value = false
  }
}

const handleBack = goBack

const openYaml = async (target: ResourceYamlTarget) => {
  try {
    yamlContent.value = await fetchResourceYaml(target)
  } catch {
    return
  }
  yamlTarget.value = target
  yamlVisible.value = true
}

const openScYaml = () => openYaml({ clusterId: clusterId.value, kind: 'storageclass', name: name.value })
const openPvcYaml = (row: PVCVO) => openYaml({ clusterId: clusterId.value, kind: 'persistentvolumeclaim', namespace: row.namespace, name: row.name })

const goNamespace = (namespace: string) => router.push({ path: `/cluster/${clusterId.value}/namespace/detail`, query: { namespace } })

const toggleDefault = async () => {
  if (!sc.value) return
  const makeDefault = !sc.value.isDefault
  try {
    await ElMessageBox.confirm(
      makeDefault
        ? `将 "${sc.value.name}" 设为默认存储类后，未指定存储类的 PVC 会使用它，同时会取消其他存储类的默认标记。`
        : '取消后，集群将没有默认存储类，未指定存储类的 PVC 会一直处于等待状态。',
      makeDefault ? '设为默认存储类' : '取消默认存储类',
      { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await storageClassApi.update({
      clusterId: clusterId.value,
      name: sc.value.name,
      alias: sc.value.alias,
      describe: sc.value.describe,
      storageType: sc.value.storageTypeAuto ? '' : sc.value.storageTypeKey,
      isDefault: makeDefault
    })
    ElMessage.success('设置成功')
    fetchDetail()
  } catch {
    // 请求拦截器已提示错误
  }
}

const handleDelete = async () => {
  if (!sc.value) return
  const inUse = claims.value.length || sc.value.pvCount
  try {
    await ElMessageBox.confirm(
      inUse
        ? `存储类 "${sc.value.name}" 还有 ${claims.value.length} 个持久卷声明、${sc.value.pvCount ?? 0} 个持久卷在使用。删除后已有的卷不受影响，但不能再用它创建新卷。确定删除吗？`
        : `确定要删除存储类 "${sc.value.name}" 吗？此操作不可恢复。`,
      '确认删除',
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' }
    )
  } catch {
    return
  }
  try {
    await storageClassApi.deleteStorageClass({ clusterId: clusterId.value, name: sc.value.name })
    ElMessage.success('删除成功')
    router.push(`/cluster/${clusterId.value}/storage/storageclass`)
  } catch {
    // 请求拦截器已提示错误
  }
}

const typeTag = (key?: string) =>
  (({ block: 'primary', file: 'success', object: 'warning' }) as Record<string, 'primary' | 'success' | 'warning'>)[key || ''] || 'info'

const reclaimText = (policy?: string, short = false) => {
  const map: Record<string, string> = { Delete: '删除', Retain: '保留', Recycle: '回收' }
  if (!policy) return '-'
  return short ? map[policy] || policy : `${map[policy] || policy}（${policy}）`
}

const bindingText = (mode?: string) =>
  mode === 'WaitForFirstConsumer' ? '首次使用时绑定' : mode === 'Immediate' ? '立即绑定' : mode || '-'

const phaseText = (phase?: string) =>
  ({ Available: '可用', Bound: '已绑定', Released: '已释放', Failed: '失败', Pending: '等待中', Lost: '丢失' } as Record<string, string>)[phase || ''] || phase || '-'

const phaseTag = (phase?: string) =>
  (({ Available: 'primary', Bound: 'success', Released: 'warning', Failed: 'danger', Pending: 'warning', Lost: 'danger' }) as Record<string, 'primary' | 'success' | 'warning' | 'danger'>)[phase || ''] || 'info'

onMounted(fetchDetail)
</script>

<style scoped>
.storageclass-detail {
  padding: 20px;
}

.mono {
  font-family: 'SF Mono', Consolas, Monaco, monospace;
  font-size: 13px;
}

.sc-auto {
  margin-left: 8px;
  padding: 1px 8px;
  border-radius: 999px;
  background: #f1f5f9;
  color: #64748b;
  font-size: 12px;
  font-style: normal;
  font-weight: 400;
}

.labels-content {
  max-height: 320px;
  overflow-y: auto;
}

.meta-type {
  margin-right: 6px;
  padding: 0 6px;
  border-radius: 4px;
  background: #eaf2ff;
  color: #2563eb;
  font-family: inherit;
  font-size: 11px;
  font-weight: 500;
}

.card-header__hint {
  font-size: 12px;
  font-weight: 400;
  color: #94a3b8;
}

.card-header__action {
  margin-left: auto;
  border-radius: 999px !important;
}

.pv-name {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.pv-name__main {
  overflow: hidden;
  color: #1e293b;
  font-weight: 500;
  text-overflow: ellipsis;
}

.pv-name__alias {
  font-size: 12px;
  color: #94a3b8;
}
</style>
