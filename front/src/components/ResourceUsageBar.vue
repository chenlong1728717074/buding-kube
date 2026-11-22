<template>
  <el-tooltip :disabled="used === null || used === undefined" placement="top" :show-after="200">
    <template #content>
      <div class="ru-tip">
        <div>用量：{{ formatUsage(kind, used) }}</div>
        <div>预留：{{ request ? formatUsage(kind, request) : '未设置' }}</div>
        <div>上限：{{ limit ? formatUsage(kind, limit) : '不限制' }}</div>
        <div v-if="percent !== null">占{{ limit ? '上限' : '预留' }}的 {{ percent }}%</div>
      </div>
    </template>
    <div :class="['ru', { 'is-compact': compact }]">
      <span v-if="used === null || used === undefined" class="ru-empty">-</span>
      <template v-else>
        <span class="ru-value">{{ formatUsage(kind, used) }}<em v-if="limit">/ {{ formatUsage(kind, limit) }}</em></span>
        <div v-if="percent !== null" class="ru-bar">
          <i :class="level" :style="{ width: `${Math.min(percent, 100)}%` }" />
        </div>
      </template>
    </div>
  </el-tooltip>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatUsage, usagePercent, type UsageKind } from '@/utils/resource'

// 资源用量: 数值 + 占上限(没有上限时占预留)的进度条
const props = withDefaults(
  defineProps<{
    kind: UsageKind
    used?: number | null
    request?: number
    limit?: number
    compact?: boolean
  }>(),
  { used: null, request: 0, limit: 0, compact: false }
)

const percent = computed(() => usagePercent(props.used, props.request, props.limit))

// 超过上限 90% 有被限流(CPU)或 OOM(内存)的风险; 没有上限时超过预留只是提示
const level = computed(() => {
  const p = percent.value ?? 0
  if (props.limit) return p >= 90 ? 'is-danger' : p >= 70 ? 'is-warning' : 'is-ok'
  return p > 100 ? 'is-warning' : 'is-ok'
})
</script>

<style scoped>
.ru {
  display: inline-flex;
  flex-direction: column;
  align-items: stretch;
  gap: 4px;
  min-width: 88px;
  line-height: 1.2;
}

.ru.is-compact {
  min-width: 70px;
}

.ru-value {
  color: #1e293b;
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12.5px;
  white-space: nowrap;
}

.ru-value em {
  margin-left: 4px;
  color: #94a3b8;
  font-style: normal;
}

.ru-empty {
  color: #94a3b8;
}

.ru-bar {
  height: 4px;
  border-radius: 2px;
  background: #e8eef8;
  overflow: hidden;
}

.ru-bar i {
  display: block;
  height: 100%;
  border-radius: 2px;
  transition: width 0.3s ease;
}

.ru-bar i.is-ok {
  background: #3b74f2;
}

.ru-bar i.is-warning {
  background: #f59e0b;
}

.ru-bar i.is-danger {
  background: #ef4444;
}

.ru-tip {
  line-height: 1.7;
}
</style>
