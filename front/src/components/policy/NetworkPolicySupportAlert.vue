<template>
  <el-alert
    v-if="info && (info.support !== 'supported' || showSupported)"
    :type="TYPE[info.support]"
    :closable="info.support === 'supported'"
    show-icon
    class="nps-alert"
    :title="info.message"
  />
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { capabilityApi, type NetworkPolicySupportVO } from '@/api/capability'

// 网络插件是否支持网络策略: 不支持或无法识别时提醒, 支持时按需显示
const props = withDefaults(defineProps<{ clusterId: string; showSupported?: boolean }>(), { showSupported: false })

const TYPE = { supported: 'success', unsupported: 'warning', unknown: 'info' } as const
const info = ref<NetworkPolicySupportVO | null>(null)

watch(
  () => props.clusterId,
  async id => {
    info.value = null
    if (!id) return
    const result = await capabilityApi.networkPolicy(id)
    // 等待期间切换了集群时丢弃旧结果
    if (id === props.clusterId) info.value = result
  },
  { immediate: true }
)
</script>

<style scoped>
.nps-alert {
  margin-bottom: 14px;
}
</style>
