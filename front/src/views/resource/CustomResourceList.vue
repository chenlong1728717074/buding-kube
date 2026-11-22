<template>
  <div v-if="resourceRef" class="cr-page">
    <ResourceTablePage
      :resource="resourceRef"
      :kind-label="crd.kind"
      :title="crd.kind"
      :subtitle="`自定义资源 ${crd.name}`"
      icon="crd"
      :namespaced="crd.namespaced"
      :yaml-template="(ns: string) => customResourceYaml(`${crd.group}/${version}`, crd.kind, crd.namespaced ? ns : undefined)"
      :empty-text="`还没有 ${crd.kind} 资源`"
      back
      @back="goBack"
    >
      <template #header>
        <el-select v-if="crd.versions.length > 1" v-model="version" style="width: 140px" title="API 版本">
          <el-option v-for="v in crd.versions" :key="v" :label="v" :value="v" />
        </el-select>
      </template>
    </ResourceTablePage>
  </div>
  <div v-else v-loading="loading" class="cr-page cr-page--empty">
    <el-empty v-if="!loading" :description="errorText || 'CRD 不存在'" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import ResourceTablePage from '@/components/resource/ResourceTablePage.vue'
import { crdApi } from '@/api/crd'
import type { ResourceRef } from '@/api/resource'
import { useClusterStore } from '@/stores/cluster'
import { useTabBack } from '@/composables/useTabBack'
import { customResourceYaml } from '@/utils/resourceTemplates'

// CRD 的实例列表; 地址参数 name 为 CRD 名称(例如 certificates.cert-manager.io)
const route = useRoute()
const goBack = useTabBack()
const clusterStore = useClusterStore()

const crdName = (route.query.name as string) || ''
const loading = ref(true)
const errorText = ref('')
const version = ref('')
const crd = reactive({ name: crdName, group: '', kind: '', plural: '', namespaced: true, versions: [] as string[] })

const resourceRef = computed<ResourceRef | null>(() =>
  crd.kind && version.value
    ? { kind: 'custom', group: crd.group, version: version.value, resource: crd.plural, kindName: crd.kind, namespaced: crd.namespaced }
    : null
)

onMounted(async () => {
  if (!crdName || !clusterStore.currentClusterId) {
    loading.value = false
    errorText.value = '缺少 CRD 名称'
    return
  }
  try {
    // 只取 CRD 概要, 不加载 schema
    const { data } = await crdApi.getCrd(clusterStore.currentClusterId, crdName)
    crd.group = data.group
    crd.kind = data.kind
    crd.plural = data.plural
    crd.namespaced = data.scope !== 'Cluster'
    crd.versions = data.versions || []
    version.value = crd.versions[0] || ''
    if (!version.value) errorText.value = '该 CRD 没有可用的 API 版本'
  } catch {
    errorText.value = 'CRD 不存在或已被删除'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.cr-page--empty {
  min-height: 300px;
}
</style>
