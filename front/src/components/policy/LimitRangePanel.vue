<template>
  <div v-loading="loading">
    <div class="lp-actions">
      <el-button size="small" type="primary" @click="openForm('')">
        <el-icon><Plus /></el-icon>
        {{ items.length ? '新建默认资源限制' : '设置默认资源限制' }}
      </el-button>
    </div>
    <el-empty v-if="!items.length" description="没有默认资源限制，容器没有声明资源时不受限制" :image-size="72" />
    <div v-for="l in items" :key="l.metadata.name" class="lp-card">
      <div class="lp-head">
        <span class="lp-name">{{ l.metadata.name }}</span>
        <div class="lp-ops">
          <el-button size="small" @click="openForm(l.metadata.name)">编辑</el-button>
          <el-button size="small" type="danger" plain @click="remove(l.metadata.name)">删除</el-button>
        </div>
      </div>
      <el-table :data="limitRows(l.spec?.limits)" size="small" style="width: 100%" :header-cell-class-name="tableHeaderClass">
        <el-table-column label="对象类型" width="120">
          <template #default="{ row }">{{ LIMIT_TYPE_TEXT[row.type] || row.type }}</template>
        </el-table-column>
        <el-table-column label="资源" width="110">
          <template #default="{ row }">{{ RESOURCE_NAME_TEXT[row.resource] || row.resource }}</template>
        </el-table-column>
        <el-table-column v-for="f in LIMIT_FIELDS" :key="f.key" :label="f.label" min-width="110">
          <template #default="{ row }"><span class="mono">{{ row.values[f.key] ?? '-' }}</span></template>
        </el-table-column>
      </el-table>
    </div>
    <LimitRangeFormDialog v-model="formVisible" :cluster-id="clusterId" :namespace="namespace" :edit-name="editName" @success="load" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import LimitRangeFormDialog from '@/components/policy/LimitRangeFormDialog.vue'
import { resourceApi, type KubeObject } from '@/api/resource'
import { tableHeaderClass } from '@/utils/table'
import { LIMIT_FIELDS, LIMIT_TYPE_TEXT, RESOURCE_NAME_TEXT, limitRows } from '@/utils/policy'

const props = defineProps<{ clusterId: string; namespace: string }>()
const emit = defineEmits<{ change: [count: number] }>()

const LR_REF = { kind: 'limitrange' as const }
const loading = ref(false)
const items = ref<KubeObject[]>([])
const formVisible = ref(false)
const editName = ref('')

const load = async () => {
  loading.value = true
  try {
    const { data } = await resourceApi.objects({ ...LR_REF, clusterId: props.clusterId, namespace: props.namespace })
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
    await ElMessageBox.confirm(`确定要删除默认资源限制 "${name}" 吗？`, '确认删除', { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' })
  } catch {
    return
  }
  try {
    await resourceApi.delete({ ...LR_REF, clusterId: props.clusterId, namespace: props.namespace, name })
    ElMessage.success('删除成功')
    load()
  } catch {
    // 请求拦截器已提示错误
  }
}

onMounted(load)
</script>

<style scoped>
.lp-actions {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
}

.lp-card {
  padding: 14px 16px;
  border: 1px solid #e6ecf5;
  border-radius: 12px;
  background: #fbfcfe;
}

.lp-card + .lp-card {
  margin-top: 12px;
}

.lp-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.lp-name {
  color: #1e293b;
  font-weight: 600;
}

.lp-ops {
  margin-left: auto;
}

.mono {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
