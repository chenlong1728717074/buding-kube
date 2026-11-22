<template>
  <div v-loading="loading">
    <div class="qp-actions">
      <el-button size="small" type="primary" @click="openForm('')">
        <el-icon><Plus /></el-icon>
        {{ items.length ? '新建配额' : '设置资源配额' }}
      </el-button>
    </div>
    <el-empty v-if="!items.length" description="没有资源配额，命名空间可以使用集群的全部资源" :image-size="72" />
    <div v-for="q in items" :key="q.metadata.name" class="qp-card">
      <div class="qp-head">
        <span class="qp-name">{{ q.metadata.name }}</span>
        <span v-if="q.spec?.scopes?.length" class="qp-scope">作用范围：{{ q.spec.scopes.join('、') }}</span>
        <div class="qp-ops">
          <el-button size="small" @click="openForm(q.metadata.name)">编辑</el-button>
          <el-button size="small" type="danger" plain @click="remove(q.metadata.name)">删除</el-button>
        </div>
      </div>
      <div class="qp-grid">
        <div v-for="r in quotaUsageRows(q.status?.hard || q.spec?.hard, q.status?.used)" :key="r.key" class="qp-item">
          <div class="qp-item__head">
            <span>{{ r.label }}</span>
            <span class="mono">{{ r.used }} / {{ r.hard }}</span>
          </div>
          <div class="qp-bar"><i :class="percentLevel(r.percent)" :style="{ width: `${Math.min(r.percent ?? 0, 100)}%` }" /></div>
          <span class="qp-key mono">{{ r.key }}</span>
        </div>
      </div>
    </div>
    <QuotaFormDialog v-model="formVisible" :cluster-id="clusterId" :namespace="namespace" :edit-name="editName" @success="load" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import QuotaFormDialog from '@/components/policy/QuotaFormDialog.vue'
import { resourceApi, type KubeObject } from '@/api/resource'
import { quotaUsageRows } from '@/utils/policy'
import { percentLevel } from '@/utils/node'

const props = defineProps<{ clusterId: string; namespace: string }>()
const emit = defineEmits<{ change: [count: number] }>()

const QUOTA_REF = { kind: 'resourcequota' as const }
const loading = ref(false)
const items = ref<KubeObject[]>([])
const formVisible = ref(false)
const editName = ref('')

const load = async () => {
  loading.value = true
  try {
    const { data } = await resourceApi.objects({ ...QUOTA_REF, clusterId: props.clusterId, namespace: props.namespace })
    items.value = data || []
  } catch {
    items.value = []
  } finally {
    loading.value = false
    emit('change', items.value.length)
  }
}

const openForm = (name: string) => {
  editName.value = name
  formVisible.value = true
}

const remove = async (name: string) => {
  try {
    await ElMessageBox.confirm(`确定要删除资源配额 "${name}" 吗？`, '确认删除', { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' })
  } catch {
    return
  }
  try {
    await resourceApi.delete({ ...QUOTA_REF, clusterId: props.clusterId, namespace: props.namespace, name })
    ElMessage.success('删除成功')
    load()
  } catch {
    // 请求拦截器已提示错误
  }
}

onMounted(load)
</script>

<style scoped>
.qp-actions {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
}

.qp-card {
  padding: 14px 16px;
  border: 1px solid #e6ecf5;
  border-radius: 12px;
  background: #fbfcfe;
}

.qp-card + .qp-card {
  margin-top: 12px;
}

.qp-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.qp-name {
  color: #1e293b;
  font-weight: 600;
}

.qp-scope {
  color: #64748b;
  font-size: 12px;
}

.qp-ops {
  margin-left: auto;
}

.qp-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.qp-item {
  padding: 10px 12px;
  border-radius: 8px;
  background: #fff;
}

.qp-item__head {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 6px;
  color: #334155;
  font-size: 13px;
}

.qp-bar {
  height: 6px;
  border-radius: 3px;
  background: #e8eef8;
  overflow: hidden;
}

.qp-bar i {
  display: block;
  height: 100%;
  background: #3b74f2;
}

.qp-bar i.is-warning {
  background: #f59e0b;
}

.qp-bar i.is-danger {
  background: #ef4444;
}

.qp-key {
  display: block;
  margin-top: 4px;
  color: #94a3b8;
  font-size: 11px;
}

.mono {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}

@media (max-width: 1200px) {
  .qp-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
