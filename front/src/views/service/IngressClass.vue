<template>
  <ResourceTablePage
    :resource="{ kind: 'ingressclass' }"
    kind-label="IngressClass"
    title="IngressClass"
    subtitle="集群中可用的 Ingress 控制器，创建 Ingress 时选择"
    icon="ing"
    :namespaced="false"
    :yaml-template="INGRESS_CLASS_YAML"
  />
</template>

<script setup lang="ts">
import ResourceTablePage from '@/components/resource/ResourceTablePage.vue'
import { INGRESS_CLASS_YAML } from '@/utils/resourceTemplates'
</script>
