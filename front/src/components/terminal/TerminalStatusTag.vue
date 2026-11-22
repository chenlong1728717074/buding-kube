<template>
  <span :class="['ts-tag', `is-${status}`]"><i />{{ TEXT[status] }}</span>
</template>

<script setup lang="ts">
import type { TerminalStatus } from '@/utils/terminal'

defineProps<{ status: TerminalStatus }>()

const TEXT: Record<TerminalStatus, string> = { connecting: '连接中', connected: '已连接', closed: '已断开' }
</script>

<style scoped>
.ts-tag {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 10px;
  border-radius: 10px;
  background: rgba(148, 163, 184, 0.15);
  color: #94a3b8;
  font-size: 12px;
  white-space: nowrap;
}

.ts-tag i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
}

.ts-tag.is-connected {
  background: rgba(74, 222, 128, 0.15);
  color: #4ade80;
}

.ts-tag.is-connecting {
  background: rgba(250, 204, 21, 0.15);
  color: #facc15;
}

.ts-tag.is-connecting i {
  animation: ts-blink 1s infinite;
}

@keyframes ts-blink {
  50% {
    opacity: 0.3;
  }
}
</style>
