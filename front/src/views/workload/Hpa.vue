<template>
  <div>
    <ResourceTablePage
      ref="page"
      :resource="HPA_REF"
      kind-label="HorizontalPodAutoscaler"
      title="弹性伸缩"
      subtitle="根据 CPU / 内存使用率自动调整 Deployment、StatefulSet 的副本数（需要 metrics-server）"
      icon="hpa"
      create-text="创建弹性伸缩"
      :actions="[{ command: 'edit', label: '编辑' }]"
      :yaml-template="HPA_YAML"
      empty-text="暂无弹性伸缩"
      @create="openCreate"
      @action="handleAction"
    >
      <template #summary="{ obj }">
        <div class="hpa-summary">
          <div><label>目标</label><strong>{{ obj.spec?.scaleTargetRef?.kind }} / {{ obj.spec?.scaleTargetRef?.name }}</strong></div>
          <div><label>副本范围</label><strong>{{ obj.spec?.minReplicas ?? 1 }} ~ {{ obj.spec?.maxReplicas }}</strong></div>
          <div><label>当前 / 期望副本</label><strong>{{ obj.status?.currentReplicas ?? '-' }} / {{ obj.status?.desiredReplicas ?? '-' }}</strong></div>
        </div>
        <el-table :data="hpaMetricRows(obj)" size="small" style="width: 100%">
          <el-table-column label="指标" prop="label" min-width="180" />
          <el-table-column label="当前" prop="current" min-width="140" />
          <el-table-column label="目标" prop="target" min-width="140" />
        </el-table>
      </template>
    </ResourceTablePage>
    <HpaFormDialog v-model="formVisible" :cluster-id="clusterId" :namespace="formNamespace" :edit-name="editName" @success="page?.refresh()" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import ResourceTablePage from '@/components/resource/ResourceTablePage.vue'
import HpaFormDialog from '@/components/workload/HpaFormDialog.vue'
import type { ResourceTableRow } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { HPA_YAML } from '@/utils/resourceTemplates'
import { hpaMetricRows } from '@/utils/hpa'

const HPA_REF = { kind: 'horizontalpodautoscaler' as const }
const clusterStore = useClusterStore()
const clusterId = computed(() => clusterStore.currentClusterId)
const page = ref<InstanceType<typeof ResourceTablePage> | null>(null)
const formVisible = ref(false)
const formNamespace = ref('')
const editName = ref('')

const openCreate = (namespace: string) => {
  formNamespace.value = namespace
  editName.value = ''
  formVisible.value = true
}

const handleAction = (command: string, row: ResourceTableRow) => {
  if (command !== 'edit') return
  formNamespace.value = row.namespace
  editName.value = row.name
  formVisible.value = true
}
</script>

<style scoped>
.hpa-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 12px;
}

.hpa-summary div {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #f8fafc;
}

.hpa-summary label {
  color: #64748b;
  font-size: 12px;
}
</style>
