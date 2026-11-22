<template>
  <div v-loading="loading">
    <NetworkPolicySupportAlert :cluster-id="clusterId" />
    <div class="npp-actions">
      <el-button size="small" type="primary" @click="openForm('')">
        <el-icon><Plus /></el-icon>
        创建网络策略
      </el-button>
    </div>
    <el-table :data="items" style="width: 100%" :header-cell-class-name="tableHeaderClass">
      <el-table-column label="名称" min-width="180">
        <template #default="{ row }"><span class="npp-name">{{ row.metadata.name }}</span></template>
      </el-table-column>
      <el-table-column label="应用于" min-width="200">
        <template #default="{ row }"><span class="mono">{{ podSelectorText(row.spec) }}</span></template>
      </el-table-column>
      <el-table-column label="入站" min-width="120">
        <template #default="{ row }">{{ directionText(row.spec, 'Ingress') }}</template>
      </el-table-column>
      <el-table-column label="出站" min-width="120">
        <template #default="{ row }">{{ directionText(row.spec, 'Egress') }}</template>
      </el-table-column>
      <el-table-column label="创建时间" min-width="160">
        <template #default="{ row }">{{ formatDateTime(row.metadata.creationTimestamp) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openForm(row.metadata.name)">编辑</el-button>
          <el-button link type="danger" @click="remove(row.metadata.name)">删除</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="没有网络策略，容器组之间的网络不受限制" :image-size="72" />
      </template>
    </el-table>
    <NetworkPolicyFormDialog v-model="formVisible" :cluster-id="clusterId" :namespace="namespace" :edit-name="editName" @success="load" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import NetworkPolicyFormDialog from '@/components/policy/NetworkPolicyFormDialog.vue'
import NetworkPolicySupportAlert from '@/components/policy/NetworkPolicySupportAlert.vue'
import { resourceApi, type KubeObject } from '@/api/resource'
import { tableHeaderClass } from '@/utils/table'
import { formatDateTime } from '@/utils/job'
import { directionText, podSelectorText } from '@/utils/networkPolicy'

const props = defineProps<{ clusterId: string; namespace: string }>()
const emit = defineEmits<{ change: [count: number] }>()

const NP_REF = { kind: 'networkpolicy' as const }
const loading = ref(false)
const items = ref<KubeObject[]>([])
const formVisible = ref(false)
const editName = ref('')

const load = async () => {
  loading.value = true
  try {
    const { data } = await resourceApi.objects({ ...NP_REF, clusterId: props.clusterId, namespace: props.namespace })
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
    await ElMessageBox.confirm(`确定要删除网络策略 "${name}" 吗？删除后相关容器组的流量限制会立即解除。`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await resourceApi.delete({ ...NP_REF, clusterId: props.clusterId, namespace: props.namespace, name })
    ElMessage.success('删除成功')
    load()
  } catch {
    // 请求拦截器已提示错误
  }
}

onMounted(load)
</script>

<style scoped>
.npp-actions {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
}

.npp-name {
  color: #1e293b;
  font-weight: 500;
}

.mono {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12.5px;
}
</style>
