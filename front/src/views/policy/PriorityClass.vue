<template>
  <ResourceTablePage
    :resource="{ kind: 'priorityclass' }"
    kind-label="PriorityClass"
    title="优先级"
    subtitle="容器组的调度优先级，资源不足时高优先级的容器组可以抢占低优先级的"
    icon="etcd"
    :namespaced="false"
    :yaml-template="PRIORITY_CLASS_YAML"
  />
</template>

<script setup lang="ts">
import ResourceTablePage from '@/components/resource/ResourceTablePage.vue'
import { PRIORITY_CLASS_YAML } from '@/utils/resourceTemplates'
</script>
