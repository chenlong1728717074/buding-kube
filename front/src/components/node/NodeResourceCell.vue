<template>
  <el-tooltip placement="top" :show-after="200">
    <template #content>
      <div class="nr-tip">
        <div>实时用量：{{ used === null || used === undefined ? '无数据（需要 metrics-server）' : `${fmt(used)}（${usedPercent}%）` }}</div>
        <div>已分配（预留）：{{ fmt(requests) }}（{{ requestPercent ?? 0 }}%）</div>
        <div>可分配：{{ fmt(allocatable) }}</div>
      </div>
    </template>
    <div class="nr">
      <div class="nr-line">
        <span class="nr-label">用量</span>
        <div class="nr-bar"><i :class="percentLevel(usedPercent)" :style="{ width: `${Math.min(usedPercent ?? 0, 100)}%` }" /></div>
        <span class="nr-num">{{ usedPercent === null ? '-' : `${usedPercent}%` }}</span>
      </div>
      <div class="nr-line">
        <span class="nr-label">分配</span>
        <div class="nr-bar"><i class="is-alloc" :style="{ width: `${Math.min(requestPercent ?? 0, 100)}%` }" /></div>
        <span class="nr-num">{{ requestPercent === null ? '-' : `${requestPercent}%` }}</span>
      </div>
    </div>
  </el-tooltip>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { formatCpu, formatMemory } from '@/utils/resource'
import { percentLevel, percentOf } from '@/utils/node'

// 节点列表中的资源单元格: 实时用量和已分配(容器组预留合计)占可分配量的比例
const props = withDefaults(
  defineProps<{
    kind: 'cpu' | 'memory'
    used?: number | null
    requests: number
    allocatable: number
  }>(),
  { used: null }
)

const fmt = (v?: number | null) => (props.kind === 'cpu' ? formatCpu(v) : formatMemory(v))
const usedPercent = computed(() => percentOf(props.used, props.allocatable))
const requestPercent = computed(() => percentOf(props.requests, props.allocatable))
</script>

<style scoped>
.nr {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 130px;
}

.nr-line {
  display: flex;
  align-items: center;
  gap: 6px;
}

.nr-label {
  flex: 0 0 26px;
  color: #94a3b8;
  font-size: 12px;
}

.nr-bar {
  flex: 1;
  height: 6px;
  border-radius: 3px;
  background: #e8eef8;
  overflow: hidden;
}

.nr-bar i {
  display: block;
  height: 100%;
  border-radius: 3px;
}

.nr-bar i.is-ok {
  background: #3b74f2;
}

.nr-bar i.is-warning {
  background: #f59e0b;
}

.nr-bar i.is-danger {
  background: #ef4444;
}

.nr-bar i.is-alloc {
  background: #94a3b8;
}

.nr-num {
  flex: 0 0 36px;
  color: #334155;
  font-family: 'JetBrains Mono', Consolas, Menlo, monospace;
  font-size: 12px;
  text-align: right;
}

.nr-tip {
  line-height: 1.7;
}
</style>
