<template>
  <div v-loading="loading">
    <el-empty v-if="!items.length" description="还没有设置弹性伸缩，副本数需要手动调整" :image-size="80">
      <el-button type="primary" @click="openCreate">
        <el-icon><Plus /></el-icon>
        设置弹性伸缩
      </el-button>
    </el-empty>
    <div v-for="h in items" :key="h.metadata.name" class="hp-card">
      <div class="hp-head">
        <span class="hp-name">{{ h.metadata.name }}</span>
        <el-tag v-if="limited(h)" type="warning" size="small">达到副本上下限</el-tag>
        <div class="hp-actions">
          <el-button size="small" @click="openEdit(h.metadata.name)">编辑</el-button>
          <el-button size="small" type="danger" plain @click="remove(h.metadata.name)">删除</el-button>
        </div>
      </div>
      <div class="hp-stats">
        <div class="hp-stat"><label>副本范围</label><strong>{{ h.spec.minReplicas ?? 1 }} ~ {{ h.spec.maxReplicas }}</strong></div>
        <div class="hp-stat"><label>当前副本</label><strong>{{ h.status?.currentReplicas ?? '-' }}</strong></div>
        <div class="hp-stat"><label>期望副本</label><strong>{{ h.status?.desiredReplicas ?? '-' }}</strong></div>
        <div class="hp-stat"><label>最近伸缩</label><strong>{{ formatDateTime(h.status?.lastScaleTime) }}</strong></div>
        <div class="hp-stat"><label>缩容稳定窗口</label><strong>{{ h.spec.behavior?.scaleDown?.stabilizationWindowSeconds ?? 300 }} 秒</strong></div>
      </div>
      <el-table :data="hpaMetricRows(h)" size="small" style="width: 100%" :header-cell-class-name="tableHeaderClass">
        <el-table-column label="指标" min-width="200" prop="label" />
        <el-table-column label="当前" min-width="160" prop="current" />
        <el-table-column label="目标" min-width="160" prop="target" />
      </el-table>
      <div class="hp-conds">
        <span v-for="c in h.status?.conditions || []" :key="c.type" :class="['hp-cond', conditionOk(c) ? 'is-ok' : 'is-bad']" :title="c.message">
          {{ HPA_CONDITION_TEXT[c.type] || c.type }}：{{ c.status === 'True' ? '是' : '否' }}<em v-if="c.status !== 'True' || c.type === 'ScalingLimited'">（{{ c.reason }}）</em>
        </span>
      </div>
    </div>
    <HpaFormDialog
      v-model="formVisible"
      :cluster-id="clusterId"
      :namespace="namespace"
      :target="{ kind: targetKind, name }"
      :edit-name="editName"
      @success="load"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import HpaFormDialog from '@/components/workload/HpaFormDialog.vue'
import { resourceApi, type KubeObject } from '@/api/resource'
import { tableHeaderClass } from '@/utils/table'
import { formatDateTime } from '@/utils/job'
import { HPA_CONDITION_TEXT, hpaMetricRows, hpaTargetKind } from '@/utils/hpa'

// 工作负载详情的「弹性伸缩」: 显示指向该工作负载的 HPA
const props = defineProps<{
  clusterId: string
  namespace: string
  kind: 'deployment' | 'statefulset'
  name: string
}>()

const emit = defineEmits<{ change: [count: number] }>()

const HPA_REF = { kind: 'horizontalpodautoscaler' as const }
const targetKind = computed(() => hpaTargetKind(props.kind))
const loading = ref(false)
const items = ref<KubeObject[]>([])
const formVisible = ref(false)
const editName = ref('')

// ScalingLimited 为 True 表示已经达到上下限, 属于需要关注的状态
const conditionOk = (c: KubeObject) => (c.type === 'ScalingLimited' ? c.status !== 'True' : c.status === 'True')
const limited = (h: KubeObject) => (h.status?.conditions || []).some((c: KubeObject) => c.type === 'ScalingLimited' && c.status === 'True')

const load = async () => {
  loading.value = true
  try {
    const { data } = await resourceApi.objects({ ...HPA_REF, clusterId: props.clusterId, namespace: props.namespace })
    items.value = (data || []).filter(h => h.spec?.scaleTargetRef?.kind === targetKind.value && h.spec?.scaleTargetRef?.name === props.name)
  } catch {
    items.value = []
  } finally {
    loading.value = false
    emit('change', items.value.length)
  }
}

const openCreate = () => {
  editName.value = ''
  formVisible.value = true
}

const openEdit = (name: string) => {
  editName.value = name
  formVisible.value = true
}

const remove = async (name: string) => {
  try {
    await ElMessageBox.confirm(`删除弹性伸缩 "${name}" 后，副本数保持当前值，需要手动调整。确定删除吗？`, '确认删除', {
      type: 'warning',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  try {
    await resourceApi.delete({ ...HPA_REF, clusterId: props.clusterId, namespace: props.namespace, name })
    ElMessage.success('删除成功')
    load()
  } catch {
    // 请求拦截器已提示错误
  }
}

onMounted(load)
defineExpose({ reload: load })
</script>

<style scoped>
.hp-card {
  padding: 16px 18px;
  border: 1px solid #e6ecf5;
  border-radius: 12px;
  background: #fbfcfe;
}

.hp-card + .hp-card {
  margin-top: 14px;
}

.hp-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}

.hp-name {
  color: #1e293b;
  font-size: 15px;
  font-weight: 600;
}

.hp-actions {
  margin-left: auto;
}

.hp-stats {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 14px;
}

.hp-stat {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #fff;
}

.hp-stat label {
  color: #64748b;
  font-size: 12px;
}

.hp-stat strong {
  color: #1e293b;
  font-size: 15px;
}

.hp-conds {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.hp-cond {
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
}

.hp-cond em {
  font-style: normal;
  opacity: 0.8;
}

.hp-cond.is-ok {
  background: #f0fdf4;
  color: #16a34a;
}

.hp-cond.is-bad {
  background: #fef2f2;
  color: #dc2626;
}
</style>
