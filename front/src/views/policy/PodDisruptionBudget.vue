<template>
  <ResourceTablePage
    :resource="{ kind: 'poddisruptionbudget' }"
    kind-label="PodDisruptionBudget"
    title="中断预算"
    subtitle="限制排空节点、集群升级等主动操作时同时不可用的容器组数量"
    icon="quota"
    
    :yaml-template="PDB_YAML"
  />
</template>

<script setup lang="ts">
import ResourceTablePage from '@/components/resource/ResourceTablePage.vue'
import { PDB_YAML } from '@/utils/resourceTemplates'
</script>
