<template>
  <ResourceTablePage
    :resource="{ kind: 'replicaset' }"
    kind-label="ReplicaSet"
    title="ReplicaSet"
    subtitle="副本集，通常由 Deployment 自动管理，每个版本对应一个副本集"
    icon="rs"
    :yaml-template="REPLICASET_YAML"
  >
    <template #summary="{ obj }">
      <div class="rs-summary">
        <div><label>期望副本</label><strong>{{ obj.spec?.replicas ?? 0 }}</strong></div>
        <div><label>就绪</label><strong>{{ obj.status?.readyReplicas ?? 0 }}</strong></div>
        <div><label>可用</label><strong>{{ obj.status?.availableReplicas ?? 0 }}</strong></div>
        <div><label>版本号</label><strong>{{ obj.metadata?.annotations?.['deployment.kubernetes.io/revision'] || '-' }}</strong></div>
      </div>
      <p class="rs-images">
        镜像：<span v-for="c in obj.spec?.template?.spec?.containers || []" :key="c.name" class="rs-image">{{ c.name }} = {{ c.image }}</span>
      </p>
    </template>
  </ResourceTablePage>
</template>

<script setup lang="ts">
import ResourceTablePage from '@/components/resource/ResourceTablePage.vue'
import { REPLICASET_YAML } from '@/utils/resourceTemplates'
</script>

<style scoped>
.rs-summary {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 10px;
}

.rs-summary div {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 8px;
  background: #f8fafc;
}

.rs-summary label {
  color: #64748b;
  font-size: 12px;
}

.rs-images {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  color: #64748b;
  font-size: 13px;
}

.rs-image {
  padding: 1px 8px;
  border-radius: 6px;
  background: #eef4ff;
  color: #1d4ed8;
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12px;
}
</style>
