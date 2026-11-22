<template>
  <div>
    <ResourceTablePage
      ref="page"
      :resource="{ kind: 'resourcequota' }"
      kind-label="ResourceQuota"
      title="资源配额"
      subtitle="限制命名空间可以使用的 CPU、内存、存储和对象数量总量"
      icon="quota"
      create-text="设置资源配额"
      :actions="[{ command: 'edit', label: '编辑' }]"
      :yaml-template="QUOTA_YAML"
      @create="openCreate"
      @action="handleAction"
    >
      <template #summary="{ obj }">
        <el-table :data="quotaUsageRows(obj.status?.hard || obj.spec?.hard, obj.status?.used)" size="small" style="width: 100%">
          <el-table-column label="配额项" prop="label" min-width="160" />
          <el-table-column label="已用 / 上限" min-width="150">
            <template #default="{ row }"><span class="mono">{{ row.used }} / {{ row.hard }}</span></template>
          </el-table-column>
          <el-table-column label="使用率" min-width="160">
            <template #default="{ row }">
              <el-progress v-if="row.percent !== null" :percentage="Math.min(row.percent, 100)" :status="row.percent >= 90 ? 'exception' : row.percent >= 70 ? 'warning' : undefined" />
              <span v-else>-</span>
            </template>
          </el-table-column>
        </el-table>
      </template>
    </ResourceTablePage>
    <QuotaFormDialog v-model="formVisible" :cluster-id="clusterId" :namespace="formNamespace" :edit-name="editName" @success="page?.refresh()" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import ResourceTablePage from '@/components/resource/ResourceTablePage.vue'
import QuotaFormDialog from '@/components/policy/QuotaFormDialog.vue'
import type { ResourceTableRow } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { QUOTA_YAML } from '@/utils/resourceTemplates'
import { quotaUsageRows } from '@/utils/policy'

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
.mono {
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
}
</style>
