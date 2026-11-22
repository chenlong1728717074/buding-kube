<template>
  <div>
    <ResourceTablePage
      ref="page"
      :resource="{ kind: 'networkpolicy' }"
      kind-label="NetworkPolicy"
      title="网络策略"
      subtitle="控制容器组之间及与外部的网络访问，需要网络插件（Calico、Cilium 等）支持"
      icon="netpol"
      create-text="创建网络策略"
      :actions="[{ command: 'edit', label: '编辑' }]"
      :yaml-template="NETWORK_POLICY_YAML"
      @create="openCreate"
      @action="handleAction"
    >
      <template #notice>
        <NetworkPolicySupportAlert :cluster-id="clusterId" show-supported />
      </template>
      <template #summary="{ obj }">
        <div class="np-summary">
          <div><label>应用于</label><strong>{{ podSelectorText(obj.spec) }}</strong></div>
          <div><label>入站</label><strong>{{ directionText(obj.spec, 'Ingress') }}</strong></div>
          <div><label>出站</label><strong>{{ directionText(obj.spec, 'Egress') }}</strong></div>
        </div>
      </template>
    </ResourceTablePage>
    <NetworkPolicyFormDialog v-model="formVisible" :cluster-id="clusterId" :namespace="formNamespace" :edit-name="editName" @success="page?.refresh()" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import ResourceTablePage from '@/components/resource/ResourceTablePage.vue'
import NetworkPolicyFormDialog from '@/components/policy/NetworkPolicyFormDialog.vue'
import NetworkPolicySupportAlert from '@/components/policy/NetworkPolicySupportAlert.vue'
import type { ResourceTableRow } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { NETWORK_POLICY_YAML } from '@/utils/resourceTemplates'
import { directionText, podSelectorText } from '@/utils/networkPolicy'

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
.np-summary {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.np-summary div {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #f8fafc;
}

.np-summary label {
  color: #64748b;
  font-size: 12px;
}

.np-summary strong {
  word-break: break-all;
}
</style>
