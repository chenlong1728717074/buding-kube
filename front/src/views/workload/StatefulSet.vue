<template>
  <WorkloadListPage kind="statefulset" />
</template>

<script setup lang="ts">
// 列表逻辑在 WorkloadListPage 中, 三种工作负载共用
import WorkloadListPage from '@/components/workload/WorkloadListPage.vue'
</script>
