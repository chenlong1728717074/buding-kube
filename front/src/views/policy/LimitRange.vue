<template>
  <div>
    <ResourceTablePage
      ref="page"
      :resource="{ kind: 'limitrange' }"
      kind-label="LimitRange"
      title="默认资源限制"
      subtitle="为容器设置默认的资源预留 / 上限，以及允许声明的范围"
      icon="limits"
      create-text="设置默认资源限制"
      :actions="[{ command: 'edit', label: '编辑' }]"
      :yaml-template="LIMIT_RANGE_YAML"
      @create="openCreate"
      @action="handleAction"
    >
      <template #summary="{ obj }">
        <el-table :data="limitRows(obj.spec?.limits)" size="small" style="width: 100%">
          <el-table-column label="对象类型" width="110">
            <template #default="{ row }">{{ LIMIT_TYPE_TEXT[row.type] || row.type }}</template>
          </el-table-column>
          <el-table-column label="资源" width="90">
            <template #default="{ row }">{{ RESOURCE_NAME_TEXT[row.resource] || row.resource }}</template>
          </el-table-column>
          <el-table-column v-for="f in LIMIT_FIELDS" :key="f.key" :label="f.label" min-width="100">
            <template #default="{ row }">{{ row.values[f.key] ?? '-' }}</template>
          </el-table-column>
        </el-table>
      </template>
    </ResourceTablePage>
    <LimitRangeFormDialog v-model="formVisible" :cluster-id="clusterId" :namespace="formNamespace" :edit-name="editName" @success="page?.refresh()" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import ResourceTablePage from '@/components/resource/ResourceTablePage.vue'
import LimitRangeFormDialog from '@/components/policy/LimitRangeFormDialog.vue'
import type { ResourceTableRow } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { LIMIT_RANGE_YAML } from '@/utils/resourceTemplates'
import { LIMIT_FIELDS, LIMIT_TYPE_TEXT, RESOURCE_NAME_TEXT, limitRows } from '@/utils/policy'

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
